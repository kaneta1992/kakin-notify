package imap

import (
	"bufio"
)

type reader struct {
	*bufio.Reader
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
			// log.Printf(string(token))
			return string(token), nil
		case '\r':
			err := self.UnreadByte()
			warning(err)
			// log.Printf(string(token))
			return string(token), nil
		}
		token = append(token, char)
	}
}

func (self *reader) readTag() (string, error) {
	token, err := self.readToken()
	warning(err)
	switch token[:1] {
	case "?":
		return token, nil
	default:
		return "", fmt.Errorf("[%s] not tag", token)
	}
}
