package part

import (
	"encoding/json"
	"github.com/cyrilix/robocar-base/mqttdevice"
	"github.com/cyrilix/robocar-base/service"
	"github.com/cyrilix/robocar-base/types"
	mqtt "github.com/eclipse/paho.mqtt.golang"
	log "github.com/sirupsen/logrus"
	"sync"
	"time"
)

func NewPart(client mqtt.Client, pub mqttdevice.Publisher, throttleTopic, driveModeTopic, rcThrottleTopic string,
	minValue, maxValue float64, publishPilotFrequency int) *ThrottlePart {
	return &ThrottlePart{
		client:                client,
		pub:                   pub,
		throttleTopic:         throttleTopic,
		driveModeTopic:        driveModeTopic,
		rcThrottleTopic:       rcThrottleTopic,
		minThrottle:           minValue,
		maxThrottle:           maxValue,
		driveMode:             types.DriveModeUser,
		publishPilotFrequency: publishPilotFrequency,
	}

}

type ThrottlePart struct {
	client                   mqtt.Client
	pub                      mqttdevice.Publisher
	throttleTopic            string
	minThrottle, maxThrottle float64

	muDriveMode sync.RWMutex
	driveMode   types.DriveMode

	cancel                          chan interface{}
	publishPilotFrequency           int
	driveModeTopic, rcThrottleTopic string
}

func (p *ThrottlePart) Start() error {
	if err := registerCallbacks(p); err != nil {
		log.Printf("unable to rgeister callbacks: %v", err)
		return err
	}

	p.cancel = make(chan interface{})
	ticker := time.NewTicker(1 * time.Second / time.Duration(p.publishPilotFrequency))
	for {
		select {
		case <-ticker.C:
			p.publishPilotValue()
		case <-p.cancel:
			break
		}
	}
}

func (p *ThrottlePart) publishPilotValue() {
	p.muDriveMode.RLock()
	defer p.muDriveMode.RUnlock()

	if p.driveMode != types.DriveModePilot {
		return
	}

	p.pub.Publish(p.throttleTopic, mqttdevice.NewMqttValue(types.Throttle{
		Value:      p.minThrottle,
		Confidence: 1.0,
	}))
}

func (p *ThrottlePart) Stop() {
	close(p.cancel)
	service.StopService("throttle", p.client, p.driveModeTopic, p.rcThrottleTopic)
}

func (p *ThrottlePart) onDriveMode(_ mqtt.Client, message mqtt.Message) {
	m := types.ParseString(string(message.Payload()))

	p.muDriveMode.Lock()
	defer p.muDriveMode.Unlock()
	p.driveMode = m
}

func (p *ThrottlePart) onRCThrottle(_ mqtt.Client, message mqtt.Message) {
	payload := message.Payload()
	var throttle types.Throttle
	err := json.Unmarshal(payload, &throttle)
	if err != nil {
		log.Errorf("unable to parse throttle json: %v", err)
		return
	}

	p.muDriveMode.RLock()
	defer p.muDriveMode.RUnlock()
	if p.driveMode == types.DriveModeUser {
		p.pub.Publish(p.throttleTopic, mqttdevice.NewMqttValue(throttle))
	}
}

var registerCallbacks = func (p *ThrottlePart) error {
	err := service.RegisterCallback(p.client, p.driveModeTopic, p.onDriveMode)
	if err != nil {
		return err
	}

	err = service.RegisterCallback(p.client, p.rcThrottleTopic, p.onRCThrottle)
	if err != nil {
		return err
	}
	return nil
}
