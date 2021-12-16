// Copyright 2020 VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package yamlmeta

import "github.com/k14s/ytt/pkg/filepos"

// NewDocumentSet creates a new DocumentSet instance based on the given prototype
func NewDocumentSet(val *DocumentSetProto) *DocumentSet {
	newDocSet := &DocumentSet{
		Comments: NewComments(val.Comments),
		Position: val.Position,
	}
	for _, item := range val.Items {
		newDocSet.Items = append(newDocSet.Items, NewDocument(item))
	}
	return newDocSet
}

// NewDocument creates a new Document instance based on the given prototype
func NewDocument(val *DocumentProto) *Document {
	return &Document{
		Comments: NewComments(val.Comments),
		Value:    NewValue(val.Value),
		Position: val.Position,
	}
}

// NewMap creates a new Map instance based on the given prototype
func NewMap(val *MapProto) *Map {
	newMap := &Map{
		Comments: NewComments(val.Comments),
		Position: val.Position,
	}
	for _, item := range val.Items {
		newMap.Items = append(newMap.Items, NewMapItem(item))
	}
	return newMap
}

// NewMapItem creates a new MapItem instance based on the given prototype
func NewMapItem(val *MapItemProto) *MapItem {
	return &MapItem{
		Comments: NewComments(val.Comments),
		Key:      val.Key,
		Value:    NewValue(val.Value),
		Position: val.Position,
	}
}

// NewArray creates a new Array instance based on the given prototype
func NewArray(val *ArrayProto) *Array {
	newArray := &Array{
		Comments: NewComments(val.Comments),
		Position: val.Position,
	}
	for _, item := range val.Items {
		newArray.Items = append(newArray.Items, NewArrayItem(item))
	}
	return newArray
}

// NewArrayItem creates a new ArrayItem instance based on the given prototype
func NewArrayItem(val *ArrayItemProto) *ArrayItem {
	return &ArrayItem{
		Comments: NewComments(val.Comments),
		Value:    NewValue(val.Value),
		Position: val.Position,
	}
}

// NewComments creates a new slice of Comment based on the given prototype
func NewComments(val []*CommentProto) []*Comment {
	newComments := []*Comment{}
	for _, comment := range val {
		newComments = append(newComments, NewComment(comment))
	}
	return newComments
}

// NewComment creates a new Comment based on the given prototype
func NewComment(val *CommentProto) *Comment {
	return &Comment{
		Data:     val.Data,
		Position: val.Position,
	}
}

// NewValue creates a new Document, Map, Array, or scalar, depending on the type of the given prototype
func NewValue(val interface{}) interface{} {
	switch typedVal := val.(type) {
	case *DocumentProto:
		return NewDocument(typedVal)
	case *MapProto:
		return NewMap(typedVal)
	case *ArrayProto:
		return NewArray(typedVal)
	default:
		return val
	}
}

// DocumentSetProto is a prototype for a DocumentSet
type DocumentSetProto struct {
	Items    []*DocumentProto
	Comments []*CommentProto
	Position *filepos.Position
}

// DocumentProto is a prototype for a Document
type DocumentProto struct {
	Value    interface{}
	Comments []*CommentProto
	Position *filepos.Position
}

// MapProto is a prototype for a Map
type MapProto struct {
	Items    []*MapItemProto
	Comments []*CommentProto
	Position *filepos.Position
}

// MapItemProto is a prototype for a MapItem
type MapItemProto struct {
	Key      interface{}
	Value    interface{}
	Comments []*CommentProto
	Position *filepos.Position
}

// ArrayProto is a prototype for an Array
type ArrayProto struct {
	Items    []*ArrayItemProto
	Comments []*CommentProto
	Position *filepos.Position
}

// ArrayItemProto is a prototype for an ArrayItem
type ArrayItemProto struct {
	Value    interface{}
	Comments []*CommentProto
	Position *filepos.Position
}

// CommentProto is a prototype for a Comment
type CommentProto struct {
	Data     string
	Position *filepos.Position
}
