
package network

import (
    "JaonedServer/utils"
"net"
    "sync"
)

type clients interface {
    addClient(connection net.Conn, userId int32)
    removeClient(userId int32)
}

type clientsImpl struct {
    clients map[int32]*client
    rwMutex sync.RWMutex
}

type client struct {
    connection net.Conn
}

func createClients() clients {
    return &clientsImpl{}
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

func (impl *clientsImpl) removeClient(userId int32) {
    impl.rwMutex.Lock()

    _, found := impl.clients[userId]
    utils.Assert(found)

    delete(impl.clients, userId)

    impl.rwMutex.Unlock()
}
