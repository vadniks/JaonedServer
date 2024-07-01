
package network

import (
    "JaonedServer/database"
    "JaonedServer/utils"
    "net"
    goSync "sync"
)

type clients interface {
    addClient(connection net.Conn, xClient *client)
    getClient(connection net.Conn) *client // nillable
    removeClient(connection net.Conn) bool
}

type clientsImpl struct {
    clients map[net.Conn]*client
    rwMutex goSync.RWMutex
}

type client struct {
    *database.User
}

var clientsInitialized = false

func createClients() clients {
    utils.Assert(!clientsInitialized)
    clientsInitialized = true

    return &clientsImpl{
        make(map[net.Conn]*client),
        goSync.RWMutex{},
    }
}

func (impl *clientsImpl) addClient(connection net.Conn, xClient *client) {
    impl.rwMutex.Lock()

    _, found := impl.clients[connection]
    utils.Assert(!found)

    impl.clients[connection] = xClient

    impl.rwMutex.Unlock()
}

func (impl *clientsImpl) getClient(connection net.Conn) *client { // nillable
    return impl.clients[connection]
}

func (impl *clientsImpl) removeClient(connection net.Conn) bool {
    impl.rwMutex.Lock()

    _, found := impl.clients[connection]
    delete(impl.clients, connection)

    impl.rwMutex.Unlock()
    return found
}
