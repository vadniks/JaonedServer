
package network

import (
    "JaonedServer/utils"
    "net"
    "reflect"
goSync "sync"
)

type clients interface {
    addClient(connection net.Conn, userId int32)
    removeClient(userId int32) bool
    removeClient2(connection net.Conn) bool
}

type clientsImpl struct {
    clients map[int32]*client
    rwMutex goSync.RWMutex
}

type client struct {
    connection net.Conn
}

var clientsInitialized = false

func createClients() clients {
    utils.Assert(!clientsInitialized)
    clientsInitialized = true

    return &clientsImpl{
        make(map[int32]*client),
        goSync.RWMutex{},
    }
}

func (impl *clientsImpl) addClient(connection net.Conn, userId int32) {
    impl.rwMutex.Lock()

    _, found := impl.clients[userId]
    utils.Assert(!found)

    impl.clients[userId] = &client{
        connection,
    }

    impl.rwMutex.Unlock()
}

func (impl *clientsImpl) removeClient(userId int32) bool {
    impl.rwMutex.Lock()

    _, found := impl.clients[userId]
    delete(impl.clients, userId)

    impl.rwMutex.Unlock()
    return found
}

func (impl *clientsImpl) removeClient2(connection net.Conn) bool {
    impl.rwMutex.Lock()

    found := false
    for userId, xClient := range impl.clients {
        if reflect.DeepEqual(xClient.connection, connection) {
            found = true
            delete(impl.clients, userId)
            break
        }
    }

    impl.rwMutex.Unlock()
    return found
}
