// Copyright 2020 VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package texttemplate

import (
	"fmt"

	"github.com/k14s/ytt/pkg/filepos"
	"github.com/k14s/ytt/pkg/template"
)

type NodeRoot struct {
	Items []interface{}

	annotations interface{}
}

type NodeText struct {
	Position *filepos.Position
	Content  string

	annotations interface{}
	startOffset int
}

type NodeCode struct {
	Position *filepos.Position
	Content  string

	startOffset int
}

// NodeCode is not a evaluation node
var _ = []template.EvaluationNode{&NodeRoot{}, &NodeText{}}

func (n *NodeRoot) AsString() string {
	var result string
	for _, item := range n.Items {
		switch typedItem := item.(type) {
		case *NodeText:
			result += typedItem.Content
		default:
			panic(fmt.Sprintf("unknown node type %T", typedItem))
		}
	}
	return result
}

func (n *NodeRoot) GetValues() []interface{} {
	var result []interface{}
	for _, item := range n.Items {
		result = append(result, item)
	}
	return result
}

func (n *NodeRoot) SetValue(val interface{}) error {
	return fmt.Errorf("cannot set value on a noderoot")
}
func (n *NodeRoot) AddValue(val interface{}) error { n.Items = append(n.Items, val); return nil }
func (n *NodeRoot) ResetValue()                    { n.Items = nil }

func (n *NodeRoot) DeepCopyAsInterface() interface{} {
	var newItems []interface{}
	for _, item := range n.Items {
		if typedItem, ok := item.(interface{ DeepCopyAsInterface() interface{} }); ok {
			newItems = append(newItems, typedItem.DeepCopyAsInterface())
		} else {
			panic(fmt.Sprintf("unknown node type %T", typedItem))
		}
	}
	return &NodeRoot{Items: newItems, annotations: annotationsDeepCopy(n.annotations)}
}

func (n *NodeRoot) GetAnnotations() interface{}     { return n.annotations }
func (n *NodeRoot) SetAnnotations(anns interface{}) { n.annotations = anns }

func (n *NodeText) GetValues() []interface{} { return []interface{}{n.Content} }

func (n *NodeText) SetValue(val interface{}) error {
	if typedVal, ok := val.(string); ok {
		n.Content = typedVal
		return nil
	}
	return fmt.Errorf("cannot set non-string value (%T), consider using str(...) to convert to string", val)
}

func (n *NodeText) AddValue(val interface{}) error { n.Content = val.(string); return nil }
func (n *NodeText) ResetValue()                    { n.Content = "" }

func (n *NodeText) DeepCopyAsInterface() interface{} {
	return &NodeText{Position: n.Position, Content: n.Content, annotations: annotationsDeepCopy(n.annotations)}
}

func (n *NodeText) GetAnnotations() interface{}     { return n.annotations }
func (n *NodeText) SetAnnotations(anns interface{}) { n.annotations = anns }

func annotationsDeepCopy(anns interface{}) interface{} {
	if anns == nil {
		return nil
	}
	return anns.(interface{ DeepCopyAsInterface() interface{} }).DeepCopyAsInterface()
}
