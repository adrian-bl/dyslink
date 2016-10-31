/*
 * Copyright (c) 2016 Adrian Ulrich
 *
 * All rights reserved. This program and the accompanying materials
 * are made available under the terms of the Eclipse Public License v1.0
 * which accompanies this distribution, and is available at
 * http://www.eclipse.org/legal/epl-v10.html
 *
 */

package dyslink

import (
	"encoding/json"
	"fmt"
	mqtt "github.com/eclipse/paho.mqtt.golang"
	"time"
)

// Fixme: this is clunky
var f mqtt.MessageHandler = func(client mqtt.Client, msg mqtt.Message) {
	hdr := &commandHeader{}
	err := json.Unmarshal(msg.Payload(), &hdr)
	if err == nil {
		if hdr.Command == MessageEnvSensorData && hdr.Data != nil {
			raw, _ := json.Marshal(hdr.Data)
			envstate := &EnvironmentState{}
			json.Unmarshal(raw, &envstate)
			fmt.Printf(">> REJ: %+v\n", envstate)
		} else if hdr.Command == MessageCurrentState && hdr.ProductState != nil {
			raw, _ := json.Marshal(hdr.ProductState)
			prodstate := &ProductState{}
			json.Unmarshal(raw, &prodstate)
			fmt.Printf("> %+v\n", prodstate)
		} else {
			fmt.Printf("Warning: Unknwon state update: %s, json=%s\n", hdr.Command, msg.Payload())
		}
	}
}

type Client interface {
	Connect() error
	Disconnect(uint)
	WifiBootstrap(string, string) error
	SetState(*FanState) error
	RequestCurrentState() error
}

type client struct {
	MqttClient mqtt.Client
	opts       *ClientOpts
}

// Returns a new client
func NewClient(opts *ClientOpts) Client {
	c := &client{opts: opts}
	return c
}

// Establishes a new connection
func (c *client) Connect() error {
	mqttOpts := mqtt.NewClientOptions().AddBroker(c.opts.DeviceAddress)
	mqttOpts.SetUsername(c.opts.Username)
	mqttOpts.SetPassword(c.opts.Password)
	mqttOpts.SetDefaultPublishHandler(f)
	mqttClient := mqtt.NewClient(mqttOpts)
	if token := mqttClient.Connect(); token.Wait() && token.Error() != nil {
		return token.Error()
	}
	mqttClient.Subscribe(c.getDeviceTopic("status/current"), 0, nil)
	c.MqttClient = mqttClient
	return nil
}

// Disconnect disconnects the client
// The quiesce parameter defines how long we are going
// to wait for the connection tear down
func (c *client) Disconnect(quiesce uint) {
	c.MqttClient.Disconnect(quiesce)
	c.MqttClient = nil
}

// Helper function to bootstrap a unconfigured device.
// Not implemented yet
func (c *client) WifiBootstrap(essid string, password string) error {
	return nil
}

// SetState sets the fan to given state
func (c *client) SetState(state *FanState) error {
	cmd := &commandHeader{Command: "STATE-SET", Data: state}
	return c.sendCommand(cmd)
}

// RequestCurrentState asks the connected device to return ENVIRONMENTAL-CURRENT-SENSORT-DATA
// and CURRENT-STATE messages
func (c *client) RequestCurrentState() error {
	cmd := &commandHeader{Command: "REQUEST-CURRENT-STATE"}
	return c.sendCommand(cmd)
}

// sendCommand delivers given command to the device
func (c *client) sendCommand(cmd *commandHeader) error {
	cmd.TimeString = time.Now().UTC().Format(time.RFC3339Nano)

	raw, err := json.Marshal(cmd)
	fmt.Printf("SENDTO: %s\n", raw)
	if err == nil {
		if token := c.MqttClient.Publish(c.getDeviceTopic("command"), 1, false, raw); token.Wait() && token.Error() != nil {
			err = token.Error()
		}
	}
	return err
}

// getDeviceTopic returns the topic we are supposed to send for
// this connection
func (c *client) getDeviceTopic(command string) string {
	return fmt.Sprintf("%s/%s/%s", c.opts.Model, c.opts.Username, command)
}
