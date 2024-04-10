package lru

import "container/list"

// Cache v1 LRU缓存，不是并发安全的
type Cache struct {
	maxBytes int64      // 允许的最大内存
	nbytes   int64      // 目前已经使用的内存
	ll       *list.List // 直接使用标准库中的双向链表
	cache    map[string]*list.Element
	// 可选项在清除entry时执行
	OnEvicted func(key string, value Value) // 某条记录被移除时的回调函数
}

// 双向链表节点的数据类型, 仍保存每个值对应的key的好处在于, 淘汰队首节点时, 只需要用key从字典中删除对应的映射
type entry struct {
	key   string
	value Value
}

// Value 使用Len计算所需字节数
type Value interface {
	Len() int
}

// New 新建一个Cache实例
func New(maxBytes int64, onEvicted func(string, Value)) *Cache {
	return &Cache{
		maxBytes:  maxBytes,
		ll:        list.New(),
		cache:     make(map[string]*list.Element),
		OnEvicted: onEvicted,
	}
}

// Get 查找功能
/*
查找功能主要是两个步骤
1. 从字典中找到对应的双向链表的节点
2. 将该节点移动到队尾
这样做的目的是适用LRU,最近使用的节点会被移动到队尾,而最久未使用的节点会被移除
*/
func (c *Cache) Get(key string) (value Value, ok bool) {
	if ele, ok := c.cache[key]; ok {
		// 将此项移到队尾
		c.ll.MoveToBack(ele)
		// 拿到节点的信息返回
		kv := ele.Value.(*entry)
		return kv.value, true
	}
	return
}

// RemoveOldest 删除最近最少访问的节点
func (c *Cache) RemoveOldest() {
	// 取到队首节点
	ele := c.ll.Back()
	if ele != nil {
		c.ll.Remove(ele)
		kv := ele.Value.(*entry)
		delete(c.cache, kv.key)
		// 调整已经占用的内存
		c.nbytes -= int64(len(kv.key)) + int64(kv.value.Len())
		// 如果回调函数不为空, 则调用回调函数
		if c.OnEvicted != nil {
			c.OnEvicted(kv.key, kv.value)
		}
	}
}

// Add 添加/修改功能
func (c *Cache) Add(key string, value Value) {
	if ele, ok := c.cache[key]; ok {
		// 如果存在, 则更新值并将该节点移到队尾
		c.ll.MoveToBack(ele)
		kv := ele.Value.(*entry)
		// 更新内存占用
		c.nbytes += int64(value.Len()) - int64(kv.value.Len())
		// 更新节点的值
		kv.value = value
	} else {
		// 如果不存在, 则插入新节点
		ele := c.ll.PushFront(&entry{key, value})
		// 更新内存占用
		c.cache[key] = ele
		c.nbytes += int64(len(key)) + int64(value.Len())
	}
	// 如果超出了最大内存, 则移除最少访问的节点
	for c.maxBytes != 0 && c.nbytes > c.maxBytes {
		c.RemoveOldest()
	}
}

// Len 获取当前缓存的节点数
func (c *Cache) Len() int {
	return c.ll.Len()
}
