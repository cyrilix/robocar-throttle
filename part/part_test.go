package part

import (
	"encoding/json"
	"github.com/cyrilix/robocar-base/mqttdevice"
	"github.com/cyrilix/robocar-base/testtools"
	"github.com/cyrilix/robocar-base/types"
	"testing"
	"time"
)

func TestDefaultThrottle(t *testing.T) {
	oldRegister := registerCallbacks
	defer func(){
		registerCallbacks = oldRegister
	}()
	registerCallbacks = func(p *ThrottlePart) error {
		return nil
	}

	throttleTopic := "topic/throttle"
	driveModeTopic := "topic/driveMode"
	rcThrottleTopic := "topic/rcThrottle"

	minValue := 0.56

	pub := testtools.NewFakePublisher()
	p := NewPart(nil, pub, throttleTopic, driveModeTopic, rcThrottleTopic, minValue, 1., 200)

	cases := []struct {
		driveMode        string
		rcThrottle       types.Throttle
		expectedThrottle types.Throttle
	}{
		{types.ToString(types.DriveModeUser), types.Throttle{Value: 0.3, Confidence: 1.0},types.Throttle{Value: 0.3, Confidence: 1.0}},
		{types.ToString(types.DriveModePilot), types.Throttle{Value: 0.5, Confidence: 1.0}, types.Throttle{Value: minValue, Confidence: 1.0}},
		{types.ToString(types.DriveModePilot), types.Throttle{Value: 0.4, Confidence: 1.0}, types.Throttle{Value: minValue, Confidence: 1.0}},
		{types.ToString(types.DriveModeUser), types.Throttle{Value: 0.5, Confidence: 1.0}, types.Throttle{Value: 0.5, Confidence: 1.0}},
		{types.ToString(types.DriveModeUser), types.Throttle{Value: 0.4, Confidence: 1.0}, types.Throttle{Value: 0.4, Confidence: 1.0}},
		{types.ToString(types.DriveModeUser), types.Throttle{Value: 0.6, Confidence: 1.0}, types.Throttle{Value: 0.6, Confidence: 1.0}},
	}

	go p.Start()
	defer func(){close(p.cancel)}()

	for _, c := range cases {

		p.onDriveMode(nil, testtools.NewFakeMessage(driveModeTopic, mqttdevice.NewMqttValue(c.driveMode)))
		p.onRCThrottle(nil, testtools.NewFakeMessage(rcThrottleTopic, mqttdevice.NewMqttValue(c.rcThrottle)))

		time.Sleep(10 * time.Millisecond)

		for i := 3; i >= 0; i-- {

			mqttValue := pub.PublishedEvent(throttleTopic)
			var throttle types.Throttle
			err := json.Unmarshal(mqttValue, &throttle)
			if err != nil {
				t.Errorf("unable to unmarshall response: %v", err)
				t.Fail()
			}

			if throttle != c.expectedThrottle {
				t.Errorf("bad throttle value for mode %v: %v, wants %v", c.driveMode, throttle.Value, c.expectedThrottle)
			}
			if throttle.Confidence != 1. {
				t.Errorf("bad throtlle confidence: %v, wants %v", throttle.Confidence, 1.)
			}

			time.Sleep(1 * time.Millisecond)
		}
	}
}
