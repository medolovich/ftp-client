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

type Client struct {
	conn *net.TCPConn
}

func (c *Client) drop() error {
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
	if strings.TrimSpace(command) != "" {
		if err := c.send(fmt.Sprintf(command)); err != nil {
			return "", err
		}
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
			return "", fmt.Errorf("Unexpected code %s: %s", code, message)
		}
	}

	return message, nil
}

func (c *Client) Login(ftpLogin, ftpPassword string) error {
	passLogin := func(ftpLogin string) error {
		_, err := c.sendCommand(fmt.Sprintf("USER %s", ftpLogin), statusCodeNeedPassword)
		return err
	}

	passPassword := func(ftpPassword string) error {
		_, err := c.sendCommand(fmt.Sprintf("PASS %s", ftpPassword), statusCodeLoginSuccess)
		return err
	}

	if err := passLogin(ftpLogin); err != nil {
		return fmt.Errorf("Passing login failed: %v", err)
	}
	if err := passPassword(ftpPassword); err != nil {
		return fmt.Errorf("Passing password failed: %v", err)
	}
	// Set binary transport type
	if err := c.SetBinaryType(); err != nil {
		return err
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
	regex := "(?i)" + `^entering passive mode \((\d+,\d+,\d+,\d+,\d+,\d+)\)`
	r := regexp.MustCompile(regex)
	all := r.FindStringSubmatch(msg)
	if len(all) < 2 {
		return "", fmt.Errorf("Parsing string \"%s\" by regex \"%s\" failed", msg, regex)
	}
	needle := all[1]
	octets := strings.Split(needle, ",")
	host := strings.Join(octets[:4], ".")
	multiplier, _ := strconv.Atoi(octets[4])
	summand, _ := strconv.Atoi(octets[5])
	base := 256
	port := multiplier*base + summand
	return fmt.Sprintf("%s:%d", host, port), nil
}

func (c *Client) List() (string, error) {
	addr, err := c.enterPassive()
	if err != nil {
		return "", err
	}
	addConn, err := net.Dial("tcp", addr)
	if err != nil {
		return "", err
	}
	defer addConn.Close()
	_, err = c.sendCommand("LIST", statusCodeAboutToOpenConnection, statusCodeClosingDataConnection)
	if err != nil {
		return "", err
	}
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

func NewClient(ftpServerHost, ftpServerPort string) (*Client, error) {
	conn, err := net.DialTimeout("tcp", fmt.Sprintf("%s:%s", ftpServerHost, ftpServerPort), 5*time.Second)
	if err != nil {
		return nil, err
	}
	tcpConn, ok := conn.(*net.TCPConn)
	if !ok {
		log.Fatal("conn assertion failed")
	}
	c := Client{conn: tcpConn}
	// Receive the greeting message from vsftpd or something like that
	// to not to receive it again in the future
	_, err = c.sendCommand("", statusCodeReadyForNewUser)
	if err != nil {
		return nil, err
	}
	return &c, nil
}
