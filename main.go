package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	upnp "github.com/NebulousLabs/go-upnp"
	mqtt "github.com/eclipse/paho.mqtt.golang"
)

const externalIPService string = "http://ipecho.net/plain"

var (
	mqttBroker = flag.String("mqttBroker", "", "MQTT broker URI (mandatory). E.g.:192.168.1.1:1883")
	topic      = flag.String("topic", "", "Topic where hub-ctrl messages will be received (mandatory)")
	user       = flag.String("user", "", "MQTT username")
	pwd        = flag.String("password", "", "MQTT password")
	period     = flag.Int("period", 3, "Periodic time in hours to recheck the external IP address")
)

var router *upnp.IGD

//Message Used to hold MQTT JSON messages
type Message struct {
	Type        string
	Port        int
	Description string
}

//Connect to the MQTT broker
func connectMQTT() (mqtt.Client, error) {
	opts := mqtt.NewClientOptions().AddBroker("tcp://" + *mqttBroker)

	if *user != "" && *pwd != "" {
		opts.SetUsername(*user).SetPassword(*pwd)
	}

	client := mqtt.NewClient(opts)

	if token := client.Connect(); token.Wait() && token.Error() != nil {
		return nil, fmt.Errorf("%s", token.Error())
	}

	return client, nil
}

//Callback for MQTT messages received through the subscribed topic
func mqttCallback(client mqtt.Client, msg mqtt.Message) {

	var jsonMessage Message
	log.Printf("Message received: %s", msg.Payload())

	err := json.Unmarshal(msg.Payload(), &jsonMessage)
	if err != nil {
		log.Printf("Error parsing JSON: %s", err)
	}

	typeMsg := jsonMessage.Type
	port := jsonMessage.Port

	switch typeMsg {
	case "FWD":
		desc := jsonMessage.Description

		err = router.Forward(uint16(port), desc)
		if err != nil {
			log.Printf("Error forwarding port: %s", err)
		}

		log.Printf("Port %d forwarded", port)
	case "CLR":
		err = router.Clear(uint16(port))
		if err != nil {
			log.Printf("Error deleting forwarded port: %s", err)
		}

		log.Printf("Port %d cleared", port)
	}

}

func init() {
	flag.Parse()
}

func main() {

	//Check command line parameters
	if *mqttBroker == "" || *topic == "" {
		flag.PrintDefaults()
		os.Exit(1)
	}

	clientMQTT, err := connectMQTT()
	if err != nil {
		log.Fatalf("Error connecting to MQTT broker: %s", err)
	}

	log.Printf("Connected to MQTT broker at %s", *mqttBroker)

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	router, err = upnp.DiscoverCtx(ctx)
	if err != nil {
		log.Fatalf("Error discovering router: %s", err)
	}

	log.Printf("Router discovered")

	if token := clientMQTT.Subscribe(*topic+"/portmapping", 0, mqttCallback); token.Wait() && token.Error() != nil {
		log.Fatalf("Error subscribing to topic %s : %s", *topic+"/portmapping", err)
	}

	log.Printf("Subscribed to topic %s", *topic+"/portmapping")

	for {

		var msg string

		ip, err := router.ExternalIP()
		if err != nil {
			log.Printf("Error getting external IP from router: %s", err)
		}

		log.Printf("UPNP external IP: %s", ip)
		msg = ip

		if token := clientMQTT.Publish(*topic+"/externalip", 0, false, msg); token.Wait() && token.Error() != nil {
			log.Printf("Error publishing message to MQTT broker: %s", token.Error())
		}

		time.Sleep(time.Duration(*period) * time.Hour)

	}

}
