package imap
import (
    "crypto/tls"
    "bufio"
    "io"
    "log"
    "strings"
    "fmt"
)

type Imap struct {
    userId      string
    passward    string
    mailBox     string
    conn        *tls.Conn
    response    chan string
    r           *bufio.Reader
    w           io.Writer
}

func Create() *Imap {
    log.Printf("connecting...")
    conn, err := tls.Dial("tcp", "imap.gmail.com:993", nil)
    check(err)
    log.Printf("connected!")

    var w io.Writer = conn
    var r io.Reader = conn

    im := &Imap {
        conn:       conn,
        r:          bufio.NewReader(r),
        w:          w,
    }

    return im
}

func (self *Imap) Login(id string, pass string, mail string) {
    self.userId = id
    self.passward = pass
    self.mailBox = mail 

    self.write(fmt.Sprintf("? LOGIN %s %s", id, pass))
    self.write(fmt.Sprintf("? SELECT %s", mail))

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
            self.readFetch()
        case "+ ":
            self.readToEOL()
            self.idle()
        default:
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
        encode_text, err := self.r.ReadString(')')
        check(err)
        encode_text = strings.TrimRight(encode_text, ")")
        encode_text = strings.Replace(string(encode_text), "\r\n", "", -1)
        self.response  <- string(encode_text)
        self.readToEOL()
    default:
        self.readToEOL()
    }
}

func (self *Imap) idle() {
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

            go self.notify(num)

        default:
            self.readToEOL()
        }
    }
}

func (self *Imap) notify(number string) {
    im := Create()
    im.Login(self.userId, self.passward, self.mailBox)
    im.write("? FETCH " + number + " BODY[1]")
    
    ch := make(chan string)
    go im.read(ch)
    response := <- ch
    
    im.Logout()

    self.response <- response
}
