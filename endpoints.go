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
	"net/http"
	"strings"
)

// endpoint represents a Keycloak Admin API endpoint with its HTTP method and path template.
// Path templates use curly-brace placeholders that match Keycloak's API documentation.
type endpoint struct {
	Method string // HTTP method (GET, POST, PUT, DELETE)
	Path   string // URL path template with {placeholders}
}

// Keycloak Admin API endpoints for Groups resource.
// These endpoints map directly to the official Keycloak Admin REST API.
// See: https://www.keycloak.org/docs-api/latest/rest-api/index.html#_groups
var (
	endpointGroupsList       = endpoint{http.MethodGet, "/admin/realms/{realm}/groups"}
	endpointGroupsCreate     = endpoint{http.MethodPost, "/admin/realms/{realm}/groups"}
	endpointGroupsCount      = endpoint{http.MethodGet, "/admin/realms/{realm}/groups/count"}
	endpointGroupGet         = endpoint{http.MethodGet, "/admin/realms/{realm}/groups/{groupID}"}
	endpointGroupUpdate      = endpoint{http.MethodPut, "/admin/realms/{realm}/groups/{groupID}"}
	endpointGroupDelete      = endpoint{http.MethodDelete, "/admin/realms/{realm}/groups/{groupID}"}
	endpointGroupChildren    = endpoint{http.MethodGet, "/admin/realms/{realm}/groups/{groupID}/children"}
	endpointGroupChildCreate = endpoint{http.MethodPost, "/admin/realms/{realm}/groups/{groupID}/children"}
	endpointGroupMembers     = endpoint{http.MethodGet, "/admin/realms/{realm}/groups/{groupID}/members"}
	endpointGroupPermsGet    = endpoint{http.MethodGet, "/admin/realms/{realm}/groups/{groupID}/management/permissions"}
	endpointGroupPermsUpdate = endpoint{http.MethodPut, "/admin/realms/{realm}/groups/{groupID}/management/permissions"}
)

// buildURL constructs a full URL from an endpoint template by replacing placeholders with actual values.
// The realm is automatically substituted from the client configuration.
// Additional parameters can be provided via the params map using keys that match the placeholder names
// (without curly braces). For example, to replace {groupID}, use params["groupID"].
//
// Example:
//
//	url := c.buildURL(endpointGroupGet, map[string]string{"groupID": "123"})
//	// Returns: https://keycloak.example.com/admin/realms/my-realm/groups/123
func (c *Client) buildURL(ep endpoint, params map[string]string) string {
	path := ep.Path

	// Always replace realm placeholder with client's configured realm
	path = strings.ReplaceAll(path, "{realm}", c.realm)

	// Replace additional placeholders if provided
	for key, value := range params {
		placeholder := "{" + key + "}"
		path = strings.ReplaceAll(path, placeholder, value)
	}

	return c.baseURL + path
}
