package lib

import (
	log "github.com/sirupsen/logrus"
	"os"
)

func InitLogrus () {
	if err := setFile(RuntimeConf.SavePath + "/run.log"); err != nil {
		log.Fatalf("%v\n", err)
		return
	}

	// set log level
	switch RuntimeConf.Mode {
	case "debug":
		log.SetLevel(log.DebugLevel)
		break
	case "info":
		log.SetLevel(log.InfoLevel)
		break
	default:
		log.SetLevel(log.TraceLevel)
	}
}

func setFile(filePath string) (err error) {
	_, err = os.Stat(filePath)

	// create file if not exists
	if os.IsNotExist(err) {
		file, err := os.Create(filePath)
		if err != nil {
			return err
		}

		defer func() {
			err = file.Close()
		}()
	}

	file, err := os.OpenFile(filePath, os.O_APPEND|os.O_WRONLY, os.ModeAppend)
	if err != nil {
		return
	}

	log.SetOutput(file)
	return
}
