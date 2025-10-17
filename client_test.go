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

// Package keycloak provides tests for client configuration and functional options.
// Tests cover option validation, client initialization, and configuration defaults.
package keycloak

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/stretchr/testify/assert"
)

func TestWithPageSize(t *testing.T) {
	tests := []struct {
		name      string
		size      int
		wantErr   bool
		wantValue int
	}{
		{
			name:      "valid page size",
			size:      100,
			wantErr:   false,
			wantValue: 100,
		},
		{
			name:    "zero page size",
			size:    0,
			wantErr: true,
		},
		{
			name:    "negative page size",
			size:    -1,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := &Client{pageSize: defaultSize}
			err := WithPageSize(tt.size)(client)

			if (err != nil) != tt.wantErr {
				t.Errorf("WithPageSize() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && client.pageSize != tt.wantValue {
				t.Errorf("WithPageSize() pageSize = %v, want %v", client.pageSize, tt.wantValue)
			}
		})
	}
}

func TestWithTimeout(t *testing.T) {
	tests := []struct {
		name    string
		timeout time.Duration
		wantErr bool
	}{
		{
			name:    "valid timeout",
			timeout: 30 * time.Second,
			wantErr: false,
		},
		{
			name:    "zero timeout",
			timeout: 0,
			wantErr: false,
		},
		{
			name:    "negative timeout",
			timeout: -1 * time.Second,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := &Client{resty: newTestRestyClient()}
			err := WithTimeout(tt.timeout)(client)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				// Verify timeout was actually set
				actualTimeout := client.resty.GetClient().Timeout
				assert.Equal(t, tt.timeout, actualTimeout, "timeout should be set correctly")
			}
		})
	}
}

func TestWithRetry(t *testing.T) {
	tests := []struct {
		name        string
		count       int
		waitTime    time.Duration
		maxWaitTime time.Duration
		wantErr     bool
	}{
		{
			name:        "valid retry config",
			count:       3,
			waitTime:    5 * time.Second,
			maxWaitTime: 30 * time.Second,
			wantErr:     false,
		},
		{
			name:        "zero retries",
			count:       0,
			waitTime:    5 * time.Second,
			maxWaitTime: 30 * time.Second,
			wantErr:     false,
		},
		{
			name:        "negative retries",
			count:       -1,
			waitTime:    5 * time.Second,
			maxWaitTime: 30 * time.Second,
			wantErr:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := &Client{resty: newTestRestyClient()}
			err := WithRetry(tt.count, tt.waitTime, tt.maxWaitTime)(client)

			if (err != nil) != tt.wantErr {
				t.Errorf("WithRetry() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestWithDebug(t *testing.T) {
	tests := []struct {
		name  string
		debug bool
	}{
		{
			name:  "enable debug",
			debug: true,
		},
		{
			name:  "disable debug",
			debug: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := &Client{resty: newTestRestyClient()}
			err := WithDebug(tt.debug)(client)
			assert.NoError(t, err)
			assert.Equal(t, tt.debug, client.resty.Debug, "debug mode should be set correctly")
		})
	}
}

func TestWithHeaders(t *testing.T) {
	headers := map[string]string{
		"X-Request-ID": "12345",
		"X-Custom":     "value",
	}

	client := &Client{resty: newTestRestyClient()}
	err := WithHeaders(headers)(client)
	assert.NoError(t, err)

	// Verify headers were actually set
	for key, expectedValue := range headers {
		actualValue := client.resty.Header.Get(key)
		assert.Equal(t, expectedValue, actualValue, "header %s should be set correctly", key)
	}
}

func TestWithUserAgent(t *testing.T) {
	userAgent := "my-app/1.0"
	client := &Client{resty: newTestRestyClient()}
	err := WithUserAgent(userAgent)(client)
	assert.NoError(t, err)

	// Verify user agent was set
	actualUserAgent := client.resty.Header.Get("User-Agent")
	assert.Equal(t, userAgent, actualUserAgent, "user agent should be set correctly")
}

func TestWithProxy(t *testing.T) {
	proxyURL := "http://proxy.example.com:8080"
	client := &Client{resty: newTestRestyClient()}
	err := WithProxy(proxyURL)(client)
	assert.NoError(t, err)

	// Note: Proxy is set on transport, difficult to verify without making actual request
	// At least verify no error occurred
}

func TestWithHTTPClient(t *testing.T) {
	tests := []struct {
		name       string
		httpClient *http.Client
		wantErr    bool
	}{
		{
			name:       "valid http client",
			httpClient: &http.Client{},
			wantErr:    false,
		},
		{
			name:       "nil http client",
			httpClient: nil,
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := &Client{resty: newTestRestyClient()}
			err := WithHTTPClient(tt.httpClient)(client)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				// Verify HTTP client was set
				assert.Equal(t, tt.httpClient, client.resty.GetClient())
			}
		})
	}
}

func TestNew(t *testing.T) {
	tests := []struct {
		name    string
		config  Config
		options []Option
		wantErr bool
	}{
		{
			name: "missing URL",
			config: Config{
				Realm:        "test-realm",
				ClientID:     "test-client",
				ClientSecret: "test-secret",
			},
			wantErr: true,
		},
		{
			name: "missing realm",
			config: Config{
				URL:          "https://keycloak.example.com",
				ClientID:     "test-client",
				ClientSecret: "test-secret",
			},
			wantErr: true,
		},
		{
			name: "missing client ID",
			config: Config{
				URL:          "https://keycloak.example.com",
				Realm:        "test-realm",
				ClientSecret: "test-secret",
			},
			wantErr: true,
		},
		{
			name: "missing client secret",
			config: Config{
				URL:      "https://keycloak.example.com",
				Realm:    "test-realm",
				ClientID: "test-client",
			},
			wantErr: true,
		},
		{
			name: "invalid page size option",
			config: Config{
				URL:          "https://keycloak.example.com",
				Realm:        "test-realm",
				ClientID:     "test-client",
				ClientSecret: "test-secret",
			},
			options: []Option{WithPageSize(-1)},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			client, err := New(ctx, tt.config, tt.options...)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, client)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, client)
				assert.NotNil(t, client.Groups)
			}
		})
	}
}

// newTestRestyClient creates a basic resty client for testing
func newTestRestyClient() *resty.Client {
	return resty.New()
}
