// Copyright 2025 Company.info B.V.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package keycloak

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMapper(t *testing.T) {
	type TestStruct struct {
		Name   string `json:"name,omitempty"`
		Age    int    `json:"age,omitempty"`
		Active bool   `json:"active,omitempty"`
		Score  int64  `json:"score,omitempty"`
	}

	type NestedStruct struct {
		ID     string `json:"id,omitempty"`
		Nested struct {
			Value string `json:"value,omitempty"`
		} `json:"nested,omitempty"`
	}

	tests := []struct {
		name      string
		input     any
		expected  map[string]string
		expectErr bool
	}{
		{
			name: "simple struct with all fields",
			input: TestStruct{
				Name:   "John",
				Age:    30,
				Active: true,
				Score:  100,
			},
			expected: map[string]string{
				"name":   "John",
				"age":    "30",
				"active": "true",
				"score":  "100",
			},
			expectErr: false,
		},
		{
			name: "struct with some empty fields",
			input: TestStruct{
				Name: "Jane",
				Age:  25,
			},
			expected: map[string]string{
				"name": "Jane",
				"age":  "25",
			},
			expectErr: false,
		},
		{
			name:      "empty struct",
			input:     TestStruct{},
			expected:  map[string]string{},
			expectErr: false,
		},
		{
			name: "struct with nested object",
			input: NestedStruct{
				ID: "123",
			},
			expected: map[string]string{
				"id":     "123",
				"nested": "map[]",
			},
			expectErr: false,
		},
		{
			name: "struct with zero values",
			input: TestStruct{
				Name:   "",
				Age:    0,
				Active: false,
				Score:  0,
			},
			expected:  map[string]string{},
			expectErr: false,
		},
		{
			name: "struct with negative numbers",
			input: TestStruct{
				Name:  "Test",
				Age:   -1,
				Score: -100,
			},
			expected: map[string]string{
				"name":  "Test",
				"age":   "-1",
				"score": "-100",
			},
			expectErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := mapper(tt.input)
			if tt.expectErr {
				assert.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestMapperWithUnmarshallableType(t *testing.T) {
	type UnmarshallableStruct struct {
		Channel chan int
	}

	// Channels cannot be marshalled to JSON
	result, err := mapper(UnmarshallableStruct{Channel: make(chan int)})
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "failed to marshal struct")
}
