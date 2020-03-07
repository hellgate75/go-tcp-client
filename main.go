package main

import (
	"flag"
	"fmt"
	"github.com/gookit/color"
	"github.com/hellgate75/go-tcp-client/client/worker"
	"github.com/hellgate75/go-tcp-client/common"
	"github.com/hellgate75/go-tcp-client/log"
	"os"
	"strings"
	"time"
)

var Logger log.Logger = log.NewAppLogger("go-tcp-client", "INFO")

var certs string = ""
var rootCA string = ""
var keys string = ""
var host string = ""
var port string = ""
var verbosity string = ""
var readTimeout int64 = 0
var fSet *flag.FlagSet

func init() {
	fSet = flag.NewFlagSet("go-tcp-client", flag.ContinueOnError)
	fSet.StringVar(&certs, "certs", "certs/server.pem", "Comma separated pem server certificate list")
	fSet.StringVar(&rootCA, "root-ca", "certs/ca.crt", "Root CA certificate for insecure server config")
	fSet.StringVar(&keys, "keys", "certs/server.key", "Comma separated server certs keys list")
	fSet.StringVar(&host, "ip", common.DEFAULT_CLIENT_IP_ADDRESS, "Server ip address")
	fSet.StringVar(&port, "port", common.DEFAULT_PORT, "Server port")
	fSet.StringVar(&verbosity, "verbosity", "INFO", "Logger verbosity level [TRACE,DEBUG,INFO,ERROR,FATAL] ")
	fSet.Int64Var(&readTimeout, "readTimeout", 5, "Message Read timeout in seconds, used to keep listening for answer from clients")
	worker.MainAccess = true
}

func main() {
	var commands []string = make([]string, 0)
	var args []string = os.Args[1:]
	var hasToken bool = false
	var counter int = 0
	for _, arg := range args {
		if "-h" == arg || "--help" == arg {
			fSet.Usage()
			os.Exit(0)
		}
		if "-" == arg[0:1] {
			if counter < 2 {
				hasToken = true
				counter = 0
			} else {
				commands = append(commands, arg)
			}
		} else if !hasToken {
			counter += 1
			commands = append(commands, arg)
		} else {
			hasToken = false
		}

	}
	if errParse := fSet.Parse(os.Args[1:]); errParse != nil {
		Logger.Errorf("Error in arguments parse: %s", errParse.Error())
		fSet.Usage()
		os.Exit(1)
	}
	common.DEFAULT_TIMEOUT = time.Duration(readTimeout) * time.Second
	if string(Logger.GetVerbosity()) != strings.ToUpper(verbosity) {
		Logger.Debugf("Changing logger verbosity to: %s", strings.ToUpper(verbosity))
		Logger.SetVerbosity(log.VerbosityLevelFromString(strings.ToUpper(verbosity)))
	}
	if string(worker.Logger.GetVerbosity()) != strings.ToUpper(verbosity) {
		Logger.Debugf("Changing worker logger verbosity to: %s", strings.ToUpper(verbosity))
		worker.Logger.SetVerbosity(log.VerbosityLevelFromString(strings.ToUpper(verbosity)))
	}
	var lenght int = len(certs)
	if lenght > len(keys) {
		lenght = len(keys)
	}
	var certPair common.CertificateKeyPair = common.CertificateKeyPair{
		Cert: certs,
		Key:  keys,
	}
	client := worker.NewClient(certPair, rootCA, host, port)
	if len(commands) > 0 {
		var cmd string = commands[0]

		if strings.ToLower(cmd) == "help" ||
			strings.ToLower(cmd) == "--help" ||
			strings.ToLower(cmd) == "-h" {
			list := client.GetHelp()
			fmt.Println("List of commands:")
			for _, item := range list {
				color.Yellow.Printf("- %s", item)
			}
			return

		}
		Logger.Debugf("Summary:\nIp: %s\nPort: %s\ncerts: %v\nkeys: %v\n", host, port, certs, keys)
		defer client.Close()
		errOpen := client.Open(true)
		if errOpen != nil {
			Logger.Errorf("Client start-up error: %s\n", errOpen.Error())
			panic(errOpen.Error())
		}

		if "shutdown" == cmd || "restart" == cmd {
			client.SendText(cmd)
			Logger.Warnf("Called: %s. It will change the server state!!", cmd)
			var repeat bool = true
			var counter int = 0
			for repeat && counter < 2 {
				time.Sleep(2 * time.Second)
				out, errCmd := client.ReadAnswer()
				if errCmd == nil && len(out) >= 2 {
					counter += 1
					if out[0:2] == "ok" {
						Logger.Successf("Called: %s. Success reported from server!!", cmd)
						repeat = false
					} else if out[0:2] == "ko" {
						Logger.Failuref("Called: %s. Errors reported from server, Details -> ", out)
						repeat = false
					} else {
						Logger.Infof("Called: %s. Message reported from server, Details -> ", out)
					}
				} else {
					Logger.Errorf("Error reported waiting for answer: %s", errCmd.Error())
					repeat = false
				}
			}
			return
		} else if "exit" == cmd {
			exitClient(client)
			time.Sleep(2 * time.Second)
			out, errCmd := client.ReadAnswer()
			if errCmd == nil && len(out) >= 2 {
				if out[0:2] == "ok" {
					Logger.Successf("Called: %s. Success reported from server!!", cmd)
				} else if out[0:2] == "ko" {
					Logger.Failuref("Called: %s. Errors reported from server, Details -> ", out)
				} else {
					Logger.Infof("Called: %s. Message reported from server, Details -> ", out)
				}
			}
			return
		}

		var commandArgs []string = commands[1:]
		Logger.Debugf("Command Args: (len: %v) %v", len(commandArgs), commandArgs)
		var params []interface{} = make([]interface{}, 0)
		if "shell" == strings.ToLower(cmd) {
			if len(commandArgs) > 0 {
				params = append(params, commandArgs[0])
			}
			if len(commandArgs) > 1 {
				params = append(params, strings.Join(commandArgs[1:], " "))
			}
		} else {
			for _, val := range commandArgs {
				params = append(params, val)
			}
		}
		Logger.Tracef("Params: (len: %v) %v", len(params), params)
		err1 := client.ApplyCommand(cmd, params...)
		if err1 != nil {
			Logger.Errorf("Error sending command %s, Details: %s", cmd, err1.Error())
			exitClient(client)
			return
		}
		time.Sleep(3 * time.Second)
		answer, err2 := client.ReadAnswer()
		if err2 != nil {
			Logger.Errorf("Error reading response for command %s, Details: %s", cmd, err2.Error())
			exitClient(client)
			return
		}
		exitClient(client)
		if len(answer) >= 2 && "ko" == answer[0:2] {
			Logger.Errorf("Command Message '%s' sent but failed!!", cmd)
			Logger.Errorf("Response: %v", answer)
		} else {
			Logger.Successf("Command Message '%s' sent and executed successfully!!", cmd)
			Logger.Debugf("Response: %v", answer)
		}
	}
	exitClient(client)
}

var exited bool = false

func exitClient(client common.TCPClient) {
	if !exited {
		client.SendText("exit")
		client.Close()
		time.Sleep(2 * time.Second)
		Logger.Success("Exit!!")
		exited = true
	}
}
