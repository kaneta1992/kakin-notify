package kakin

import (
	"github.com/jhillyerd/enmime"
	"github.com/kaneta1992/kakin-notify/imap"
	"log"
	"regexp"
	"strings"
)

type Kakin struct {
	client   *imap.Client
	addr     string
	user     string
	passward string
	mailbox  string
}

func Create(addr string, user string, pass string, mailbox string) (*Kakin, error) {
	client, err := imap.Create(addr)
	if err != nil {
		return nil, err
	}

	kakin := &Kakin{
		client,
		addr,
		user,
		pass,
		mailbox,
	}

	_, err = kakin.client.Login(user, pass)
	if err != nil {
		return nil, err
	}

	_, err = kakin.client.Select(mailbox)
	if err != nil {
		return nil, err
	}

	return kakin, nil
}

func analyzeGoogleMail(text string) (string, error) {
	env, err := enmime.ReadEnvelope(strings.NewReader(text))
	if err != nil {
		return "", err
	}

	log.Printf(env.Text)

	assined := regexp.MustCompile("合計: (.*)\r\n")
	group := assined.FindStringSubmatch(env.Text)
	if group != nil {
		return group[1], nil
	} else {
		return "", nil
	}
}

func (self *Kakin) Start(callback func(string)) {
	_, err := self.client.Idle(func(exists int) {
		go func() {
			c, _ := imap.Create(self.addr)
			c.Login(self.user, self.passward)
			c.Select(self.mailbox)
			fetch, err := c.Fetch(exists, "RFC822")
			money, err := analyzeGoogleMail(fetch.Text)
			if err != nil {
				return
			}
			callback(money)
		}()
	})
	if err != nil {
		log.Printf("Idle Error: %s", err)
	}
	self.client.Close()
}
