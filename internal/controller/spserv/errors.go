package spserv

import "errors"

var (
	ErrNetOpenConn = errors.New("error of opening the connection")
	ErrSendingMsg  = errors.New("error of sending the message")
)
