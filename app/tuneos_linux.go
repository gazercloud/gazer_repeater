package app

import (
	"http-server.org/gazer/logger"
	"syscall"
)

func TuneFDs() {
	logger.Println("[app]", "TimeFDs begin")
	var rLimit syscall.Rlimit
	err := syscall.Getrlimit(syscall.RLIMIT_NOFILE, &rLimit)
	if err != nil {
		logger.Println("[app]", "TuneFDs Getrlimit1 error: ", err)
	}
	logger.Println("[app]", "Current limits:", rLimit)
	rLimit.Max = 999999
	rLimit.Cur = 999999
	err = syscall.Setrlimit(syscall.RLIMIT_NOFILE, &rLimit)
	if err != nil {
		logger.Println("[app]", "TuneFDs Setrlimit error: ", err)
	}
	err = syscall.Getrlimit(syscall.RLIMIT_NOFILE, &rLimit)
	if err != nil {
		logger.Println("[app]", "TuneFDs Getrlimit2 error: ", err)
	}
	logger.Println("[app]", "TimeFDs end")
}
