
package network

import (
    "JaonedServer/database"
    "JaonedServer/utils"
    "net"
    "reflect"
    "unsafe"
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

    fromServer = -1
)

type sync interface {
    logIn(connection net.Conn, msg *message) bool
    register(connection net.Conn, msg *message) bool
    routeMessage(connection net.Conn, msg *message) bool
    clientDisconnected(connection net.Conn)
}

type syncImpl struct {
    db database.Database
    network Network
    xClients clients
}

var syncInitialized = false

func createSync(network Network) sync {
    utils.Assert(!syncInitialized)
    syncInitialized = true

    return &syncImpl{
        database.Init(),
        network,
        createClients(),
    }
}

func (impl *syncImpl) logIn(connection net.Conn, msg *message) bool {
    utils.Assert(msg.body != nil && msg.size == maxUsernameSize + maxPasswordSize)

    username := msg.body[0:maxUsernameSize]
    password := msg.body[maxUsernameSize:(maxUsernameSize + maxPasswordSize)]

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
        body = unsafe.Slice((*byte) (unsafe.Pointer(&(user.Id))), 4)
        impl.xClients.addClient(connection, user.Id)
    }

    impl.network.sendMessage(connection, &message{
        int32(len(body)),
        flagLogIn,
        fromServer,
        body,
    })

    return !authenticated
}

func (impl *syncImpl) register(connection net.Conn, msg *message) bool {
    utils.Assert(msg.body != nil && msg.size == maxUsernameSize + maxPasswordSize)

    username := msg.body[0:maxUsernameSize]
    password := msg.body[maxUsernameSize:(maxUsernameSize + maxPasswordSize)]

    successful := impl.db.AddUser(username, password)

    var body []byte
    if !successful {
        body = nil
    } else {
        body = make([]byte, 1)
    }

    impl.network.sendMessage(connection, &message{
        int32(len(body)),
        flagRegister,
        fromServer,
        body,
    })

    return true
}

func (impl *syncImpl) routeMessage(connection net.Conn, msg *message) bool {
    disconnect := false

    switch msg.flag {
        case flagLogIn:
            disconnect = impl.logIn(connection, msg)
        case flagShutdown:
            if msg.from == 0 { impl.network.shutdown() }
            disconnect = true
    }

    if disconnect {
        impl.xClients.removeClient2(connection)
    }

    return disconnect
}

func (impl *syncImpl) clientDisconnected(connection net.Conn) {
    impl.xClients.removeClient2(connection)
}
