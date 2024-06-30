
package network

import (
    "JaonedServer/utils"
    "errors"
    "fmt"
    "net"
    goSync "sync"
    "sync/atomic"
    "time"
    "unsafe"
)

type Network interface {
    ProcessClients()
    processClient(connection net.Conn)
    receive(connection net.Conn, buffer []byte) utils.Triple
    receiveMessage(connection net.Conn) (*message, error)
    send(connection net.Conn, buffer []byte) utils.Triple
    sendMessage(connection net.Conn, msg *message) utils.Triple
    packMessage(msg *message) []byte
    shutdown()
}

type networkImpl struct {
    acceptingClients atomic.Bool
    receivingMessages atomic.Bool
    waitGroup goSync.WaitGroup
    xSync sync
}

type message struct {
    size int32
    flag actionFlag
    from int32
    body []byte
}

const (
	messageHeadSize = 4 + 4 + 4 // 12
    maxMessageBodySize = 128 - messageHeadSize // 116
    maxMessageSize = 12 + maxMessageBodySize // 128
)

var networkInitialized = false

func Init() Network {
    utils.Assert(!networkInitialized)
    networkInitialized = true

    impl := &networkImpl{}
    impl.xSync = createSync(impl)
    return impl
}

func (impl *networkImpl) ProcessClients() {
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

func (impl *networkImpl) updateConnectionIdleTimeout(connection net.Conn) {
    utils.Assert(connection.SetDeadline(time.UnixMilli(int64(utils.CurrentTimeMillis() + 15 * 60 * 1000))) == nil)
}

func (impl *networkImpl) processClient(connection net.Conn) {
    impl.waitGroup.Add(1)

    for impl.receivingMessages.Load() {
        msg, err := impl.receiveMessage(connection)

        if err != nil { break }
        if msg == nil { continue }

        if impl.xSync.routeMessage(connection, msg) { break }
    }

    impl.waitGroup.Done()
}

func (impl *networkImpl) receive(connection net.Conn, buffer []byte) utils.Triple {
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

func (impl *networkImpl) receiveMessage(connection net.Conn) (*message, error) { // nillable
    head := make([]byte, messageHeadSize)
    result := impl.receive(connection, head)

    if result == utils.Negative { return nil, errors.New("") }
    if result == utils.Neutral { return nil, nil }

    msg := new(message)

    copy(unsafe.Slice((*byte) (unsafe.Pointer(&(msg.size))), 4), unsafe.Slice(&(head[0]), 4))
    copy(unsafe.Slice((*byte) (unsafe.Pointer(&(msg.flag))), 4), unsafe.Slice(&(head[4]), 4))
    copy(unsafe.Slice((*byte) (unsafe.Pointer(&(msg.from))), 4), unsafe.Slice(&(head[8]), 4))

    utils.Assert(msg.size <= maxMessageBodySize)

    if msg.size > 0 {
        body := make([]byte, msg.size)
        result := impl.receive(connection, body)

        if result == utils.Negative { return nil, errors.New("") }
        if result == utils.Neutral { return nil, nil }

        msg.body = body
    } else {
        msg.body = nil
    }

    return msg, nil
}

func (impl *networkImpl) send(connection net.Conn, buffer []byte) utils.Triple {
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

func (impl *networkImpl) sendMessage(connection net.Conn, msg *message) utils.Triple {
    return impl.send(connection, impl.packMessage(msg))
}

//goland:noinspection GoRedundantConversion
func (impl *networkImpl) packMessage(msg *message) []byte {
    utils.Assert(msg.body != nil && int(msg.size) == len(msg.body) || msg.body == nil && msg.size == 0)

    bytes := make([]byte, messageHeadSize + msg.size)

    copy(unsafe.Slice(&(bytes[0]), 4), unsafe.Slice((*byte) (unsafe.Pointer(&(msg.size))), 4))
    copy(unsafe.Slice(&(bytes[4]), 4), unsafe.Slice((*byte) (unsafe.Pointer(&(msg.flag))), 4))
    copy(unsafe.Slice(&(bytes[8]), 4), unsafe.Slice((*byte) (unsafe.Pointer(&(msg.from))), 4))

    if msg.body != nil { copy(unsafe.Slice(&(bytes[12]), msg.size), unsafe.Slice(&(msg.body[0]), msg.size)) }

    return bytes
}

func (impl *networkImpl) shutdown() {
    impl.acceptingClients.Store(false)
    impl.receivingMessages.Store(false)
}
