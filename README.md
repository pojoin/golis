# golis
golis is a simple socket framework writened by golang.

##Quick Start
######Download and install

    go get github.com/hechuangqiang/golis

######Create file `echoServer.go`
```go
package main

import "github.com/hechuangqiang/golis"

func main() {
    s := golis.NewServer()
}
```

