package ez_cli

import (
	"ezBastion/pkg/servicemanager"
	"ezBastion/pkg/setupmanager"
	"github.com/urfave/cli"
	"golang.org/x/sys/windows/svc"
	"log"
)

func EZCli(SERVICENAME, SERVICEFULLNAME, exePath, confPath string, ms servicemanager.MainService) []cli.Command {

	ezcli := []cli.Command{
		{
			Name:  "init",
			Usage: "Generate config file and certificate.",
			Action: func(c *cli.Context) error {
				err := setupmanager.Setup(exePath, confPath, SERVICENAME)
				return err
			},
		}, {
			Name:  "debug",
			Usage: "Start in a console with extra log.",
			Action: func(c *cli.Context) error {
				servicemanager.RunService(SERVICENAME, true, ms)
				return nil
			},
		}, {
			Name:  "install",
			Usage: "Add windows service.",
			Action: func(c *cli.Context) error {
				// not for worker -> SERVICENAME
				err := servicemanager.InstallService(SERVICENAME, SERVICEFULLNAME, exePath)
				if err != nil {
					log.Fatalf("Install service: %v", err)
				}
				return err
			},
		}, {
			Name:  "remove",
			Usage: "Remove windows service.",
			Action: func(c *cli.Context) error {
				err := servicemanager.RemoveService(SERVICENAME)
				if err != nil {
					log.Fatalf("Remove service: %v", err)
				}
				return err
			},
		}, {
			Name:  "start",
			Usage: "Start windows service.",
			Action: func(c *cli.Context) error {
				err := servicemanager.StartService(SERVICENAME)
				if err != nil {
					log.Fatalf("start service: %v", err)
				}
				return err
			},
		}, {
			Name:  "stop",
			Usage: "Stop windows service.",
			Action: func(c *cli.Context) error {
				err := servicemanager.ControlService(SERVICENAME, svc.Stop, svc.Stopped)
				if err != nil {
					log.Fatalf("stop service: %v", err)
				}
				return err
			},
		},
	}
	return ezcli
}
