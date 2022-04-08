package traffic_logger

import (
	"github.com/gazercloud/gazer_repeater/logger"
	"os"
	"path/filepath"
	"strings"
	"time"
)

func CurrentExePath() string {
	dir, _ := filepath.Abs(filepath.Dir(os.Args[0]))
	return dir
}

func AppendStringToLen(text string, l int) string {
	needToAdd := l - len(text)
	if needToAdd > 0 {
		return text + strings.Repeat(" ", needToAdd)
	}
	return text
}

func Write(remoteAddr string, urlString string, comment string) {
	directory := CurrentExePath() + "/traffic"
	fileName := time.Now().Format("2006-01-02") + ".log"
	fullPath := directory + "/" + fileName

	if _, err := os.Stat(directory); os.IsNotExist(err) {
		err := os.MkdirAll(directory, 0777)
		if err != nil {
			logger.Println("Can not create log directory")
		}
	}

	f, err := os.OpenFile(fullPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
	if err == nil {

		timeString := time.Now().UTC().Format("2006-01-02 15:04:05.999")
		line := AppendStringToLen(timeString, 23) + " | " + AppendStringToLen(remoteAddr, 15) + " | " + AppendStringToLen(urlString, 40) + " " + comment + "\r\n"

		if _, err := f.WriteString(line); err != nil {
			logger.Println("traffic_logger WriteString error", err)
		}
		err = f.Close()
		if err != nil {
			logger.Println("traffic_logger Close error", err)
		}
	} else {
		logger.Println("traffic_logger OpenFile error", err)
	}

}
