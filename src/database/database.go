
package database

import (
    "JaonedServer/utils"
    "database/sql"
    _ "github.com/lib/pq"
    "reflect"
)

type User struct {
    Id int32
    Username []byte
    Password []byte
}

type Database interface {
    Close()
    FindUser(username []byte) *User // nillable
    AddUser(username []byte, password []byte)
    RemoveUser(username []byte)
    GetAllUsers() []*User
    UserExists(username []byte) bool
}

type databaseImpl struct {
    db *sql.DB
    users []*User // TODO: test only
}

var initialized = false

func Init() Database {
    utils.Assert(!initialized)
    initialized = true

    //db, err := sql.Open("postgres", "postgres://server:server@localhost:5432/db")
    //utils.Assert(err == nil)

    users := make([]*User, 2)
    users[0] = &User{
        0,
        []byte{'a', 'd', 'm', 'i', 'n', 0, 0, 0},
        []byte{'p', 'a', 's', 's', 0, 0, 0, 0},
    }
    users[0] = &User{
        1,
        []byte{'u', 's', 'e', 'r', 0, 0, 0, 0},
        []byte{'p', 'a', 's', 's', 0, 0, 0, 0},
    }

    return &databaseImpl{
        nil,
        users,
    }
}

func (impl *databaseImpl) Close() {

}

func (impl *databaseImpl) FindUser(username []byte) *User {
    for _, user := range impl.users {
        if reflect.DeepEqual(username, user.Username) {
            return user
        }
    }
    return nil
}

func (impl *databaseImpl) AddUser(username []byte, password []byte) {

}

func (impl *databaseImpl) RemoveUser(username []byte) {

}

func (impl *databaseImpl) GetAllUsers() []*User {
    return nil
}

func (impl *databaseImpl) UserExists(username []byte) bool {
    return false
}
