package part

import (
	"github.com/cyrilix/robocar-base/service"
	"github.com/cyrilix/robocar-protobuf/go/events"
	mqtt "github.com/eclipse/paho.mqtt.golang"
	"go.uber.org/zap"
	"google.golang.org/protobuf/proto"
	"sync"
	"time"
)

func NewPart(client mqtt.Client, throttleTopic, driveModeTopic, rcThrottleTopic string, minValue, maxValue float32, publishPilotFrequency int) *ThrottlePart {
	return &ThrottlePart{
		client:                client,
		throttleTopic:         throttleTopic,
		driveModeTopic:        driveModeTopic,
		rcThrottleTopic:       rcThrottleTopic,
		minThrottle:           minValue,
		maxThrottle:           maxValue,
		driveMode:             events.DriveMode_USER,
		publishPilotFrequency: publishPilotFrequency,
	}

}

type ThrottlePart struct {
	client                   mqtt.Client
	throttleTopic            string
	minThrottle, maxThrottle float32

	muDriveMode sync.RWMutex
	driveMode   events.DriveMode

	cancel                          chan interface{}
	publishPilotFrequency           int
	driveModeTopic, rcThrottleTopic string
}

func (p *ThrottlePart) Start() error {
	if err := registerCallbacks(p); err != nil {
		zap.S().Errorf("unable to register callbacks: %v", err)
		return err
	}

	p.cancel = make(chan interface{})
	ticker := time.NewTicker(1 * time.Second / time.Duration(p.publishPilotFrequency))
	for {
		select {
		case <-ticker.C:
			p.onPublishPilotValue()
		case <-p.cancel:
			break
		}
	}
}

func (p *ThrottlePart) onPublishPilotValue() {
	p.muDriveMode.RLock()
	defer p.muDriveMode.RUnlock()

	if p.driveMode != events.DriveMode_PILOT {
		return
	}

	throttleMsg := events.ThrottleMessage{
		Throttle:   p.minThrottle,
		Confidence: 1.0,
	}
	payload, err := proto.Marshal(&throttleMsg)
	if err != nil {
		zap.S().Errorf("unable to marshal %T protobuf content: %err", throttleMsg, err)
		return
	}

	publish(p.client, p.throttleTopic, &payload)
}

func (p *ThrottlePart) Stop() {
	close(p.cancel)
	service.StopService("throttle", p.client, p.driveModeTopic, p.rcThrottleTopic)
}

func (p *ThrottlePart) onDriveMode(_ mqtt.Client, message mqtt.Message) {
	var msg events.DriveModeMessage
	err := proto.Unmarshal(message.Payload(), &msg)
	if err != nil {
		zap.S().Errorf("unable to unmarshal protobuf %T message: %v", msg, err)
		return
	}

	p.muDriveMode.Lock()
	defer p.muDriveMode.Unlock()
	p.driveMode = msg.GetDriveMode()
}

func (p *ThrottlePart) onRCThrottle(_ mqtt.Client, message mqtt.Message) {
	p.muDriveMode.RLock()
	defer p.muDriveMode.RUnlock()
	if p.driveMode == events.DriveMode_USER {
		// Republish same content
		payload := message.Payload()
		var throttleMsg events.ThrottleMessage
		err := proto.Unmarshal(payload, &throttleMsg)
		if err != nil {
			zap.S().Errorf("unable to unmarshall throttle msg to check throttle value: %v", err)
			return
		}
		zap.S().Debugf("publish new throttle value from rc: %v", throttleMsg.GetThrottle())
		if throttleMsg.GetThrottle() > p.maxThrottle {
			zap.S().Debugf("throttle upper that max value allowed, patch value from %v to %v", throttleMsg.GetThrottle(), p.maxThrottle)
			throttleMsg.Throttle = p.maxThrottle
			payloadPatched, err := proto.Marshal(&throttleMsg)
			if err != nil {
				zap.S().Errorf("unable to marshall throttle msg: %v", err)
				return
			}
			publish(p.client, p.throttleTopic, &payloadPatched)
			return
		}
		publish(p.client, p.throttleTopic, &payload)
	}
}

var registerCallbacks = func(p *ThrottlePart) error {
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

var publish = func(client mqtt.Client, topic string, payload *[]byte) {
	client.Publish(topic, 0, false, *payload)
}
