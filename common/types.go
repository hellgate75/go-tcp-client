package common

import (
	"crypto/tls"
)

type Sender interface {
	SendMessage(conn *tls.Conn, params ...interface{}) error
	Helper() string
}
