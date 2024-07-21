
package network

import (
    "JaonedServer/database"
    "JaonedServer/utils"
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
    logIn(connection net.Conn, message *Message) bool
    register(connection net.Conn, message *Message) bool
    shutdown(connection net.Conn) bool
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

func (impl *SyncImpl) routeMessage(connection net.Conn, message *Message) bool {
    disconnect := false

    switch message.flag {
        case flagLogIn:
            disconnect = impl.logIn(connection, message)
        case flagRegister:
            disconnect = impl.register(connection, message)
        case flagShutdown:
            disconnect = impl.shutdown(connection)
    }

    if disconnect {
        impl.clients.removeClient(connection)
    }

    return disconnect
}

func (impl *SyncImpl) clientDisconnected(connection net.Conn) {
    impl.clients.removeClient(connection)
}
