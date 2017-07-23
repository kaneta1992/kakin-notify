package main
import (
    "crypto/tls"
    "bufio"
    "io"
    "io/ioutil"
    "log"
    "strings"
    "encoding/base64"
    "regexp"
    "gopkg.in/yaml.v2"
    "fmt"
)

func check(err error) {
    if err != nil {
        log.Fatal(err)
    }
}

func readToEOL(br (*bufio.Reader)) {
    _, _, err := br.ReadLine()
    check(err)
}

func notify(number string) {
    conn := connect()

    write(conn, fmt.Sprintf("? LOGIN %s %s", config.UserId, config.Passward))
    write(conn, fmt.Sprintf("? SELECT %s", config.MailBox))
    write(conn, "? FETCH " + number + " BODY[1]")
    write(conn, "? LOGOUT")
    
    ch := make(chan string)
    go read(ch, conn)
    response := <- ch

    assined := regexp.MustCompile("合計: (.*)\n")
    group := assined.FindStringSubmatch(string(response))

    if group != nil {
        log.Printf(group[1])
    }
}

func idle(ch chan string, br (*bufio.Reader)) {
    for {
        token, err := br.ReadString(' ')
        if err != nil {
            log.Printf("idle EOF")
            return
        }
        switch token {
        case "* ":
            num, err := br.ReadString(' ')
            check(err)
            num = strings.TrimRight(num, " ")
            log.Printf(string(num))

            token, _, err := br.ReadLine()
            if string(token) != "EXISTS" {
                log.Printf("FETCH")
                continue
            }

            go notify(num)

        default:
            readToEOL(br)
        }
    }
}

func readFetch(ch chan string, br (*bufio.Reader)) {
    _, err := br.ReadString(' ')
    token, err := br.ReadString(' ')
    check(err)
    switch token {
    case "FETCH ":
        readToEOL(br)
        encode_text, err := br.ReadString(')')
        check(err)
        encode_text = strings.TrimRight(encode_text, ")")
        encode_text = strings.Replace(string(encode_text), "\r\n", "", -1)
        decode_text, err := base64.StdEncoding.DecodeString(encode_text)
        readToEOL(br)

        if err != nil {
            ch <- string(encode_text)
            log.Printf(string(encode_text))
            return
        }

        ch <- string(decode_text)
        log.Printf(string(decode_text))
    default:
        readToEOL(br)
    }
}

func read(ch chan string, conn *tls.Conn) {
    var r io.Reader = conn
    br := bufio.NewReader(r)
    for {
        token, err := br.ReadString(' ')
        if err != nil {
            conn.Close()
            log.Printf("close conection")
            ch <- "close"
            return
        }

        check(err)
        switch token {
        case "* ":
            readFetch(ch, br)
        case "+ ":
            readToEOL(br)
            idle(ch, br)
        default:
            readToEOL(br)
        }
    }
}

func write(w io.Writer, message string) {
    n, err := w.Write([]byte(message + "\r\n"))
    check(err)
    log.Printf("client: wrote %q (%d bytes)", message, n)
}

func connect() (*tls.Conn) {
    log.Printf("connecting...")
    conn, err := tls.Dial("tcp", "imap.gmail.com:993", nil)
    check(err)
    log.Printf("connected!")

    return conn
}

type Config struct {
    UserId          string
    Passward        string
    MailBox         string
    SlackToken      string
    SlackName       string
    SlackIconUrl    string
}

var config Config

func main() {

    buf, err := ioutil.ReadFile("config.yml")
    check(err)
    err = yaml.Unmarshal(buf, &config)
    check(err)

    for {
        conn := connect()

        ch := make(chan string)
        go read(ch ,conn)

        write(conn, fmt.Sprintf("? LOGIN %s %s", config.UserId, config.Passward))
        write(conn, fmt.Sprintf("? SELECT %s", config.MailBox))
        write(conn, "? IDLE")

        <- ch
        conn.Close()
    }
}
