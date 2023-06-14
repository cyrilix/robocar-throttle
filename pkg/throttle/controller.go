package throttle

import (
	"github.com/cyrilix/robocar-base/service"
	"github.com/cyrilix/robocar-protobuf/go/events"
	"github.com/cyrilix/robocar-throttle/pkg/brake"
	"github.com/cyrilix/robocar-throttle/pkg/types"
	mqtt "github.com/eclipse/paho.mqtt.golang"
	"go.uber.org/zap"
	"google.golang.org/protobuf/proto"
	"sync"
	"time"
)

func New(client mqtt.Client, throttleTopic, driveModeTopic, rcThrottleTopic, steeringTopic, throttleFeedbackTopic,
	speedZoneTopic string,
	maxValue types.Throttle, publishPilotFrequency int, opts ...Option) *Controller {
	c := &Controller{
		client:                client,
		throttleTopic:         throttleTopic,
		driveModeTopic:        driveModeTopic,
		rcThrottleTopic:       rcThrottleTopic,
		steeringTopic:         steeringTopic,
		throttleFeedbackTopic: throttleFeedbackTopic,
		speedZoneTopic:        speedZoneTopic,
		maxThrottle:           maxValue,
		driveMode:             events.DriveMode_USER,
		publishPilotFrequency: publishPilotFrequency,
		processor:             &SteeringProcessor{minThrottle: 0.1, maxThrottle: maxValue},
		brakeCtrl:             &brake.DisabledController{},
	}
	for _, o := range opts {
		o(c)
	}
	return c
}

type Option func(c *Controller)

func WithBrakeController(bc brake.Controller) Option {
	return func(c *Controller) {
		c.brakeCtrl = bc
	}
}

func WithThrottleProcessor(p Processor) Option {
	return func(c *Controller) {
		c.processor = p
	}
}

type Controller struct {
	client        mqtt.Client
	throttleTopic string
	maxThrottle   types.Throttle
	processor     Processor

	muDriveMode sync.RWMutex
	driveMode   events.DriveMode

	muSteering sync.RWMutex
	steering   types.Steering

	brakeCtrl brake.Controller

	cancel                                                                chan interface{}
	publishPilotFrequency                                                 int
	driveModeTopic, rcThrottleTopic, steeringTopic, throttleFeedbackTopic string
	speedZoneTopic                                                        string
}

func (c *Controller) Start() error {
	if err := registerCallbacks(c); err != nil {
		zap.S().Errorf("unable to register callbacks: %v", err)
		return err
	}

	c.cancel = make(chan interface{})
	ticker := time.NewTicker(1 * time.Second / time.Duration(c.publishPilotFrequency))
	for {
		select {
		case <-ticker.C:
			c.onPublishPilotValue()
		case <-c.cancel:
			return nil
		}
	}
}

func (c *Controller) onPublishPilotValue() {
	c.muDriveMode.RLock()
	defer c.muDriveMode.RUnlock()

	if c.driveMode != events.DriveMode_PILOT {
		return
	}

	throttleFromSteering := c.processor.Process(c.readSteering())

	throttleMsg := events.ThrottleMessage{
		Throttle:   float32(c.brakeCtrl.AdjustThrottle(throttleFromSteering)),
		Confidence: 1.0,
	}
	payload, err := proto.Marshal(&throttleMsg)
	if err != nil {
		zap.S().Errorf("unable to marshal %v protobuf content: %v", throttleMsg.String(), err)
		return
	}

	publish(c.client, c.throttleTopic, payload)

}

func (c *Controller) readSteering() types.Steering {
	c.muSteering.RLock()
	defer c.muSteering.RUnlock()
	return c.steering
}

func (c *Controller) Stop() {
	close(c.cancel)
	service.StopService("throttle", c.client, c.driveModeTopic, c.rcThrottleTopic, c.steeringTopic,
		c.throttleFeedbackTopic, c.speedZoneTopic)
}

func (c *Controller) onThrottleFeedback(_ mqtt.Client, message mqtt.Message) {
	var msg events.ThrottleMessage
	err := proto.Unmarshal(message.Payload(), &msg)
	if err != nil {
		zap.S().Errorf("unable to unmarshal protobuf %T message: %v", &msg, err)
		return
	}
	c.brakeCtrl.SetRealThrottle(types.Throttle(msg.GetThrottle()))
}

func (c *Controller) onDriveMode(_ mqtt.Client, message mqtt.Message) {
	var msg events.DriveModeMessage
	err := proto.Unmarshal(message.Payload(), &msg)
	if err != nil {
		zap.S().Errorf("unable to unmarshal protobuf %T message: %v", &msg, err)
		return
	}

	c.muDriveMode.Lock()
	defer c.muDriveMode.Unlock()
	c.driveMode = msg.GetDriveMode()
}

func (c *Controller) onRCThrottle(_ mqtt.Client, message mqtt.Message) {
	c.muDriveMode.RLock()
	defer c.muDriveMode.RUnlock()
	if c.driveMode == events.DriveMode_USER {
		// Republish same content
		payload := message.Payload()
		var throttleMsg events.ThrottleMessage
		err := proto.Unmarshal(payload, &throttleMsg)
		if err != nil {
			zap.S().Errorf("unable to unmarshall throttle msg to check throttle value: %v", err)
			return
		}
		zap.S().Debugf("publish new throttle value from rc: %v", throttleMsg.GetThrottle())
		if types.Throttle(throttleMsg.GetThrottle()) > c.maxThrottle {
			zap.S().Debugf("throttle upper that max value allowed, patch value from %v to %v", throttleMsg.GetThrottle(), c.maxThrottle)
			throttleMsg.Throttle = float32(c.maxThrottle)
			payloadPatched, err := proto.Marshal(&throttleMsg)
			if err != nil {
				zap.S().Errorf("unable to marshall throttle msg: %v", err)
				return
			}
			publish(c.client, c.throttleTopic, payloadPatched)
			return
		}
		publish(c.client, c.throttleTopic, payload)
	}
}

func (c *Controller) onSteering(_ mqtt.Client, message mqtt.Message) {
	var steeringMsg events.SteeringMessage
	payload := message.Payload()
	err := proto.Unmarshal(payload, &steeringMsg)
	if err != nil {
		zap.S().Errorf("unable to unmarshal steering message, skip value: %v", err)
		return
	}
	c.muSteering.Lock()
	defer c.muSteering.Unlock()
	c.steering = types.Steering(steeringMsg.GetSteering())
}

func (c *Controller) onSpeedZone(_ mqtt.Client, message mqtt.Message) {
	var szMsg events.SpeedZoneMessage
	payload := message.Payload()
	err := proto.Unmarshal(payload, &szMsg)
	if err != nil {
		zap.S().Errorf("unable to unmarshal speedZone message, skip value: %v", err)
		return
	}
	c.processor.SetSpeedZone(szMsg.GetSpeedZone())
}

var registerCallbacks = func(p *Controller) error {
	err := service.RegisterCallback(p.client, p.driveModeTopic, p.onDriveMode)
	if err != nil {
		return err
	}

	err = service.RegisterCallback(p.client, p.rcThrottleTopic, p.onRCThrottle)
	if err != nil {
		return err
	}

	err = service.RegisterCallback(p.client, p.steeringTopic, p.onSteering)
	if err != nil {
		return err
	}
	err = service.RegisterCallback(p.client, p.throttleFeedbackTopic, p.onThrottleFeedback)
	if err != nil {
		return err
	}
	err = service.RegisterCallback(p.client, p.speedZoneTopic, p.onSpeedZone)
	if err != nil {
		return err
	}
	return nil
}

var publish = func(client mqtt.Client, topic string, payload []byte) {
	client.Publish(topic, 0, false, payload)
}
