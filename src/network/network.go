
package network

import (
    "JaonedServer/utils"
    "fmt"
    "net"
    "sync"
    "sync/atomic"
    "time"
)

type Network interface {
    ProcessClients()
    processClient(connection net.Conn, clientId int32)
    receive(connection net.Conn, buffer []byte) utils.Triple
    send(connection net.Conn, buffer []byte) utils.Triple
}

type networkImpl struct {
    acceptingClients atomic.Bool
    receivingMessages atomic.Bool
    waitGroup sync.WaitGroup
}

var initialized = false

func Init() Network {
    utils.Assert(!initialized)
    initialized = true
    return &networkImpl{}
}

func (impl *networkImpl) ProcessClients() {
    listener, err := net.Listen("tcp", fmt.Sprintf("%s:%d", "127.0.0.1", 8080))
    utils.Assert(err == nil)

    impl.acceptingClients.Store(true)
    impl.receivingMessages.Store(true)

    var clientId int32 = 0
    for impl.acceptingClients.Load() {
        connection, err := listener.Accept()
        if err != nil { continue }

        go impl.processClient(connection, clientId)
    }

    impl.waitGroup.Wait()
    utils.Assert(listener.Close() == nil)
}

func (impl *networkImpl) processClient(connection net.Conn, clientId int32) {
    impl.waitGroup.Add(1)
    utils.Assert(connection.SetDeadline(time.UnixMilli(int64(utils.CurrentTimeMillis() + 100))) == nil)

    for impl.receivingMessages.Load() {

    }
}

func (impl *networkImpl) receive(connection net.Conn, buffer []byte) utils.Triple {
    utils.Assert(len(buffer) > 0)

    count, err := connection.Read(buffer)
    if err != nil { return utils.Negative }

    if count == len(buffer) {
        return utils.Positive
    } else {
        return utils.Neutral
    }
}

func (impl *networkImpl) send(connection net.Conn, buffer []byte) utils.Triple {
    utils.Assert(len(buffer) > 0)

    count, err := connection.Write(buffer)
    if err != nil { return utils.Negative }

    if count == len(buffer) {
        return utils.Positive
    } else {
        return utils.Neutral
    }
}
