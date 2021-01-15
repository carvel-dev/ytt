// Copyright 2020 VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package yaml

import (
	"io"
)

// Set the reader error and return 0.
func yamlParserSetReaderError(parser *yamlParserT, problem string, offset int, value int) bool {
	parser.error = yamlReaderError
	parser.problem = problem
	parser.problemOffset = offset
	parser.problemValue = value
	return false
}

// Byte order marks.
const (
	bomUtf8    = "\xef\xbb\xbf"
	bomUtf16le = "\xff\xfe"
	bomUtf16be = "\xfe\xff"
)

// Determine the input stream encoding by checking the BOM symbol. If no BOM is
// found, the UTF-8 encoding is assumed. Return 1 on success, 0 on failure.
func yamlParserDetermineEncoding(parser *yamlParserT) bool {
	// Ensure that we had enough bytes in the raw buffer.
	for !parser.eof && len(parser.rawBuffer)-parser.rawBufferPos < 3 {
		if !yamlParserUpdateRawBuffer(parser) {
			return false
		}
	}

	// Determine the encoding.
	buf := parser.rawBuffer
	pos := parser.rawBufferPos
	avail := len(buf) - pos
	if avail >= 2 && buf[pos] == bomUtf16le[0] && buf[pos+1] == bomUtf16le[1] {
		parser.encoding = yamlUtf16leEncoding
		parser.rawBufferPos += 2
		parser.offset += 2
	} else if avail >= 2 && buf[pos] == bomUtf16be[0] && buf[pos+1] == bomUtf16be[1] {
		parser.encoding = yamlUtf16beEncoding
		parser.rawBufferPos += 2
		parser.offset += 2
	} else if avail >= 3 && buf[pos] == bomUtf8[0] && buf[pos+1] == bomUtf8[1] && buf[pos+2] == bomUtf8[2] {
		parser.encoding = yamlUtf8Encoding
		parser.rawBufferPos += 3
		parser.offset += 3
	} else {
		parser.encoding = yamlUtf8Encoding
	}
	return true
}

// Update the raw buffer.
func yamlParserUpdateRawBuffer(parser *yamlParserT) bool {
	sizeRead := 0

	// Return if the raw buffer is full.
	if parser.rawBufferPos == 0 && len(parser.rawBuffer) == cap(parser.rawBuffer) {
		return true
	}

	// Return on EOF.
	if parser.eof {
		return true
	}

	// Move the remaining bytes in the raw buffer to the beginning.
	if parser.rawBufferPos > 0 && parser.rawBufferPos < len(parser.rawBuffer) {
		copy(parser.rawBuffer, parser.rawBuffer[parser.rawBufferPos:])
	}
	parser.rawBuffer = parser.rawBuffer[:len(parser.rawBuffer)-parser.rawBufferPos]
	parser.rawBufferPos = 0

	// Call the read handler to fill the buffer.
	sizeRead, err := parser.readHandler(parser, parser.rawBuffer[len(parser.rawBuffer):cap(parser.rawBuffer)])
	parser.rawBuffer = parser.rawBuffer[:len(parser.rawBuffer)+sizeRead]
	if err == io.EOF {
		parser.eof = true
	} else if err != nil {
		return yamlParserSetReaderError(parser, "input error: "+err.Error(), parser.offset, -1)
	}
	return true
}

// Ensure that the buffer contains at least `length` characters.
// Return true on success, false on failure.
//
// The length is supposed to be significantly less that the buffer size.
func yamlParserUpdateBuffer(parser *yamlParserT, length int) bool {
	if parser.readHandler == nil {
		panic("read handler must be set")
	}

	// [Go] This function was changed to guarantee the requested length size at EOF.
	// The fact we need to do this is pretty awful, but the description above implies
	// for that to be the case, and there are tests

	// If the EOF flag is set and the raw buffer is empty, do nothing.
	if parser.eof && parser.rawBufferPos == len(parser.rawBuffer) {
		// [Go] ACTUALLY! Read the documentation of this function above.
		// This is just broken. To return true, we need to have the
		// given length in the buffer. Not doing that means every single
		// check that calls this function to make sure the buffer has a
		// given length is Go) panicking; or C) accessing invalid memory.
		//return true
	}

	// Return if the buffer contains enough characters.
	if parser.unread >= length {
		return true
	}

	// Determine the input encoding if it is not known yet.
	if parser.encoding == yamlAnyEncoding {
		if !yamlParserDetermineEncoding(parser) {
			return false
		}
	}

	// Move the unread characters to the beginning of the buffer.
	bufferLen := len(parser.buffer)
	if parser.bufferPos > 0 && parser.bufferPos < bufferLen {
		copy(parser.buffer, parser.buffer[parser.bufferPos:])
		bufferLen -= parser.bufferPos
		parser.bufferPos = 0
	} else if parser.bufferPos == bufferLen {
		bufferLen = 0
		parser.bufferPos = 0
	}

	// Open the whole buffer for writing, and cut it before returning.
	parser.buffer = parser.buffer[:cap(parser.buffer)]

	// Fill the buffer until it has enough characters.
	first := true
	for parser.unread < length {

		// Fill the raw buffer if necessary.
		if !first || parser.rawBufferPos == len(parser.rawBuffer) {
			if !yamlParserUpdateRawBuffer(parser) {
				parser.buffer = parser.buffer[:bufferLen]
				return false
			}
		}
		first = false

		// Decode the raw buffer.
	inner:
		for parser.rawBufferPos != len(parser.rawBuffer) {
			var value rune
			var width int

			rawUnread := len(parser.rawBuffer) - parser.rawBufferPos

			// Decode the next character.
			switch parser.encoding {
			case yamlUtf8Encoding:
				// Decode a UTF-8 character.  Check RFC 3629
				// (http://www.ietf.org/rfc/rfc3629.txt) for more details.
				//
				// The following table (taken from the RFC) is used for
				// decoding.
				//
				//    Char. number range |        UTF-8 octet sequence
				//      (hexadecimal)    |              (binary)
				//   --------------------+------------------------------------
				//   0000 0000-0000 007F | 0xxxxxxx
				//   0000 0080-0000 07FF | 110xxxxx 10xxxxxx
				//   0000 0800-0000 FFFF | 1110xxxx 10xxxxxx 10xxxxxx
				//   0001 0000-0010 FFFF | 11110xxx 10xxxxxx 10xxxxxx 10xxxxxx
				//
				// Additionally, the characters in the range 0xD800-0xDFFF
				// are prohibited as they are reserved for use with UTF-16
				// surrogate pairs.

				// Determine the length of the UTF-8 sequence.
				octet := parser.rawBuffer[parser.rawBufferPos]
				switch {
				case octet&0x80 == 0x00:
					width = 1
				case octet&0xE0 == 0xC0:
					width = 2
				case octet&0xF0 == 0xE0:
					width = 3
				case octet&0xF8 == 0xF0:
					width = 4
				default:
					// The leading octet is invalid.
					return yamlParserSetReaderError(parser,
						"invalid leading UTF-8 octet",
						parser.offset, int(octet))
				}

				// Check if the raw buffer contains an incomplete character.
				if width > rawUnread {
					if parser.eof {
						return yamlParserSetReaderError(parser,
							"incomplete UTF-8 octet sequence",
							parser.offset, -1)
					}
					break inner
				}

				// Decode the leading octet.
				switch {
				case octet&0x80 == 0x00:
					value = rune(octet & 0x7F)
				case octet&0xE0 == 0xC0:
					value = rune(octet & 0x1F)
				case octet&0xF0 == 0xE0:
					value = rune(octet & 0x0F)
				case octet&0xF8 == 0xF0:
					value = rune(octet & 0x07)
				default:
					value = 0
				}

				// Check and decode the trailing octets.
				for k := 1; k < width; k++ {
					octet = parser.rawBuffer[parser.rawBufferPos+k]

					// Check if the octet is valid.
					if (octet & 0xC0) != 0x80 {
						return yamlParserSetReaderError(parser,
							"invalid trailing UTF-8 octet",
							parser.offset+k, int(octet))
					}

					// Decode the octet.
					value = (value << 6) + rune(octet&0x3F)
				}

				// Check the length of the sequence against the value.
				switch {
				case width == 1:
				case width == 2 && value >= 0x80:
				case width == 3 && value >= 0x800:
				case width == 4 && value >= 0x10000:
				default:
					return yamlParserSetReaderError(parser,
						"invalid length of a UTF-8 sequence",
						parser.offset, -1)
				}

				// Check the range of the value.
				if value >= 0xD800 && value <= 0xDFFF || value > 0x10FFFF {
					return yamlParserSetReaderError(parser,
						"invalid Unicode character",
						parser.offset, int(value))
				}

			case yamlUtf16leEncoding, yamlUtf16beEncoding:
				var low, high int
				if parser.encoding == yamlUtf16leEncoding {
					low, high = 0, 1
				} else {
					low, high = 1, 0
				}

				// The UTF-16 encoding is not as simple as one might
				// naively think.  Check RFC 2781
				// (http://www.ietf.org/rfc/rfc2781.txt).
				//
				// Normally, two subsequent bytes describe a Unicode
				// character.  However a special technique (called a
				// surrogate pair) is used for specifying character
				// values larger than 0xFFFF.
				//
				// A surrogate pair consists of two pseudo-characters:
				//      high surrogate area (0xD800-0xDBFF)
				//      low surrogate area (0xDC00-0xDFFF)
				//
				// The following formulas are used for decoding
				// and encoding characters using surrogate pairs:
				//
				//  U  = U' + 0x10000   (0x01 00 00 <= U <= 0x10 FF FF)
				//  U' = yyyyyyyyyyxxxxxxxxxx   (0 <= U' <= 0x0F FF FF)
				//  W1 = 110110yyyyyyyyyy
				//  W2 = 110111xxxxxxxxxx
				//
				// where U is the character value, W1 is the high surrogate
				// area, W2 is the low surrogate area.

				// Check for incomplete UTF-16 character.
				if rawUnread < 2 {
					if parser.eof {
						return yamlParserSetReaderError(parser,
							"incomplete UTF-16 character",
							parser.offset, -1)
					}
					break inner
				}

				// Get the character.
				value = rune(parser.rawBuffer[parser.rawBufferPos+low]) +
					(rune(parser.rawBuffer[parser.rawBufferPos+high]) << 8)

				// Check for unexpected low surrogate area.
				if value&0xFC00 == 0xDC00 {
					return yamlParserSetReaderError(parser,
						"unexpected low surrogate area",
						parser.offset, int(value))
				}

				// Check for a high surrogate area.
				if value&0xFC00 == 0xD800 {
					width = 4

					// Check for incomplete surrogate pair.
					if rawUnread < 4 {
						if parser.eof {
							return yamlParserSetReaderError(parser,
								"incomplete UTF-16 surrogate pair",
								parser.offset, -1)
						}
						break inner
					}

					// Get the next character.
					value2 := rune(parser.rawBuffer[parser.rawBufferPos+low+2]) +
						(rune(parser.rawBuffer[parser.rawBufferPos+high+2]) << 8)

					// Check for a low surrogate area.
					if value2&0xFC00 != 0xDC00 {
						return yamlParserSetReaderError(parser,
							"expected low surrogate area",
							parser.offset+2, int(value2))
					}

					// Generate the value of the surrogate pair.
					value = 0x10000 + ((value & 0x3FF) << 10) + (value2 & 0x3FF)
				} else {
					width = 2
				}

			default:
				panic("impossible")
			}

			// Check if the character is in the allowed range:
			//      #x9 | #xA | #xD | [#x20-#x7E]               (8 bit)
			//      | #x85 | [#xA0-#xD7FF] | [#xE000-#xFFFD]    (16 bit)
			//      | [#x10000-#x10FFFF]                        (32 bit)
			switch {
			case value == 0x09:
			case value == 0x0A:
			case value == 0x0D:
			case value >= 0x20 && value <= 0x7E:
			case value == 0x85:
			case value >= 0xA0 && value <= 0xD7FF:
			case value >= 0xE000 && value <= 0xFFFD:
			case value >= 0x10000 && value <= 0x10FFFF:
			default:
				return yamlParserSetReaderError(parser,
					"control characters are not allowed",
					parser.offset, int(value))
			}

			// Move the raw pointers.
			parser.rawBufferPos += width
			parser.offset += width

			// Finally put the character into the buffer.
			if value <= 0x7F {
				// 0000 0000-0000 007F . 0xxxxxxx
				parser.buffer[bufferLen+0] = byte(value)
				bufferLen++
			} else if value <= 0x7FF {
				// 0000 0080-0000 07FF . 110xxxxx 10xxxxxx
				parser.buffer[bufferLen+0] = byte(0xC0 + (value >> 6))
				parser.buffer[bufferLen+1] = byte(0x80 + (value & 0x3F))
				bufferLen += 2
			} else if value <= 0xFFFF {
				// 0000 0800-0000 FFFF . 1110xxxx 10xxxxxx 10xxxxxx
				parser.buffer[bufferLen+0] = byte(0xE0 + (value >> 12))
				parser.buffer[bufferLen+1] = byte(0x80 + ((value >> 6) & 0x3F))
				parser.buffer[bufferLen+2] = byte(0x80 + (value & 0x3F))
				bufferLen += 3
			} else {
				// 0001 0000-0010 FFFF . 11110xxx 10xxxxxx 10xxxxxx 10xxxxxx
				parser.buffer[bufferLen+0] = byte(0xF0 + (value >> 18))
				parser.buffer[bufferLen+1] = byte(0x80 + ((value >> 12) & 0x3F))
				parser.buffer[bufferLen+2] = byte(0x80 + ((value >> 6) & 0x3F))
				parser.buffer[bufferLen+3] = byte(0x80 + (value & 0x3F))
				bufferLen += 4
			}

			parser.unread++
		}

		// On EOF, put NUL into the buffer and return.
		if parser.eof {
			parser.buffer[bufferLen] = 0
			bufferLen++
			parser.unread++
			break
		}
	}
	// [Go] Read the documentation of this function above. To return true,
	// we need to have the given length in the buffer. Not doing that means
	// every single check that calls this function to make sure the buffer
	// has a given length is Go) panicking; or C) accessing invalid memory.
	// This happens here due to the EOF above breaking early.
	for bufferLen < length {
		parser.buffer[bufferLen] = 0
		bufferLen++
	}
	parser.buffer = parser.buffer[:bufferLen]
	return true
}
