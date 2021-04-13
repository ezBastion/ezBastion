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

package main

import (
	"fmt"
	"github.com/pelletier/go-toml"
	"io/ioutil"
	"net"
	"strings"
)

func startMsg(msg string) {
	if len(msg) < 30 {
		msg = fmt.Sprintf("%s%s", msg, strings.Repeat(" ", 30-len(msg)))
	}
	if isVerbose {
		fmt.Print(msg)
	}
}
func endMsg(err error) {
	if isVerbose {
		if err == nil {
			fmt.Println("[OK]")
		} else {
			fmt.Println("[ERROR] ", err)
		}

	}
}
func checkToml() error {
	startMsg("Get config file:")
	raw, readerror := ioutil.ReadFile(confPath)
	if readerror != nil {
		endMsg(err)
		return readerror
	}
	endMsg(nil)
	startMsg("Check config structure:")
	err = toml.Unmarshal(raw, &conf)
	if err != nil {
		endMsg(readerror)
		return err
	}
	endMsg(nil)
	return nil
}
func tcpPing(addr string, port int) error {
	startMsg("ezb_pki TCP connection:")
	_, err := net.Dial("tcp", fmt.Sprintf("%s:%d", addr, port))
	if err != nil {
		endMsg(err)
		return err
	}
	endMsg(nil)
	return nil
}
