package lib

import (
	"os"
	log "github.com/sirupsen/logrus"
)

func Exists(path string) bool {
	_, err := os.Stat(path)
	if err != nil {
		if os.IsExist(err) {
			return true
		}
		return false
	}
	return true
}

func checkSavePath(path string) (err error) {
	if !Exists(path) {
		if err := os.Mkdir(path, 0755); err != nil {
			log.WithFields(log.Fields{
				path
			}).Fatalf("%v", NewError("path is un"))
		}
	}
}