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
	"strings"
)

// HTTPErrorResponse represents an error response from the Keycloak API.
// It captures error details from failed HTTP requests.
type HTTPErrorResponse struct {
	Error       string `json:"error,omitempty"`             // Error type or code
	Message     string `json:"errorMessage,omitempty"`      // Human-readable error message
	Description string `json:"error_description,omitempty"` // Detailed error description
}

// Empty returns true if the error response contains no error information.
func (e HTTPErrorResponse) Empty() bool {
	return len(e.Error) <= 0 || len(e.Message) <= 0 || len(e.Description) <= 0
}

// String returns a formatted string representation of the error.
// It concatenates all available error fields with proper formatting.
func (e HTTPErrorResponse) String() string {
	var res strings.Builder
	if len(e.Error) > 0 {
		res.WriteString(e.Error)
	}
	if len(e.Message) > 0 {
		if res.Len() > 0 {
			res.WriteString(": ")
		}
		res.WriteString(e.Message)
	}
	if len(e.Description) > 0 {
		if res.Len() > 0 {
			res.WriteString(": ")
		}
		res.WriteString(e.Description)
	}
	return res.String()
}
