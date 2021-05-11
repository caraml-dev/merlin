package jsonpath

import "sync"

type Storage struct {
	syncMap sync.Map
}

func NewStorage() *Storage {
	return &Storage{}
}

func (c *Storage) Get(jsonPath string) *Compiled {
	o, ok := c.syncMap.Load(jsonPath)
	if !ok {
		return nil
	}

	return o.(*Compiled)
}

func (c *Storage) Set(jsonPath string, compiledJsonPath *Compiled) {
	c.syncMap.Store(jsonPath, compiledJsonPath)
}

func (c *Storage) AddAll(jsonPaths map[string]*Compiled) {
	for k, v := range jsonPaths {
		c.Set(k, v)
	}
}
