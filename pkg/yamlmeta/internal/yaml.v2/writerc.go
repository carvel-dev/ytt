// Copyright 2020 VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package yaml

// Set the writer error and return false.
func yamlEmitterSetWriterError(emitter *yamlEmitterT, problem string) bool {
	emitter.error = yamlWriterError
	emitter.problem = problem
	return false
}

// Flush the output buffer.
func yamlEmitterFlush(emitter *yamlEmitterT) bool {
	if emitter.writeHandler == nil {
		panic("write handler not set")
	}

	// Check if the buffer is empty.
	if emitter.bufferPos == 0 {
		return true
	}

	if err := emitter.writeHandler(emitter, emitter.buffer[:emitter.bufferPos]); err != nil {
		return yamlEmitterSetWriterError(emitter, "write error: "+err.Error())
	}
	emitter.bufferPos = 0
	return true
}
