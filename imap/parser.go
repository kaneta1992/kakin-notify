package imap

import (
	"errors"
	"log"
	"strconv"
)

type parser struct {
	*reader
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

type ResponseRecent struct {
	Recent int
}

type ResponseIdle struct {
	MailCount int
}

func (self *parser) parseFetch() ResponseFetch {
	token, err := self.readNextBlock()
	check(err)
	num, err := strconv.Atoi(token)
	check(err)
	err = self.skipToEOL()

	data, err := self.readBytes(num)
	check(err)

	return ResponseFetch{data}
}

func (self *parser) parseUntag() (interface{}, error) {
	token, err := self.readToken()
	if err != nil {
		return nil, err
	}

	// 先頭のtokenが数字の場合とそうでない場合がある
	num, err := strconv.Atoi(token)
	if err == nil {
		token, err := self.readToken()
		if err != nil {
			return nil, err
		}
		switch token {
		case "FETCH":
			return self.parseFetch(), nil
		case "EXISTS":
			return ResponseExists{num}, nil
		case "RECENT":
			return ResponseRecent{num}, nil
		}
	}

	// 数字以外のuntagは今回全部無視する
	return nil, nil
}

func (self *parser) parseTag() (interface{}, error) {
	token, err := self.readToken()
	if err != nil {
		return nil, err
	}
	switch token {
	case "OK", "NG", "BAD":
		log.Printf("return Status")
		return ResponseStatus{token}, nil
	default:
		return nil, errors.New("Fatal parse tag")
	}
}
