/*
Copyright (c) 2024 Seldon Technologies Ltd.

Use of this software is governed by
(1) the license included in the LICENSE file or
(2) if the license included in the LICENSE file is the Business Source License 1.1,
the Change License after the Change Date as each is defined in accordance with the LICENSE file.
*/

package schema

import (
	"bytes"
	"testing"
)

func TestTrimSchemaID(t *testing.T) {
	tests := []struct {
		name     string
		payload  []byte
		expected []byte
	}{
		{
			name:     "schema registry format with magic byte",
			payload:  []byte{0x0, 0x0, 0x0, 0x0, 0x1, 0x0, 0x48, 0x65, 0x6c, 0x6c, 0x6f},
			expected: []byte{0x48, 0x65, 0x6c, 0x6c, 0x6f},
		},
		{
			name:     "non-schema registry format",
			payload:  []byte{0x1, 0x2, 0x3, 0x4, 0x5},
			expected: []byte{0x1, 0x2, 0x3, 0x4, 0x5},
		},
		{
			name:     "empty payload with magic byte",
			payload:  []byte{0x0, 0x0, 0x0, 0x0, 0x1, 0x0},
			expected: []byte{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := TrimSchemaID(tt.payload)
			if !bytes.Equal(result, tt.expected) {
				t.Errorf("TrimSchemaID() = %v, want %v", result, tt.expected)
			}
		})
	}
}
