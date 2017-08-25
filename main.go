package main

import (
	"flag"
	"fmt"
	"github.com/bluele/slack"
	"github.com/kaneta1992/kakin-notify/kakin"
	_ "github.com/mattn/go-sqlite3"
	"github.com/naoina/genmai"
	"github.com/okzk/stats"
	"github.com/utahta/go-linenotify"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"log"
	"strconv"
	"time"
)

type Config struct {
	UserId    string
	Passward  string
	MailBox   string
	SlackInfo []SlackInfo
	LineInfo  []LineInfo
}

type SlackInfo struct {
	SlackToken   string
	SlackName    string
	SlackIconUrl string
	SlackChannel string
}

type LineInfo struct {
	LineToken string
}

func check(err error) {
	if err != nil {
		log.Fatalf("Fatal: %v", err)
	}
}

func slackNotify(message string) {
	for _, row := range config.SlackInfo {
		hook := slack.NewWebHook(row.SlackToken)
		err := hook.PostMessage(&slack.WebHookPostPayload{
			Attachments: []*slack.Attachment{
				{Text: message, Color: "danger"},
			},
			Channel:  row.SlackChannel,
			Username: row.SlackName,
			IconUrl:  row.SlackIconUrl,
		})
		check(err)
	}
}

func lineNotify(message string) {
	for _, row := range config.LineInfo {
		c := linenotify.New()
		c.Notify(row.LineToken, message, "", "", nil)
	}
}

type Wastes struct {
	Id        int64 `db:"pk"`
	Price     int
	CreatedAt time.Time
}

func (self *Wastes) BeforeInsert() error {
	self.CreatedAt = time.Now()
	return nil
}

func notify(money string) {

	message := "私は課金しました"
	if money != "" {
		message = fmt.Sprintf("私は%s円課金しました", money)

		// 課金記録
		price, _ := strconv.Atoi(money)
		if _, err := db.Insert(&Wastes{Price: price}); err != nil {
			panic(err)
		}
	}
	slackNotify(message)
	lineNotify(message)
}

var config Config
var db *genmai.DB

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	var configPath string
	flag.StringVar(&configPath, "c", "config.yml", "config file path")
	flag.StringVar(&configPath, "config", "config.yml", "config file path")
	flag.Parse()

	// テーブル作成
	var err error
	if db, err = genmai.New(&genmai.SQLite3Dialect{}, "kakin.db"); err != nil {
		panic(err)
	}
	defer db.Close()
	if err := db.CreateTableIfNotExists(&Wastes{}); err != nil {
		panic(err)
	}

	// config yml
	buf, err := ioutil.ReadFile(configPath)
	check(err)
	err = yaml.Unmarshal(buf, &config)
	check(err)

	t := stats.SchedulePeriodically(30*time.Minute, func(s *stats.Stats) { log.Printf("gostatus: %v", s) })
	defer t.Stop()

	for {
		k, err := kakin.Create("imap.gmail.com:993", config.UserId, config.Passward, config.MailBox)
		if err != nil {
			log.Printf("Kakin Create Error: %s", err)
			time.Sleep(1 * time.Second)
			continue
		}
		k.Start(notify)
	}
}
