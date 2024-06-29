
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
    routeMessage(connection net.Conn, msg *message) bool
}

type syncImpl struct {
    db database.Database
    network Network
}

var syncInitialized = false

func createSync(network Network) sync {
    utils.Assert(!syncInitialized)
    syncInitialized = true

    return &syncImpl{
        database.Init(),
        network,
    }
}

func (impl *syncImpl) logIn(connection net.Conn, msg *message) bool {
    utils.Assert(msg.body != nil && msg.size == maxUsernameSize + maxPasswordSize)

    username := msg.body[0:maxUsernameSize]
    password := msg.body[maxUsernameSize:(maxUsernameSize + maxPasswordSize)]

    isAdmin := reflect.DeepEqual(username, []byte{0, 0, 0, 0, 0, 0, 0, 0}) && reflect.DeepEqual(password, []byte{1, 1, 1, 1, 1, 1, 1, 1}) // TODO: test only
    isSomeone := reflect.DeepEqual(username, []byte{2, 2, 2, 2, 2, 2, 2, 2}) && reflect.DeepEqual(password, []byte{3, 3, 3, 3, 3, 3, 3, 3}) // TODO: test only

    authenticated := isAdmin || isSomeone

    impl.network.send(connection, impl.network.packMessage(&message{
        0,
        flagLogIn,
        fromServer,
        func() []byte {
            if !authenticated { return nil }

            var id int32
            if isAdmin {
                id = 0
            } else {
                id = 1
            }

            return unsafe.Slice((*byte) (unsafe.Pointer(&(id))), 4)
        }(),
    }))

    return authenticated
}

func (impl *syncImpl) routeMessage(connection net.Conn, msg *message) bool {
    switch msg.flag {
        case flagLogIn:
            return impl.logIn(connection, msg)
        case flagShutdown:
            if msg.from == 0 { impl.network.shutdown() }
            return true
    }

    return false
}
