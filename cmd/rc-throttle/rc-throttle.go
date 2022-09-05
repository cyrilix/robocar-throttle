package main

import (
	"flag"
	"github.com/cyrilix/robocar-base/cli"
	"github.com/cyrilix/robocar-throttle/pkg/brake"
	"github.com/cyrilix/robocar-throttle/pkg/throttle"
	"github.com/cyrilix/robocar-throttle/pkg/types"
	"go.uber.org/zap"
	"log"
	"os"
)

const (
	DefaultClientId    = "robocar-throttle"
	DefaultThrottleMin = 0.3
)

func main() {
	var mqttBroker, username, password, clientId string
	var throttleTopic, driveModeTopic, rcThrottleTopic, steeringTopic, throttleFeedbackTopic string
	var minThrottle, maxThrottle float64
	var publishPilotFrequency int
	var brakeConfig string
	var enableBrake bool

	err := cli.SetFloat64DefaultValueFromEnv(&minThrottle, "THROTTLE_MIN", DefaultThrottleMin)
	if err != nil {
		zap.S().Errorf("unable to parse min throttle value arg: %v", err)
	}
	err = cli.SetFloat64DefaultValueFromEnv(&maxThrottle, "THROTTLE_MAX", minThrottle)
	if err != nil {
		zap.S().Errorf("unable to parse max throttle value arg: %v", err)
	}

	mqttQos := cli.InitIntFlag("MQTT_QOS", 0)
	_, mqttRetain := os.LookupEnv("MQTT_RETAIN")

	cli.InitMqttFlags(DefaultClientId, &mqttBroker, &username, &password, &clientId, &mqttQos, &mqttRetain)

	flag.StringVar(&throttleTopic, "mqtt-topic-throttle", os.Getenv("MQTT_TOPIC_THROTTLE"), "Mqtt topic to publish throttle result, use MQTT_TOPIC_THROTTLE if args not set")
	flag.StringVar(&driveModeTopic, "mqtt-topic-drive-mode", os.Getenv("MQTT_TOPIC_DRIVE_MODE"), "Mqtt topic that contains DriveMode value, use MQTT_TOPIC_DRIVE_MODE if args not set")
	flag.StringVar(&rcThrottleTopic, "mqtt-topic-rc-throttle", os.Getenv("MQTT_TOPIC_RC_THROTTLE"), "Mqtt topic that contains RC Throttle value, use MQTT_TOPIC_RC_THROTTLE if args not set")
	flag.StringVar(&steeringTopic, "mqtt-topic-steering", os.Getenv("MQTT_TOPIC_STEERING"), "Mqtt topic that contains steering value, use MQTT_TOPIC_STEERING if args not set")
	flag.StringVar(&throttleFeedbackTopic, "mqtt-topic-throttle-feedback", os.Getenv("MQTT_TOPIC_THROTTLE_FEEDBACK"), "Mqtt topic where to publish throttle feedback, use MQTT_TOPIC_THROTTLE_FEEDBACK if args not set")

	flag.Float64Var(&minThrottle, "throttle-min", minThrottle, "Minimum throttle value, use THROTTLE_MIN if args not set")
	flag.Float64Var(&maxThrottle, "throttle-max", maxThrottle, "Minimum throttle value, use THROTTLE_MAX if args not set")
	flag.IntVar(&publishPilotFrequency, "update-pwm-frequency", 2, "Number of throttle event to publish when pilot mode is enabled")

	flag.BoolVar(&enableBrake, "enable-brake-feature", false, "Enable brake to slow car on throttle changes")
	flag.StringVar(&brakeConfig, "brake-configuration", "", "Json file to use to configure brake adaptation when --enable-brake is `true`")
	logLevel := zap.LevelFlag("log", zap.InfoLevel, "log level")

	flag.Parse()
	if len(os.Args) <= 1 {
		flag.PrintDefaults()
		os.Exit(1)
	}

	config := zap.NewDevelopmentConfig()
	config.Level = zap.NewAtomicLevelAt(*logLevel)
	lgr, err := config.Build()
	if err != nil {
		log.Fatalf("unable to init logger: %v", err)
	}
	defer func() {
		if err := lgr.Sync(); err != nil {
			log.Printf("unable to Sync logger: %v\n", err)
		}
	}()
	zap.ReplaceGlobals(lgr)

	zap.S().Infof("Topic throttle          : %s", throttleTopic)
	zap.S().Infof("Topic rc-throttle       : %s", rcThrottleTopic)
	zap.S().Infof("Topic throttle feedback : %s", throttleFeedbackTopic)
	zap.S().Infof("Topic steering          : %s", steeringTopic)
	zap.S().Infof("Topic drive mode        : %s", driveModeTopic)
	zap.S().Infof("Min throttle            : %v", minThrottle)
	zap.S().Infof("Max throttle            : %v", maxThrottle)
	zap.S().Infof("Publish frequency       : %vHz", publishPilotFrequency)
	zap.S().Infof("Brake enabled           : %v", enableBrake)

	client, err := cli.Connect(mqttBroker, username, password, clientId)
	if err != nil {
		zap.S().Fatalf("unable to connect to mqtt bus: %v", err)
	}
	defer client.Disconnect(50)

	var brakeCtrl brake.Controller
	if enableBrake {
		brakeCtrl = brake.NewCustomControllerWithJsonConfig(brakeConfig)
	} else {
		brakeCtrl = &brake.DisabledController{}
	}
	p := throttle.New(client, throttleTopic, driveModeTopic, rcThrottleTopic, steeringTopic, throttleFeedbackTopic,
		types.Throttle(minThrottle), types.Throttle(maxThrottle), 2, throttle.WithBrakeController(brakeCtrl))
	defer p.Stop()

	cli.HandleExit(p)

	err = p.Start()
	if err != nil {
		zap.S().Fatalf("unable to start service: %v", err)
	}
}
