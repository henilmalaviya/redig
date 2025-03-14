package resp

import "strconv"

const (
	SimpleStringPrefix = "+"
	ErrorPrefix        = "-"
	ErrorFullPrefix    = ErrorPrefix + "ERR" + " "
	BulkStringPrefix   = "$"
	IntegerPrefix      = ":"
	ArrayPrefix        = "*"
	CRLF               = "\r\n"
)

type Response interface {
	ToString() string
}

type SimpleString struct {
	Value string
}

func (s SimpleString) ToString() string {
	return SimpleStringPrefix + s.Value + CRLF
}

func NewSimpleString(s string) SimpleString {
	return SimpleString{Value: s}
}

func NewOKResponse() SimpleString {
	return NewSimpleString("OK")
}

type Error struct {
	Message string
}

func (e Error) ToString() string {
	return ErrorFullPrefix + e.Message + CRLF
}

func NewError(s string) Error {
	return Error{Message: s}
}

type Integer struct {
	Value int
}

func (i Integer) ToString() string {
	return IntegerPrefix + strconv.Itoa(i.Value) + CRLF
}

func NewInteger(i int) Integer {
	return Integer{Value: i}
}

func NewIntegerFromBool(b bool) Integer {
	if b {
		return NewInteger(1)
	}
	return NewInteger(0)
}

type BulkString struct {
	Value string
}

func (b BulkString) ToString() string {
	if b.Value == "" {
		return BulkStringPrefix + "-1" + CRLF
	}

	return BulkStringPrefix + strconv.Itoa(len(b.Value)) + CRLF + b.Value + CRLF
}

func NewBulkString(s string) BulkString {
	return BulkString{Value: s}
}

type Array struct {
	Elements []Response
}

func (a Array) ToString() string {
	result := ArrayPrefix + strconv.Itoa(len(a.Elements)) + CRLF

	for _, element := range a.Elements {
		result += element.ToString()
	}

	return result
}

func NewArray(elements []Response) Array {
	return Array{Elements: elements}
}
