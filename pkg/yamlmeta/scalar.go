// Copyright 2020 VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package yamlmeta

import "github.com/k14s/ytt/pkg/filepos"

// GetComments is not implemented.
func (s *Scalar) GetComments() []*Comment {
	panic("implement me")
}

// addComments is not implemented.
func (s *Scalar) addComments(comment *Comment) {
	panic("implement me")
}

// GetMeta is not implemented.
func (s *Scalar) GetMeta(name string) interface{} {
	panic("implement me")
}

// SetMeta is not implemented.
func (s *Scalar) SetMeta(name string, data interface{}) {
	panic("implement me")
}

// GetAnnotations is not implemented.
func (s *Scalar) GetAnnotations() interface{} {
	panic("implement me")
}

// SetAnnotations is not implemented.
func (s *Scalar) SetAnnotations(i interface{}) {
	panic("implement me")
}

// DeepCopyAsInterface is not implemented.
func (s *Scalar) DeepCopyAsInterface() interface{} {
	panic("implement me")
}

// DeepCopyAsNode is not implemented.
func (s *Scalar) DeepCopyAsNode() Node {
	panic("implement me")
}

// DisplayName is not implemented.
func (s *Scalar) DisplayName() string {
	panic("implement me")
}

// sealed is not implemented.
func (s *Scalar) sealed() {
	panic("implement me")
}

// SetPosition sets the position for this scalar value.
func (s *Scalar) SetPosition(position *filepos.Position) { s.Position = position }

// SetValue resets the value of this scalar to be `val`.
func (s *Scalar) SetValue(val interface{}) error {
	s.ResetValue()
	return s.AddValue(val)
}

// ResetValue clears the value of this scalar (i.e. sets it to nil)
func (s *Scalar) ResetValue() { s.Value = nil }

// AddValue sets the value of this scalar to `val`
func (s *Scalar) AddValue(val interface{}) error {
	s.Value = val
	return nil
}

// GetValues returns the value of this scalar as the only element in a slice.
func (s *Scalar) GetValues() []interface{} { return []interface{}{s.Value} }

// GetPosition returns the location where this scalar value originated.
func (s *Scalar) GetPosition() *filepos.Position { return s.Position }
