
package main

import (
    "fmt"
    "net"
    "net/http"
    "net/http/cookiejar"
    "os"
    "time"

    "github.com/PuerkitoBio/goquery"
    "golang.org/x/net/proxy"
)

// setTransport use socket5 proxy
func setTransport(client *http.Client, addr string) error {

    dialer, err := proxy.SOCKS5("tcp", addr,
        nil,
        &net.Dialer{
            Timeout:   30 * time.Second,
            KeepAlive: 30 * time.Second,
        },
    )
    if err != nil {
        return err
    }

    client.Transport = &http.Transport{
        Proxy:               nil,
        Dial:                dialer.Dial,
        TLSHandshakeTimeout: 10 * time.Second,
    }

    return nil
}

func main() {

    cookieJar, _ := cookiejar.New(nil)

    client := &http.Client{
        Jar: cookieJar,
    }

    if err := setTransport(client, "127.0.0.1:1080"); err != nil {
        fmt.Println(err)
        os.Exit(1)
    }

    // request login page
    resp, err := client.Get("https://accounts.pixiv.net/login?lang=zh&source=pc&view_type=page&ref=wwwtop_accounts_index")
    if err != nil {
        fmt.Println(err)
        os.Exit(1)
    }

    htmlDoc, err := goquery.NewDocumentFromResponse(resp)

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