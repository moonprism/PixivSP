package main

import (
    "encoding/json"
    "errors"
    "github.com/PuerkitoBio/goquery"
    "github.com/moonprism/PixivSP/lib"
    "io/ioutil"
    "log"
    "net/http"
    "net/http/cookiejar"
    "net/url"
)

const (
    PixivLoginLink  = "https://accounts.pixiv.net/login?lang=zh&source=pc&view_type=page&ref=wwwtop_accounts_index"
    PixivLoginToLink  = "https://accounts.pixiv.net/api/login?lang=zh"
    PixivBookmarkLink  = "https://www.pixiv.net/bookmark.php"
)

type Pixiv struct {
    UserID string
    password string
    Client *http.Client
    IndexUrl *url.URL
    ProcessChan chan *Illust
}

type Illust struct{
    ID string
    AuthorID string
    LikeNum int
    // generate link
    Link string
    // error info
    Error error
}

func NewPixiv(id string, passwd string) (p *Pixiv) {

    p = &Pixiv{
        UserID: id,
        password: passwd,
        ProcessChan: make(chan *Illust, 100),
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

func (p *Pixiv) SetProxy(host string, port int) {
    if  lib.ProxyConf.ProxyOn == true {
        if err := lib.SetTransport(p.Client, lib.ProxyConf.ProxyHost+":"+lib.ProxyConf.ProxyPort); err != nil {
            log.Fatalf("proxy - fatal: %v", err)
        }
    }
}

func (p *Pixiv) ParseBookmark(page int) (ills []Illust, err error) {

    for _, ill := range ills {
        go p.ParseIllust(&ill)
    }

    processNum := 0

    for ill := range p.ProcessChan {
        if ill.Link != "" && ill.Error == nil {
            processNum ++
        }
    }

    return
}

func (p *Pixiv) ParseIllust(ill *Illust) {

    // process over
    p.ProcessChan <- ill
}

func (p *Pixiv) Login() (err error) {
    // request login page
    resp, err := p.Client.Get(PixivLoginLink)
    if err != nil {
        log.Printf("login - get pixiv login url failed: %v", err)
        return
    }

    // parse html doc
    htmlDoc, err := goquery.NewDocumentFromReader(resp.Body)
    if err != nil {
        log.Printf("login - goquery parse failed: %v", err)
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
    resp, err = p.Client.PostForm(PixivLoginToLink, url.Values{
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