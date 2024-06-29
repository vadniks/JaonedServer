
package network

import "JaonedServer/utils"

type actionFlag int32

const (
    flagLogIn actionFlag = 0
    flagRegister actionFlag = 1
    flagFinish actionFlag = 2
    flagError actionFlag = 3
    flagSuccess actionFlag = 4
)

type sync interface {

}

type syncImpl struct {

}

var syncInitialized = false

func createSync() sync {
    utils.Assert(!syncInitialized)
    syncInitialized = true
    return &syncImpl{}
}


