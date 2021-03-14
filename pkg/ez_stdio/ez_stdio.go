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
//     along with ezBastion.  If not, see <https://www.gnu.org/licenses/>.`

package ez_stdio

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/ezBastion/ezb_lib/logmanager"
)

// askForConfirmation : Reads the stdin for an confirmation aka answer - ONLY yes/no
func AskForConfirmation(s string) bool {
	reader := bufio.NewReader(os.Stdin)
	logmanager.Debug(fmt.Sprintf("AskForConfirmation : %s",s))
	for {
		fmt.Printf("\n%s [y/n]: ", s)

		response, err := reader.ReadString('\n')
		if err != nil {
			fmt.Println(err)
		}

		response = strings.ToLower(strings.TrimSpace(response))

		if response == "y" || response == "yes" {
			logmanager.Debug("AskForConfirmation : TRUE")
			return true
		} else if response == "n" || response == "no" {
			logmanager.Debug("AskForConfirmation : FALSE")
			return false
		}
	}
}

// askForValue : Reads the stdin for an answer
func AskForValue(s, def string, pattern string) string {
	reader := bufio.NewReader(os.Stdin)
	re := regexp.MustCompile(pattern)
	logmanager.Debug(fmt.Sprintf("AskForValue : %s with default %s",s, def))
	for {
		fmt.Printf("%s [%s]: ", s, def)

		response, err := reader.ReadString('\n')
		if err != nil {
			fmt.Println(err)
		}

		response = strings.TrimSpace(response)
		if response == "" {
			return def
		} else if re.MatchString(response) {
			logmanager.Debug(fmt.Sprintf("AskForValue return : %s",response))
			return response
		} else {
			fmt.Printf("[%s] wrong format, must match (%s)\n", response, pattern)
		}
	}
}

func AskForStringValue(s string) string {
	reader := bufio.NewReader(os.Stdin)
	logmanager.Debug(fmt.Sprintf("AskForStringValue : %s",s))
	for {
		fmt.Printf("%s ", s)

		response, err := reader.ReadString('\n')
		if err != nil {
			fmt.Println(err)
		}

		response = strings.TrimSpace(response)
		logmanager.Debug(fmt.Sprintf("AskForStringValue return : %s",response))
		return response
	}
}
