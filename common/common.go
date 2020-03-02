package common

import (
	"bytes"
)

const (
	DEFAULT_IP_ADDRESS        string = "0.0.0.0"
	DEFAULT_CLIENT_IP_ADDRESS string = "127.0.0.1"
	DEFAULT_PORT              string = "49022"
)

type CertificateKeyPair struct {
	Cert string
	Key  string
}

type TCPClient interface {
	Open(insecureSkipVerify bool) error

	ServerOS() string

	IsOpen() bool

	Send(message bytes.Buffer) error

	SendText(message string) error

	ApplyCommand(command string, params ...interface{}) error

	ReadAnswer() (string, error)

	ReadDataPack() ([]byte, error)

	GetHelp() []string

	Clone() TCPClient

	Close() error
}
