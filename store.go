package mast

import (
	"encoding/json"
	"fmt"
	"reflect"
)

type contentHash string

type link interface {
	Load(m *Mast) (*mastNode, error)
}

type stringNodeT = struct {
	Key   []json.RawMessage
	Value []json.RawMessage
	Link  []contentHash `json:",omitempty"`
}

func defaultUnmarshal(bytes []byte, i interface{}) error {
	return json.Unmarshal(bytes, i)
}

func defaultMarshal(i interface{}) ([]byte, error) {
	return json.Marshal(i)
}

func (m *Mast) load(link link) (*mastNode, error) {
	return link.Load(m)
}

func (l contentHash) Load(m *Mast) (*mastNode, error) {
	nodeBytes, err := m.persist.Load(string(l))
	if err != nil {
		return nil, fmt.Errorf("failed loading %s: %w", l, err)
	}
	var stringNode stringNodeT
	err = m.unmarshal(nodeBytes, &stringNode)
	if err != nil {
		return nil, fmt.Errorf("failed unmarshaling %s: %w", l, err)
	}
	if m.debug {
		fmt.Printf("loaded stringNode: %v\n", stringNode)
	}
	if len(stringNode.Key) != len(stringNode.Value) {
		return nil, fmt.Errorf("cannot unmarshal %s: mismatched keys and values", l)
	}
	node := mastNode{
		make([]interface{}, len(stringNode.Key)),
		make([]interface{}, len(stringNode.Value)),
		make([]link, len(stringNode.Key)+1),
	}
	for i := 0; i < len(stringNode.Key); i++ {
		aType := reflect.TypeOf(m.zeroKey)
		aCopy := reflect.New(aType)
		err := m.unmarshal(stringNode.Key[i], aCopy.Interface())
		if err != nil {
			return nil, fmt.Errorf("cannot unmarshal key[%d] in %s: %w", i, l, err)
		}
		newKey := aCopy.Elem().Interface()

		var newValue interface{}
		if m.zeroValue != nil {
			aType = reflect.TypeOf(m.zeroValue)
			aCopy = reflect.New(aType)
			err = m.unmarshal(stringNode.Value[i], aCopy.Interface())
			if err != nil {
				return nil, fmt.Errorf("cannot unmarshal value[%d] in %s: %w", i, l, err)
			}
			newValue = aCopy.Elem().Interface()
		} else {
			newValue = nil
		}

		node.Key[i] = newKey
		node.Value[i] = newValue
	}
	if stringNode.Link != nil {
		for i, l := range stringNode.Link {
			if l == "" {
				node.Link[i] = nil
			} else {
				node.Link[i] = l
			}
		}
	}
	if m.debug {
		fmt.Printf("loaded node %s->%v\n", l, node)
	}
	return &node, nil
}

func (node *mastNode) Load(m *Mast) (*mastNode, error) {
	return node, nil
}

func (m *Mast) store(node *mastNode) (link, error) {
	validateNode(node, m)
	if len(node.Link) == 1 && node.Link[0] == nil {
		return nil, fmt.Errorf("bug! shouldn't be storing empty nodes")
	}
	return node, nil
}
