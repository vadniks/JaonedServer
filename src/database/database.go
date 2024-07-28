
package database

import (
    "JaonedServer/utils"
    "database/sql"
    _ "github.com/lib/pq"
    "reflect"
    "unsafe"
)

const (
	MaxCredentialSize = 16
    MaxBoardTitleSize = 16
)

const ( // TODO: test only
    ElementPointsSet ElementType = 0
    ElementLine ElementType = 1
    ElementText ElementType = 2
    ElementImage ElementType = 3
)

type User struct {
    Username []byte
    Password []byte
    IsAdmin bool
}

type Board struct {
    Id int32
    Color int32
    // Size int32
    Title []byte
}

type ElementType int32 // TODO: test only

type Element struct { // TODO: test only
    Type ElementType
    Bytes []byte
}

type Database interface {
    Close()

    FindUser(username []byte) *User // nillable
    AddUser(username []byte, password []byte) bool
    RemoveUser(username []byte) bool
    GetAllUsers() []*User
    UserExists(username []byte) bool

    AddBoard(username []byte, board *Board)
    GetBoard(username []byte, id int32) *Board // nillable
    GetBoards(username []byte) []*Board // nillable
    RemoveBoard(username []byte, id int32) bool

    AddElement(element Element, board int32, username []byte)
    RemoveLastElement(board int32, username []byte)
    GetElements(board int32, username []byte) []Element
}

type DatabaseImpl struct {
    db *sql.DB
    users []*User // TODO: test only
    boards map[[MaxCredentialSize]byte][]*Board // TODO: test only
    elements map[[MaxCredentialSize + 4]byte][]Element // TODO: test only
}

var initialized = false

func Init() Database {
    utils.Assert(!initialized)
    initialized = true

    //db, err := sql.Open("postgres", "postgres://server:server@localhost:5432/db") // TODO
    //utils.Assert(err == nil)

    users := make([]*User, 2)
    users[0] = &User{
        []byte{'a', 'd', 'm', 'i', 'n', 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
        []byte{'p', 'a', 's', 's', 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
        true,
    }
    users[1] = &User{
        []byte{'u', 's', 'e', 'r', 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
        []byte{'p', 'a', 's', 's', 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
        false,
    }

    adminBoards := make([]*Board, 2)
    adminBoards[0] = &Board{
        0,
        0x7f101010,
        []byte{'T', 'e', 's', 't', ' ', '1', 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
    }
    adminBoards[1] = &Board{
        1,
        0x7f252525,
        []byte{'T', 'e', 's', 't', ' ', '2', 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
    }

    boards := make(map[[16]byte][]*Board)
    boards[[MaxCredentialSize]byte{'a', 'd', 'm', 'i', 'n', 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}] = adminBoards

    return &DatabaseImpl{
        nil,
        users,
        boards,
        make(map[[MaxCredentialSize + 4]byte][]Element),
    }
}

func (impl *DatabaseImpl) Close() {

}

func (impl *DatabaseImpl) FindUser(username []byte) *User { // TODO: stub
    for _, user := range impl.users {
        if reflect.DeepEqual(username, user.Username) {
            return user
        }
    }
    return nil
}

func (impl *DatabaseImpl) AddUser(username []byte, password []byte) bool { // TODO: stub
    if impl.UserExists(username) { return false }

    found := false

    for _, user := range impl.users {
        if reflect.DeepEqual(user.Username, username) { found = true }
    }

    if found { return false }

    impl.users = append(impl.users, &User{
        username,
        password,
        false,
    })

    return true
}

func (impl *DatabaseImpl) RemoveUser(username []byte) bool {
    return false
}

func (impl *DatabaseImpl) GetAllUsers() []*User {
    return nil
}

func (impl *DatabaseImpl) UserExists(username []byte) bool { // TODO: stub
    return impl.FindUser(username) != nil
}

func (impl *DatabaseImpl) AddBoard(username []byte, board *Board) { // TODO: stub
    xUsername := [MaxCredentialSize]byte(username)
    if impl.boards[xUsername] == nil { impl.boards[xUsername] = make([]*Board, 0) }

    var maxId int32 = 0
    for _, board := range impl.boards[xUsername] {
        if maxId < board.Id { maxId = board.Id }
    }

    board.Id = maxId + 1
    impl.boards[xUsername] = append(impl.boards[xUsername], board)
}

func (impl *DatabaseImpl) GetBoard(username []byte, id int32) *Board { // nillable // TODO: stub
    xUsername := [16]byte(username)
    if impl.boards[xUsername] != nil {
        for _, board := range impl.boards[xUsername] {
            if board.Id == id { return board }
        }
        return nil
    } else {
        return nil
    }
}

func (impl *DatabaseImpl) GetBoards(username []byte) []*Board { // nillable // TODO: stub
    xUsername := [16]byte(username)
    return impl.boards[xUsername]
}

func (impl *DatabaseImpl) RemoveBoard(username []byte, id int32) bool { // TODO: stub
    xUsername := [16]byte(username)
    var newBoards []*Board

    if impl.boards[xUsername] == nil { return false }

    for _, board := range impl.boards[xUsername] {
        if board.Id != id {
            newBoards = append(newBoards, board)
        }
    }

    previousLength := len(impl.boards[xUsername])
    impl.boards[xUsername] = newBoards
    return len(newBoards) < previousLength
}

func (impl *DatabaseImpl) AddElement(element Element, board int32, username []byte) { // TODO: stub
    key := [MaxCredentialSize + 4]byte{}
    copy(unsafe.Slice(&(key[0]), MaxCredentialSize), username)
    copy(unsafe.Slice(&(key[MaxCredentialSize]), 4), unsafe.Slice((*byte) (unsafe.Pointer(&(board))), 4))

    impl.elements[key] = append(impl.elements[key], element)
}

func (impl *DatabaseImpl) RemoveLastElement(board int32, username []byte) { // TODO: stub
    key := [MaxCredentialSize + 4]byte{}
    copy(unsafe.Slice(&(key[0]), MaxCredentialSize), username)
    copy(unsafe.Slice(&(key[MaxCredentialSize]), 4), unsafe.Slice((*byte) (unsafe.Pointer(&(board))), 4))

    impl.elements[key] = impl.elements[key][:(len(impl.elements[key]) - 1)]
}

func (impl *DatabaseImpl) GetElements(board int32, username []byte) []Element { // TODO: stub
    key := [MaxCredentialSize + 4]byte{}
    copy(unsafe.Slice(&(key[0]), MaxCredentialSize), username)
    copy(unsafe.Slice(&(key[MaxCredentialSize]), 4), unsafe.Slice((*byte) (unsafe.Pointer(&(board))), 4))

    return impl.elements[key]
}
