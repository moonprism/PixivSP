package main

import (
    "encoding/json"
    "errors"
    "fmt"
    "github.com/PuerkitoBio/goquery"
    "github.com/moonprism/PixivSP/lib"
    "io/ioutil"
    "log"
    "net/http"
    "net/http/cookiejar"
    "net/url"
    "strconv"
)

const (
    // pixiv login form
    PixivLoginLink  = "https://accounts.pixiv.net/login?lang=zh&source=pc&view_type=page&ref=wwwtop_accounts_index"
    PixivLoginToLink  = "https://accounts.pixiv.net/api/login?lang=zh"
    // bookmark format link
    PixivBookmarkLink  = "https://www.pixiv.net/bookmark.php?rest=show&p=%d"
)

type Pixiv struct {
    // user info
    UserID      string
    password    string

    Client      *http.Client
    IndexUrl    *url.URL

    ProcessChan chan *Illust
}

type Illust struct{
    // illust info
    ID      string
    Name    string
    Like    int
    Tags     string

    // author info
    MemberId    string
    MemberName  string

    // generate link
    Link    string
    // processerror info
    Error   error
}

func NewPixiv(id string, passwd string) (p *Pixiv) {

    p = &Pixiv{
        UserID: id,
        password: passwd,
        ProcessChan: make(chan *Illust, 10),
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

func (p *Pixiv) SetProxy(host string, port string) {

    if err := lib.SetTransport(p.Client, host+":"+port); err != nil {
        log.Fatalf("proxy - fatal: %v", err)
    }
}

func (p *Pixiv) requestGet(link string) (doc *goquery.Document, err error) {
    resp, err := p.Client.Get(link)
    if err != nil {
        log.Printf("request url %s failed: %v", link, err)
    }

    // parse html doc
    doc, err = goquery.NewDocumentFromReader(resp.Body)
    if err != nil {
        log.Printf("parse html %s failed: %v", link, err)
    }
    return
}

func (p *Pixiv) ParseBookmark(page int) (err error) {

    link := fmt.Sprintf(PixivBookmarkLink, page)
    htmlDoc, err := p.requestGet(link)
    if err != nil {
        return
    }

    if userName := htmlDoc.Find("a.user-name").Text(); userName == "" {
        err = errors.New("login failed maybe ~")
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

func (p *Pixiv) ParseIllust(ill *Illust) {

    // todo process
    p.ProcessChan <- ill
}

func (p *Pixiv) Login() (err error) {
    // request login page
    htmlDoc, err := p.requestGet(PixivLoginLink)
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

    // login
    resp, err := p.Client.PostForm(PixivLoginToLink, url.Values{
        "pixiv_id" : {lib.PixivConf.PixivUser},
        "password" : {lib.PixivConf.PixivPasswd},
        "post_key" : {postKey},
    })

    if err != nil {
        log.Printf("login - post login form failed: %v", err)
        return
    }

    body, _ := ioutil.ReadAll(resp.Body)

    // is login success
    if body != nil {
        var respData struct {
            Error string
            Message string
            Body map[string]interface{}
        }

        if err = json.Unmarshal(body, &respData); err != nil {
            log.Printf("login - parse response date failed: %v", err)
            return
        }

        // dump login error info
        if respData.Body["validation_errors"] != nil {
            err = errors.New("login failed")
            log.Printf("%v: %v", err, respData.Body["validation_errors"])
            return
        }
    }
    return
}