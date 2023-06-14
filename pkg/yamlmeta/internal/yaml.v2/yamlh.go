// Copyright 2020 VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package yaml

import (
	"fmt"
	"io"
)

// The version directive data.
type yamlVersionDirectiveT struct {
	major int8 // The major version number.
	minor int8 // The minor version number.
}

// The tag directive data.
type yamlTagDirectiveT struct {
	handle []byte // The tag handle.
	prefix []byte // The tag prefix.
}

type yamlEncodingT int

// The stream encoding.
const (
	// Let the parser choose the encoding.
	yamlAnyEncoding yamlEncodingT = iota

	yamlUtf8Encoding    // The default UTF-8 encoding.
	yamlUtf16leEncoding // The UTF-16-LE encoding with BOM.
	yamlUtf16beEncoding // The UTF-16-BE encoding with BOM.
)

type yamlBreakT int

// Line break types.
const (
	// Let the parser choose the break type.
	yamlAnyBreak yamlBreakT = iota

	yamlCrBreak   // Use CR for line breaks (Mac style).
	yamlLnBreak   // Use LN for line breaks (Unix style).
	yamlCrlnBreak // Use CR LN for line breaks (DOS style).
)

type yamlErrorTypeT int

// Many bad things could happen with the parser and emitter.
const (
	// No error is produced.
	yamlNoError yamlErrorTypeT = iota

	yamlMemoryError   // Cannot allocate or reallocate a block of memory.
	yamlReaderError   // Cannot read or decode the input stream.
	yamlScannerError  // Cannot scan the input stream.
	yamlParserError   // Cannot parse the input stream.
	yamlComposerError // Cannot compose a YAML document.
	yamlWriterError   // Cannot write to the output stream.
	yamlEmitterError  // Cannot emit a YAML stream.
)

// The pointer position.
type yamlMarkT struct {
	index  int // The position index.
	line   int // The position line.
	column int // The position column.
}

// Node Styles

type yamlStyleT int8

type yamlScalarStyleT yamlStyleT

// Scalar styles.
const (
	// Let the emitter choose the style.
	yamlAnyScalarStyle yamlScalarStyleT = iota

	yamlPlainScalarStyle        // The plain scalar style.
	yamlSingleQuotedScalarStyle // The single-quoted scalar style.
	yamlDoubleQuotedScalarStyle // The double-quoted scalar style.
	yamlLiteralScalarStyle      // The literal scalar style.
	yamlFoldedScalarStyle       // The folded scalar style.
)

type yamlSequenceStyleT yamlStyleT

// Sequence styles.
const (
	// Let the emitter choose the style.
	yamlAnySequenceStyle yamlSequenceStyleT = iota

	yamlBlockSequenceStyle // The block sequence style.
	yamlFlowSequenceStyle  // The flow sequence style.
)

type yamlMappingStyleT yamlStyleT

// Mapping styles.
const (
	// Let the emitter choose the style.
	yamlAnyMappingStyle yamlMappingStyleT = iota

	yamlBlockMappingStyle // The block mapping style.
	yamlFlowMappingStyle  // The flow mapping style.
)

// Tokens

type yamlTokenTypeT int

// Token types.
const (
	// An empty token.
	yamlNoToken yamlTokenTypeT = iota

	yamlStreamStartToken // A STREAM-START token.
	yamlStreamEndToken   // A STREAM-END token.

	yamlVersionDirectiveToken // A VERSION-DIRECTIVE token.
	yamlTagDirectiveToken     // A TAG-DIRECTIVE token.
	yamlDocumentStartToken    // A DOCUMENT-START token.
	yamlDocumentEndToken      // A DOCUMENT-END token.

	yamlBlockSequenceStartToken // A BLOCK-SEQUENCE-START token.
	yamlBlockMappingStartToken  // A BLOCK-SEQUENCE-END token.
	yamlBlockEndToken           // A BLOCK-END token.

	yamlFlowSequenceStartToken // A FLOW-SEQUENCE-START token.
	yamlFlowSequenceEndToken   // A FLOW-SEQUENCE-END token.
	yamlFlowMappingStartToken  // A FLOW-MAPPING-START token.
	yamlFlowMappingEndToken    // A FLOW-MAPPING-END token.

	yamlBlockEntryToken // A BLOCK-ENTRY token.
	yamlFlowEntryToken  // A FLOW-ENTRY token.
	yamlKeyToken        // A KEY token.
	yamlValueToken      // A VALUE token.

	yamlAliasToken  // An ALIAS token.
	yamlAnchorToken // An ANCHOR token.
	yamlTagToken    // A TAG token.
	yamlScalarToken // A SCALAR token.
	yamlCommentToken
)

func (tt yamlTokenTypeT) String() string {
	switch tt {
	case yamlNoToken:
		return "yaml_NO_TOKEN"
	case yamlStreamStartToken:
		return "yaml_STREAM_START_TOKEN"
	case yamlStreamEndToken:
		return "yaml_STREAM_END_TOKEN"
	case yamlVersionDirectiveToken:
		return "yaml_VERSION_DIRECTIVE_TOKEN"
	case yamlTagDirectiveToken:
		return "yaml_TAG_DIRECTIVE_TOKEN"
	case yamlDocumentStartToken:
		return "yaml_DOCUMENT_START_TOKEN"
	case yamlDocumentEndToken:
		return "yaml_DOCUMENT_END_TOKEN"
	case yamlBlockSequenceStartToken:
		return "yaml_BLOCK_SEQUENCE_START_TOKEN"
	case yamlBlockMappingStartToken:
		return "yaml_BLOCK_MAPPING_START_TOKEN"
	case yamlBlockEndToken:
		return "yaml_BLOCK_END_TOKEN"
	case yamlFlowSequenceStartToken:
		return "yaml_FLOW_SEQUENCE_START_TOKEN"
	case yamlFlowSequenceEndToken:
		return "yaml_FLOW_SEQUENCE_END_TOKEN"
	case yamlFlowMappingStartToken:
		return "yaml_FLOW_MAPPING_START_TOKEN"
	case yamlFlowMappingEndToken:
		return "yaml_FLOW_MAPPING_END_TOKEN"
	case yamlBlockEntryToken:
		return "yaml_BLOCK_ENTRY_TOKEN"
	case yamlFlowEntryToken:
		return "yaml_FLOW_ENTRY_TOKEN"
	case yamlKeyToken:
		return "yaml_KEY_TOKEN"
	case yamlValueToken:
		return "yaml_VALUE_TOKEN"
	case yamlAliasToken:
		return "yaml_ALIAS_TOKEN"
	case yamlAnchorToken:
		return "yaml_ANCHOR_TOKEN"
	case yamlTagToken:
		return "yaml_TAG_TOKEN"
	case yamlScalarToken:
		return "yaml_SCALAR_TOKEN"
	case yamlCommentToken:
		return "yaml_COMMENT_TOKEN"
	}
	return "<unknown token>"
}

// The token structure.
type yamlTokenT struct {
	// The token type.
	typ yamlTokenTypeT

	// The start/end of the token.
	startMark, endMark yamlMarkT

	// The stream encoding (for yaml_STREAM_START_TOKEN).
	encoding yamlEncodingT

	// The alias/anchor/scalar value or tag/tag directive handle
	// (for yaml_ALIAS_TOKEN, yaml_ANCHOR_TOKEN, yaml_SCALAR_TOKEN, yaml_TAG_TOKEN, yaml_TAG_DIRECTIVE_TOKEN).
	value []byte

	// The tag suffix (for yaml_TAG_TOKEN).
	suffix []byte

	// The tag directive prefix (for yaml_TAG_DIRECTIVE_TOKEN).
	prefix []byte

	// The scalar style (for yaml_SCALAR_TOKEN).
	style yamlScalarStyleT

	// The version directive major/minor (for yaml_VERSION_DIRECTIVE_TOKEN).
	major, minor int8
}

func (t *yamlTokenT) String() string {
	return fmt.Sprintf("Token(typ=%s, value=%s)", t.typ.String(), string(t.value))
}

// Events

type yamlEventTypeT int8

// Event types.
const (
	// An empty event.
	yamlNoEvent yamlEventTypeT = iota

	yamlStreamStartEvent   // A STREAM-START event.
	yamlStreamEndEvent     // A STREAM-END event.
	yamlDocumentStartEvent // A DOCUMENT-START event.
	yamlDocumentEndEvent   // A DOCUMENT-END event.
	yamlAliasEvent         // An ALIAS event.
	yamlScalarEvent        // A SCALAR event.
	yamlSequenceStartEvent // A SEQUENCE-START event.
	yamlSequenceEndEvent   // A SEQUENCE-END event.
	yamlMappingStartEvent  // A MAPPING-START event.
	yamlMappingEndEvent    // A MAPPING-END event.
)

var eventStrings = []string{
	yamlNoEvent:            "none",
	yamlStreamStartEvent:   "stream start",
	yamlStreamEndEvent:     "stream end",
	yamlDocumentStartEvent: "document start",
	yamlDocumentEndEvent:   "document end",
	yamlAliasEvent:         "alias",
	yamlScalarEvent:        "scalar",
	yamlSequenceStartEvent: "sequence start",
	yamlSequenceEndEvent:   "sequence end",
	yamlMappingStartEvent:  "mapping start",
	yamlMappingEndEvent:    "mapping end",
}

func (e yamlEventTypeT) String() string {
	if e < 0 || int(e) >= len(eventStrings) {
		return fmt.Sprintf("unknown event %d", e)
	}
	return eventStrings[e]
}

// The event structure.
type yamlEventT struct {

	// The event type.
	typ yamlEventTypeT

	// The start and end of the event.
	startMark, endMark yamlMarkT

	// The document encoding (for yaml_STREAM_START_EVENT).
	encoding yamlEncodingT

	// The version directive (for yaml_DOCUMENT_START_EVENT).
	versionDirective *yamlVersionDirectiveT

	// The list of tag directives (for yaml_DOCUMENT_START_EVENT).
	tagDirectives []yamlTagDirectiveT

	// The anchor (for yaml_SCALAR_EVENT, yaml_SEQUENCE_START_EVENT, yaml_MAPPING_START_EVENT, yaml_ALIAS_EVENT).
	anchor []byte

	// The tag (for yaml_SCALAR_EVENT, yaml_SEQUENCE_START_EVENT, yaml_MAPPING_START_EVENT).
	tag []byte

	// The scalar value (for yaml_SCALAR_EVENT).
	value []byte

	// Is the document start/end indicator implicit, or the tag optional?
	// (for yaml_DOCUMENT_START_EVENT, yaml_DOCUMENT_END_EVENT, yaml_SEQUENCE_START_EVENT, yaml_MAPPING_START_EVENT, yaml_SCALAR_EVENT).
	implicit bool

	// Is the tag optional for any non-plain style? (for yaml_SCALAR_EVENT).
	quotedImplicit bool

	// The style (for yaml_SCALAR_EVENT, yaml_SEQUENCE_START_EVENT, yaml_MAPPING_START_EVENT).
	style yamlStyleT
}

func (e *yamlEventT) scalarStyle() yamlScalarStyleT     { return yamlScalarStyleT(e.style) }
func (e *yamlEventT) sequenceStyle() yamlSequenceStyleT { return yamlSequenceStyleT(e.style) }
func (e *yamlEventT) mappingStyle() yamlMappingStyleT   { return yamlMappingStyleT(e.style) }

// Nodes

const (
	yamlNullTag      = "tag:yaml.org,2002:null"      // The tag !!null with the only possible value: null.
	yamlBoolTag      = "tag:yaml.org,2002:bool"      // The tag !!bool with the values: true and false.
	yamlStrTag       = "tag:yaml.org,2002:str"       // The tag !!str for string values.
	yamlIntTag       = "tag:yaml.org,2002:int"       // The tag !!int for integer values.
	yamlFloatTag     = "tag:yaml.org,2002:float"     // The tag !!float for float values.
	yamlTimestampTag = "tag:yaml.org,2002:timestamp" // The tag !!timestamp for date and time values.

	yamlSeqTag = "tag:yaml.org,2002:seq" // The tag !!seq is used to denote sequences.
	yamlMapTag = "tag:yaml.org,2002:map" // The tag !!map is used to denote mapping.

	// Not in original libyaml.
	yamlBinaryTag = "tag:yaml.org,2002:binary"
	yamlMergeTag  = "tag:yaml.org,2002:merge"

	yamlDefaultScalarTag   = yamlStrTag // The default scalar tag is !!str.
	yamlDefaultSequenceTag = yamlSeqTag // The default sequence tag is !!seq.
	yamlDefaultMappingTag  = yamlMapTag // The default mapping tag is !!map.
)

type yamlNodeTypeT int

// Node types.
const (
	// An empty node.
	yamlNoNode yamlNodeTypeT = iota

	yamlScalarNode   // A scalar node.
	yamlSequenceNode // A sequence node.
	yamlMappingNode  // A mapping node.
)

// An element of a sequence node.
type yamlNodeItemT int

// An element of a mapping node.
type yamlNodePairT struct {
	key   int // The key of the element.
	value int // The value of the element.
}

// The node structure.
type yamlNodeT struct {
	typ yamlNodeTypeT // The node type.
	tag []byte        // The node tag.

	// The node data.

	// The scalar parameters (for yaml_SCALAR_NODE).
	scalar struct {
		value  []byte           // The scalar value.
		length int              // The length of the scalar value.
		style  yamlScalarStyleT // The scalar style.
	}

	// The sequence parameters (for YAML_SEQUENCE_NODE).
	sequence struct {
		itemsData []yamlNodeItemT    // The stack of sequence items.
		style     yamlSequenceStyleT // The sequence style.
	}

	// The mapping parameters (for yaml_MAPPING_NODE).
	mapping struct {
		pairsData  []yamlNodePairT   // The stack of mapping pairs (key, value).
		pairsStart *yamlNodePairT    // The beginning of the stack.
		pairsEnd   *yamlNodePairT    // The end of the stack.
		pairsTop   *yamlNodePairT    // The top of the stack.
		style      yamlMappingStyleT // The mapping style.
	}

	startMark yamlMarkT // The beginning of the node.
	endMark   yamlMarkT // The end of the node.

}

// The document structure.
type yamlDocumentT struct {

	// The document nodes.
	nodes []yamlNodeT

	// The version directive.
	versionDirective *yamlVersionDirectiveT

	// The list of tag directives.
	tagDirectivesData  []yamlTagDirectiveT
	tagDirectivesStart int // The beginning of the tag directives list.
	tagDirectivesEnd   int // The end of the tag directives list.

	startImplicit int // Is the document start indicator implicit?
	endImplicit   int // Is the document end indicator implicit?

	// The start/end of the document.
	startMark, endMark yamlMarkT
}

// The prototype of a read handler.
//
// The read handler is called when the parser needs to read more bytes from the
// source. The handler should write not more than size bytes to the buffer.
// The number of written bytes should be set to the size_read variable.
//
// [in,out]   data        A pointer to an application data specified by
//                        yaml_parser_set_input().
// [out]      buffer      The buffer to write the data from the source.
// [in]       size        The size of the buffer.
// [out]      size_read   The actual number of bytes read from the source.
//
// On success, the handler should return 1.  If the handler failed,
// the returned value should be 0. On EOF, the handler should set the
// size_read to 0 and return 1.
type yamlReadHandlerT func(parser *yamlParserT, buffer []byte) (n int, err error)

// This structure holds information about a potential simple key.
type yamlSimpleKeyT struct {
	possible    bool      // Is a simple key possible?
	required    bool      // Is a simple key required?
	tokenNumber int       // The number of the token.
	mark        yamlMarkT // The position mark.
}

// The states of the parser.
type yamlParserStateT int

const (
	yamlParseStreamStartState yamlParserStateT = iota

	yamlParseImplicitDocumentStartState         // Expect the beginning of an implicit document.
	yamlParseDocumentStartState                 // Expect DOCUMENT-START.
	yamlParseDocumentContentState               // Expect the content of a document.
	yamlParseDocumentEndState                   // Expect DOCUMENT-END.
	yamlParseBlockNodeState                     // Expect a block node.
	yamlParseBlockNodeOrIndentlessSequenceState // Expect a block node or indentless sequence.
	yamlParseFlowNodeState                      // Expect a flow node.
	yamlParseBlockSequenceFirstEntryState       // Expect the first entry of a block sequence.
	yamlParseBlockSequenceEntryState            // Expect an entry of a block sequence.
	yamlParseIndentlessSequenceEntryState       // Expect an entry of an indentless sequence.
	yamlParseBlockMappingFirstKeyState          // Expect the first key of a block mapping.
	yamlParseBlockMappingKeyState               // Expect a block mapping key.
	yamlParseBlockMappingValueState             // Expect a block mapping value.
	yamlParseFlowSequenceFirstEntryState        // Expect the first entry of a flow sequence.
	yamlParseFlowSequenceEntryState             // Expect an entry of a flow sequence.
	yamlParseFlowSequenceEntryMappingKeyState   // Expect a key of an ordered mapping.
	yamlParseFlowSequenceEntryMappingValueState // Expect a value of an ordered mapping.
	yamlParseFlowSequenceEntryMappingEndState   // Expect the and of an ordered mapping entry.
	yamlParseFlowMappingFirstKeyState           // Expect the first key of a flow mapping.
	yamlParseFlowMappingKeyState                // Expect a key of a flow mapping.
	yamlParseFlowMappingValueState              // Expect a value of a flow mapping.
	yamlParseFlowMappingEmptyValueState         // Expect an empty value of a flow mapping.
	yamlParseEndState                           // Expect nothing.
)

func (ps yamlParserStateT) String() string {
	switch ps {
	case yamlParseStreamStartState:
		return "yaml_PARSE_STREAM_START_STATE"
	case yamlParseImplicitDocumentStartState:
		return "yaml_PARSE_IMPLICIT_DOCUMENT_START_STATE"
	case yamlParseDocumentStartState:
		return "yaml_PARSE_DOCUMENT_START_STATE"
	case yamlParseDocumentContentState:
		return "yaml_PARSE_DOCUMENT_CONTENT_STATE"
	case yamlParseDocumentEndState:
		return "yaml_PARSE_DOCUMENT_END_STATE"
	case yamlParseBlockNodeState:
		return "yaml_PARSE_BLOCK_NODE_STATE"
	case yamlParseBlockNodeOrIndentlessSequenceState:
		return "yaml_PARSE_BLOCK_NODE_OR_INDENTLESS_SEQUENCE_STATE"
	case yamlParseFlowNodeState:
		return "yaml_PARSE_FLOW_NODE_STATE"
	case yamlParseBlockSequenceFirstEntryState:
		return "yaml_PARSE_BLOCK_SEQUENCE_FIRST_ENTRY_STATE"
	case yamlParseBlockSequenceEntryState:
		return "yaml_PARSE_BLOCK_SEQUENCE_ENTRY_STATE"
	case yamlParseIndentlessSequenceEntryState:
		return "yaml_PARSE_INDENTLESS_SEQUENCE_ENTRY_STATE"
	case yamlParseBlockMappingFirstKeyState:
		return "yaml_PARSE_BLOCK_MAPPING_FIRST_KEY_STATE"
	case yamlParseBlockMappingKeyState:
		return "yaml_PARSE_BLOCK_MAPPING_KEY_STATE"
	case yamlParseBlockMappingValueState:
		return "yaml_PARSE_BLOCK_MAPPING_VALUE_STATE"
	case yamlParseFlowSequenceFirstEntryState:
		return "yaml_PARSE_FLOW_SEQUENCE_FIRST_ENTRY_STATE"
	case yamlParseFlowSequenceEntryState:
		return "yaml_PARSE_FLOW_SEQUENCE_ENTRY_STATE"
	case yamlParseFlowSequenceEntryMappingKeyState:
		return "yaml_PARSE_FLOW_SEQUENCE_ENTRY_MAPPING_KEY_STATE"
	case yamlParseFlowSequenceEntryMappingValueState:
		return "yaml_PARSE_FLOW_SEQUENCE_ENTRY_MAPPING_VALUE_STATE"
	case yamlParseFlowSequenceEntryMappingEndState:
		return "yaml_PARSE_FLOW_SEQUENCE_ENTRY_MAPPING_END_STATE"
	case yamlParseFlowMappingFirstKeyState:
		return "yaml_PARSE_FLOW_MAPPING_FIRST_KEY_STATE"
	case yamlParseFlowMappingKeyState:
		return "yaml_PARSE_FLOW_MAPPING_KEY_STATE"
	case yamlParseFlowMappingValueState:
		return "yaml_PARSE_FLOW_MAPPING_VALUE_STATE"
	case yamlParseFlowMappingEmptyValueState:
		return "yaml_PARSE_FLOW_MAPPING_EMPTY_VALUE_STATE"
	case yamlParseEndState:
		return "yaml_PARSE_END_STATE"
	}
	return "<unknown parser state>"
}

// This structure holds aliases data.
type yamlAliasDataT struct {
	anchor []byte    // The anchor.
	index  int       // The node id.
	mark   yamlMarkT // The anchor mark.
}

// The parser structure.
//
// All members are internal. Manage the structure using the
// yaml_parser_ family of functions.
type yamlParserT struct {

	// Error handling

	error yamlErrorTypeT // Error type.

	problem string // Error description.

	// The byte about which the problem occurred.
	problemOffset int
	problemValue  int
	problemMark   yamlMarkT

	// The error context.
	context     string
	contextMark yamlMarkT

	// Reader stuff

	readHandler yamlReadHandlerT // Read handler.

	inputReader io.Reader // File input data.
	input       []byte    // String input data.
	inputPos    int

	eof bool // EOF flag

	buffer    []byte // The working buffer.
	bufferPos int    // The current position of the buffer.

	unread int // The number of unread characters in the buffer.

	rawBuffer    []byte // The raw buffer.
	rawBufferPos int    // The current position of the buffer.

	encoding yamlEncodingT // The input encoding.

	offset int       // The offset of the current position (in bytes).
	mark   yamlMarkT // The mark of the current position.

	// Scanner stuff

	streamStartProduced bool // Have we started to scan the input stream?
	streamEndProduced   bool // Have we reached the end of the input stream?

	flowLevel int // The number of unclosed '[' and '{' indicators.

	tokens         []yamlTokenT // The tokens queue.
	tokensHead     int          // The head of the tokens queue.
	tokensParsed   int          // The number of tokens fetched from the queue.
	tokenAvailable bool         // Does the tokens queue contain a token ready for dequeueing.

	indent  int   // The current indentation level.
	indents []int // The indentation levels stack.

	simpleKeyAllowed bool             // May a simple key occur at the current position?
	simpleKeys       []yamlSimpleKeyT // The stack of simple keys.

	// Parser stuff

	state         yamlParserStateT    // The current parser state.
	states        []yamlParserStateT  // The parser states stack.
	marks         []yamlMarkT         // The stack of marks.
	tagDirectives []yamlTagDirectiveT // The list of TAG directives.

	// Dumper stuff

	aliases []yamlAliasDataT // The alias data.

	document *yamlDocumentT // The currently parsed document.

	comments            []yamlTokenT
	pendingSeqItemEvent *yamlEventT
}

// Emitter Definitions

// The prototype of a write handler.
//
// The write handler is called when the emitter needs to flush the accumulated
// characters to the output.  The handler should write @a size bytes of the
// @a buffer to the output.
//
// @param[in,out]   data        A pointer to an application data specified by
//                              yaml_emitter_set_output().
// @param[in]       buffer      The buffer with bytes to be written.
// @param[in]       size        The size of the buffer.
//
// @returns On success, the handler should return @c 1.  If the handler failed,
// the returned value should be @c 0.
//
type yamlWriteHandlerT func(emitter *yamlEmitterT, buffer []byte) error

type yamlEmitterStateT int

// The emitter states.
const (
	// Expect STREAM-START.
	yamlEmitStreamStartState yamlEmitterStateT = iota

	yamlEmitFirstDocumentStartState      // Expect the first DOCUMENT-START or STREAM-END.
	yamlEmitDocumentStartState           // Expect DOCUMENT-START or STREAM-END.
	yamlEmitDocumentContentState         // Expect the content of a document.
	yamlEmitDocumentEndState             // Expect DOCUMENT-END.
	yamlEmitFlowSequenceFirstItemState   // Expect the first item of a flow sequence.
	yamlEmitFlowSequenceItemState        // Expect an item of a flow sequence.
	yamlEmitFlowMappingFirstKeyState     // Expect the first key of a flow mapping.
	yamlEmitFlowMappingKeyState          // Expect a key of a flow mapping.
	yamlEmitFlowMappingSimpleValueState  // Expect a value for a simple key of a flow mapping.
	yamlEmitFlowMappingValueState        // Expect a value of a flow mapping.
	yamlEmitBlockSequenceFirstItemState  // Expect the first item of a block sequence.
	yamlEmitBlockSequenceItemState       // Expect an item of a block sequence.
	yamlEmitBlockMappingFirstKeyState    // Expect the first key of a block mapping.
	yamlEmitBlockMappingKeyState         // Expect the key of a block mapping.
	yamlEmitBlockMappingSimpleValueState // Expect a value for a simple key of a block mapping.
	yamlEmitBlockMappingValueState       // Expect a value of a block mapping.
	yamlEmitEndState                     // Expect nothing.
)

// The emitter structure.
//
// All members are internal.  Manage the structure using the @c yaml_emitter_
// family of functions.
type yamlEmitterT struct {

	// Error handling

	error   yamlErrorTypeT // Error type.
	problem string         // Error description.

	// Writer stuff

	writeHandler yamlWriteHandlerT // Write handler.

	outputBuffer *[]byte   // String output data.
	outputWriter io.Writer // File output data.

	buffer    []byte // The working buffer.
	bufferPos int    // The current position of the buffer.

	rawBuffer    []byte // The raw buffer.
	rawBufferPos int    // The current position of the buffer.

	encoding yamlEncodingT // The stream encoding.

	// Emitter stuff

	canonical  bool       // If the output is in the canonical style?
	bestIndent int        // The number of indentation spaces.
	bestWidth  int        // The preferred width of the output lines.
	unicode    bool       // Allow unescaped non-ASCII characters?
	lineBreak  yamlBreakT // The preferred line break.

	state  yamlEmitterStateT   // The current emitter state.
	states []yamlEmitterStateT // The stack of states.

	events     []yamlEventT // The event queue.
	eventsHead int          // The head of the event queue.

	indents []int // The stack of indentation levels.

	tagDirectives []yamlTagDirectiveT // The list of tag directives.

	indent int // The current indentation level.

	flowLevel int // The current flow level.

	rootContext      bool // Is it the document root context?
	sequenceContext  bool // Is it a sequence context?
	mappingContext   bool // Is it a mapping context?
	simpleKeyContext bool // Is it a simple mapping key context?

	line       int  // The current line.
	column     int  // The current column.
	whitespace bool // If the last character was a whitespace?
	indention  bool // If the last character was an indentation character (' ', '-', '?', ':')?
	openEnded  bool // If an explicit document end is required?

	// Anchor analysis.
	anchorData struct {
		anchor []byte // The anchor value.
		alias  bool   // Is it an alias?
	}

	// Tag analysis.
	tagData struct {
		handle []byte // The tag handle.
		suffix []byte // The tag suffix.
	}

	// Scalar analysis.
	scalarData struct {
		value               []byte           // The scalar value.
		multiline           bool             // Does the scalar contain line breaks?
		flowPlainAllowed    bool             // Can the scalar be expessed in the flow plain style?
		blockPlainAllowed   bool             // Can the scalar be expressed in the block plain style?
		singleQuotedAllowed bool             // Can the scalar be expressed in the single quoted style?
		blockAllowed        bool             // Can the scalar be expressed in the literal or folded styles?
		style               yamlScalarStyleT // The output style.
	}

	// Dumper stuff

	opened bool // If the stream was already opened?
	closed bool // If the stream was already closed?

	// The information associated with the document nodes.
	anchors *struct {
		references int  // The number of references.
		anchor     int  // The anchor id.
		serialized bool // If the node has been emitted?
	}

	lastAnchorID int // The last assigned anchor id.

	document *yamlDocumentT // The currently emitted document.
}
