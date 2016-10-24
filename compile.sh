#!/bin/sh
export GOPATH=`pwd`

MQTT="github.com/eclipse/paho.mqtt.golang"

if [ ! -d $MQTT ] ; then
	echo "Fetching $MQTT"
	go get $MQTT
fi

echo "Building..."
go build dyslink.go
