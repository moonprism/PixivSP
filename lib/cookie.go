package lib

import (
	"encoding/gob"
	"net/http"
	"net/url"
	"os"
)

const FileName  = "cookie_"

// SaveCookies save cookie info to local file
func SaveCookies(u *url.URL, cookies []*http.Cookie) error {
	fileName := RuntimeConf.SaveFilePath + FileName + u.Host

	_ = os.Remove(fileName)

	file, err := os.Create(fileName)
	if err != nil {
		return err
	}

	defer file.Close()

	encoder := gob.NewEncoder(file)

	return encoder.Encode(cookies)
}

// LoadCookies set cookieJar from file
func LoadCookies(u *url.URL) (cookies []*http.Cookie, err error) {
	fileName := RuntimeConf.SaveFilePath + FileName + u.Host

	file, err := os.Open(fileName)
	if err != nil {
		return
	}

	defer file.Close()

	decoder := gob.NewDecoder(file)
	err = decoder.Decode(&cookies)

	return
}