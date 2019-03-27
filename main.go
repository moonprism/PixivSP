package main

import (
	"fmt"
	"github.com/davecgh/go-spew/spew"
	"github.com/moonprism/PixivSP/tools"
	"os"
	"sort"
	"strings"
)

func main() {
	p := NewPixiv(tools.PixivConf.PixivUser, tools.PixivConf.PixivPasswd)

	if tools.ProxyConf.ProxyOn {
		if err := p.SetProxy(tools.ProxyConf.ProxyHost, tools.ProxyConf.ProxyPort); err != nil {
			spew.Sdump(err)
			return
		}
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

	var percentages = make(map[string]int)

	for {
		select {
		case i := <-p.ProcessChan:
			if i.Error == nil {
				// make struct
			} else if i.Times < 3 {
				//go p.ParseIllust(i)
			} else {
				// failed logs
			}
			break
		case s := <-p.SaveProgress:
			percentages[s.ID] = s.Percentage

			var ids []string
			for id := range percentages {
				ids = append(ids, id)
			}
			sort.Strings(ids)
			for _, id := range ids {
				fmt.Printf("%s.png %s%d%%\n", id, Bar(percentages[id], 30), percentages[id])
			}
			fmt.Printf("\033[%dA\033[K", len(ids))
		}
	}

}
// 这段测试代码网上找的，以后这个chan的数据直接写到websocket里
func Bar(vl int, width int) string {
	return fmt.Sprintf("%s%*c", strings.Repeat("█", vl/10), vl/10-width+1,
		([]rune(" ▏▎▍▌▋▋▊▉█"))[vl%10])
}
