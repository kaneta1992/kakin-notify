package main
import (
    "io/ioutil"
    "log"
    "gopkg.in/yaml.v2"
    "./imap"
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
            log.Printf(response)
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
