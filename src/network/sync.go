
package network

import (
    "JaonedServer/database"
    "JaonedServer/utils"
    "math"
"net"
    "reflect"
)

type Flag int32

const (
    flagError Flag = 0
    flagLogIn Flag = 1
    flagRegister Flag = 2
    flagShutdown Flag = 3
    flagPointsSet Flag = 4
    flagLine Flag = 5
    flagText Flag = 6
    flagImage Flag = 7

    maxCredentialSize = 16
)

type Sync interface {
    sendBytes(connection net.Conn, bytes []byte, flag Flag)
    logIn(connection net.Conn, message *Message) bool
    register(connection net.Conn, message *Message) bool
    shutdown(connection net.Conn) bool
    processPendingMessages(connection net.Conn, message *Message) []byte // nillable
    pointsSet(connection net.Conn, message *Message) bool
    line(connection net.Conn, message *Message) bool
    text(connection net.Conn, message *Message) bool
    image(connection net.Conn, message *Message) bool
    routeMessage(connection net.Conn, message *Message) bool
    clientDisconnected(connection net.Conn)
}

type SyncImpl struct {
    db database.Database
    network Network
    clients Clients
}

var syncInitialized = false

func createSync(network Network) Sync {
    utils.Assert(!syncInitialized)
    syncInitialized = true

    return &SyncImpl{
        database.Init(),
        network,
        createClients(),
    }
}

func (impl *SyncImpl) sendBytes(connection net.Conn, bytes []byte, flag Flag) {
    var start int32 = 0
    var index int32 = 0
    count := int32(math.Ceil(float64(len(bytes)) / float64(maxMessageBodySize)))
    timestamp := int64(utils.CurrentTimeMillis())

    for {
        message := &Message{
            flag,
            index,
            count,
            timestamp,
            bytes[start:(start + maxMessageBodySize)],
        }

        index++
        if len(message.body) == 0 { break }
        impl.network.sendMessage(connection, message)
    }
}

func (impl *SyncImpl) logIn(connection net.Conn, message *Message) bool {
    utils.Assert(message.body != nil && len(message.body) == maxCredentialSize * 2)

    if impl.clients.getClient(connection) != nil {
        impl.network.sendMessage(connection, &Message{
            flagLogIn,
            0,
            1,
            int64(utils.CurrentTimeMillis()),
            nil,
        })
        return true
    }

    username := message.body[0:maxCredentialSize]
    password := message.body[maxCredentialSize:(maxCredentialSize + maxCredentialSize)]

    user := impl.db.FindUser(username)
    var authenticated bool

    if user != nil {
        authenticated = reflect.DeepEqual(password, user.Password)
    } else {
        authenticated = false
    }

    var body []byte
    if !authenticated {
        body = nil
    } else {
        body = make([]byte, 1)
        impl.clients.addClient(connection, &Client{user})
    }

    impl.network.sendMessage(connection, &Message{
        flagLogIn,
        0,
        1,
        int64(utils.CurrentTimeMillis()),
        body,
    })

    return !authenticated
}

func (impl *SyncImpl) register(connection net.Conn, message *Message) bool {
    utils.Assert(message.body != nil && len(message.body) == maxCredentialSize * 2)

    if impl.clients.getClient(connection) != nil {
        impl.network.sendMessage(connection, &Message{
            flagRegister,
            0,
            1,
            int64(utils.CurrentTimeMillis()),
            nil,
        })
        return true
    }

    username := message.body[0:maxCredentialSize]
    password := message.body[maxCredentialSize:(maxCredentialSize + maxCredentialSize)]

    successful := impl.db.AddUser(username, password)

    var body []byte
    if !successful {
        body = nil
    } else {
        body = make([]byte, 1)
    }

    impl.network.sendMessage(connection, &Message{
        flagRegister,
        0,
        1,
        int64(utils.CurrentTimeMillis()),
        body,
    })

    return true
}

func (impl *SyncImpl) shutdown(connection net.Conn) bool {
    if client := impl.clients.getClient(connection); client != nil && client.IsAdmin {
        impl.network.shutdown()
    } else {
        impl.network.sendMessage(connection, &Message{
            flagError,
            0,
            1,
            int64(utils.CurrentTimeMillis()),
            nil,
        })
    }
    return true
}

func (impl *SyncImpl) processPendingMessages(connection net.Conn, message *Message) []byte { // nillable
    impl.clients.enqueueMessageToClient(connection, message)
    if message.index < message.count - 1 { return nil }

    var bytes []byte
    for impl.clients.clientHasMessages(connection) {
        bytes = append(bytes, impl.clients.dequeueMessageFromClient(connection).body...)
    }

    return bytes
}

func (impl *SyncImpl) pointsSet(connection net.Conn, message *Message) bool {
    bytes := impl.processPendingMessages(connection, message)
    if bytes == nil { return false }

    impl.sendBytes(connection, bytes, flagPointsSet) // TODO: test only
    return false
}

func (impl *SyncImpl) line(connection net.Conn, message *Message) bool {
    bytes := impl.processPendingMessages(connection, message)
    if bytes == nil { return false }

    impl.sendBytes(connection, bytes, flagLine) // TODO: test only
    return false
}

func (impl *SyncImpl) text(connection net.Conn, message *Message) bool {
    bytes := impl.processPendingMessages(connection, message)
    if bytes == nil { return false }

    impl.sendBytes(connection, bytes, flagText) // TODO: test only
    return false
}

func (impl *SyncImpl) image(connection net.Conn, message *Message) bool {
    bytes := impl.processPendingMessages(connection, message)
    if bytes == nil { return false }

    impl.sendBytes(connection, bytes, flagImage) // TODO: test only
    return false
}

func (impl *SyncImpl) routeMessage(connection net.Conn, message *Message) bool {
    disconnect := false

    switch message.flag {
        case flagLogIn:
            disconnect = impl.logIn(connection, message)
        case flagRegister:
            disconnect = impl.register(connection, message)
        case flagShutdown:
            disconnect = impl.shutdown(connection)
        case flagPointsSet:
            disconnect = impl.pointsSet(connection, message)
        case flagLine:
            disconnect = impl.line(connection, message)
        case flagText:
            disconnect = impl.text(connection, message)
        case flagImage:
            disconnect = impl.image(connection, message)
    }

    if disconnect {
        impl.clients.removeClient(connection)
    }

    return disconnect
}

func (impl *SyncImpl) clientDisconnected(connection net.Conn) {
    impl.clients.removeClient(connection)
}
