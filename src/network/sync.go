
package network

import (
    "JaonedServer/utils"
    "net"
)

type actionFlag int32

const (
    flagLogIn actionFlag = 0
    flagRegister actionFlag = 1
    flagFinish actionFlag = 2
    flagError actionFlag = 3
    flagSuccess actionFlag = 4
)

type sync interface {
    routeMessage(connection net.Conn, msg *message) bool
}

type syncImpl struct {

}

var syncInitialized = false

func createSync() sync {
    utils.Assert(!syncInitialized)
    syncInitialized = true
    return &syncImpl{}
}

func (impl *syncImpl) routeMessage(connection net.Conn, msg *message) bool {
    return false
}
