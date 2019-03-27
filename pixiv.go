package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"golang.org/x/net/proxy"
	"io"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"
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

type Pixiv struct {
	// user info
	UserID   string
	password string

	Client   *http.Client
	IndexUrl *url.URL

	ProcessChan		chan *Illust
	SavePath		string
	SaveProgress	chan *IllustSaveProgress
}

type IllustSaveProgress struct {
	ID			string	`json:"id"`
	Percentage	int		`json:"percentage"`
}

type Illust struct {
	// illust info
	ID   string
	Name string
	Like int
	Tags string

	// author info
	MemberId   string
	MemberName string

	// generate link
	Link string
	// processerror info
	Error error
	// in channel times
	Times int

	CurrentSaveProgress	*IllustSaveProgress
}

func NewPixiv(id string, passwd string) (p *Pixiv) {
	p = &Pixiv{
		UserID:      id,
		password:    passwd,
		ProcessChan: make(chan *Illust, 30),
		SaveProgress: make(chan *IllustSaveProgress, 100),
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
	}

	defer func() {
		resp.Body.Close()
		//if p := recover(); p != nil {
		//	doc = nil
		//	err = errors.New("http request panic")
		//}
	}()

	// parse html doc
	doc, err = goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		log.Printf("parse html %s failed: %v", link, err)
	}
	return
}

func (p *Pixiv) ParseBookmark(page int) (err error) {
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
		ill := &Illust{}

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

		go p.ParseIllust(ill)
	})
	return
}

func (p *Pixiv) ParseIllust(i *Illust) {
	i.Error = p.DownLoadIllust(i)
	i.Times++
	// todo process
	p.ProcessChan <- i
}

type WriteCounter struct {
	Total	uint64
	Size	uint64
	lastPercentage int
	*Pixiv
	*Illust
}

func (wc *WriteCounter) Write(p []byte) (n int, err error) {
	n = len(p)
	wc.Size += uint64(n)
	per := int(wc.Size / wc.Total * 100)
	if per > wc.lastPercentage {
		wc.Pixiv.SaveProgress <- &IllustSaveProgress{
			ID:			wc.Illust.ID,
			Percentage:	per,
		}
		wc.lastPercentage = per
	}
	return
}

func (p *Pixiv) DownLoadIllust(i *Illust) (err error) {
	fileName := p.SavePath+"/"+i.ID
	illustPageUrl := fmt.Sprintf(PixivIllustLink, i.ID)

	doc, err := p.GetResponseDoc(illustPageUrl)
	if err != nil {
		return
	}
	html, _ := doc.Html()
	imgRegexp := regexp.MustCompile(`,"original":"(.+?)"}`)
	imgSrcInfo := imgRegexp.FindStringSubmatch(html)
	if len(imgSrcInfo) < 2 {
		return errors.New("no find image "+i.ID)
	}
	imgSrc := strings.Replace(imgSrcInfo[1], "\\", "", -1)
	resp, err := p.Client.Get(imgSrc)
	if err != nil {
		return
	}
	defer resp.Body.Close()

	fSize, err := strconv.ParseInt(resp.Header.Get("Content-Length"), 10, 32)
	if err != nil {
		return
	}

	file, err := os.Create(fileName+".tmp")
	if err != nil {
		return
	}
	defer file.Close()

	counter := &WriteCounter{
		Total: uint64(fSize),
		Pixiv: p,
		Illust: i,
	}
	_, err = io.Copy(file, io.TeeReader(resp.Body, counter))
	if err != nil {
		return err
	}

	err = os.Rename(fileName+".tmp", fileName+".png")
	if err != nil {
		return err
	}

	return
}

func (p *Pixiv) Login() (err error) {
	// request login page
	htmlDoc, err := p.GetResponseDoc(PixivLoginLink)
	if err != nil {
		return
	}

	// find post key
	postKey, isSetKey := htmlDoc.Find("input[name='post_key']").First().Attr("value")
	if isSetKey != true {
		err = errors.New("not found post_key")
		log.Printf("login - %v", err)
		return
	}

	_, err = p.PostLoginForm(postKey)
	return
}

type PixivLoginResponse struct {
	Error   bool
	Message string
	Body    map[string]interface{}
}

func (p *Pixiv) PostLoginForm(postKey string) (response *PixivLoginResponse, err error) {
	// login
	resp, err := p.Client.PostForm(PixivLoginToLink, url.Values{
		"pixiv_id": {p.UserID},
		"password": {p.password},
		"post_key": {postKey},
	})
	if err != nil {
		log.Printf("login - post login form failed: %v", err)
		return
	}

	defer func() {
		resp.Body.Close()
		//if p := recover(); p != nil {
		//	response = nil
		//	err = errors.New("http request panic")
		//}
	}()

	body, _ := ioutil.ReadAll(resp.Body)

	// is login success
	if body != nil {

		if err = json.Unmarshal(body, &response); err != nil {
			log.Printf("login - parse response date failed: %v", err)
			return
		}

		// dump login error info
		if response.Body["validation_errors"] != nil {
			err = errors.New("login failed")
			log.Printf("%v: %v", err, response.Body["validation_errors"])
			return
		}
	}
	return
}