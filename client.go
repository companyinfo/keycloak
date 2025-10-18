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

// Package keycloak provides an idiomatic Go client for the Keycloak Admin REST API.
package keycloak

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/go-resty/resty/v2"
	"golang.org/x/oauth2/clientcredentials"
)

const defaultSize = 50

// Client is the main entry point for the Keycloak Admin API.
// It provides access to resource-specific clients for managing different Keycloak resources.
//
// Example usage:
//
//	client, err := keycloak.New(ctx, config,
//	    keycloak.WithPageSize(100),
//	    keycloak.WithTimeout(30*time.Second),
//	)
//	if err != nil {
//	    return err
//	}
//
//	// Access Groups resource
//	groupID, err := client.Groups.Create(ctx, "My Group", attributes)
//	groups, err := client.Groups.List(ctx, nil, false)
type Client struct {
	// Groups provides access to group management operations
	Groups GroupsClient

	// Internal shared state
	resty    *resty.Client
	config   Config
	baseURL  string
	realm    string
	pageSize int
}

// Config contains the required configuration for creating a Keycloak client.
// Only required fields are included; optional configuration uses functional options.
type Config struct {
	URL          string // Base URL of the Keycloak server (required, e.g., https://keycloak.example.com)
	Realm        string // Keycloak realm name (required)
	ClientID     string // OAuth2 client ID (required)
	ClientSecret string // OAuth2 client secret (required)
}

// Option is a functional option for configuring the Client.
type Option func(*Client) error

// WithPageSize sets the default page size for paginated requests.
// Default is 50 if not specified.
//
// Example:
//
//	client, err := keycloak.New(ctx, config, keycloak.WithPageSize(100))
func WithPageSize(size int) Option {
	return func(c *Client) error {
		if size <= 0 {
			return fmt.Errorf("page size must be positive, got %d", size)
		}
		c.pageSize = size
		return nil
	}
}

// WithHTTPClient sets a custom HTTP client for the underlying transport.
// This is useful for custom timeouts, proxies, or TLS configuration.
// Note: This will override the OAuth2 client, so you need to handle authentication separately.
//
// Example:
//
//	httpClient := &http.Client{
//	    Transport: &http.Transport{
//	        TLSClientConfig: &tls.Config{InsecureSkipVerify: false},
//	    },
//	    Timeout: 30 * time.Second,
//	}
//	client, err := keycloak.New(ctx, config, keycloak.WithHTTPClient(httpClient))
func WithHTTPClient(httpClient *http.Client) Option {
	return func(c *Client) error {
		if httpClient == nil {
			return fmt.Errorf("http client cannot be nil")
		}
		c.resty = resty.NewWithClient(httpClient)
		return nil
	}
}

// WithTimeout sets the request timeout for all API calls.
// Default is no timeout if not specified.
//
// Example:
//
//	client, err := keycloak.New(ctx, config, keycloak.WithTimeout(30*time.Second))
func WithTimeout(timeout time.Duration) Option {
	return func(c *Client) error {
		if timeout < 0 {
			return fmt.Errorf("timeout must be non-negative, got %v", timeout)
		}
		c.resty.SetTimeout(timeout)
		return nil
	}
}

// WithRetry configures retry behavior for failed requests.
//
// Example:
//
//	client, err := keycloak.New(ctx, config,
//	    keycloak.WithRetry(3, 5*time.Second, 30*time.Second),
//	)
func WithRetry(count int, waitTime, maxWaitTime time.Duration) Option {
	return func(c *Client) error {
		if count < 0 {
			return fmt.Errorf("retry count must be non-negative, got %d", count)
		}
		c.resty.
			SetRetryCount(count).
			SetRetryWaitTime(waitTime).
			SetRetryMaxWaitTime(maxWaitTime)
		return nil
	}
}

// WithDebug enables debug mode, logging all requests and responses.
//
// Example:
//
//	client, err := keycloak.New(ctx, config, keycloak.WithDebug(true))
func WithDebug(debug bool) Option {
	return func(c *Client) error {
		c.resty.SetDebug(debug)
		return nil
	}
}

// WithHeaders adds custom headers to all requests.
//
// Example:
//
//	client, err := keycloak.New(ctx, config,
//	    keycloak.WithHeaders(map[string]string{
//	        "X-Request-ID": requestID,
//	    }),
//	)
func WithHeaders(headers map[string]string) Option {
	return func(c *Client) error {
		c.resty.SetHeaders(headers)
		return nil
	}
}

// WithUserAgent sets a custom User-Agent header.
//
// Example:
//
//	client, err := keycloak.New(ctx, config,
//	    keycloak.WithUserAgent("my-app/1.0"),
//	)
func WithUserAgent(userAgent string) Option {
	return func(c *Client) error {
		c.resty.SetHeader("User-Agent", userAgent)
		return nil
	}
}

// WithProxy sets a proxy URL for all requests.
//
// Example:
//
//	client, err := keycloak.New(ctx, config,
//	    keycloak.WithProxy("http://proxy.example.com:8080"),
//	)
func WithProxy(proxyURL string) Option {
	return func(c *Client) error {
		c.resty.SetProxy(proxyURL)
		return nil
	}
}

// New creates a new Keycloak client with the provided configuration and options.
// It establishes OAuth2 authentication using the client credentials flow
// and returns a ready-to-use client.
//
// The client automatically manages token refresh and includes the access token
// in all API requests.
//
// Example:
//
//	client, err := keycloak.New(ctx,
//	    keycloak.Config{
//	        URL:          "https://keycloak.example.com",
//	        Realm:        "my-realm",
//	        ClientID:     "admin-cli",
//	        ClientSecret: "secret",
//	    },
//	    keycloak.WithPageSize(100),
//	    keycloak.WithTimeout(30*time.Second),
//	    keycloak.WithRetry(3, 5*time.Second, 30*time.Second),
//	)
//	if err != nil {
//	    return fmt.Errorf("failed to create client: %w", err)
//	}
func New(ctx context.Context, config Config, opts ...Option) (*Client, error) {
	// Validate required config
	if config.URL == "" {
		return nil, fmt.Errorf("URL is required")
	}
	if config.Realm == "" {
		return nil, fmt.Errorf("realm is required")
	}
	if config.ClientID == "" {
		return nil, fmt.Errorf("clientID is required")
	}
	if config.ClientSecret == "" {
		return nil, fmt.Errorf("clientSecret is required")
	}

	authAdminRealms := "admin/realms"
	authRealms := "realms"
	realmURL, err := url.JoinPath(config.URL, authRealms, config.Realm)
	if err != nil {
		return nil, fmt.Errorf("invalid base URL: %w", err)
	}

	oidcProvider, err := oidc.NewProvider(ctx, realmURL)
	if err != nil {
		return nil, fmt.Errorf("login failed: %w", err)
	}

	oauthClient := clientcredentials.Config{
		ClientID:     config.ClientID,
		ClientSecret: config.ClientSecret,
		TokenURL:     oidcProvider.Endpoint().TokenURL,
	}

	// Initialize client with defaults
	client := &Client{
		resty:    resty.NewWithClient(oauthClient.Client(ctx)),
		config:   config,
		baseURL:  config.URL,
		realm:    config.Realm,
		pageSize: defaultSize, // default, can be overridden by options
	}

	// Apply functional options
	for _, opt := range opts {
		if err := opt(client); err != nil {
			return nil, fmt.Errorf("failed to apply option: %w", err)
		}
	}

	// Initialize resource clients (after all options applied)
	client.Groups = newGroupsClient(client, authAdminRealms)

	return client, nil
}
