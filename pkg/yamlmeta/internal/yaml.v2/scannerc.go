// Copyright 2020 VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package yaml

import (
	"bytes"
	"fmt"
)

// Introduction
// ************
//
// The following notes assume that you are familiar with the YAML specification
// (http://yaml.org/spec/1.2/spec.html).  We mostly follow it, although in
// some cases we are less restrictive that it requires.
//
// The process of transforming a YAML stream into a sequence of events is
// divided on two steps: Scanning and Parsing.
//
// The Scanner transforms the input stream into a sequence of tokens, while the
// parser transform the sequence of tokens produced by the Scanner into a
// sequence of parsing events.
//
// The Scanner is rather clever and complicated. The Parser, on the contrary,
// is a straightforward implementation of a recursive-descendant parser (or,
// LL(1) parser, as it is usually called).
//
// Actually there are two issues of Scanning that might be called "clever", the
// rest is quite straightforward.  The issues are "block collection start" and
// "simple keys".  Both issues are explained below in details.
//
// Here the Scanning step is explained and implemented.  We start with the list
// of all the tokens produced by the Scanner together with short descriptions.
//
// Now, tokens:
//
//      STREAM-START(encoding)          # The stream start.
//      STREAM-END                      # The stream end.
//      VERSION-DIRECTIVE(major,minor)  # The '%YAML' directive.
//      TAG-DIRECTIVE(handle,prefix)    # The '%TAG' directive.
//      DOCUMENT-START                  # '---'
//      DOCUMENT-END                    # '...'
//      BLOCK-SEQUENCE-START            # Indentation increase denoting a block
//      BLOCK-MAPPING-START             # sequence or a block mapping.
//      BLOCK-END                       # Indentation decrease.
//      FLOW-SEQUENCE-START             # '['
//      FLOW-SEQUENCE-END               # ']'
//      BLOCK-SEQUENCE-START            # '{'
//      BLOCK-SEQUENCE-END              # '}'
//      BLOCK-ENTRY                     # '-'
//      FLOW-ENTRY                      # ','
//      KEY                             # '?' or nothing (simple keys).
//      VALUE                           # ':'
//      ALIAS(anchor)                   # '*anchor'
//      ANCHOR(anchor)                  # '&anchor'
//      TAG(handle,suffix)              # '!handle!suffix'
//      SCALAR(value,style)             # A scalar.
//
// The following two tokens are "virtual" tokens denoting the beginning and the
// end of the stream:
//
//      STREAM-START(encoding)
//      STREAM-END
//
// We pass the information about the input stream encoding with the
// STREAM-START token.
//
// The next two tokens are responsible for tags:
//
//      VERSION-DIRECTIVE(major,minor)
//      TAG-DIRECTIVE(handle,prefix)
//
// Example:
//
//      %YAML   1.1
//      %TAG    !   !foo
//      %TAG    !yaml!  tag:yaml.org,2002:
//      ---
//
// The correspoding sequence of tokens:
//
//      STREAM-START(utf-8)
//      VERSION-DIRECTIVE(1,1)
//      TAG-DIRECTIVE("!","!foo")
//      TAG-DIRECTIVE("!yaml","tag:yaml.org,2002:")
//      DOCUMENT-START
//      STREAM-END
//
// Note that the VERSION-DIRECTIVE and TAG-DIRECTIVE tokens occupy a whole
// line.
//
// The document start and end indicators are represented by:
//
//      DOCUMENT-START
//      DOCUMENT-END
//
// Note that if a YAML stream contains an implicit document (without '---'
// and '...' indicators), no DOCUMENT-START and DOCUMENT-END tokens will be
// produced.
//
// In the following examples, we present whole documents together with the
// produced tokens.
//
//      1. An implicit document:
//
//          'a scalar'
//
//      Tokens:
//
//          STREAM-START(utf-8)
//          SCALAR("a scalar",single-quoted)
//          STREAM-END
//
//      2. An explicit document:
//
//          ---
//          'a scalar'
//          ...
//
//      Tokens:
//
//          STREAM-START(utf-8)
//          DOCUMENT-START
//          SCALAR("a scalar",single-quoted)
//          DOCUMENT-END
//          STREAM-END
//
//      3. Several documents in a stream:
//
//          'a scalar'
//          ---
//          'another scalar'
//          ---
//          'yet another scalar'
//
//      Tokens:
//
//          STREAM-START(utf-8)
//          SCALAR("a scalar",single-quoted)
//          DOCUMENT-START
//          SCALAR("another scalar",single-quoted)
//          DOCUMENT-START
//          SCALAR("yet another scalar",single-quoted)
//          STREAM-END
//
// We have already introduced the SCALAR token above.  The following tokens are
// used to describe aliases, anchors, tag, and scalars:
//
//      ALIAS(anchor)
//      ANCHOR(anchor)
//      TAG(handle,suffix)
//      SCALAR(value,style)
//
// The following series of examples illustrate the usage of these tokens:
//
//      1. A recursive sequence:
//
//          &A [ *A ]
//
//      Tokens:
//
//          STREAM-START(utf-8)
//          ANCHOR("A")
//          FLOW-SEQUENCE-START
//          ALIAS("A")
//          FLOW-SEQUENCE-END
//          STREAM-END
//
//      2. A tagged scalar:
//
//          !!float "3.14"  # A good approximation.
//
//      Tokens:
//
//          STREAM-START(utf-8)
//          TAG("!!","float")
//          SCALAR("3.14",double-quoted)
//          STREAM-END
//
//      3. Various scalar styles:
//
//          --- # Implicit empty plain scalars do not produce tokens.
//          --- a plain scalar
//          --- 'a single-quoted scalar'
//          --- "a double-quoted scalar"
//          --- |-
//            a literal scalar
//          --- >-
//            a folded
//            scalar
//
//      Tokens:
//
//          STREAM-START(utf-8)
//          DOCUMENT-START
//          DOCUMENT-START
//          SCALAR("a plain scalar",plain)
//          DOCUMENT-START
//          SCALAR("a single-quoted scalar",single-quoted)
//          DOCUMENT-START
//          SCALAR("a double-quoted scalar",double-quoted)
//          DOCUMENT-START
//          SCALAR("a literal scalar",literal)
//          DOCUMENT-START
//          SCALAR("a folded scalar",folded)
//          STREAM-END
//
// Now it's time to review collection-related tokens. We will start with
// flow collections:
//
//      FLOW-SEQUENCE-START
//      FLOW-SEQUENCE-END
//      FLOW-MAPPING-START
//      FLOW-MAPPING-END
//      FLOW-ENTRY
//      KEY
//      VALUE
//
// The tokens FLOW-SEQUENCE-START, FLOW-SEQUENCE-END, FLOW-MAPPING-START, and
// FLOW-MAPPING-END represent the indicators '[', ']', '{', and '}'
// correspondingly.  FLOW-ENTRY represent the ',' indicator.  Finally the
// indicators '?' and ':', which are used for denoting mapping keys and values,
// are represented by the KEY and VALUE tokens.
//
// The following examples show flow collections:
//
//      1. A flow sequence:
//
//          [item 1, item 2, item 3]
//
//      Tokens:
//
//          STREAM-START(utf-8)
//          FLOW-SEQUENCE-START
//          SCALAR("item 1",plain)
//          FLOW-ENTRY
//          SCALAR("item 2",plain)
//          FLOW-ENTRY
//          SCALAR("item 3",plain)
//          FLOW-SEQUENCE-END
//          STREAM-END
//
//      2. A flow mapping:
//
//          {
//              a simple key: a value,  # Note that the KEY token is produced.
//              ? a complex key: another value,
//          }
//
//      Tokens:
//
//          STREAM-START(utf-8)
//          FLOW-MAPPING-START
//          KEY
//          SCALAR("a simple key",plain)
//          VALUE
//          SCALAR("a value",plain)
//          FLOW-ENTRY
//          KEY
//          SCALAR("a complex key",plain)
//          VALUE
//          SCALAR("another value",plain)
//          FLOW-ENTRY
//          FLOW-MAPPING-END
//          STREAM-END
//
// A simple key is a key which is not denoted by the '?' indicator.  Note that
// the Scanner still produce the KEY token whenever it encounters a simple key.
//
// For scanning block collections, the following tokens are used (note that we
// repeat KEY and VALUE here):
//
//      BLOCK-SEQUENCE-START
//      BLOCK-MAPPING-START
//      BLOCK-END
//      BLOCK-ENTRY
//      KEY
//      VALUE
//
// The tokens BLOCK-SEQUENCE-START and BLOCK-MAPPING-START denote indentation
// increase that precedes a block collection (cf. the INDENT token in Python).
// The token BLOCK-END denote indentation decrease that ends a block collection
// (cf. the DEDENT token in Python).  However YAML has some syntax pecularities
// that makes detections of these tokens more complex.
//
// The tokens BLOCK-ENTRY, KEY, and VALUE are used to represent the indicators
// '-', '?', and ':' correspondingly.
//
// The following examples show how the tokens BLOCK-SEQUENCE-START,
// BLOCK-MAPPING-START, and BLOCK-END are emitted by the Scanner:
//
//      1. Block sequences:
//
//          - item 1
//          - item 2
//          -
//            - item 3.1
//            - item 3.2
//          -
//            key 1: value 1
//            key 2: value 2
//
//      Tokens:
//
//          STREAM-START(utf-8)
//          BLOCK-SEQUENCE-START
//          BLOCK-ENTRY
//          SCALAR("item 1",plain)
//          BLOCK-ENTRY
//          SCALAR("item 2",plain)
//          BLOCK-ENTRY
//          BLOCK-SEQUENCE-START
//          BLOCK-ENTRY
//          SCALAR("item 3.1",plain)
//          BLOCK-ENTRY
//          SCALAR("item 3.2",plain)
//          BLOCK-END
//          BLOCK-ENTRY
//          BLOCK-MAPPING-START
//          KEY
//          SCALAR("key 1",plain)
//          VALUE
//          SCALAR("value 1",plain)
//          KEY
//          SCALAR("key 2",plain)
//          VALUE
//          SCALAR("value 2",plain)
//          BLOCK-END
//          BLOCK-END
//          STREAM-END
//
//      2. Block mappings:
//
//          a simple key: a value   # The KEY token is produced here.
//          ? a complex key
//          : another value
//          a mapping:
//            key 1: value 1
//            key 2: value 2
//          a sequence:
//            - item 1
//            - item 2
//
//      Tokens:
//
//          STREAM-START(utf-8)
//          BLOCK-MAPPING-START
//          KEY
//          SCALAR("a simple key",plain)
//          VALUE
//          SCALAR("a value",plain)
//          KEY
//          SCALAR("a complex key",plain)
//          VALUE
//          SCALAR("another value",plain)
//          KEY
//          SCALAR("a mapping",plain)
//          BLOCK-MAPPING-START
//          KEY
//          SCALAR("key 1",plain)
//          VALUE
//          SCALAR("value 1",plain)
//          KEY
//          SCALAR("key 2",plain)
//          VALUE
//          SCALAR("value 2",plain)
//          BLOCK-END
//          KEY
//          SCALAR("a sequence",plain)
//          VALUE
//          BLOCK-SEQUENCE-START
//          BLOCK-ENTRY
//          SCALAR("item 1",plain)
//          BLOCK-ENTRY
//          SCALAR("item 2",plain)
//          BLOCK-END
//          BLOCK-END
//          STREAM-END
//
// YAML does not always require to start a new block collection from a new
// line.  If the current line contains only '-', '?', and ':' indicators, a new
// block collection may start at the current line.  The following examples
// illustrate this case:
//
//      1. Collections in a sequence:
//
//          - - item 1
//            - item 2
//          - key 1: value 1
//            key 2: value 2
//          - ? complex key
//            : complex value
//
//      Tokens:
//
//          STREAM-START(utf-8)
//          BLOCK-SEQUENCE-START
//          BLOCK-ENTRY
//          BLOCK-SEQUENCE-START
//          BLOCK-ENTRY
//          SCALAR("item 1",plain)
//          BLOCK-ENTRY
//          SCALAR("item 2",plain)
//          BLOCK-END
//          BLOCK-ENTRY
//          BLOCK-MAPPING-START
//          KEY
//          SCALAR("key 1",plain)
//          VALUE
//          SCALAR("value 1",plain)
//          KEY
//          SCALAR("key 2",plain)
//          VALUE
//          SCALAR("value 2",plain)
//          BLOCK-END
//          BLOCK-ENTRY
//          BLOCK-MAPPING-START
//          KEY
//          SCALAR("complex key")
//          VALUE
//          SCALAR("complex value")
//          BLOCK-END
//          BLOCK-END
//          STREAM-END
//
//      2. Collections in a mapping:
//
//          ? a sequence
//          : - item 1
//            - item 2
//          ? a mapping
//          : key 1: value 1
//            key 2: value 2
//
//      Tokens:
//
//          STREAM-START(utf-8)
//          BLOCK-MAPPING-START
//          KEY
//          SCALAR("a sequence",plain)
//          VALUE
//          BLOCK-SEQUENCE-START
//          BLOCK-ENTRY
//          SCALAR("item 1",plain)
//          BLOCK-ENTRY
//          SCALAR("item 2",plain)
//          BLOCK-END
//          KEY
//          SCALAR("a mapping",plain)
//          VALUE
//          BLOCK-MAPPING-START
//          KEY
//          SCALAR("key 1",plain)
//          VALUE
//          SCALAR("value 1",plain)
//          KEY
//          SCALAR("key 2",plain)
//          VALUE
//          SCALAR("value 2",plain)
//          BLOCK-END
//          BLOCK-END
//          STREAM-END
//
// YAML also permits non-indented sequences if they are included into a block
// mapping.  In this case, the token BLOCK-SEQUENCE-START is not produced:
//
//      key:
//      - item 1    # BLOCK-SEQUENCE-START is NOT produced here.
//      - item 2
//
// Tokens:
//
//      STREAM-START(utf-8)
//      BLOCK-MAPPING-START
//      KEY
//      SCALAR("key",plain)
//      VALUE
//      BLOCK-ENTRY
//      SCALAR("item 1",plain)
//      BLOCK-ENTRY
//      SCALAR("item 2",plain)
//      BLOCK-END
//

// Ensure that the buffer contains the required number of characters.
// Return true on success, false on failure (reader error or memory error).
func cache(parser *yamlParserT, length int) bool {
	// [Go] This was inlined: !cache(A, B) -> unread < B && !update(A, B)
	return parser.unread >= length || yamlParserUpdateBuffer(parser, length)
}

// Advance the buffer pointer.
func skip(parser *yamlParserT) {
	parser.mark.index++
	parser.mark.column++
	parser.unread--
	parser.bufferPos += width(parser.buffer[parser.bufferPos])
}

func skipLine(parser *yamlParserT) {
	if isCrlf(parser.buffer, parser.bufferPos) {
		parser.mark.index += 2
		parser.mark.column = 0
		parser.mark.line++
		parser.unread -= 2
		parser.bufferPos += 2
	} else if isBreak(parser.buffer, parser.bufferPos) {
		parser.mark.index++
		parser.mark.column = 0
		parser.mark.line++
		parser.unread--
		parser.bufferPos += width(parser.buffer[parser.bufferPos])
	}
}

// Copy a character to a string buffer and advance pointers.
func read(parser *yamlParserT, s []byte) []byte {
	w := width(parser.buffer[parser.bufferPos])
	if w == 0 {
		panic("invalid character sequence")
	}
	if len(s) == 0 {
		s = make([]byte, 0, 32)
	}
	if w == 1 && len(s)+w <= cap(s) {
		s = s[:len(s)+1]
		s[len(s)-1] = parser.buffer[parser.bufferPos]
		parser.bufferPos++
	} else {
		s = append(s, parser.buffer[parser.bufferPos:parser.bufferPos+w]...)
		parser.bufferPos += w
	}
	parser.mark.index++
	parser.mark.column++
	parser.unread--
	return s
}

// Copy a line break character to a string buffer and advance pointers.
func readLine(parser *yamlParserT, s []byte) []byte {
	buf := parser.buffer
	pos := parser.bufferPos
	switch {
	case buf[pos] == '\r' && buf[pos+1] == '\n':
		// CR LF . LF
		s = append(s, '\n')
		parser.bufferPos += 2
		parser.mark.index++
		parser.unread--
	case buf[pos] == '\r' || buf[pos] == '\n':
		// CR|LF . LF
		s = append(s, '\n')
		parser.bufferPos++
	case buf[pos] == '\xC2' && buf[pos+1] == '\x85':
		// NEL . LF
		s = append(s, '\n')
		parser.bufferPos += 2
	case buf[pos] == '\xE2' && buf[pos+1] == '\x80' && (buf[pos+2] == '\xA8' || buf[pos+2] == '\xA9'):
		// LS|PS . LS|PS
		s = append(s, buf[parser.bufferPos:pos+3]...)
		parser.bufferPos += 3
	default:
		return s
	}
	parser.mark.index++
	parser.mark.column = 0
	parser.mark.line++
	parser.unread--
	return s
}

// Get the next token.
func yamlParserScan(parser *yamlParserT, token *yamlTokenT) bool {
	// Erase the token object.
	*token = yamlTokenT{} // [Go] Is this necessary?

	// No tokens after STREAM-END or error.
	if parser.streamEndProduced || parser.error != yamlNoError {
		return true
	}

	// Ensure that the tokens queue contains enough tokens.
	if !parser.tokenAvailable {
		if !yamlParserFetchMoreTokens(parser) {
			return false
		}
	}

	// Fetch the next token from the queue.
	*token = parser.tokens[parser.tokensHead]
	parser.tokensHead++
	parser.tokensParsed++
	parser.tokenAvailable = false

	if token.typ == yamlStreamEndToken {
		parser.streamEndProduced = true
	}
	return true
}

// Set the scanner error and return false.
func yamlParserSetScannerError(parser *yamlParserT, context string, contextMark yamlMarkT, problem string) bool {
	parser.error = yamlScannerError
	parser.context = context
	parser.contextMark = contextMark
	parser.problem = problem
	parser.problemMark = parser.mark
	return false
}

func yamlParserSetScannerTagError(parser *yamlParserT, directive bool, contextMark yamlMarkT, problem string) bool {
	context := "while parsing a tag"
	if directive {
		context = "while parsing a %TAG directive"
	}
	return yamlParserSetScannerError(parser, context, contextMark, problem)
}

func trace(args ...interface{}) func() {
	pargs := append([]interface{}{"+++"}, args...)
	fmt.Println(pargs...)
	pargs = append([]interface{}{"---"}, args...)
	return func() { fmt.Println(pargs...) }
}

// Ensure that the tokens queue contains at least one token which can be
// returned to the Parser.
func yamlParserFetchMoreTokens(parser *yamlParserT) bool {
	// While we need more tokens to fetch, do it.
	for {
		// Check if we really need to fetch more tokens.
		needMoreTokens := false

		if parser.tokensHead == len(parser.tokens) {
			// Queue is empty.
			needMoreTokens = true
		} else {
			// Check if any potential simple key may occupy the head position.
			if !yamlParserStaleSimpleKeys(parser) {
				return false
			}

			for i := range parser.simpleKeys {
				simpleKey := &parser.simpleKeys[i]
				if simpleKey.possible && simpleKey.tokenNumber == parser.tokensParsed {
					needMoreTokens = true
					break
				}
			}
		}

		// We are finished.
		if !needMoreTokens {
			break
		}
		// Fetch the next token.
		if !yamlParserFetchNextToken(parser) {
			return false
		}
	}

	parser.tokenAvailable = true
	return true
}

// The dispatcher for token fetchers.
func yamlParserFetchNextToken(parser *yamlParserT) bool {
	// Ensure that the buffer is initialized.
	if parser.unread < 1 && !yamlParserUpdateBuffer(parser, 1) {
		return false
	}

	// Check if we just started scanning.  Fetch STREAM-START then.
	if !parser.streamStartProduced {
		return yamlParserFetchStreamStart(parser)
	}

	// Eat whitespaces and comments until we reach the next token.
	if !yamlParserScanToNextToken(parser) {
		return false
	}

	// Remove obsolete potential simple keys.
	if !yamlParserStaleSimpleKeys(parser) {
		return false
	}

	// Check the indentation level against the current column.
	if !yamlParserUnrollIndent(parser, parser.mark.column) {
		return false
	}

	// Ensure that the buffer contains at least 4 characters.  4 is the length
	// of the longest indicators ('--- ' and '... ').
	if parser.unread < 4 && !yamlParserUpdateBuffer(parser, 4) {
		return false
	}

	// Is it the end of the stream?
	if isZ(parser.buffer, parser.bufferPos) {
		return yamlParserFetchStreamEnd(parser)
	}

	// Is it a directive?
	if parser.mark.column == 0 && parser.buffer[parser.bufferPos] == '%' {
		return yamlParserFetchDirective(parser)
	}

	buf := parser.buffer
	pos := parser.bufferPos

	// Is it the document start indicator?
	if parser.mark.column == 0 && buf[pos] == '-' && buf[pos+1] == '-' && buf[pos+2] == '-' && isBlankz(buf, pos+3) {
		return yamlParserFetchDocumentIndicator(parser, yamlDocumentStartToken)
	}

	// Is it the document end indicator?
	if parser.mark.column == 0 && buf[pos] == '.' && buf[pos+1] == '.' && buf[pos+2] == '.' && isBlankz(buf, pos+3) {
		return yamlParserFetchDocumentIndicator(parser, yamlDocumentEndToken)
	}

	// Is it the flow sequence start indicator?
	if buf[pos] == '[' {
		return yamlParserFetchFlowCollectionStart(parser, yamlFlowSequenceStartToken)
	}

	// Is it the flow mapping start indicator?
	if parser.buffer[parser.bufferPos] == '{' {
		return yamlParserFetchFlowCollectionStart(parser, yamlFlowMappingStartToken)
	}

	// Is it the flow sequence end indicator?
	if parser.buffer[parser.bufferPos] == ']' {
		return yamlParserFetchFlowCollectionEnd(parser,
			yamlFlowSequenceEndToken)
	}

	// Is it the flow mapping end indicator?
	if parser.buffer[parser.bufferPos] == '}' {
		return yamlParserFetchFlowCollectionEnd(parser,
			yamlFlowMappingEndToken)
	}

	// Is it the flow entry indicator?
	if parser.buffer[parser.bufferPos] == ',' {
		return yamlParserFetchFlowEntry(parser)
	}

	// Is it the block entry indicator?
	if parser.buffer[parser.bufferPos] == '-' && isBlankz(parser.buffer, parser.bufferPos+1) {
		return yamlParserFetchBlockEntry(parser)
	}

	// Is it the key indicator?
	if parser.buffer[parser.bufferPos] == '?' && (parser.flowLevel > 0 || isBlankz(parser.buffer, parser.bufferPos+1)) {
		return yamlParserFetchKey(parser)
	}

	// Is it the value indicator?
	if parser.buffer[parser.bufferPos] == ':' && (parser.flowLevel > 0 || isBlankz(parser.buffer, parser.bufferPos+1)) {
		return yamlParserFetchValue(parser)
	}

	// Is it an alias?
	if parser.buffer[parser.bufferPos] == '*' {
		return yamlParserFetchAnchor(parser, yamlAliasToken)
	}

	// Is it an anchor?
	if parser.buffer[parser.bufferPos] == '&' {
		return yamlParserFetchAnchor(parser, yamlAnchorToken)
	}

	// Is it a tag?
	if parser.buffer[parser.bufferPos] == '!' {
		return yamlParserFetchTag(parser)
	}

	// Is it a literal scalar?
	if parser.buffer[parser.bufferPos] == '|' && parser.flowLevel == 0 {
		return yamlParserFetchBlockScalar(parser, true)
	}

	// Is it a folded scalar?
	if parser.buffer[parser.bufferPos] == '>' && parser.flowLevel == 0 {
		return yamlParserFetchBlockScalar(parser, false)
	}

	// Is it a single-quoted scalar?
	if parser.buffer[parser.bufferPos] == '\'' {
		return yamlParserFetchFlowScalar(parser, true)
	}

	// Is it a double-quoted scalar?
	if parser.buffer[parser.bufferPos] == '"' {
		return yamlParserFetchFlowScalar(parser, false)
	}

	// Is it a plain scalar?
	//
	// A plain scalar may start with any non-blank characters except
	//
	//      '-', '?', ':', ',', '[', ']', '{', '}',
	//      '#', '&', '*', '!', '|', '>', '\'', '\"',
	//      '%', '@', '`'.
	//
	// In the block context (and, for the '-' indicator, in the flow context
	// too), it may also start with the characters
	//
	//      '-', '?', ':'
	//
	// if it is followed by a non-space character.
	//
	// The last rule is more restrictive than the specification requires.
	// [Go] Make this logic more reasonable.
	//switch parser.buffer[parser.buffer_pos] {
	//case '-', '?', ':', ',', '?', '-', ',', ':', ']', '[', '}', '{', '&', '#', '!', '*', '>', '|', '"', '\'', '@', '%', '-', '`':
	//}
	if !(isBlankz(parser.buffer, parser.bufferPos) || parser.buffer[parser.bufferPos] == '-' ||
		parser.buffer[parser.bufferPos] == '?' || parser.buffer[parser.bufferPos] == ':' ||
		parser.buffer[parser.bufferPos] == ',' || parser.buffer[parser.bufferPos] == '[' ||
		parser.buffer[parser.bufferPos] == ']' || parser.buffer[parser.bufferPos] == '{' ||
		parser.buffer[parser.bufferPos] == '}' || parser.buffer[parser.bufferPos] == '#' ||
		parser.buffer[parser.bufferPos] == '&' || parser.buffer[parser.bufferPos] == '*' ||
		parser.buffer[parser.bufferPos] == '!' || parser.buffer[parser.bufferPos] == '|' ||
		parser.buffer[parser.bufferPos] == '>' || parser.buffer[parser.bufferPos] == '\'' ||
		parser.buffer[parser.bufferPos] == '"' || parser.buffer[parser.bufferPos] == '%' ||
		parser.buffer[parser.bufferPos] == '@' || parser.buffer[parser.bufferPos] == '`') ||
		(parser.buffer[parser.bufferPos] == '-' && !isBlank(parser.buffer, parser.bufferPos+1)) ||
		(parser.flowLevel == 0 &&
			(parser.buffer[parser.bufferPos] == '?' || parser.buffer[parser.bufferPos] == ':') &&
			!isBlankz(parser.buffer, parser.bufferPos+1)) {
		return yamlParserFetchPlainScalar(parser)
	}

	// If we don't determine the token type so far, it is an error.
	return yamlParserSetScannerError(parser,
		"while scanning for the next token", parser.mark,
		"found character that cannot start any token")
}

// Check the list of potential simple keys and remove the positions that
// cannot contain simple keys anymore.
func yamlParserStaleSimpleKeys(parser *yamlParserT) bool {
	// Check for a potential simple key for each flow level.
	for i := range parser.simpleKeys {
		simpleKey := &parser.simpleKeys[i]

		// The specification requires that a simple key
		//
		//  - is limited to a single line,
		//  - is shorter than 1024 characters.
		if simpleKey.possible && (simpleKey.mark.line < parser.mark.line || simpleKey.mark.index+1024 < parser.mark.index) {

			// Check if the potential simple key to be removed is required.
			if simpleKey.required {
				return yamlParserSetScannerError(parser,
					"while scanning a simple key", simpleKey.mark,
					"could not find expected ':'")
			}
			simpleKey.possible = false
		}
	}
	return true
}

// Check if a simple key may start at the current position and add it if
// needed.
func yamlParserSaveSimpleKey(parser *yamlParserT) bool {
	// A simple key is required at the current position if the scanner is in
	// the block context and the current column coincides with the indentation
	// level.

	required := parser.flowLevel == 0 && parser.indent == parser.mark.column

	//
	// If the current position may start a simple key, save it.
	//
	if parser.simpleKeyAllowed {
		simpleKey := yamlSimpleKeyT{
			possible:    true,
			required:    required,
			tokenNumber: parser.tokensParsed + (len(parser.tokens) - parser.tokensHead),
		}
		simpleKey.mark = parser.mark

		if !yamlParserRemoveSimpleKey(parser) {
			return false
		}
		parser.simpleKeys[len(parser.simpleKeys)-1] = simpleKey
	}
	return true
}

// Remove a potential simple key at the current flow level.
func yamlParserRemoveSimpleKey(parser *yamlParserT) bool {
	i := len(parser.simpleKeys) - 1
	if parser.simpleKeys[i].possible {
		// If the key is required, it is an error.
		if parser.simpleKeys[i].required {
			return yamlParserSetScannerError(parser,
				"while scanning a simple key", parser.simpleKeys[i].mark,
				"could not find expected ':'")
		}
	}
	// Remove the key from the stack.
	parser.simpleKeys[i].possible = false
	return true
}

// Increase the flow level and resize the simple key list if needed.
func yamlParserIncreaseFlowLevel(parser *yamlParserT) bool {
	// Reset the simple key on the next level.
	parser.simpleKeys = append(parser.simpleKeys, yamlSimpleKeyT{})

	// Increase the flow level.
	parser.flowLevel++
	return true
}

// Decrease the flow level.
func yamlParserDecreaseFlowLevel(parser *yamlParserT) bool {
	if parser.flowLevel > 0 {
		parser.flowLevel--
		parser.simpleKeys = parser.simpleKeys[:len(parser.simpleKeys)-1]
	}
	return true
}

// Push the current indentation level to the stack and set the new level
// the current column is greater than the indentation level.  In this case,
// append or insert the specified token into the token queue.
func yamlParserRollIndent(parser *yamlParserT, column, number int, typ yamlTokenTypeT, mark yamlMarkT) bool {
	// In the flow context, do nothing.
	if parser.flowLevel > 0 {
		return true
	}

	if parser.indent < column {
		// Push the current indentation level to the stack and set the new
		// indentation level.
		parser.indents = append(parser.indents, parser.indent)
		parser.indent = column

		// Create a token and insert it into the queue.
		token := yamlTokenT{
			typ:       typ,
			startMark: mark,
			endMark:   mark,
		}
		if number > -1 {
			number -= parser.tokensParsed
		}
		yamlInsertToken(parser, number, &token)
	}
	return true
}

// Pop indentation levels from the indents stack until the current level
// becomes less or equal to the column.  For each indentation level, append
// the BLOCK-END token.
func yamlParserUnrollIndent(parser *yamlParserT, column int) bool {
	// In the flow context, do nothing.
	if parser.flowLevel > 0 {
		return true
	}

	// Loop through the indentation levels in the stack.
	for parser.indent > column {
		// Create a token and append it to the queue.
		token := yamlTokenT{
			typ:       yamlBlockEndToken,
			startMark: parser.mark,
			endMark:   parser.mark,
		}
		yamlInsertToken(parser, -1, &token)

		// Pop the indentation level.
		parser.indent = parser.indents[len(parser.indents)-1]
		parser.indents = parser.indents[:len(parser.indents)-1]
	}
	return true
}

// Initialize the scanner and produce the STREAM-START token.
func yamlParserFetchStreamStart(parser *yamlParserT) bool {

	// Set the initial indentation.
	parser.indent = -1

	// Initialize the simple key stack.
	parser.simpleKeys = append(parser.simpleKeys, yamlSimpleKeyT{})

	// A simple key is allowed at the beginning of the stream.
	parser.simpleKeyAllowed = true

	// We have started.
	parser.streamStartProduced = true

	// Create the STREAM-START token and append it to the queue.
	token := yamlTokenT{
		typ:       yamlStreamStartToken,
		startMark: parser.mark,
		endMark:   parser.mark,
		encoding:  parser.encoding,
	}
	yamlInsertToken(parser, -1, &token)
	return true
}

// Produce the STREAM-END token and shut down the scanner.
func yamlParserFetchStreamEnd(parser *yamlParserT) bool {

	// Force new line.
	if parser.mark.column != 0 {
		parser.mark.column = 0
		parser.mark.line++
	}

	// Reset the indentation level.
	if !yamlParserUnrollIndent(parser, -1) {
		return false
	}

	// Reset simple keys.
	if !yamlParserRemoveSimpleKey(parser) {
		return false
	}

	parser.simpleKeyAllowed = false

	// Create the STREAM-END token and append it to the queue.
	token := yamlTokenT{
		typ:       yamlStreamEndToken,
		startMark: parser.mark,
		endMark:   parser.mark,
	}
	yamlInsertToken(parser, -1, &token)
	return true
}

// Produce a VERSION-DIRECTIVE or TAG-DIRECTIVE token.
func yamlParserFetchDirective(parser *yamlParserT) bool {
	// Reset the indentation level.
	if !yamlParserUnrollIndent(parser, -1) {
		return false
	}

	// Reset simple keys.
	if !yamlParserRemoveSimpleKey(parser) {
		return false
	}

	parser.simpleKeyAllowed = false

	// Create the YAML-DIRECTIVE or TAG-DIRECTIVE token.
	token := yamlTokenT{}
	if !yamlParserScanDirective(parser, &token) {
		return false
	}
	// Append the token to the queue.
	yamlInsertToken(parser, -1, &token)
	return true
}

// Produce the DOCUMENT-START or DOCUMENT-END token.
func yamlParserFetchDocumentIndicator(parser *yamlParserT, typ yamlTokenTypeT) bool {
	// Reset the indentation level.
	if !yamlParserUnrollIndent(parser, -1) {
		return false
	}

	// Reset simple keys.
	if !yamlParserRemoveSimpleKey(parser) {
		return false
	}

	parser.simpleKeyAllowed = false

	// Consume the token.
	startMark := parser.mark

	skip(parser)
	skip(parser)
	skip(parser)

	endMark := parser.mark

	// Create the DOCUMENT-START or DOCUMENT-END token.
	token := yamlTokenT{
		typ:       typ,
		startMark: startMark,
		endMark:   endMark,
	}
	// Append the token to the queue.
	yamlInsertToken(parser, -1, &token)
	return true
}

// Produce the FLOW-SEQUENCE-START or FLOW-MAPPING-START token.
func yamlParserFetchFlowCollectionStart(parser *yamlParserT, typ yamlTokenTypeT) bool {
	// The indicators '[' and '{' may start a simple key.
	if !yamlParserSaveSimpleKey(parser) {
		return false
	}

	// Increase the flow level.
	if !yamlParserIncreaseFlowLevel(parser) {
		return false
	}

	// A simple key may follow the indicators '[' and '{'.
	parser.simpleKeyAllowed = true

	// Consume the token.
	startMark := parser.mark
	skip(parser)
	endMark := parser.mark

	// Create the FLOW-SEQUENCE-START of FLOW-MAPPING-START token.
	token := yamlTokenT{
		typ:       typ,
		startMark: startMark,
		endMark:   endMark,
	}
	// Append the token to the queue.
	yamlInsertToken(parser, -1, &token)
	return true
}

// Produce the FLOW-SEQUENCE-END or FLOW-MAPPING-END token.
func yamlParserFetchFlowCollectionEnd(parser *yamlParserT, typ yamlTokenTypeT) bool {
	// Reset any potential simple key on the current flow level.
	if !yamlParserRemoveSimpleKey(parser) {
		return false
	}

	// Decrease the flow level.
	if !yamlParserDecreaseFlowLevel(parser) {
		return false
	}

	// No simple keys after the indicators ']' and '}'.
	parser.simpleKeyAllowed = false

	// Consume the token.

	startMark := parser.mark
	skip(parser)
	endMark := parser.mark

	// Create the FLOW-SEQUENCE-END of FLOW-MAPPING-END token.
	token := yamlTokenT{
		typ:       typ,
		startMark: startMark,
		endMark:   endMark,
	}
	// Append the token to the queue.
	yamlInsertToken(parser, -1, &token)
	return true
}

// Produce the FLOW-ENTRY token.
func yamlParserFetchFlowEntry(parser *yamlParserT) bool {
	// Reset any potential simple keys on the current flow level.
	if !yamlParserRemoveSimpleKey(parser) {
		return false
	}

	// Simple keys are allowed after ','.
	parser.simpleKeyAllowed = true

	// Consume the token.
	startMark := parser.mark
	skip(parser)
	endMark := parser.mark

	// Create the FLOW-ENTRY token and append it to the queue.
	token := yamlTokenT{
		typ:       yamlFlowEntryToken,
		startMark: startMark,
		endMark:   endMark,
	}
	yamlInsertToken(parser, -1, &token)
	return true
}

// Produce the BLOCK-ENTRY token.
func yamlParserFetchBlockEntry(parser *yamlParserT) bool {
	// Check if the scanner is in the block context.
	if parser.flowLevel == 0 {
		// Check if we are allowed to start a new entry.
		if !parser.simpleKeyAllowed {
			return yamlParserSetScannerError(parser, "", parser.mark,
				"block sequence entries are not allowed in this context")
		}
		// Add the BLOCK-SEQUENCE-START token if needed.
		if !yamlParserRollIndent(parser, parser.mark.column, -1, yamlBlockSequenceStartToken, parser.mark) {
			return false
		}
	} else {
		// It is an error for the '-' indicator to occur in the flow context,
		// but we let the Parser detect and report about it because the Parser
		// is able to point to the context.
	}

	// Reset any potential simple keys on the current flow level.
	if !yamlParserRemoveSimpleKey(parser) {
		return false
	}

	// Simple keys are allowed after '-'.
	parser.simpleKeyAllowed = true

	// Consume the token.
	startMark := parser.mark
	skip(parser)
	endMark := parser.mark

	// Create the BLOCK-ENTRY token and append it to the queue.
	token := yamlTokenT{
		typ:       yamlBlockEntryToken,
		startMark: startMark,
		endMark:   endMark,
	}
	yamlInsertToken(parser, -1, &token)
	return true
}

// Produce the KEY token.
func yamlParserFetchKey(parser *yamlParserT) bool {

	// In the block context, additional checks are required.
	if parser.flowLevel == 0 {
		// Check if we are allowed to start a new key (not nessesary simple).
		if !parser.simpleKeyAllowed {
			return yamlParserSetScannerError(parser, "", parser.mark,
				"mapping keys are not allowed in this context")
		}
		// Add the BLOCK-MAPPING-START token if needed.
		if !yamlParserRollIndent(parser, parser.mark.column, -1, yamlBlockMappingStartToken, parser.mark) {
			return false
		}
	}

	// Reset any potential simple keys on the current flow level.
	if !yamlParserRemoveSimpleKey(parser) {
		return false
	}

	// Simple keys are allowed after '?' in the block context.
	parser.simpleKeyAllowed = parser.flowLevel == 0

	// Consume the token.
	startMark := parser.mark
	skip(parser)
	endMark := parser.mark

	// Create the KEY token and append it to the queue.
	token := yamlTokenT{
		typ:       yamlKeyToken,
		startMark: startMark,
		endMark:   endMark,
	}
	yamlInsertToken(parser, -1, &token)
	return true
}

// Produce the VALUE token.
func yamlParserFetchValue(parser *yamlParserT) bool {

	simpleKey := &parser.simpleKeys[len(parser.simpleKeys)-1]

	// Have we found a simple key?
	if simpleKey.possible {
		// Create the KEY token and insert it into the queue.
		token := yamlTokenT{
			typ:       yamlKeyToken,
			startMark: simpleKey.mark,
			endMark:   simpleKey.mark,
		}
		yamlInsertToken(parser, simpleKey.tokenNumber-parser.tokensParsed, &token)

		// In the block context, we may need to add the BLOCK-MAPPING-START token.
		if !yamlParserRollIndent(parser, simpleKey.mark.column,
			simpleKey.tokenNumber,
			yamlBlockMappingStartToken, simpleKey.mark) {
			return false
		}

		// Remove the simple key.
		simpleKey.possible = false

		// A simple key cannot follow another simple key.
		parser.simpleKeyAllowed = false

	} else {
		// The ':' indicator follows a complex key.

		// In the block context, extra checks are required.
		if parser.flowLevel == 0 {

			// Check if we are allowed to start a complex value.
			if !parser.simpleKeyAllowed {
				return yamlParserSetScannerError(parser, "", parser.mark,
					"mapping values are not allowed in this context")
			}

			// Add the BLOCK-MAPPING-START token if needed.
			if !yamlParserRollIndent(parser, parser.mark.column, -1, yamlBlockMappingStartToken, parser.mark) {
				return false
			}
		}

		// Simple keys after ':' are allowed in the block context.
		parser.simpleKeyAllowed = parser.flowLevel == 0
	}

	// Consume the token.
	startMark := parser.mark
	skip(parser)
	endMark := parser.mark

	// Create the VALUE token and append it to the queue.
	token := yamlTokenT{
		typ:       yamlValueToken,
		startMark: startMark,
		endMark:   endMark,
	}
	yamlInsertToken(parser, -1, &token)
	return true
}

// Produce the ALIAS or ANCHOR token.
func yamlParserFetchAnchor(parser *yamlParserT, typ yamlTokenTypeT) bool {
	// An anchor or an alias could be a simple key.
	if !yamlParserSaveSimpleKey(parser) {
		return false
	}

	// A simple key cannot follow an anchor or an alias.
	parser.simpleKeyAllowed = false

	// Create the ALIAS or ANCHOR token and append it to the queue.
	var token yamlTokenT
	if !yamlParserScanAnchor(parser, &token, typ) {
		return false
	}
	yamlInsertToken(parser, -1, &token)
	return true
}

// Produce the TAG token.
func yamlParserFetchTag(parser *yamlParserT) bool {
	// A tag could be a simple key.
	if !yamlParserSaveSimpleKey(parser) {
		return false
	}

	// A simple key cannot follow a tag.
	parser.simpleKeyAllowed = false

	// Create the TAG token and append it to the queue.
	var token yamlTokenT
	if !yamlParserScanTag(parser, &token) {
		return false
	}
	yamlInsertToken(parser, -1, &token)
	return true
}

// Produce the SCALAR(...,literal) or SCALAR(...,folded) tokens.
func yamlParserFetchBlockScalar(parser *yamlParserT, literal bool) bool {
	// Remove any potential simple keys.
	if !yamlParserRemoveSimpleKey(parser) {
		return false
	}

	// A simple key may follow a block scalar.
	parser.simpleKeyAllowed = true

	// Create the SCALAR token and append it to the queue.
	var token yamlTokenT
	if !yamlParserScanBlockScalar(parser, &token, literal) {
		return false
	}
	yamlInsertToken(parser, -1, &token)
	return true
}

// Produce the SCALAR(...,single-quoted) or SCALAR(...,double-quoted) tokens.
func yamlParserFetchFlowScalar(parser *yamlParserT, single bool) bool {
	// A plain scalar could be a simple key.
	if !yamlParserSaveSimpleKey(parser) {
		return false
	}

	// A simple key cannot follow a flow scalar.
	parser.simpleKeyAllowed = false

	// Create the SCALAR token and append it to the queue.
	var token yamlTokenT
	if !yamlParserScanFlowScalar(parser, &token, single) {
		return false
	}
	yamlInsertToken(parser, -1, &token)
	return true
}

// Produce the SCALAR(...,plain) token.
func yamlParserFetchPlainScalar(parser *yamlParserT) bool {
	// A plain scalar could be a simple key.
	if !yamlParserSaveSimpleKey(parser) {
		return false
	}

	// A simple key cannot follow a flow scalar.
	parser.simpleKeyAllowed = false

	// Create the SCALAR token and append it to the queue.
	var token yamlTokenT
	if !yamlParserScanPlainScalar(parser, &token) {
		return false
	}
	yamlInsertToken(parser, -1, &token)
	return true
}

func yamlParserFetchComment(parser *yamlParserT) bool {
	startMark := parser.mark
	var comment []byte

	for !isBreakz(parser.buffer, parser.bufferPos) {
		comment = read(parser, comment)
		if parser.unread < 1 && !yamlParserUpdateBuffer(parser, 1) {
			break
		}
	}

	parser.comments = append(parser.comments, yamlTokenT{
		typ:       yamlCommentToken,
		startMark: startMark,
		endMark:   parser.mark,
		value:     comment[1:], // skip #
		style:     yamlPlainScalarStyle,
	})

	return true
}

// Eat whitespaces and comments until the next token is found.
func yamlParserScanToNextToken(parser *yamlParserT) bool {

	// Until the next token is not found.
	for {
		// Allow the BOM mark to start a line.
		if parser.unread < 1 && !yamlParserUpdateBuffer(parser, 1) {
			return false
		}
		if parser.mark.column == 0 && isBom(parser.buffer, parser.bufferPos) {
			skip(parser)
		}

		// Eat whitespaces.
		// Tabs are allowed:
		//  - in the flow context
		//  - in the block context, but not at the beginning of the line or
		//  after '-', '?', or ':' (complex value).
		if parser.unread < 1 && !yamlParserUpdateBuffer(parser, 1) {
			return false
		}

		for parser.buffer[parser.bufferPos] == ' ' || ((parser.flowLevel > 0 || !parser.simpleKeyAllowed) && parser.buffer[parser.bufferPos] == '\t') {
			skip(parser)
			if parser.unread < 1 && !yamlParserUpdateBuffer(parser, 1) {
				return false
			}
		}

		if parser.buffer[parser.bufferPos] == '#' {
			if !yamlParserFetchComment(parser) {
				return false
			}
		}

		// If it is a line break, eat it.
		if isBreak(parser.buffer, parser.bufferPos) {
			if parser.unread < 2 && !yamlParserUpdateBuffer(parser, 2) {
				return false
			}
			skipLine(parser)

			// In the block context, a new line may start a simple key.
			if parser.flowLevel == 0 {
				parser.simpleKeyAllowed = true
			}
		} else {
			break // We have found a token.
		}
	}

	return true
}

// Scan a YAML-DIRECTIVE or TAG-DIRECTIVE token.
//
// Scope:
//      %YAML    1.1    # a comment \n
//      ^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^
//      %TAG    !yaml!  tag:yaml.org,2002:  \n
//      ^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^
//
func yamlParserScanDirective(parser *yamlParserT, token *yamlTokenT) bool {
	// Eat '%'.
	startMark := parser.mark
	skip(parser)

	// Scan the directive name.
	var name []byte
	if !yamlParserScanDirectiveName(parser, startMark, &name) {
		return false
	}

	// Is it a YAML directive?
	if bytes.Equal(name, []byte("YAML")) {
		// Scan the VERSION directive value.
		var major, minor int8
		if !yamlParserScanVersionDirectiveValue(parser, startMark, &major, &minor) {
			return false
		}
		endMark := parser.mark

		// Create a VERSION-DIRECTIVE token.
		*token = yamlTokenT{
			typ:       yamlVersionDirectiveToken,
			startMark: startMark,
			endMark:   endMark,
			major:     major,
			minor:     minor,
		}

		// Is it a TAG directive?
	} else if bytes.Equal(name, []byte("TAG")) {
		// Scan the TAG directive value.
		var handle, prefix []byte
		if !yamlParserScanTagDirectiveValue(parser, startMark, &handle, &prefix) {
			return false
		}
		endMark := parser.mark

		// Create a TAG-DIRECTIVE token.
		*token = yamlTokenT{
			typ:       yamlTagDirectiveToken,
			startMark: startMark,
			endMark:   endMark,
			value:     handle,
			prefix:    prefix,
		}

		// Unknown directive.
	} else {
		yamlParserSetScannerError(parser, "while scanning a directive",
			startMark, "found unknown directive name")
		return false
	}

	// Eat the rest of the line including any comments.
	if parser.unread < 1 && !yamlParserUpdateBuffer(parser, 1) {
		return false
	}

	for isBlank(parser.buffer, parser.bufferPos) {
		skip(parser)
		if parser.unread < 1 && !yamlParserUpdateBuffer(parser, 1) {
			return false
		}
	}

	if parser.buffer[parser.bufferPos] == '#' {
		for !isBreakz(parser.buffer, parser.bufferPos) {
			skip(parser)
			if parser.unread < 1 && !yamlParserUpdateBuffer(parser, 1) {
				return false
			}
		}
	}

	// Check if we are at the end of the line.
	if !isBreakz(parser.buffer, parser.bufferPos) {
		yamlParserSetScannerError(parser, "while scanning a directive",
			startMark, "did not find expected comment or line break")
		return false
	}

	// Eat a line break.
	if isBreak(parser.buffer, parser.bufferPos) {
		if parser.unread < 2 && !yamlParserUpdateBuffer(parser, 2) {
			return false
		}
		skipLine(parser)
	}

	return true
}

// Scan the directive name.
//
// Scope:
//      %YAML   1.1     # a comment \n
//       ^^^^
//      %TAG    !yaml!  tag:yaml.org,2002:  \n
//       ^^^
//
func yamlParserScanDirectiveName(parser *yamlParserT, startMark yamlMarkT, name *[]byte) bool {
	// Consume the directive name.
	if parser.unread < 1 && !yamlParserUpdateBuffer(parser, 1) {
		return false
	}

	var s []byte
	for isAlpha(parser.buffer, parser.bufferPos) {
		s = read(parser, s)
		if parser.unread < 1 && !yamlParserUpdateBuffer(parser, 1) {
			return false
		}
	}

	// Check if the name is empty.
	if len(s) == 0 {
		yamlParserSetScannerError(parser, "while scanning a directive",
			startMark, "could not find expected directive name")
		return false
	}

	// Check for an blank character after the name.
	if !isBlankz(parser.buffer, parser.bufferPos) {
		yamlParserSetScannerError(parser, "while scanning a directive",
			startMark, "found unexpected non-alphabetical character")
		return false
	}
	*name = s
	return true
}

// Scan the value of VERSION-DIRECTIVE.
//
// Scope:
//      %YAML   1.1     # a comment \n
//           ^^^^^^
func yamlParserScanVersionDirectiveValue(parser *yamlParserT, startMark yamlMarkT, major, minor *int8) bool {
	// Eat whitespaces.
	if parser.unread < 1 && !yamlParserUpdateBuffer(parser, 1) {
		return false
	}
	for isBlank(parser.buffer, parser.bufferPos) {
		skip(parser)
		if parser.unread < 1 && !yamlParserUpdateBuffer(parser, 1) {
			return false
		}
	}

	// Consume the major version number.
	if !yamlParserScanVersionDirectiveNumber(parser, startMark, major) {
		return false
	}

	// Eat '.'.
	if parser.buffer[parser.bufferPos] != '.' {
		return yamlParserSetScannerError(parser, "while scanning a %YAML directive",
			startMark, "did not find expected digit or '.' character")
	}

	skip(parser)

	// Consume the minor version number.
	if !yamlParserScanVersionDirectiveNumber(parser, startMark, minor) {
		return false
	}
	return true
}

const maxNumberLength = 2

// Scan the version number of VERSION-DIRECTIVE.
//
// Scope:
//      %YAML   1.1     # a comment \n
//              ^
//      %YAML   1.1     # a comment \n
//                ^
func yamlParserScanVersionDirectiveNumber(parser *yamlParserT, startMark yamlMarkT, number *int8) bool {

	// Repeat while the next character is digit.
	if parser.unread < 1 && !yamlParserUpdateBuffer(parser, 1) {
		return false
	}
	var value, length int8
	for isDigit(parser.buffer, parser.bufferPos) {
		// Check if the number is too long.
		length++
		if length > maxNumberLength {
			return yamlParserSetScannerError(parser, "while scanning a %YAML directive",
				startMark, "found extremely long version number")
		}
		value = value*10 + int8(asDigit(parser.buffer, parser.bufferPos))
		skip(parser)
		if parser.unread < 1 && !yamlParserUpdateBuffer(parser, 1) {
			return false
		}
	}

	// Check if the number was present.
	if length == 0 {
		return yamlParserSetScannerError(parser, "while scanning a %YAML directive",
			startMark, "did not find expected version number")
	}
	*number = value
	return true
}

// Scan the value of a TAG-DIRECTIVE token.
//
// Scope:
//      %TAG    !yaml!  tag:yaml.org,2002:  \n
//          ^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^
//
func yamlParserScanTagDirectiveValue(parser *yamlParserT, startMark yamlMarkT, handle, prefix *[]byte) bool {
	var handleValue, prefixValue []byte

	// Eat whitespaces.
	if parser.unread < 1 && !yamlParserUpdateBuffer(parser, 1) {
		return false
	}

	for isBlank(parser.buffer, parser.bufferPos) {
		skip(parser)
		if parser.unread < 1 && !yamlParserUpdateBuffer(parser, 1) {
			return false
		}
	}

	// Scan a handle.
	if !yamlParserScanTagHandle(parser, true, startMark, &handleValue) {
		return false
	}

	// Expect a whitespace.
	if parser.unread < 1 && !yamlParserUpdateBuffer(parser, 1) {
		return false
	}
	if !isBlank(parser.buffer, parser.bufferPos) {
		yamlParserSetScannerError(parser, "while scanning a %TAG directive",
			startMark, "did not find expected whitespace")
		return false
	}

	// Eat whitespaces.
	for isBlank(parser.buffer, parser.bufferPos) {
		skip(parser)
		if parser.unread < 1 && !yamlParserUpdateBuffer(parser, 1) {
			return false
		}
	}

	// Scan a prefix.
	if !yamlParserScanTagURI(parser, true, nil, startMark, &prefixValue) {
		return false
	}

	// Expect a whitespace or line break.
	if parser.unread < 1 && !yamlParserUpdateBuffer(parser, 1) {
		return false
	}
	if !isBlankz(parser.buffer, parser.bufferPos) {
		yamlParserSetScannerError(parser, "while scanning a %TAG directive",
			startMark, "did not find expected whitespace or line break")
		return false
	}

	*handle = handleValue
	*prefix = prefixValue
	return true
}

func yamlParserScanAnchor(parser *yamlParserT, token *yamlTokenT, typ yamlTokenTypeT) bool {
	var s []byte

	// Eat the indicator character.
	startMark := parser.mark
	skip(parser)

	// Consume the value.
	if parser.unread < 1 && !yamlParserUpdateBuffer(parser, 1) {
		return false
	}

	for isAlpha(parser.buffer, parser.bufferPos) {
		s = read(parser, s)
		if parser.unread < 1 && !yamlParserUpdateBuffer(parser, 1) {
			return false
		}
	}

	endMark := parser.mark

	/*
	 * Check if length of the anchor is greater than 0 and it is followed by
	 * a whitespace character or one of the indicators:
	 *
	 *      '?', ':', ',', ']', '}', '%', '@', '`'.
	 */

	if len(s) == 0 ||
		!(isBlankz(parser.buffer, parser.bufferPos) || parser.buffer[parser.bufferPos] == '?' ||
			parser.buffer[parser.bufferPos] == ':' || parser.buffer[parser.bufferPos] == ',' ||
			parser.buffer[parser.bufferPos] == ']' || parser.buffer[parser.bufferPos] == '}' ||
			parser.buffer[parser.bufferPos] == '%' || parser.buffer[parser.bufferPos] == '@' ||
			parser.buffer[parser.bufferPos] == '`') {
		context := "while scanning an alias"
		if typ == yamlAnchorToken {
			context = "while scanning an anchor"
		}
		yamlParserSetScannerError(parser, context, startMark,
			"did not find expected alphabetic or numeric character")
		return false
	}

	// Create a token.
	*token = yamlTokenT{
		typ:       typ,
		startMark: startMark,
		endMark:   endMark,
		value:     s,
	}

	return true
}

/*
 * Scan a TAG token.
 */

func yamlParserScanTag(parser *yamlParserT, token *yamlTokenT) bool {
	var handle, suffix []byte

	startMark := parser.mark

	// Check if the tag is in the canonical form.
	if parser.unread < 2 && !yamlParserUpdateBuffer(parser, 2) {
		return false
	}

	if parser.buffer[parser.bufferPos+1] == '<' {
		// Keep the handle as ''

		// Eat '!<'
		skip(parser)
		skip(parser)

		// Consume the tag value.
		if !yamlParserScanTagURI(parser, false, nil, startMark, &suffix) {
			return false
		}

		// Check for '>' and eat it.
		if parser.buffer[parser.bufferPos] != '>' {
			yamlParserSetScannerError(parser, "while scanning a tag",
				startMark, "did not find the expected '>'")
			return false
		}

		skip(parser)
	} else {
		// The tag has either the '!suffix' or the '!handle!suffix' form.

		// First, try to scan a handle.
		if !yamlParserScanTagHandle(parser, false, startMark, &handle) {
			return false
		}

		// Check if it is, indeed, handle.
		if handle[0] == '!' && len(handle) > 1 && handle[len(handle)-1] == '!' {
			// Scan the suffix now.
			if !yamlParserScanTagURI(parser, false, nil, startMark, &suffix) {
				return false
			}
		} else {
			// It wasn't a handle after all.  Scan the rest of the tag.
			if !yamlParserScanTagURI(parser, false, handle, startMark, &suffix) {
				return false
			}

			// Set the handle to '!'.
			handle = []byte{'!'}

			// A special case: the '!' tag.  Set the handle to '' and the
			// suffix to '!'.
			if len(suffix) == 0 {
				handle, suffix = suffix, handle
			}
		}
	}

	// Check the character which ends the tag.
	if parser.unread < 1 && !yamlParserUpdateBuffer(parser, 1) {
		return false
	}
	if !isBlankz(parser.buffer, parser.bufferPos) {
		yamlParserSetScannerError(parser, "while scanning a tag",
			startMark, "did not find expected whitespace or line break")
		return false
	}

	endMark := parser.mark

	// Create a token.
	*token = yamlTokenT{
		typ:       yamlTagToken,
		startMark: startMark,
		endMark:   endMark,
		value:     handle,
		suffix:    suffix,
	}
	return true
}

// Scan a tag handle.
func yamlParserScanTagHandle(parser *yamlParserT, directive bool, startMark yamlMarkT, handle *[]byte) bool {
	// Check the initial '!' character.
	if parser.unread < 1 && !yamlParserUpdateBuffer(parser, 1) {
		return false
	}
	if parser.buffer[parser.bufferPos] != '!' {
		yamlParserSetScannerTagError(parser, directive,
			startMark, "did not find expected '!'")
		return false
	}

	var s []byte

	// Copy the '!' character.
	s = read(parser, s)

	// Copy all subsequent alphabetical and numerical characters.
	if parser.unread < 1 && !yamlParserUpdateBuffer(parser, 1) {
		return false
	}
	for isAlpha(parser.buffer, parser.bufferPos) {
		s = read(parser, s)
		if parser.unread < 1 && !yamlParserUpdateBuffer(parser, 1) {
			return false
		}
	}

	// Check if the trailing character is '!' and copy it.
	if parser.buffer[parser.bufferPos] == '!' {
		s = read(parser, s)
	} else {
		// It's either the '!' tag or not really a tag handle.  If it's a %TAG
		// directive, it's an error.  If it's a tag token, it must be a part of URI.
		if directive && string(s) != "!" {
			yamlParserSetScannerTagError(parser, directive,
				startMark, "did not find expected '!'")
			return false
		}
	}

	*handle = s
	return true
}

// Scan a tag.
func yamlParserScanTagURI(parser *yamlParserT, directive bool, head []byte, startMark yamlMarkT, uri *[]byte) bool {
	//size_t length = head ? strlen((char *)head) : 0
	var s []byte
	hasTag := len(head) > 0

	// Copy the head if needed.
	//
	// Note that we don't copy the leading '!' character.
	if len(head) > 1 {
		s = append(s, head[1:]...)
	}

	// Scan the tag.
	if parser.unread < 1 && !yamlParserUpdateBuffer(parser, 1) {
		return false
	}

	// The set of characters that may appear in URI is as follows:
	//
	//      '0'-'9', 'A'-'Z', 'a'-'z', '_', '-', ';', '/', '?', ':', '@', '&',
	//      '=', '+', '$', ',', '.', '!', '~', '*', '\'', '(', ')', '[', ']',
	//      '%'.
	// [Go] Convert this into more reasonable logic.
	for isAlpha(parser.buffer, parser.bufferPos) || parser.buffer[parser.bufferPos] == ';' ||
		parser.buffer[parser.bufferPos] == '/' || parser.buffer[parser.bufferPos] == '?' ||
		parser.buffer[parser.bufferPos] == ':' || parser.buffer[parser.bufferPos] == '@' ||
		parser.buffer[parser.bufferPos] == '&' || parser.buffer[parser.bufferPos] == '=' ||
		parser.buffer[parser.bufferPos] == '+' || parser.buffer[parser.bufferPos] == '$' ||
		parser.buffer[parser.bufferPos] == ',' || parser.buffer[parser.bufferPos] == '.' ||
		parser.buffer[parser.bufferPos] == '!' || parser.buffer[parser.bufferPos] == '~' ||
		parser.buffer[parser.bufferPos] == '*' || parser.buffer[parser.bufferPos] == '\'' ||
		parser.buffer[parser.bufferPos] == '(' || parser.buffer[parser.bufferPos] == ')' ||
		parser.buffer[parser.bufferPos] == '[' || parser.buffer[parser.bufferPos] == ']' ||
		parser.buffer[parser.bufferPos] == '%' {
		// Check if it is a URI-escape sequence.
		if parser.buffer[parser.bufferPos] == '%' {
			if !yamlParserScanURIEscapes(parser, directive, startMark, &s) {
				return false
			}
		} else {
			s = read(parser, s)
		}
		if parser.unread < 1 && !yamlParserUpdateBuffer(parser, 1) {
			return false
		}
		hasTag = true
	}

	if !hasTag {
		yamlParserSetScannerTagError(parser, directive,
			startMark, "did not find expected tag URI")
		return false
	}
	*uri = s
	return true
}

// Decode an URI-escape sequence corresponding to a single UTF-8 character.
func yamlParserScanURIEscapes(parser *yamlParserT, directive bool, startMark yamlMarkT, s *[]byte) bool {

	// Decode the required number of characters.
	w := 1024
	for w > 0 {
		// Check for a URI-escaped octet.
		if parser.unread < 3 && !yamlParserUpdateBuffer(parser, 3) {
			return false
		}

		if !(parser.buffer[parser.bufferPos] == '%' &&
			isHex(parser.buffer, parser.bufferPos+1) &&
			isHex(parser.buffer, parser.bufferPos+2)) {
			return yamlParserSetScannerTagError(parser, directive,
				startMark, "did not find URI escaped octet")
		}

		// Get the octet.
		octet := byte((asHex(parser.buffer, parser.bufferPos+1) << 4) + asHex(parser.buffer, parser.bufferPos+2))

		// If it is the leading octet, determine the length of the UTF-8 sequence.
		if w == 1024 {
			w = width(octet)
			if w == 0 {
				return yamlParserSetScannerTagError(parser, directive,
					startMark, "found an incorrect leading UTF-8 octet")
			}
		} else {
			// Check if the trailing octet is correct.
			if octet&0xC0 != 0x80 {
				return yamlParserSetScannerTagError(parser, directive,
					startMark, "found an incorrect trailing UTF-8 octet")
			}
		}

		// Copy the octet and move the pointers.
		*s = append(*s, octet)
		skip(parser)
		skip(parser)
		skip(parser)
		w--
	}
	return true
}

// Scan a block scalar.
func yamlParserScanBlockScalar(parser *yamlParserT, token *yamlTokenT, literal bool) bool {
	// Eat the indicator '|' or '>'.
	startMark := parser.mark
	skip(parser)

	// Scan the additional block scalar indicators.
	if parser.unread < 1 && !yamlParserUpdateBuffer(parser, 1) {
		return false
	}

	// Check for a chomping indicator.
	var chomping, increment int
	if parser.buffer[parser.bufferPos] == '+' || parser.buffer[parser.bufferPos] == '-' {
		// Set the chomping method and eat the indicator.
		if parser.buffer[parser.bufferPos] == '+' {
			chomping = +1
		} else {
			chomping = -1
		}
		skip(parser)

		// Check for an indentation indicator.
		if parser.unread < 1 && !yamlParserUpdateBuffer(parser, 1) {
			return false
		}
		if isDigit(parser.buffer, parser.bufferPos) {
			// Check that the indentation is greater than 0.
			if parser.buffer[parser.bufferPos] == '0' {
				yamlParserSetScannerError(parser, "while scanning a block scalar",
					startMark, "found an indentation indicator equal to 0")
				return false
			}

			// Get the indentation level and eat the indicator.
			increment = asDigit(parser.buffer, parser.bufferPos)
			skip(parser)
		}

	} else if isDigit(parser.buffer, parser.bufferPos) {
		// Do the same as above, but in the opposite order.

		if parser.buffer[parser.bufferPos] == '0' {
			yamlParserSetScannerError(parser, "while scanning a block scalar",
				startMark, "found an indentation indicator equal to 0")
			return false
		}
		increment = asDigit(parser.buffer, parser.bufferPos)
		skip(parser)

		if parser.unread < 1 && !yamlParserUpdateBuffer(parser, 1) {
			return false
		}
		if parser.buffer[parser.bufferPos] == '+' || parser.buffer[parser.bufferPos] == '-' {
			if parser.buffer[parser.bufferPos] == '+' {
				chomping = +1
			} else {
				chomping = -1
			}
			skip(parser)
		}
	}

	// Eat whitespaces and comments to the end of the line.
	if parser.unread < 1 && !yamlParserUpdateBuffer(parser, 1) {
		return false
	}
	for isBlank(parser.buffer, parser.bufferPos) {
		skip(parser)
		if parser.unread < 1 && !yamlParserUpdateBuffer(parser, 1) {
			return false
		}
	}
	if parser.buffer[parser.bufferPos] == '#' {
		for !isBreakz(parser.buffer, parser.bufferPos) {
			skip(parser)
			if parser.unread < 1 && !yamlParserUpdateBuffer(parser, 1) {
				return false
			}
		}
	}

	// Check if we are at the end of the line.
	if !isBreakz(parser.buffer, parser.bufferPos) {
		yamlParserSetScannerError(parser, "while scanning a block scalar",
			startMark, "did not find expected comment or line break")
		return false
	}

	// Eat a line break.
	if isBreak(parser.buffer, parser.bufferPos) {
		if parser.unread < 2 && !yamlParserUpdateBuffer(parser, 2) {
			return false
		}
		skipLine(parser)
	}

	endMark := parser.mark

	// Set the indentation level if it was specified.
	var indent int
	if increment > 0 {
		if parser.indent >= 0 {
			indent = parser.indent + increment
		} else {
			indent = increment
		}
	}

	// Scan the leading line breaks and determine the indentation level if needed.
	var s, leadingBreak, trailingBreaks []byte
	if !yamlParserScanBlockScalarBreaks(parser, &indent, &trailingBreaks, startMark, &endMark) {
		return false
	}

	// Scan the block scalar content.
	if parser.unread < 1 && !yamlParserUpdateBuffer(parser, 1) {
		return false
	}
	var leadingBlank, trailingBlank bool
	for parser.mark.column == indent && !isZ(parser.buffer, parser.bufferPos) {
		// We are at the beginning of a non-empty line.

		// Is it a trailing whitespace?
		trailingBlank = isBlank(parser.buffer, parser.bufferPos)

		// Check if we need to fold the leading line break.
		if !literal && !leadingBlank && !trailingBlank && len(leadingBreak) > 0 && leadingBreak[0] == '\n' {
			// Do we need to join the lines by space?
			if len(trailingBreaks) == 0 {
				s = append(s, ' ')
			}
		} else {
			s = append(s, leadingBreak...)
		}
		leadingBreak = leadingBreak[:0]

		// Append the remaining line breaks.
		s = append(s, trailingBreaks...)
		trailingBreaks = trailingBreaks[:0]

		// Is it a leading whitespace?
		leadingBlank = isBlank(parser.buffer, parser.bufferPos)

		// Consume the current line.
		for !isBreakz(parser.buffer, parser.bufferPos) {
			s = read(parser, s)
			if parser.unread < 1 && !yamlParserUpdateBuffer(parser, 1) {
				return false
			}
		}

		// Consume the line break.
		if parser.unread < 2 && !yamlParserUpdateBuffer(parser, 2) {
			return false
		}

		leadingBreak = readLine(parser, leadingBreak)

		// Eat the following indentation spaces and line breaks.
		if !yamlParserScanBlockScalarBreaks(parser, &indent, &trailingBreaks, startMark, &endMark) {
			return false
		}
	}

	// Chomp the tail.
	if chomping != -1 {
		s = append(s, leadingBreak...)
	}
	if chomping == 1 {
		s = append(s, trailingBreaks...)
	}

	// Create a token.
	*token = yamlTokenT{
		typ:       yamlScalarToken,
		startMark: startMark,
		endMark:   endMark,
		value:     s,
		style:     yamlLiteralScalarStyle,
	}
	if !literal {
		token.style = yamlFoldedScalarStyle
	}
	return true
}

// Scan indentation spaces and line breaks for a block scalar.  Determine the
// indentation level if needed.
func yamlParserScanBlockScalarBreaks(parser *yamlParserT, indent *int, breaks *[]byte, startMark yamlMarkT, endMark *yamlMarkT) bool {
	*endMark = parser.mark

	// Eat the indentation spaces and line breaks.
	maxIndent := 0
	for {
		// Eat the indentation spaces.
		if parser.unread < 1 && !yamlParserUpdateBuffer(parser, 1) {
			return false
		}
		for (*indent == 0 || parser.mark.column < *indent) && isSpace(parser.buffer, parser.bufferPos) {
			skip(parser)
			if parser.unread < 1 && !yamlParserUpdateBuffer(parser, 1) {
				return false
			}
		}
		if parser.mark.column > maxIndent {
			maxIndent = parser.mark.column
		}

		// Check for a tab character messing the indentation.
		if (*indent == 0 || parser.mark.column < *indent) && isTab(parser.buffer, parser.bufferPos) {
			return yamlParserSetScannerError(parser, "while scanning a block scalar",
				startMark, "found a tab character where an indentation space is expected")
		}

		// Have we found a non-empty line?
		if !isBreak(parser.buffer, parser.bufferPos) {
			break
		}

		// Consume the line break.
		if parser.unread < 2 && !yamlParserUpdateBuffer(parser, 2) {
			return false
		}
		// [Go] Should really be returning breaks instead.
		*breaks = readLine(parser, *breaks)
		*endMark = parser.mark
	}

	// Determine the indentation level if needed.
	if *indent == 0 {
		*indent = maxIndent
		if *indent < parser.indent+1 {
			*indent = parser.indent + 1
		}
		if *indent < 1 {
			*indent = 1
		}
	}
	return true
}

// Scan a quoted scalar.
func yamlParserScanFlowScalar(parser *yamlParserT, token *yamlTokenT, single bool) bool {
	// Eat the left quote.
	startMark := parser.mark
	skip(parser)

	// Consume the content of the quoted scalar.
	var s, leadingBreak, trailingBreaks, whitespaces []byte
	for {
		// Check that there are no document indicators at the beginning of the line.
		if parser.unread < 4 && !yamlParserUpdateBuffer(parser, 4) {
			return false
		}

		if parser.mark.column == 0 &&
			((parser.buffer[parser.bufferPos+0] == '-' &&
				parser.buffer[parser.bufferPos+1] == '-' &&
				parser.buffer[parser.bufferPos+2] == '-') ||
				(parser.buffer[parser.bufferPos+0] == '.' &&
					parser.buffer[parser.bufferPos+1] == '.' &&
					parser.buffer[parser.bufferPos+2] == '.')) &&
			isBlankz(parser.buffer, parser.bufferPos+3) {
			yamlParserSetScannerError(parser, "while scanning a quoted scalar",
				startMark, "found unexpected document indicator")
			return false
		}

		// Check for EOF.
		if isZ(parser.buffer, parser.bufferPos) {
			yamlParserSetScannerError(parser, "while scanning a quoted scalar",
				startMark, "found unexpected end of stream")
			return false
		}

		// Consume non-blank characters.
		leadingBlanks := false
		for !isBlankz(parser.buffer, parser.bufferPos) {
			if single && parser.buffer[parser.bufferPos] == '\'' && parser.buffer[parser.bufferPos+1] == '\'' {
				// Is is an escaped single quote.
				s = append(s, '\'')
				skip(parser)
				skip(parser)

			} else if single && parser.buffer[parser.bufferPos] == '\'' {
				// It is a right single quote.
				break
			} else if !single && parser.buffer[parser.bufferPos] == '"' {
				// It is a right double quote.
				break

			} else if !single && parser.buffer[parser.bufferPos] == '\\' && isBreak(parser.buffer, parser.bufferPos+1) {
				// It is an escaped line break.
				if parser.unread < 3 && !yamlParserUpdateBuffer(parser, 3) {
					return false
				}
				skip(parser)
				skipLine(parser)
				leadingBlanks = true
				break

			} else if !single && parser.buffer[parser.bufferPos] == '\\' {
				// It is an escape sequence.
				codeLength := 0

				// Check the escape character.
				switch parser.buffer[parser.bufferPos+1] {
				case '0':
					s = append(s, 0)
				case 'a':
					s = append(s, '\x07')
				case 'b':
					s = append(s, '\x08')
				case 't', '\t':
					s = append(s, '\x09')
				case 'n':
					s = append(s, '\x0A')
				case 'v':
					s = append(s, '\x0B')
				case 'f':
					s = append(s, '\x0C')
				case 'r':
					s = append(s, '\x0D')
				case 'e':
					s = append(s, '\x1B')
				case ' ':
					s = append(s, '\x20')
				case '"':
					s = append(s, '"')
				case '\'':
					s = append(s, '\'')
				case '\\':
					s = append(s, '\\')
				case 'N': // NEL (#x85)
					s = append(s, '\xC2')
					s = append(s, '\x85')
				case '_': // #xA0
					s = append(s, '\xC2')
					s = append(s, '\xA0')
				case 'L': // LS (#x2028)
					s = append(s, '\xE2')
					s = append(s, '\x80')
					s = append(s, '\xA8')
				case 'P': // PS (#x2029)
					s = append(s, '\xE2')
					s = append(s, '\x80')
					s = append(s, '\xA9')
				case 'x':
					codeLength = 2
				case 'u':
					codeLength = 4
				case 'U':
					codeLength = 8
				default:
					yamlParserSetScannerError(parser, "while parsing a quoted scalar",
						startMark, "found unknown escape character")
					return false
				}

				skip(parser)
				skip(parser)

				// Consume an arbitrary escape code.
				if codeLength > 0 {
					var value int

					// Scan the character value.
					if parser.unread < codeLength && !yamlParserUpdateBuffer(parser, codeLength) {
						return false
					}
					for k := 0; k < codeLength; k++ {
						if !isHex(parser.buffer, parser.bufferPos+k) {
							yamlParserSetScannerError(parser, "while parsing a quoted scalar",
								startMark, "did not find expected hexdecimal number")
							return false
						}
						value = (value << 4) + asHex(parser.buffer, parser.bufferPos+k)
					}

					// Check the value and write the character.
					if (value >= 0xD800 && value <= 0xDFFF) || value > 0x10FFFF {
						yamlParserSetScannerError(parser, "while parsing a quoted scalar",
							startMark, "found invalid Unicode character escape code")
						return false
					}
					if value <= 0x7F {
						s = append(s, byte(value))
					} else if value <= 0x7FF {
						s = append(s, byte(0xC0+(value>>6)))
						s = append(s, byte(0x80+(value&0x3F)))
					} else if value <= 0xFFFF {
						s = append(s, byte(0xE0+(value>>12)))
						s = append(s, byte(0x80+((value>>6)&0x3F)))
						s = append(s, byte(0x80+(value&0x3F)))
					} else {
						s = append(s, byte(0xF0+(value>>18)))
						s = append(s, byte(0x80+((value>>12)&0x3F)))
						s = append(s, byte(0x80+((value>>6)&0x3F)))
						s = append(s, byte(0x80+(value&0x3F)))
					}

					// Advance the pointer.
					for k := 0; k < codeLength; k++ {
						skip(parser)
					}
				}
			} else {
				// It is a non-escaped non-blank character.
				s = read(parser, s)
			}
			if parser.unread < 2 && !yamlParserUpdateBuffer(parser, 2) {
				return false
			}
		}

		if parser.unread < 1 && !yamlParserUpdateBuffer(parser, 1) {
			return false
		}

		// Check if we are at the end of the scalar.
		if single {
			if parser.buffer[parser.bufferPos] == '\'' {
				break
			}
		} else {
			if parser.buffer[parser.bufferPos] == '"' {
				break
			}
		}

		// Consume blank characters.
		for isBlank(parser.buffer, parser.bufferPos) || isBreak(parser.buffer, parser.bufferPos) {
			if isBlank(parser.buffer, parser.bufferPos) {
				// Consume a space or a tab character.
				if !leadingBlanks {
					whitespaces = read(parser, whitespaces)
				} else {
					skip(parser)
				}
			} else {
				if parser.unread < 2 && !yamlParserUpdateBuffer(parser, 2) {
					return false
				}

				// Check if it is a first line break.
				if !leadingBlanks {
					whitespaces = whitespaces[:0]
					leadingBreak = readLine(parser, leadingBreak)
					leadingBlanks = true
				} else {
					trailingBreaks = readLine(parser, trailingBreaks)
				}
			}
			if parser.unread < 1 && !yamlParserUpdateBuffer(parser, 1) {
				return false
			}
		}

		// Join the whitespaces or fold line breaks.
		if leadingBlanks {
			// Do we need to fold line breaks?
			if len(leadingBreak) > 0 && leadingBreak[0] == '\n' {
				if len(trailingBreaks) == 0 {
					s = append(s, ' ')
				} else {
					s = append(s, trailingBreaks...)
				}
			} else {
				s = append(s, leadingBreak...)
				s = append(s, trailingBreaks...)
			}
			trailingBreaks = trailingBreaks[:0]
			leadingBreak = leadingBreak[:0]
		} else {
			s = append(s, whitespaces...)
			whitespaces = whitespaces[:0]
		}
	}

	// Eat the right quote.
	skip(parser)
	endMark := parser.mark

	// Create a token.
	*token = yamlTokenT{
		typ:       yamlScalarToken,
		startMark: startMark,
		endMark:   endMark,
		value:     s,
		style:     yamlSingleQuotedScalarStyle,
	}
	if !single {
		token.style = yamlDoubleQuotedScalarStyle
	}
	return true
}

// Scan a plain scalar.
func yamlParserScanPlainScalar(parser *yamlParserT, token *yamlTokenT) bool {

	var s, leadingBreak, trailingBreaks, whitespaces []byte
	var leadingBlanks bool
	var indent = parser.indent + 1

	startMark := parser.mark
	endMark := parser.mark

	// Consume the content of the plain scalar.
	for {
		// Check for a document indicator.
		if parser.unread < 4 && !yamlParserUpdateBuffer(parser, 4) {
			return false
		}
		if parser.mark.column == 0 &&
			((parser.buffer[parser.bufferPos+0] == '-' &&
				parser.buffer[parser.bufferPos+1] == '-' &&
				parser.buffer[parser.bufferPos+2] == '-') ||
				(parser.buffer[parser.bufferPos+0] == '.' &&
					parser.buffer[parser.bufferPos+1] == '.' &&
					parser.buffer[parser.bufferPos+2] == '.')) &&
			isBlankz(parser.buffer, parser.bufferPos+3) {
			break
		}

		// Check for a comment.
		if parser.buffer[parser.bufferPos] == '#' {
			break
		}

		// Consume non-blank characters.
		for !isBlankz(parser.buffer, parser.bufferPos) {

			// Check for indicators that may end a plain scalar.
			if (parser.buffer[parser.bufferPos] == ':' && isBlankz(parser.buffer, parser.bufferPos+1)) ||
				(parser.flowLevel > 0 &&
					(parser.buffer[parser.bufferPos] == ',' ||
						parser.buffer[parser.bufferPos] == '?' || parser.buffer[parser.bufferPos] == '[' ||
						parser.buffer[parser.bufferPos] == ']' || parser.buffer[parser.bufferPos] == '{' ||
						parser.buffer[parser.bufferPos] == '}')) {
				break
			}

			// Check if we need to join whitespaces and breaks.
			if leadingBlanks || len(whitespaces) > 0 {
				if leadingBlanks {
					// Do we need to fold line breaks?
					if leadingBreak[0] == '\n' {
						if len(trailingBreaks) == 0 {
							s = append(s, ' ')
						} else {
							s = append(s, trailingBreaks...)
						}
					} else {
						s = append(s, leadingBreak...)
						s = append(s, trailingBreaks...)
					}
					trailingBreaks = trailingBreaks[:0]
					leadingBreak = leadingBreak[:0]
					leadingBlanks = false
				} else {
					s = append(s, whitespaces...)
					whitespaces = whitespaces[:0]
				}
			}

			// Copy the character.
			s = read(parser, s)

			endMark = parser.mark
			if parser.unread < 2 && !yamlParserUpdateBuffer(parser, 2) {
				return false
			}
		}

		// Is it the end?
		if !(isBlank(parser.buffer, parser.bufferPos) || isBreak(parser.buffer, parser.bufferPos)) {
			break
		}

		// Consume blank characters.
		if parser.unread < 1 && !yamlParserUpdateBuffer(parser, 1) {
			return false
		}

		for isBlank(parser.buffer, parser.bufferPos) || isBreak(parser.buffer, parser.bufferPos) {
			if isBlank(parser.buffer, parser.bufferPos) {

				// Check for tab characters that abuse indentation.
				if leadingBlanks && parser.mark.column < indent && isTab(parser.buffer, parser.bufferPos) {
					yamlParserSetScannerError(parser, "while scanning a plain scalar",
						startMark, "found a tab character that violates indentation")
					return false
				}

				// Consume a space or a tab character.
				if !leadingBlanks {
					whitespaces = read(parser, whitespaces)
				} else {
					skip(parser)
				}
			} else {
				if parser.unread < 2 && !yamlParserUpdateBuffer(parser, 2) {
					return false
				}

				// Check if it is a first line break.
				if !leadingBlanks {
					whitespaces = whitespaces[:0]
					leadingBreak = readLine(parser, leadingBreak)
					leadingBlanks = true
				} else {
					trailingBreaks = readLine(parser, trailingBreaks)
				}
			}
			if parser.unread < 1 && !yamlParserUpdateBuffer(parser, 1) {
				return false
			}
		}

		// Check indentation level.
		if parser.flowLevel == 0 && parser.mark.column < indent {
			break
		}
	}

	// Create a token.
	*token = yamlTokenT{
		typ:       yamlScalarToken,
		startMark: startMark,
		endMark:   endMark,
		value:     s,
		style:     yamlPlainScalarStyle,
	}

	// Note that we change the 'simple_key_allowed' flag.
	if leadingBlanks {
		parser.simpleKeyAllowed = true
	}
	return true
}
