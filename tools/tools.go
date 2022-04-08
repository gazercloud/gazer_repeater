package tools

import (
	"log"
	"os"
)

func setLogFile() {
	f, err := os.OpenFile("c:/Private/system/log.log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
	}
	log.SetOutput(f)
}
