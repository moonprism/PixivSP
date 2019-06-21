package lib

import (
	"log"

	"github.com/go-ini/ini"
)

var config *ini.File

var PixivConf = &struct {
	PixivUser	string
	PixivPassword	string
}{}

var ProxyConf = &struct {
	ProxyOn	bool
	ProxyHost	string
	ProxyPort	string
}{}

var RuntimeConf = &struct {
	SavePath	string
	Mode	string
}{}

func init() {
	var err error
	config, err = ini.Load("config/app.ini")
	if err != nil {
		log.Fatalf("setting, config/app.ini parse failed: %v", err)
	}

	MapTo("pixiv", PixivConf)
	MapTo("proxy", ProxyConf)
	MapTo("runtime", RuntimeConf)
}

func MapTo(section string, v interface{}) {
	if err := config.Section(section).MapTo(v); err != nil {
		log.Fatalf("setting, config/app.ini parse failed: %v", err)
	}
}
