package tools

import (
	"github.com/go-ini/ini"
	"log"
)

var config *ini.File

var PixivConf = &struct{
	PixivUser	string
	PixivPasswd string
}{}

var ProxyConf = &struct{
	ProxyOn		bool
	ProxyHost	string
	ProxyPort	string
}{}

var RuntimeConf = &struct{
	SaveFilePath	string
}{}

func init() {
	var err error
	config, err = ini.Load("conf/app.ini")
	if err != nil {
		log.Fatalf("setting, conf/app.ini parse failed: %v", err)
	}

	MapTo("pixiv", PixivConf)
	MapTo("proxy", ProxyConf)
	MapTo("runtime", RuntimeConf)
}

func MapTo(section string, v interface{}) {
	if err := config.Section(section).MapTo(v); err != nil {
		log.Fatalf("setting, conf/app.ini parse failed: %v", err)
	}
}