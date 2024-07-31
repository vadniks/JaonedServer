
package database

import (
    "JaonedServer/utils"
    "database/sql"
    _ "github.com/lib/pq"
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
    Title []byte
}

type ElementType int32

type Element struct {
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
    RemoveAllElements(board int32, username []byte)
}

type DatabaseImpl struct {
    db *sql.DB
}

var initialized = false

func Init() Database {
    utils.Assert(!initialized)
    initialized = true

    db, err := sql.Open("postgres", "postgres://postgres:postgres@localhost:5432/db?sslmode=disable")
    utils.Assert(err == nil)

    utils.Assert(db.Ping() == nil)

    _, err = db.Exec(`
        create table if not exists users(
            username bytea not null,
            password bytea not null,
            admin int not null,
            primary key(username)
        )
    `)
    if err != nil { println(err.Error()) }
    utils.Assert(err == nil)

    _, err = db.Exec(`
        create table if not exists boards(
            id serial not null,
            color int not null,
            title bytea,
            primary key(id)
        )
    `)
    if err != nil { println(err.Error()) }
    utils.Assert(err == nil)

    _, err = db.Exec(`
        create table if not exists userAndBoard(
            username bytea not null,
            boardId int not null,
            foreign key(username) references users(username) on delete cascade,
            foreign key(boardId) references boards(id) on delete cascade,
            primary key(username, boardId)
        )
    `)
    if err != nil { println(err.Error()) }
    utils.Assert(err == nil)

    _, err = db.Exec(`
        create table if not exists elements(
            id serial not null,
            type int not null,
            bytes bytea not null,
            primary key(id)
        )
    `)
    if err != nil { println(err.Error()) }
    utils.Assert(err == nil)

    return &DatabaseImpl{
        db,
    }
}

func (impl *DatabaseImpl) Close() {
    utils.Assert(impl.db.Close() == nil)
}

func (impl *DatabaseImpl) FindUser(username []byte) *User { // nillable
    //utils.Assert(impl.db.Query("select ") == nil)
    return nil
}

func (impl *DatabaseImpl) AddUser(username []byte, password []byte) bool {
    return false
}

func (impl *DatabaseImpl) RemoveUser(username []byte) bool {
    return false
}

func (impl *DatabaseImpl) GetAllUsers() []*User {
    return nil
}

func (impl *DatabaseImpl) UserExists(username []byte) bool {
    return false
}

func (impl *DatabaseImpl) AddBoard(username []byte, board *Board) {

}

func (impl *DatabaseImpl) GetBoard(username []byte, id int32) *Board { // nillable
    return nil
}

func (impl *DatabaseImpl) GetBoards(username []byte) []*Board { // nillable
    return nil
}

func (impl *DatabaseImpl) RemoveBoard(username []byte, id int32) bool {
    return false
}

func (impl *DatabaseImpl) AddElement(element Element, board int32, username []byte) {

}

func (impl *DatabaseImpl) RemoveLastElement(board int32, username []byte) {

}

func (impl *DatabaseImpl) GetElements(board int32, username []byte) []Element {
    return nil
}

func (impl *DatabaseImpl) RemoveAllElements(board int32, username []byte) {

}
