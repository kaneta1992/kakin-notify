package main
import (
    "io/ioutil"
    "log"
    "gopkg.in/yaml.v2"
    "./imap"
    "encoding/base64"
    "regexp"
)

func check(err error) {
    if err != nil {
        log.Fatal(err)
    }
}

type Config struct {
    UserId          string
    Passward        string
    MailBox         string
    SlackToken      string
    SlackName       string
    SlackIconUrl    string
}

var config Config

func responseLoop(im *imap.Imap) {
    ch := make(chan string)
    im.Listen(ch)

    for {
        response := <- ch
        switch response {
        case "close":
            im.Logout()
            return
        default:
            decode_text, err := base64.StdEncoding.DecodeString(response)
            if err != nil {
                log.Printf(response)
                continue
            }
            log.Printf(string(decode_text))

            assined := regexp.MustCompile("合計: (.*)\n")
            group := assined.FindStringSubmatch(string(decode_text))
            if group != nil {
                log.Printf(group[1])
            }
        }
    }
}

func main() {
    log.SetFlags(log.LstdFlags | log.Lshortfile) 
    buf, err := ioutil.ReadFile("config.yml")
    check(err)
    err = yaml.Unmarshal(buf, &config)
    check(err)

    for {
        im := imap.Create()
        im.Login(config.UserId, config.Passward, config.MailBox)

        responseLoop(im)
    }
}
