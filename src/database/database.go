
package database

import (
    "JaonedServer/utils"
    "database/sql"
    _ "github.com/lib/pq"
    "reflect"
)

const (
	MaxCredentialSize = 16
    MaxBoardTitleSize = 16
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
}

type DatabaseImpl struct {
    db *sql.DB
    users []*User // TODO: test only
    boards map[[MaxCredentialSize]byte][]*Board // TODO: test only
}

var initialized = false

func Init() Database {
    utils.Assert(!initialized)
    initialized = true

    //db, err := sql.Open("postgres", "postgres://server:server@localhost:5432/db")
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
        0x10101010,
        []byte{'T', 'e', 's', 't', ' ', '1', 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
    }
    adminBoards[1] = &Board{
        1,
        0x25252525,
        []byte{'T', 'e', 's', 't', ' ', '2', 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
    }

    boards := make(map[[16]byte][]*Board)
    boards[[MaxCredentialSize]byte{'a', 'd', 'm', 'i', 'n', 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}] = adminBoards

    return &DatabaseImpl{
        nil,
        users,
        boards,
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

    impl.boards[xUsername] = newBoards
    return len(newBoards) < len(impl.boards[xUsername])
}
