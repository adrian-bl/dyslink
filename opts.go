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

const (
	TypeModelN475 = "475" // pure link cool (non-desk)
	TypeModelN469 = "469" // pure link cool round/desk
	TypeModelN455 = "455" // pure hot & cool
)

type MessageCallback struct {
	Error   error
	Message interface{}
}

type ClientOpts struct {
	Username      string // The username to use for this connection
	Password      string // The password to use for this connection
	DeviceAddress string // The ip+port of the device in the tcp://IP:PORT format
	Model         string // One of the TypeModel* constants
	CallbackChan  chan<- *MessageCallback
	Debug         bool
}
