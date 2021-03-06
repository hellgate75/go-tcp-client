<p align="right">
 <img src="https://github.com/hellgate75/go-tcp-client/workflows/Go/badge.svg?branch=master"></img>
&nbsp;&nbsp;<img src="https://api.travis-ci.com/hellgate75/go-tcp-client.svg?branch=master" alt="trevis-ci" width="98" height="20" />&nbsp;&nbsp;<a href="https://travis-ci.com/hellgate75/go-tcp-client">Check last build on Travis-CI</a>
 </p>
<p align="center">
<image width="150" height="50" src="images/kube-go.png"></image>&nbsp;
<image width="260" height="410" src="images/golang-logo.png">
&nbsp;<image width="130" height="50" src="images/tls-logo.png"></image>
</p><br/>
<br/>

# Go TCP Client

Go TLS TCP Client library

## Goals

Definition of a secure layer transfer channel using TLS Serve/Client PEM certificates/keys only protocol. 


## Reference Repositories

* [Go TCP Server](https://github.com/hellgate75/go-tcp-server) Server side application



## How does it work?

Client starts with a single client certificate/key pair. 

Call ```--help``` or ```-h``` from command line to print out the Server command help. 

Call ```help``` to print the list of available plugged-in commands with the syntax. 



Source test script is :
```
[test-client-with-args.sh](/test-client-with-args.sh)
```
It accepts optional parameters or the help request.



## Available plug-ins

Here a sample screen of commands help request:

<p align="center">
<image src="images/commands-screen.png">
</p><br/>

Available commands are:

* transfer-file: transfers files, create folders and copy folder files into to the remote server

* shell: can run interactively (command prompt, without any parameter) or can run not interactively commands and script remotely, copying script files remotely, executing them and deleting the remote files at the end of the execution.


## Use interactive shell:

In order to enter command shell on the remote server you can use the sample script: 

```
./test-client-with-args.sh shell
```

<p align="center">
	<image src="images/commands-interactive shell-1.png">
</p><br/>

<p align="center">
	<image src="images/commands-interactive shell-2.png">
</p><br/>

<p align="center">
	<image src="images/commands-interactive shell-3.png">
</p><br/>



## Need sample certificates?

You can produce test client/server certificates using following provided script:

```
makecert.sh [-d] or using any of provided parameters. 
```

Run it without any parameter to read the usage.



## Build client

Build client:

```
go.exe install -v -gcflags "-N -l" github.com/hellgate75/go-tcp-client
```



## Integrate the client in you application


See following sample

```
package my-package

import(
	"github.com/hellgate75/go-tcp-client/client/worker"
	"github.com/hellgate75/go-tcp-client/common"
)

func myfunc() {
	var certPair common.CertificateKeyPair = common.CertificateKeyPair{
		Cert: "/etc/ssl/client.pem",
		Key: "/etc/ssl/client.key",
	}
	var host string = "my-remote-server-host-or-ip"
	var port string = "49022 or your custom port"
	var insecureSkipVerify bool = false //depends on the server configuration, hopefully you use mandatory certificate check!!
	var client common.TCPClient = worker.NewClient(certPair, host, port)
	client.Open(insecureSkipVerify)
	................
	................
	................
}

```

Plug-in commands are available for call using the client method ```common.TCPClient::ApplyCommand(command string, params ...interface{}) error```




## Get  client and server

Get server:

```
go get -u github.com/hellgate75/go-tcp-server
```

Get client:

```
go get -u github.com/hellgate75/go-tcp-client
```


Enjoy the experience.



## License

The library is licensed with [LGPL v. 3.0](/LICENSE) clauses, with prior authorization of author before any production or commercial use. Use of this library or any extension is prohibited due to high risk of damages due to improper use. No warranty is provided for improper or unauthorized use of this library or any implementation.

Any request can be prompted to the author [Fabrizio Torelli](https://www.linkedin.com/in/fabriziotorelli) at the following email address:

[hellgate75@gmail.com](mailto:hellgate75@gmail.com)




