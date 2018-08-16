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
var sensorType = dht.AM2301

func getBoardSN() string {
	path := "/proc/device-tree/serial-number"
	dat, _ := ioutil.ReadFile(path)
	value := string(dat)
	value = strings.TrimSuffix(value, "\u0000")
	return value
}

func getSensorData(gpioPin int) Data {
	temperature, humidity, err := dht.ReadDHTxx(sensorType, gpioPin, false)
	if temperature != 0 && humidity != 0 && err == nil {
		thingData := Data{
			Temp: temperature,
			Hum:  humidity,
		}
		return thingData
	}
	return Data{}
}

func cmpSensorData(currentData Data, newData Data) bool {
	correctTemp := newData.Temp-currentData.Temp <= 5 && newData.Temp != 0
	correctHum := newData.Hum-currentData.Hum <= 5 && newData.Temp != 0
	return correctTemp && correctHum
}

func main() {
	mqttHostFlag := flag.String("mqtt", "192.168.178.108", "MQTT server address")
	gpioFlag := flag.Int("gpio", 4, "Default GPIO")
	periodicFlag := flag.Int("periodic", 2, "Gather data every number sec") 
	flag.Parse()
	mqttHost := *mqttHostFlag
	gpioPin := *gpioFlag
	periodicInt := time.Duration(*periodicFlag)
        periodicTime := periodicInt * 1000 * time.Millisecond
	const TOPIC = "/sensors/climat"
	port := 1883
	fmt.Println("Connecting to MQTT server")
	connOpts := &mqtt.ClientOptions{
		ClientID:      serialNumber,
		CleanSession:  true,
		AutoReconnect: true,
	}

	brokerURL := fmt.Sprintf("tcp://%s:%d", mqttHost, port)
	fmt.Printf("MQTT url to server: %v\n", brokerURL)
	connOpts.AddBroker(brokerURL)

	client := mqtt.NewClient(connOpts)
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		panic(token.Error())
	}
	fmt.Println("Connection to MQTT server established")
	fmt.Println("Start fetching data from sensor")

	thingData := Data{}
	for {
		newThingData := getSensorData(gpioPin)
		if newThingData != thingData {
			if cmpSensorData(thingData, newThingData) {
				newThingData.ThingID = serialNumber
				data, _ := json.Marshal(newThingData)
				client.Publish(TOPIC, 0, false, data)
			}
			fmt.Println("New data was sent to server")
			thingData = newThingData
		}
		time.Sleep(periodicTime)
	}
}
