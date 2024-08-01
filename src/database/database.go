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

const (
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

type Database interface { // true - success
    Close()

    FindUser(username []byte) *User // nillable
    AddUser(username []byte, password []byte) bool
    RemoveUser(username []byte) bool

    AddBoard(username []byte, board *Board) bool
    GetBoard(username []byte, id int32) *Board // nillable
    GetBoards(username []byte) []*Board // nillable
    RemoveBoard(username []byte, id int32) bool

    AddElement(element Element, board int32) bool
    RemoveLastElement(board int32) bool
    GetElements(board int32) []*Element // nillable
    RemoveAllElements(board int32) bool
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
            boardId int not null,
            timestamp bigint not null,
            foreign key(boardId) references boards(id) on delete cascade,
            primary key(id)
        )
    `)
    if err != nil { println(err.Error()) }
    utils.Assert(err == nil)

    impl := &DatabaseImpl{db}

    if impl.FindUser([]byte{'a', 'd', 'm', 'i', 'n', 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}) == nil {
        impl.AddUser(
            []byte{'a', 'd', 'm', 'i', 'n', 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
            []byte{'p', 'a', 's', 's', 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
        )
    }

    return impl
}

func (impl *DatabaseImpl) Close() {
    utils.Assert(impl.db.Close() == nil)
}

func (impl *DatabaseImpl) FindUser(username []byte) *User { // nillable
    row := impl.db.QueryRow("select * from users where username = $1", username)

    user := &User{}
    if row.Scan(&(user.Username), &(user.Password), &(user.IsAdmin)) != nil { return nil }

    return user
}

func (impl *DatabaseImpl) AddUser(username []byte, password []byte) bool {
    _, err := impl.db.Exec("insert into users(username, password, admin) values($1, $2, $3)", username, password, 0)
    return err == nil
}

func (impl *DatabaseImpl) RemoveUser(username []byte) bool {
    _, err := impl.db.Exec("delete from users where username = $1", username)
    return err == nil
}

func (impl *DatabaseImpl) AddBoard(username []byte, board *Board) bool {
    row := impl.db.QueryRow("insert into boards(color, title) values($1, $2) returning id", board.Color, board.Title)
    var boardId int32
    if row.Scan(&boardId) != nil { return false }

    _, err := impl.db.Exec("insert into userAndBoard(username, boardId) values($1, $2)", username, boardId)
    return err == nil
}

func (impl *DatabaseImpl) GetBoard(username []byte, id int32) *Board { // nillable
    row := impl.db.QueryRow("select b.id, b.color, b.title from boards b inner join userAndBoard uab on b.id = uab.boardId where uab.username = $1 and b.id = $2", username, id)

    board := &Board{}
    if row.Scan(&(board.Id), &(board.Color), &(board.Title)) != nil { return nil }

    return board
}

func (impl *DatabaseImpl) GetBoards(username []byte) []*Board { // nillable
    rows, err := impl.db.Query("select b.id, b.color, b.title from boards b inner join userAndBoard uab on b.id = uab.boardId where uab.username = $1", username)
    if err != nil { return nil }

    boards := make([]*Board, 0)

    for rows.Next() {
        board := &Board{}
        if rows.Scan(&(board.Id), &(board.Color), &(board.Title)) != nil { return nil }
        boards = append(boards, board)
    }

    return boards
}

func (impl *DatabaseImpl) RemoveBoard(username []byte, id int32) bool {
    _, err := impl.db.Exec("delete from boards where id = $1", id)
    if err != nil { return false }

    _, err = impl.db.Exec("delete from userAndBoard where username = $1 and boardId = $2", username, id)
    return err == nil
}

func (impl *DatabaseImpl) AddElement(element Element, board int32) bool {
    _, err := impl.db.Exec("insert into elements(type, bytes, boardId, timestamp) values($1, $2, $3, $4)", element.Type, element.Bytes, board, utils.CurrentTimeMillis())
    return err == nil
}

func (impl *DatabaseImpl) RemoveLastElement(board int32) bool {
    _, err := impl.db.Exec("delete from elements where timestamp = (select max(timestamp) from elements)")
    return err == nil
}

func (impl *DatabaseImpl) GetElements(board int32) []*Element { // nillable
    rows, err := impl.db.Query("select type, bytes from elements where boardId = $1", board)
    if err != nil { return nil }

    elements := make([]*Element, 0)

    for rows.Next() {
        element := &Element{}
        if rows.Scan(&(element.Type), &(element.Bytes)) != nil { return nil }
        elements = append(elements, element)
    }

    return elements
}

func (impl *DatabaseImpl) RemoveAllElements(board int32) bool {
    _, err := impl.db.Exec("delete from elements where boardId = $1", board)
    return err == nil
}
