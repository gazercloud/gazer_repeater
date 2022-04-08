package main

import (
	"github.com/gazercloud/gazer_repeater/app"
	"github.com/gazercloud/gazer_repeater/application"
	"github.com/gazercloud/gazer_repeater/logger"
)

func main() {
	application.Name = "gazer_web"
	application.ServiceName = "gazer_web"
	application.ServiceDisplayName = "gazer_web"
	application.ServiceDescription = "gazer_web"
	application.ServiceRunFunc = app.RunAsService
	application.ServiceStopFunc = app.StopService

	logger.Init(logger.CurrentExePath() + "/logs")

	if !application.TryService() {
		app.RunDesktop()
	}
}
