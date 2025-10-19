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
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.companyinfo.dev/ptr"
)

func TestFindGroupByAttribute(t *testing.T) {
	tests := []struct {
		name      string
		groups    []*Group
		attribute GroupAttribute
		wantGroup *Group
		wantFound bool
	}{
		{
			name: "find group with matching attribute",
			groups: []*Group{
				{
					ID:   ptr.String("group-1"),
					Name: ptr.String("Group 1"),
					Attributes: &map[string][]string{
						"customID": {"12345"},
					},
				},
				{
					ID:   ptr.String("group-2"),
					Name: ptr.String("Group 2"),
					Attributes: &map[string][]string{
						"customID": {"67890"},
					},
				},
			},
			attribute: GroupAttribute{
				Key:   "customID",
				Value: "12345",
			},
			wantGroup: &Group{
				ID:   ptr.String("group-1"),
				Name: ptr.String("Group 1"),
				Attributes: &map[string][]string{
					"customID": {"12345"},
				},
			},
			wantFound: true,
		},
		{
			name: "attribute not found",
			groups: []*Group{
				{
					ID:   ptr.String("group-1"),
					Name: ptr.String("Group 1"),
					Attributes: &map[string][]string{
						"customID": {"12345"},
					},
				},
			},
			attribute: GroupAttribute{
				Key:   "customID",
				Value: "99999",
			},
			wantGroup: nil,
			wantFound: false,
		},
		{
			name: "attribute key not present",
			groups: []*Group{
				{
					ID:   ptr.String("group-1"),
					Name: ptr.String("Group 1"),
					Attributes: &map[string][]string{
						"otherKey": {"12345"},
					},
				},
			},
			attribute: GroupAttribute{
				Key:   "customID",
				Value: "12345",
			},
			wantGroup: nil,
			wantFound: false,
		},
		{
			name: "group with nil attributes",
			groups: []*Group{
				{
					ID:         ptr.String("group-1"),
					Name:       ptr.String("Group 1"),
					Attributes: nil,
				},
			},
			attribute: GroupAttribute{
				Key:   "customID",
				Value: "12345",
			},
			wantGroup: nil,
			wantFound: false,
		},
		{
			name: "nil group in slice",
			groups: []*Group{
				nil,
				{
					ID:   ptr.String("group-2"),
					Name: ptr.String("Group 2"),
					Attributes: &map[string][]string{
						"customID": {"12345"},
					},
				},
			},
			attribute: GroupAttribute{
				Key:   "customID",
				Value: "12345",
			},
			wantGroup: &Group{
				ID:   ptr.String("group-2"),
				Name: ptr.String("Group 2"),
				Attributes: &map[string][]string{
					"customID": {"12345"},
				},
			},
			wantFound: true,
		},
		{
			name: "attribute with multiple values",
			groups: []*Group{
				{
					ID:   ptr.String("group-1"),
					Name: ptr.String("Group 1"),
					Attributes: &map[string][]string{
						"customID": {"12345", "67890"},
					},
				},
			},
			attribute: GroupAttribute{
				Key:   "customID",
				Value: "12345",
			},
			wantGroup: nil,
			wantFound: false,
		},
		{
			name:   "empty groups slice",
			groups: []*Group{},
			attribute: GroupAttribute{
				Key:   "customID",
				Value: "12345",
			},
			wantGroup: nil,
			wantFound: false,
		},
		{
			name: "attribute with empty value",
			groups: []*Group{
				{
					ID:   ptr.String("group-1"),
					Name: ptr.String("Group 1"),
					Attributes: &map[string][]string{
						"customID": {""},
					},
				},
			},
			attribute: GroupAttribute{
				Key:   "customID",
				Value: "",
			},
			wantGroup: &Group{
				ID:   ptr.String("group-1"),
				Name: ptr.String("Group 1"),
				Attributes: &map[string][]string{
					"customID": {""},
				},
			},
			wantFound: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotGroup, gotFound := findGroupByAttribute(tt.groups, tt.attribute)
			assert.Equal(t, tt.wantFound, gotFound)
			if tt.wantFound {
				assert.NotNil(t, gotGroup)
				assert.Equal(t, *tt.wantGroup.ID, *gotGroup.ID)
			} else {
				assert.Nil(t, gotGroup)
			}
		})
	}
}

func TestGetID(t *testing.T) {
	tests := []struct {
		name       string
		location   string
		expectedID string
	}{
		{
			name:       "extract ID from Location header",
			location:   "https://keycloak.example.com/admin/realms/test-realm/groups/test-group-id-123",
			expectedID: "test-group-id-123",
		},
		{
			name:       "extract ID with special characters",
			location:   "https://keycloak.example.com/admin/realms/test-realm/groups/abc-def-123-456",
			expectedID: "abc-def-123-456",
		},
		{
			name:       "no Location header",
			location:   "",
			expectedID: "",
		},
		{
			name:       "empty Location header",
			location:   "",
			expectedID: "",
		},
		{
			name:       "Location header with trailing slash",
			location:   "https://keycloak.example.com/admin/realms/test-realm/groups/test-group-id/",
			expectedID: "test-group-id",
		},
		{
			name:       "Location header with only base URL",
			location:   "https://keycloak.example.com",
			expectedID: "keycloak.example.com",
		},
		{
			name:       "Location with UUID format",
			location:   "https://keycloak.example.com/admin/realms/test-realm/groups/550e8400-e29b-41d4-a716-446655440000",
			expectedID: "550e8400-e29b-41d4-a716-446655440000",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a mock HTTP server that returns the Location header
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if tt.location != "" {
					w.Header().Set("Location", tt.location)
				}
				w.WriteHeader(http.StatusCreated)
			}))
			defer server.Close()

			// Create resty client and make request
			restyClient := newTestRestyClient()
			resp, err := restyClient.R().Post(server.URL)
			assert.NoError(t, err)

			// Test the getID function
			result := getID(resp)
			assert.Equal(t, tt.expectedID, result)
		})
	}
}

func TestBuildURL(t *testing.T) {
	tests := []struct {
		name        string
		baseURL     string
		realm       string
		endpoint    endpoint
		params      map[string]string
		expectedURL string
	}{
		{
			name:        "groups list endpoint",
			baseURL:     "https://keycloak.example.com",
			realm:       "test-realm",
			endpoint:    endpointGroupsList,
			params:      nil,
			expectedURL: "https://keycloak.example.com/admin/realms/test-realm/groups",
		},
		{
			name:        "group get endpoint with ID",
			baseURL:     "https://keycloak.example.com",
			realm:       "test-realm",
			endpoint:    endpointGroupGet,
			params:      map[string]string{"groupID": "group-id-123"},
			expectedURL: "https://keycloak.example.com/admin/realms/test-realm/groups/group-id-123",
		},
		{
			name:        "groups count endpoint",
			baseURL:     "https://keycloak.example.com",
			realm:       "test-realm",
			endpoint:    endpointGroupsCount,
			params:      nil,
			expectedURL: "https://keycloak.example.com/admin/realms/test-realm/groups/count",
		},
		{
			name:        "group children endpoint",
			baseURL:     "https://keycloak.example.com",
			realm:       "test-realm",
			endpoint:    endpointGroupChildren,
			params:      map[string]string{"groupID": "parent-id"},
			expectedURL: "https://keycloak.example.com/admin/realms/test-realm/groups/parent-id/children",
		},
		{
			name:        "group permissions get endpoint",
			baseURL:     "https://keycloak.example.com",
			realm:       "test-realm",
			endpoint:    endpointGroupPermsGet,
			params:      map[string]string{"groupID": "group-id"},
			expectedURL: "https://keycloak.example.com/admin/realms/test-realm/groups/group-id/management/permissions",
		},
		{
			name:        "group members endpoint",
			baseURL:     "https://keycloak.example.com",
			realm:       "test-realm",
			endpoint:    endpointGroupMembers,
			params:      map[string]string{"groupID": "group-id"},
			expectedURL: "https://keycloak.example.com/admin/realms/test-realm/groups/group-id/members",
		},
		{
			name:        "baseURL with trailing slash",
			baseURL:     "https://keycloak.example.com/",
			realm:       "test-realm",
			endpoint:    endpointGroupsList,
			params:      nil,
			expectedURL: "https://keycloak.example.com//admin/realms/test-realm/groups",
		},
		{
			name:        "realm with special characters",
			baseURL:     "https://keycloak.example.com",
			realm:       "test-realm-123",
			endpoint:    endpointGroupsList,
			params:      nil,
			expectedURL: "https://keycloak.example.com/admin/realms/test-realm-123/groups",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := &Client{
				baseURL: tt.baseURL,
				realm:   tt.realm,
			}

			result := client.buildURL(tt.endpoint, tt.params)
			assert.Equal(t, tt.expectedURL, result)
		})
	}
}

func TestGroupsClient_GetSubGroupByID(t *testing.T) {
	tests := []struct {
		name        string
		parentGroup Group
		subGroupID  string
		wantErr     error
		wantGroup   *Group
	}{
		{
			name: "find existing subgroup",
			parentGroup: Group{
				ID:   ptr.String("parent-id"),
				Name: ptr.String("Parent Group"),
				SubGroups: &[]*Group{
					{
						ID:   ptr.String("sub-1"),
						Name: ptr.String("Subgroup 1"),
					},
					{
						ID:   ptr.String("sub-2"),
						Name: ptr.String("Subgroup 2"),
					},
				},
			},
			subGroupID: "sub-2",
			wantErr:    nil,
			wantGroup: &Group{
				ID:   ptr.String("sub-2"),
				Name: ptr.String("Subgroup 2"),
			},
		},
		{
			name: "subgroup not found",
			parentGroup: Group{
				ID:   ptr.String("parent-id"),
				Name: ptr.String("Parent Group"),
				SubGroups: &[]*Group{
					{
						ID:   ptr.String("sub-1"),
						Name: ptr.String("Subgroup 1"),
					},
				},
			},
			subGroupID: "non-existent",
			wantErr:    ErrGroupNotFound,
			wantGroup:  nil,
		},
		{
			name: "nil subgroup in list",
			parentGroup: Group{
				ID:   ptr.String("parent-id"),
				Name: ptr.String("Parent Group"),
				SubGroups: &[]*Group{
					nil,
					{
						ID:   ptr.String("sub-2"),
						Name: ptr.String("Subgroup 2"),
					},
				},
			},
			subGroupID: "sub-2",
			wantErr:    nil,
			wantGroup: &Group{
				ID:   ptr.String("sub-2"),
				Name: ptr.String("Subgroup 2"),
			},
		},
		{
			name: "subgroup with nil ID",
			parentGroup: Group{
				ID:   ptr.String("parent-id"),
				Name: ptr.String("Parent Group"),
				SubGroups: &[]*Group{
					{
						ID:   nil,
						Name: ptr.String("Subgroup 1"),
					},
					{
						ID:   ptr.String("sub-2"),
						Name: ptr.String("Subgroup 2"),
					},
				},
			},
			subGroupID: "sub-2",
			wantErr:    nil,
			wantGroup: &Group{
				ID:   ptr.String("sub-2"),
				Name: ptr.String("Subgroup 2"),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := &Client{}
			gc := &groupsClient{client: client}

			group, err := gc.GetSubGroupByID(tt.parentGroup, tt.subGroupID)

			if tt.wantErr != nil {
				assert.Equal(t, tt.wantErr, err)
				assert.Nil(t, group)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, group)
				assert.Equal(t, *tt.wantGroup.ID, *group.ID)
			}
		})
	}
}

func TestGroupsClient_GetSubGroupByAttribute(t *testing.T) {
	tests := []struct {
		name        string
		parentGroup Group
		attribute   GroupAttribute
		wantErr     bool
		wantGroup   *Group
	}{
		{
			name: "find subgroup by attribute",
			parentGroup: Group{
				ID:   ptr.String("parent-id"),
				Name: ptr.String("Parent Group"),
				SubGroups: &[]*Group{
					{
						ID:   ptr.String("sub-1"),
						Name: ptr.String("Subgroup 1"),
						Attributes: &map[string][]string{
							"type": {"basic"},
						},
					},
					{
						ID:   ptr.String("sub-2"),
						Name: ptr.String("Subgroup 2"),
						Attributes: &map[string][]string{
							"type": {"premium"},
						},
					},
				},
			},
			attribute: GroupAttribute{
				Key:   "type",
				Value: "premium",
			},
			wantErr: false,
			wantGroup: &Group{
				ID:   ptr.String("sub-2"),
				Name: ptr.String("Subgroup 2"),
				Attributes: &map[string][]string{
					"type": {"premium"},
				},
			},
		},
		{
			name: "subgroup not found by attribute",
			parentGroup: Group{
				ID:   ptr.String("parent-id"),
				Name: ptr.String("Parent Group"),
				SubGroups: &[]*Group{
					{
						ID:   ptr.String("sub-1"),
						Name: ptr.String("Subgroup 1"),
						Attributes: &map[string][]string{
							"type": {"basic"},
						},
					},
				},
			},
			attribute: GroupAttribute{
				Key:   "type",
				Value: "enterprise",
			},
			wantErr:   true,
			wantGroup: nil,
		},
		{
			name: "parent has no subgroups",
			parentGroup: Group{
				ID:        ptr.String("parent-id"),
				Name:      ptr.String("Parent Group"),
				SubGroups: nil,
			},
			attribute: GroupAttribute{
				Key:   "type",
				Value: "premium",
			},
			wantErr:   true,
			wantGroup: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := &Client{}
			gc := &groupsClient{client: client}

			group, err := gc.GetSubGroupByAttribute(tt.parentGroup, tt.attribute)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, group)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, group)
				assert.Equal(t, *tt.wantGroup.ID, *group.ID)
			}
		})
	}
}
