package main

import (
    "encoding/json"
    "fmt"
    "github.com/PuerkitoBio/goquery"
    "github.com/moonprism/PixivSP/lib"
    "io/ioutil"
    "log"
    "net/http"
    "net/http/cookiejar"
    "net/url"
)

const PixivLoginUrl  = "https://accounts.pixiv.net/login?lang=zh&source=pc&view_type=page&ref=wwwtop_accounts_index"
const PixivLoginToUrl  = "https://accounts.pixiv.net/api/login?lang=zh"
const PixivBookmark  = "https://www.pixiv.net/bookmark.php"

type LoginResponse struct {
    Error string
    Message string
    Body map[string]interface{}
}

var cookieUrl *url.URL

func initial() {
    cookieUrl, _ = url.Parse(PixivLoginUrl)
}

func main() {

    initial()

    jar, _ := cookiejar.New(nil)

    client := &http.Client{
        Jar: jar,
    }

    if  lib.ProxyConf.ProxyOn == true {
        if err := lib.SetTransport(client, lib.ProxyConf.ProxyHost+":"+lib.ProxyConf.ProxyPort); err != nil {
            fmt.Println(err)
            return
        }
    }

    //cookies, err := lib.LoadCookies(cookieUrl)

    //if err != nil {
    //    login(client)
    //} else {
    //    client.Jar.SetCookies(cookieUrl, cookies)
    //}

    login(client)

    resp, err := client.Get(PixivBookmark)

    if err != nil {
        log.Fatalf("get bookmark failed: %v", err)
    }

    htmlDoc, err := goquery.NewDocumentFromReader(resp.Body)

    if err != nil {
        fmt.Println(err)
        return
    }

    userName := htmlDoc.Find("a.user-name").Text()

   if userName == "" {

   } else {
       println(userName)
   }
}

func login(client *http.Client) {
    // request login page
    resp, err := client.Get(PixivLoginUrl)
    if err != nil {
        fmt.Println(err)
        return
    }

    htmlDoc, err := goquery.NewDocumentFromReader(resp.Body)

    if err != nil {
        fmt.Println(err)
        return
    }

    postKey, isSetKey := htmlDoc.Find("input[name='post_key']").First().Attr("value")

    if isSetKey != true {
        // not found login form post key

        return
    }
    // login
    resp, err = client.PostForm(PixivLoginToUrl, url.Values{
        "pixiv_id" : {lib.PixivConf.PixivUser},
        "password" : {lib.PixivConf.PixivPasswd},
        "post_key" : {postKey},
    })

    if err != nil {
        log.Fatalf("post login form failed: %v", err)
    }

    body, _ := ioutil.ReadAll(resp.Body)

    // login status
    if string(body) != "" {
        var respData LoginResponse
        _ = json.Unmarshal(body, &respData)

        if respData.Body["validation_errors"] != nil {
            log.Fatalf("login error : %v", respData.Body["validation_errors"])
        }
    }

    lib.SaveCookies(cookieUrl, client.Jar.Cookies(cookieUrl))
}