// Copyright 2024 The Carvel Authors.
// SPDX-License-Identifier: Apache-2.0

package yamlmeta_test

import (
	"testing"

	"carvel.dev/ytt/pkg/yamlmeta"
)

func TestMapKeyOverridePlainYAML(t *testing.T) {
	t.Run("when no maps have duplicate keys, is a no op", func(t *testing.T) {
		docSet := &yamlmeta.DocumentSet{
			Items: []*yamlmeta.Document{{
				Value: &yamlmeta.Map{
					Items: []*yamlmeta.MapItem{
						{Key: "foo", Value: 1},
						{Key: "bar", Value: 2},
					},
				},
			}},
		}
		expectedDocSet := docSet.DeepCopyAsNode()
		docSet.OverrideMapKeys()

		printer := yamlmeta.NewPrinterWithOpts(nil, yamlmeta.PrinterOpts{ExcludeRefs: true})

		result := printer.PrintStr(docSet)
		expected := printer.PrintStr(expectedDocSet)
		assertEqual(t, result, expected)
	})
	t.Run("when there are duplicates, last map item overrides", func(t *testing.T) {
		docSet := &yamlmeta.DocumentSet{
			Items: []*yamlmeta.Document{{
				Value: &yamlmeta.Map{
					Items: []*yamlmeta.MapItem{
						{Key: "foo", Value: 1},
						{Key: "foo", Value: 2},
						{Key: "foo", Value: 3},
						{Key: "foo", Value: 4},
					},
				},
			}},
		}
		expectedDocSet := &yamlmeta.DocumentSet{
			Items: []*yamlmeta.Document{{
				Value: &yamlmeta.Map{
					Items: []*yamlmeta.MapItem{
						{Key: "foo", Value: 4},
					},
				},
			}},
		}
		docSet.OverrideMapKeys()

		printer := yamlmeta.NewPrinterWithOpts(nil, yamlmeta.PrinterOpts{ExcludeRefs: true})

		result := printer.PrintStr(docSet)
		expected := printer.PrintStr(expectedDocSet)
		assertEqual(t, result, expected)
	})
	t.Run("when there are multiple keys with duplicates, last map item for each key overrides the others", func(t *testing.T) {
		docSet := &yamlmeta.DocumentSet{
			Items: []*yamlmeta.Document{{
				Value: &yamlmeta.Map{
					Items: []*yamlmeta.MapItem{
						{Key: "foo", Value: 1},
						{Key: "bar", Value: 2},
						{Key: "ree", Value: 3},
						{Key: "ree", Value: 4},
						{Key: "foo", Value: 5},
						{Key: "bar", Value: 6},
					},
				},
			}},
		}
		expectedDocSet := &yamlmeta.DocumentSet{
			Items: []*yamlmeta.Document{{
				Value: &yamlmeta.Map{
					Items: []*yamlmeta.MapItem{
						{Key: "foo", Value: 5},
						{Key: "bar", Value: 6},
						{Key: "ree", Value: 4},
					},
				},
			}},
		}
		docSet.OverrideMapKeys()

		printer := yamlmeta.NewPrinterWithOpts(nil, yamlmeta.PrinterOpts{ExcludeRefs: true})

		result := printer.PrintStr(docSet)
		expected := printer.PrintStr(expectedDocSet)
		assertEqual(t, result, expected)
	})
}
