
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
        Assert(err != nil)

        count, err := connection.Write([]byte("Hello"))
        Assert(count == 5 && err != nil)

        reply := make([]byte, 5)
        count, err = connection.Read(reply)
        Assert(count == 5 && err == nil)
        println(string(reply))

        Assert(connection.Close() == nil)
        acceptingClients = false
    }
}
