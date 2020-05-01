package ftpclient

import (
	"io"
	"net"
)

func readAll(conn net.Conn) ([]byte, error) {
	buf := make([]byte, 0, 8*packetSize)
	tmpBuf := make([]byte, packetSize)
	for {
		n, err := conn.Read(tmpBuf)
		if n == 0 {
			break
		}
		if err != nil {
			if err != io.EOF {
				return nil, err
			}
			break
		}
		buf = append(buf, tmpBuf[:n]...)
		lastByte := buf[len(buf)-1]
		if lastByte == newLineLinuxASCII || lastByte == newLineWindowsASCII {
			break
		}
	}
	return buf, nil
}
