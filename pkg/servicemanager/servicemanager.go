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
	"ezBastion/pkg/setupmanager"
	"fmt"
	"golang.org/x/sys/windows/svc/debug"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
	"golang.org/x/sys/windows/svc"
	"golang.org/x/sys/windows/svc/eventlog"
	"golang.org/x/sys/windows/svc/mgr"
)

var (
	err  error
	elog debug.Log
	ms   MainService
)

type MyService struct{}

type MainService interface {
	StartMainService(serverchan *chan bool)
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
	err = s.Start("is", "auto-started")
	if err != nil {
		log.Errorln(fmt.Sprintf("could not start service %s, error : %s", name, err.Error()))
		return fmt.Errorf("could not start service: %v", err)
	}
	return nil
}

// ControlService controls the service targeted by name
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
func InstallService(name, desc, exePath string) error {
	var errormsg string
	exeFullPath, err := setupmanager.ExeFullPath()
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
	s, err = m.CreateService(name, exeFullPath, mgr.Config{DisplayName: desc}, "is", "auto-started")
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

func (m *MyService) Execute(args []string, r <-chan svc.ChangeRequest, changes chan<- svc.Status) (ssec bool, errno uint32) {
	const cmdsAccepted = svc.AcceptStop | svc.AcceptShutdown
	changes <- svc.Status{State: svc.StartPending}

	changes <- svc.Status{State: svc.Running, Accepts: cmdsAccepted}
	serverchan := make(chan bool)
	go ms.StartMainService(&serverchan)
loop:
	for {
		select {
		case c := <-r:
			switch c.Cmd {
			case svc.Interrogate:
				elog.Info(1, "Interrogate")
				changes <- c.CurrentStatus

				time.Sleep(100 * time.Millisecond)
				changes <- c.CurrentStatus
			case svc.Stop, svc.Shutdown:

				close(serverchan)
				break loop
			default:
				elog.Error(1, fmt.Sprintf("unexpected control request #%d", c))
			}
		}
	}
	changes <- svc.Status{State: svc.StopPending}
	return
}

func RunService(name string, isDebug bool, MS MainService) {
	var err error
	if isDebug {
		elog = debug.New(name)
	} else {
		elog, err = eventlog.Open(name)
		if err != nil {
			return
		}
	}
	defer elog.Close()

	elog.Info(1, fmt.Sprintf("starting %s service", name))
	run := svc.Run
	if isDebug {
		run = debug.Run
	}
	ms = MS
	err = run(name, &MyService{})
	if err != nil {
		elog.Error(1, fmt.Sprintf("%s service failed: %v", name, err))
		return
	}
	elog.Info(1, fmt.Sprintf("%s service stopped", name))
}
