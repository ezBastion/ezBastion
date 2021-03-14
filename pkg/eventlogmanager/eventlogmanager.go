// +build windows

// This file is part of ezBastion.

//     ezBastion is free software: you can redistribute it and/or modify
//     it under the terms of the GNU Affero General Public License as published by
//     the Free Software Foundation, either version 3 of the License, or
//     (at your option) any later version.

//     ezBastion is distributed in the hope that it will be useful,
//     but WITHOUT ANY WARRANTY; without even the implied warranty of
//     MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
//     GNU Affero General Public License for more details.

//     You should have received a copy of the GNU Affero General Public License
//     along with ezBastion.  If not, see <https://www.gnu.org/licenses/>.

// Package eventlogmanager add helper for eventlogs on windows
package eventlogmanager

import (
	"errors"
	"fmt"

	"golang.org/x/sys/windows/svc/debug"
	"golang.org/x/sys/windows/svc/eventlog"
)

var Elog debug.Log
var Eventname string
var Status int

func init() {
	Status = -1
}

// Open open a eventlog specified by name, returning nil or an error
func Open(name string) error {
	var err error

	Elog, err = eventlog.Open(name)
	if err != nil {
		Status = 255
		return fmt.Errorf("Cannot Open %s with error %s", name, err.Error())
	}
	Status = 0
	Eventname = name
	return nil
}

// Close closes the event
func Close() error {
	if Status == 0 {
		return Elog.Close()
	}
	return errors.New("Cannot close a non created event")
}
