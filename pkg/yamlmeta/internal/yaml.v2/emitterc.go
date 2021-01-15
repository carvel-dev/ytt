// Copyright 2020 VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package yaml

import (
	"bytes"
	"fmt"
)

// Flush the buffer if needed.
func flush(emitter *yamlEmitterT) bool {
	if emitter.bufferPos+5 >= len(emitter.buffer) {
		return yamlEmitterFlush(emitter)
	}
	return true
}

// Put a character to the output buffer.
func put(emitter *yamlEmitterT, value byte) bool {
	if emitter.bufferPos+5 >= len(emitter.buffer) && !yamlEmitterFlush(emitter) {
		return false
	}
	emitter.buffer[emitter.bufferPos] = value
	emitter.bufferPos++
	emitter.column++
	return true
}

// Put a line break to the output buffer.
func putBreak(emitter *yamlEmitterT) bool {
	if emitter.bufferPos+5 >= len(emitter.buffer) && !yamlEmitterFlush(emitter) {
		return false
	}
	switch emitter.lineBreak {
	case yamlCrBreak:
		emitter.buffer[emitter.bufferPos] = '\r'
		emitter.bufferPos++
	case yamlLnBreak:
		emitter.buffer[emitter.bufferPos] = '\n'
		emitter.bufferPos++
	case yamlCrlnBreak:
		emitter.buffer[emitter.bufferPos+0] = '\r'
		emitter.buffer[emitter.bufferPos+1] = '\n'
		emitter.bufferPos += 2
	default:
		panic("unknown line break setting")
	}
	emitter.column = 0
	emitter.line++
	return true
}

// Copy a character from a string into buffer.
func write(emitter *yamlEmitterT, s []byte, i *int) bool {
	if emitter.bufferPos+5 >= len(emitter.buffer) && !yamlEmitterFlush(emitter) {
		return false
	}
	p := emitter.bufferPos
	w := width(s[*i])
	switch w {
	case 4:
		emitter.buffer[p+3] = s[*i+3]
		fallthrough
	case 3:
		emitter.buffer[p+2] = s[*i+2]
		fallthrough
	case 2:
		emitter.buffer[p+1] = s[*i+1]
		fallthrough
	case 1:
		emitter.buffer[p+0] = s[*i+0]
	default:
		panic("unknown character width")
	}
	emitter.column++
	emitter.bufferPos += w
	*i += w
	return true
}

// Write a whole string into buffer.
func writeAll(emitter *yamlEmitterT, s []byte) bool {
	for i := 0; i < len(s); {
		if !write(emitter, s, &i) {
			return false
		}
	}
	return true
}

// Copy a line break character from a string into buffer.
func writeBreak(emitter *yamlEmitterT, s []byte, i *int) bool {
	if s[*i] == '\n' {
		if !putBreak(emitter) {
			return false
		}
		*i++
	} else {
		if !write(emitter, s, i) {
			return false
		}
		emitter.column = 0
		emitter.line++
	}
	return true
}

// Set an emitter error and return false.
func yamlEmitterSetEmitterError(emitter *yamlEmitterT, problem string) bool {
	emitter.error = yamlEmitterError
	emitter.problem = problem
	return false
}

// Emit an event.
func yamlEmitterEmit(emitter *yamlEmitterT, event *yamlEventT) bool {
	emitter.events = append(emitter.events, *event)
	for !yamlEmitterNeedMoreEvents(emitter) {
		event := &emitter.events[emitter.eventsHead]
		if !yamlEmitterAnalyzeEvent(emitter, event) {
			return false
		}
		if !yamlEmitterStateMachine(emitter, event) {
			return false
		}
		yamlEventDelete(event)
		emitter.eventsHead++
	}
	return true
}

// Check if we need to accumulate more events before emitting.
//
// We accumulate extra
//  - 1 event for DOCUMENT-START
//  - 2 events for SEQUENCE-START
//  - 3 events for MAPPING-START
//
func yamlEmitterNeedMoreEvents(emitter *yamlEmitterT) bool {
	if emitter.eventsHead == len(emitter.events) {
		return true
	}
	var accumulate int
	switch emitter.events[emitter.eventsHead].typ {
	case yamlDocumentStartEvent:
		accumulate = 1
		break
	case yamlSequenceStartEvent:
		accumulate = 2
		break
	case yamlMappingStartEvent:
		accumulate = 3
		break
	default:
		return false
	}
	if len(emitter.events)-emitter.eventsHead > accumulate {
		return false
	}
	var level int
	for i := emitter.eventsHead; i < len(emitter.events); i++ {
		switch emitter.events[i].typ {
		case yamlStreamStartEvent, yamlDocumentStartEvent, yamlSequenceStartEvent, yamlMappingStartEvent:
			level++
		case yamlStreamEndEvent, yamlDocumentEndEvent, yamlSequenceEndEvent, yamlMappingEndEvent:
			level--
		}
		if level == 0 {
			return false
		}
	}
	return true
}

// Append a directive to the directives stack.
func yamlEmitterAppendTagDirective(emitter *yamlEmitterT, value *yamlTagDirectiveT, allowDuplicates bool) bool {
	for i := 0; i < len(emitter.tagDirectives); i++ {
		if bytes.Equal(value.handle, emitter.tagDirectives[i].handle) {
			if allowDuplicates {
				return true
			}
			return yamlEmitterSetEmitterError(emitter, "duplicate %TAG directive")
		}
	}

	// [Go] Do we actually need to copy this given garbage collection
	// and the lack of deallocating destructors?
	tagCopy := yamlTagDirectiveT{
		handle: make([]byte, len(value.handle)),
		prefix: make([]byte, len(value.prefix)),
	}
	copy(tagCopy.handle, value.handle)
	copy(tagCopy.prefix, value.prefix)
	emitter.tagDirectives = append(emitter.tagDirectives, tagCopy)
	return true
}

// Increase the indentation level.
func yamlEmitterIncreaseIndent(emitter *yamlEmitterT, flow, indentless bool) bool {
	emitter.indents = append(emitter.indents, emitter.indent)
	if emitter.indent < 0 {
		if flow {
			emitter.indent = emitter.bestIndent
		} else {
			emitter.indent = 0
		}
	} else if !indentless {
		emitter.indent += emitter.bestIndent
	}
	return true
}

// State dispatcher.
func yamlEmitterStateMachine(emitter *yamlEmitterT, event *yamlEventT) bool {
	switch emitter.state {
	default:
	case yamlEmitStreamStartState:
		return yamlEmitterEmitStreamStart(emitter, event)

	case yamlEmitFirstDocumentStartState:
		return yamlEmitterEmitDocumentStart(emitter, event, true)

	case yamlEmitDocumentStartState:
		return yamlEmitterEmitDocumentStart(emitter, event, false)

	case yamlEmitDocumentContentState:
		return yamlEmitterEmitDocumentContent(emitter, event)

	case yamlEmitDocumentEndState:
		return yamlEmitterEmitDocumentEnd(emitter, event)

	case yamlEmitFlowSequenceFirstItemState:
		return yamlEmitterEmitFlowSequenceItem(emitter, event, true)

	case yamlEmitFlowSequenceItemState:
		return yamlEmitterEmitFlowSequenceItem(emitter, event, false)

	case yamlEmitFlowMappingFirstKeyState:
		return yamlEmitterEmitFlowMappingKey(emitter, event, true)

	case yamlEmitFlowMappingKeyState:
		return yamlEmitterEmitFlowMappingKey(emitter, event, false)

	case yamlEmitFlowMappingSimpleValueState:
		return yamlEmitterEmitFlowMappingValue(emitter, event, true)

	case yamlEmitFlowMappingValueState:
		return yamlEmitterEmitFlowMappingValue(emitter, event, false)

	case yamlEmitBlockSequenceFirstItemState:
		return yamlEmitterEmitBlockSequenceItem(emitter, event, true)

	case yamlEmitBlockSequenceItemState:
		return yamlEmitterEmitBlockSequenceItem(emitter, event, false)

	case yamlEmitBlockMappingFirstKeyState:
		return yamlEmitterEmitBlockMappingKey(emitter, event, true)

	case yamlEmitBlockMappingKeyState:
		return yamlEmitterEmitBlockMappingKey(emitter, event, false)

	case yamlEmitBlockMappingSimpleValueState:
		return yamlEmitterEmitBlockMappingValue(emitter, event, true)

	case yamlEmitBlockMappingValueState:
		return yamlEmitterEmitBlockMappingValue(emitter, event, false)

	case yamlEmitEndState:
		return yamlEmitterSetEmitterError(emitter, "expected nothing after STREAM-END")
	}
	panic("invalid emitter state")
}

// Expect STREAM-START.
func yamlEmitterEmitStreamStart(emitter *yamlEmitterT, event *yamlEventT) bool {
	if event.typ != yamlStreamStartEvent {
		return yamlEmitterSetEmitterError(emitter, "expected STREAM-START")
	}
	if emitter.encoding == yamlAnyEncoding {
		emitter.encoding = event.encoding
		if emitter.encoding == yamlAnyEncoding {
			emitter.encoding = yamlUtf8Encoding
		}
	}
	if emitter.bestIndent < 2 || emitter.bestIndent > 9 {
		emitter.bestIndent = 2
	}
	if emitter.bestWidth >= 0 && emitter.bestWidth <= emitter.bestIndent*2 {
		emitter.bestWidth = 80
	}
	if emitter.bestWidth < 0 {
		emitter.bestWidth = 1<<31 - 1
	}
	if emitter.lineBreak == yamlAnyBreak {
		emitter.lineBreak = yamlLnBreak
	}

	emitter.indent = -1
	emitter.line = 0
	emitter.column = 0
	emitter.whitespace = true
	emitter.indention = true

	if emitter.encoding != yamlUtf8Encoding {
		if !yamlEmitterWriteBom(emitter) {
			return false
		}
	}
	emitter.state = yamlEmitFirstDocumentStartState
	return true
}

// Expect DOCUMENT-START or STREAM-END.
func yamlEmitterEmitDocumentStart(emitter *yamlEmitterT, event *yamlEventT, first bool) bool {

	if event.typ == yamlDocumentStartEvent {

		if event.versionDirective != nil {
			if !yamlEmitterAnalyzeVersionDirective(emitter, event.versionDirective) {
				return false
			}
		}

		for i := 0; i < len(event.tagDirectives); i++ {
			tagDirective := &event.tagDirectives[i]
			if !yamlEmitterAnalyzeTagDirective(emitter, tagDirective) {
				return false
			}
			if !yamlEmitterAppendTagDirective(emitter, tagDirective, false) {
				return false
			}
		}

		for i := 0; i < len(defaultTagDirectives); i++ {
			tagDirective := &defaultTagDirectives[i]
			if !yamlEmitterAppendTagDirective(emitter, tagDirective, true) {
				return false
			}
		}

		implicit := event.implicit
		if !first || emitter.canonical {
			implicit = false
		}

		if emitter.openEnded && (event.versionDirective != nil || len(event.tagDirectives) > 0) {
			if !yamlEmitterWriteIndicator(emitter, []byte("..."), true, false, false) {
				return false
			}
			if !yamlEmitterWriteIndent(emitter) {
				return false
			}
		}

		if event.versionDirective != nil {
			implicit = false
			if !yamlEmitterWriteIndicator(emitter, []byte("%YAML"), true, false, false) {
				return false
			}
			if !yamlEmitterWriteIndicator(emitter, []byte("1.1"), true, false, false) {
				return false
			}
			if !yamlEmitterWriteIndent(emitter) {
				return false
			}
		}

		if len(event.tagDirectives) > 0 {
			implicit = false
			for i := 0; i < len(event.tagDirectives); i++ {
				tagDirective := &event.tagDirectives[i]
				if !yamlEmitterWriteIndicator(emitter, []byte("%TAG"), true, false, false) {
					return false
				}
				if !yamlEmitterWriteTagHandle(emitter, tagDirective.handle) {
					return false
				}
				if !yamlEmitterWriteTagContent(emitter, tagDirective.prefix, true) {
					return false
				}
				if !yamlEmitterWriteIndent(emitter) {
					return false
				}
			}
		}

		if yamlEmitterCheckEmptyDocument(emitter) {
			implicit = false
		}
		if !implicit {
			if !yamlEmitterWriteIndent(emitter) {
				return false
			}
			if !yamlEmitterWriteIndicator(emitter, []byte("---"), true, false, false) {
				return false
			}
			if emitter.canonical {
				if !yamlEmitterWriteIndent(emitter) {
					return false
				}
			}
		}

		emitter.state = yamlEmitDocumentContentState
		return true
	}

	if event.typ == yamlStreamEndEvent {
		if emitter.openEnded {
			if !yamlEmitterWriteIndicator(emitter, []byte("..."), true, false, false) {
				return false
			}
			if !yamlEmitterWriteIndent(emitter) {
				return false
			}
		}
		if !yamlEmitterFlush(emitter) {
			return false
		}
		emitter.state = yamlEmitEndState
		return true
	}

	return yamlEmitterSetEmitterError(emitter, "expected DOCUMENT-START or STREAM-END")
}

// Expect the root node.
func yamlEmitterEmitDocumentContent(emitter *yamlEmitterT, event *yamlEventT) bool {
	emitter.states = append(emitter.states, yamlEmitDocumentEndState)
	return yamlEmitterEmitNode(emitter, event, true, false, false, false)
}

// Expect DOCUMENT-END.
func yamlEmitterEmitDocumentEnd(emitter *yamlEmitterT, event *yamlEventT) bool {
	if event.typ != yamlDocumentEndEvent {
		return yamlEmitterSetEmitterError(emitter, "expected DOCUMENT-END")
	}
	if !yamlEmitterWriteIndent(emitter) {
		return false
	}
	if !event.implicit {
		// [Go] Allocate the slice elsewhere.
		if !yamlEmitterWriteIndicator(emitter, []byte("..."), true, false, false) {
			return false
		}
		if !yamlEmitterWriteIndent(emitter) {
			return false
		}
	}
	if !yamlEmitterFlush(emitter) {
		return false
	}
	emitter.state = yamlEmitDocumentStartState
	emitter.tagDirectives = emitter.tagDirectives[:0]
	return true
}

// Expect a flow item node.
func yamlEmitterEmitFlowSequenceItem(emitter *yamlEmitterT, event *yamlEventT, first bool) bool {
	if first {
		if !yamlEmitterWriteIndicator(emitter, []byte{'['}, true, true, false) {
			return false
		}
		if !yamlEmitterIncreaseIndent(emitter, true, false) {
			return false
		}
		emitter.flowLevel++
	}

	if event.typ == yamlSequenceEndEvent {
		emitter.flowLevel--
		emitter.indent = emitter.indents[len(emitter.indents)-1]
		emitter.indents = emitter.indents[:len(emitter.indents)-1]
		if emitter.canonical && !first {
			if !yamlEmitterWriteIndicator(emitter, []byte{','}, false, false, false) {
				return false
			}
			if !yamlEmitterWriteIndent(emitter) {
				return false
			}
		}
		if !yamlEmitterWriteIndicator(emitter, []byte{']'}, false, false, false) {
			return false
		}
		emitter.state = emitter.states[len(emitter.states)-1]
		emitter.states = emitter.states[:len(emitter.states)-1]

		return true
	}

	if !first {
		if !yamlEmitterWriteIndicator(emitter, []byte{','}, false, false, false) {
			return false
		}
	}

	if emitter.canonical || emitter.column > emitter.bestWidth {
		if !yamlEmitterWriteIndent(emitter) {
			return false
		}
	}
	emitter.states = append(emitter.states, yamlEmitFlowSequenceItemState)
	return yamlEmitterEmitNode(emitter, event, false, true, false, false)
}

// Expect a flow key node.
func yamlEmitterEmitFlowMappingKey(emitter *yamlEmitterT, event *yamlEventT, first bool) bool {
	if first {
		if !yamlEmitterWriteIndicator(emitter, []byte{'{'}, true, true, false) {
			return false
		}
		if !yamlEmitterIncreaseIndent(emitter, true, false) {
			return false
		}
		emitter.flowLevel++
	}

	if event.typ == yamlMappingEndEvent {
		emitter.flowLevel--
		emitter.indent = emitter.indents[len(emitter.indents)-1]
		emitter.indents = emitter.indents[:len(emitter.indents)-1]
		if emitter.canonical && !first {
			if !yamlEmitterWriteIndicator(emitter, []byte{','}, false, false, false) {
				return false
			}
			if !yamlEmitterWriteIndent(emitter) {
				return false
			}
		}
		if !yamlEmitterWriteIndicator(emitter, []byte{'}'}, false, false, false) {
			return false
		}
		emitter.state = emitter.states[len(emitter.states)-1]
		emitter.states = emitter.states[:len(emitter.states)-1]
		return true
	}

	if !first {
		if !yamlEmitterWriteIndicator(emitter, []byte{','}, false, false, false) {
			return false
		}
	}
	if emitter.canonical || emitter.column > emitter.bestWidth {
		if !yamlEmitterWriteIndent(emitter) {
			return false
		}
	}

	if !emitter.canonical && yamlEmitterCheckSimpleKey(emitter) {
		emitter.states = append(emitter.states, yamlEmitFlowMappingSimpleValueState)
		return yamlEmitterEmitNode(emitter, event, false, false, true, true)
	}
	if !yamlEmitterWriteIndicator(emitter, []byte{'?'}, true, false, false) {
		return false
	}
	emitter.states = append(emitter.states, yamlEmitFlowMappingValueState)
	return yamlEmitterEmitNode(emitter, event, false, false, true, false)
}

// Expect a flow value node.
func yamlEmitterEmitFlowMappingValue(emitter *yamlEmitterT, event *yamlEventT, simple bool) bool {
	if simple {
		if !yamlEmitterWriteIndicator(emitter, []byte{':'}, false, false, false) {
			return false
		}
	} else {
		if emitter.canonical || emitter.column > emitter.bestWidth {
			if !yamlEmitterWriteIndent(emitter) {
				return false
			}
		}
		if !yamlEmitterWriteIndicator(emitter, []byte{':'}, true, false, false) {
			return false
		}
	}
	emitter.states = append(emitter.states, yamlEmitFlowMappingKeyState)
	return yamlEmitterEmitNode(emitter, event, false, false, true, false)
}

// Expect a block item node.
func yamlEmitterEmitBlockSequenceItem(emitter *yamlEmitterT, event *yamlEventT, first bool) bool {
	if first {
		if !yamlEmitterIncreaseIndent(emitter, false, emitter.mappingContext && !emitter.indention) {
			return false
		}
	}
	if event.typ == yamlSequenceEndEvent {
		emitter.indent = emitter.indents[len(emitter.indents)-1]
		emitter.indents = emitter.indents[:len(emitter.indents)-1]
		emitter.state = emitter.states[len(emitter.states)-1]
		emitter.states = emitter.states[:len(emitter.states)-1]
		return true
	}
	if !yamlEmitterWriteIndent(emitter) {
		return false
	}
	if !yamlEmitterWriteIndicator(emitter, []byte{'-'}, true, false, true) {
		return false
	}
	emitter.states = append(emitter.states, yamlEmitBlockSequenceItemState)
	return yamlEmitterEmitNode(emitter, event, false, true, false, false)
}

// Expect a block key node.
func yamlEmitterEmitBlockMappingKey(emitter *yamlEmitterT, event *yamlEventT, first bool) bool {
	if first {
		if !yamlEmitterIncreaseIndent(emitter, false, false) {
			return false
		}
	}
	if event.typ == yamlMappingEndEvent {
		emitter.indent = emitter.indents[len(emitter.indents)-1]
		emitter.indents = emitter.indents[:len(emitter.indents)-1]
		emitter.state = emitter.states[len(emitter.states)-1]
		emitter.states = emitter.states[:len(emitter.states)-1]
		return true
	}
	if !yamlEmitterWriteIndent(emitter) {
		return false
	}
	if yamlEmitterCheckSimpleKey(emitter) {
		emitter.states = append(emitter.states, yamlEmitBlockMappingSimpleValueState)
		return yamlEmitterEmitNode(emitter, event, false, false, true, true)
	}
	if !yamlEmitterWriteIndicator(emitter, []byte{'?'}, true, false, true) {
		return false
	}
	emitter.states = append(emitter.states, yamlEmitBlockMappingValueState)
	return yamlEmitterEmitNode(emitter, event, false, false, true, false)
}

// Expect a block value node.
func yamlEmitterEmitBlockMappingValue(emitter *yamlEmitterT, event *yamlEventT, simple bool) bool {
	if simple {
		if !yamlEmitterWriteIndicator(emitter, []byte{':'}, false, false, false) {
			return false
		}
	} else {
		if !yamlEmitterWriteIndent(emitter) {
			return false
		}
		if !yamlEmitterWriteIndicator(emitter, []byte{':'}, true, false, true) {
			return false
		}
	}
	emitter.states = append(emitter.states, yamlEmitBlockMappingKeyState)
	return yamlEmitterEmitNode(emitter, event, false, false, true, false)
}

// Expect a node.
func yamlEmitterEmitNode(emitter *yamlEmitterT, event *yamlEventT,
	root bool, sequence bool, mapping bool, simpleKey bool) bool {

	emitter.rootContext = root
	emitter.sequenceContext = sequence
	emitter.mappingContext = mapping
	emitter.simpleKeyContext = simpleKey

	switch event.typ {
	case yamlAliasEvent:
		return yamlEmitterEmitAlias(emitter, event)
	case yamlScalarEvent:
		return yamlEmitterEmitScalar(emitter, event)
	case yamlSequenceStartEvent:
		return yamlEmitterEmitSequenceStart(emitter, event)
	case yamlMappingStartEvent:
		return yamlEmitterEmitMappingStart(emitter, event)
	default:
		return yamlEmitterSetEmitterError(emitter,
			fmt.Sprintf("expected SCALAR, SEQUENCE-START, MAPPING-START, or ALIAS, but got %v", event.typ))
	}
}

// Expect ALIAS.
func yamlEmitterEmitAlias(emitter *yamlEmitterT, event *yamlEventT) bool {
	if !yamlEmitterProcessAnchor(emitter) {
		return false
	}
	emitter.state = emitter.states[len(emitter.states)-1]
	emitter.states = emitter.states[:len(emitter.states)-1]
	return true
}

// Expect SCALAR.
func yamlEmitterEmitScalar(emitter *yamlEmitterT, event *yamlEventT) bool {
	if !yamlEmitterSelectScalarStyle(emitter, event) {
		return false
	}
	if !yamlEmitterProcessAnchor(emitter) {
		return false
	}
	if !yamlEmitterProcessTag(emitter) {
		return false
	}
	if !yamlEmitterIncreaseIndent(emitter, true, false) {
		return false
	}
	if !yamlEmitterProcessScalar(emitter) {
		return false
	}
	emitter.indent = emitter.indents[len(emitter.indents)-1]
	emitter.indents = emitter.indents[:len(emitter.indents)-1]
	emitter.state = emitter.states[len(emitter.states)-1]
	emitter.states = emitter.states[:len(emitter.states)-1]
	return true
}

// Expect SEQUENCE-START.
func yamlEmitterEmitSequenceStart(emitter *yamlEmitterT, event *yamlEventT) bool {
	if !yamlEmitterProcessAnchor(emitter) {
		return false
	}
	if !yamlEmitterProcessTag(emitter) {
		return false
	}
	if emitter.flowLevel > 0 || emitter.canonical || event.sequenceStyle() == yamlFlowSequenceStyle ||
		yamlEmitterCheckEmptySequence(emitter) {
		emitter.state = yamlEmitFlowSequenceFirstItemState
	} else {
		emitter.state = yamlEmitBlockSequenceFirstItemState
	}
	return true
}

// Expect MAPPING-START.
func yamlEmitterEmitMappingStart(emitter *yamlEmitterT, event *yamlEventT) bool {
	if !yamlEmitterProcessAnchor(emitter) {
		return false
	}
	if !yamlEmitterProcessTag(emitter) {
		return false
	}
	if emitter.flowLevel > 0 || emitter.canonical || event.mappingStyle() == yamlFlowMappingStyle ||
		yamlEmitterCheckEmptyMapping(emitter) {
		emitter.state = yamlEmitFlowMappingFirstKeyState
	} else {
		emitter.state = yamlEmitBlockMappingFirstKeyState
	}
	return true
}

// Check if the document content is an empty scalar.
func yamlEmitterCheckEmptyDocument(emitter *yamlEmitterT) bool {
	return false // [Go] Huh?
}

// Check if the next events represent an empty sequence.
func yamlEmitterCheckEmptySequence(emitter *yamlEmitterT) bool {
	if len(emitter.events)-emitter.eventsHead < 2 {
		return false
	}
	return emitter.events[emitter.eventsHead].typ == yamlSequenceStartEvent &&
		emitter.events[emitter.eventsHead+1].typ == yamlSequenceEndEvent
}

// Check if the next events represent an empty mapping.
func yamlEmitterCheckEmptyMapping(emitter *yamlEmitterT) bool {
	if len(emitter.events)-emitter.eventsHead < 2 {
		return false
	}
	return emitter.events[emitter.eventsHead].typ == yamlMappingStartEvent &&
		emitter.events[emitter.eventsHead+1].typ == yamlMappingEndEvent
}

// Check if the next node can be expressed as a simple key.
func yamlEmitterCheckSimpleKey(emitter *yamlEmitterT) bool {
	length := 0
	switch emitter.events[emitter.eventsHead].typ {
	case yamlAliasEvent:
		length += len(emitter.anchorData.anchor)
	case yamlScalarEvent:
		if emitter.scalarData.multiline {
			return false
		}
		length += len(emitter.anchorData.anchor) +
			len(emitter.tagData.handle) +
			len(emitter.tagData.suffix) +
			len(emitter.scalarData.value)
	case yamlSequenceStartEvent:
		if !yamlEmitterCheckEmptySequence(emitter) {
			return false
		}
		length += len(emitter.anchorData.anchor) +
			len(emitter.tagData.handle) +
			len(emitter.tagData.suffix)
	case yamlMappingStartEvent:
		if !yamlEmitterCheckEmptyMapping(emitter) {
			return false
		}
		length += len(emitter.anchorData.anchor) +
			len(emitter.tagData.handle) +
			len(emitter.tagData.suffix)
	default:
		return false
	}
	return length <= 128
}

// Determine an acceptable scalar style.
func yamlEmitterSelectScalarStyle(emitter *yamlEmitterT, event *yamlEventT) bool {

	noTag := len(emitter.tagData.handle) == 0 && len(emitter.tagData.suffix) == 0
	if noTag && !event.implicit && !event.quotedImplicit {
		return yamlEmitterSetEmitterError(emitter, "neither tag nor implicit flags are specified")
	}

	style := event.scalarStyle()
	if style == yamlAnyScalarStyle {
		style = yamlPlainScalarStyle
	}
	if emitter.canonical {
		style = yamlDoubleQuotedScalarStyle
	}
	if emitter.simpleKeyContext && emitter.scalarData.multiline {
		style = yamlDoubleQuotedScalarStyle
	}

	if style == yamlPlainScalarStyle {
		if emitter.flowLevel > 0 && !emitter.scalarData.flowPlainAllowed ||
			emitter.flowLevel == 0 && !emitter.scalarData.blockPlainAllowed {
			style = yamlSingleQuotedScalarStyle
		}
		if len(emitter.scalarData.value) == 0 && (emitter.flowLevel > 0 || emitter.simpleKeyContext) {
			style = yamlSingleQuotedScalarStyle
		}
		if noTag && !event.implicit {
			style = yamlSingleQuotedScalarStyle
		}
	}
	if style == yamlSingleQuotedScalarStyle {
		if !emitter.scalarData.singleQuotedAllowed {
			style = yamlDoubleQuotedScalarStyle
		}
	}
	if style == yamlLiteralScalarStyle || style == yamlFoldedScalarStyle {
		if !emitter.scalarData.blockAllowed || emitter.flowLevel > 0 || emitter.simpleKeyContext {
			style = yamlDoubleQuotedScalarStyle
		}
	}

	if noTag && !event.quotedImplicit && style != yamlPlainScalarStyle {
		emitter.tagData.handle = []byte{'!'}
	}
	emitter.scalarData.style = style
	return true
}

// Write an anchor.
func yamlEmitterProcessAnchor(emitter *yamlEmitterT) bool {
	if emitter.anchorData.anchor == nil {
		return true
	}
	c := []byte{'&'}
	if emitter.anchorData.alias {
		c[0] = '*'
	}
	if !yamlEmitterWriteIndicator(emitter, c, true, false, false) {
		return false
	}
	return yamlEmitterWriteAnchor(emitter, emitter.anchorData.anchor)
}

// Write a tag.
func yamlEmitterProcessTag(emitter *yamlEmitterT) bool {
	if len(emitter.tagData.handle) == 0 && len(emitter.tagData.suffix) == 0 {
		return true
	}
	if len(emitter.tagData.handle) > 0 {
		if !yamlEmitterWriteTagHandle(emitter, emitter.tagData.handle) {
			return false
		}
		if len(emitter.tagData.suffix) > 0 {
			if !yamlEmitterWriteTagContent(emitter, emitter.tagData.suffix, false) {
				return false
			}
		}
	} else {
		// [Go] Allocate these slices elsewhere.
		if !yamlEmitterWriteIndicator(emitter, []byte("!<"), true, false, false) {
			return false
		}
		if !yamlEmitterWriteTagContent(emitter, emitter.tagData.suffix, false) {
			return false
		}
		if !yamlEmitterWriteIndicator(emitter, []byte{'>'}, false, false, false) {
			return false
		}
	}
	return true
}

// Write a scalar.
func yamlEmitterProcessScalar(emitter *yamlEmitterT) bool {
	switch emitter.scalarData.style {
	case yamlPlainScalarStyle:
		return yamlEmitterWritePlainScalar(emitter, emitter.scalarData.value, !emitter.simpleKeyContext)

	case yamlSingleQuotedScalarStyle:
		return yamlEmitterWriteSingleQuotedScalar(emitter, emitter.scalarData.value, !emitter.simpleKeyContext)

	case yamlDoubleQuotedScalarStyle:
		return yamlEmitterWriteDoubleQuotedScalar(emitter, emitter.scalarData.value, !emitter.simpleKeyContext)

	case yamlLiteralScalarStyle:
		return yamlEmitterWriteLiteralScalar(emitter, emitter.scalarData.value)

	case yamlFoldedScalarStyle:
		return yamlEmitterWriteFoldedScalar(emitter, emitter.scalarData.value)
	}
	panic("unknown scalar style")
}

// Check if a %YAML directive is valid.
func yamlEmitterAnalyzeVersionDirective(emitter *yamlEmitterT, versionDirective *yamlVersionDirectiveT) bool {
	if versionDirective.major != 1 || versionDirective.minor != 1 {
		return yamlEmitterSetEmitterError(emitter, "incompatible %YAML directive")
	}
	return true
}

// Check if a %TAG directive is valid.
func yamlEmitterAnalyzeTagDirective(emitter *yamlEmitterT, tagDirective *yamlTagDirectiveT) bool {
	handle := tagDirective.handle
	prefix := tagDirective.prefix
	if len(handle) == 0 {
		return yamlEmitterSetEmitterError(emitter, "tag handle must not be empty")
	}
	if handle[0] != '!' {
		return yamlEmitterSetEmitterError(emitter, "tag handle must start with '!'")
	}
	if handle[len(handle)-1] != '!' {
		return yamlEmitterSetEmitterError(emitter, "tag handle must end with '!'")
	}
	for i := 1; i < len(handle)-1; i += width(handle[i]) {
		if !isAlpha(handle, i) {
			return yamlEmitterSetEmitterError(emitter, "tag handle must contain alphanumerical characters only")
		}
	}
	if len(prefix) == 0 {
		return yamlEmitterSetEmitterError(emitter, "tag prefix must not be empty")
	}
	return true
}

// Check if an anchor is valid.
func yamlEmitterAnalyzeAnchor(emitter *yamlEmitterT, anchor []byte, alias bool) bool {
	if len(anchor) == 0 {
		problem := "anchor value must not be empty"
		if alias {
			problem = "alias value must not be empty"
		}
		return yamlEmitterSetEmitterError(emitter, problem)
	}
	for i := 0; i < len(anchor); i += width(anchor[i]) {
		if !isAlpha(anchor, i) {
			problem := "anchor value must contain alphanumerical characters only"
			if alias {
				problem = "alias value must contain alphanumerical characters only"
			}
			return yamlEmitterSetEmitterError(emitter, problem)
		}
	}
	emitter.anchorData.anchor = anchor
	emitter.anchorData.alias = alias
	return true
}

// Check if a tag is valid.
func yamlEmitterAnalyzeTag(emitter *yamlEmitterT, tag []byte) bool {
	if len(tag) == 0 {
		return yamlEmitterSetEmitterError(emitter, "tag value must not be empty")
	}
	for i := 0; i < len(emitter.tagDirectives); i++ {
		tagDirective := &emitter.tagDirectives[i]
		if bytes.HasPrefix(tag, tagDirective.prefix) {
			emitter.tagData.handle = tagDirective.handle
			emitter.tagData.suffix = tag[len(tagDirective.prefix):]
			return true
		}
	}
	emitter.tagData.suffix = tag
	return true
}

// Check if a scalar is valid.
func yamlEmitterAnalyzeScalar(emitter *yamlEmitterT, value []byte) bool {
	var (
		blockIndicators   = false
		flowIndicators    = false
		lineBreaks        = false
		specialCharacters = false

		leadingSpace  = false
		leadingBreak  = false
		trailingSpace = false
		trailingBreak = false
		breakSpace    = false
		spaceBreak    = false

		precededByWhitespace = false
		followedByWhitespace = false
		previousSpace        = false
		previousBreak        = false
	)

	emitter.scalarData.value = value

	if len(value) == 0 {
		emitter.scalarData.multiline = false
		emitter.scalarData.flowPlainAllowed = false
		emitter.scalarData.blockPlainAllowed = true
		emitter.scalarData.singleQuotedAllowed = true
		emitter.scalarData.blockAllowed = false
		return true
	}

	if len(value) >= 3 && ((value[0] == '-' && value[1] == '-' && value[2] == '-') || (value[0] == '.' && value[1] == '.' && value[2] == '.')) {
		blockIndicators = true
		flowIndicators = true
	}

	precededByWhitespace = true
	for i, w := 0, 0; i < len(value); i += w {
		w = width(value[i])
		followedByWhitespace = i+w >= len(value) || isBlank(value, i+w)

		if i == 0 {
			switch value[i] {
			case '#', ',', '[', ']', '{', '}', '&', '*', '!', '|', '>', '\'', '"', '%', '@', '`':
				flowIndicators = true
				blockIndicators = true
			case '?', ':':
				flowIndicators = true
				if followedByWhitespace {
					blockIndicators = true
				}
			case '-':
				if followedByWhitespace {
					flowIndicators = true
					blockIndicators = true
				}
			}
		} else {
			switch value[i] {
			case ',', '?', '[', ']', '{', '}':
				flowIndicators = true
			case ':':
				flowIndicators = true
				if followedByWhitespace {
					blockIndicators = true
				}
			case '#':
				if precededByWhitespace {
					flowIndicators = true
					blockIndicators = true
				}
			}
		}

		if !isPrintable(value, i) || !isASCII(value, i) && !emitter.unicode {
			specialCharacters = true
		}
		if isSpace(value, i) {
			if i == 0 {
				leadingSpace = true
			}
			if i+width(value[i]) == len(value) {
				trailingSpace = true
			}
			if previousBreak {
				breakSpace = true
			}
			previousSpace = true
			previousBreak = false
		} else if isBreak(value, i) {
			lineBreaks = true
			if i == 0 {
				leadingBreak = true
			}
			if i+width(value[i]) == len(value) {
				trailingBreak = true
			}
			if previousSpace {
				spaceBreak = true
			}
			previousSpace = false
			previousBreak = true
		} else {
			previousSpace = false
			previousBreak = false
		}

		// [Go]: Why 'z'? Couldn't be the end of the string as that's the loop condition.
		precededByWhitespace = isBlankz(value, i)
	}

	emitter.scalarData.multiline = lineBreaks
	emitter.scalarData.flowPlainAllowed = true
	emitter.scalarData.blockPlainAllowed = true
	emitter.scalarData.singleQuotedAllowed = true
	emitter.scalarData.blockAllowed = true

	if leadingSpace || leadingBreak || trailingSpace || trailingBreak {
		emitter.scalarData.flowPlainAllowed = false
		emitter.scalarData.blockPlainAllowed = false
	}
	if trailingSpace {
		emitter.scalarData.blockAllowed = false
	}
	if breakSpace {
		emitter.scalarData.flowPlainAllowed = false
		emitter.scalarData.blockPlainAllowed = false
		emitter.scalarData.singleQuotedAllowed = false
	}
	if spaceBreak || specialCharacters {
		emitter.scalarData.flowPlainAllowed = false
		emitter.scalarData.blockPlainAllowed = false
		emitter.scalarData.singleQuotedAllowed = false
		emitter.scalarData.blockAllowed = false
	}
	if lineBreaks {
		emitter.scalarData.flowPlainAllowed = false
		emitter.scalarData.blockPlainAllowed = false
	}
	if flowIndicators {
		emitter.scalarData.flowPlainAllowed = false
	}
	if blockIndicators {
		emitter.scalarData.blockPlainAllowed = false
	}
	return true
}

// Check if the event data is valid.
func yamlEmitterAnalyzeEvent(emitter *yamlEmitterT, event *yamlEventT) bool {

	emitter.anchorData.anchor = nil
	emitter.tagData.handle = nil
	emitter.tagData.suffix = nil
	emitter.scalarData.value = nil

	switch event.typ {
	case yamlAliasEvent:
		if !yamlEmitterAnalyzeAnchor(emitter, event.anchor, true) {
			return false
		}

	case yamlScalarEvent:
		if len(event.anchor) > 0 {
			if !yamlEmitterAnalyzeAnchor(emitter, event.anchor, false) {
				return false
			}
		}
		if len(event.tag) > 0 && (emitter.canonical || (!event.implicit && !event.quotedImplicit)) {
			if !yamlEmitterAnalyzeTag(emitter, event.tag) {
				return false
			}
		}
		if !yamlEmitterAnalyzeScalar(emitter, event.value) {
			return false
		}

	case yamlSequenceStartEvent:
		if len(event.anchor) > 0 {
			if !yamlEmitterAnalyzeAnchor(emitter, event.anchor, false) {
				return false
			}
		}
		if len(event.tag) > 0 && (emitter.canonical || !event.implicit) {
			if !yamlEmitterAnalyzeTag(emitter, event.tag) {
				return false
			}
		}

	case yamlMappingStartEvent:
		if len(event.anchor) > 0 {
			if !yamlEmitterAnalyzeAnchor(emitter, event.anchor, false) {
				return false
			}
		}
		if len(event.tag) > 0 && (emitter.canonical || !event.implicit) {
			if !yamlEmitterAnalyzeTag(emitter, event.tag) {
				return false
			}
		}
	}
	return true
}

// Write the BOM character.
func yamlEmitterWriteBom(emitter *yamlEmitterT) bool {
	if !flush(emitter) {
		return false
	}
	pos := emitter.bufferPos
	emitter.buffer[pos+0] = '\xEF'
	emitter.buffer[pos+1] = '\xBB'
	emitter.buffer[pos+2] = '\xBF'
	emitter.bufferPos += 3
	return true
}

func yamlEmitterWriteIndent(emitter *yamlEmitterT) bool {
	indent := emitter.indent
	if indent < 0 {
		indent = 0
	}
	if !emitter.indention || emitter.column > indent || (emitter.column == indent && !emitter.whitespace) {
		if !putBreak(emitter) {
			return false
		}
	}
	for emitter.column < indent {
		if !put(emitter, ' ') {
			return false
		}
	}
	emitter.whitespace = true
	emitter.indention = true
	return true
}

func yamlEmitterWriteIndicator(emitter *yamlEmitterT, indicator []byte, needWhitespace, isWhitespace, isIndention bool) bool {
	if needWhitespace && !emitter.whitespace {
		if !put(emitter, ' ') {
			return false
		}
	}
	if !writeAll(emitter, indicator) {
		return false
	}
	emitter.whitespace = isWhitespace
	emitter.indention = (emitter.indention && isIndention)
	emitter.openEnded = false
	return true
}

func yamlEmitterWriteAnchor(emitter *yamlEmitterT, value []byte) bool {
	if !writeAll(emitter, value) {
		return false
	}
	emitter.whitespace = false
	emitter.indention = false
	return true
}

func yamlEmitterWriteTagHandle(emitter *yamlEmitterT, value []byte) bool {
	if !emitter.whitespace {
		if !put(emitter, ' ') {
			return false
		}
	}
	if !writeAll(emitter, value) {
		return false
	}
	emitter.whitespace = false
	emitter.indention = false
	return true
}

func yamlEmitterWriteTagContent(emitter *yamlEmitterT, value []byte, needWhitespace bool) bool {
	if needWhitespace && !emitter.whitespace {
		if !put(emitter, ' ') {
			return false
		}
	}
	for i := 0; i < len(value); {
		var mustWrite bool
		switch value[i] {
		case ';', '/', '?', ':', '@', '&', '=', '+', '$', ',', '_', '.', '~', '*', '\'', '(', ')', '[', ']':
			mustWrite = true
		default:
			mustWrite = isAlpha(value, i)
		}
		if mustWrite {
			if !write(emitter, value, &i) {
				return false
			}
		} else {
			w := width(value[i])
			for k := 0; k < w; k++ {
				octet := value[i]
				i++
				if !put(emitter, '%') {
					return false
				}

				c := octet >> 4
				if c < 10 {
					c += '0'
				} else {
					c += 'A' - 10
				}
				if !put(emitter, c) {
					return false
				}

				c = octet & 0x0f
				if c < 10 {
					c += '0'
				} else {
					c += 'A' - 10
				}
				if !put(emitter, c) {
					return false
				}
			}
		}
	}
	emitter.whitespace = false
	emitter.indention = false
	return true
}

func yamlEmitterWritePlainScalar(emitter *yamlEmitterT, value []byte, allowBreaks bool) bool {
	if !emitter.whitespace {
		if !put(emitter, ' ') {
			return false
		}
	}

	spaces := false
	breaks := false
	for i := 0; i < len(value); {
		if isSpace(value, i) {
			if allowBreaks && !spaces && emitter.column > emitter.bestWidth && !isSpace(value, i+1) {
				if !yamlEmitterWriteIndent(emitter) {
					return false
				}
				i += width(value[i])
			} else {
				if !write(emitter, value, &i) {
					return false
				}
			}
			spaces = true
		} else if isBreak(value, i) {
			if !breaks && value[i] == '\n' {
				if !putBreak(emitter) {
					return false
				}
			}
			if !writeBreak(emitter, value, &i) {
				return false
			}
			emitter.indention = true
			breaks = true
		} else {
			if breaks {
				if !yamlEmitterWriteIndent(emitter) {
					return false
				}
			}
			if !write(emitter, value, &i) {
				return false
			}
			emitter.indention = false
			spaces = false
			breaks = false
		}
	}

	emitter.whitespace = false
	emitter.indention = false
	if emitter.rootContext {
		emitter.openEnded = true
	}

	return true
}

func yamlEmitterWriteSingleQuotedScalar(emitter *yamlEmitterT, value []byte, allowBreaks bool) bool {

	if !yamlEmitterWriteIndicator(emitter, []byte{'\''}, true, false, false) {
		return false
	}

	spaces := false
	breaks := false
	for i := 0; i < len(value); {
		if isSpace(value, i) {
			if allowBreaks && !spaces && emitter.column > emitter.bestWidth && i > 0 && i < len(value)-1 && !isSpace(value, i+1) {
				if !yamlEmitterWriteIndent(emitter) {
					return false
				}
				i += width(value[i])
			} else {
				if !write(emitter, value, &i) {
					return false
				}
			}
			spaces = true
		} else if isBreak(value, i) {
			if !breaks && value[i] == '\n' {
				if !putBreak(emitter) {
					return false
				}
			}
			if !writeBreak(emitter, value, &i) {
				return false
			}
			emitter.indention = true
			breaks = true
		} else {
			if breaks {
				if !yamlEmitterWriteIndent(emitter) {
					return false
				}
			}
			if value[i] == '\'' {
				if !put(emitter, '\'') {
					return false
				}
			}
			if !write(emitter, value, &i) {
				return false
			}
			emitter.indention = false
			spaces = false
			breaks = false
		}
	}
	if !yamlEmitterWriteIndicator(emitter, []byte{'\''}, false, false, false) {
		return false
	}
	emitter.whitespace = false
	emitter.indention = false
	return true
}

func yamlEmitterWriteDoubleQuotedScalar(emitter *yamlEmitterT, value []byte, allowBreaks bool) bool {
	spaces := false
	if !yamlEmitterWriteIndicator(emitter, []byte{'"'}, true, false, false) {
		return false
	}

	for i := 0; i < len(value); {
		if !isPrintable(value, i) || (!emitter.unicode && !isASCII(value, i)) ||
			isBom(value, i) || isBreak(value, i) ||
			value[i] == '"' || value[i] == '\\' {

			octet := value[i]

			var w int
			var v rune
			switch {
			case octet&0x80 == 0x00:
				w, v = 1, rune(octet&0x7F)
			case octet&0xE0 == 0xC0:
				w, v = 2, rune(octet&0x1F)
			case octet&0xF0 == 0xE0:
				w, v = 3, rune(octet&0x0F)
			case octet&0xF8 == 0xF0:
				w, v = 4, rune(octet&0x07)
			}
			for k := 1; k < w; k++ {
				octet = value[i+k]
				v = (v << 6) + (rune(octet) & 0x3F)
			}
			i += w

			if !put(emitter, '\\') {
				return false
			}

			var ok bool
			switch v {
			case 0x00:
				ok = put(emitter, '0')
			case 0x07:
				ok = put(emitter, 'a')
			case 0x08:
				ok = put(emitter, 'b')
			case 0x09:
				ok = put(emitter, 't')
			case 0x0A:
				ok = put(emitter, 'n')
			case 0x0b:
				ok = put(emitter, 'v')
			case 0x0c:
				ok = put(emitter, 'f')
			case 0x0d:
				ok = put(emitter, 'r')
			case 0x1b:
				ok = put(emitter, 'e')
			case 0x22:
				ok = put(emitter, '"')
			case 0x5c:
				ok = put(emitter, '\\')
			case 0x85:
				ok = put(emitter, 'N')
			case 0xA0:
				ok = put(emitter, '_')
			case 0x2028:
				ok = put(emitter, 'L')
			case 0x2029:
				ok = put(emitter, 'P')
			default:
				if v <= 0xFF {
					ok = put(emitter, 'x')
					w = 2
				} else if v <= 0xFFFF {
					ok = put(emitter, 'u')
					w = 4
				} else {
					ok = put(emitter, 'U')
					w = 8
				}
				for k := (w - 1) * 4; ok && k >= 0; k -= 4 {
					digit := byte((v >> uint(k)) & 0x0F)
					if digit < 10 {
						ok = put(emitter, digit+'0')
					} else {
						ok = put(emitter, digit+'A'-10)
					}
				}
			}
			if !ok {
				return false
			}
			spaces = false
		} else if isSpace(value, i) {
			if allowBreaks && !spaces && emitter.column > emitter.bestWidth && i > 0 && i < len(value)-1 {
				if !yamlEmitterWriteIndent(emitter) {
					return false
				}
				if isSpace(value, i+1) {
					if !put(emitter, '\\') {
						return false
					}
				}
				i += width(value[i])
			} else if !write(emitter, value, &i) {
				return false
			}
			spaces = true
		} else {
			if !write(emitter, value, &i) {
				return false
			}
			spaces = false
		}
	}
	if !yamlEmitterWriteIndicator(emitter, []byte{'"'}, false, false, false) {
		return false
	}
	emitter.whitespace = false
	emitter.indention = false
	return true
}

func yamlEmitterWriteBlockScalarHints(emitter *yamlEmitterT, value []byte) bool {
	if isSpace(value, 0) || isBreak(value, 0) {
		indentHint := []byte{'0' + byte(emitter.bestIndent)}
		if !yamlEmitterWriteIndicator(emitter, indentHint, false, false, false) {
			return false
		}
	}

	emitter.openEnded = false

	var chompHint [1]byte
	if len(value) == 0 {
		chompHint[0] = '-'
	} else {
		i := len(value) - 1
		for value[i]&0xC0 == 0x80 {
			i--
		}
		if !isBreak(value, i) {
			chompHint[0] = '-'
		} else if i == 0 {
			chompHint[0] = '+'
			emitter.openEnded = true
		} else {
			i--
			for value[i]&0xC0 == 0x80 {
				i--
			}
			if isBreak(value, i) {
				chompHint[0] = '+'
				emitter.openEnded = true
			}
		}
	}
	if chompHint[0] != 0 {
		if !yamlEmitterWriteIndicator(emitter, chompHint[:], false, false, false) {
			return false
		}
	}
	return true
}

func yamlEmitterWriteLiteralScalar(emitter *yamlEmitterT, value []byte) bool {
	if !yamlEmitterWriteIndicator(emitter, []byte{'|'}, true, false, false) {
		return false
	}
	if !yamlEmitterWriteBlockScalarHints(emitter, value) {
		return false
	}
	if !putBreak(emitter) {
		return false
	}
	emitter.indention = true
	emitter.whitespace = true
	breaks := true
	for i := 0; i < len(value); {
		if isBreak(value, i) {
			if !writeBreak(emitter, value, &i) {
				return false
			}
			emitter.indention = true
			breaks = true
		} else {
			if breaks {
				if !yamlEmitterWriteIndent(emitter) {
					return false
				}
			}
			if !write(emitter, value, &i) {
				return false
			}
			emitter.indention = false
			breaks = false
		}
	}

	return true
}

func yamlEmitterWriteFoldedScalar(emitter *yamlEmitterT, value []byte) bool {
	if !yamlEmitterWriteIndicator(emitter, []byte{'>'}, true, false, false) {
		return false
	}
	if !yamlEmitterWriteBlockScalarHints(emitter, value) {
		return false
	}

	if !putBreak(emitter) {
		return false
	}
	emitter.indention = true
	emitter.whitespace = true

	breaks := true
	leadingSpaces := true
	for i := 0; i < len(value); {
		if isBreak(value, i) {
			if !breaks && !leadingSpaces && value[i] == '\n' {
				k := 0
				for isBreak(value, k) {
					k += width(value[k])
				}
				if !isBlankz(value, k) {
					if !putBreak(emitter) {
						return false
					}
				}
			}
			if !writeBreak(emitter, value, &i) {
				return false
			}
			emitter.indention = true
			breaks = true
		} else {
			if breaks {
				if !yamlEmitterWriteIndent(emitter) {
					return false
				}
				leadingSpaces = isBlank(value, i)
			}
			if !breaks && isSpace(value, i) && !isSpace(value, i+1) && emitter.column > emitter.bestWidth {
				if !yamlEmitterWriteIndent(emitter) {
					return false
				}
				i += width(value[i])
			} else {
				if !write(emitter, value, &i) {
					return false
				}
			}
			emitter.indention = false
			breaks = false
		}
	}
	return true
}
