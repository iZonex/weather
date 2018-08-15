package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"strings"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	dht "github.com/iZonex/go-dht"
)

type Data struct {
	ThingID string  `json:"thing_id"`
	Temp    float32 `json:"temperature"`
	Hum     float32 `json:"humidity"`
}

var serialNumber = getBoardSN()
var sensorType = dht.DHT22

func getBoardSN() string {
	path := "/proc/device-tree/serial-number"
	dat, _ := ioutil.ReadFile(path)
	value := string(dat)
	value = strings.TrimSuffix(value, "\u0000")
	return value
}

func getSensorData() Data {
	temperature, humidity, _ := dht.ReadDHTxx(sensorType, 4, false)
	if temperature != -1 && humidity != -1 {
		thingData := Data{
			Temp: temperature,
			Hum:  humidity,
		}
		return thingData
	}
	return Data{}
}

func cmpSensorData(currentData Data, newData Data) bool {
	correctTemp := newData.Temp-currentData.Temp <= 5
	correctHum := newData.Hum-currentData.Hum <= 5
	return correctTemp && correctHum
}

func main() {
	mqttHostFlag := flag.String("mqtt", "localhost", "MQTT server address")
	flag.Parse()
	mqttHost := *mqttHostFlag
	const TOPIC = "/sensors/climat"
	port := 1883
	connOpts := &mqtt.ClientOptions{
		ClientID:      serialNumber,
		CleanSession:  true,
		AutoReconnect: true,
	}

	brokerURL := fmt.Sprintf("tcp://%s:%d", mqttHost, port)
	connOpts.AddBroker(brokerURL)

	client := mqtt.NewClient(connOpts)
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		panic(token.Error())
	}

	thingData := Data{}
	for {
		newThingData := getSensorData()
		if newThingData != thingData {
			if cmpSensorData(thingData, newThingData) {
				newThingData.ThingID = serialNumber
				data, _ := json.Marshal(newThingData)
				client.Publish(TOPIC, 0, false, data)
			}
			thingData = newThingData
		}
		time.Sleep(2000 * time.Millisecond)
	}
}
