package tools

import (
	"encoding/gob"
	"net/http"
	"net/url"
	"os"
)

const suffix = ".cookie"

// SaveCookies save cookie info to local file
func SaveCookies(u *url.URL, cookies []*http.Cookie) error {
	fileName := RuntimeConf.CookieSavePath + "/" + u.Host + suffix

	_ = os.Remove(fileName)

	file, err := os.Create(fileName)
	if err != nil {
		return err
	}

	defer func () {
		err = file.Close()
	}()

	encoder := gob.NewEncoder(file)

	return encoder.Encode(cookies)
}

// LoadCookies set cookieJar from file
func LoadCookies(u *url.URL) (cookies []*http.Cookie, err error) {
	fileName := RuntimeConf.CookieSavePath + "/" + u.Host + suffix

	file, err := os.Open(fileName)
	if err != nil {
		return
	}

	defer func() {
		err = file.Close()
	}()

	decoder := gob.NewDecoder(file)
	err = decoder.Decode(&cookies)

	return
}
