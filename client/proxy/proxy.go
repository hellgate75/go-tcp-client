package proxy

import (
	"github.com/hellgate75/go-tcp-modules/client/proxy"
	"github.com/hellgate75/go-tcp-client/common"
	"github.com/hellgate75/go-tcp-client/log"
)

var Logger log.Logger = nil

func GetSender(command string) (common.Sender, error) {
    proxy.Logger = Logger
    return proxy.GetSender(command)
}

func Help() []string {
	return proxy.Help()
}
