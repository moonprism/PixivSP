package main

import (
	"github.com/davecgh/go-spew/spew"
	"github.com/moonprism/PixivSP/lib"
	"os"
)

func main() {
	p := NewPixiv(lib.PixivConf.PixivUser, lib.PixivConf.PixivPasswd)

	p.SetProxy(lib.ProxyConf.ProxyHost, lib.ProxyConf.ProxyPort)

	cookies, err := lib.LoadCookies(p.IndexUrl)
	if os.IsNotExist(err) {
		p.Login()
	}

	p.SetCookies(cookies)
	p.ParseBookmark(1)

	for {
		select {
		case i := <- p.ProcessChan:
			if i.Link != "" && i.Error == nil {
				// make struct
			}
			spew.Dump(i)
		}
	}
	
}