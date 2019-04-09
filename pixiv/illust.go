package pixiv

import (
	"github.com/go-xorm/xorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
	"github.com/moonprism/PixivSP/lib"
	log "github.com/sirupsen/logrus"
	"time"
)

// illustration model
type Illust struct {
	// illustration info
	ID	int64	`xorm:"primary_key"`
	Title	string
	Likes	string
	Url	string
	Height	int
	Width	int

	// author info
	AuthorID	string
	AuthorName	string
	// generate status

	Tags	[]string

	CreatedAt time.Time `xorm:"created"`
	UpdatedAt time.Time `xorm:"updated"`
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
	var err error
	engine, err := xorm.NewEngine("mysql", lib.MysqlConf.DSN)
	if err != nil {
		log.Errorf("new engine %v", err)
	}

	if exist, _ := engine.IsTableExist(&Illust{}); !exist {
		err := engine.CreateTables(&Illust{})
		if err != nil {
			log.Errorf("create table %v", err)
		}
	}
}

