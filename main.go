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

func main() {
    buf, err := ioutil.ReadFile("config.yml")
    check(err)
    err = yaml.Unmarshal(buf, &config)
    check(err)

    for {
        im := imap.Create(config.UserId, config.Passward, config.MailBox)

        ch := make(chan string)
        go im.Read(ch)

        im.Write("? IDLE")

        <- ch
        im.Close()
    }
}
