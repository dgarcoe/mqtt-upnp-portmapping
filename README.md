# mqtt-upnp-portmapping
Performs UPnP portmapping requested through a MQTT topic. It also publishes the external IP address through MQTT.

## Installation
A Dockerfile for ARM is provided. To install the container in your Raspberry you need of course to have a working docker and docker-compose installation. Afterwards, create a .env file in the directory of the application with a content like the following one, updating your data for the MQTT broker as needed:

```
MQTT_BROKER=ip:port
MQTT_TOPIC=upnp
MQTT_USER=user1
MQTT_PWD=sup3rs3cr3tpwd
PERIOD=3
```
Then, simply run the docker-compose up command and wait until the container is running:

```
docker-compose up -d
```

Otherwise, you can install the software with the following commands:

```
$ go get -u github.com/eclipse/paho.mqtt.golang
$ go get -u github.com/NebulousLabs/go-upnp
$ go install
```

## Running the application

If you are not using the docker container you can use the following command to run the application:

```
sudo ./mqtt-hub-ctrl -mqttBroker ip:port -topic upnp -user user1 -password sup3rs3cr3tpwd -period 3
```
The application will use two topics based on the string passed when running it. For example, if the topic is set to upnp the topics used by the application will be upnp/externalip and upnp/portmapping.

The application will publish the external IP in the topic upnp/externalip each 'period' hours.

To forward a port, a JSON message with the following fields should be sent through the upnp/portmapping topic.

```
{"Type":"FWD","Port":8080,"Description":"Web server"}
```
To clear a forwarded port a JSON message with the following fields should be sent through the upnp/portmapping topic.

```
{"Type":"CLR","Port":8080}
```
