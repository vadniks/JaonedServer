
package network

import (
    "JaonedServer/database"
    "JaonedServer/utils"
    "net"
    "reflect"
)

type actionFlag int32

const (
    flagLogIn actionFlag = 0
    flagRegister actionFlag = 1
    flagFinish actionFlag = 2
    flagError actionFlag = 3
    flagSuccess actionFlag = 4
    flagShutdown actionFlag = 5

    maxUsernameSize = 8
    maxPasswordSize = 8
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
    utils.Assert(message.body != nil && message.size == maxUsernameSize + maxPasswordSize)

    if impl.clients.getClient(connection) != nil {
        impl.network.sendMessage(connection, &Message{
            0,
            flagLogIn,
            nil,
        })
        return true
    }

    username := message.body[0:maxUsernameSize]
    password := message.body[maxUsernameSize:(maxUsernameSize + maxPasswordSize)]

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
        int32(len(body)),
        flagLogIn,
        body,
    })

    return !authenticated
}

func (impl *SyncImpl) register(connection net.Conn, message *Message) bool {
    utils.Assert(message.body != nil && message.size == maxUsernameSize + maxPasswordSize)

    if impl.clients.getClient(connection) != nil {
        impl.network.sendMessage(connection, &Message{
            0,
            flagRegister,
            nil,
        })
        return true
    }

    username := message.body[0:maxUsernameSize]
    password := message.body[maxUsernameSize:(maxUsernameSize + maxPasswordSize)]

    successful := impl.db.AddUser(username, password)

    var body []byte
    if !successful {
        body = nil
    } else {
        body = make([]byte, 1)
    }

    impl.network.sendMessage(connection, &Message{
        int32(len(body)),
        flagRegister,
        body,
    })

    return true
}

func (impl *SyncImpl) shutdown(connection net.Conn) bool {
    if client := impl.clients.getClient(connection); client != nil && client.IsAdmin {
        impl.network.shutdown()
    } else {
        impl.network.sendMessage(connection, &Message{
            0,
            flagError,
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
