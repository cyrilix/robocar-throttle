# robocar-throttle

Microservice part to manage throttle
     
## Usage
```
rc-throttle <OPTIONS>

  -mqtt-broker string
        Broker Uri, use MQTT_BROKER env if arg not set (default "tcp://127.0.0.1:1883")
  -mqtt-client-id string
        Mqtt client id, use MQTT_CLIENT_ID env if args not set (default "robocar-throttle")
  -mqtt-password string
        Broker Password, MQTT_PASSWORD env if args not set
  -mqtt-qos int
        Qos to pusblish message, use MQTT_QOS env if arg not set
  -mqtt-retain
        Retain mqtt message, if not set, true if MQTT_RETAIN env variable is set
  -mqtt-topic-throttle string
        Mqtt topic to publish throttle result, use MQTT_TOPIC_THROTTLE if args not set
  -mqtt-username string
        Broker Username, use MQTT_USERNAME env if arg not set
  -throttle-max float
        Minimum throttle value, use THROTTLE_MAX if args not set (default 0.3)
  -throttle-min float
        Minimum throttle value, use THROTTLE_MIN if args not set (default 0.3)
```

## Docker build

```bash
export DOCKER_CLI_EXPERIMENTAL=enabled
docker buildx build . --platform linux/amd64,linux/arm/7,linux/arm64 -t cyrilix/robocar-throttle
```
