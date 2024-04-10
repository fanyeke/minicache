package geecache

import (
	"fmt"
	"log"
	pb "minicache/geecache/geecachepb"
	"minicache/geecache/singleflght"
	"sync"
)

type Getter interface {
	Get(key string) ([]byte, error)
}

type GetterFunc func(key string) ([]byte, error)

func (f GetterFunc) Get(key string) ([]byte, error) {
	return f(key)
}

// Group 一个并发安全的缓存结构
type Group struct {
	name      string // 缓存的名称
	getter    Getter // 缓存未命中时获取源数据的回调
	mainCache cache  // 缓存的实现, 在之前实现的缓存外套了一层name, 便于缓存不同类型的数据
	peers     PeerPicker
	loader    *singleflght.Group
}

var (
	mu     sync.RWMutex
	groups = make(map[string]*Group)
)

// NewGroup 创建一个新的Group实例
func NewGroup(name string, cacheBytes int64, getter Getter) *Group {
	// 函数不能为空
	if getter == nil {
		panic("nil Getter")
	}
	mu.Lock()
	defer mu.Unlock()
	g := &Group{
		name:      name,
		getter:    getter,
		mainCache: cache{cacheBytes: cacheBytes},
		loader:    &singleflght.Group{},
	}
	groups[name] = g
	return g
}

// GetGroup 返回指定名称的Group
func GetGroup(name string) *Group {
	// 只用了读锁, 因为不涉及冲突变量的写操作
	mu.RLock()
	g := groups[name]
	mu.RUnlock()
	return g
}

// Get 实现根据key获取内容的功能
func (g *Group) Get(key string) (ByteView, error) {
	// key值不能为空
	if key == "" {
		return ByteView{}, fmt.Errorf("key is required")
	}
	// 从缓存中获取数据
	if v, ok := g.mainCache.get(key); ok {
		log.Println("[GeeCache] hit")
		return v, nil
	}
	// 如果没有命中, 则调用load方法
	return g.load(key)
}

// load 实现缓存未命中时的回调函数
func (g *Group) load(key string) (value ByteView, err error) {
	// 每个密钥只被获取一次（本地或远程）。
	// 无论有多少并发调用者。
	viewi, err := g.loader.Do(key, func() (interface{}, error) {
		if g.peers != nil {
			if peer, ok := g.peers.PickPeer(key); ok {
				// 从其他节点获取缓存
				if value, err = g.getFromPeer(peer, key); err == nil {
					return value, nil
				}
				log.Println("[GeeCache] Failed to get from peer", err)
			}
		}
		// 其他节点也没找到, 从本地数据库拿去数据
		return g.getLocally(key)
	})

	if err == nil {
		return viewi.(ByteView), nil
	}
	return
}

// getLocally 实现缓存未命中时的回调函数
func (g *Group) getLocally(key string) (ByteView, error) {
	// 调用用户自定义的回调函数
	bytes, err := g.getter.Get(key)
	if err != nil {
		return ByteView{}, err
	}
	// 将获取到的数据封装成ByteView
	value := ByteView{b: cloneBytes(bytes)}
	// 将数据添加到缓存中
	g.populateCache(key, value)
	return value, nil
}

// populateCache 将数据添加到缓存中
func (g *Group) populateCache(key string, value ByteView) {
	g.mainCache.add(key, value)
}

func (g *Group) RegisterPeers(peers PeerPicker) {
	if g.peers != nil {
		panic("RegisterPeerPicker called more than once")
	}
	g.peers = peers
	fmt.Println("成功注册")
}

func (g *Group) getFromPeer(peer PeerGetter, key string) (ByteView, error) {
	req := &pb.Request{
		Group: g.name,
		Key:   key,
	}
	res := &pb.Response{}
	err := peer.Get(req, res)
	if err != nil {
		return ByteView{}, err
	}
	return ByteView{b: res.Value}, nil
}
