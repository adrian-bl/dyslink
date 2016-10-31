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

// Messages related to fan states
const (
	MessageCurrentState  = "CURRENT-STATE"                     // incoming state data
	MessageEnvSensorData = "ENVIRONMENTAL-CURRENT-SENSOR-DATA" // incoming environmental data
	MessageStateChange   = "STATE-CHANGE"                      // incoming state change data
)

// States of fan modules
const (
	FanModeOff   = "OFF"
	FanModeAuto  = "AUTO"
	FanModeOn    = "FAN"
	NightModeOn  = "ON"
	NightModeOff = "OFF"
	OscillateOn  = "ON"
	OscillateOff = "OFF"
)

// The command-json sent to the device
type commandHeader struct {
	Command      string      `json:"msg"`
	TimeString   string      `json:"time"`
	ModeReason   string      `json:"mode-reason,omitempty"`
	Data         interface{} `json:"data,omitempty"`
	ProductState interface{} `json:"product-state,omitempty"`
}

// A fan status message
// Note that bools and ints are strings, that's because
// they also take the special `StateKep` pragma :-/
type FanState struct {
	FanMode     string `json:"fmod,omitempty"`
	FanSpeed    string `json:"fnsp,omitempty"`
	Oscillate   string `json:"oson,omitempty"`
	SleepTimer  string `json:"sltm,omitempty"`
	Rhtm        string `json:"rhtm,omitempty"` // collect data (??)
	ResetFilter string `json:"rstf,omitempty"` // resets lifetime of filter?
	Qtar        string `json:"qtar,omitempty"`
	NightMode   string `json:"nmod,omitempty"`
}

// A product status message
// Similar to FanState, but this is something we
// receive from a subscription
type ProductState struct {
	FanMode     string `mapstructure:"fmod"`
	FanSpeed    string `mapstructure:"fnsp"`
	Oscillate   string `mapstructure:"oson"`
	SleepTimer  string `mapstructure:"sltm"`
	Rhtm        string `mapstructure:"rhtm"` // collect data (??)
	ResetFilter string `mapstructure:"rstf"` // resets lifetime of filter?
	Qtar        string `mapstructure:"qtar"`
	NightMode   string `mapstructure:"nmod"`
	FilterLife  string `mapstructure:"filf"`
	UnknownErcd string `mapstructure:"ercd"`
	UnknownWacd string `mapstructure:"wacd"`
}

// The current environment data as reported by the device
type EnvironmentState struct {
	Temperature string `mapstructure:"tact"`
	Humidity    string `mapstructure:"hact"`
	Particle    string `mapstructure:"pact"`
	UnknownVact string `mapstructure:"vact"`
	SleepTimer  string `mapstructure:"sltm"`
}
