// Copyright 2026 The Carvel Authors.
// SPDX-License-Identifier: Apache-2.0

package yaml

import (
	"fmt"
	"reflect"
	"testing"
)

// TestEmitterEventQueueIsRewound checks that yamlEmitterEmit resets eventsHead
// back to zero once the event queue drains. That keeps the queue's backing
// array bounded by nesting depth instead of growing one entry per node.
func TestEmitterEventQueueIsRewound(t *testing.T) {
	const keys = 2000

	doc := make(map[string]any, keys)
	for i := range keys {
		doc[fmt.Sprintf("key%05d", i)] = i
	}

	e := newEncoder()
	defer e.destroy()

	e.marshalDoc("", reflect.ValueOf(doc))

	if e.emitter.eventsHead != 0 {
		t.Errorf("eventsHead = %d, want 0", e.emitter.eventsHead)
	}
	// Without the rewind this grows to roughly 2*keys.
	if c := cap(e.emitter.events); c > 32 {
		t.Errorf("cap(emitter.events) = %d after emitting %d keys, want <= 32", c, keys)
	}
}
