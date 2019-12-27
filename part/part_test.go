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
		driveMode        types.DriveMode
		rcThrottle       float64
		expectedThrottle float64
	}{
		{types.DriveModeUser, 0.3, 0.3},
		{types.DriveModePilot, 0.5, minValue},
		{types.DriveModePilot, 0.4, minValue},
		{types.DriveModeUser, 0.5, 0.5},
		{types.DriveModeUser, 0.4, 0.4},
		{types.DriveModeUser, 0.6, 0.6},
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

			if throttle.Value != c.expectedThrottle {
				t.Errorf("bad throttle value for mode %v: %v, wants %v", c.driveMode, throttle.Value, c.expectedThrottle)
			}
			if throttle.Confidence != 1. {
				t.Errorf("bad throtlle confidence: %v, wants %v", throttle.Confidence, 1.)
			}

			time.Sleep(1 * time.Millisecond)
		}
	}
}
