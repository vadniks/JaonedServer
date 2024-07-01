
package network

import (
    "JaonedServer/database"
    "JaonedServer/utils"
    "net"
    "sync"
)

type Clients interface {
    addClient(connection net.Conn, client *Client)
    getClient(connection net.Conn) *Client // nillable
    removeClient(connection net.Conn) bool
}

type ClientsImpl struct {
    clients map[net.Conn]*Client
    rwMutex sync.RWMutex
}

type Client struct {
    *database.User
}

var clientsInitialized = false

func createClients() Clients {
    utils.Assert(!clientsInitialized)
    clientsInitialized = true

    return &ClientsImpl{
        make(map[net.Conn]*Client),
        sync.RWMutex{},
    }
}

func (impl *ClientsImpl) addClient(connection net.Conn, client *Client) {
    impl.rwMutex.Lock()

    _, found := impl.clients[connection]
    utils.Assert(!found)

    impl.clients[connection] = client

    impl.rwMutex.Unlock()
}

func (impl *ClientsImpl) getClient(connection net.Conn) *Client { // nillable
    return impl.clients[connection]
}

func (impl *ClientsImpl) removeClient(connection net.Conn) bool {
    impl.rwMutex.Lock()

    _, found := impl.clients[connection]
    delete(impl.clients, connection)

    impl.rwMutex.Unlock()
    return found
}
