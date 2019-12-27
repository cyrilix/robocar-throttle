package part

import (
	"github.com/cyrilix/robocar-base/mqttdevice"
	mqtt "github.com/eclipse/paho.mqtt.golang"
	"time"
)

type CommandValue struct {
	Value      float64
	Confidence float64
}
type Steering CommandValue
type Throttle CommandValue

func NewPart(pub mqttdevice.Publisher, throttleTopic string, minValue, maxValue float64) *ThrottlePart {
	return &ThrottlePart{
		pub:           pub,
		throttleTopic: throttleTopic,
		minThrottle:   minValue,
		maxThrottle:   maxValue,
	}

}

type ThrottlePart struct {
	client                   mqtt.Client
	pub                      mqttdevice.Publisher
	throttleTopic            string
	minThrottle, maxThrottle float64
	cancel                   chan interface{}
}

func (p *ThrottlePart) Start() error {
	p.cancel = make(chan interface{})
	ticker := time.NewTicker(500 * time.Millisecond)
	for {

		p.pub.Publish(p.throttleTopic, mqttdevice.NewMqttValue(Throttle{
			Value:      p.minThrottle,
			Confidence: 1.0,
		}))

		select {
		case <-ticker.C:
		case <-p.cancel:
			break
		}
	}
}

func (p *ThrottlePart) Stop() {
	close(p.cancel)
}
