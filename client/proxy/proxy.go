package proxy

import (
	"errors"
	"fmt"
	"github.com/hellgate75/go-tcp-client/client/proxy/shell"
	"github.com/hellgate75/go-tcp-client/client/proxy/transfer"
	"github.com/hellgate75/go-tcp-client/common"
	"github.com/hellgate75/go-tcp-client/log"
)

var sendersMap map[string]common.Sender = make(map[string]common.Sender)
var Logger log.Logger = nil
var filled bool = false



func initMap() {
	sendersMap["transfer-file"] = transfer.New()
	sendersMap["shell"] = shell.New()
	filled = true
}

func GetSender(command string) (common.Sender, error) {
	if !filled {
		initMap()
	}
	if sender, ok := sendersMap[command]; ok {
		sender.SetLogger(Logger)
		return sender, nil
	} else {
		return nil, errors.New(fmt.Sprintf("Sender unavailable: %s", command))
	}
}

func Help() []string {
	if !filled {
		initMap()
	}
	var list []string = make([]string, 0)
	for _, sender := range sendersMap {
		list = append(list, sender.Helper())
	}
	return list
}
