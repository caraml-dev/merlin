package expression

import (
	"sync"

	"github.com/antonmedv/expr/vm"
)

type Storage struct {
	syncMap sync.Map
}

func NewStorage() *Storage {
	return &Storage{}
}

func (c *Storage) Get(expression string) *vm.Program {
	o, ok := c.syncMap.Load(expression)
	if !ok {
		return nil
	}

	return o.(*vm.Program)
}

func (c *Storage) Set(expression string, compiledExpression *vm.Program) {
	c.syncMap.Store(expression, compiledExpression)
}

func (c *Storage) AddAll(expressionMap map[string]*vm.Program) {
	for k, v := range expressionMap {
		c.Set(k, v)
	}
}
