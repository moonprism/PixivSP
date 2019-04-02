package pixiv

import (
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
)

// illustration model
type Illust struct {
	// illustration info
	ID	int64	`gorm:"primary_key"`
	Name	string
	Likes	string

	// author info
	AuthorID	string
	AuthorName	string
	// generate status

}

type IllustChan struct {
	IllustID	int64
	UserID	int64

	Index	int

	// generate link
	Link	string
	// progress error info
	Error	error
	// in channel times
	Times	int

	CurrentProgress *IllustSaveProgress
}

// IllustSaveProgress fro display progress bar
type IllustSaveProgress struct {
	ID         string `json:"id"`
	Percentage int    `json:"percentage"`
}

func init() {
	db, _ := gorm.Open("mysql", "")

	defer db.Close()
}
