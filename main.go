package main
import (
    "io/ioutil"
    "log"
    "gopkg.in/yaml.v2"
    "./imap"
    _"encoding/base64"
    _"regexp"
    "github.com/bluele/slack"
    _"fmt"
    "github.com/okzk/stats"
    "time"
)

type Config struct {
    UserId          string
    Passward        string
    MailBox         string
    SlackInfo       []SlackInfo
}

type SlackInfo struct {
    SlackToken      string
    SlackName       string
    SlackIconUrl    string
    SlackChannel    string
}

func check(err error) {
    if err != nil {
        log.Fatalf("Fatal: %v", err)
    }
}

func warning(err error) {
    if err != nil {
        log.Printf("Warning: %v", err)
    }
}

func notify(message string) {
    for _, row := range config.SlackInfo {
        hook := slack.NewWebHook(row.SlackToken)
        err := hook.PostMessage(&slack.WebHookPostPayload{
            Attachments: []*slack.Attachment{
                {Text: message, Color: "danger"},
            },
            Channel: row.SlackChannel,
            Username: row.SlackName,
            IconUrl: row.SlackIconUrl,
        })
        check(err)
    }
}

func ora(im *imap.Imap) {
    time.Sleep(10 * time.Second)
    im.Logout()
}

func responseLoop(im *imap.Imap) {
    // ch := make(chan string)
    // im.Listen(ch)

    go ora(im)

    for {
        //token := im.ReadLine()
        _, err := im.ReadToken()
        warning(err)
    }

    // for {
    //     response := <- ch
    //     switch response {
    //     case "close":
    //         im.Logout()
    //         return
    //     default:
    //         decode, err := base64.StdEncoding.DecodeString(response)
    //         decode_text := string(decode)
    //         if err != nil {
    //             decode_text = response
    //         }
    //         log.Printf(string(decode_text))

    //         assined := regexp.MustCompile("合計: (.*)\r\n")
    //         group := assined.FindStringSubmatch(string(decode_text))
    //         if group != nil {
    //             log.Printf(group[1])
    //             notify(fmt.Sprintf("私は%s課金しました", group[1]))
    //         } else {
    //             notify("私は課金しました?")
    //         }
    //     }
    // }
}

var config Config

func main() {
    log.SetFlags(log.LstdFlags | log.Lshortfile) 

	t := stats.SchedulePeriodically(10 * time.Minute, func(s *stats.Stats) { log.Printf("gostatus: %v", s) })
	defer t.Stop()

    buf, err := ioutil.ReadFile("config.yml")
    check(err)
    err = yaml.Unmarshal(buf, &config)
    check(err)

    for {
        im := imap.Create("imap.gmail.com:993")
        im.Login(config.UserId, config.Passward, config.MailBox)
        responseLoop(im)
    }
}
