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

	if err = encoder.Encode(cookies); err != nil {
		return err
	}

	return nil
}

// LoadCookies set cookieJar from file
func LoadCookies(u *url.URL) (cookies []*http.Cookie, err error) {
	fileName := RuntimeConf.SaveFilePath + FileName + u.Host

	file, err := os.Open(fileName)
	defer file.Close()

	if err != nil {
		return
	}

	decoder := gob.NewDecoder(file)

	if err := decoder.Decode(&cookies); err != nil {
		return
	}

	return
}