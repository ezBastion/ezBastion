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

// +build windows

package servicemanager

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
	"golang.org/x/sys/windows/svc"
	eventlog "golang.org/x/sys/windows/svc/eventlog"
	"golang.org/x/sys/windows/svc/mgr"
)

// EventLog management
var exPath string
var err error

func init() {
	ex, _ := os.Executable()
	exPath = filepath.Dir(ex)
}

func exePath() (string, error) {

	prog := os.Args[0]
	p, err := filepath.Abs(prog)
	if err != nil {
		return "", err
	}
	fi, err := os.Stat(p)
	if err == nil {
		if !fi.Mode().IsDir() {
			return p, nil
		}
		err = fmt.Errorf("%s is directory", p)
	}
	if filepath.Ext(p) == "" {
		p += ".exe"
		fi, err := os.Stat(p)
		if err == nil {
			if !fi.Mode().IsDir() {
				return p, nil
			}
			err = fmt.Errorf("%s is directory", p)
		}
	}
	return "", err
}

// StartService starts the windows service targeted by name
func StartService(name string) error {

	m, err := mgr.Connect()
	if err != nil {
		log.Errorln(fmt.Sprintf("could not connect the service control manager error : %s", err.Error()))
		return err
	}
	defer m.Disconnect()
	s, err := m.OpenService(name)
	if err != nil {
		log.Errorln(fmt.Sprintf("could not access service (OpenService): %s", name))
		return fmt.Errorf("could not access service: %v", err)
	}
	defer s.Close()
	err = s.Start("is", "manual-started")
	if err != nil {
		log.Errorln(fmt.Sprintf("could not start service %s, error : %s", name, err.Error()))
		return fmt.Errorf("could not start service: %v", err)
	}
	return nil
}

// ControlService controls the service targetede by name
func ControlService(name string, c svc.Cmd, to svc.State) error {

	m, err := mgr.Connect()
	if err != nil {
		log.Errorln(fmt.Sprintf("could not connect the service %s, error : %s", name, err.Error()))
		return err
	}
	defer m.Disconnect()
	s, err := m.OpenService(name)
	if err != nil {
		log.Errorln(fmt.Sprintf("could not access service (OpenService): %s", name))
		return fmt.Errorf("could not access service: %v", err)
	}
	defer s.Close()
	status, err := s.Control(c)
	if err != nil {
		log.Errorln(fmt.Sprintf("could not send control=%d: %s", c, err.Error()))
		return fmt.Errorf("could not send control=%d: %v", c, err)
	}
	timeout := time.Now().Add(10 * time.Second)
	for status.State != to {
		if timeout.Before(time.Now()) {
			log.Errorln(fmt.Sprintf("timeout waiting for service to go to state=%d", to))
			return fmt.Errorf("timeout waiting for service to go to state=%d", to)
		}
		time.Sleep(300 * time.Millisecond)
		status, err = s.Query()
		if err != nil {
			log.Errorln(fmt.Sprintf("could not send control=%d: %s", c, err.Error()))
			return fmt.Errorf("could not retrieve service status: %v", err)
		}
	}
	return nil
}

// InstallService installs the service targeted by name
func InstallService(name, desc string) error {
	var errormsg string

	exepath, err := exePath()
	if err != nil {
		errormsg = err.Error()
		log.Errorln(errormsg)
		return err
	}
	m, err := mgr.Connect()
	if err != nil {
		errormsg = err.Error()
		log.Errorln(errormsg)
		return err
	}
	defer m.Disconnect()
	s, err := m.OpenService(name)
	if err == nil {
		s.Close()
		errormsg = fmt.Sprintf("service %s already exists", name)
		log.Errorln(errormsg)
		return fmt.Errorf(errormsg)
	}
	s, err = m.CreateService(name, exepath, mgr.Config{DisplayName: desc}, "is", "auto-started")
	if err != nil {
		errormsg = err.Error()
		log.Errorln(errormsg)
		return err
	}
	defer s.Close()
	err = eventlog.InstallAsEventCreate(name, eventlog.Error|eventlog.Warning|eventlog.Info)
	if err != nil {
		if strings.Contains(err.Error(), "ezb_vault registry key already exists") {
			log.Warnln("Event registry key already exists, skipping warn and continue")
			return nil
		}
		s.Delete()
		errormsg = fmt.Sprintf("SetupEventLogSource() failed: %s", err)
		log.Errorln(errormsg)
		return fmt.Errorf(errormsg)
	}
	return nil
}

// RemoveService remove the service trageted by name
func RemoveService(name string) error {
	var errormsg string

	m, err := mgr.Connect()
	if err != nil {
		errormsg = err.Error()
		log.Errorln(errormsg)
		return err
	}
	defer m.Disconnect()
	s, err := m.OpenService(name)
	if err != nil {
		errormsg = fmt.Sprintf("service %s is not installed", name)
		log.Errorln(errormsg)
		return fmt.Errorf(errormsg)
	}
	defer s.Close()
	err = s.Delete()
	if err != nil {
		log.Errorln(err.Error())
		return err
	}
	err = eventlog.Remove(name)
	if err != nil {
		errormsg = fmt.Sprintf("RemoveEventLogSource() failed: %s", err)
		log.Errorln(errormsg)
		return fmt.Errorf(errormsg)
	}
	return nil
}
