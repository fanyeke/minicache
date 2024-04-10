package consistenthash

import (
	"hash/crc32"
	"sort"
	"strconv"
)

type Hash func(data []byte) uint32

type Map struct {
	hash     Hash
	replicas int   // 虚拟节点的倍数
	keys     []int // 哈希环
	hashMap  map[int]string
}

func New(replicas int, fn Hash) *Map {
	m := &Map{
		replicas: replicas,
		hash:     fn,
		hashMap:  make(map[int]string),
	}
	if m.hash == nil {
		m.hash = crc32.ChecksumIEEE
	}
	return m
}

// Add 添加真实节点
func (m *Map) Add(keys ...string) {
	for _, key := range keys {
		// 添加虚拟节点
		for i := 0; i < m.replicas; i++ {
			hash := int(m.hash([]byte(strconv.Itoa(i) + key)))
			m.keys = append(m.keys, hash) // 把虚拟节点加入到哈希环当中
			m.hashMap[hash] = key         // 维护映射关系
		}
	}
	sort.Ints(m.keys)
}

// Get 选择节点
func (m *Map) Get(key string) string {
	if len(m.keys) == 0 {
		return ""
	}
	// 获得key的哈希值
	hash := int(m.hash([]byte(key)))
	// 顺时针找到keys值大于哈希值的第一个虚拟节点
	idx := sort.Search(len(m.keys), func(i int) bool {
		return m.keys[i] >= hash
	})
	// 根据映射找到真实节点
	return m.hashMap[m.keys[idx%len(m.keys)]]
}
