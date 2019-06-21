package lib

import (
	"os"
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

func CheckSavePath(path string) (err error) {
	if !Exists(path) {
		err = os.Mkdir(path, 0755)
	}
	return
}