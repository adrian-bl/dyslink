package main

/*
 * Copyright (c) 2019 Adrian Ulrich <adrian@blinkenlights.ch>
 *
 * All rights reserved. This program and the accompanying materials
 * are made available under the terms of the Eclipse Public License v1.0
 * which accompanies this distribution, and is available at
 * http://www.eclipse.org/legal/epl-v10.html
 *
 */

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"sync"

	"github.com/adrian-bl/dyslink/lib/dyslink"
)

var (
	flagHost   = flag.String("host", "10.0.42.137:1883", "The ip:port combination to connect to")
	flagUser   = flag.String("user", "", "The user to use. Part of setup SSID, example: NN4-CH-HEA0322B")
	flagPass   = flag.String("password", "", "The passwort to use. See sticker on the manual (or under your fans filter)")
	flagListen = flag.String("listen", "127.0.0.1:9033", "ip:port to listen on")
)

var state = struct {
	sync.RWMutex
	Fan dyslink.ProductState     `json:"Fan"`
	Env dyslink.EnvironmentState `json:"Env"`
}{}

func main() {
	flag.Parse()

	cb := make(chan *dyslink.MessageCallback)
	opts := &dyslink.ClientOpts{
		Model:         dyslink.TypeModelN475,
		Username:      *flagUser,
		Password:      *flagPass,
		DeviceAddress: fmt.Sprintf("tcp://%s", *flagHost),
		CallbackChan:  cb,
	}

	c := dyslink.NewClient(opts)
	if err := c.Connect(); err != nil {
		log.Fatalf("failed to connect to '%s': %v", *flagHost, err)
	}

	ctx := context.Background()

	go func() {
		if err := serveHttp(ctx, c, *flagListen); err != nil {
			log.Printf("serveHttp err: %v", err)
		}
	}()
	go monitorStatus(ctx, c, cb)
	<-ctx.Done()
}

func monitorStatus(ctx context.Context, c dyslink.Client, cb chan *dyslink.MessageCallback) {

	c.RequestCurrentState()
	for {
		select {
		case <-ctx.Done():
			return
		case msg := <-cb:
			if msg.Error == nil {
				fmt.Printf("> %+v\n", msg)
				state.Lock()
				if v, ok := msg.Message.(*dyslink.ProductState); ok {
					state.Fan = *v
				}
				if v, ok := msg.Message.(*dyslink.EnvironmentState); ok {
					state.Env = *v
				}
				state.Unlock()
			}
		}
	}
}

// serveHttp setups the http server.
func serveHttp(ctx context.Context, c dyslink.Client, addr string) error {
	srv := &http.Server{
		Addr: addr,
	}

	go func() {
		<-ctx.Done()
		srv.Shutdown(ctx)
	}()
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		handleHttp(c, w, r)
	})
	return srv.ListenAndServe()
}

// handleHttp dispatches http requests.
func handleHttp(c dyslink.Client, w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" && r.URL.String() == "/" {
		serveIndex(w)
	} else if r.Method == "GET" && r.URL.String() == "/getstate.json" {
		serveState(w)
	} else if r.Method == "POST" && r.URL.String() == "/setstate.json" {
		setState(c, w, r)
	} else {
		w.WriteHeader(404)
	}
}

// serveState serves the current fan state as json.
func serveState(w http.ResponseWriter) {
	state.RLock()
	defer state.RUnlock()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(state)
}

func setState(c dyslink.Client, w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	w.WriteHeader(200)

	state := &dyslink.FanState{}
	if v := r.Form["mode"]; len(v) == 1 {
		switch v[0] {
		case "OFF":
			state.FanMode = dyslink.FanModeOff
		case "FAN":
			state.FanMode = dyslink.FanModeOn
		case "AUTO":
			state.FanMode = dyslink.FanModeAuto
		}
	}

	if v := r.Form["speed"]; len(v) == 1 {
		state.FanSpeed = v[0]
	}

	if v := r.Form["rotate"]; len(v) == 1 {
		switch v[0] {
		case "ON":
			state.Oscillate = dyslink.OscillateOn
		case "OFF":
			state.Oscillate = dyslink.OscillateOff
		}
	}

	c.SetState(state)
}

// serveIndex serves the main html.
func serveIndex(w http.ResponseWriter) {
	w.WriteHeader(200)
	w.Write([]byte(`<html>
<head>
<title>Dyslink</title>
<meta name="viewport" content="width=device-width, initial-scale=1.0">

<script src="https://ajax.googleapis.com/ajax/libs/jquery/3.4.1/jquery.min.js"></script>

<style>
.toptitle {
  text-align: center;
  font-size: 2.5em;
  color: #606060;
  text-shadow: 2px 2px 12px #202020;
}
.title {
  text-align: center;
  font-size: 2em;
  color: #505050;
  text-shadow: 2px 2px 8px #101010;
}
.select {
  font-size: 1.2em;
  display: block;
  margin: 0 auto;
}
</style>

</head>
<body bgcolor="#222222">
<div class="toptitle">Fan Web UI</div>
<br><br>
<div id="ui" style="visibility: hidden;">

<div class="title">Mode</div>
<select class="select" id="mode">
  <option value="OFF">Off</option>
  <option value="FAN">On</option>
  <option value="AUTO">Auto</option>
</select>
<br>

<div class="title">Fan Speed</div>
<select class="select" id="speed">
  <option value="1">1</option>
  <option value="2">2</option>
  <option value="3">3</option>
  <option value="4">4</option>
  <option value="5">5</option>
  <option value="6">6</option>
  <option value="7">7</option>
  <option value="8">8</option>
  <option value="9">9</option>
  <option value="10">10</option>
</select>
<br>

<div class="title">Rotation</div>
<select class="select" id="rotate">
  <option value="OFF">Off</option>
  <option value="ON">Rotate</option>
</select>
</div>

<script>

var busy = 0;

function setFan() {
  busy = 1;
  $.ajax({
    type: "POST",
    url: "setstate.json",
    data: {
      mode: $("#mode").val(),
      speed: $("#speed").val(),
      rotate: $("#rotate").val(),
    },
  });
}

$("#mode").change(function()   { setFan(); });
$("#speed").change(function()  { setFan(); });
$("#rotate").change(function() { setFan(); });

(function poll() {
    $.ajax({
        url: "getstate.json",
        type: "GET",
        success: function(data) {
            if (busy == 0) {
              restoreUI(data);
            }
        },
        dataType: "json",
        complete: setTimeout(function() {poll()}, 500),
        timeout: 2000
    })
})();

function restoreUI(data) {
  fs = parseInt(data.Fan.FanSpeed)
  if (!isNaN(fs)) {
    $('#speed').val(fs);
  }
  $('#rotate').val(data.Fan.Oscillate);
  $('#mode').val(data.Fan.FanMode);
  $('#ui').css("visibility", "visible");
  busy = 0;
}

</script>

</body>
</html>
`))
}
