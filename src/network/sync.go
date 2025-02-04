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
    "math"
    "net"
    "reflect"
    "unsafe"
)

type Flag int32

const (
    flagError Flag = 0
    flagLogIn Flag = 1
    flagRegister Flag = 2
    flagShutdown Flag = 3
    flagCreateBoard Flag = 4
    flagGetBoard Flag = 5
    flagGetBoards Flag = 6
    flagDeleteBoard Flag = 7
    flagPointsSet Flag = 8
    flagLine Flag = 9
    flagText Flag = 10
    flagImage Flag = 11
    flagUndo Flag = 12
    flagClear Flag = 13
    flagSelectBoard Flag = 14
    flagGetBoardElements Flag = 15

    maxCredentialSize = database.MaxCredentialSize
)

type Sync interface {
    terminate()
    sendBytes(connection net.Conn, bytes []byte, flag Flag)
    logIn(connection net.Conn, message *Message) bool
    register(connection net.Conn, message *Message) bool
    shutdown(connection net.Conn) bool
    processPendingMessages(connection net.Conn, message *Message) []byte // nillable
    packBoard(board *database.Board) []byte
    unpackBoard(bytes []byte) *database.Board
    createBoard(connection net.Conn, message *Message) bool
    getBoard(connection net.Conn, message *Message) bool
    getBoards(connection net.Conn) bool
    deleteBoard(connection net.Conn, message *Message) bool
    pointsSet(connection net.Conn, message *Message) bool
    line(connection net.Conn, message *Message) bool
    text(connection net.Conn, message *Message) bool
    image(connection net.Conn, message *Message) bool
    undo(connection net.Conn) bool
    clear(connection net.Conn) bool
    selectBoard(connection net.Conn, message *Message) bool
    boardElements(connection net.Conn) bool
    routeMessage(connection net.Conn, message *Message) bool
    clientDisconnected(connection net.Conn)
}

type SyncImpl struct {
    db database.Database
    network Network
    clients Clients
}

var syncInitialized = false

func createSync(network Network) Sync {
    utils.Assert(!syncInitialized)
    syncInitialized = true

    return &SyncImpl{
        database.Init(),
        network,
        createClients(),
    }
}

func (impl *SyncImpl) terminate() {
    impl.db.Close()
}

func (impl *SyncImpl) sendBytes(connection net.Conn, bytes []byte, flag Flag) {
    var start int32 = 0
    var index int32 = 0
    count := int32(math.Ceil(float64(len(bytes)) / float64(maxMessageBodySize)))
    timestamp := int64(utils.CurrentTimeMillis())

    for {
        if start >= int32(len(bytes)) { break }

        end := start + maxMessageBodySize
        if end >= int32(len(bytes)) { end = int32(len(bytes)) }

        message := &Message{
            flag,
            index,
            count,
            timestamp,
            bytes[start:end],
        }

        start += int32(len(message.body))
        index++

        if len(message.body) == 0 { break }
        impl.network.sendMessage(connection, message)
    }
}

func (impl *SyncImpl) logIn(connection net.Conn, message *Message) bool {
    utils.Assert(message.body != nil && len(message.body) == maxCredentialSize * 2)

    if impl.clients.getClient(connection) != nil {
        impl.network.sendMessage(connection, &Message{
            flagLogIn,
            0,
            1,
            int64(utils.CurrentTimeMillis()),
            nil,
        })
        return true
    }

    username := message.body[0:maxCredentialSize]
    password := message.body[maxCredentialSize:(maxCredentialSize + maxCredentialSize)]

    user := impl.db.FindUser(username)
    var authenticated bool

    if user != nil {
        authenticated = reflect.DeepEqual(password, user.Password)
    } else {
        authenticated = false
    }

    var body []byte
    if !authenticated {
        body = nil
    } else {
        body = make([]byte, 1)
        impl.clients.addClient(connection, &Client{user, make(map[int64][]*Message), -1})
    }

    impl.network.sendMessage(connection, &Message{
        flagLogIn,
        0,
        1,
        int64(utils.CurrentTimeMillis()),
        body,
    })

    return !authenticated
}

func (impl *SyncImpl) register(connection net.Conn, message *Message) bool {
    utils.Assert(message.body != nil && len(message.body) == maxCredentialSize * 2)

    if impl.clients.getClient(connection) != nil {
        impl.network.sendMessage(connection, &Message{
            flagRegister,
            0,
            1,
            int64(utils.CurrentTimeMillis()),
            nil,
        })
        return true
    }

    username := message.body[0:maxCredentialSize]
    password := message.body[maxCredentialSize:(maxCredentialSize + maxCredentialSize)]

    successful := impl.db.AddUser(username, password)

    var body []byte
    if !successful {
        body = nil
    } else {
        body = make([]byte, 1)
    }

    impl.network.sendMessage(connection, &Message{
        flagRegister,
        0,
        1,
        int64(utils.CurrentTimeMillis()),
        body,
    })

    return true
}

func (impl *SyncImpl) shutdown(connection net.Conn) bool {
    if client := impl.clients.getClient(connection); client != nil && client.IsAdmin {
        impl.network.shutdown()
    } else {
        impl.network.sendMessage(connection, &Message{
            flagError,
            0,
            1,
            int64(utils.CurrentTimeMillis()),
            nil,
        })
    }
    return true
}

func (impl *SyncImpl) processPendingMessages(connection net.Conn, message *Message) []byte { // nillable
    impl.clients.enqueueMessageToClient(connection, message)
    if message.index < message.count - 1 { return nil }

    var bytes []byte
    for impl.clients.clientHasMessages(connection) {
        bytes = append(bytes, impl.clients.dequeueMessageFromClient(connection, message.timestamp).body...)
    }

    return bytes
}

func (impl *SyncImpl) packBoard(board *database.Board) []byte {
    size := len(board.Title)
    utils.Assert(size <= database.MaxBoardTitleSize)

    bytes := make([]byte, 4 + 4 + 4 + size)
    copy(unsafe.Slice(&(bytes[0]), 4), unsafe.Slice((*byte) (unsafe.Pointer(&(board.Id))), 4))
    copy(unsafe.Slice(&(bytes[4]), 4), unsafe.Slice((*byte) (unsafe.Pointer(&(board.Color))), 4))
    copy(unsafe.Slice(&(bytes[8]), 4), unsafe.Slice((*byte) (unsafe.Pointer(&(size))), 4))
    copy(unsafe.Slice(&(bytes[12]), size), board.Title)
    return bytes
}

func (impl *SyncImpl) unpackBoard(bytes []byte) *database.Board {
    board := new(database.Board)
    var size int32

    copy(unsafe.Slice((*byte) (unsafe.Pointer(&(board.Id))), 4), unsafe.Slice(&(bytes[0]), 4))
    copy(unsafe.Slice((*byte) (unsafe.Pointer(&(board.Color))), 4), unsafe.Slice(&(bytes[4]), 4))
    copy(unsafe.Slice((*byte) (unsafe.Pointer(&(size))), 4), unsafe.Slice(&(bytes[8]), 4))

    board.Title = make([]byte, size)
    copy(board.Title, unsafe.Slice(&(bytes[12]), size))

    return board
}

func (impl *SyncImpl) createBoard(connection net.Conn, message *Message) bool {
    client := impl.clients.getClient(connection)
    if client == nil { return true }

    impl.db.AddBoard(client.Username, impl.unpackBoard(message.body))

    impl.network.sendMessage(connection, &Message{
        flagCreateBoard,
        0,
        1,
        int64(utils.CurrentTimeMillis()),
        []byte{1},
    })

    return false
}

func (impl *SyncImpl) getBoard(connection net.Conn, message *Message) bool {
    client := impl.clients.getClient(connection)
    if client == nil { return true }

    var id int32
    copy(unsafe.Slice((*byte) (unsafe.Pointer(&(id))), 4), unsafe.Slice(&(message.body[0]), 4))

    board := impl.db.GetBoard(client.Username, id)

    if board == nil {
        impl.network.sendMessage(connection, &Message{
            flagGetBoard,
            0,
            1,
            int64(utils.CurrentTimeMillis()),
            nil,
        })
    } else {
        impl.network.sendMessage(connection, &Message{
            flagGetBoard,
            0,
            1,
            int64(utils.CurrentTimeMillis()),
            impl.packBoard(board),
        })
    }

    return false
}

func (impl *SyncImpl) getBoards(connection net.Conn) bool {
    client := impl.clients.getClient(connection)
    if client == nil { return true }

    boards := impl.db.GetBoards(client.Username)

    if len(boards) == 0 {
        impl.network.sendMessage(connection, &Message{
            flagGetBoards,
            0,
            1,
            int64(utils.CurrentTimeMillis()),
            nil,
        })
    } else {
        var index int32 = 0
        timestamp := int64(utils.CurrentTimeMillis())

        for _, board := range boards {
            impl.network.sendMessage(connection, &Message{
                flagGetBoards,
                index,
                int32(len(boards)),
                timestamp,
                impl.packBoard(board),
            })
            index++
        }
    }

    return false
}

func (impl *SyncImpl) deleteBoard(connection net.Conn, message *Message) bool {
    client := impl.clients.getClient(connection)
    if client == nil { return true }

    var id int32
    copy(unsafe.Slice((*byte) (unsafe.Pointer(&(id))), 4), unsafe.Slice(&(message.body[0]), 4))

    var result []byte
    if impl.db.RemoveBoard(client.Username, id) {
        result = []byte{1}
    } else {
        result = nil
    }

    impl.network.sendMessage(connection, &Message{
        flagDeleteBoard,
        0,
        1,
        int64(utils.CurrentTimeMillis()),
        result,
    })

    return false
}

func (impl *SyncImpl) pointsSet(connection net.Conn, message *Message) bool {
    client := impl.clients.getClient(connection)
    if client == nil { return true }

    bytes := impl.processPendingMessages(connection, message)
    if bytes == nil { return false }

    impl.db.AddElement(database.Element{database.ElementPointsSet, bytes}, client.board)
    return false
}

func (impl *SyncImpl) line(connection net.Conn, message *Message) bool {
    client := impl.clients.getClient(connection)
    if client == nil { return true }

    bytes := impl.processPendingMessages(connection, message)
    if bytes == nil { return false }

    impl.db.AddElement(database.Element{database.ElementLine, bytes}, client.board)
    return false
}

func (impl *SyncImpl) text(connection net.Conn, message *Message) bool {
    client := impl.clients.getClient(connection)
    if client == nil { return true }

    bytes := impl.processPendingMessages(connection, message)
    if bytes == nil { return false }

    impl.db.AddElement(database.Element{database.ElementText, bytes}, client.board)
    return false
}

func (impl *SyncImpl) image(connection net.Conn, message *Message) bool {
    client := impl.clients.getClient(connection)
    if client == nil { return true }

    bytes := impl.processPendingMessages(connection, message)
    if bytes == nil { return false }

    impl.db.AddElement(database.Element{database.ElementImage, bytes}, client.board)
    return false
}

func (impl *SyncImpl) undo(connection net.Conn) bool {
    client := impl.clients.getClient(connection)
    if client == nil { return true }

    impl.db.RemoveLastElement(client.board)
    return false
}

func (impl *SyncImpl) clear(connection net.Conn) bool {
    client := impl.clients.getClient(connection)
    if client == nil { return true }

    impl.db.RemoveAllElements(client.board)
    return false
}

func (impl *SyncImpl) selectBoard(connection net.Conn, message *Message) bool {
    client := impl.clients.getClient(connection)
    if client == nil { return true }

    var id int32
    copy(unsafe.Slice((*byte) (unsafe.Pointer(&(id))), 4), unsafe.Slice(&(message.body[0]), 4))

    client.board = id

    return false
}

func (impl *SyncImpl) boardElements(connection net.Conn) bool {
    client := impl.clients.getClient(connection)
    if client == nil { return true }

    for _, element := range impl.db.GetElements(client.board) {
        bytes := make([]byte, 4 + len(element.Bytes))
        copy(unsafe.Slice(&(bytes[0]), 4), unsafe.Slice((*byte) (unsafe.Pointer(&(element.Type))), 4))
        copy(unsafe.Slice(&(bytes[4]), len(element.Bytes)), element.Bytes)

        impl.sendBytes(connection, bytes, flagGetBoardElements)
    }

    impl.network.sendMessage(connection, &Message{
        flagGetBoardElements,
        0,
        1,
        int64(utils.CurrentTimeMillis()),
        nil,
    })

    return false
}

func (impl *SyncImpl) routeMessage(connection net.Conn, message *Message) bool {
    disconnect := false

    switch message.flag {
        case flagLogIn:
            disconnect = impl.logIn(connection, message)
        case flagRegister:
            disconnect = impl.register(connection, message)
        case flagShutdown:
            disconnect = impl.shutdown(connection)
        case flagCreateBoard:
            disconnect = impl.createBoard(connection, message)
        case flagGetBoard:
            disconnect = impl.getBoard(connection, message)
        case flagGetBoards:
            disconnect = impl.getBoards(connection)
        case flagDeleteBoard:
            disconnect = impl.deleteBoard(connection, message)
        case flagPointsSet:
            disconnect = impl.pointsSet(connection, message)
        case flagLine:
            disconnect = impl.line(connection, message)
        case flagText:
            disconnect = impl.text(connection, message)
        case flagImage:
            disconnect = impl.image(connection, message)
        case flagUndo:
            disconnect = impl.undo(connection)
        case flagClear:
            disconnect = impl.clear(connection)
        case flagSelectBoard:
            disconnect = impl.selectBoard(connection, message)
        case flagGetBoardElements:
            disconnect = impl.boardElements(connection)
    }

    if disconnect {
        impl.clients.removeClient(connection)
    }

    return disconnect
}

func (impl *SyncImpl) clientDisconnected(connection net.Conn) {
    impl.clients.removeClient(connection)
}
