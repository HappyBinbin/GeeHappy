package geecache

import pb "geecache/protobuf"

type PeerPicker interface {
	PickPeer(key string) (PeerGetter, bool)
}

type PeerGetter interface {
	Get(in *pb.Request, out *pb.Response) error
}
