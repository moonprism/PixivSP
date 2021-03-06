package pixiv

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/moonprism/PixivSP/lib"
	"golang.org/x/net/proxy"
	log "github.com/sirupsen/logrus"
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

var (
	ErrLoginFailed = errors.New("login failed")
)

type Pixiv struct {
	// user info
	UserID   int64

	Client   *http.Client
	IndexUrl *url.URL

	ProcessChan  chan *Illust
	ProgressChan chan *IllustSaveProgress
}

func New(userID int64) (p *Pixiv) {
	p = &Pixiv{
		UserID:	userID,
		ProcessChan:	make(chan *Illust, 30),
		ProgressChan:	make(chan *IllustSaveProgress, 100),
	}

	jar, _ := cookiejar.New(nil)

	client := &http.Client{
		Jar: jar,
	}

	p.Client = client
	p.IndexUrl, _ = url.Parse("https://www.pixiv.net")
	return
}

func (p *Pixiv) SetCookies(cookies []*http.Cookie) {
	p.Client.Jar.SetCookies(p.IndexUrl, cookies)
}

func (p *Pixiv) GetCookies() []*http.Cookie {
	return p.Client.Jar.Cookies(p.IndexUrl)
}

func (p *Pixiv) SetSavePath(path string) {
	p.SavePath = path
}

func (p *Pixiv) SetProxy(host string, port string) (err error) {
	dialer, err := proxy.SOCKS5("tcp", host+":"+port,
		nil,
		&net.Dialer{
			Timeout:   30 * time.Second,
			KeepAlive: 30 * time.Second,
		},
	)
	if err != nil {
		return
	}

	p.Client.Transport = &http.Transport{
		Proxy:               nil,
		Dial:                dialer.Dial,
		TLSHandshakeTimeout: 10 * time.Second,
	}
	return
}

func (p *Pixiv) GetResponseDoc(link string) (doc *goquery.Document, err error) {
	resp, err := p.Client.Get(link)
	if err != nil {
		log.Printf("request url %s failed: %v", link, err)
		return
	}

	defer func() {
		err = resp.Body.Close()
	}()

	// parse html doc
	doc, err = goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		log.Warningf("parse html %s failed: %v", link, err)
	}
	return
}

func (p *Pixiv) GetResponseJson(link string) (err error) {
	resp, err := p.Client.Get(link)
	if err != nil {
		return
	}

	defer func() {
		err = resp.Body.Close()
	}()


}

func (p *Pixiv) ParseBookmark(page int) (illustQuantity int, nextPage int, err error) {
	link := fmt.Sprintf(PixivBookmarkLink, page)
	htmlDoc, err := p.GetResponseDoc(link)
	if err != nil {
		return
	}

	if userName := htmlDoc.Find("a.user-name").Text(); userName == "" {
		err = errors.New("session expired maybe ~")
		return
	}

	htmlDoc.Find("li.image-item").Each(func(i int, selection *goquery.Selection) {
		ill := &Illust{
			Page: page,
		}

		imgDoc := selection.Find("img").First()

		illID, isSet := imgDoc.Attr("data-id")
		if !isSet {
			return
		}
		ill.ID = illID
		ill.Name = selection.Find("h1.title").First().Text()
		ill.Tags, _ = imgDoc.Attr("data-tags")

		authorDoc := selection.Find("a.user").First()

		ill.MemberId, _ = authorDoc.Attr("data-user_id")
		ill.MemberName, _ = authorDoc.Attr("data-user_name")

		count := selection.Find("a.bookmark-count").First().Text()
		ill.Like, _ = strconv.Atoi(count)

		illustQuantity++
		ill.Index = illustQuantity

		go p.ParseIllust(ill)
	})

	// search next page link
	if nextPageStr := htmlDoc.Find("ul.page-list li.current").Next().Find("a").First().Text(); nextPageStr != "" {
		nextPage, err = strconv.Atoi(nextPageStr)
		if err != nil {
			return
		}
	}

	return
}

func (p *Pixiv) ParseIllust(i *Illust) {
	i.Error = p.DownloadIllust(i)
	i.Times++
	p.ProcessChan <- i
}

// WriteCounter user for display write file progress bar
type WriteCounter struct {
	Total          uint64
	Size           uint64
	lastPercentage int
	*Pixiv
	*Illust
}

func (wc *WriteCounter) Write(p []byte) (n int, err error) {
	// todo 超时后的断点续传
	n = len(p)
	wc.Size += uint64(n)
	per := int(float32(wc.Size) / float32(wc.Total) * 100)
	if per > wc.lastPercentage {
		wc.Illust.CurrentProgress = &IllustSaveProgress{
			ID:         wc.Illust.ID,
			Percentage: per,
		}
		wc.Pixiv.ProgressChan <- wc.Illust.CurrentProgress
		wc.lastPercentage = per
	}
	return
}

func (p *Pixiv) parseIllustSrc(illustID string) (src string, pageURL string, err error) {
	pageURL = fmt.Sprintf(PixivIllustLink, illustID)
	doc, err := p.GetResponseDoc(pageURL)
	if err != nil {
		return
	}
	html, _ := doc.Html()
	imgRegexp := regexp.MustCompile(`,"original":"(.+?)"}`)
	imgSrcInfo := imgRegexp.FindStringSubmatch(html)
	if len(imgSrcInfo) < 2 {
		err = errors.New("no find image " + illustID)
		return
	}
	src = strings.Replace(imgSrcInfo[1], "\\", "", -1)
	return
}

func (p *Pixiv) DownloadIllust(i *Illust) (err error) {
	tmpName := p.SavePath + "/" + i.ID + ".tmp"
	fileName := p.SavePath + "/" + i.ID + ".png"

	if tools.Exists(fileName) {
		return
	}

	if tools.Exists(tmpName) {
		err = os.Remove(tmpName)
		if err != nil {
			return
		}
	}

	imgSrc, illustPageURL, err := p.parseIllustSrc(i.ID)
	if err != nil {
		return
	}

	req, err := http.NewRequest("GET", imgSrc, nil)
	if err != nil {
		return
	}
	req.Header.Set("Referer", illustPageURL)
	resp, err := p.Client.Do(req)
	if err != nil {
		return
	}
	defer func() {
		err = resp.Body.Close()
	}()

	fSize, err := strconv.ParseInt(resp.Header.Get("Content-Length"), 10, 32)
	if err != nil {
		return
	}

	file, err := os.Create(tmpName)
	if err != nil {
		return
	}
	defer func() {
		err = file.Close()
	}()

	counter := &WriteCounter{
		Total:  uint64(fSize),
		Pixiv:  p,
		Illust: i,
	}
	_, err = io.Copy(file, io.TeeReader(resp.Body, counter))
	if err != nil {
		return
	}

	err = os.Rename(tmpName, fileName)
	if err != nil {
		return
	}

	return
}

func (p *Pixiv) Login(id string, password string) (err error) {
	// request login page
	htmlDoc, err := p.GetResponseDoc(PixivLoginLink)
	if err != nil {
		return
	}

	// find post key
	postKey, isSetKey := htmlDoc.Find("input[name='post_key']").First().Attr("value")
	if isSetKey != true {
		err =
		log.Errorf("login - %v", err)
		return
	}

	type PixivLoginResponse struct {
		Error   bool
		Message string
		Body    map[string]interface{}
	}
	var response PixivLoginResponse

	// login
	resp, err := p.Client.PostForm(PixivLoginToLink, url.Values{
		"pixiv_id": {id},
		"password": {password},
		"post_key": {postKey},
	})
	if err != nil {
		log.Errorf("login - post login form failed: %v", err)
		return
	}

	defer func() {
		err = resp.Body.Close()
	}()

	body, _ := ioutil.ReadAll(resp.Body)

	// is login success
	if body != nil {

		if err = json.Unmarshal(body, &response); err != nil {
			log.Warningf("login - parse response date failed: %v", err)
			return
		}

		// dump login error info
		if response.Body["validation_errors"] != nil {
			err = errors.New("login failed")
			log.Warningf("%v: %v", err, response.Body["validation_errors"])
			return
		}
	}
	return
}
