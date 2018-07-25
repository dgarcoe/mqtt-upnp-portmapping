FROM resin/raspberry-pi-golang AS build-env
ADD . /src
RUN cd /src && go get -u github.com/eclipse/paho.mqtt.golang && go get -u github.com/NebulousLabs/go-upnp && go build -ldflags "-linkmode external -extldflags -static" -x -o mqtt-upnp-portmapping .

FROM hypriot/rpi-alpine-scratch
WORKDIR /app
COPY --from=build-env /src/mqtt-upnp-portmapping /app/
ENTRYPOINT ["./mqtt-upnp-portmapping"]
