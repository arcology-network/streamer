package actor

import (
	scommon "github.com/arcology-network/streamer/common"
)

type BusinessChain struct {
	nodes   []*ControlledNode
	filters []*Filter // Pre-Filter（chain-level）
}

func NewBusinessChain(nodes []*ControlledNode, filters []*Filter) *BusinessChain {
	return &BusinessChain{
		nodes:   nodes,
		filters: filters,
	}
}

func (c *BusinessChain) GetNode(idx int) *ControlledNode {
	if idx >= len(c.nodes) || idx < 0 {
		return nil
	}
	return c.nodes[idx]
}

func (c *BusinessChain) OnMessages(msgs []*scommon.Message, execCtx *ExecutionContext) error {
	filtered := msgs

	// 1. Pre-Filter Chain
	for _, f := range c.filters {
		filtered = f.Apply(filtered)
		if filtered == nil || len(filtered) == 0 {
			return nil
		}
	}

	// 2. Explicit and deterministic business execution path
	for _, node := range c.nodes {
		err := node.Handle(msgs, execCtx)
		if err != nil {
			return err
		}
	}
	return nil
}
