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

func TestMakeURL(t *testing.T) {
	tests := []struct {
		name     string
		path     []string
		expected string
	}{
		{
			name:     "single segment",
			path:     []string{"api"},
			expected: "api",
		},
		{
			name:     "multiple segments",
			path:     []string{"api", "v1", "users"},
			expected: "api/v1/users",
		},
		{
			name:     "empty segments",
			path:     []string{"", "api", "", "users"},
			expected: "/api//users",
		},
		{
			name:     "no segments",
			path:     []string{},
			expected: "",
		},
		{
			name:     "single empty segment",
			path:     []string{""},
			expected: "",
		},
		{
			name:     "segments with special characters",
			path:     []string{"api", "users", "123", "profile"},
			expected: "api/users/123/profile",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := makeURL(tt.path...)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestStringP(t *testing.T) {
	tests := []struct {
		name  string
		value string
	}{
		{
			name:  "non-empty string",
			value: "test",
		},
		{
			name:  "empty string",
			value: "",
		},
		{
			name:  "string with spaces",
			value: "hello world",
		},
		{
			name:  "string with special characters",
			value: "test@example.com",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := StringP(tt.value)
			require.NotNil(t, result)
			assert.Equal(t, tt.value, *result)
		})
	}
}

func TestPString(t *testing.T) {
	tests := []struct {
		name     string
		value    *string
		expected string
	}{
		{
			name:     "non-nil pointer",
			value:    StringP("test"),
			expected: "test",
		},
		{
			name:     "nil pointer",
			value:    nil,
			expected: "",
		},
		{
			name:     "pointer to empty string",
			value:    StringP(""),
			expected: "",
		},
		{
			name:     "pointer to string with spaces",
			value:    StringP("hello world"),
			expected: "hello world",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := PString(tt.value)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestIntP(t *testing.T) {
	tests := []struct {
		name  string
		value int
	}{
		{
			name:  "positive integer",
			value: 42,
		},
		{
			name:  "negative integer",
			value: -10,
		},
		{
			name:  "zero",
			value: 0,
		},
		{
			name:  "large integer",
			value: 999999,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IntP(tt.value)
			require.NotNil(t, result)
			assert.Equal(t, tt.value, *result)
		})
	}
}

func TestInt32P(t *testing.T) {
	tests := []struct {
		name  string
		value int32
	}{
		{
			name:  "positive int32",
			value: 42,
		},
		{
			name:  "negative int32",
			value: -10,
		},
		{
			name:  "zero",
			value: 0,
		},
		{
			name:  "max int32",
			value: 2147483647,
		},
		{
			name:  "min int32",
			value: -2147483648,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Int32P(tt.value)
			require.NotNil(t, result)
			assert.Equal(t, tt.value, *result)
		})
	}
}

func TestInt64P(t *testing.T) {
	tests := []struct {
		name  string
		value int64
	}{
		{
			name:  "positive int64",
			value: 42,
		},
		{
			name:  "negative int64",
			value: -10,
		},
		{
			name:  "zero",
			value: 0,
		},
		{
			name:  "large int64",
			value: 9223372036854775807,
		},
		{
			name:  "small int64",
			value: -9223372036854775808,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Int64P(tt.value)
			require.NotNil(t, result)
			assert.Equal(t, tt.value, *result)
		})
	}
}

func TestBoolP(t *testing.T) {
	tests := []struct {
		name  string
		value bool
	}{
		{
			name:  "true",
			value: true,
		},
		{
			name:  "false",
			value: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := BoolP(tt.value)
			require.NotNil(t, result)
			assert.Equal(t, tt.value, *result)
		})
	}
}

func TestNilOrEmpty(t *testing.T) {
	tests := []struct {
		name     string
		value    *string
		expected bool
	}{
		{
			name:     "nil pointer",
			value:    nil,
			expected: true,
		},
		{
			name:     "empty string",
			value:    StringP(""),
			expected: true,
		},
		{
			name:     "non-empty string",
			value:    StringP("test"),
			expected: false,
		},
		{
			name:     "whitespace string",
			value:    StringP(" "),
			expected: false,
		},
		{
			name:     "string with multiple characters",
			value:    StringP("hello world"),
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := NilOrEmpty(tt.value)
			assert.Equal(t, tt.expected, result)
		})
	}
}

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
