package mqttdevice

import (
	"encoding/json"
	"fmt"
	"github.com/cyrilix/robocar-base/types"
	MQTT "github.com/eclipse/paho.mqtt.golang"
	"log"
	"strconv"
)

type Publisher interface {
	Publish(topic string, payload MqttValue)
}

type Subscriber interface {
	Subscribe(topic string, mh MQTT.MessageHandler)
}

type MQTTPubSub interface {
	Publisher
	Subscriber
}

type pahoMqttPubSub struct {
	client   MQTT.Client
	qos      int
	retain   bool
}

func NewPahoMqttPubSub(client MQTT.Client, qos int, retain bool) MQTTPubSub {
	p := pahoMqttPubSub{client: client, qos: qos, retain: retain}
	return &p
}

func Connect(uri, username, password, clientId string) (MQTT.Client, error) {
	//create a ClientOptions struct setting the broker address, clientid, turn
	//off trace output and set the default message handler
	opts := MQTT.NewClientOptions().AddBroker(uri)
	opts.SetUsername(username)
	opts.SetPassword(password)
	opts.SetClientID(clientId)
	opts.SetAutoReconnect(true)
	opts.SetDefaultPublishHandler(
		//define a function for the default message handler
		func(client MQTT.Client, msg MQTT.Message) {
			fmt.Printf("TOPIC: %s\n", msg.Topic())
			fmt.Printf("MSG: %s\n", msg.Payload())
		})

	//create and start a client using the above ClientOptions
	client := MQTT.NewClient(opts)
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		return nil, fmt.Errorf("unable to connect to mqtt bus: %v", token.Error())
	}
	return client, nil
}

// Publish message to broker
func (p *pahoMqttPubSub) Publish(topic string, payload MqttValue) {
	tokenResp := p.client.Publish(topic, byte(p.qos), p.retain, string(payload))
	if tokenResp.Error() != nil {
		log.Fatalf("%+v\n", tokenResp.Error())
	}
}

// Register func to execute on message
func (p *pahoMqttPubSub) Subscribe(topic string, callback MQTT.MessageHandler) {
	tokenResp := p.client.Subscribe(topic, byte(p.qos), callback)
	if tokenResp.Error() != nil {
		log.Fatalf("%+v\n", tokenResp.Error())
	}
}

type MqttValue []byte

func NewMqttValue(v interface{}) MqttValue {
	switch val := v.(type) {
	case string:
		return MqttValue(val)
	case float32, float64:
		return MqttValue(fmt.Sprintf("%0.2f", val))
	case int, int8, int16, int32, int64:
		return MqttValue(fmt.Sprintf("%d", val))
	case bool:
		if val {
			return []byte("ON")
		} else {
			return []byte("OFF")
		}
	case []byte:
		return val
	case MqttValue:
		return val
	default:
		jsonValue, err := json.Marshal(v)
		if err != nil {
			log.Printf("unable to mashall to json value '%v': %v", v, err)
			return nil
		}
		return jsonValue
	}
}

func (m *MqttValue) IntValue() (int, error) {
	return strconv.Atoi(string(*m))
}

func (m *MqttValue) Float32Value() (float32, error) {
	val := string(*m)
	r, err := strconv.ParseFloat(val, 32)
	return float32(r), err
}
func (m *MqttValue) Float64Value() (float64, error) {
	val := string(*m)
	return strconv.ParseFloat(val, 64)
}
func (m *MqttValue) StringValue() (string, error) {
	return string(*m), nil
}
func (m *MqttValue) DriveModeValue() (types.DriveMode, error) {
	val, err := m.IntValue()
	if err != nil {
		return types.DriveModeInvalid, err
	}
	return types.DriveMode(val), nil
}
func (m *MqttValue) ByteSliceValue() ([]byte, error) {
	return *m, nil
}
func (m *MqttValue) BoolValue() (bool, error) {
	val := string(*m)
	switch val {
	case "ON":
		return true, nil
	case "OFF":
		return false, nil
	default:
		return false, fmt.Errorf("value %v can't be converted to bool", val)
	}
}
