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
	"math"
)

// Messages related to fan states
const (
	MessageCurrentState         = "CURRENT-STATE"                     // incoming state data
	MessageEnvSensorData        = "ENVIRONMENTAL-CURRENT-SENSOR-DATA" // incoming environmental data
	MessageStateChange          = "STATE-CHANGE"                      // incoming state change data
	MessageJoinNetwork          = "JOIN-NETWORK"
	MessageAuthoriseUserRequest = "AUTHORISE-USER-REQUEST"
	MessageCloseAccessPoint     = "CLOSE-ACCESS-POINT"
	MessageDeviceCredentials    = "DEVICE-CREDENTIALS"
)

// States of fan modules
const (
	FanModeOff        = "OFF"
	FanModeAuto       = "AUTO"
	FanModeOn         = "FAN"
	FanSpeedAuto      = "AUTO"
	NightModeOn       = "ON"
	NightModeOff      = "OFF"
	OscillateOn       = "ON"
	OscillateOff      = "OFF"
	QualityLow        = "0001"
	QualityNormal     = "0003"
	QualityHigh       = "0004"
	HeatModeOn        = "HEAT"
	HeatModeOff       = "OFF"
	StandbyMonitorOn  = "ON"
	StandbyMonitorOff = "OFF"
	FocusedModeOn     = "ON"
	FocusedModeOff    = "OFF"
)

// The command-json sent to the device
type commandHeader struct {
	Command      string      `json:"msg"`
	TimeString   string      `json:"time"`
	ModeReason   string      `json:"mode-reason,omitempty"`
	Data         interface{} `json:"data,omitempty"`
	ProductState interface{} `json:"product-state,omitempty"`
	RequestId    string      `json:"requestId,omitempty"`
	Id           string      `json:"id,omitempty"`
	WifiSsid     string      `json:"ssid,omitempty"`
	WifiPassword string      `json:"password,omitempty"`
}

// A fan status message
// Note that bools and ints are strings, that's because
// they also take the special `StateKep` pragma :-/
type FanState struct {
	FanMode           string `json:"fmod,omitempty"`
	FanSpeed          string `json:"fnsp,omitempty"`
	Oscillate         string `json:"oson,omitempty"`
	SleepTimer        string `json:"sltm,omitempty"`
	StandbyMonitoring string `json:"rhtm,omitempty"` // always run + capture environment data
	ResetFilter       string `json:"rstf,omitempty"` // resets lifetime of filter?
	QualityTarget     string `json:"qtar,omitempty"` // the air-target in auto-mode
	NightMode         string `json:"nmod,omitempty"`
	HeatMode          string `json:"hmod,omitempty"`
	HeatTarget        string `json:"hmax,omitempty"`
	FocusedMode       string `json:"ffoc,omitempty"`
}

// A product status message
// Similar to FanState, but this is something we
// receive from a subscription
type ProductState struct {
	FanMode           string `mapstructure:"fmod"`
	FanState          string `mapstructure:"fnst"`
	FanSpeed          string `mapstructure:"fnsp"`
	Oscillate         string `mapstructure:"oson"`
	SleepTimer        string `mapstructure:"sltm"`
	StandbyMonitoring string `mapstructure:"rhtm"`
	ResetFilter       string `mapstructure:"rstf"` // resets lifetime of filter?
	QualityTarget     string `mapstructure:"qtar"`
	NightMode         string `mapstructure:"nmod"`
	HeatMode          string `mapstructure:"hmod"`
	HeatState         string `mapstructure:"hsta"`
	HeatTarget        string `mapstructure:"hmax"`
	FilterLife        string `mapstructure:"filf"`
	FocusedMode       string `mapstructure:"ffoc"`
	UnknownErcd       string `mapstructure:"ercd"`
	UnknownWacd       string `mapstructure:"wacd"`
	UnknownRhtm       string `mapstructure:"rhtm"`
	UnknownTilt       string `mapstructure:"tilt"`
}

// The current environment data as reported by the device
type EnvironmentState struct {
	Temperature string `mapstructure:"tact"`
	Humidity    string `mapstructure:"hact"`
	Particle    string `mapstructure:"pact"`
	UnknownVact string `mapstructure:"vact"`
	SleepTimer  string `mapstructure:"sltm"`
}

// Reply for a credentials request (note: this is sent in a commandHeader)
type DeviceCredentials struct {
	SerialNumber string `json:"serialNumber"`
	Password     string `json:"apPasswordHash"`
}

func round(f float64) int {
	if math.Abs(f) < 0.5 {
		return 0
	}
	return int(f + math.Copysign(0.5000, f))
}

func ConvertTempToFahr(temp int) int {
	return 33 + round(float64(temp-2736)/5.54)
}

func ConvertTempFromFahr(temp int) int {
	return 2736 + round(float64(temp-33)*5.54)
}
