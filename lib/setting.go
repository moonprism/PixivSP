package lib

import (
	"github.com/go-ini/ini"
	"log"
)

var config *ini.File

func init() {
	var err error
	config, err = ini.Load("conf/app.ini")
	if err != nil {
		log.Fatalf("setting, conf/app.ini parse failed: %v", err)
	}
}

func MapTo(section string, v interface{}) {
	if err := config.Section(section).MapTo(v); err != nil {
		log.Fatalf("setting, conf/app.ini parse failed: %v", err)
	}
}