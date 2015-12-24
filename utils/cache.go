package utils

import (
	"sync"
	"time"
)

var Cache = &cache{caches: make(map[string]*expirableContent)}

type cache struct {
	sync.Mutex
	caches map[string]*expirableContent
}

type expirableContent struct {
	content   interface{}
	expiresAt time.Time
}

func (self *cache) Fetch(key string, dur time.Duration, f func() (interface{}, error)) (interface{}, error) {
	self.Lock()
	defer self.Unlock()

	if res, found := self.caches[key]; found {
		if res.expiresAt.After(time.Now().UTC()) {
			return res.content, nil
		} else {
			delete(self.caches, key)
		}
	}

	content, err := f()
	if err != nil {
		return nil, err
	}

	newContent := &expirableContent{content: content}
	newContent.expiresAt = time.Now().Add(dur)

	self.caches[key] = newContent
	return newContent.content, nil
}
