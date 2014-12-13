#PODEM Algorithm

This project consists of a PODEM ATPG written in the go programming language. 

All source code is inside a single package, though it is spread out across different file to make it somewhat modular. 

##Install
To build and run this code, the most import thing is to have go installed on your computer.  Visit https://golang.org/doc/install to follow OS-specific instructions on how to install but I'll also outline some basics below.

###Ubuntu
To install on Ubuntu, run `sudo apt-get install golang`

Ubuntu will go ahead and install go.  To test an installation, you can save the following code to a file called hello.go:

```go
package main

import "fmt"

func main() {
        fmt.Println("Hello, world")
}
```

In the directory this file is in, run `go run *.go`. This will compile and run the code and you should see `Hello, world` on the command line.

If it all runs properly, you're all set to run the podem code and can skip down to that area!  Otherwise, continue reading here.

If you got an error saying the gopath must be set up, you can use the instructions at this link to fix that. https://golang.org/doc/code.html#GOPATH

In a nutshell, you'll want to run 
```
mkdir $HOME/go
export GOPATH=$HOME/go
export PATH=$PATH:$GOPATH/bin
```

You can also put the following in your ~/.bashrc to have it always set up the GOPATH. 
```
export GOPATH=$HOME/go
export PATH=$PATH:$GOPATH/bin
```

If there's any other errors, google will be your best friend.

###Windows
The link provided has concise and helpful steps to install.  

Go to https://golang.org/dl/

Download the MSI installer for your OS and follow the steps provided.  
