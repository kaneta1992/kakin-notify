package imap

import (
	"bufio"
	"crypto/tls"
	"fmt"
	"io"
	"log"
	"strings"
)

type Imap struct {
	addr     string
	userId   string
	passward string
	mailBox  string
	conn     *tls.Conn
	response chan string
	r        *bufio.Reader
	w        io.Writer
}

func Create(addr string) *Imap {
	log.Printf("connecting...")
	conn, err := tls.Dial("tcp", addr, nil)
	check(err)
	log.Printf("connected!")

	var w io.Writer = conn
	var r io.Reader = conn

	im := &Imap{
		addr: addr,
		conn: conn,
		r:    bufio.NewReader(r),
		w:    w,
	}

	return im
}

func (self *Imap) Login(id string, pass string, mail string) {
	self.userId = id
	self.passward = pass
	self.mailBox = mail

	ch := make(chan string)
	self.write(fmt.Sprintf("? LOGIN %s %s", id, pass))
	go self.getStatus(ch)
	if status := <-ch; status != "OK" {
		panic("login error")
	}

	self.write(fmt.Sprintf("? SELECT %s", mail))
	go self.getStatus(ch)
	if status := <-ch; status != "OK" {
		panic("select error")
	}
	log.Printf("login")
}

func (self *Imap) Logout() {
	self.write("? LOGOUT")
	log.Printf("logout")
	self.conn.Close()
	log.Printf("close conection")
}

func (self *Imap) Listen(ch chan string) {
	self.write("? IDLE")
	go self.read(ch)
}

func check(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

func (self *Imap) write(message string) {
	n, _ := self.w.Write([]byte(message + "\r\n"))
	log.Printf("client: wrote %q (%d bytes)", message, n)
}

func (self *Imap) getStatus(ch chan string) {
	for {
		token, err := self.r.ReadString(' ')
		check(err)

		switch token {
		case "? ":
			token, err := self.r.ReadString(' ')
			check(err)
			status := strings.TrimRight(token, " ")
			ch <- status
			self.readToEOL()
			return
		default:
			self.readToEOL()
		}
	}
}

func (self *Imap) read(ch chan string) {
	self.response = ch
	for {
		token, err := self.r.ReadString(' ')
		if err != nil {
			self.response <- "close"
			return
		}

		switch token {
		case "* ":
			// FETCHコマンド以外実行されない前提
			self.readFetch()
		case "+ ":
			// IDLE
			self.readToEOL()
			self.idle()
		default:
			// その他はエラー処理なぞせず捨てる
			self.readToEOL()
		}
	}
}

func (self *Imap) readToEOL() {
	_, _, err := self.r.ReadLine()
	check(err)
}

func (self *Imap) readFetch() {
	_, err := self.r.ReadString(' ')
	check(err)
	token, err := self.r.ReadString(' ')
	check(err)
	switch token {
	case "FETCH ":
		self.readToEOL()
		// BODYを読み込む
		encode_text, err := self.r.ReadString(')')
		check(err)
		encode_text = strings.TrimRight(encode_text, ")")
		// crlfでいくつかに区切られているので結合する
		encode_text = strings.Replace(string(encode_text), "\r\n", "", -1)
		self.response <- string(encode_text)
		self.readToEOL()
	default:
		self.readToEOL()
	}
}

func (self *Imap) idle() {
	log.Printf("start idle...")
	for {
		token, err := self.r.ReadString(' ')
		if err != nil {
			log.Printf("idle EOF")
			return
		}
		switch token {
		case "* ":
			num, err := self.r.ReadString(' ')
			check(err)
			num = strings.TrimRight(num, " ")
			log.Printf(string(num))

			token, _, err := self.r.ReadLine()
			if string(token) != "EXISTS" {
				log.Printf(string(token))
				continue
			}
			// IDLEしているgoroutineを止めないために新しいgroutineで実行する
			go self.notify(num)
		default:
			self.readToEOL()
		}
	}
}

func (self *Imap) notify(number string) {
	// EXISTSを検出したらメール本文をチャネルに通知する
	im := Create(self.addr)
	im.Login(self.userId, self.passward, self.mailBox)
	im.write("? FETCH " + number + " BODY[1]")

	ch := make(chan string)
	go im.read(ch)
	response := <-ch

	im.Logout()

	self.response <- response
}
