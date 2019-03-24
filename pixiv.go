package main

import (
    "fmt"
    "github.com/moonprism/PixivSP/lib"
    "net/http"
    "net/http/cookiejar"
    "os"

    "github.com/PuerkitoBio/goquery"
)

var pixiv = &struct{
    PixivUser string
    PixivPasswd string
}{}

var proxy = &struct{
    ProxyOn bool
    ProxyHost string
    ProxyPort string
}{}

func initial() {
    lib.MapTo("pixiv", pixiv)
    
    lib.MapTo("proxy", proxy)
}

func main() {

    initial()

    cookieJar, _ := cookiejar.New(nil)

    client := &http.Client{
        Jar: cookieJar,
    }

    if  proxy.ProxyOn == true {
        if err := lib.SetTransport(client, proxy.ProxyHost+":"+proxy.ProxyPort); err != nil {
            fmt.Println(err)
            os.Exit(1)
        }
    }


    // request login page
    resp, err := client.Get("https://accounts.pixiv.net/login?lang=zh&source=pc&view_type=page&ref=wwwtop_accounts_index")
    if err != nil {
        fmt.Println(err)
        os.Exit(1)
    }

    htmlDoc, err := goquery.NewDocumentFromReader(resp.Body)

    if err != nil {
        fmt.Println(err)
        os.Exit(1)
    }

    postKey, isSetKey := htmlDoc.Find("input[name='post_key']").First().Attr("value")

    if isSetKey != true {
        // not found login form post key
    }

    println(postKey)

}