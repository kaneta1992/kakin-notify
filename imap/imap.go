package imap
import (
    "crypto/tls"
    "bufio"
    "io"
    "log"
    "strings"
    "fmt"
    "time"
    "net"
)

type Imap struct {
    addr        string
    userId      string
    passward    string
    mailBox     string
    conn        *tls.Conn
    response    chan string
    r           *bufio.Reader
    w           io.Writer
}

func Create(addr string) *Imap {
    log.Printf("connecting...")
    conn, err := tls.Dial("tcp", addr, nil)
    check(err)
    log.Printf("connected!")

    var w io.Writer = conn
    var r io.Reader = conn

    im := &Imap {
        addr:       addr,
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

    // ch := make(chan string)
    self.write(fmt.Sprintf("? LOGIN %s %s", id, pass))
    // go self.getStatus(ch)
    // if status := <- ch; status != "OK" {
    //     panic("login error")
    // }

    self.write(fmt.Sprintf("? SELECT %s", mail))
    // go self.getStatus(ch)
    // if status := <- ch; status != "OK" {
    //     panic("select error")
    // }
    log.Printf("login")
}

func (self *Imap) Logout() {
    self.write("? LOGOUT")
    log.Printf("logout")
    time.Sleep(10 * time.Second)
    self.conn.Close()
    log.Printf("close conection")
}

func (self *Imap) Listen(ch chan string) {
    self.write("? IDLE")
    go self.read(ch)
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

func (self *Imap) write(message string) {
    n, err := self.w.Write([]byte(message + "\r\n"))
    log.Printf("client: wrote %q (%d bytes)", message, n)
    warning(err)
}

func (self *Imap) getStatus(ch chan string) {
    for {
        token, err := self.r.ReadString(' ')
        check(err)
        log.Printf(string(token))

        switch token {
        case "? ":
            token, err := self.r.ReadString(' ')
            check(err)
            log.Printf(string(token))
            status := strings.TrimRight(token, " ")
            ch <- status
            self.readToEOL()
            return
        default:
            self.readToEOL()
        }
    }
}

func (self *Imap) ReadToken() (string, error) {
    var token []byte
    defer func() {
        self.conn.SetReadDeadline(time.Time{})
    }()

    for {
        self.conn.SetReadDeadline(time.Now().Add(5 * time.Second))
        char, err := self.r.ReadByte()
        if err != nil {
            // タイムアウト
            if e, ok := err.(net.Error); ok && e.Timeout() {
                log.Printf("readToken timeout %v", err)
                continue
            }
            // EOF
            log.Printf("error %v", err)
            return "", err
        }
        switch char {
        case ' ':
            log.Printf(string(token))
            return string(token), nil
        case '\r':
            err := self.r.UnreadByte()
            warning(err)
            self.readToEOL()
            log.Printf(string(token))
            return string(token), nil
        }
        token = append(token, char)
    }
}

func (self *Imap) ReadLine() (string){
    // token, err := self.r.ReadString('\n')
    token, _, err := self.r.ReadLine()
    check(err)
    return string(token)
}

func (self *Imap) read(ch chan string) {
    self.response = ch
    for {
        // TODO: いい感じにtoken単位で読むようにしよう
        token, err := self.r.ReadString(' ')
        log.Printf(string(token))
        warning(err)
        if err != nil {
            log.Printf("send close to channel")
            self.response <- "close"
            log.Printf("return read")
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
    token, _, err := self.r.ReadLine()
    check(err)
    log.Printf(string(token))
}

func (self *Imap) readFetch() {
    token, err := self.r.ReadString(' ')
    check(err)
    log.Printf(string(token))
    token, err = self.r.ReadString(' ')
    check(err)
    log.Printf(string(token))
    switch token {
    case "FETCH ":
        self.readToEOL()
        // BODYを読み込む
        encode_text, err := self.r.ReadString(')')
        check(err)
        encode_text = strings.TrimRight(encode_text, ")")
        // crlfでいくつかに区切られているので結合する
        encode_text = strings.Replace(string(encode_text), "\r\n", "", -1)
        self.response  <- string(encode_text)
        self.readToEOL()
    default:
        self.readToEOL()
    }
}

func (self *Imap) idle() {
    log.Printf("start idle...")
    for {
        token, err := self.r.ReadString(' ')
        warning(err)
        if err != nil {
            log.Printf("idle EOF")
            log.Printf(string(token))
            return
        }
        log.Printf(string(token))

        switch token {
        case "* ":
            num, err := self.r.ReadString(' ')
            check(err)
            num = strings.TrimRight(num, " ")
            log.Printf(string(num))

            token, _, err := self.r.ReadLine()
            warning(err)
            log.Printf(string(token))

            if string(token) != "EXISTS" {
                continue
            }         
            // IDLEしているgoroutineを止めないために新しいgroutineで実行する
            go self.notify(num)
        default:
            self.readToEOL()
        }
    }
}

// EXISTSを検出したらメール本文をチャネルに通知する
func (self *Imap) notify(number string) {
    log.Printf(number + " start notify")
    im := Create(self.addr)
    im.Login(self.userId, self.passward, self.mailBox)
    im.write("? FETCH " + number + " BODY[1]")
    
    ch := make(chan string)
    go im.read(ch)
    response := <- ch
    
    im.Logout()

    self.response <- response
    log.Printf(number + " end notify")
}
