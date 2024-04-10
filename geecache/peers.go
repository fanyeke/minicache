package geecache

import pb "minicache/geecache/geecachepb"

// PeerPicker 根据传入的key选择相对应的节点
type PeerPicker interface {
	PickPeer(key string) (peer PeerGetter, ok bool)
}

// PeerGetter 从对应的节点的group中查找缓存值, 对应之前实现http的服务端,这里实现的是http的客户端
type PeerGetter interface {
	Get(in *pb.Request, out *pb.Response) error
}
