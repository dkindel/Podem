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

###Other distros
Go can be installed from source by following the steps found at this link:
https://golang.org/doc/install/source

##Running PODEM
Now that go has been installed to your machine, PODEM can be run.  The following sections will describe how to run it.

###Setting up
Copy all the `*.go` files to a single directory.  This can be named Podem, Main, or whatever else is desired.  No other `*.go` files other than what is included for podem should be in this directory.  It **may** cause issues in compiling, depending on package names and function names but it is better to only include the files that'll be needed.  

###To run
If you `cd` to the podem directory, there are 2 ways to build the code. One is to run `go build`.  The other is to simply skip the build step and go straight to running the code.  

To run the code, run `go run *.go [filename] [-debug]`. This will build and automatically run the code.  The parameters are explained in the next section.

5 example tests have been provided, c1, c2, c3, c4, and c5. c1 is the same as c17 in past projects in this class with some minor modifications.  c2 is the same as c17 but with an extra input on gate 8.  c3 is a completely different circuit with a lot of types of gates and many gates in different levels.  c4 is the same circuit as c3 but the fault list runs through all possible variations of gate substitutions. c5 is a large circuit. This one will prove that it works for circuits larger than a few gates.

####Parameters

#####filename
The filename paramter is the **base** name of the files to run.  That is, if you're running test c1, this requires the files c1.lev and c1.flt to be in the same directories as the source `*.go` files.  Thus, the command line to run against c1 is:
```
go run *.go c1
```

The .lev and .flt file descriptions are described later.

#####-debug
The debug flag is just a string of "-debug" and is case insensitive.  The importance of this flag is that a lot of debug information can be discovered by setting this.  Backtrace information, periodic circuit states, and objectives can be displayed along with come other information.  If you're looking for specifics of how the code runs, this is the flag you want to set.  

###.lev
The .lev file is the levelized circuit.  A couple example test files are included along with this source code.  It is the same format as the .lev files used in the previous projects with 2 minor changes. The final `END` line is removed and each level is now represented in steps of 1 instead of 5.  There was no reason to implement it in steps of 5 before and just made things a little more complicated.

###.flt
This is the custom-made fault file.  It is read by first reading the first line.  That is just an integer letting the program know how many faults are described in the .flt file.  Every other line has 2 numbers.  The first number is the gate number with the fault.  The second number is the number of the gate type that has been substituted in. 

For instance, if the .flt resembles
```
1
8 8
```
This means that there is 1 fault to be injected and the fault is that gate 8 has been switched with an OR gate.  Gate numbers can be easily found in kindel\_dave\_ckt.go (or as follows):

```go
const (
        JUNK = iota
        T_input
        T_output
        T_xor
        T_xnor
        T_dff
        T_and
        T_nand
        T_or
        T_nor
        T_not
        T_buf
      )
```

The numbering starts as JUNK = 0, T\_input = 1, and so on.  

###Output
If the debug flag is set, a lot of information will be output and most will be pretty clear.  But regardless, the last couple lines will either declare a success or a failure of the PODEM algorithm.  If there's a failure, a failure message will display.  If there's a success, a success message will display along with the input the vectors that will cause the fault to be propogated.

For example, a successful output would be 
```
-------------------------------------------------------
Running PODEM on fault of faulty gate 8 with gate type 8
SUCCESS! D has been propogated
The vector that sensitizes faulty gate 8 with gate type 8 is:
1  =  2
2  =  0
3  =  1
4  =  1
5  =  2
```
