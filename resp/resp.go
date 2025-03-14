package resp

import "strconv"

const (
	SimpleStringPrefix = "+"
	ErrorPrefix        = "-"
	BulkStringPrefix   = "$"
	CRLF               = "\r\n"
)

type ResponseType = string
type ResponseValue = string

const (
	SimpleStringType ResponseType = "simple_string"
	ErrorType        ResponseType = "error"
	BulkStringType   ResponseType = "bulk_string"
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
		return ErrorPrefix + "ERR " + r.Value + CRLF
	case BulkStringType:
		if r.Value == "" {
			return BulkStringPrefix + "-1" + CRLF
		}
		return BulkStringPrefix + strconv.Itoa(len(r.Value)) + CRLF + r.Value + CRLF
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
