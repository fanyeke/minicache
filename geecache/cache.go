package geecache

import (
	"minicache/geecache/lru"
	"sync"
)

// 一个并发安全的缓存结构
type cache struct {
	mu         sync.Mutex // 给LRUCache加了一层,方便实现并发安全
	lru        *lru.Cache // 封装了之前实现的 lru.Cache
	cacheBytes int64      // 缓存最大容量
}

func (c *cache) add(key string, value ByteView) {
	c.mu.Lock()
	defer c.mu.Unlock()
	// 惰性初始化
	if c.lru == nil {
		c.lru = lru.New(c.cacheBytes, nil)
	}
	c.lru.Add(key, value)
}

func (c *cache) get(key string) (value ByteView, ok bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.lru == nil {
		return
	}
	if v, ok := c.lru.Get(key); ok {
		return v.(ByteView), ok
	}
	return
}
