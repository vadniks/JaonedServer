/*
 * JaonedServer - an online drawing board
 * Copyright (C) 2024 Vadim Nikolaev (https://github.com/vadniks).
 *
 * This program is free software: you can redistribute it and/or modify
 * it under the terms of the GNU General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * This program is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU General Public License for more details.
 *
 * You should have received a copy of the GNU General Public License
 * along with this program.  If not, see <https://www.gnu.org/licenses/>.
 */

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
    dequeueMessageFromClient(connection net.Conn, timestamp int64) *Message // nillable
    selectBoard(connection net.Conn, board int32)
    getBoard(connection net.Conn) int32 // might be negative
}

type ClientsImpl struct {
    clients map[net.Conn]*Client
    rwMutex sync.RWMutex
}

type Client struct {
    *database.User
    pendingMessages map[int64][]*Message
    board int32
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
    if impl.clients[connection].pendingMessages[message.timestamp] == nil {
        impl.clients[connection].pendingMessages[message.timestamp] = make([]*Message, 0)
    }

    impl.clients[connection].pendingMessages[message.timestamp] = append(impl.clients[connection].pendingMessages[message.timestamp], message)
}

func (impl *ClientsImpl) dequeueMessageFromClient(connection net.Conn, timestamp int64) *Message { // nillable
    if impl.clients[connection].pendingMessages == nil { return nil }

    message := impl.clients[connection].pendingMessages[timestamp][0]

    impl.clients[connection].pendingMessages[timestamp] = impl.clients[connection].pendingMessages[timestamp][1:]
    if len(impl.clients[connection].pendingMessages[timestamp]) == 0 { delete(impl.clients[connection].pendingMessages, timestamp) }

    return message
}

func (impl *ClientsImpl) selectBoard(connection net.Conn, board int32) {
    impl.clients[connection].board = board
}

func (impl *ClientsImpl) getBoard(connection net.Conn) int32 { // might be negative
    return impl.clients[connection].board
}
