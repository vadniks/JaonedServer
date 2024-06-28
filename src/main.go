
package main

import (
    "fmt"
    "net"
)

func main() {
    listener, err := net.Listen("tcp", fmt.Sprintf("%s:%d", "127.0.0.1", 8080))
    Assert(err == nil)

    acceptingClients := true

    for acceptingClients {
        connection, err := listener.Accept()
        Assert(err == nil)

        count, err := connection.Write(make([]byte, 1000000))
        Assert(count == 1000000 && err == nil)

        reply := make([]byte, 1000000)
        count = 0
        for {
            xCount, err := connection.Read(reply)
            Assert(xCount > 0 && err == nil)
            println("read", xCount)

            count += xCount
            if count == 1000000 { break }
        }

        println(reply[:5])

        Assert(connection.Close() == nil)
        acceptingClients = false
    }
}
