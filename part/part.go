package part

import (
	"github.com/cyrilix/robocar-base/mqttdevice"
	"github.com/cyrilix/robocar-base/service"
	"github.com/cyrilix/robocar-base/types"
	mqtt "github.com/eclipse/paho.mqtt.golang"
	"log"
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
	payload := message.Payload()
	value := mqttdevice.NewMqttValue(payload)
	m, err := value.DriveModeValue()
	if err != nil {
		log.Printf("invalid drive mode: %v", err)
		return
	}

	p.muDriveMode.Lock()
	defer p.muDriveMode.Unlock()
	p.driveMode = m
}

func (p *ThrottlePart) onRCThrottle(_ mqtt.Client, message mqtt.Message) {
	payload := message.Payload()
	value := mqttdevice.NewMqttValue(payload)
	val, err := value.Float64Value()
	if err != nil {
		log.Printf("invalid throttle value from arduino: %v", err)
		return
	}

	p.muDriveMode.RLock()
	defer p.muDriveMode.RUnlock()
	if p.driveMode == types.DriveModeUser {
		p.pub.Publish(p.throttleTopic, mqttdevice.NewMqttValue(types.Throttle{Value: val, Confidence: 1.0}))
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
