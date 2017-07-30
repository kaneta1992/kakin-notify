package imap

import (
	"errors"
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

func (self *parser) parseFetch() (ResponseFetch, error) {
	token, err := self.readNextBlock()
	if err != nil {
		return ResponseFetch{}, err
	}
	num, err := strconv.Atoi(token)
	if err != nil {
		return ResponseFetch{}, err
	}
	err = self.skipToEOL()
	if err != nil {
		return ResponseFetch{}, err
	}
	data, err := self.readBytes(num)
	if err != nil {
		return ResponseFetch{}, err
	}

	return ResponseFetch{data}, nil
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
			return self.parseFetch()
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
		return ResponseStatus{token}, nil
	default:
		return nil, errors.New("Fatal parse tag")
	}
}
