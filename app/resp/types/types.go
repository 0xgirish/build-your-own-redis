package types

import (
	"bytes"
	"fmt"
	"io"
)

var EmptyBulkString = BulkString("")

var _ error = (*Error)(nil)
var _ Type = (*Int)(nil)
var _ Type = (*String)(nil)
var _ Type = (*BulkString)(nil)
var _ Type = (*Array)(nil)

type Type interface {
	ToResp() string
	FromResp(*bytes.Buffer) error
}

type (
	Int        int    // (:)
	String     string // (+)
	BulkString string // ($)
	Error      string // (-)
)

// Array of any type
type Array struct {
	elements []Type
}

func (i *Int) ToResp() string {
	return fmt.Sprintf(":%d", *i)
}

func (i *Int) FromResp(buffer *bytes.Buffer) error {
	if _, err := fmt.Fscanf(buffer, ":%d\r\n", i); err != nil && err != io.EOF {
		return err
	}

	return nil
}

func (s *String) ToResp() string {
	return fmt.Sprintf("+%s\r\n", *s)
}

func (s *String) FromResp(buffer *bytes.Buffer) error {
	if _, err := fmt.Fscanf(buffer, ":%s\r\n", s); err != nil && err != io.EOF {
		return err
	}

	return nil
}

func (bs *BulkString) ToResp() string {
	if bs == nil {
		return "$-1\r\n"
	}

	return fmt.Sprintf("$%d\r\n%s\r\n", len(*bs), *bs)
}

func (bs *BulkString) FromResp(buffer *bytes.Buffer) error {
	var size int
	if _, err := fmt.Fscanf(buffer, "$%d\r\n", &size); err != nil && err != io.EOF {
		return err
	}

	if size <= 0 {
		return nil
	}

	till := size + 2 // including crlf

	var b = make([]byte, till)
	if n, err := buffer.Read(b); err != nil || n != till {
		return fmt.Errorf("failed to read bulk string worth of bytes %d (including crlf), error: %w", till, err)
	}

	*bs = BulkString(b[:size])
	return nil
}

func (err *Error) ToResp() string {
	return fmt.Sprintf("-%s\r\n", err)
}

func (e *Error) FromResp(buffer *bytes.Buffer) error {
	if _, err := fmt.Fscanf(buffer, ":%s\r\n", e); err != nil && err != io.EOF {
		return err
	}

	return nil
}

func (err *Error) Error() string {
	return string(*err)
}

func (arr *Array) ToResp() string {
	buffer := bytes.NewBufferString("")

	var size = 0
	if arr == nil || arr.elements == nil {
		size = -1
	} else if arr != nil {
		size = len(arr.elements)
	}

	fmt.Fprintf(buffer, "*%d\r\n", size)
	if arr != nil {
		for _, e := range arr.elements {
			fmt.Fprintf(buffer, e.ToResp())
		}
	}

	return buffer.String()
}

func (arr *Array) FromResp(buffer *bytes.Buffer) error {
	var size int
	if _, err := fmt.Fscanf(buffer, "*%d\r\n", &size); err != nil {
		return fmt.Errorf("failed to read array size: %w", err)
	}

	if size < 0 {
		return nil
	}

	if arr.elements == nil {
		arr.elements = make([]Type, 0, size)
	}

	for i := 0; i < size; i++ {
		unread := buffer.Bytes()

		var scanned Type

		switch unread[0] {
		case ':':
			var i Int
			scanned = &i
		case '$':
			var bs BulkString
			scanned = &bs
		case '+':
			var s String
			scanned = &s
		case '-':
			var err Error
			scanned = &err
		case '*':
			var arr Array
			scanned = &arr
		default:
			return fmt.Errorf("invalid data, wrong byte type: %v", unread[0])
		}

		if err := scanned.FromResp(buffer); err != nil {
			return fmt.Errorf("invalid data failed to read at index: %d, %w", i, err)
		}

		arr.elements = append(arr.elements, scanned)
	}

	return nil
}

func (arr *Array) Len() int {
	if arr == nil || arr.elements == nil {
		return -1
	}

	return len(arr.elements)
}

func (arr *Array) Index(i int) Type {
	if arr == nil || len(arr.elements) <= i {
		return nil
	}

	return arr.elements[i]
}

func (arr *Array) CastBulkStringFrom(i int) []BulkString {
	var res = make([]BulkString, 0, len(arr.elements[i:]))

	for _, s := range arr.elements[i:] {
		ss, ok := s.(*BulkString)
		if ok {
			res = append(res, *ss)
		}
	}

	return res
}
