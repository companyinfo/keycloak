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
	"encoding/json"
	"fmt"
	"strings"
)

const (
	urlSeparator string = "/"
)

// makeURL joins path segments with forward slashes to construct URLs.
func makeURL(path ...string) string {
	return strings.Join(path, urlSeparator)
}

// StringP returns a pointer to the given string value.
// Useful for optional string fields that require pointers.
func StringP(value string) *string {
	return &value
}

// PString dereferences a string pointer, returning empty string if nil.
// Safe way to get string value from optional pointer fields.
func PString(value *string) string {
	if value == nil {
		return ""
	}
	return *value
}

// IntP returns a pointer to the given int value.
// Useful for optional int fields that require pointers.
func IntP(value int) *int {
	return &value
}

// Int32P returns a pointer to the given int32 value.
// Useful for optional int32 fields that require pointers.
func Int32P(value int32) *int32 {
	return &value
}

// Int64P returns a pointer to the given int64 value.
// Useful for optional int64 fields that require pointers.
func Int64P(value int64) *int64 {
	return &value
}

// BoolP returns a pointer to the given bool value.
// Useful for optional bool fields that require pointers.
func BoolP(value bool) *bool {
	return &value
}

// NilOrEmpty returns true if the string pointer is nil or points to an empty string.
func NilOrEmpty(value *string) bool {
	return value == nil || len(*value) == 0
}

// mapper converts a struct to a map[string]string for use as query parameters.
// The struct fields must have json tags with "omitempty" for proper serialization.
// Note: Fields with `json:"name,string,omitempty"` will have quotes in values.
func mapper(s interface{}) (map[string]string, error) {
	b, err := json.Marshal(s)
	if err != nil {
		return nil, err
	}
	var res map[string]interface{}
	err = json.Unmarshal(b, &res)
	if err != nil {
		return nil, err
	}

	resStr := make(map[string]string, len(res))
	for key, elem := range res {
		resStr[key] = fmt.Sprintf("%v", elem)
	}
	return resStr, nil
}
