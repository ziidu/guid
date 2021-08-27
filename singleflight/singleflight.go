package singleflight

import "sync"

type call struct {
	wait sync.WaitGroup
	val  interface{}
	err  error
}

type Group struct {
	mu sync.Mutex
	m  map[string]*call
}

func (g *Group) Do(key string, fn func() (interface{}, error)) (interface{}, error) {
	g.mu.Lock()
	if g.m == nil {
		g.m = make(map[string]*call)
	}
	if c, ok := g.m[key]; ok {
		g.mu.Unlock()
		c.wait.Wait()
		return c.val, c.err
	}
	c := new(call)
	c.wait.Add(1)
	g.m[key] = c
	g.mu.Unlock()
	c.val, c.err = fn()
	c.wait.Done()

	g.mu.Lock()
	delete(g.m, key)
	g.mu.Unlock()
	return c.val, c.err
}

func (g *Group) DoTask(key string, fn func()) {
	g.mu.Lock()
	if g.m == nil {
		g.m = make(map[string]*call)
	}
	if c, ok := g.m[key]; ok {
		g.mu.Unlock()
		c.wait.Wait()
		return
	}
	c := new(call)
	c.wait.Add(1)
	g.m[key] = c
	g.mu.Unlock()
	fn()
	c.wait.Done()

	g.mu.Lock()
	delete(g.m, key)
	g.mu.Unlock()
}
