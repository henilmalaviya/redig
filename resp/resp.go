package resp

import "strconv"

const (
	SimpleStringPrefix = "+"
	ErrorPrefix        = "-"
	ErrorFullPrefix    = ErrorPrefix + "ERR" + " "
	BulkStringPrefix   = "$"
	IntegerPrefix      = ":"
	CRLF               = "\r\n"
)

type ResponseType = string
type ResponseValue = string

const (
	SimpleStringType ResponseType = "simple_string"
	ErrorType        ResponseType = "error"
	BulkStringType   ResponseType = "bulk_string"
	IntegerType      ResponseType = "integer"
)

type Response struct {
	Type  ResponseType
	Value ResponseValue
}

func (r Response) String() string {
	switch r.Type {
	case SimpleStringType:
		return SimpleStringPrefix + r.Value + CRLF
	case ErrorType:
		return ErrorFullPrefix + r.Value + CRLF
	case BulkStringType:
		if r.Value == "" {
			return BulkStringPrefix + "-1" + CRLF
		}
		return BulkStringPrefix + strconv.Itoa(len(r.Value)) + CRLF + r.Value + CRLF
	case IntegerType:
		return IntegerPrefix + r.Value + CRLF
	default:
		return ErrorPrefix + "unknown response type" + CRLF
	}
}

func (r Response) Bytes() []byte {
	return []byte(r.String())
}

func NewResponse(t ResponseType, v ResponseValue) Response {
	return Response{
		Type:  t,
		Value: v,
	}
}

func NewOKResponse() Response {
	return NewResponse(SimpleStringType, "OK")
}

func NewErrorResponse(msg string) Response {
	return NewResponse(ErrorType, msg)
}

func NewIntegerResponse(i int) Response {
	return NewResponse(IntegerType, strconv.Itoa(i))
}

func NewIntegerResponseFromBool(b bool) Response {
	var i int = 0
	if b {
		i = 1
	}
	return NewIntegerResponse(i)
}
