package main

import (
	"encoding/json"
	"fmt"
	"github.com/davecgh/go-spew/spew"
	"github.com/moonprism/PixivSP/lib"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"io/ioutil"
	"net/url"
	"os"
	"sort"
)

const (
	// pixiv login form
	PixivLoginLink   = "https://accounts.pixiv.net/login?lang=zh&source=pc&view_type=page&ref=wwwtop_accounts_index"
	PixivLoginToLink = "https://accounts.pixiv.net/api/login?lang=zh"
	// bookmark format link
	PixivBookmarkLink = "https://www.pixiv.net/bookmark.php?rest=show&p=%d"
	// illust page link format
	PixivIllustLink = "https://www.pixiv.net/member_illust.php?mode=medium&illust_id=%s"
)

func login(c *lib.HttpClient) (err error) {
	// request login page
	htmlDoc, err := c.GetResponseDoc(PixivLoginLink)
	if err != nil {
		return
	}

	// find post key
	postKey, isSetKey := htmlDoc.Find("input[name='post_key']").First().Attr("value")
	if isSetKey != true {
		return errors.New("parse login post_key failed")
	}

	type PixivLoginResponse struct {
		Error   bool
		Message string
		Body    map[string]interface{}
	}
	var response PixivLoginResponse

	// login
	resp, err := c.PostForm(PixivLoginToLink, url.Values{
		"pixiv_id": {lib.PixivConf.PixivUser},
		"password": {lib.PixivConf.PixivPassword},
		"post_key": {postKey},
	})
	if err != nil {
		return lib.NewError(fmt.Sprintf("post login form failed: %v", err))
	}

	defer func() {
		err = resp.Body.Close()
	}()

	body, _ := ioutil.ReadAll(resp.Body)

	// is login success
	if body != nil {

		if err = json.Unmarshal(body, &response); err != nil {
			log.Errorf("login failed parse response json: %v", err)
			return
		}

		// dump login error info
		if response.Body["validation_errors"] != nil {
			log.Errorf("login failed %v", response.Body["validation_errors"])
			return
		}
	}
	return
}

func checkLogin(c *lib.HttpClient) {
	// check cookie file exists
	indexUrl, _ := url.Parse("https://www.pixiv.net")
	cookie, err := lib.LoadCookies(indexUrl)
	if os.IsNotExist(err) {
		// generate cookie
		if err = login(c); err != nil {
			log.Errorln(err.Error())
			return
		}
		lib.SaveCookies(indexUrl, c.Jar.Cookies(indexUrl))
	}
	// check cookie expire

}

func main() {

	if err := lib.CheckSavePath(lib.RuntimeConf.SavePath); err != nil {
		log.Fatalf("init %v", err)
	}

	lib.InitLogrus()

	client := lib.NewHttpClient()

	if lib.ProxyConf.ProxyOn {
		if err := client.SetProxy(lib.ProxyConf.ProxyHost, lib.ProxyConf.ProxyPort); err != nil {
			log.Fatalf("set proxy %v", err)
			return
		}
	}


		err := p.Login(lib.PixivConf.PixivUser, lib.PixivConf.PixivPassword)
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

	var illustQuantity, page int

	for {
		page++
		num, nextPage, err := p.ParseBookmark(page)
		if err != nil {
			log.Fatalf("parse bookmark error: %v", err)
		}
		if nextPage == 0  {
			break
		}

		illustQuantity += num
		log.WithFields(log.Fields{
			"num": num,
			"nextPage": nextPage,
		}).Info("parse next page")
	}

	var percentages = make(map[string]int)

	for {
		select {
		case i := <-p.ProcessChan:
			if i.Error == nil {
				illustQuantity--
			} else if i.Times <= 3 {
				log.WithFields(log.Fields{
					"times": i.Times,
					"illust": i.ID,
				}).Warningf("download image failed: %v", i.Error)
				go p.ParseIllust(i)
			} else {
				illustQuantity--
				log.WithFields(log.Fields{
					"illust": i.ID,
				}).Errorf("download image failed: %v", i.Error)
			}
			if illustQuantity == 0 {
				log.Info("over")
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
