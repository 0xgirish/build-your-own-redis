package resp

import (
	"bufio"
	"bytes"
	"fmt"
	"io"

	"0xgirish.eth/redis/app/resp/types"
)

// Scanner provides a convenient interface for reading data such as file or net.Conn.
// Successive calls to the Scan method will step through the resp tokens of a connection,
// skipping the bytes between the tokens.
// After each Scan, Token must be called to retrieve it or it will reset after next Scan call.
type Scanner struct {
	scanner *bufio.Scanner

	token types.Type
	err   error
}

// NewScanner returns a new Scanner to read from r.
func NewScanner(r io.Reader) *Scanner {
	s := &Scanner{
		scanner: bufio.NewScanner(r),
	}

	s.scanner.Split(s.Split)
	return s
}

func (s *Scanner) Scan() bool {
	return s.scanner.Scan()
}

func (s *Scanner) Token() types.Type {
	return s.token
}

func (s *Scanner) Err() error {
	return s.err
}

// Split implements bufio.SplitFunc
func (t *Scanner) Split(data []byte, atEOF bool) (advance int, token []byte, err error) {
	defer func() {
		t.err = err
	}()

	if atEOF && len(data) == 0 {
		t.token = nil
		return 0, nil, nil
	}

	switch data[0] {
	case ':':
		var i types.Int
		t.token = &i
	case '$':
		var bs types.BulkString
		t.token = &bs
	case '+':
		var s types.String
		t.token = &s
	case '-':
		var err types.Error
		t.token = &err
	case '*':
		var arr types.Array
		t.token = &arr
	default:
		return 0, nil, fmt.Errorf("invalid data, wrong byte type: %v", data[0])
	}

	buffer := bytes.NewBuffer(data)
	err = t.token.FromResp(buffer)
	if err == nil {
		n := len(data) - buffer.Len()
		return n, data[:n], nil
	}

	if atEOF {
		// invalid data
		return 0, nil, err
	}

	// Request more data
	return 0, nil, nil
}

// dropCR drops a terminal \r from the data.
// credits: go std `bufio/scan.go`
func dropCR(data []byte) []byte {
	if len(data) > 0 && data[len(data)-1] == '\r' {
		return data[0 : len(data)-1]
	}
	return data
}