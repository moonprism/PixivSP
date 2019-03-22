package lib

import (
	"github.com/go-ini/ini"
	"log"
)

func init() {
	config, err := ini.Load("conf/app.ini")

	if err != nil {
		log.Fatalf("setting, conf/app.ini parse failed: %v", err)
	}
}