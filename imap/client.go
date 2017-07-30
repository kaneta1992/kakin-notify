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

func (self *Client) sendSync(command string) ([]interface{}, error) {
	log.Printf("start sync")
	ch := make(chan interface{})
	err := self.send(command, ch)
	if err != nil {
		return nil, err
	}

	responseList := make([]interface{}, 0)
	for {
		response := <-ch
		switch t := response.(type) {
		case ResponseStatus:
			return append(responseList, t), nil
		default:
			if response != nil {
				responseList = append(responseList, t)
			}
		}
	}
}

func (self *Client) getResponseStatus(responseList []interface{}) (ResponseStatus, error) {
	if len(responseList) == 0 {
		return ResponseStatus{}, errors.New("Not exist response")
	}
	for _, response := range responseList {
		switch t := response.(type) {
		case ResponseStatus:
			return t, nil
		}
	}
	return ResponseStatus{}, errors.New("Not found ResponseStatus")
}

func (self *Client) Login(id string, pass string) (ResponseStatus, error) {
	log.Printf("start login")
	list, err := self.sendSync(fmt.Sprintf("? LOGIN %s %s", id, pass))
	check(err)
	return self.getResponseStatus(list)
}

func (self *Client) Select(mailbox string) (ResponseStatus, error) {
	list, err := self.sendSync(fmt.Sprintf("? SELECT %s", mailbox))
	log.Printf("%v", list)
	check(err)
	return self.getResponseStatus(list)
}

func (self *Client) Fetch(number string, format string) (ResponseFetch, error) {
	list, err := self.sendSync(fmt.Sprintf("? FETCH %s %s", number, format))
	check(err)
	log.Printf("%v", list)
	for _, response := range list {
		switch t := response.(type) {
		case ResponseFetch:
			return t, nil
		}
	}
	return ResponseFetch{}, errors.New("not found ResponseFetch")
}

func (self *Client) Idle(callback func(int)) (ResponseStatus, error) {
	ch := make(chan interface{})
	err := self.send("? IDLE", ch)
	check(err)

	for {
		response := <-ch
		switch t := response.(type) {
		case ResponseStatus:
			return t, nil
		case ResponseExists:
			callback(t.Exists)
		}
	}
}

func (self *Client) Logout() (ResponseStatus, error) {
	list, err := self.sendSync(fmt.Sprintf("? LOGOUT"))
	check(err)
	return self.getResponseStatus(list)
}

func (self *Client) responseReceiver(ch chan interface{}) {
	for {
		token, err := self.readToken()
		check(err)
		switch token {
		case "*":
			response, err := self.parseUntag()
			check(err)
			err = self.skipToEOL()
			check(err)
			ch <- response
		case "?":
			response, err := self.parseTag()
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
