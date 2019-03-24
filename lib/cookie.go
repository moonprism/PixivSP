package lib

import (
	"encoding/gob"
	"log"
	"net/http"
	"net/url"
	"os"
)

const FileName  = "cookie_"

// SaveCookies save cookie info to local file
func SaveCookies(u *url.URL, cookies []*http.Cookie) {
	fileName := RuntimeConf.SaveFilePath + FileName + u.Host

	_ = os.Remove(fileName)

	file, err := os.Create(fileName)
	if err != nil {
		log.Fatalf("create file %s failed: %v", fileName, err)
	}

	defer file.Close()

	encoder := gob.NewEncoder(file)

	if err = encoder.Encode(cookies); err != nil {
		log.Fatalf("ecode cookies failed: %v", err)
	}
}

// LoadCookies set cookieJar from file
func LoadCookies(u *url.URL) (coo []*http.Cookie, err error) {
	fileName := RuntimeConf.SaveFilePath + FileName + u.Host

	file, err := os.Open(fileName)
	defer file.Close()

	if err != nil {
		if os.IsNotExist(err) {
			return nil, err
		}
		log.Fatalf("open file %s failed: %v", fileName, err)
	}

	var cookies []*http.Cookie

	decoder := gob.NewDecoder(file)

	if err := decoder.Decode(&cookies); err != nil {
		log.Fatalf("decode cookies failed: %v", err)
	}

	return cookies, nil
}