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
	"fmt"
	"github.com/mitchellh/mapstructure"
)

// parseStateChangePayload is a butt ugly version to parse
// unsolicited status changes sent to us
func parseStateChangePayload(p interface{}) (*ProductState, error) {
	m, found := p.(map[string]interface{})
	if found == false {
		return nil, fmt.Errorf("Unexpected interface type")
	}

	state := &ProductState{}
	newson := make(map[string]string)

	for key, intf := range m {
		ilist, found := intf.([]interface{})
		if found == true && len(ilist) == 2 {
			str, found := ilist[1].(string)
			if found == true {
				newson[key] = str
			}
		}
	}

	err := mapstructure.Decode(newson, &state)
	return state, err
}
