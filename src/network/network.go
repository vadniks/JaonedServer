
package network

import (
    "JaonedServer/utils"
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
    send(connection net.Conn, buffer []byte) utils.Triple
    packMessage(msg *message) []byte
    unpackMessage(bytes []byte) *message
}

type networkImpl struct {
    acceptingClients atomic.Bool
    receivingMessages atomic.Bool
    waitGroup goSync.WaitGroup
}

type message struct {
    size int32
    flag actionFlag
    from int32
    body []byte
}

const messageHeadSize = 12

var networkInitialized = false

func Init() Network {
    utils.Assert(!networkInitialized)
    networkInitialized = true
    return &networkImpl{}
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
    utils.Assert(connection.SetDeadline(time.UnixMilli(int64(utils.CurrentTimeMillis() + 100))) == nil)
}

func (impl *networkImpl) processClient(connection net.Conn) {
    impl.waitGroup.Add(1)

    for impl.receivingMessages.Load() {
        impl.updateConnectionIdleTimeout(connection)

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

//goland:noinspection GoRedundantConversion
func (impl *networkImpl) unpackMessage(bytes []byte) *message {
    size := int32(len(bytes))
    utils.Assert(size >= messageHeadSize)

    msg := new(message)

    copy(unsafe.Slice((*byte) (unsafe.Pointer(&(msg.size))), 4), unsafe.Slice(&(bytes[0]), 4))
    copy(unsafe.Slice((*byte) (unsafe.Pointer(&(msg.flag))), 4), unsafe.Slice(&(bytes[4]), 4))
    copy(unsafe.Slice((*byte) (unsafe.Pointer(&(msg.from))), 4), unsafe.Slice(&(bytes[8]), 4))

    if msg.size > 0 {
        msg.body = make([]byte, msg.size)
        copy(unsafe.Slice(&(msg.body[0]), msg.size), unsafe.Slice(&(bytes[12]), msg.size))
    } else {
        msg.body = nil
    }

    return msg
}
