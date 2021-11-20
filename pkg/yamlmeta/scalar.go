// Copyright 2020 VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package yamlmeta

import "github.com/k14s/ytt/pkg/filepos"

func (s *Scalar) GetComments() []*Comment {
	panic("implement me")
}

func (s *Scalar) addComments(comment *Comment) {
	panic("implement me")
}

func (s *Scalar) GetMeta(name string) interface{} {
	panic("implement me")
}

func (s *Scalar) SetMeta(name string, data interface{}) {
	panic("implement me")
}

func (s *Scalar) GetAnnotations() interface{} {
	panic("implement me")
}

func (s *Scalar) SetAnnotations(i interface{}) {
	panic("implement me")
}

func (s *Scalar) DeepCopyAsInterface() interface{} {
	panic("implement me")
}

func (s *Scalar) DeepCopyAsNode() Node {
	panic("implement me")
}

func (s *Scalar) DisplayName() string {
	panic("implement me")
}

func (s *Scalar) sealed() {
	panic("implement me")
}
func (s *Scalar) SetPosition(position *filepos.Position) { s.Position = position }

func (s *Scalar) SetValue(val interface{}) error {
	s.ResetValue()
	return s.AddValue(val)
}

func (s *Scalar) ResetValue() { s.Value = nil }

func (s *Scalar) AddValue(val interface{}) error {
	s.Value = val
	return nil
}

func (s *Scalar) GetValues() []interface{} { return []interface{}{s.Value} }

func (s *Scalar) GetPosition() *filepos.Position { return s.Position }
