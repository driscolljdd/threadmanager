# Process Management Library

This library extends an [errgroup](https://golang.org/x/sync/errgroup) to provide a very easy way to control a group of processes.
The library greatly simplifies the process of advanced process management.

## Features

- Automatic catch of unix quit signals and clear signalling to all processes to end
- Universal message channel for every process which takes any form of message
- Start or stop threads at any time during program execution with complete ease
- Simple, well documented code which you can adapt to perfectly fit your needs
- Thread safe; the manager struct could be declared globally and all processes could be used to start / stop other processes

## Installation

Install:
```shell
go get -u github.com/driscolljdd/threads
```

Import:

```go
import "github.com/driscolljdd/threads"
```

## QuickStart

Please view the Example.go file in this package for an easy to follow simple application which handles multiple threads.