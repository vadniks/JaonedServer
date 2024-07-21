
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
    clientHasMessages(connection net.Conn) bool
    enqueueMessageToClient(connection net.Conn, message *Message)
    dequeueMessageToClient(connection net.Conn) *Message // nillable
}

type ClientsImpl struct {
    clients map[net.Conn]*Client
    rwMutex sync.RWMutex
    pendingMessages map[net.Conn][]*Message
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
        make(map[net.Conn][]*Message),
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

func (impl *ClientsImpl) clientHasMessages(connection net.Conn) bool {
    return len(impl.pendingMessages[connection]) > 0
}

func (impl *ClientsImpl) enqueueMessageToClient(connection net.Conn, message *Message) {
    if impl.pendingMessages[connection] == nil {
        impl.pendingMessages[connection] = make([]*Message, 0)
    }

    impl.pendingMessages[connection] = append(impl.pendingMessages[connection], message)
}

func (impl *ClientsImpl) dequeueMessageToClient(connection net.Conn) *Message { // nillable
    if impl.pendingMessages[connection] == nil { return nil }

    message := impl.pendingMessages[connection][0]
    impl.pendingMessages[connection] = impl.pendingMessages[connection][1:]
    return message
}
