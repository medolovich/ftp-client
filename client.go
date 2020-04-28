package ftpclient

import (
	"fmt"
	"log"
	"net"
	"regexp"
	"strconv"
	"strings"
	"time"
)

const (
	ftpServerHost = "localhost" // TODO: set from env
	ftpServerPort = 21          // TODO: set from env

	ftpResponseCodePartEndIndex      = 3
	ftpResponseMessagePartStartIndex = 4

	newLineChar = "\n" // cross-platforming will die here
	tabChar     = "\t"

	statusCodeAboutToOpenConnection    = "150"
	statusCodeActionSuccess            = "200"
	statusCodeReadyForNewUser          = "220"
	statusCodeClosingControlConnection = "221"
	statusCodeClosingDataConnection    = "226"
	statusCodeEnterPassiveMode         = "227"
	statusCodeLoginSuccess             = "230"
	statusCodeRequestFileActionOK      = "250"
	statusCodePathnameCreated          = "257"
	statusCodeNeedPassword             = "331"
	statusCodeRequestFilePending       = "350"
)

type Client struct {
	conn  *net.TCPConn
	files chan []byte
}

func (c *Client) drop() error {
	close(c.files)
	return c.conn.Close()
}

func (c *Client) send(msg string) error {
	_, err := c.conn.Write([]byte(msg + newLineChar))
	return err
}

func (c *Client) receive() (code string, message string, err error) {
	extractCode := func(raw string) string {
		return raw[:ftpResponseCodePartEndIndex]
	}

	extractMessage := func(raw string) string {
		return raw[ftpResponseMessagePartStartIndex:]
	}

	buf, err := readAll(c.conn)
	if err != nil {
		return "", "", err
	}
	raw := string(buf)
	return extractCode(raw), extractMessage(raw), nil
}

func (c *Client) sendCommand(command string, expectedCodes ...string) (string, error) {
	if err := c.send(fmt.Sprintf(command)); err != nil {
		return "", err
	}

	var message string
	for _, expectedCode := range expectedCodes {
		var code string
		var err error
		code, message, err = c.receive()
		if err != nil {
			return "", err
		}

		if code != expectedCode {
			return "", fmt.Errorf("unexpected code %s: %s", code, message)
		}
	}

	return message, nil
}

func (c *Client) passLogin(ftpLogin string) error {
	_, err := c.sendCommand(fmt.Sprintf("USER %s", ftpLogin), statusCodeNeedPassword)
	return err
}

func (c *Client) passPassword(ftpPassword string) error {
	_, err := c.sendCommand(fmt.Sprintf("PASS %s", ftpPassword), statusCodeLoginSuccess)
	return err
}

func (c *Client) Login(ftpLogin, ftpPassword string) error {
	if err := c.passLogin(ftpLogin); err != nil {
		return fmt.Errorf("Passing login failed: %v", err)
	}
	if err := c.passPassword(ftpPassword); err != nil {
		return fmt.Errorf("Passing password failed: %v", err)
	}
	return nil
}

func (c *Client) Quit() error {
	if _, err := c.sendCommand("QUIT", statusCodeClosingControlConnection); err != nil {
		return err
	}
	return c.drop()
}

func (c *Client) enterPassive() (string, error) {
	msg, err := c.sendCommand("PASV", statusCodeEnterPassiveMode)
	if err != nil {
		return "", err
	}
	r := regexp.MustCompile("(?i)" + `^entering passive mode \((\d+,\d+,\d+,\d+,\d+,\d+)\)`)
	all := r.FindStringSubmatch(msg)
	if len(all) == 0 {
		return "", fmt.Errorf("regex parse string \"%s\" error", msg)
	}
	needle := all[1]
	octets := strings.Split(needle, ",")
	host := strings.Join(octets[:4], ".")
	multiplier, _ := strconv.Atoi(octets[4])
	summand, _ := strconv.Atoi(octets[5])
	port := multiplier*256 + summand
	return fmt.Sprintf("%s:%d", host, port), nil
}

func (c *Client) List() (string, error) { // TODO: parse output into good old files slice
	addr, err := c.enterPassive()
	if err != nil {
		return "", err
	}
	addConn, err := net.Dial("tcp", addr)
	if err != nil {
		return "", err
	}
	_, err = c.sendCommand("LIST", statusCodeAboutToOpenConnection, statusCodeClosingDataConnection)
	if err != nil {
		return "", err
	}
	defer addConn.Close()
	buf, err := readAll(addConn)
	if err != nil {
		return "", err
	}
	return string(buf), nil
}

func (c *Client) ChangeDir(dirName string) error {
	_, err := c.sendCommand(fmt.Sprintf("CWD %s", dirName), statusCodeRequestFileActionOK)
	return err
}

func (c *Client) MakeDir(dirName string) error {
	_, err := c.sendCommand(fmt.Sprintf("MKD %s", dirName), statusCodePathnameCreated)
	return err
}

func (c *Client) RemoveDir(dirName string) error {
	_, err := c.sendCommand(fmt.Sprintf("RMD %s", dirName), statusCodeRequestFileActionOK)
	return err
}

func (c *Client) PWD() (string, error) {
	msg, err := c.sendCommand("PWD", statusCodePathnameCreated)
	return strings.Split(msg, "\"")[1], err
}

func (c *Client) RenameFrom(from string) error {
	_, err := c.sendCommand(fmt.Sprintf("RNFR %s", from), statusCodeRequestFilePending)
	return err
}

func (c *Client) RenameTo(from string) error {
	_, err := c.sendCommand(fmt.Sprintf("RNTO %s", from), statusCodeRequestFileActionOK)
	return err
}

func (c *Client) Delete(fileName string) error {
	_, err := c.sendCommand(fmt.Sprintf("DELE %s", fileName), statusCodeRequestFileActionOK)
	return err
}

func (c *Client) SetBinaryType() error {
	_, err := c.sendCommand("TYPE I", statusCodeActionSuccess)
	return err
}

func NewClient() (*Client, error) {
	conn, err := net.DialTimeout("tcp", fmt.Sprintf("%s:%d", ftpServerHost, ftpServerPort), 5*time.Second)
	if err != nil {
		return nil, err
	}
	tcpConn, ok := conn.(*net.TCPConn)
	if !ok {
		log.Fatal("conn assertion failed")
	}
	c := Client{
		conn:  tcpConn,
		files: make(chan []byte),
	}
	// Receive the greeting message from vsftpd or something like that
	// to not to receive it again in the future
	_, err = c.sendCommand("", statusCodeReadyForNewUser)
	return &c, err
}
