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
	"go.companyinfo.dev/ptr"
)

func TestMapperWithPointers(t *testing.T) {
	type PointerStruct struct {
		StringPtr *string `json:"stringPtr,omitempty"`
		IntPtr    *int    `json:"intPtr,omitempty"`
		BoolPtr   *bool   `json:"boolPtr,omitempty"`
	}

	tests := []struct {
		name     string
		input    PointerStruct
		expected map[string]string
	}{
		{
			name: "all pointers set",
			input: PointerStruct{
				StringPtr: ptr.String("test"),
				IntPtr:    ptr.Int(42),
				BoolPtr:   ptr.Bool(true),
			},
			expected: map[string]string{
				"stringPtr": "test",
				"intPtr":    "42",
				"boolPtr":   "true",
			},
		},
		{
			name: "some pointers nil",
			input: PointerStruct{
				StringPtr: ptr.String("test"),
				IntPtr:    nil,
				BoolPtr:   ptr.Bool(false),
			},
			expected: map[string]string{
				"stringPtr": "test",
				"boolPtr":   "false",
			},
		},
		{
			name:     "all pointers nil",
			input:    PointerStruct{},
			expected: map[string]string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := mapper(tt.input)
			require.NoError(t, err)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestMapperWithBooleanStrings(t *testing.T) {
	type BoolStringStruct struct {
		Flag1 *bool `json:"flag1,string,omitempty"`
		Flag2 *bool `json:"flag2,string,omitempty"`
		Flag3 *bool `json:"flag3,omitempty"`
	}

	tests := []struct {
		name  string
		input BoolStringStruct
	}{
		{
			name: "mixed boolean flags",
			input: BoolStringStruct{
				Flag1: ptr.Bool(true),
				Flag2: ptr.Bool(false),
				Flag3: ptr.Bool(true),
			},
		},
		{
			name: "all false",
			input: BoolStringStruct{
				Flag1: ptr.Bool(false),
				Flag2: ptr.Bool(false),
				Flag3: ptr.Bool(false),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := mapper(tt.input)
			require.NoError(t, err)
			assert.NotNil(t, result)
			// Verify boolean values are converted to strings
			for k, v := range result {
				assert.NotEmpty(t, k)
				assert.NotEmpty(t, v)
			}
		})
	}
}

func TestMapperWithIntegerStrings(t *testing.T) {
	type IntStringStruct struct {
		Count1 *int `json:"count1,string,omitempty"`
		Count2 *int `json:"count2,string,omitempty"`
		Count3 *int `json:"count3,omitempty"`
	}

	tests := []struct {
		name  string
		input IntStringStruct
	}{
		{
			name: "mixed integer values",
			input: IntStringStruct{
				Count1: ptr.Int(100),
				Count2: ptr.Int(0),
				Count3: ptr.Int(-50),
			},
		},
		{
			name: "large numbers",
			input: IntStringStruct{
				Count1: ptr.Int(999999),
				Count2: ptr.Int(1),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := mapper(tt.input)
			require.NoError(t, err)
			assert.NotNil(t, result)
		})
	}
}

func TestMapperWithComplexNestedStructs(t *testing.T) {
	type InnerStruct struct {
		Value string `json:"value,omitempty"`
	}

	type OuterStruct struct {
		ID    string      `json:"id,omitempty"`
		Inner InnerStruct `json:"inner,omitempty"`
	}

	tests := []struct {
		name  string
		input OuterStruct
	}{
		{
			name: "nested struct with values",
			input: OuterStruct{
				ID: "outer-1",
				Inner: InnerStruct{
					Value: "inner-value",
				},
			},
		},
		{
			name: "nested struct empty inner",
			input: OuterStruct{
				ID:    "outer-1",
				Inner: InnerStruct{},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := mapper(tt.input)
			require.NoError(t, err)
			assert.NotNil(t, result)
			if tt.input.ID != "" {
				assert.Contains(t, result, "id")
			}
		})
	}
}

func TestMapperWithSlices(t *testing.T) {
	type SliceStruct struct {
		Items []string `json:"items,omitempty"`
		IDs   []int    `json:"ids,omitempty"`
	}

	tests := []struct {
		name  string
		input SliceStruct
	}{
		{
			name: "with slices",
			input: SliceStruct{
				Items: []string{"a", "b", "c"},
				IDs:   []int{1, 2, 3},
			},
		},
		{
			name: "empty slices",
			input: SliceStruct{
				Items: []string{},
				IDs:   []int{},
			},
		},
		{
			name:  "nil slices",
			input: SliceStruct{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := mapper(tt.input)
			require.NoError(t, err)
			assert.NotNil(t, result)
		})
	}
}

func TestMapperWithMaps(t *testing.T) {
	type MapStruct struct {
		Attributes map[string]string `json:"attributes,omitempty"`
		Counts     map[string]int    `json:"counts,omitempty"`
	}

	tests := []struct {
		name  string
		input MapStruct
	}{
		{
			name: "with maps",
			input: MapStruct{
				Attributes: map[string]string{
					"key1": "value1",
					"key2": "value2",
				},
				Counts: map[string]int{
					"count1": 10,
				},
			},
		},
		{
			name: "empty maps",
			input: MapStruct{
				Attributes: map[string]string{},
				Counts:     map[string]int{},
			},
		},
		{
			name:  "nil maps",
			input: MapStruct{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := mapper(tt.input)
			require.NoError(t, err)
			assert.NotNil(t, result)
		})
	}
}

func TestMapperWithSpecialCharacters(t *testing.T) {
	type SpecialStruct struct {
		Name        string `json:"name,omitempty"`
		Description string `json:"description,omitempty"`
		Path        string `json:"path,omitempty"`
	}

	tests := []struct {
		name  string
		input SpecialStruct
	}{
		{
			name: "unicode characters",
			input: SpecialStruct{
				Name:        "Test 测试",
				Description: "Ñoño & friends",
				Path:        "/path/to/资源",
			},
		},
		{
			name: "special characters",
			input: SpecialStruct{
				Name:        "Name with spaces & symbols!",
				Description: "Description with \"quotes\" and 'apostrophes'",
				Path:        "/path/with-dashes_and_underscores",
			},
		},
		{
			name: "newlines and tabs",
			input: SpecialStruct{
				Name:        "Name\nwith\nnewlines",
				Description: "Description\twith\ttabs",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := mapper(tt.input)
			require.NoError(t, err)
			assert.NotNil(t, result)
			// Verify special characters are preserved
			if tt.input.Name != "" {
				assert.Contains(t, result, "name")
			}
		})
	}
}

func TestMapperWithFloats(t *testing.T) {
	type FloatStruct struct {
		Percentage float64 `json:"percentage,omitempty"`
		Score      float32 `json:"score,omitempty"`
	}

	tests := []struct {
		name  string
		input FloatStruct
	}{
		{
			name: "positive floats",
			input: FloatStruct{
				Percentage: 99.99,
				Score:      85.5,
			},
		},
		{
			name: "negative floats",
			input: FloatStruct{
				Percentage: -10.5,
				Score:      -25.3,
			},
		},
		{
			name: "zero values",
			input: FloatStruct{
				Percentage: 0.0,
				Score:      0.0,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := mapper(tt.input)
			require.NoError(t, err)
			assert.NotNil(t, result)
		})
	}
}

func TestMapperErrorCases(t *testing.T) {
	tests := []struct {
		name      string
		input     interface{}
		wantError bool
	}{
		{
			name: "channel - cannot marshal",
			input: struct {
				Ch chan int `json:"ch"`
			}{
				Ch: make(chan int),
			},
			wantError: true,
		},
		{
			name: "function - cannot marshal",
			input: struct {
				Fn func() `json:"fn"`
			}{
				Fn: func() {},
			},
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := mapper(tt.input)
			if tt.wantError {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
			}
		})
	}
}

func TestMapperConsistency(t *testing.T) {
	type TestStruct struct {
		Name  string `json:"name,omitempty"`
		Count int    `json:"count,omitempty"`
		Flag  bool   `json:"flag,omitempty"`
	}

	input := TestStruct{
		Name:  "test",
		Count: 42,
		Flag:  true,
	}

	// Call mapper multiple times with same input
	result1, err1 := mapper(input)
	require.NoError(t, err1)

	result2, err2 := mapper(input)
	require.NoError(t, err2)

	// Results should be identical
	assert.Equal(t, result1, result2)
}
