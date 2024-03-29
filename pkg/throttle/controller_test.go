package throttle

import (
	"github.com/cyrilix/robocar-base/testtools"
	"github.com/cyrilix/robocar-protobuf/go/events"
	"github.com/cyrilix/robocar-throttle/pkg/brake"
	"github.com/cyrilix/robocar-throttle/pkg/types"
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
	publish = func(client mqtt.Client, topic string, payload []byte) {
		muEventsPublished.Lock()
		defer muEventsPublished.Unlock()
		eventsPublished[topic] = payload
	}

	throttleTopic := "topic/throttle"
	driveModeTopic := "topic/driveMode"
	rcThrottleTopic := "topic/rcThrottle"
	steeringTopic := "topic/rcThrottle"
	throttleFeedbackTopic := "topic/feedback/throttle"
	maxThrottleCtrlTopic := "topic/max/throttle"
	speedZoneTopic := "topic/speedZone"

	p := New(nil, throttleTopic, driveModeTopic, rcThrottleTopic, steeringTopic, throttleFeedbackTopic,
		maxThrottleCtrlTopic, speedZoneTopic, 1., 200)

	cases := []*struct {
		name             string
		maxThrottle      types.Throttle
		driveMode        events.DriveModeMessage
		rcThrottle       events.ThrottleMessage
		expectedThrottle events.ThrottleMessage
	}{
		{"test1", 1., events.DriveModeMessage{DriveMode: events.DriveMode_USER}, events.ThrottleMessage{Throttle: 0.3, Confidence: 1.0}, events.ThrottleMessage{Throttle: 0.3, Confidence: 1.0}},
		{"test2", 1., events.DriveModeMessage{DriveMode: events.DriveMode_PILOT}, events.ThrottleMessage{Throttle: 0.5, Confidence: 1.0}, events.ThrottleMessage{Throttle: 1.0, Confidence: 1.0}},
		{"test3", 1., events.DriveModeMessage{DriveMode: events.DriveMode_PILOT}, events.ThrottleMessage{Throttle: 0.4, Confidence: 1.0}, events.ThrottleMessage{Throttle: 1.0, Confidence: 1.0}},
		{"test4", 1., events.DriveModeMessage{DriveMode: events.DriveMode_USER}, events.ThrottleMessage{Throttle: 0.5, Confidence: 1.0}, events.ThrottleMessage{Throttle: 0.5, Confidence: 1.0}},
		{"test5", 1., events.DriveModeMessage{DriveMode: events.DriveMode_USER}, events.ThrottleMessage{Throttle: 0.4, Confidence: 1.0}, events.ThrottleMessage{Throttle: 0.4, Confidence: 1.0}},
		{"test6", 1., events.DriveModeMessage{DriveMode: events.DriveMode_USER}, events.ThrottleMessage{Throttle: 0.6, Confidence: 1.0}, events.ThrottleMessage{Throttle: 0.6, Confidence: 1.0}},
		{"test7", 1., events.DriveModeMessage{DriveMode: events.DriveMode_COPILOT}, events.ThrottleMessage{Throttle: 0.3, Confidence: 1.0}, events.ThrottleMessage{Throttle: 0.3, Confidence: 1.0}},
		{"test8", 1., events.DriveModeMessage{DriveMode: events.DriveMode_COPILOT}, events.ThrottleMessage{Throttle: 0.5, Confidence: 1.0}, events.ThrottleMessage{Throttle: 0.5, Confidence: 1.0}},
		{"test9", 1., events.DriveModeMessage{DriveMode: events.DriveMode_COPILOT}, events.ThrottleMessage{Throttle: 0.4, Confidence: 1.0}, events.ThrottleMessage{Throttle: 0.4, Confidence: 1.0}},
		{"test10", 1., events.DriveModeMessage{DriveMode: events.DriveMode_COPILOT}, events.ThrottleMessage{Throttle: 0.6, Confidence: 1.0}, events.ThrottleMessage{Throttle: 0.6, Confidence: 1.0}},
		{"rescale throttle on user mode", 0.4, events.DriveModeMessage{DriveMode: events.DriveMode_USER}, events.ThrottleMessage{Throttle: 0.6, Confidence: 1.0}, events.ThrottleMessage{Throttle: 0.24000001, Confidence: 1.0}},
		{"rescale throttle on copilot mode", 0.4, events.DriveModeMessage{DriveMode: events.DriveMode_COPILOT}, events.ThrottleMessage{Throttle: 0.6, Confidence: 1.0}, events.ThrottleMessage{Throttle: 0.24000001, Confidence: 1.0}},
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
					t.Errorf("bad msg value for mode %v: %v, wants %v", c.driveMode.String(), msg.GetThrottle(), c.expectedThrottle.GetThrottle())
				}
				if msg.GetConfidence() != 1. {
					t.Errorf("bad throtlle confidence: %v, wants %v", msg.GetConfidence(), 1.)
				}

				time.Sleep(1 * time.Millisecond)
			}
		})
	}
}
func TestController_Start(t *testing.T) {
	oldRegister := registerCallbacks
	oldPublish := publish
	defer func() {
		registerCallbacks = oldRegister
		publish = oldPublish
	}()
	registerCallbacks = func(p *Controller) error {
		return nil
	}
	publishPilotFrequency := 10
	waitPublish := sync.WaitGroup{}
	var muEventsPublished sync.Mutex
	eventsPublished := make(map[string][]byte)
	publish = func(client mqtt.Client, topic string, payload []byte) {
		muEventsPublished.Lock()
		defer muEventsPublished.Unlock()
		eventsPublished[topic] = payload
		waitPublish.Done()
	}

	throttleTopic := "topic/throttle"
	steeringTopic := "topic/steering"
	driveModeTopic := "topic/driveMode"
	rcThrottleTopic := "topic/rcThrottle"
	throttleFeedbackTopic := "topic/feedback/throttle"
	maxThrottleCtrlTopic := "topic/max/throttle"
	speedZoneTopic := "topic/speedZone"

	type fields struct {
		max                   types.Throttle
		min                   types.Throttle
		driveMode             events.DriveMode
		publishPilotFrequency int
		brakeCtl              brake.Controller
	}
	type msgEvents struct {
		driveMode        *events.DriveModeMessage
		steering         *events.SteeringMessage
		rcThrottle       *events.ThrottleMessage
		throttleFeedback *events.ThrottleMessage
		maxThrottleCtrl  *events.ThrottleMessage
	}

	tests := []struct {
		name      string
		fields    fields
		msgEvents msgEvents
		want      *events.ThrottleMessage
		wantErr   bool
	}{
		{
			name: "On user drive mode, throttle from rc",
			fields: fields{
				max:                   0.8,
				min:                   0.3,
				driveMode:             events.DriveMode_USER,
				publishPilotFrequency: publishPilotFrequency,
				brakeCtl:              &brake.DisabledController{},
			},
			msgEvents: msgEvents{
				driveMode:        &events.DriveModeMessage{DriveMode: events.DriveMode_USER},
				steering:         &events.SteeringMessage{Steering: 0.0, Confidence: 1.0},
				rcThrottle:       &events.ThrottleMessage{Throttle: 0.5, Confidence: 1.0},
				throttleFeedback: &events.ThrottleMessage{Throttle: 0.4, Confidence: 1.0},
				maxThrottleCtrl:  &events.ThrottleMessage{Throttle: 0.8, Confidence: 1.0},
			},
			want: &events.ThrottleMessage{Throttle: 0.4, Confidence: 1.0},
		},
		{
			name: "On user drive mode, rescale throttle with max allowed value",
			fields: fields{
				driveMode:             events.DriveMode_USER,
				max:                   0.8,
				min:                   0.3,
				publishPilotFrequency: publishPilotFrequency,
				brakeCtl:              &brake.DisabledController{},
			},
			msgEvents: msgEvents{
				driveMode:        &events.DriveModeMessage{DriveMode: events.DriveMode_USER},
				steering:         &events.SteeringMessage{Steering: 0.0, Confidence: 1.0},
				rcThrottle:       &events.ThrottleMessage{Throttle: 0.9, Confidence: 1.0},
				throttleFeedback: &events.ThrottleMessage{Throttle: 0.8, Confidence: 1.0},
				maxThrottleCtrl:  &events.ThrottleMessage{Throttle: 0.8, Confidence: 1.0},
			},
			want: &events.ThrottleMessage{Throttle: 0.71999997, Confidence: 1.0},
		},
		{
			name: "On user drive mode, throttle can be < to  min allowed value",
			fields: fields{
				driveMode:             events.DriveMode_USER,
				max:                   0.8,
				min:                   0.3,
				publishPilotFrequency: publishPilotFrequency,
				brakeCtl:              &brake.DisabledController{},
			},
			msgEvents: msgEvents{
				driveMode:        &events.DriveModeMessage{DriveMode: events.DriveMode_USER},
				steering:         &events.SteeringMessage{Steering: 0.0, Confidence: 1.0},
				rcThrottle:       &events.ThrottleMessage{Throttle: 0.1, Confidence: 1.0},
				throttleFeedback: &events.ThrottleMessage{Throttle: 0.4, Confidence: 1.0},
				maxThrottleCtrl:  &events.ThrottleMessage{Throttle: 0.8, Confidence: 1.0},
			},
			want: &events.ThrottleMessage{Throttle: 0.080000006, Confidence: 1.0},
		},
		{
			name: "On user drive mode, brake doesn't rescale throttle",
			fields: fields{
				driveMode:             events.DriveMode_USER,
				max:                   0.8,
				min:                   0.3,
				publishPilotFrequency: publishPilotFrequency,
				brakeCtl:              &brake.DisabledController{},
			},
			msgEvents: msgEvents{
				driveMode:        &events.DriveModeMessage{DriveMode: events.DriveMode_USER},
				steering:         &events.SteeringMessage{Steering: 0.0, Confidence: 1.0},
				rcThrottle:       &events.ThrottleMessage{Throttle: -0.8, Confidence: 1.0},
				throttleFeedback: &events.ThrottleMessage{Throttle: 0.4, Confidence: 1.0},
				maxThrottleCtrl:  &events.ThrottleMessage{Throttle: 0.8, Confidence: 1.0},
			},
			want: &events.ThrottleMessage{Throttle: -0.8, Confidence: 1.0},
		},
		{
			name: "On copilot drive mode, throttle from rc",
			fields: fields{
				driveMode:             events.DriveMode_COPILOT,
				max:                   0.8,
				min:                   0.3,
				publishPilotFrequency: publishPilotFrequency,
				brakeCtl:              &brake.DisabledController{},
			},
			msgEvents: msgEvents{
				driveMode:        &events.DriveModeMessage{DriveMode: events.DriveMode_COPILOT},
				steering:         &events.SteeringMessage{Steering: 0.0, Confidence: 1.0},
				rcThrottle:       &events.ThrottleMessage{Throttle: 0.5, Confidence: 1.0},
				throttleFeedback: &events.ThrottleMessage{Throttle: 0.4, Confidence: 1.0},
				maxThrottleCtrl:  &events.ThrottleMessage{Throttle: 0.8, Confidence: 1.0},
			},
			want: &events.ThrottleMessage{Throttle: 0.4, Confidence: 1.0},
		},
		{
			name: "On copilot drive mode, rescale throttle with max allowed value",
			fields: fields{
				driveMode:             events.DriveMode_COPILOT,
				max:                   0.8,
				min:                   0.3,
				publishPilotFrequency: publishPilotFrequency,
				brakeCtl:              &brake.DisabledController{},
			},
			msgEvents: msgEvents{
				driveMode:        &events.DriveModeMessage{DriveMode: events.DriveMode_COPILOT},
				steering:         &events.SteeringMessage{Steering: 0.0, Confidence: 1.0},
				rcThrottle:       &events.ThrottleMessage{Throttle: 0.9, Confidence: 1.0},
				throttleFeedback: &events.ThrottleMessage{Throttle: 0.8, Confidence: 1.0},
				maxThrottleCtrl:  &events.ThrottleMessage{Throttle: 0.8, Confidence: 1.0},
			},
			want: &events.ThrottleMessage{Throttle: 0.71999997, Confidence: 1.0},
		},
		{
			name: "On copilot drive mode, limit throttle to max allowed value",
			fields: fields{
				driveMode:             events.DriveMode_COPILOT,
				max:                   0.8,
				min:                   0.3,
				publishPilotFrequency: publishPilotFrequency,
				brakeCtl:              &brake.DisabledController{},
			},
			msgEvents: msgEvents{
				driveMode:        &events.DriveModeMessage{DriveMode: events.DriveMode_COPILOT},
				steering:         &events.SteeringMessage{Steering: 0.0, Confidence: 1.0},
				rcThrottle:       &events.ThrottleMessage{Throttle: 0.9, Confidence: 1.0},
				throttleFeedback: &events.ThrottleMessage{Throttle: 0.8, Confidence: 1.0},
				maxThrottleCtrl:  &events.ThrottleMessage{Throttle: 0.8, Confidence: 1.0},
			},
			want: &events.ThrottleMessage{Throttle: 0.71999997, Confidence: 1.0},
		},
		{
			name: "On copilot drive mode, throttle can be < to  min allowed value",
			fields: fields{
				driveMode:             events.DriveMode_COPILOT,
				max:                   0.8,
				min:                   0.3,
				publishPilotFrequency: publishPilotFrequency,
				brakeCtl:              &brake.DisabledController{},
			},
			msgEvents: msgEvents{
				driveMode:        &events.DriveModeMessage{DriveMode: events.DriveMode_COPILOT},
				steering:         &events.SteeringMessage{Steering: 0.0, Confidence: 1.0},
				rcThrottle:       &events.ThrottleMessage{Throttle: 0.1, Confidence: 1.0},
				throttleFeedback: &events.ThrottleMessage{Throttle: 0.4, Confidence: 1.0},
				maxThrottleCtrl:  &events.ThrottleMessage{Throttle: 0.8, Confidence: 1.0},
			},
			want: &events.ThrottleMessage{Throttle: 0.080000006, Confidence: 1.0},
		},
		{
			name: "On user drive mode, brake doesn't rescale throttle",
			fields: fields{
				driveMode:             events.DriveMode_COPILOT,
				max:                   0.8,
				min:                   0.3,
				publishPilotFrequency: publishPilotFrequency,
				brakeCtl:              &brake.DisabledController{},
			},
			msgEvents: msgEvents{
				driveMode:        &events.DriveModeMessage{DriveMode: events.DriveMode_COPILOT},
				steering:         &events.SteeringMessage{Steering: 0.0, Confidence: 1.0},
				rcThrottle:       &events.ThrottleMessage{Throttle: -0.8, Confidence: 1.0},
				throttleFeedback: &events.ThrottleMessage{Throttle: 0.4, Confidence: 1.0},
				maxThrottleCtrl:  &events.ThrottleMessage{Throttle: 0.8, Confidence: 1.0},
			},
			want: &events.ThrottleMessage{Throttle: -0.8, Confidence: 1.0},
		},
		{
			name: "On pilot drive mode and straight steering, use max throttle allowed",
			fields: fields{
				driveMode:             events.DriveMode_PILOT,
				max:                   0.8,
				min:                   0.3,
				publishPilotFrequency: publishPilotFrequency,
				brakeCtl:              &brake.DisabledController{},
			},
			msgEvents: msgEvents{
				driveMode:        &events.DriveModeMessage{DriveMode: events.DriveMode_PILOT},
				steering:         &events.SteeringMessage{Steering: 0.0, Confidence: 1.0},
				rcThrottle:       &events.ThrottleMessage{Throttle: 0.5, Confidence: 1.0},
				throttleFeedback: &events.ThrottleMessage{Throttle: 0.4, Confidence: 1.0},
				maxThrottleCtrl:  &events.ThrottleMessage{Throttle: 0.8, Confidence: 1.0},
			},
			want: &events.ThrottleMessage{Throttle: 0.8, Confidence: 1.0},
		},
		{
			name: "On pilot drive mode and on left steering, use min throttle allowed",
			fields: fields{
				driveMode:             events.DriveMode_PILOT,
				max:                   0.8,
				min:                   0.3,
				publishPilotFrequency: publishPilotFrequency,
				brakeCtl:              &brake.DisabledController{},
			},
			msgEvents: msgEvents{
				driveMode:        &events.DriveModeMessage{DriveMode: events.DriveMode_PILOT},
				steering:         &events.SteeringMessage{Steering: -1.0, Confidence: 1.0},
				rcThrottle:       &events.ThrottleMessage{Throttle: 0.3, Confidence: 1.0},
				throttleFeedback: &events.ThrottleMessage{Throttle: 0.4, Confidence: 1.0},
				maxThrottleCtrl:  &events.ThrottleMessage{Throttle: 0.8, Confidence: 1.0},
			},
			want: &events.ThrottleMessage{Throttle: 0.3, Confidence: 1.0},
		},
		{
			name: "On pilot drive mode and on right steering, use min throttle allowed",
			fields: fields{
				driveMode:             events.DriveMode_PILOT,
				max:                   0.8,
				min:                   0.3,
				publishPilotFrequency: publishPilotFrequency,
				brakeCtl:              &brake.DisabledController{},
			},
			msgEvents: msgEvents{
				driveMode:        &events.DriveModeMessage{DriveMode: events.DriveMode_PILOT},
				steering:         &events.SteeringMessage{Steering: 1.0, Confidence: 1.0},
				rcThrottle:       &events.ThrottleMessage{Throttle: 0.3, Confidence: 1.0},
				throttleFeedback: &events.ThrottleMessage{Throttle: 0.4, Confidence: 1.0},
				maxThrottleCtrl:  &events.ThrottleMessage{Throttle: 0.8, Confidence: 1.0},
			},
			want: &events.ThrottleMessage{Throttle: 0.3, Confidence: 1.0},
		},
		{
			name: "On pilot drive mode, should brake on brutal change",
			fields: fields{
				driveMode:             events.DriveMode_PILOT,
				max:                   1.0,
				min:                   0.3,
				publishPilotFrequency: publishPilotFrequency,
				brakeCtl:              brake.NewCustomController(),
			},
			msgEvents: msgEvents{
				driveMode:        &events.DriveModeMessage{DriveMode: events.DriveMode_PILOT},
				steering:         &events.SteeringMessage{Steering: -1.0, Confidence: 1.0},
				rcThrottle:       &events.ThrottleMessage{Throttle: 0.3, Confidence: 1.0},
				throttleFeedback: &events.ThrottleMessage{Throttle: 1.0, Confidence: 1.0},
				maxThrottleCtrl:  &events.ThrottleMessage{Throttle: 0.8, Confidence: 1.0},
			},
			want: &events.ThrottleMessage{Throttle: -1.0, Confidence: 1.0},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := New(nil,
				throttleTopic, driveModeTopic, rcThrottleTopic, steeringTopic, throttleFeedbackTopic,
				maxThrottleCtrlTopic, speedZoneTopic, tt.fields.max,
				tt.fields.publishPilotFrequency,
				WithThrottleProcessor(&SteeringProcessor{
					minThrottle: tt.fields.min,
					maxThrottle: tt.fields.max,
				}),
				WithBrakeController(tt.fields.brakeCtl),
			)

			go c.Start()
			defer func() { close(c.cancel) }()
			time.Sleep(1 * time.Millisecond)

			// Publish events and wait generation of new steering message
			waitPublish.Add(1)
			c.onDriveMode(nil, testtools.NewFakeMessageFromProtobuf(driveModeTopic, tt.msgEvents.driveMode))
			c.onRCThrottle(nil, testtools.NewFakeMessageFromProtobuf(rcThrottleTopic, tt.msgEvents.rcThrottle))
			c.onSteering(nil, testtools.NewFakeMessageFromProtobuf(steeringTopic, tt.msgEvents.steering))
			c.onThrottleFeedback(nil, testtools.NewFakeMessageFromProtobuf(throttleFeedbackTopic, tt.msgEvents.throttleFeedback))
			c.onMaxThrottleCtrl(nil, testtools.NewFakeMessageFromProtobuf(maxThrottleCtrlTopic, tt.msgEvents.maxThrottleCtrl))
			waitPublish.Wait()

			var msg events.ThrottleMessage
			muEventsPublished.Lock()
			err := proto.Unmarshal(eventsPublished[throttleTopic], &msg)
			if err != nil {
				t.Errorf("unable to unmarshall response: %v", err)
				t.Fail()
			}
			muEventsPublished.Unlock()

			if msg.GetThrottle() != tt.want.GetThrottle() {
				t.Errorf("bad msg value for mode %v: %v, wants %v", c.driveMode.String(), msg.GetThrottle(), tt.want.GetThrottle())
			}

		})
	}
}
