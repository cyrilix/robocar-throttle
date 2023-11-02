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
	DefaultThrottleMin = 0.1
)

func main() {
	var mqttBroker, username, password, clientId string
	var throttleTopic, driveModeTopic, rcThrottleTopic, steeringTopic, throttleFeedbackTopic, maxThrottleCtrlTopic,
		speedZoneTopic string
	var minThrottle, maxThrottle float64
	var publishPilotFrequency int
	var brakeConfig string
	var enableBrake bool
	var enableSpeedZone bool
	var enableCustomSteeringProcessor bool
	var configFileSteeringProcessor string
	var slowZoneThrottle, normalZoneThrottle, fastZoneThrottle float64
	var moderateSteering, fullSteering float64

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
	flag.StringVar(&maxThrottleCtrlTopic, "mqtt-topic-max-throttle-ctrl", os.Getenv("MQTT_TOPIC_MAX_THROTTLE_CTRL"), "Mqtt topic where to publish max throttle value allowed, use MQTT_TOPIC_MAX_THROTTLE_CTRL if args not set")
	flag.StringVar(&steeringTopic, "mqtt-topic-steering", os.Getenv("MQTT_TOPIC_STEERING"), "Mqtt topic that contains steering value, use MQTT_TOPIC_STEERING if args not set")
	flag.StringVar(&throttleFeedbackTopic, "mqtt-topic-throttle-feedback", os.Getenv("MQTT_TOPIC_THROTTLE_FEEDBACK"), "Mqtt topic where to publish throttle feedback, use MQTT_TOPIC_THROTTLE_FEEDBACK if args not set")
	flag.StringVar(&speedZoneTopic, "mqtt-topic-speed-zone", os.Getenv("MQTT_TOPIC_SPEED_ZONE"), "Mqtt topic where to subscribe speed zone events, use MQTT_TOPIC_SPEED_ZONE if args not set")

	flag.Float64Var(&minThrottle, "throttle-min", minThrottle, "Minimum throttle value, use THROTTLE_MIN if args not set")
	flag.Float64Var(&maxThrottle, "throttle-max", maxThrottle, "Minimum throttle value, use THROTTLE_MAX if args not set")
	flag.IntVar(&publishPilotFrequency, "update-pwm-frequency", 2, "Number of throttle event to publish when pilot mode is enabled")

	flag.BoolVar(&enableBrake, "enable-brake-feature", false, "Enable brake to slow car on throttle changes")
	flag.StringVar(&brakeConfig, "brake-configuration", "", "Json file to use to configure brake adaptation when --enable-brake is `true`")

	flag.BoolVar(&enableCustomSteeringProcessor, "enable-custom-steering-processor", false, "Enable custom steering processor to estimate throttle")
	flag.StringVar(&configFileSteeringProcessor, "custom-steering-processor-config", "", "Path to json config to parameter custom steering processor")

	flag.BoolVar(&enableSpeedZone, "enable-speed-zone", false, "Enable speed zone information to estimate throttle")
	flag.Float64Var(&slowZoneThrottle, "slow-zone-throttle", 0.11, "Throttle target for slow speed zone")
	flag.Float64Var(&normalZoneThrottle, "normal-zone-throttle", 0.12, "Throttle target for normal speed zone")
	flag.Float64Var(&fastZoneThrottle, "fast-zone-throttle", 0.13, "Throttle target for fast speed zone")
	flag.Float64Var(&moderateSteering, "moderate-steering", 0.3, "Steering above is considered as moderate")
	flag.Float64Var(&fullSteering, "full-steering", 0.8, "Steering above is considered as full")

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

	zap.S().Infof("Topic throttle                 : %s", throttleTopic)
	zap.S().Infof("Topic rc-throttle              : %s", rcThrottleTopic)
	zap.S().Infof("Topic throttle feedback        : %s", throttleFeedbackTopic)
	zap.S().Infof("Topic steering                 : %s", steeringTopic)
	zap.S().Infof("Topic drive mode               : %s", driveModeTopic)
	zap.S().Infof("Topic speed zone               : %s", speedZoneTopic)
	zap.S().Infof("Min throttle                   : %v", minThrottle)
	zap.S().Infof("Max throttle                   : %v", maxThrottle)
	zap.S().Infof("Publish frequency              : %vHz", publishPilotFrequency)
	zap.S().Infof("Brake enabled                  : %v", enableBrake)
	zap.S().Infof("CustomSteeringProcessor enabled: %v", enableCustomSteeringProcessor)
	zap.S().Infof("SpeedZone enabled              : %v", enableSpeedZone)
	zap.S().Infof("SpeedZone slow throttle        : %v", slowZoneThrottle)
	zap.S().Infof("SpeedZone normal throttle      : %v", normalZoneThrottle)
	zap.S().Infof("SpeedZone fast throttle        : %v", fastZoneThrottle)
	zap.S().Infof("Steering moderate              : %v", moderateSteering)
	zap.S().Infof("Steering full                  : %v", fullSteering)

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

	if enableSpeedZone && enableCustomSteeringProcessor {
		zap.S().Panicf("invalid flag, speedZone and customSteering processor can't be enabled at the same time")
	}
	var throttleProcessor throttle.Processor
	if enableSpeedZone {
		throttleProcessor = throttle.NewSpeedZoneProcessor(
			types.Throttle(slowZoneThrottle),
			types.Throttle(normalZoneThrottle),
			types.Throttle(fastZoneThrottle),
			moderateSteering,
			fullSteering,
		)
	} else if enableCustomSteeringProcessor {
		cfg, err := throttle.NewConfigFromJson(configFileSteeringProcessor)
		if err != nil {
			zap.S().Fatalf("unable to load config '%v': %v", configFileSteeringProcessor, err)
		}
		throttleProcessor = throttle.NewCustomSteeringProcessor(cfg)
	} else {
		throttleProcessor = throttle.NewSteeringProcessor(types.Throttle(minThrottle), types.Throttle(maxThrottle))
	}

	p := throttle.New(client, throttleTopic, driveModeTopic, rcThrottleTopic, steeringTopic, throttleFeedbackTopic,
		maxThrottleCtrlTopic, speedZoneTopic, types.Throttle(maxThrottle), 2,
		throttle.WithThrottleProcessor(throttleProcessor),
		throttle.WithBrakeController(brakeCtrl))
	defer p.Stop()

	cli.HandleExit(p)

	err = p.Start()
	if err != nil {
		zap.S().Fatalf("unable to start service: %v", err)
	}
}
