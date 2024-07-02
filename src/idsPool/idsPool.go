
package idsPool

import (
    "JaonedServer/utils"
    "math"
    "math/big"
    "sync"
    )

type IdsPool interface {
    SetId(id int32, taken bool)
    TakeId() *int32
    ReturnId(id int32)
}

type IdsPoolImpl struct {
    size int32
    ids big.Int
    mutex sync.Mutex
}

func InitIdsPool(size int32) IdsPool {
    impl := &IdsPoolImpl{size, big.Int{}, sync.Mutex{}}

    impl.ids.SetBit(&(impl.ids), int(size), 1)

    utils.Assert(
        float64(len(impl.ids.Bytes())) == math.Ceil(float64(size) / float64(8)) &&
        impl.ids.Bit(0) == 0,
    )

    return impl
}

func (impl *IdsPoolImpl) SetId(id int32, taken bool) {
    impl.mutex.Lock()

    utils.Assert(id < impl.size)
    impl.ids.SetBit(&(impl.ids), int(id), uint(func() byte { if taken { return 1 } else { return 0 } }()))

    impl.mutex.Unlock()
}

func (impl *IdsPoolImpl) TakeId() *int32 { // nillable result
    impl.mutex.Lock()

    for i := int32(0); i < impl.size; i++ {
        if impl.ids.Bit(int(i)) == 0 {
            impl.ids.SetBit(&(impl.ids), int(i), 1)
            impl.mutex.Unlock()
            return &i
        }
    }

    impl.mutex.Unlock()
    return nil
}

func (impl *IdsPoolImpl) ReturnId(id int32) {
    impl.mutex.Lock()

    utils.Assert(id < impl.size && impl.ids.Bit(int(id)) == 1)
    impl.ids.SetBit(&(impl.ids), int(id), 0)

    impl.mutex.Unlock()
}
