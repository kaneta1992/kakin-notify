package imap

import (
	"bufio"
	"errors"
	"log"
)

var (
	ErrNotExistBlock = errors.New("Not exist Block")
)

type reader struct {
	*bufio.Reader
}

func (self *reader) readNextBlock() (string, error) {
	var token []byte

	status := false
	for {
		char, err := self.ReadByte()
		if err != nil {
			return "", err
		}
		switch char {
		case '\r':
			if status == false {
				_ = self.UnreadByte()
				return "", ErrNotExistBlock
			}
		case '{':
			status = true
			continue
		case '}':
			return string(token), nil
		}
		if status == true {
			token = append(token, char)
		}
	}
}

func (self *reader) readToken() (string, error) {
	var token []byte

	for {
		char, err := self.ReadByte()
		if err != nil {
			return "", err
		}
		switch char {
		case ' ':
			log.Printf(string(token))
			return string(token), nil
		case '\r':
			log.Printf(string(token))
			err := self.UnreadByte()
			return string(token), err
		}
		token = append(token, char)
	}
}

func (self *reader) readBytes(num int) (string, error) {
	var buffer []byte

	for i := 0; i < num; i++ {
		char, err := self.ReadByte()
		if err != nil {
			return "", err
		}
		buffer = append(buffer, char)
	}
	return string(buffer), nil
}

func (self *reader) skipToEOL() error {
	token, _, err := self.ReadLine()
	log.Printf(string(token))
	return err
}
