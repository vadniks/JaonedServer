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

package utils

import "time"

type Triple int8

const (
    Positive Triple = 1
    Neutral Triple = 0
    Negative Triple = -1
)

func Assert(condition bool) { if !condition { panic(any("")) } }
func CurrentTimeMillis() uint64 { return uint64(time.Now().UnixMilli()) }
