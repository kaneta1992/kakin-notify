package imap

import (
	"bufio"
	"crypto/tls"
	"errors"
	"fmt"
	"io"
	"log"
)

type Client struct {
	*parser
	addr string
	conn *tls.Conn
}

type ResponseStatus struct {
	Status string
}

type ResponseFetch struct {
	Text string
}

type ResponseExists struct {
	Exists int
}

type ResponseIdle struct {
	MailCount int
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

func Create(addr string) *Client {
	log.Printf("connecting...")
	conn, err := tls.Dial("tcp", addr, nil)
	check(err)
	log.Printf("connected!")
	var r io.Reader = conn
	client := &Client{
		&parser{&reader{bufio.NewReader(r)}},
		addr,
		conn,
	}

	return client
}

func (self *Client) send(command string, ch chan interface{}) error {
	_, err := self.conn.Write([]byte(command + "\r\n"))
	go self.responseReceiver(ch)
	return err
}

func (self *Client) sendSync(command string) (interface{}, error) {
	log.Printf("start sync")
	ch := make(chan interface{})
	err := self.send(command, ch)
	if err != nil {
		return nil, err
	}

	var data interface{}
	for {
		response := <-ch
		switch t := response.(type) {
		case ResponseStatus:
			if t.Status != "OK" {
				return nil, errors.New("Fatal recieve response")
			}
			log.Printf("end sync %s", t.Status)
			return data, nil
		default:
			log.Printf("response default")
			data = t
		}
	}
}

func (self *Client) Login(id string, pass string) (ResponseStatus, error) {
	log.Printf("start login")
	response, err := self.sendSync(fmt.Sprintf("? LOGIN %s %s", id, pass))
	check(err)
	if t, ok := response.(ResponseStatus); ok {
		log.Printf("end login")
		return t, nil
	}
	return ResponseStatus{}, errors.New("Cast error")
}

func (self *Client) Select(mailbox string) (ResponseStatus, error) {
	response, err := self.sendSync(fmt.Sprintf("? SELECT %s", mailbox))
	check(err)
	if t, ok := response.(ResponseStatus); ok {
		return t, nil
	}
	return ResponseStatus{}, errors.New("Cast error")
}

func (self *Client) Fetch(number string, format string) (ResponseFetch, error) {
	response, err := self.sendSync(fmt.Sprintf("? FETCH %s %s", number, format))
	check(err)
	if t, ok := response.(ResponseFetch); ok {
		return t, nil
	}
	return ResponseFetch{}, errors.New("Cast error")
}

func (self *Client) Idle() (ResponseIdle, error) {
	ch := make(chan interface{})
	err := self.send("? IDLE", ch)
	check(err)
	for {
		//response := <- ch
		// TODO: FETCH or END
	}
	return ResponseIdle{}, err
}

func (self *Client) Logout() (ResponseStatus, error) {
	response, err := self.sendSync(fmt.Sprintf("? LOGOUT"))
	check(err)
	if t, ok := response.(ResponseStatus); ok {
		return t, nil
	}
	return ResponseStatus{}, errors.New("Cast error")
}

func (self *Client) responseReceiver(ch chan interface{}) {
	for {
		token, err := self.readToken()
		check(err)
		switch token {
		case "*":
			log.Printf("start parse untag")
			response, err := self.parseUntag()
			log.Printf("end parse untag")
			check(err)
			err = self.skipToEOL()
			check(err)
			ch <- response
		case "+":
		case "?":
			log.Printf("start parse tag")
			response, err := self.parseTag()
			log.Printf("end parse tag")
			check(err)
			err = self.skipToEOL()
			check(err)
			ch <- response
			return
		default:
			err = self.skipToEOL()
			check(err)
		}
	}
}
