package main

import (
	"./kakin"
	"fmt"
	"github.com/bluele/slack"
	"github.com/okzk/stats"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"log"
	"time"
)

type Config struct {
	UserId    string
	Passward  string
	MailBox   string
	SlackInfo []SlackInfo
}

type SlackInfo struct {
	SlackToken   string
	SlackName    string
	SlackIconUrl string
	SlackChannel string
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
}

func notify(money string) {
	message := "私は課金しました"
	if money != "" {
		message = fmt.Sprintf("私は%s課金しました", money)
	}
	slackNotify(message)
}

var config Config

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	t := stats.SchedulePeriodically(10*time.Minute, func(s *stats.Stats) { log.Printf("gostatus: %v", s) })
	defer t.Stop()

	buf, err := ioutil.ReadFile("config.yml")
	check(err)
	err = yaml.Unmarshal(buf, &config)
	check(err)

	for {
		k, _ := kakin.Create("imap.gmail.com:993", config.UserId, config.Passward, config.MailBox)
		k.Start(notify)
	}
}
