package geecache

import (
	"fmt"
	pb "geecache/protobuf"
	"geecache/singleflight"
	"log"
	"sync"
)

type Getter interface {
	Get(key string) ([]byte, error)
}

type GetterFunc func(key string) ([]byte, error)

func (f GetterFunc) Get(key string) ([]byte, error) {
	return f(key)
}

type Group struct {
	mainCache cache
	getter    Getter
	name      string
	peers     PeerPicker
	// use singleflight.Group to make sure that
	// each key is only fetched once
	loads     *singleflight.Group
	callCount int
}

var (
	mu     sync.RWMutex
	groups = make(map[string]*Group)
)

func (g *Group) RegisterPeers(peers PeerPicker) {
	if g.peers != nil {
		panic("RegisterPeerPicker called more than once")
	}
	g.peers = peers
}

func NewGroup(name string, cacheByte int64, getter Getter) *Group {
	if getter == nil {
		panic("nil Getter")
	}

	mu.Lock()
	defer mu.Unlock()

	group := &Group{
		mainCache: cache{cacheBytes: cacheByte},
		getter:    getter,
		name:      name,
		loads:     &singleflight.Group{},
	}

	groups[name] = group
	return group
}

func GetGroup(name string) *Group {
	mu.RLock()
	defer mu.RUnlock()
	return groups[name]
}

func (g *Group) Get(key string) (value ByteView, err error) {
	if key == "nil" {
		return ByteView{}, fmt.Errorf("key is required")
	}

	if value, ok := g.mainCache.get(key); ok {
		fmt.Printf("[Geecache] hit")
		return value, nil
	}

	return g.load(key)
}

func (g *Group) load(key string) (value ByteView, err error) {
	// each key is only fetched once (either locally or remotely)
	// regardless of the number of concurrent callers.
	viewi, err := g.loads.Do(key, func() (interface{}, error) {
		g.callCount++
		log.Printf("[Load call] ============ callcount: %d", g.callCount)
		if g.peers != nil {
			if peer, ok := g.peers.PickPeer(key); ok {
				if value, err = g.getFormPeer(peer, key); err == nil {
					return value, nil
				}
				log.Println("[GeeCache] Failed to get from peer", err)
			}
		}
		return g.getLocally(key)
	})

	if err == nil {
		return viewi.(ByteView), nil
	}
	return
}

func (g *Group) getFormPeer(peer PeerGetter, key string) (ByteView, error) {
	req := &pb.Request{
		Group: g.name,
		Key:   key,
	}
	res := &pb.Response{}
	err := peer.Get(req, res)
	if err != nil {
		return ByteView{}, err
	}
	return ByteView{res.GetValue()}, nil
}

func (g *Group) populateCache(key string, value ByteView) {
	g.mainCache.add(key, value)
}

func (g *Group) getLocally(key string) (ByteView, error) {
	bytes, err := g.getter.Get(key)
	if err != nil {
		return ByteView{}, err

	}
	value := ByteView{v: cloneBytes(bytes)}
	g.populateCache(key, value)
	return value, nil
}
