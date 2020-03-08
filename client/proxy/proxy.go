package proxy

import (
	"fmt"
	"github.com/hellgate75/go-tcp-client/common"
	"github.com/hellgate75/go-tcp-common/log"
	"github.com/hellgate75/go-tcp-modules/client/proxy"
	"os"
	"path/filepath"
	"plugin"
	"io/ioutil"
	"strings"
)

var Logger log.Logger = nil

var UsePlugins bool = false
var PluginLibrariesFolder string = getDefaultPluginsFolder()
var PluginLibrariesExtension = "so"

func GetSender(command string) (common.Sender, error) {
	proxy.Logger = Logger
	if UsePlugins {
		fullPath := fmt.Sprintf("%s%s%s.%s", PluginLibrariesFolder, string(os.PathSeparator), command, PluginLibrariesExtension)
		Logger.Debugf("client.proxy.GetSender() -> Loading library: %s", fullPath)
		plugin, err := plugin.Open(fullPath)
		if err == nil {
			sym, err2 := plugin.Lookup("GetSender")
			if err2 != nil {
				sender, errPlugin := sym.(func(string)(common.Sender, error))(command)
				if errPlugin != nil {
					return nil, errPlugin
				}
				sender.SetLogger(Logger)
				return sender, nil
			}
		}
	}
	sender, errOrig := proxy.GetSender(command)
	if errOrig != nil {
		return nil, errOrig
	}
	sender.SetLogger(Logger)
	return sender, nil
}

func Help() []string {
	var out []string = make([]string, 0)
	if UsePlugins {
		dirName := PluginLibrariesFolder
		_, err0 := os.Stat(dirName)
		if err0 == nil {
			lst, err1 := ioutil.ReadDir(dirName)
			if err1 == nil {
				n := len(PluginLibrariesExtension)
				for _,fi := range lst {
					fileName := fi.Name()
					fileNameLen := len(fileName)
					posix := fileNameLen - n
					if ! fi.IsDir() && posix > 0 && strings.ToLower(fileName[posix:]) == strings.ToLower("." + PluginLibrariesExtension) {
						fullPath := fmt.Sprintf("%s%s%s", PluginLibrariesFolder, string(os.PathSeparator), fileNameLen)
						Logger.Debugf("client.proxy.Help() -> Loading help from library: %s", fullPath)
						plugin, err := plugin.Open(fullPath)
						if err == nil {
							sym, err2 := plugin.Lookup("Help")
							if err2 != nil {
								out = append(out, sym.(func()([]string))()...)
							}
						}
					}
				}
			}
		}
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