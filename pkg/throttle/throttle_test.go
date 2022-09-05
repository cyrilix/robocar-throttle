package throttle

import (
	"github.com/cyrilix/robocar-base/testtools"
	"github.com/cyrilix/robocar-protobuf/go/events"
	mqtt "github.com/eclipse/paho.mqtt.golang"
	"google.golang.org/protobuf/proto"
	"sync"
	"testing"
	"time"
)

func TestDefaultThrottle(t *testing.T) {
	oldRegister := registerCallbacks
	oldPublish := publish
	defer func() {
		registerCallbacks = oldRegister
		publish = oldPublish
	}()
	registerCallbacks = func(p *Controller) error {
		return nil
	}

	var muEventsPublished sync.Mutex
	eventsPublished := make(map[string][]byte)
	publish = func(client mqtt.Client, topic string, payload *[]byte) {
		muEventsPublished.Lock()
		defer muEventsPublished.Unlock()
		eventsPublished[topic] = *payload
	}

	throttleTopic := "topic/throttle"
	driveModeTopic := "topic/driveMode"
	rcThrottleTopic := "topic/rcThrottle"

	minValue := float32(0.56)

	p := New(nil, throttleTopic, driveModeTopic, rcThrottleTopic, minValue, 1., 200)

	cases := []struct {
		name             string
		maxThrottle      float32
		driveMode        events.DriveModeMessage
		rcThrottle       events.ThrottleMessage
		expectedThrottle events.ThrottleMessage
	}{
		{"test1", 1., events.DriveModeMessage{DriveMode: events.DriveMode_USER}, events.ThrottleMessage{Throttle: 0.3, Confidence: 1.0}, events.ThrottleMessage{Throttle: 0.3, Confidence: 1.0}},
		{"test2", 1., events.DriveModeMessage{DriveMode: events.DriveMode_PILOT}, events.ThrottleMessage{Throttle: 0.5, Confidence: 1.0}, events.ThrottleMessage{Throttle: minValue, Confidence: 1.0}},
		{"test3", 1., events.DriveModeMessage{DriveMode: events.DriveMode_PILOT}, events.ThrottleMessage{Throttle: 0.4, Confidence: 1.0}, events.ThrottleMessage{Throttle: minValue, Confidence: 1.0}},
		{"test4", 1., events.DriveModeMessage{DriveMode: events.DriveMode_USER}, events.ThrottleMessage{Throttle: 0.5, Confidence: 1.0}, events.ThrottleMessage{Throttle: 0.5, Confidence: 1.0}},
		{"test5", 1., events.DriveModeMessage{DriveMode: events.DriveMode_USER}, events.ThrottleMessage{Throttle: 0.4, Confidence: 1.0}, events.ThrottleMessage{Throttle: 0.4, Confidence: 1.0}},
		{"test6", 1., events.DriveModeMessage{DriveMode: events.DriveMode_USER}, events.ThrottleMessage{Throttle: 0.6, Confidence: 1.0}, events.ThrottleMessage{Throttle: 0.6, Confidence: 1.0}},
		{"limit max throttle on user mode", 0.4, events.DriveModeMessage{DriveMode: events.DriveMode_USER}, events.ThrottleMessage{Throttle: 0.6, Confidence: 1.0}, events.ThrottleMessage{Throttle: 0.4, Confidence: 1.0}},
	}

	go p.Start()
	defer func() { close(p.cancel) }()

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			p.maxThrottle = c.maxThrottle
			p.onDriveMode(nil, testtools.NewFakeMessageFromProtobuf(driveModeTopic, &c.driveMode))
			p.onRCThrottle(nil, testtools.NewFakeMessageFromProtobuf(rcThrottleTopic, &c.rcThrottle))

			time.Sleep(10 * time.Millisecond)

			for i := 3; i >= 0; i-- {

				var msg events.ThrottleMessage
				muEventsPublished.Lock()
				err := proto.Unmarshal(eventsPublished[throttleTopic], &msg)
				if err != nil {
					t.Errorf("unable to unmarshall response: %v", err)
					t.Fail()
				}
				muEventsPublished.Unlock()

				if msg.GetThrottle() != c.expectedThrottle.GetThrottle() {
					t.Errorf("bad msg value for mode %v: %v, wants %v", c.driveMode, msg.GetThrottle(), c.expectedThrottle.GetThrottle())
				}
				if msg.GetConfidence() != 1. {
					t.Errorf("bad throtlle confidence: %v, wants %v", msg.GetConfidence(), 1.)
				}

				time.Sleep(1 * time.Millisecond)
			}
		})
	}
}
