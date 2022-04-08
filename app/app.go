package app

import (
	"fmt"
	"github.com/gazercloud/gazer_repeater/credentials"
	"github.com/gazercloud/gazer_repeater/logger"
	"github.com/gazercloud/gazer_repeater/srv_public/public"
	"github.com/gazercloud/gazer_repeater/srv_repeater"
	"github.com/gazercloud/gazer_repeater/starter"
	"github.com/gazercloud/gazer_repeater/state"
	"github.com/gazercloud/gazer_repeater/traffic_control"
	"time"
)

type ISystem interface {
	Start()
	Stop()
	State() *state.System
}

var system ISystem

func Start() {
	logger.Println("")
	logger.Println("")
	logger.Println("")
	logger.Println("Application Started")
	logger.Println("")
	logger.Println("")
	logger.Println("")

	TuneFDs()
	traffic_control.Start()

	st := starter.NewHttpServer()

	st.Start()
	for st.Started {
		time.Sleep(100 * time.Millisecond)
	}

	logger.Println("[app]", "Role:", credentials.ServerRole)

	if credentials.ServerRole == "repeater" {
		system = srv_repeater.NewSrvRepeater()
	}

	if credentials.ServerRole == "public" {
		system = public.NewSrvPublic()
	}

	if system != nil {
		system.Start()
	} else {
		logger.Println("[app]", "UNKNOWN ROLE")
	}
}

func Stop() {
	traffic_control.Stop()
	if system != nil {
		system.Stop()
	}
}

func RunDesktop() {
	logger.Println("[app]", "Running as console application")
	Start()
	fmt.Scanln()
	logger.Println("[app]", "Console application exit")
}

func RunAsService() error {
	Start()
	return nil
}

func StopService() {
	Stop()
}
