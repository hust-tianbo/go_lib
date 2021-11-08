package bluele_cache

import (
	"time"

	bl "github.com/bluele/gcache"
)

var defaultCapacity = 4096

type blueleCache struct {
	D bl.Cache
}

func NewBlueleCacheLRU() *blueleCache {
	return &blueleCache{D: bl.New(defaultCapacity).LRU().Build()}
}

func NewBuleleCacheLRUWithCapacity(capacity int) *blueleCache {
	return &blueleCache{D: bl.New(capacity).LRU().Build()}
}

func (br *blueleCache) Add(key, v interface{}) error {
	return br.D.Set(key, v)
}

func (br *blueleCache) AddWithExpire(key, v interface{}, lifeSpan time.Duration) error {
	return br.D.SetWithExpire(key, v, lifeSpan)
}

func (br *blueleCache) Get(key interface{}) (interface{}, bool) {
	v, err := br.D.Get(key.(string))
	if err != nil {
		return nil, false
	}
	return v, true
}

func (br *blueleCache) Remove(key interface{}) bool {
	ok := br.D.Remove(key)
	return ok
}

func (br *blueleCache) Len() int {
	return br.D.Len(false)
}

func (br *blueleCache) Clear() {
	br.D.Purge()
}
