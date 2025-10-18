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

// Group represents a Keycloak group with all its properties.
// Groups can contain subgroups (hierarchical structure) and have custom attributes.
// The ID, Name, and Attributes fields are the most commonly used.
// This struct maps to Keycloak's GroupRepresentation.
type Group struct {
	ID            *string              `json:"id,omitempty"`            // Unique identifier for the group
	Name          *string              `json:"name,omitempty"`          // Display name of the group
	Description   *string              `json:"description,omitempty"`   // Description of the group
	Path          *string              `json:"path,omitempty"`          // Full path in the group hierarchy (e.g., /parent/child)
	ParentID      *string              `json:"parentId,omitempty"`      // ID of the parent group (if this is a subgroup)
	SubGroupCount *int64               `json:"subGroupCount,omitempty"` // Count of direct subgroups
	SubGroups     *[]*Group            `json:"subGroups,omitempty"`     // Child groups in the hierarchy
	Attributes    *map[string][]string `json:"attributes,omitempty"`    // Custom key-value attributes (values are arrays)
	Access        *map[string]bool     `json:"access,omitempty"`        // Access permissions for this group
	ClientRoles   *map[string][]string `json:"clientRoles,omitempty"`   // Client-specific roles assigned to the group
	RealmRoles    *[]string            `json:"realmRoles,omitempty"`    // Realm-level roles assigned to the group
}

// GroupAttribute represents a key-value pair for searching groups by attributes.
// Use this to search for groups with specific attribute values.
type GroupAttribute struct {
	Key   string `json:"key"`   // The attribute key to search for
	Value string `json:"value"` // The expected attribute value
}

// SearchGroupParams represents the optional parameters for querying groups.
// All fields are optional; nil/zero values will use Keycloak defaults.
// Used with GET /admin/realms/{realm}/groups endpoint.
type SearchGroupParams struct {
	BriefRepresentation *bool   `json:"briefRepresentation,string,omitempty"` // If true, return groups without detailed attributes (default: true)
	PopulateHierarchy   *bool   `json:"populateHierarchy,string,omitempty"`   // If true, include subgroup hierarchy in response (default: true)
	Exact               *bool   `json:"exact,string,omitempty"`               // If true, search query must match exactly (default: false)
	First               *int    `json:"first,string,omitempty"`               // Offset for pagination (default: null)
	Full                *bool   `json:"full,string,omitempty"`                // If true, return full group representation
	Max                 *int    `json:"max,string,omitempty"`                 // Maximum number of results to return (default: null)
	Q                   *string `json:"q,omitempty"`                          // General query string (default: null)
	Search              *string `json:"search,omitempty"`                     // Search by group name (default: null). SubGroups only returned when search/q is provided
	SubGroupsCount      *bool   `json:"subGroupsCount,string,omitempty"`      // If true, return the count of subgroups for each group (default: true)
}

// CountGroupParams represents the optional parameters for counting groups.
// Used with GET /admin/realms/{realm}/groups/count endpoint.
type CountGroupParams struct {
	Search *string `json:"search,omitempty"` // Filter count by group name search (default: null)
	Top    *bool   `json:"top,omitempty"`    // If true, only count top-level groups (default: false)
}

// CountGroupResponse represents the response from the count groups endpoint.
type CountGroupResponse struct {
	Count int `json:"count"` // Total number of groups matching the query
}

// SubGroupSearchParams represents the optional parameters for querying subgroups.
// These parameters are used with the /groups/{group-id}/children endpoint.
type SubGroupSearchParams struct {
	BriefRepresentation *bool   `json:"briefRepresentation,string,omitempty"` // If true, return brief group representations (default: false)
	Exact               *bool   `json:"exact,string,omitempty"`               // If true, search must match exactly (default: null)
	First               *int    `json:"first,string,omitempty"`               // Pagination offset (default: null)
	Max                 *int    `json:"max,string,omitempty"`                 // Maximum results to return (default: 10)
	Search              *string `json:"search,omitempty"`                     // Search by group name (substring or exact based on 'exact' param) (default: null)
	SubGroupsCount      *bool   `json:"subGroupsCount,string,omitempty"`      // If true, return count of subgroups for each result (default: true)
}

// GroupMembersParams represents the optional parameters for querying group members.
// Used with GET /admin/realms/{realm}/groups/{group-id}/members endpoint.
type GroupMembersParams struct {
	BriefRepresentation *bool `json:"briefRepresentation,string,omitempty"` // If true, return only basic user information (default: null)
	First               *int  `json:"first,string,omitempty"`               // Pagination offset (default: null)
	Max                 *int  `json:"max,string,omitempty"`                 // Maximum results to return (default: 100)
}

// ManagementPermissionReference represents the authorization permissions status for a group.
// Used with /admin/realms/{realm}/groups/{group-id}/management/permissions endpoint.
type ManagementPermissionReference struct {
	Enabled          *bool              `json:"enabled,omitempty"`          // Whether authorization permissions are enabled
	Resource         *string            `json:"resource,omitempty"`         // Resource identifier
	ScopePermissions *map[string]string `json:"scopePermissions,omitempty"` // Scope permissions mapping
}
