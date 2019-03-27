package main

import (
	"fmt"
	"github.com/davecgh/go-spew/spew"
	"github.com/moonprism/PixivSP/tools"
	"os"
	"strings"
)

func main() {
	p := NewPixiv(tools.PixivConf.PixivUser, tools.PixivConf.PixivPasswd)

	if tools.ProxyConf.ProxyOn {
		p.SetProxy(tools.ProxyConf.ProxyHost, tools.ProxyConf.ProxyPort)
	}

	cookies, err := tools.LoadCookies(p.IndexUrl)
	if os.IsNotExist(err) {
		err := p.Login()
		if err != nil {
			spew.Dump(err)
			return
		}
		err = tools.SaveCookies(p.IndexUrl, p.GetCookies())
		if err != nil {
			spew.Dump(err)
		}
	}

	p.SetCookies(cookies)
	p.SetSavePath(tools.RuntimeConf.SaveFilePath)

	p.ParseBookmark(1)

	for {
		select {
		case i := <-p.ProcessChan:
			if i.Link != "" && i.Error == nil {
				// make struct
			} else if i.Times < 3 {
				//go p.ParseIllust(i)
			} else {
				// failed logs
			}
			spew.Dump(i)
			break
		case s := <-p.SaveProgress:
			fmt.Printf("\r%s", strings.Repeat(" ", 35))
			fmt.Printf("\rDownloading... %d%%", s.Percentage)
		}
	}

}
