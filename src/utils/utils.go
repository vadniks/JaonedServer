
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
