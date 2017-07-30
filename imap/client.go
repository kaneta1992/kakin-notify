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

func Create(addr string) (*Client, error) {
	log.Printf("create")
	conn, err := tls.Dial("tcp", addr, nil)

	var r io.Reader = conn
	client := &Client{
		&parser{&reader{bufio.NewReader(r)}},
		addr,
		conn,
	}

	return client, err
}

func (self *Client) Close() {
	log.Printf("close")
	self.conn.Close()
}

func (self *Client) send(command string, ch chan interface{}) error {
	log.Printf(command)
	_, err := self.conn.Write([]byte(command + "\r\n"))
	go self.responseReceiver(ch)
	return err
}

func (self *Client) sendSync(command string) ([]interface{}, error) {
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
		case error:
			return nil, t
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
	if err != nil {
		return ResponseStatus{}, err
	}
	return self.getResponseStatus(list)
}

func (self *Client) Logout() (ResponseStatus, error) {
	log.Printf("logout")
	list, err := self.sendSync(fmt.Sprintf("? LOGOUT"))
	if err != nil {
		return ResponseStatus{}, err
	}
	return self.getResponseStatus(list)
}

func (self *Client) Select(mailbox string) (ResponseStatus, error) {
	list, err := self.sendSync(fmt.Sprintf("? SELECT %s", mailbox))
	if err != nil {
		return ResponseStatus{}, err
	}
	return self.getResponseStatus(list)
}

func (self *Client) Fetch(number int, format string) (ResponseFetch, error) {
	list, err := self.sendSync(fmt.Sprintf("? FETCH %d %s", number, format))
	if err != nil {
		return ResponseFetch{}, err
	}
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
	if err != nil {
		return ResponseStatus{}, err
	}

	for {
		response := <-ch
		switch t := response.(type) {
		case ResponseStatus:
			return t, nil
		case ResponseExists:
			callback(t.Exists)
		case ResponseFetch:
			// log.Printf("%d: %f", t.Number, t.Text)
		case error:
			return ResponseStatus{}, t
		}
	}
}

func (self *Client) Done() error {
	_, err := self.conn.Write([]byte("Done\r\n"))
	return err
}

func (self *Client) responseReceiver(ch chan interface{}) {
	for {
		token, err := self.readToken()
		if err != nil {
			ch <- err
			return
		}
		switch token {
		case "*":
			response, err := self.parseUntag()
			if err != nil {
				ch <- err
				return
			}
			self.skipToEOL()
			ch <- response
		case "?":
			response, err := self.parseTag()
			if err != nil {
				ch <- err
				return
			}
			self.skipToEOL()
			ch <- response
			return
		default:
			self.skipToEOL()
		}
	}
}
