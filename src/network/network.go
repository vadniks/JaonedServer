
package network

import (
    "JaonedServer/utils"
    "errors"
    "fmt"
    "net"
    "sync"
    "sync/atomic"
    "time"
    "unsafe"
)

type Network interface {
    ProcessClients()
    processClient(connection net.Conn)
    receive(connection net.Conn, buffer []byte) utils.Triple
    receiveMessage(connection net.Conn) (*Message, error)
    send(connection net.Conn, buffer []byte) utils.Triple
    sendMessage(connection net.Conn, message *Message) utils.Triple
    packMessage(message *Message) []byte
    shutdown()
}

type NetworkImpl struct {
    acceptingClients atomic.Bool
    receivingMessages atomic.Bool
    waitGroup sync.WaitGroup
    sync Sync
}

type Message struct {
    size int32
    flag actionFlag
    body []byte
}

const (
	messageHeadSize = 4 + 4 // 8
    maxMessageBodySize = 128 - messageHeadSize // 120
    maxMessageSize = messageHeadSize + maxMessageBodySize // 128
)

var networkInitialized = false

func Init() Network {
    utils.Assert(!networkInitialized)
    networkInitialized = true

    impl := &NetworkImpl{}
    impl.sync = createSync(impl)
    return impl
}

func (impl *NetworkImpl) ProcessClients() {
    listener, err := net.Listen("tcp", fmt.Sprintf("%s:%d", "127.0.0.1", 8080))
    utils.Assert(err == nil)

    impl.acceptingClients.Store(true)
    impl.receivingMessages.Store(true)

    for impl.acceptingClients.Load() {
        connection, err := listener.Accept()
        if err != nil { continue }

        go impl.processClient(connection)
    }

    impl.waitGroup.Wait()
    utils.Assert(listener.Close() == nil)
}

func (impl *NetworkImpl) updateConnectionIdleTimeout(connection net.Conn) {
    utils.Assert(connection.SetDeadline(time.UnixMilli(int64(utils.CurrentTimeMillis() + 15 * 60 * 1000))) == nil) // 15 minutes
}

func (impl *NetworkImpl) processClient(connection net.Conn) {
    impl.waitGroup.Add(1)

    for impl.receivingMessages.Load() {
        message, err := impl.receiveMessage(connection)

        if err != nil { break }
        if message == nil { continue }

        if impl.sync.routeMessage(connection, message) { break }
    }

    impl.sync.clientDisconnected(connection)
    utils.Assert(connection.Close() == nil)
    impl.waitGroup.Done()
}

func (impl *NetworkImpl) receive(connection net.Conn, buffer []byte) utils.Triple {
    utils.Assert(len(buffer) > 0)

    count, err := connection.Read(buffer)
    if err != nil { return utils.Negative }

    impl.updateConnectionIdleTimeout(connection)
    if count == len(buffer) {
        return utils.Positive
    } else {
        return utils.Neutral
    }
}

func (impl *NetworkImpl) receiveMessage(connection net.Conn) (*Message, error) { // nillable
    head := make([]byte, messageHeadSize)
    result := impl.receive(connection, head)

    if result == utils.Negative { return nil, errors.New("") }
    if result == utils.Neutral { return nil, nil }

    message := new(Message)

    copy(unsafe.Slice((*byte) (unsafe.Pointer(&(message.size))), 4), unsafe.Slice(&(head[0]), 4))
    copy(unsafe.Slice((*byte) (unsafe.Pointer(&(message.flag))), 4), unsafe.Slice(&(head[4]), 4))

    utils.Assert(message.size <= maxMessageBodySize)

    if message.size > 0 {
        body := make([]byte, message.size)
        result := impl.receive(connection, body)

        if result == utils.Negative { return nil, errors.New("") }
        if result == utils.Neutral { return nil, nil }

        message.body = body
    } else {
        message.body = nil
    }

    return message, nil
}

func (impl *NetworkImpl) send(connection net.Conn, buffer []byte) utils.Triple {
    utils.Assert(len(buffer) > 0)

    count, err := connection.Write(buffer)
    if err != nil { return utils.Negative }

    impl.updateConnectionIdleTimeout(connection)
    if count == len(buffer) {
        return utils.Positive
    } else {
        return utils.Neutral
    }
}

func (impl *NetworkImpl) sendMessage(connection net.Conn, message *Message) utils.Triple {
    return impl.send(connection, impl.packMessage(message))
}

func (impl *NetworkImpl) packMessage(message *Message) []byte {
    utils.Assert(message.body != nil && int(message.size) == len(message.body) || message.body == nil && message.size == 0)

    bytes := make([]byte, messageHeadSize + message.size)

    copy(unsafe.Slice(&(bytes[0]), 4), unsafe.Slice((*byte) (unsafe.Pointer(&(message.size))), 4))
    copy(unsafe.Slice(&(bytes[4]), 4), unsafe.Slice((*byte) (unsafe.Pointer(&(message.flag))), 4))

    if message.body != nil { copy(unsafe.Slice(&(bytes[8]), message.size), unsafe.Slice(&(message.body[0]), message.size)) }

    return bytes
}

func (impl *NetworkImpl) shutdown() {
    impl.acceptingClients.Store(false)
    impl.receivingMessages.Store(false)
}
