package worker

import (
	"bytes"
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"github.com/hellgate75/go-tcp-client/client/proxy"
	"github.com/hellgate75/go-tcp-client/common"
	"github.com/hellgate75/go-tcp-common/log"
	"io/ioutil"
	"time"
)

var Logger log.Logger = log.NewLogger("go-tcp-client", "INFO")

var MainAccess bool = false
type tcpClient struct {
	Cert      common.CertificateKeyPair
	CaCert    string
	IpAddress string
	Port      string
	conn      *tls.Conn
	OS        string
}

func (tcpClient *tcpClient) Open(insecureSkipVerify bool) error {
	config := tls.Config{InsecureSkipVerify: insecureSkipVerify}

	if tcpClient.Cert.Key != "" &&  tcpClient.Cert.Cert != "" {
		Logger.Debugf("client: using client key: <%s>, cert: <%s> ", tcpClient.Cert.Key, tcpClient.Cert.Cert)
		cert, err := tls.LoadX509KeyPair(tcpClient.Cert.Cert, tcpClient.Cert.Key)
		if err != nil {
			Logger.Errorf("client: Unable to load key : %s and certificate: %s", tcpClient.Cert.Key, tcpClient.Cert.Cert)
			Logger.Fatalf("client: loadkeys: %s", err)
		}
		config.Certificates=[]tls.Certificate{cert}
	}

	if "" != tcpClient.CaCert {
		Logger.Debugf("client: Using CA cert: <%s> and insecure skip verify", tcpClient.CaCert)
		rootCAs, _ := x509.SystemCertPool()
		if rootCAs == nil {
			rootCAs = x509.NewCertPool()
		}

		// Read in the cert file
		certs, err := ioutil.ReadFile(tcpClient.CaCert)
		if err != nil {
			Logger.Fatalf("Failed to append %q to RootCAs: %v", tcpClient.CaCert, err)
		}

		// Append our cert to the system pool
		if ok := rootCAs.AppendCertsFromPEM(certs); !ok {
			Logger.Warn("No certs appended, using system certificates only")
		} else {
			config.RootCAs = rootCAs
			config.InsecureSkipVerify = true
		}

	}
	service := fmt.Sprintf("%s:%s", tcpClient.IpAddress, tcpClient.Port)
	Logger.Debugf("Connecting to service: %s", service)
	conn, err := tls.Dial("tcp", service, &config)
	if err != nil {
		Logger.Fatalf("client: dial: %s", err)
		return errors.New(fmt.Sprintf("client: dial: %s", err))
	}
	tcpClient.conn = conn
	Logger.Debugf("client: connected to: %v", conn.RemoteAddr())
	state := conn.ConnectionState()
	Logger.Trace("Uaing certificates: ")
	for _, v := range state.PeerCertificates {
		bytes, errBts := x509.MarshalPKIXPublicKey(v.PublicKey)
		if errBts == nil {
			Logger.Trace("Public Key: ", string(bytes))
		} else {
			Logger.Trace("Public Key: Unavailable")
		}
		Logger.Trace(v.Subject)
	}
	Logger.Trace("client: handshake: ", state.HandshakeComplete)
	Logger.Trace("client: mutual: ", state.NegotiatedProtocolIsMutual)
	Logger.Debug("client: Connected!!")
	time.Sleep(3 * time.Second)
	common.WriteString("os-name", conn)
	Logger.Debug("client: Waiting for remote server OS type...")
	os, errWelcome := common.ReadString(conn)
	if errWelcome != nil {
		Logger.Error("Error acquiring OS: ", errWelcome.Error())
		return err
	}
	tcpClient.OS = os
	Logger.Debugf("Remote server os: %s", os)
	return nil
}

func (tcpClient *tcpClient) ServerOS() string {
	return tcpClient.OS
}

func (tcpClient *tcpClient) IsOpen() bool {
	return tcpClient.conn != nil
}

func (tcpClient *tcpClient) Send(message bytes.Buffer) error {
	if tcpClient.conn == nil {
		return nil
	}
	n, err := common.Write(message.Bytes(), tcpClient.conn)
	if err != nil {
		Logger.Errorf("client: write: %s", err.Error())
		return errors.New(fmt.Sprintf("client: write: %s", err.Error()))
	}
	Logger.Debugf("client: wrote %s (wrote: %d bytes)", message, n)
	if n == 0 {
		return errors.New(fmt.Sprintf("client: written bytes: %d", n))
	}
	return nil
}

func (tcpClient *tcpClient) SendText(message string) error {
	if tcpClient.conn == nil {
		return nil
	}
	n, err := common.WriteString(message, tcpClient.conn)
	if err != nil {
		Logger.Errorf("client: write: %s", err.Error())
		return errors.New(fmt.Sprintf("client: write: %s", err.Error()))
	}
	Logger.Debugf("client: wrote %q (wrote: %d bytes)", message, n)
	if n == 0 {
		return errors.New(fmt.Sprintf("client: written bytes: %d", n))
	}
	return nil
}

func (tcpClient *tcpClient) ApplyCommand(command string, params ...interface{}) error {
	sender, err := proxy.GetSender(command)
	if err != nil {
		Logger.Errorf("client: apply command: %s", err.Error())
		return errors.New(fmt.Sprintf("client: write: %s", err.Error()))
	}
	if ! MainAccess {
		sender.SetLogger(Logger)
	}
	Logger.Debugf("Logger is affiliated for the current sender("+command+"): %v", Logger.IsAffiliated())
	err = sender.SendMessage(tcpClient.conn, params...)
	if err != nil {
		Logger.Errorf("client: command (%s): %s", command, err.Error())
		return errors.New(fmt.Sprintf("client: command (%s): %s", command, err.Error()))
	}
	return nil
}

func (tcpClient *tcpClient) GetHelp() []string {
	return proxy.Help()
}

func (tc *tcpClient) Clone() common.TCPClient {
	return &tcpClient{
		Cert:      tc.Cert,
		IpAddress: tc.IpAddress,
		Port:      tc.Port,
		CaCert:    tc.CaCert,
	}
}

func (tcpClient *tcpClient) Close() error {
	if tcpClient.conn != nil {
		tcpClient.SendText("exit")
		tcpClient.conn.Close()
		tcpClient.conn = nil
	}
	return nil
}

func NewClient(cert common.CertificateKeyPair, caCert string, ipAddress string, port string) common.TCPClient {
	return &tcpClient{
		Cert:      cert,
		IpAddress: ipAddress,
		Port:      port,
		CaCert:    caCert,
	}
}

func (tcpClient *tcpClient) ReadAnswer() (string, error) {
	if tcpClient.conn == nil {
		return "", nil
	}
	return common.ReadString(tcpClient.conn)
}

func (tcpClient *tcpClient) ReadDataPack() ([]byte, error) {
	if tcpClient.conn == nil {
		return []byte{}, nil
	}
	return common.Read(tcpClient.conn)
}
