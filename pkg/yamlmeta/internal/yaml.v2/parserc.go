// Copyright 2020 VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package yaml

import (
	"bytes"
)

// The parser implements the following grammar:
//
// stream               ::= STREAM-START implicit_document? explicit_document* STREAM-END
// implicit_document    ::= block_node DOCUMENT-END*
// explicit_document    ::= DIRECTIVE* DOCUMENT-START block_node? DOCUMENT-END*
// block_node_or_indentless_sequence    ::=
//                          ALIAS
//                          | properties (block_content | indentless_block_sequence)?
//                          | block_content
//                          | indentless_block_sequence
// block_node           ::= ALIAS
//                          | properties block_content?
//                          | block_content
// flow_node            ::= ALIAS
//                          | properties flow_content?
//                          | flow_content
// properties           ::= TAG ANCHOR? | ANCHOR TAG?
// block_content        ::= block_collection | flow_collection | SCALAR
// flow_content         ::= flow_collection | SCALAR
// block_collection     ::= block_sequence | block_mapping
// flow_collection      ::= flow_sequence | flow_mapping
// block_sequence       ::= BLOCK-SEQUENCE-START (BLOCK-ENTRY block_node?)* BLOCK-END
// indentless_sequence  ::= (BLOCK-ENTRY block_node?)+
// block_mapping        ::= BLOCK-MAPPING_START
//                          ((KEY block_node_or_indentless_sequence?)?
//                          (VALUE block_node_or_indentless_sequence?)?)*
//                          BLOCK-END
// flow_sequence        ::= FLOW-SEQUENCE-START
//                          (flow_sequence_entry FLOW-ENTRY)*
//                          flow_sequence_entry?
//                          FLOW-SEQUENCE-END
// flow_sequence_entry  ::= flow_node | KEY flow_node? (VALUE flow_node?)?
// flow_mapping         ::= FLOW-MAPPING-START
//                          (flow_mapping_entry FLOW-ENTRY)*
//                          flow_mapping_entry?
//                          FLOW-MAPPING-END
// flow_mapping_entry   ::= flow_node | KEY flow_node? (VALUE flow_node?)?

// Peek the next token in the token queue.
func peekToken(parser *yamlParserT) *yamlTokenT {
	if parser.tokenAvailable || yamlParserFetchMoreTokens(parser) {
		return &parser.tokens[parser.tokensHead]
	}
	return nil
}

// Remove the next token from the queue (must be called after peek_token).
func skipToken(parser *yamlParserT) {
	parser.tokenAvailable = false
	parser.tokensParsed++
	parser.streamEndProduced = parser.tokens[parser.tokensHead].typ == yamlStreamEndToken
	parser.tokensHead++
}

// Get the next event.
func yamlParserParse(parser *yamlParserT, event *yamlEventT) bool {
	// Erase the event object.
	*event = yamlEventT{}

	// No events after the end of the stream or error.
	if parser.streamEndProduced || parser.error != yamlNoError || parser.state == yamlParseEndState {
		return true
	}

	// Generate the next event.
	return yamlParserStateMachine(parser, event)
}

// Set parser error.
func yamlParserSetParserError(parser *yamlParserT, problem string, problemMark yamlMarkT) bool {
	parser.error = yamlParserError
	parser.problem = problem
	parser.problemMark = problemMark
	return false
}

func yamlParserSetParserErrorContext(parser *yamlParserT, context string, contextMark yamlMarkT, problem string, problemMark yamlMarkT) bool {
	parser.error = yamlParserError
	parser.context = context
	parser.contextMark = contextMark
	parser.problem = problem
	parser.problemMark = problemMark
	return false
}

// State dispatcher.
func yamlParserStateMachine(parser *yamlParserT, event *yamlEventT) bool {
	//trace("yaml_parser_state_machine", "state:", parser.state.String())

	switch parser.state {
	case yamlParseStreamStartState:
		return yamlParserParseStreamStart(parser, event)

	case yamlParseImplicitDocumentStartState:
		return yamlParserParseDocumentStart(parser, event, true)

	case yamlParseDocumentStartState:
		return yamlParserParseDocumentStart(parser, event, false)

	case yamlParseDocumentContentState:
		return yamlParserParseDocumentContent(parser, event)

	case yamlParseDocumentEndState:
		return yamlParserParseDocumentEnd(parser, event)

	case yamlParseBlockNodeState:
		return yamlParserParseNode(parser, event, true, false)

	case yamlParseBlockNodeOrIndentlessSequenceState:
		return yamlParserParseNode(parser, event, true, true)

	case yamlParseFlowNodeState:
		return yamlParserParseNode(parser, event, false, false)

	case yamlParseBlockSequenceFirstEntryState:
		return yamlParserParseBlockSequenceEntry(parser, event, true)

	case yamlParseBlockSequenceEntryState:
		return yamlParserParseBlockSequenceEntry(parser, event, false)

	case yamlParseIndentlessSequenceEntryState:
		return yamlParserParseIndentlessSequenceEntry(parser, event)

	case yamlParseBlockMappingFirstKeyState:
		return yamlParserParseBlockMappingKey(parser, event, true)

	case yamlParseBlockMappingKeyState:
		return yamlParserParseBlockMappingKey(parser, event, false)

	case yamlParseBlockMappingValueState:
		return yamlParserParseBlockMappingValue(parser, event)

	case yamlParseFlowSequenceFirstEntryState:
		return yamlParserParseFlowSequenceEntry(parser, event, true)

	case yamlParseFlowSequenceEntryState:
		return yamlParserParseFlowSequenceEntry(parser, event, false)

	case yamlParseFlowSequenceEntryMappingKeyState:
		return yamlParserParseFlowSequenceEntryMappingKey(parser, event)

	case yamlParseFlowSequenceEntryMappingValueState:
		return yamlParserParseFlowSequenceEntryMappingValue(parser, event)

	case yamlParseFlowSequenceEntryMappingEndState:
		return yamlParserParseFlowSequenceEntryMappingEnd(parser, event)

	case yamlParseFlowMappingFirstKeyState:
		return yamlParserParseFlowMappingKey(parser, event, true)

	case yamlParseFlowMappingKeyState:
		return yamlParserParseFlowMappingKey(parser, event, false)

	case yamlParseFlowMappingValueState:
		return yamlParserParseFlowMappingValue(parser, event, false)

	case yamlParseFlowMappingEmptyValueState:
		return yamlParserParseFlowMappingValue(parser, event, true)

	default:
		panic("invalid parser state")
	}
}

// Parse the production:
// stream   ::= STREAM-START implicit_document? explicit_document* STREAM-END
//              ************
func yamlParserParseStreamStart(parser *yamlParserT, event *yamlEventT) bool {
	token := peekToken(parser)
	if token == nil {
		return false
	}
	if token.typ != yamlStreamStartToken {
		return yamlParserSetParserError(parser, "did not find expected <stream-start>", token.startMark)
	}
	parser.state = yamlParseImplicitDocumentStartState
	*event = yamlEventT{
		typ:       yamlStreamStartEvent,
		startMark: token.startMark,
		endMark:   token.endMark,
		encoding:  token.encoding,
	}
	skipToken(parser)
	return true
}

// Parse the productions:
// implicit_document    ::= block_node DOCUMENT-END*
//                          *
// explicit_document    ::= DIRECTIVE* DOCUMENT-START block_node? DOCUMENT-END*
//                          *************************
func yamlParserParseDocumentStart(parser *yamlParserT, event *yamlEventT, implicit bool) bool {

	token := peekToken(parser)
	if token == nil {
		return false
	}

	// Parse extra document end indicators.
	if !implicit {
		for token.typ == yamlDocumentEndToken {
			skipToken(parser)
			token = peekToken(parser)
			if token == nil {
				return false
			}
		}
	}

	if implicit && token.typ != yamlVersionDirectiveToken &&
		token.typ != yamlTagDirectiveToken &&
		token.typ != yamlDocumentStartToken &&
		token.typ != yamlStreamEndToken {
		// Parse an implicit document.
		if !yamlParserProcessDirectives(parser, nil, nil) {
			return false
		}
		parser.states = append(parser.states, yamlParseDocumentEndState)
		parser.state = yamlParseBlockNodeState

		*event = yamlEventT{
			typ:       yamlDocumentStartEvent,
			startMark: token.startMark,
			endMark:   token.endMark,
		}

	} else if token.typ != yamlStreamEndToken {
		// Parse an explicit document.
		var versionDirective *yamlVersionDirectiveT
		var tagDirectives []yamlTagDirectiveT
		startMark := token.startMark
		if !yamlParserProcessDirectives(parser, &versionDirective, &tagDirectives) {
			return false
		}
		token = peekToken(parser)
		if token == nil {
			return false
		}
		if token.typ != yamlDocumentStartToken {
			yamlParserSetParserError(parser,
				"did not find expected <document start>", token.startMark)
			return false
		}
		parser.states = append(parser.states, yamlParseDocumentEndState)
		parser.state = yamlParseDocumentContentState
		endMark := token.endMark

		*event = yamlEventT{
			typ:              yamlDocumentStartEvent,
			startMark:        startMark,
			endMark:          endMark,
			versionDirective: versionDirective,
			tagDirectives:    tagDirectives,
			implicit:         false,
		}
		skipToken(parser)

	} else {
		// Parse the stream end.
		parser.state = yamlParseEndState
		*event = yamlEventT{
			typ:       yamlStreamEndEvent,
			startMark: token.startMark,
			endMark:   token.endMark,
		}
		skipToken(parser)
	}

	return true
}

// Parse the productions:
// explicit_document    ::= DIRECTIVE* DOCUMENT-START block_node? DOCUMENT-END*
//                                                    ***********
//
func yamlParserParseDocumentContent(parser *yamlParserT, event *yamlEventT) bool {
	token := peekToken(parser)
	if token == nil {
		return false
	}
	if token.typ == yamlVersionDirectiveToken ||
		token.typ == yamlTagDirectiveToken ||
		token.typ == yamlDocumentStartToken ||
		token.typ == yamlDocumentEndToken ||
		token.typ == yamlStreamEndToken {
		parser.state = parser.states[len(parser.states)-1]
		parser.states = parser.states[:len(parser.states)-1]
		return yamlParserProcessEmptyScalar(parser, event,
			token.startMark)
	}
	return yamlParserParseNode(parser, event, true, false)
}

// Parse the productions:
// implicit_document    ::= block_node DOCUMENT-END*
//                                     *************
// explicit_document    ::= DIRECTIVE* DOCUMENT-START block_node? DOCUMENT-END*
//
func yamlParserParseDocumentEnd(parser *yamlParserT, event *yamlEventT) bool {
	token := peekToken(parser)
	if token == nil {
		return false
	}

	startMark := token.startMark
	endMark := token.startMark

	implicit := true
	if token.typ == yamlDocumentEndToken {
		endMark = token.endMark
		skipToken(parser)
		implicit = false
	}

	parser.tagDirectives = parser.tagDirectives[:0]

	parser.state = yamlParseDocumentStartState
	*event = yamlEventT{
		typ:       yamlDocumentEndEvent,
		startMark: startMark,
		endMark:   endMark,
		implicit:  implicit,
	}
	return true
}

// Parse the productions:
// block_node_or_indentless_sequence    ::=
//                          ALIAS
//                          *****
//                          | properties (block_content | indentless_block_sequence)?
//                            **********  *
//                          | block_content | indentless_block_sequence
//                            *
// block_node           ::= ALIAS
//                          *****
//                          | properties block_content?
//                            ********** *
//                          | block_content
//                            *
// flow_node            ::= ALIAS
//                          *****
//                          | properties flow_content?
//                            ********** *
//                          | flow_content
//                            *
// properties           ::= TAG ANCHOR? | ANCHOR TAG?
//                          *************************
// block_content        ::= block_collection | flow_collection | SCALAR
//                                                               ******
// flow_content         ::= flow_collection | SCALAR
//                                            ******
func yamlParserParseNode(parser *yamlParserT, event *yamlEventT, block, indentlessSequence bool) bool {
	//defer trace("yaml_parser_parse_node", "block:", block, "indentless_sequence:", indentless_sequence)()

	token := peekToken(parser)
	if token == nil {
		return false
	}

	if token.typ == yamlAliasToken {
		parser.state = parser.states[len(parser.states)-1]
		parser.states = parser.states[:len(parser.states)-1]
		*event = yamlEventT{
			typ:       yamlAliasEvent,
			startMark: token.startMark,
			endMark:   token.endMark,
			anchor:    token.value,
		}
		skipToken(parser)
		return true
	}

	startMark := token.startMark
	endMark := token.startMark

	var tagToken bool
	var tagHandle, tagSuffix, anchor []byte
	var tagMark yamlMarkT
	if token.typ == yamlAnchorToken {
		anchor = token.value
		startMark = token.startMark
		endMark = token.endMark
		skipToken(parser)
		token = peekToken(parser)
		if token == nil {
			return false
		}
		if token.typ == yamlTagToken {
			tagToken = true
			tagHandle = token.value
			tagSuffix = token.suffix
			tagMark = token.startMark
			endMark = token.endMark
			skipToken(parser)
			token = peekToken(parser)
			if token == nil {
				return false
			}
		}
	} else if token.typ == yamlTagToken {
		tagToken = true
		tagHandle = token.value
		tagSuffix = token.suffix
		startMark = token.startMark
		tagMark = token.startMark
		endMark = token.endMark
		skipToken(parser)
		token = peekToken(parser)
		if token == nil {
			return false
		}
		if token.typ == yamlAnchorToken {
			anchor = token.value
			endMark = token.endMark
			skipToken(parser)
			token = peekToken(parser)
			if token == nil {
				return false
			}
		}
	}

	var tag []byte
	if tagToken {
		if len(tagHandle) == 0 {
			tag = tagSuffix
			tagSuffix = nil
		} else {
			for i := range parser.tagDirectives {
				if bytes.Equal(parser.tagDirectives[i].handle, tagHandle) {
					tag = append([]byte(nil), parser.tagDirectives[i].prefix...)
					tag = append(tag, tagSuffix...)
					break
				}
			}
			if len(tag) == 0 {
				yamlParserSetParserErrorContext(parser,
					"while parsing a node", startMark,
					"found undefined tag handle", tagMark)
				return false
			}
		}
	}

	implicit := len(tag) == 0
	if indentlessSequence && token.typ == yamlBlockEntryToken {
		endMark = token.endMark
		parser.state = yamlParseIndentlessSequenceEntryState
		*event = yamlEventT{
			typ:       yamlSequenceStartEvent,
			startMark: startMark,
			endMark:   endMark,
			anchor:    anchor,
			tag:       tag,
			implicit:  implicit,
			style:     yamlStyleT(yamlBlockSequenceStyle),
		}
		return true
	}
	if token.typ == yamlScalarToken {
		var plainImplicit, quotedImplicit bool
		endMark = token.endMark
		if (len(tag) == 0 && token.style == yamlPlainScalarStyle) || (len(tag) == 1 && tag[0] == '!') {
			plainImplicit = true
		} else if len(tag) == 0 {
			quotedImplicit = true
		}
		parser.state = parser.states[len(parser.states)-1]
		parser.states = parser.states[:len(parser.states)-1]

		*event = yamlEventT{
			typ:            yamlScalarEvent,
			startMark:      startMark,
			endMark:        endMark,
			anchor:         anchor,
			tag:            tag,
			value:          token.value,
			implicit:       plainImplicit,
			quotedImplicit: quotedImplicit,
			style:          yamlStyleT(token.style),
		}
		skipToken(parser)
		return true
	}
	if token.typ == yamlFlowSequenceStartToken {
		// [Go] Some of the events below can be merged as they differ only on style.
		endMark = token.endMark
		parser.state = yamlParseFlowSequenceFirstEntryState
		*event = yamlEventT{
			typ:       yamlSequenceStartEvent,
			startMark: startMark,
			endMark:   endMark,
			anchor:    anchor,
			tag:       tag,
			implicit:  implicit,
			style:     yamlStyleT(yamlFlowSequenceStyle),
		}
		return true
	}
	if token.typ == yamlFlowMappingStartToken {
		endMark = token.endMark
		parser.state = yamlParseFlowMappingFirstKeyState
		*event = yamlEventT{
			typ:       yamlMappingStartEvent,
			startMark: startMark,
			endMark:   endMark,
			anchor:    anchor,
			tag:       tag,
			implicit:  implicit,
			style:     yamlStyleT(yamlFlowMappingStyle),
		}
		return true
	}
	if block && token.typ == yamlBlockSequenceStartToken {
		endMark = token.endMark
		parser.state = yamlParseBlockSequenceFirstEntryState
		*event = yamlEventT{
			typ:       yamlSequenceStartEvent,
			startMark: startMark,
			endMark:   endMark,
			anchor:    anchor,
			tag:       tag,
			implicit:  implicit,
			style:     yamlStyleT(yamlBlockSequenceStyle),
		}
		return true
	}
	if block && token.typ == yamlBlockMappingStartToken {
		endMark = token.endMark
		parser.state = yamlParseBlockMappingFirstKeyState
		*event = yamlEventT{
			typ:       yamlMappingStartEvent,
			startMark: startMark,
			endMark:   endMark,
			anchor:    anchor,
			tag:       tag,
			implicit:  implicit,
			style:     yamlStyleT(yamlBlockMappingStyle),
		}
		return true
	}
	if len(anchor) > 0 || len(tag) > 0 {
		parser.state = parser.states[len(parser.states)-1]
		parser.states = parser.states[:len(parser.states)-1]

		*event = yamlEventT{
			typ:            yamlScalarEvent,
			startMark:      startMark,
			endMark:        endMark,
			anchor:         anchor,
			tag:            tag,
			implicit:       implicit,
			quotedImplicit: false,
			style:          yamlStyleT(yamlPlainScalarStyle),
		}
		return true
	}

	context := "while parsing a flow node"
	if block {
		context = "while parsing a block node"
	}
	yamlParserSetParserErrorContext(parser, context, startMark,
		"did not find expected node content", token.startMark)
	return false
}

// Parse the productions:
// block_sequence ::= BLOCK-SEQUENCE-START (BLOCK-ENTRY block_node?)* BLOCK-END
//                    ********************  *********** *             *********
//
func yamlParserParseBlockSequenceEntry(parser *yamlParserT, event *yamlEventT, first bool) bool {
	if first {
		token := peekToken(parser)
		parser.marks = append(parser.marks, token.startMark)
		skipToken(parser)
	}

	token := peekToken(parser)
	if token == nil {
		return false
	}

	if token.typ == yamlBlockEntryToken {
		parser.pendingSeqItemEvent = &yamlEventT{
			startMark: token.startMark,
			endMark:   token.endMark,
		}
		mark := token.endMark
		skipToken(parser)
		token = peekToken(parser)
		if token == nil {
			return false
		}
		if token.typ != yamlBlockEntryToken && token.typ != yamlBlockEndToken {
			parser.states = append(parser.states, yamlParseBlockSequenceEntryState)
			return yamlParserParseNode(parser, event, true, false)
		}
		parser.state = yamlParseBlockSequenceEntryState
		return yamlParserProcessEmptyScalar(parser, event, mark)
	}
	if token.typ == yamlBlockEndToken {
		parser.state = parser.states[len(parser.states)-1]
		parser.states = parser.states[:len(parser.states)-1]
		parser.marks = parser.marks[:len(parser.marks)-1]

		*event = yamlEventT{
			typ:       yamlSequenceEndEvent,
			startMark: token.startMark,
			endMark:   token.endMark,
		}

		skipToken(parser)
		return true
	}

	contextMark := parser.marks[len(parser.marks)-1]
	parser.marks = parser.marks[:len(parser.marks)-1]
	return yamlParserSetParserErrorContext(parser,
		"while parsing a block collection", contextMark,
		"did not find expected '-' indicator", token.startMark)
}

// Parse the productions:
// indentless_sequence  ::= (BLOCK-ENTRY block_node?)+
//                           *********** *
func yamlParserParseIndentlessSequenceEntry(parser *yamlParserT, event *yamlEventT) bool {
	token := peekToken(parser)
	if token == nil {
		return false
	}

	if token.typ == yamlBlockEntryToken {
		parser.pendingSeqItemEvent = &yamlEventT{
			startMark: token.startMark,
			endMark:   token.endMark,
		}
		mark := token.endMark
		skipToken(parser)
		token = peekToken(parser)
		if token == nil {
			return false
		}
		if token.typ != yamlBlockEntryToken &&
			token.typ != yamlKeyToken &&
			token.typ != yamlValueToken &&
			token.typ != yamlBlockEndToken {
			parser.states = append(parser.states, yamlParseIndentlessSequenceEntryState)
			return yamlParserParseNode(parser, event, true, false)
		}
		parser.state = yamlParseIndentlessSequenceEntryState
		return yamlParserProcessEmptyScalar(parser, event, mark)
	}
	parser.state = parser.states[len(parser.states)-1]
	parser.states = parser.states[:len(parser.states)-1]

	*event = yamlEventT{
		typ:       yamlSequenceEndEvent,
		startMark: token.startMark,
		endMark:   token.startMark, // [Go] Shouldn't this be token.end_mark?
	}
	return true
}

// Parse the productions:
// block_mapping        ::= BLOCK-MAPPING_START
//                          *******************
//                          ((KEY block_node_or_indentless_sequence?)?
//                            *** *
//                          (VALUE block_node_or_indentless_sequence?)?)*
//
//                          BLOCK-END
//                          *********
//
func yamlParserParseBlockMappingKey(parser *yamlParserT, event *yamlEventT, first bool) bool {
	if first {
		token := peekToken(parser)
		parser.marks = append(parser.marks, token.startMark)
		skipToken(parser)
	}

	token := peekToken(parser)
	if token == nil {
		return false
	}

	if token.typ == yamlKeyToken {
		mark := token.endMark
		skipToken(parser)
		token = peekToken(parser)
		if token == nil {
			return false
		}
		if token.typ != yamlKeyToken &&
			token.typ != yamlValueToken &&
			token.typ != yamlBlockEndToken {
			parser.states = append(parser.states, yamlParseBlockMappingValueState)
			return yamlParserParseNode(parser, event, true, true)
		}
		parser.state = yamlParseBlockMappingValueState
		return yamlParserProcessEmptyScalar(parser, event, mark)
	} else if token.typ == yamlBlockEndToken {
		parser.state = parser.states[len(parser.states)-1]
		parser.states = parser.states[:len(parser.states)-1]
		parser.marks = parser.marks[:len(parser.marks)-1]
		*event = yamlEventT{
			typ:       yamlMappingEndEvent,
			startMark: token.startMark,
			endMark:   token.endMark,
		}
		skipToken(parser)
		return true
	}

	contextMark := parser.marks[len(parser.marks)-1]
	parser.marks = parser.marks[:len(parser.marks)-1]
	return yamlParserSetParserErrorContext(parser,
		"while parsing a block mapping", contextMark,
		"did not find expected key", token.startMark)
}

// Parse the productions:
// block_mapping        ::= BLOCK-MAPPING_START
//
//                          ((KEY block_node_or_indentless_sequence?)?
//
//                          (VALUE block_node_or_indentless_sequence?)?)*
//                           ***** *
//                          BLOCK-END
//
//
func yamlParserParseBlockMappingValue(parser *yamlParserT, event *yamlEventT) bool {
	token := peekToken(parser)
	if token == nil {
		return false
	}
	if token.typ == yamlValueToken {
		mark := token.endMark
		skipToken(parser)
		token = peekToken(parser)
		if token == nil {
			return false
		}
		if token.typ != yamlKeyToken &&
			token.typ != yamlValueToken &&
			token.typ != yamlBlockEndToken {
			parser.states = append(parser.states, yamlParseBlockMappingKeyState)
			return yamlParserParseNode(parser, event, true, true)
		}
		parser.state = yamlParseBlockMappingKeyState
		return yamlParserProcessEmptyScalar(parser, event, mark)
	}
	parser.state = yamlParseBlockMappingKeyState
	return yamlParserProcessEmptyScalar(parser, event, token.startMark)
}

// Parse the productions:
// flow_sequence        ::= FLOW-SEQUENCE-START
//                          *******************
//                          (flow_sequence_entry FLOW-ENTRY)*
//                           *                   **********
//                          flow_sequence_entry?
//                          *
//                          FLOW-SEQUENCE-END
//                          *****************
// flow_sequence_entry  ::= flow_node | KEY flow_node? (VALUE flow_node?)?
//                          *
//
func yamlParserParseFlowSequenceEntry(parser *yamlParserT, event *yamlEventT, first bool) bool {
	if first {
		token := peekToken(parser)
		parser.pendingSeqItemEvent = &yamlEventT{
			startMark: token.startMark,
			endMark:   token.endMark,
		}
		parser.marks = append(parser.marks, token.startMark)
		skipToken(parser)
	}
	token := peekToken(parser)
	if token == nil {
		return false
	}
	if token.typ != yamlFlowSequenceEndToken {
		if !first {
			if token.typ == yamlFlowEntryToken {
				parser.pendingSeqItemEvent = &yamlEventT{
					startMark: token.startMark,
					endMark:   token.endMark,
				}
				skipToken(parser)
				token = peekToken(parser)
				if token == nil {
					return false
				}
			} else {
				contextMark := parser.marks[len(parser.marks)-1]
				parser.marks = parser.marks[:len(parser.marks)-1]
				return yamlParserSetParserErrorContext(parser,
					"while parsing a flow sequence", contextMark,
					"did not find expected ',' or ']'", token.startMark)
			}
		}

		if token.typ == yamlKeyToken {
			parser.state = yamlParseFlowSequenceEntryMappingKeyState
			*event = yamlEventT{
				typ:       yamlMappingStartEvent,
				startMark: token.startMark,
				endMark:   token.endMark,
				implicit:  true,
				style:     yamlStyleT(yamlFlowMappingStyle),
			}
			skipToken(parser)
			return true
		} else if token.typ != yamlFlowSequenceEndToken {
			parser.states = append(parser.states, yamlParseFlowSequenceEntryState)
			return yamlParserParseNode(parser, event, false, false)
		}
	}

	parser.state = parser.states[len(parser.states)-1]
	parser.states = parser.states[:len(parser.states)-1]
	parser.marks = parser.marks[:len(parser.marks)-1]

	*event = yamlEventT{
		typ:       yamlSequenceEndEvent,
		startMark: token.startMark,
		endMark:   token.endMark,
	}

	skipToken(parser)
	return true
}

//
// Parse the productions:
// flow_sequence_entry  ::= flow_node | KEY flow_node? (VALUE flow_node?)?
//                                      *** *
//
func yamlParserParseFlowSequenceEntryMappingKey(parser *yamlParserT, event *yamlEventT) bool {
	token := peekToken(parser)
	if token == nil {
		return false
	}
	if token.typ != yamlValueToken &&
		token.typ != yamlFlowEntryToken &&
		token.typ != yamlFlowSequenceEndToken {
		parser.states = append(parser.states, yamlParseFlowSequenceEntryMappingValueState)
		return yamlParserParseNode(parser, event, false, false)
	}
	mark := token.endMark
	skipToken(parser)
	parser.state = yamlParseFlowSequenceEntryMappingValueState
	return yamlParserProcessEmptyScalar(parser, event, mark)
}

// Parse the productions:
// flow_sequence_entry  ::= flow_node | KEY flow_node? (VALUE flow_node?)?
//                                                      ***** *
//
func yamlParserParseFlowSequenceEntryMappingValue(parser *yamlParserT, event *yamlEventT) bool {
	token := peekToken(parser)
	if token == nil {
		return false
	}
	if token.typ == yamlValueToken {
		skipToken(parser)
		token := peekToken(parser)
		if token == nil {
			return false
		}
		if token.typ != yamlFlowEntryToken && token.typ != yamlFlowSequenceEndToken {
			parser.states = append(parser.states, yamlParseFlowSequenceEntryMappingEndState)
			return yamlParserParseNode(parser, event, false, false)
		}
	}
	parser.state = yamlParseFlowSequenceEntryMappingEndState
	return yamlParserProcessEmptyScalar(parser, event, token.startMark)
}

// Parse the productions:
// flow_sequence_entry  ::= flow_node | KEY flow_node? (VALUE flow_node?)?
//                                                                      *
//
func yamlParserParseFlowSequenceEntryMappingEnd(parser *yamlParserT, event *yamlEventT) bool {
	token := peekToken(parser)
	if token == nil {
		return false
	}
	parser.state = yamlParseFlowSequenceEntryState
	*event = yamlEventT{
		typ:       yamlMappingEndEvent,
		startMark: token.startMark,
		endMark:   token.startMark, // [Go] Shouldn't this be end_mark?
	}
	return true
}

// Parse the productions:
// flow_mapping         ::= FLOW-MAPPING-START
//                          ******************
//                          (flow_mapping_entry FLOW-ENTRY)*
//                           *                  **********
//                          flow_mapping_entry?
//                          ******************
//                          FLOW-MAPPING-END
//                          ****************
// flow_mapping_entry   ::= flow_node | KEY flow_node? (VALUE flow_node?)?
//                          *           *** *
//
func yamlParserParseFlowMappingKey(parser *yamlParserT, event *yamlEventT, first bool) bool {
	if first {
		token := peekToken(parser)
		parser.marks = append(parser.marks, token.startMark)
		skipToken(parser)
	}

	token := peekToken(parser)
	if token == nil {
		return false
	}

	if token.typ != yamlFlowMappingEndToken {
		if !first {
			if token.typ == yamlFlowEntryToken {
				skipToken(parser)
				token = peekToken(parser)
				if token == nil {
					return false
				}
			} else {
				contextMark := parser.marks[len(parser.marks)-1]
				parser.marks = parser.marks[:len(parser.marks)-1]
				return yamlParserSetParserErrorContext(parser,
					"while parsing a flow mapping", contextMark,
					"did not find expected ',' or '}'", token.startMark)
			}
		}

		if token.typ == yamlKeyToken {
			skipToken(parser)
			token = peekToken(parser)
			if token == nil {
				return false
			}
			if token.typ != yamlValueToken &&
				token.typ != yamlFlowEntryToken &&
				token.typ != yamlFlowMappingEndToken {
				parser.states = append(parser.states, yamlParseFlowMappingValueState)
				return yamlParserParseNode(parser, event, false, false)
			}
			parser.state = yamlParseFlowMappingValueState
			return yamlParserProcessEmptyScalar(parser, event, token.startMark)
		} else if token.typ != yamlFlowMappingEndToken {
			parser.states = append(parser.states, yamlParseFlowMappingEmptyValueState)
			return yamlParserParseNode(parser, event, false, false)
		}
	}

	parser.state = parser.states[len(parser.states)-1]
	parser.states = parser.states[:len(parser.states)-1]
	parser.marks = parser.marks[:len(parser.marks)-1]
	*event = yamlEventT{
		typ:       yamlMappingEndEvent,
		startMark: token.startMark,
		endMark:   token.endMark,
	}
	skipToken(parser)
	return true
}

// Parse the productions:
// flow_mapping_entry   ::= flow_node | KEY flow_node? (VALUE flow_node?)?
//                                   *                  ***** *
//
func yamlParserParseFlowMappingValue(parser *yamlParserT, event *yamlEventT, empty bool) bool {
	token := peekToken(parser)
	if token == nil {
		return false
	}
	if empty {
		parser.state = yamlParseFlowMappingKeyState
		return yamlParserProcessEmptyScalar(parser, event, token.startMark)
	}
	if token.typ == yamlValueToken {
		skipToken(parser)
		token = peekToken(parser)
		if token == nil {
			return false
		}
		if token.typ != yamlFlowEntryToken && token.typ != yamlFlowMappingEndToken {
			parser.states = append(parser.states, yamlParseFlowMappingKeyState)
			return yamlParserParseNode(parser, event, false, false)
		}
	}
	parser.state = yamlParseFlowMappingKeyState
	return yamlParserProcessEmptyScalar(parser, event, token.startMark)
}

// Generate an empty scalar event.
func yamlParserProcessEmptyScalar(parser *yamlParserT, event *yamlEventT, mark yamlMarkT) bool {
	*event = yamlEventT{
		typ:       yamlScalarEvent,
		startMark: mark,
		endMark:   mark,
		value:     nil, // Empty
		implicit:  true,
		style:     yamlStyleT(yamlPlainScalarStyle),
	}
	return true
}

var defaultTagDirectives = []yamlTagDirectiveT{
	{[]byte("!"), []byte("!")},
	{[]byte("!!"), []byte("tag:yaml.org,2002:")},
}

// Parse directives.
func yamlParserProcessDirectives(parser *yamlParserT,
	versionDirectiveRef **yamlVersionDirectiveT,
	tagDirectivesRef *[]yamlTagDirectiveT) bool {

	var versionDirective *yamlVersionDirectiveT
	var tagDirectives []yamlTagDirectiveT

	token := peekToken(parser)
	if token == nil {
		return false
	}

	for token.typ == yamlVersionDirectiveToken || token.typ == yamlTagDirectiveToken {
		if token.typ == yamlVersionDirectiveToken {
			if versionDirective != nil {
				yamlParserSetParserError(parser,
					"found duplicate %YAML directive", token.startMark)
				return false
			}
			if token.major != 1 || token.minor != 1 {
				yamlParserSetParserError(parser,
					"found incompatible YAML document", token.startMark)
				return false
			}
			versionDirective = &yamlVersionDirectiveT{
				major: token.major,
				minor: token.minor,
			}
		} else if token.typ == yamlTagDirectiveToken {
			value := yamlTagDirectiveT{
				handle: token.value,
				prefix: token.prefix,
			}
			if !yamlParserAppendTagDirective(parser, value, false, token.startMark) {
				return false
			}
			tagDirectives = append(tagDirectives, value)
		}

		skipToken(parser)
		token = peekToken(parser)
		if token == nil {
			return false
		}
	}

	for i := range defaultTagDirectives {
		if !yamlParserAppendTagDirective(parser, defaultTagDirectives[i], true, token.startMark) {
			return false
		}
	}

	if versionDirectiveRef != nil {
		*versionDirectiveRef = versionDirective
	}
	if tagDirectivesRef != nil {
		*tagDirectivesRef = tagDirectives
	}
	return true
}

// Append a tag directive to the directives stack.
func yamlParserAppendTagDirective(parser *yamlParserT, value yamlTagDirectiveT, allowDuplicates bool, mark yamlMarkT) bool {
	for i := range parser.tagDirectives {
		if bytes.Equal(value.handle, parser.tagDirectives[i].handle) {
			if allowDuplicates {
				return true
			}
			return yamlParserSetParserError(parser, "found duplicate %TAG directive", mark)
		}
	}

	// [Go] I suspect the copy is unnecessary. This was likely done
	// because there was no way to track ownership of the data.
	valueCopy := yamlTagDirectiveT{
		handle: make([]byte, len(value.handle)),
		prefix: make([]byte, len(value.prefix)),
	}
	copy(valueCopy.handle, value.handle)
	copy(valueCopy.prefix, value.prefix)
	parser.tagDirectives = append(parser.tagDirectives, valueCopy)
	return true
}
