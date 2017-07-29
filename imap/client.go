package imap

type Client struct {
	addr   string
	conn   *tls.Conn
	parser *parser
	writer io.Writer
}

type ResponseStatus struct {
	Status string
}

type ResponseText struct {
	*ResponseStatus
	Text string
}

type ResponseSelect struct {
	*ResponseStatus
	MailCount int
}

type ResponseFetch struct {
	*ResponseStatus
	Text string
}

type ResponseIdle struct {
	*ResponseStatus
	MailCount int
}

func (self *Client) commandSync(command string) (interface{}, error) {
}

func (self *Client) Login(idd string, pass string) (ResponseStatus, error) {
}

func (self *Client) Select(mailbox string) (ResponseSelect, error) {
}

func (self *Client) Fetch(number string, format string) (ResponseFetch, error) {
}

func (self *Client) Idle(onRecieve) (ResponseIdle, error) {
}

func (self *Client) Logout() (ResponseStatus, error) {
}

func (self *Client) responseReceiver() error {
}
