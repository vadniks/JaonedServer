
package database

import (
    "JaonedServer/utils"
    "database/sql"
)

type User struct {
    username string
    password string
}

type Database interface {
    Close()
    FindUser(username string, password string) *User // nillable
    AddUser(username string, password string)
    RemoveUser(username string)
    GetAllUsers() []*User
    UserExists(username string) bool
}

type databaseImpl struct {
    db *sql.DB
}

var initialized = false

func Init() Database {
    utils.Assert(!initialized)
    initialized = true

    db, err := sql.Open("postgres", "postgres://server:server@localhost:5432/db")
    utils.Assert(err == nil)

    return &databaseImpl{
        db,
    }
}

func (impl *databaseImpl) Close() {

}

func (impl *databaseImpl) FindUser(username string, password string) *User {
    return nil
}

func (impl *databaseImpl) AddUser(username string, password string) {

}

func (impl *databaseImpl) RemoveUser(username string) {

}

func (impl *databaseImpl) GetAllUsers() []*User {
    return nil
}

func (impl *databaseImpl) UserExists(username string) bool {
    return false
}
