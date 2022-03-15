// Copyright 2020 VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package yamltemplate

import "github.com/vmware-tanzu/carvel-ytt/pkg/template"

// Comment annotation names
const (
	AnnotationYAMLComment template.AnnotationName = "yaml/comment"
	AnnotationHeadComment template.AnnotationName = "yaml/head-comment"
	AnnotationLineComment template.AnnotationName = "yaml/line-comment"
)
