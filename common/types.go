package common

import (
	"crypto/tls"
	"github.com/hellgate75/go-tcp-common/log"
)

type Sender interface {
	SendMessage(conn *tls.Conn, params ...interface{}) error
	SetLogger(logger log.Logger)
	Helper() string
}
