package main

import (
	"fmt"
	"github.com/davecgh/go-spew/spew"
	"github.com/moonprism/PixivSP/tools"
	log "github.com/sirupsen/logrus"
	"os"
	"sort"
)

func init() {
	if !tools.Exists(tools.RuntimeConf.IllustSavePath) {
		os.Mkdir(tools.RuntimeConf.IllustSavePath, 0777)
	}
}

func main() {
	p := NewPixiv()

	if tools.ProxyConf.ProxyOn {
		if err := p.SetProxy(tools.ProxyConf.ProxyHost, tools.ProxyConf.ProxyPort); err != nil {
			spew.Sdump(err)
			return
		}
	}

	cookies, err := tools.LoadCookies(p.IndexUrl)
	if os.IsNotExist(err) {
		err := p.Login(tools.PixivConf.PixivUser, tools.PixivConf.PixivPasswd)
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
	p.SetSavePath(tools.RuntimeConf.IllustSavePath)

	p.ParseBookmark(1)

	var percentages = make(map[string]int)

	for {
		select {
		case i := <-p.ProcessChan:
			if i.Error == nil {

			} else if i.Times < 3 {
				log.Warningf("download image failed: %v", i.Error)
				go p.ParseIllust(i)
			}
			break
		case s := <-p.ProgressChan:
			percentages[s.ID] = s.Percentage

			var ids []string
			for id := range percentages {
				ids = append(ids, id)
			}
			sort.Strings(ids)
			for _, id := range ids {
				fmt.Printf("%s %02d%%\n", id, percentages[id])
			}
			fmt.Printf("\033[%dA\033[K", len(ids))
		}
	}

}
