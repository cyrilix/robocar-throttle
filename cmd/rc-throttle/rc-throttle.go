package main

import (
	"flag"
	"github.com/cyrilix/robocar-base/cli"
	"github.com/cyrilix/robocar-base/mqttdevice"
	"github.com/cyrilix/robocar-throttle/part"
	"log"
	"os"
)

const (
	DefaultClientId = "robocar-throttle"
	DefaultThrottleMin = 0.3
)

func main() {
	var mqttBroker, username, password, clientId string
	var throttleTopic string
	var minThrottle, maxThrottle float64

	err := cli.SetFloat64DefaultValueFromEnv(&minThrottle, "THROTTLE_MIN", DefaultThrottleMin)
	if err != nil {
		log.Printf("unable to parse min throttle value arg: %v", err)
	}
	err = cli.SetFloat64DefaultValueFromEnv(&maxThrottle, "THROTTLE_MAX", minThrottle)
	if err != nil {
		log.Printf("unable to parse max throttle value arg: %v", err)
	}

	mqttQos := cli.InitIntFlag("MQTT_QOS", 0)
	_, mqttRetain := os.LookupEnv("MQTT_RETAIN")

	cli.InitMqttFlags(DefaultClientId, &mqttBroker, &username, &password, &clientId, &mqttQos, &mqttRetain)

	flag.StringVar(&throttleTopic, "mqtt-topic-throttle", os.Getenv("MQTT_TOPIC_THROTTLE"), "Mqtt topic to publish throttle result, use MQTT_TOPIC_THROTTLE if args not set")
	flag.Float64Var(&minThrottle, "throttle-min", minThrottle, "Minimum throttle value, use THROTTLE_MIN if args not set")
	flag.Float64Var(&maxThrottle, "throttle-max", maxThrottle, "Minimum throttle value, use THROTTLE_MAX if args not set")

	flag.Parse()
	if len(os.Args) <= 1 {
		flag.PrintDefaults()
		os.Exit(1)
	}

	client, err := cli.Connect(mqttBroker, username, password, clientId)
	if err != nil {
		log.Fatalf("unable to connect to mqtt bus: %v", err)
	}
	defer client.Disconnect(50)

	pub := mqttdevice.NewPahoMqttPubSub(mqttBroker, username, password,clientId,mqttQos, mqttRetain)

	p := part.NewPart(pub, throttleTopic, minThrottle, maxThrottle)
	defer p.Stop()

	cli.HandleExit(p)

	err = p.Start()
	if err != nil {
		log.Fatalf("unable to start service: %v", err)
	}
}
