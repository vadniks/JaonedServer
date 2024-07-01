
package database

import (
    "JaonedServer/utils"
    "database/sql"
    _ "github.com/lib/pq"
    "reflect"
)

type User struct {
    Username []byte
    Password []byte
}

type Database interface {
    Close()
    FindUser(username []byte) *User // nillable
    AddUser(username []byte, password []byte) bool
    RemoveUser(username []byte) bool
    GetAllUsers() []*User
    UserExists(username []byte) bool
}

type DatabaseImpl struct {
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
        []byte{'a', 'd', 'm', 'i', 'n', 0, 0, 0},
        []byte{'p', 'a', 's', 's', 0, 0, 0, 0},
    }
    users[1] = &User{
        []byte{'u', 's', 'e', 'r', 0, 0, 0, 0},
        []byte{'p', 'a', 's', 's', 0, 0, 0, 0},
    }

    return &DatabaseImpl{
        nil,
        users,
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
