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
	"crypto/sha512"
	"encoding/base64"
	"encoding/json"
	"fmt"
	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/mitchellh/mapstructure"
	"time"
)

func sendMessageCallback(ch chan<- *MessageCallback, msg mqtt.Message, debug bool) {
	var rv interface{}
	hdr := &commandHeader{}
	err := json.Unmarshal(msg.Payload(), &hdr)
	if debug {
		fmt.Printf("<< raw: %s\n", msg.Payload())
	}
	if err == nil {
		switch hdr.Command {
		case MessageEnvSensorData:
			envstate := &EnvironmentState{}
			err = mapstructure.Decode(hdr.Data, &envstate)
			rv = envstate
		case MessageCurrentState:
			prodstate := &ProductState{}
			err = mapstructure.Decode(hdr.ProductState, &prodstate)
			rv = prodstate
		case MessageDeviceCredentials:
			devcred := &DeviceCredentials{}
			err = json.Unmarshal(msg.Payload(), &devcred)
			rv = devcred
		case MessageStateChange:
			rv, err = parseStateChangePayload(hdr.ProductState)
		default:
			fmt.Printf("Warning: Unknown state update: %s, json=%s\n", hdr.Command, msg.Payload())
		}
	}
	if ch != nil {
		ch <- &MessageCallback{Error: err, Message: rv}
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

func encodePassword(in string) string {
	bv := []byte(in)
	hasher := sha512.New()
	hasher.Write(bv)
	return base64.StdEncoding.EncodeToString(hasher.Sum(nil))
}

// Establishes a new connection
func (c *client) Connect() error {
	mqttOpts := mqtt.NewClientOptions().AddBroker(c.opts.DeviceAddress)
	mqttOpts.SetUsername(c.opts.Username)
	mqttOpts.SetPassword(encodePassword(c.opts.Password))
	mqttOpts.SetDefaultPublishHandler(
		func(client mqtt.Client, msg mqtt.Message) {
			sendMessageCallback(c.opts.CallbackChan, msg, c.opts.Debug)
		})
	mqttOpts.SetOnConnectHandler(func(mclient mqtt.Client) {
		mclient.Subscribe(c.getDeviceTopic("status/current"), 0, nil)
	})
	mqttClient := mqtt.NewClient(mqttOpts)
	if token := mqttClient.Connect(); token.Wait() && token.Error() != nil {
		return token.Error()
	}
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
func (c *client) WifiBootstrap(essid string, password string) error {
	c.opts.Username = "initialconnection" // username is part of the topic: the credentials cant/were-not used for this connection, so we are just overwriting them
	c.opts.Password = ""
	// first, subscribe to these special endpoints:
	c.MqttClient.Subscribe(c.getDeviceTopic("credentials"), 0, nil).Wait()
	// ..and assemble our commands:
	c.sendCommand(&commandHeader{Command: MessageJoinNetwork, WifiSsid: essid, WifiPassword: password, RequestId: "0123456789ABCDEF"})
	c.sendCommand(&commandHeader{Command: MessageAuthoriseUserRequest, RequestId: "01234567890ABCDEF", Id: "00000000-0000-0000-0000-000000000000"})
	c.sendCommand(&commandHeader{Command: MessageCloseAccessPoint})
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
	if c.opts.Debug {
		fmt.Printf("SENDTO: %s\n", raw)
	}
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
