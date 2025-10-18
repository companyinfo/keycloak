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
)

// mapper converts a struct to a map[string]string for use as query parameters.
// The struct fields must have json tags with "omitempty" for proper serialization.
// Note: Fields with `json:"name,string,omitempty"` will have quotes in values.
// mapper converts a struct to a map[string]string, suitable for query parameters.
//
// It marshals the struct to JSON, then unmarshals into a generic map, converting all values
// to their string representations. Fields with the `omitempty` tag will be omitted if empty.
//
// Note: This does NOT recursively flatten nested structs or handle slices/maps other than basic stringification.
//
//	Use only for flat structs intended for query encoding.
func mapper(s any) (map[string]string, error) {
	b, err := json.Marshal(s)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal struct: %w", err)
	}

	var generic map[string]any
	if err := json.Unmarshal(b, &generic); err != nil {
		return nil, fmt.Errorf("failed to unmarshal json to map: %w", err)
	}

	result := make(map[string]string, len(generic))
	for k, v := range generic {
		// Defensive: avoid "<nil>" string by explicit nil check, though JSON shouldn't produce nils here.
		if v == nil {
			result[k] = ""
			continue
		}
		result[k] = fmt.Sprintf("%v", v)
	}
	return result, nil
}
