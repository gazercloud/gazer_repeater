package main

import (
	"http-server.org/gazer/app"
	"http-server.org/gazer/application"
	"http-server.org/gazer/logger"
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
