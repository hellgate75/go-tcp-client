package proxy

import (
	"github.com/hellgate75/go-tcp-client/common"
	"github.com/hellgate75/go-tcp-common/log"
	"github.com/hellgate75/go-tcp-modules/client/proxy"
	"io/ioutil"
	"os"
	"path/filepath"
	"plugin"
	"strings"
)

var Logger log.Logger = nil

var UsePlugins bool = false
var PluginLibrariesFolder string = getDefaultPluginsFolder()
var PluginLibrariesExtension = "so"

func GetSender(command string) (common.Sender, error) {
	proxy.Logger = Logger
	if UsePlugins {
		Logger.Debugf("client.proxy.GetSender() -> Loading library for command: %s", command)
		var sender common.Sender = nil
		forEachSenderInPlugins(command, func(sendersList []common.Sender) {
			if len(sendersList) > 0 {
				sender = sendersList[0]
			}
		})
		if sender != nil {

		}
	}
	sender, errOrig := proxy.GetSender(command)
	if errOrig != nil {
		return nil, errOrig
	}
	sender.SetLogger(Logger)
	return sender, nil
}


func filterByExtension(fileName string) bool {
	n := len(PluginLibrariesExtension)
	fileNameLen := len(fileName)
	posix := fileNameLen - n
	return posix > 0 && strings.ToLower(fileName[posix:]) == strings.ToLower("." + PluginLibrariesExtension)
}

func listLibrariesInFolder(dirName string) []string {
	var out []string = make([]string, 0)
	_, err0 := os.Stat(dirName)
	if err0 == nil {
		lst, err1 := ioutil.ReadDir(dirName)
		if err1 == nil {
			for _,file := range lst {
				if file.IsDir() {
					fullDirPath := dirName + string(os.PathSeparator) + file.Name()
					newList := listLibrariesInFolder(fullDirPath)
					out = append(out, newList...)
				} else {
					if filterByExtension(file.Name()) {
						fullFilePath := dirName + string(os.PathSeparator) + file.Name()
						out = append(out, fullFilePath)

					}
				}
			}
		}
	}
	return out
}

func forEachSenderInPlugins(command string, callback func([]common.Sender)())  {
	var senders []common.Sender = make([]common.Sender, 0)
	dirName := PluginLibrariesFolder
	_, err0 := os.Stat(dirName)
	if err0 == nil {
		libraries := listLibrariesInFolder(dirName)
		for _,libraryFullPath := range libraries {
			Logger.Debugf("client.proxy.forEachSenderInPlugins() -> Loading help from library: %s", libraryFullPath)
			plugin, err := plugin.Open(libraryFullPath)
			if err == nil {
				sym, err2 := plugin.Lookup("GetSender")
				if err2 != nil {
					sender, errPlugin := sym.(func(string)(common.Sender, error))(command)
					if errPlugin != nil {
						continue
					}
					sender.SetLogger(Logger)
					senders = append(senders, sender)
				}
			}
		}
	}
	callback(senders)
}

func forEachHelpInPlugins(callback func( []func()([]string) )())  {
	var helpers []func()([]string) = make([]func()([]string), 0)
	dirName := PluginLibrariesFolder
	_, err0 := os.Stat(dirName)
	if err0 == nil {
		libraries := listLibrariesInFolder(dirName)
		for _,libraryFullPath := range libraries {
			Logger.Debugf("client.proxy.forEachHelpInPlugins() -> Loading help from library: %s", libraryFullPath)
			plugin, err := plugin.Open(libraryFullPath)
			if err == nil {
				sym, err2 := plugin.Lookup("Help")
				if err2 != nil {
					helpers = append(helpers, sym.(func()([]string)))
				}
			}
		}
	}
	callback(helpers)
}

func Help() []string {
	var out []string = make([]string, 0)
	if UsePlugins {
		forEachHelpInPlugins(func(helpers []func()([]string)) {
			for _, helper := range helpers {
				out = append(out, helper()...)
			}
		})
	}
	out = append(out, proxy.Help()...)
	return out
}

func getDefaultPluginsFolder() string {
	execPath, err := os.Executable()
	if err != nil {
		pwd, errPwd := os.Getwd()
		if errPwd != nil {
			return filepath.Dir(".") + string(os.PathSeparator) + "modules"
		}
		return filepath.Dir(pwd) + string(os.PathSeparator) + "modules"
	}
	return filepath.Dir(execPath) + string(os.PathSeparator) + "modules"
}