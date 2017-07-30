package imap

import (
	"bufio"
	"log"
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
			// EOF
			warning(err)
			return "", err
		}
		switch char {
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
			// EOF
			warning(err)
			return "", err
		}
		switch char {
		case ' ':
			log.Printf(string(token))
			return string(token), nil
		case '\r':
			err := self.UnreadByte()
			warning(err)
			log.Printf(string(token))
			return string(token), nil
		}
		token = append(token, char)
	}
}

func (self *reader) readBytes(num int) (string, error) {
	var buffer []byte

	for i := 0; i < num; i++ {
		char, err := self.ReadByte()
		if err != nil {
			// EOF
			warning(err)
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
