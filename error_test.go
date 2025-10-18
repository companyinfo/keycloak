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
)

func TestHTTPErrorResponse_Empty(t *testing.T) {
	tests := []struct {
		name     string
		err      HTTPErrorResponse
		expected bool
	}{
		{
			name:     "completely empty error",
			err:      HTTPErrorResponse{},
			expected: true,
		},
		{
			name: "error with only Error field",
			err: HTTPErrorResponse{
				Error: "invalid_request",
			},
			expected: false,
		},
		{
			name: "error with only Message field",
			err: HTTPErrorResponse{
				Message: "Invalid request",
			},
			expected: false,
		},
		{
			name: "error with only Description field",
			err: HTTPErrorResponse{
				Description: "The request is invalid",
			},
			expected: false,
		},
		{
			name: "error with all fields",
			err: HTTPErrorResponse{
				Error:       "invalid_request",
				Message:     "Invalid request",
				Description: "The request is invalid",
			},
			expected: false,
		},
		{
			name: "error with Error and Message",
			err: HTTPErrorResponse{
				Error:   "invalid_request",
				Message: "Invalid request",
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.err.Empty()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestHTTPErrorResponse_String(t *testing.T) {
	tests := []struct {
		name     string
		err      HTTPErrorResponse
		expected string
	}{
		{
			name:     "completely empty error",
			err:      HTTPErrorResponse{},
			expected: "",
		},
		{
			name: "error with only Error field",
			err: HTTPErrorResponse{
				Error: "invalid_request",
			},
			expected: "invalid_request",
		},
		{
			name: "error with only Message field",
			err: HTTPErrorResponse{
				Message: "Invalid request",
			},
			expected: "Invalid request",
		},
		{
			name: "error with only Description field",
			err: HTTPErrorResponse{
				Description: "The request is invalid",
			},
			expected: "The request is invalid",
		},
		{
			name: "error with Error and Message",
			err: HTTPErrorResponse{
				Error:   "invalid_request",
				Message: "Invalid request",
			},
			expected: "invalid_request: Invalid request",
		},
		{
			name: "error with Error and Description",
			err: HTTPErrorResponse{
				Error:       "invalid_request",
				Description: "The request is invalid",
			},
			expected: "invalid_request: The request is invalid",
		},
		{
			name: "error with Message and Description",
			err: HTTPErrorResponse{
				Message:     "Invalid request",
				Description: "The request is invalid",
			},
			expected: "Invalid request: The request is invalid",
		},
		{
			name: "error with all fields",
			err: HTTPErrorResponse{
				Error:       "invalid_request",
				Message:     "Invalid request",
				Description: "The request is invalid",
			},
			expected: "invalid_request: Invalid request: The request is invalid",
		},
		{
			name: "error with empty strings",
			err: HTTPErrorResponse{
				Error:       "",
				Message:     "",
				Description: "",
			},
			expected: "",
		},
		{
			name: "error with whitespace",
			err: HTTPErrorResponse{
				Error:       "  error  ",
				Message:     "  message  ",
				Description: "  description  ",
			},
			expected: "  error  :   message  :   description  ",
		},
		{
			name: "error with special characters",
			err: HTTPErrorResponse{
				Error:       "invalid_request",
				Message:     "Group 'test-group' not found",
				Description: "The group with ID: 12345 does not exist",
			},
			expected: "invalid_request: Group 'test-group' not found: The group with ID: 12345 does not exist",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.err.String()
			assert.Equal(t, tt.expected, result)
		})
	}
}
