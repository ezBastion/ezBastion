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

package setupmanager

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"path"
	"regexp"
	"strings"
)

//CheckFolder
func CheckFolder(exPath string) error {

	if _, err := os.Stat(path.Join(exPath, "cert")); os.IsNotExist(err) {
		err = os.MkdirAll(path.Join(exPath, "cert"), 0600)
		if err != nil {
			return err
		}
		log.Println("Make cert folder.")
	}
	if _, err := os.Stat(path.Join(exPath, "log")); os.IsNotExist(err) {
		err = os.MkdirAll(path.Join(exPath, "log"), 0600)
		if err != nil {
			return err
		}
		log.Println("Make log folder.")
	}
	if _, err := os.Stat(path.Join(exPath, "conf")); os.IsNotExist(err) {
		err = os.MkdirAll(path.Join(exPath, "conf"), 0600)
		if err != nil {
			return err
		}
		log.Println("Make conf folder.")
	}
	return nil
}

//AskForConfirmation waiting for user yes or no
func AskForConfirmation(s string) bool {
	reader := bufio.NewReader(os.Stdin)

	for {
		fmt.Printf("\n%s [y/n]: ", s)

		response, err := reader.ReadString('\n')
		if err != nil {
			fmt.Println(err)
		}

		response = strings.ToLower(strings.TrimSpace(response))

		if response == "y" || response == "yes" {
			return true
		} else if response == "n" || response == "no" {
			return false
		}
	}
}
func AskForValue(s, def string, pattern string) string {
	reader := bufio.NewReader(os.Stdin)
	re := regexp.MustCompile(pattern)
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
			return response
		} else {
			fmt.Printf("[%s] wrong format, must match (%s)\n", response, pattern)
		}
		fmt.Printf("[%s] wrong format, must match (%s)\n", response, pattern)
	}
}
