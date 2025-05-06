package spserv

import "errors"

var (
	ErrNetOpenConn      = errors.New("error of opening the connection")
	ErrSendingMsg       = errors.New("error of sending the message")
	ErrServiceCondition = errors.New("error of service's condition")
	ErrDataRequest      = errors.New("error of the request's data")
)
