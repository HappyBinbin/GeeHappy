package singleflight

import (
	"log"
	"sync"
)

type call struct {
	w   sync.WaitGroup
	val interface{}
	err error
}

type Group struct {
	mu sync.Mutex
	m  map[string]*call
}

// Do executes and returns the results of the given function, making
// sure that only one execution is in-flight for a given key at a
// time. If a duplicate comes in, the duplicate caller waits for the
// original to complete and receives the same results.
func (g *Group) Do(key string, fn func() (interface{}, error)) (interface{}, error) {
	// lazy load g.m
	g.mu.Lock()
	if g.m == nil {
		g.m = make(map[string]*call)
	}

	// currency call will wait
	if c, ok := g.m[key]; ok {
		c.w.Wait()
		g.mu.Unlock()
		return c.val, c.err
	}

	log.Println("cannot get key cache ======= ")
	// init call, lock waitGroup, unlock mu
	c := new(call)
	c.w.Add(1)
	g.m[key] = c
	g.mu.Unlock()

	// call func and unlock waitGroup
	c.val, c.err = fn()
	c.w.Done()

	g.mu.Lock()
	delete(g.m, key)
	g.mu.Unlock()

	return c.val, c.err
}
