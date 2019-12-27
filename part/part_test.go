package part

import (
	"encoding/json"
	"github.com/cyrilix/robocar-base/testtools"
	"testing"
	"time"
)

func TestDefaultThrottle(t *testing.T){
	throttleTopic := "topic/throttle"
	minValue := 0.56

	pub := testtools.NewFakePublisher()
	p := NewPart(pub, throttleTopic, minValue, 1.)

	go p.Start()
	defer p.Stop()

	time.Sleep(1 * time.Millisecond)

	mqttValue := pub.PublishedEvent(throttleTopic)
	var throttle Throttle
	err := json.Unmarshal(mqttValue, &throttle)
	if err != nil {
		t.Errorf("unable to unmarshall response: %v", err)
		t.Fail()
	}

	if throttle.Value != minValue {
		t.Errorf("bad throttle value: %v, wants %v", throttle.Value, minValue)
	}
	if throttle.Confidence != 1. {
		t.Errorf("bad throtlle confidence: %v, wants %v", throttle.Confidence, 1.)
	}
}
