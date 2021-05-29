// Copyright 2020 VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package overlay

import (
	"github.com/k14s/ytt/pkg/yamlmeta"
)

func (o Op) mergeDocument(
	leftDocSets []*yamlmeta.DocumentSet, newDoc *yamlmeta.Document,
	parentMatchChildDefaults MatchChildDefaultsAnnotation) error {

	matchChildDefaults, err := NewMatchChildDefaultsAnnotation(newDoc, parentMatchChildDefaults)
	if err != nil {
		return err
	}

	ann, err := NewDocumentMatchAnnotation(newDoc, parentMatchChildDefaults, o.ExactMatch, o.Thread)
	if err != nil {
		return err
	}

	leftIdxs, err := ann.IndexTuples(leftDocSets)
	if err != nil {
		if err, ok := err.(MatchAnnotationNumMatchError); ok && err.isConditional() {
			return nil
		}
		return err
	}

	for _, leftIdx := range leftIdxs {
		replace := true
		if leftDocSets[leftIdx[0]].Items[leftIdx[1]].Value != nil {
			replace, err = o.apply(leftDocSets[leftIdx[0]].Items[leftIdx[1]].Value, newDoc.Value, matchChildDefaults)
			if err != nil {
				return err
			}
		}
		if replace {
			leftDocSets[leftIdx[0]].Items[leftIdx[1]].Value = newDoc.Value
		}
	}

	return nil
}

func (o Op) removeDocument(
	leftDocSets []*yamlmeta.DocumentSet, newDoc *yamlmeta.Document,
	parentMatchChildDefaults MatchChildDefaultsAnnotation) error {

	ann, err := NewDocumentMatchAnnotation(newDoc, parentMatchChildDefaults, o.ExactMatch, o.Thread)
	if err != nil {
		return err
	}

	leftIdxs, err := ann.IndexTuples(leftDocSets)
	if err != nil {
		if err, ok := err.(MatchAnnotationNumMatchError); ok && err.isConditional() {
			return nil
		}
		return err
	}

	for _, leftIdx := range leftIdxs {
		leftDocSets[leftIdx[0]].Items[leftIdx[1]] = nil
	}

	// Prune out all nil documents
	for _, leftDocSet := range leftDocSets {
		updatedDocs := []*yamlmeta.Document{}

		for _, item := range leftDocSet.Items {
			if item != nil {
				updatedDocs = append(updatedDocs, item)
			}
		}

		leftDocSet.Items = updatedDocs
	}

	return nil
}

func (o Op) replaceDocument(
	leftDocSets []*yamlmeta.DocumentSet, newDoc *yamlmeta.Document,
	parentMatchChildDefaults MatchChildDefaultsAnnotation) error {

	ann, err := NewDocumentMatchAnnotation(newDoc, parentMatchChildDefaults, o.ExactMatch, o.Thread)
	if err != nil {
		return err
	}

	replaceAnn, err := NewReplaceAnnotation(newDoc, o.Thread)
	if err != nil {
		return err
	}

	leftIdxs, err := ann.IndexTuples(leftDocSets)
	if err != nil {
		if err, ok := err.(MatchAnnotationNumMatchError); ok && err.isConditional() {
			return nil
		}
		return err
	}

	for _, leftIdx := range leftIdxs {
		newVal, err := replaceAnn.Value(leftDocSets[leftIdx[0]].Items[leftIdx[1]])
		if err != nil {
			return err
		}

		leftDocSets[leftIdx[0]].Items[leftIdx[1]] = newDoc.DeepCopy()
		err = leftDocSets[leftIdx[0]].Items[leftIdx[1]].SetValue(newVal)
		if err != nil {
			return err
		}
	}

	if len(leftIdxs) == 0 && replaceAnn.OrAdd() {
		if len(leftDocSets) == 0 {
			panic("Internal inconsistency: Expected at least one doc set")
		}

		newVal, err := replaceAnn.Value(nil)
		if err != nil {
			return err
		}

		leftDocSets[0].Items = append(leftDocSets[0].Items, newDoc.DeepCopy())
		err = leftDocSets[0].Items[len(leftDocSets[0].Items)-1].SetValue(newVal)
		if err != nil {
			return err
		}
	}

	return nil
}

func (o Op) insertDocument(
	leftDocSets []*yamlmeta.DocumentSet, newDoc *yamlmeta.Document,
	parentMatchChildDefaults MatchChildDefaultsAnnotation) error {

	ann, err := NewDocumentMatchAnnotation(newDoc, parentMatchChildDefaults, o.ExactMatch, o.Thread)
	if err != nil {
		return err
	}

	leftIdxs, err := ann.IndexTuples(leftDocSets)
	if err != nil {
		if err, ok := err.(MatchAnnotationNumMatchError); ok && err.isConditional() {
			return nil
		}
		return err
	}

	insertAnn, err := NewInsertAnnotation(newDoc)
	if err != nil {
		return err
	}

	for i, leftDocSet := range leftDocSets {
		updatedDocs := []*yamlmeta.Document{}

		for j, leftItem := range leftDocSet.Items {
			matched := false
			for _, leftIdx := range leftIdxs {
				if leftIdx[0] == i && leftIdx[1] == j {
					matched = true
					if insertAnn.IsBefore() {
						updatedDocs = append(updatedDocs, newDoc.DeepCopy())
					}
					updatedDocs = append(updatedDocs, leftItem)
					if insertAnn.IsAfter() {
						updatedDocs = append(updatedDocs, newDoc.DeepCopy())
					}
					break
				}
			}
			if !matched {
				updatedDocs = append(updatedDocs, leftItem)
			}
		}

		leftDocSet.Items = updatedDocs
	}

	return nil
}

func (o Op) appendDocument(
	leftDocSets []*yamlmeta.DocumentSet, newDoc *yamlmeta.Document) error {

	// No need to traverse further
	leftDocSets[len(leftDocSets)-1].Items = append(leftDocSets[len(leftDocSets)-1].Items, newDoc.DeepCopy())
	return nil
}

func (o Op) assertDocument(
	leftDocSets []*yamlmeta.DocumentSet, newDoc *yamlmeta.Document,
	parentMatchChildDefaults MatchChildDefaultsAnnotation) error {

	matchChildDefaults, err := NewMatchChildDefaultsAnnotation(newDoc, parentMatchChildDefaults)
	if err != nil {
		return err
	}

	ann, err := NewDocumentMatchAnnotation(newDoc, parentMatchChildDefaults, o.ExactMatch, o.Thread)
	if err != nil {
		return err
	}

	testAnn, err := NewAssertAnnotation(newDoc, o.Thread)
	if err != nil {
		return err
	}

	leftIdxs, err := ann.IndexTuples(leftDocSets)
	if err != nil {
		if err, ok := err.(MatchAnnotationNumMatchError); ok && err.isConditional() {
			return nil
		}
		return err
	}

	for _, leftIdx := range leftIdxs {
		err := testAnn.Check(leftDocSets[leftIdx[0]].Items[leftIdx[1]])
		if err != nil {
			return err
		}

		_, err = o.apply(leftDocSets[leftIdx[0]].Items[leftIdx[1]].Value, newDoc.Value, matchChildDefaults)
		if err != nil {
			return err
		}
	}

	return nil
}
