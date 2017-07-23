package imap
import (
    "crypto/tls"
    "bufio"
    "io"
    "log"
    "strings"
    "encoding/base64"
    "regexp"
    "fmt"
)

type Imap struct {
    userId      string
    passward    string
    mailBox     string
    conn        *tls.Conn
    r           *bufio.Reader
    w           io.Writer
}

func Create(id string, pass string, mail string) *Imap {
    log.Printf("connecting...")
    conn, err := tls.Dial("tcp", "imap.gmail.com:993", nil)
    check(err)
    log.Printf("connected!")

    var w io.Writer = conn
    var r io.Reader = conn

    im := &Imap {
        userId:     id,
        passward:   pass,
        mailBox:    mail,
        conn:       conn,
        r:          bufio.NewReader(r),
        w:          w,
    }

    im.Write(fmt.Sprintf("? LOGIN %s %s", id, pass))
    im.Write(fmt.Sprintf("? SELECT %s", mail))

    return im
}

func check(err error) {
    if err != nil {
        log.Fatal(err)
    }
}

func (self *Imap) readToEOL() {
    _, _, err := self.r.ReadLine()
    check(err)
}

func (self *Imap) notify(number string) {
    im := Create(self.userId, self.passward, self.mailBox)

    im.Write("? FETCH " + number + " BODY[1]")
    
    ch := make(chan string)
    go im.Read(ch)
    response := <- ch

    im.Close()

    assined := regexp.MustCompile("合計: (.*)\n")
    group := assined.FindStringSubmatch(string(response))

    if group != nil {
        log.Printf(group[1])
    }
}

func (self *Imap) idle(ch chan string) {
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
                log.Printf("FETCH")
                continue
            }

            go self.notify(num)

        default:
            self.readToEOL()
        }
    }
}

func (self *Imap) readFetch(ch chan string) {
    _, err := self.r.ReadString(' ')
    token, err := self.r.ReadString(' ')
    check(err)
    switch token {
    case "FETCH ":
        self.readToEOL()
        encode_text, err := self.r.ReadString(')')
        check(err)
        encode_text = strings.TrimRight(encode_text, ")")
        encode_text = strings.Replace(string(encode_text), "\r\n", "", -1)
        decode_text, err := base64.StdEncoding.DecodeString(encode_text)
        self.readToEOL()

        if err != nil {
            ch <- string(encode_text)
            log.Printf(string(encode_text))
            return
        }

        ch <- string(decode_text)
        log.Printf(string(decode_text))
    default:
        self.readToEOL()
    }
}

func (self *Imap) Read(ch chan string) {
    for {
        token, err := self.r.ReadString(' ')
        if err != nil {
            self.Close()
            ch <- "close"
            return
        }

        check(err)
        switch token {
        case "* ":
            self.readFetch(ch)
        case "+ ":
            self.readToEOL()
            self.idle(ch)
        default:
            self.readToEOL()
        }
    }
}

func (self *Imap) Write(message string) {
    n, err := self.w.Write([]byte(message + "\r\n"))
    check(err)
    log.Printf("client: wrote %q (%d bytes)", message, n)
}

func (self *Imap) Close() {
    self.Write("? LOGOUT")
    self.conn.Close()
    log.Printf("close conection")
}
