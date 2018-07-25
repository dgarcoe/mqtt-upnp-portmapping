package main

import (
	"context"
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
	d, err := upnp.DiscoverCtx(ctx)
	if err != nil {
		log.Printf("Error discovering router: %s", err)
	}

	log.Printf("Router discovered")

	for {

		var msg string

		ip, err := d.ExternalIP()
		if err != nil {
			log.Printf("Error getting external IP from router: %s", err)
		}

		log.Printf("UPNP external IP: %s", ip)
		msg = ip

		if token := clientMQTT.Publish(*topic, 0, false, msg); token.Wait() && token.Error() != nil {
			log.Printf("Error publishing message to MQTT broker: %s", token.Error())
		}

		time.Sleep(time.Duration(*period) * time.Second)

	}

}
