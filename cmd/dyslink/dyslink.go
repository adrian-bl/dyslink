/*
 * Copyright (c) 2017 Adrian Ulrich <adrian@blinkenlights.ch>
 *
 * All rights reserved. This program and the accompanying materials
 * are made available under the terms of the Eclipse Public License v1.0
 * which accompanies this distribution, and is available at
 * http://www.eclipse.org/legal/epl-v10.html
 *
 */

package main

import (
	"flag"
	"fmt"
	"github.com/adrian-bl/dyslink"
	"os"
)

var flagHost = flag.String("host", "10.0.42.137:1883", "The ip:port combination to connect to")
var flagUser = flag.String("user", "", "The user to use. Part of setup SSID, example: NN4-CH-HEA0322B")
var flagPass = flag.String("password", "", "The passwort to use. See sticker on the manual (or under your fans filter)")
var flagBootstrap = flag.Bool("bootstrap", false, "Bootstrap a factory reseted filter, requires -boot-essid and -boot-password")
var flagBootEssid = flag.String("boot-essid", "i-did-not-read-the-manual", "The essid the fan should join to (while using -bootstrap)")
var flagBootPass = flag.String("boot-password", "", "The password of the wifi network specified via -boot-essid")
var flagHelp = flag.Bool("help", false, "Print what you are currently reading")
var flagHangAround = flag.Bool("hang", false, "Keep running after sending a command to get status updates of the fan")

var flagStateSleep = flag.String("sleep-timer", "", "Sleep timer in minutes, eg: '5'. Passing '0' cancels the timer.")
var flagStateFanSpeed = flag.String("fan-speed", "", "Set fan to this speed (1-10). 0 turns the fan off, -1 uses auto mode.")
var flagStateOscillate = flag.Bool("oscillate", false, "Enable or disable oscillation")
var flagStateNight = flag.Bool("night-mode", false, "Enable or disable night mode")
var flagHighQuality = flag.Bool("high-quality", false, "Target 'high air quality'")

// MODEL is hardcoded for now.
const MODEL = dyslink.TypeModelN475

func main() {
	flag.Parse()
	if len(os.Args) < 2 || *flagHelp == true {
		flag.Usage()
		os.Exit(1)
	}

	cb := make(chan *dyslink.MessageCallback)
	opts := &dyslink.ClientOpts{
		Model:         MODEL,
		Username:      *flagUser,
		Password:      *flagPass,
		DeviceAddress: fmt.Sprintf("tcp://%s", *flagHost),
		CallbackChan:  cb,
	}
	c := dyslink.NewClient(opts)
	err := c.Connect()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to connect to '%s' as '%s', error: %s\n", *flagHost, *flagUser, err)
		os.Exit(2)
	}

	if *flagBootstrap == true {
		*flagHangAround = true
		fmt.Printf("Bootstrapping %s into wifi network %s\n", *flagHost, *flagBootEssid)
		c.WifiBootstrap(*flagBootEssid, *flagBootPass)
		// note that this is untested in the client - but the dyslink itself works.
	} else {
		mode := dyslink.FanModeOn
		if *flagStateFanSpeed == "0" {
			mode = dyslink.FanModeOff
		}
		if *flagStateFanSpeed == "-1" {
			mode = dyslink.FanModeAuto
		}

		state := &dyslink.FanState{
			FanMode:       mode,
			Oscillate:     triGet("oscillate", "", dyslink.OscillateOn, dyslink.OscillateOff),
			FanSpeed:      *flagStateFanSpeed,
			NightMode:     triGet("night-mode", "", dyslink.NightModeOn, dyslink.NightModeOff),
			SleepTimer:    *flagStateSleep,
			QualityTarget: triGet("high-quality", "", dyslink.QualityHigh, dyslink.QualityLow),
		}
		c.SetState(state)
		fmt.Printf("Set: %#v\n", state)
	}

	if *flagHangAround == true {
		fmt.Printf("# waiting for status messages, hit CTRL+C to exit\n")
		for {
			v := <-cb
			fmt.Printf("Message: %+v\n", v.Message)
		}
	}
}

func triGet(flagName, isUndef, isTrue, isFalse string) string {
	// Used to check if the flag was passed or if a default should be used
	passedFlags := make(map[string]bool)
	flag.Visit(func(f *flag.Flag) { passedFlags[f.Name] = true })

	ok := passedFlags[flagName]

	if ok == false {
		return isUndef
	}

	val := flag.Lookup(flagName).Value.(flag.Getter).Get().(bool)
	if val == true {
		return isTrue
	}
	return isFalse
}
