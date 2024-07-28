
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
    dequeueMessageFromClient(connection net.Conn) *Message // nillable
}

type ClientsImpl struct {
    clients map[net.Conn]*Client
    rwMutex sync.RWMutex
}

type Client struct {
    *database.User
    pendingMessages []*Message
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

func (impl *ClientsImpl) clientHasMessages(connection net.Conn) bool {
    return len(impl.clients[connection].pendingMessages) > 0
}

func (impl *ClientsImpl) enqueueMessageToClient(connection net.Conn, message *Message) {
    if impl.clients[connection].pendingMessages == nil {
        impl.clients[connection].pendingMessages = make([]*Message, 0)
    }

    impl.clients[connection].pendingMessages = append(impl.clients[connection].pendingMessages, message)
}

func (impl *ClientsImpl) dequeueMessageFromClient(connection net.Conn) *Message { // nillable
    if impl.clients[connection].pendingMessages == nil { return nil }

    message := impl.clients[connection].pendingMessages[0]

    impl.clients[connection].pendingMessages = impl.clients[connection].pendingMessages[1:]
    if len(impl.clients[connection].pendingMessages) == 0 { impl.clients[connection].pendingMessages = nil }

    return message
}
