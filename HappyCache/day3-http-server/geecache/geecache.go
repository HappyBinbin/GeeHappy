package geecache

import (
	"fmt"
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
}

var (
	mu     sync.RWMutex
	groups = make(map[string]*Group)
)

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

func (g *Group) load(key string) (ByteView, error) {

	bytes, err := g.getter.Get(key)
	if err != nil {
		return ByteView{}, err
	}

	value := ByteView{cloneBytes(bytes)}

	g.mainCache.add(key, value)
	return value, nil
}
