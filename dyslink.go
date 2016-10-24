package main

import (
  "fmt"
  mqtt "github.com/eclipse/paho.mqtt.golang"
  "os"
  "strconv"
)

var f mqtt.MessageHandler = func(client mqtt.Client, msg mqtt.Message) {
  fmt.Printf("TOPIC: %s\n", msg.Topic())
  fmt.Printf("MSG: %s\n", msg.Payload())
}

func main() {

	if len(os.Args) != 2 {
		fmt.Printf("Usage: %s SPEED\n", os.Args[0])
		os.Exit(1)
	}

	oson     := "OFF"
	speed, _ := strconv.Atoi(os.Args[1])

	if speed < 0 {
		oson = "ON"
		speed = speed * -1
	}

	fmt.Printf(">> Setting fan speed to %d, rotation = %s\n", speed, oson)

	opts := mqtt.NewClientOptions().AddBroker("tcp://192.168.1.196:1883")
	opts.SetClientID("gogogirl")
	opts.SetUsername("NN4-CH-HEA0429A")
	opts.SetPassword("1Itjufh6BiOfhiQgwyGpcxQuLAx8XOsdoe3p1fiATnlazw6TC1sfhoEPuEsf8//BXA22NWTuwI5bSHurWCzlZw==")
	opts.SetDefaultPublishHandler(f)

	client := mqtt.NewClient(opts)
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		panic(token.Error())
	}

	fmt.Printf(">> connection established!\n")

	x := client.Publish("475/NN4-CH-HEA0429A/command", 1, false, fmt.Sprintf("{\"msg\":\"STATE-SET\",\"time\":\"2016-10-24T19:45:09Z\",\"mode-reason\":\"LAPP\",\"data\":{\"fmod\":\"FAN\",\"fnsp\":\"%04d\",\"oson\":\"%s\",\"sltm\":\"STET\",\"rhtm\":\"OFF\",\"rstf\":\"STET\",\"qtar\":\"0003\",\"nmod\":\"OFF\"}}", speed, oson))
	x.Wait()
	fmt.Printf(">> cmd done: err=%+v\n", x.Error())
}
