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
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.companyinfo.dev/ptr"
)

// TestGroupsClient_CreateWithServer tests Create with a mock HTTP server
func TestGroupsClient_CreateWithServer(t *testing.T) {
	tests := []struct {
		name           string
		groupName      string
		attributes     map[string][]string
		mockStatusCode int
		mockLocation   string
		mockError      *HTTPErrorResponse
		wantErr        bool
		wantID         string
	}{
		{
			name:           "successful creation",
			groupName:      "Test Group",
			attributes:     map[string][]string{"type": {"test"}},
			mockStatusCode: http.StatusCreated,
			mockLocation:   "/admin/realms/test-realm/groups/new-group-id",
			wantErr:        false,
			wantID:         "new-group-id",
		},
		{
			name:           "creation with no attributes",
			groupName:      "Simple Group",
			attributes:     nil,
			mockStatusCode: http.StatusCreated,
			mockLocation:   "/admin/realms/test-realm/groups/simple-id",
			wantErr:        false,
			wantID:         "simple-id",
		},
		{
			name:           "server returns bad request",
			groupName:      "",
			attributes:     nil,
			mockStatusCode: http.StatusBadRequest,
			mockError: &HTTPErrorResponse{
				Error:   "invalid_request",
				Message: "Group name is required",
			},
			wantErr: true,
		},
		{
			name:           "server returns conflict",
			groupName:      "Existing Group",
			attributes:     nil,
			mockStatusCode: http.StatusConflict,
			mockError: &HTTPErrorResponse{
				Error:   "conflict",
				Message: "Group already exists",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var serverURL string
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, http.MethodPost, r.Method)
				assert.Contains(t, r.URL.Path, "/admin/realms/test-realm/groups")

				// Verify request body
				var group Group
				err := json.NewDecoder(r.Body).Decode(&group)
				require.NoError(t, err)
				if tt.groupName != "" {
					assert.Equal(t, tt.groupName, *group.Name)
				}

				if tt.mockLocation != "" {
					w.Header().Set("Location", serverURL+tt.mockLocation)
				}
				w.WriteHeader(tt.mockStatusCode)
				if tt.mockError != nil {
					json.NewEncoder(w).Encode(tt.mockError)
				}
			}))
			defer server.Close()
			serverURL = server.URL

			// Create client
			client := &Client{
				baseURL:  server.URL,
				realm:    "test-realm",
				pageSize: 50,
				resty:    newTestRestyClient(),
			}
			client.resty.SetBaseURL(server.URL)
			gc := &groupsClient{
				client: client,
			}

			ctx := context.Background()
			groupID, err := gc.Create(ctx, tt.groupName, tt.attributes)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.wantID, groupID)
			}
		})
	}
}

// TestGroupsClient_ListWithServer tests List with a mock HTTP server
func TestGroupsClient_ListWithServer(t *testing.T) {
	tests := []struct {
		name                string
		search              *string
		briefRepresentation bool
		mockGroups          []*Group
		mockStatusCode      int
		wantErr             bool
		wantCount           int
	}{
		{
			name:                "list all groups",
			search:              nil,
			briefRepresentation: false,
			mockGroups: []*Group{
				{ID: ptr.String("g1"), Name: ptr.String("Group 1")},
				{ID: ptr.String("g2"), Name: ptr.String("Group 2")},
			},
			mockStatusCode: http.StatusOK,
			wantErr:        false,
			wantCount:      2,
		},
		{
			name:                "list with search",
			search:              ptr.String("Test"),
			briefRepresentation: true,
			mockGroups: []*Group{
				{ID: ptr.String("g1"), Name: ptr.String("Test Group")},
			},
			mockStatusCode: http.StatusOK,
			wantErr:        false,
			wantCount:      1,
		},
		{
			name:                "empty result",
			search:              ptr.String("NonExistent"),
			briefRepresentation: false,
			mockGroups:          []*Group{},
			mockStatusCode:      http.StatusOK,
			wantErr:             false,
			wantCount:           0,
		},
		{
			name:                "server error",
			search:              nil,
			briefRepresentation: false,
			mockStatusCode:      http.StatusInternalServerError,
			wantErr:             true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, http.MethodGet, r.Method)

				// Verify query parameters
				if tt.search != nil {
					assert.Equal(t, *tt.search, r.URL.Query().Get("search"))
				}

				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(tt.mockStatusCode)
				if tt.mockStatusCode == http.StatusOK {
					json.NewEncoder(w).Encode(tt.mockGroups)
				}
			}))
			defer server.Close()

			client := &Client{
				baseURL:  server.URL,
				realm:    "test-realm",
				pageSize: 50,
				resty:    newTestRestyClient(),
			}
			client.resty.SetBaseURL(server.URL)
			gc := &groupsClient{
				client: client,
			}

			ctx := context.Background()
			groups, err := gc.List(ctx, tt.search, tt.briefRepresentation)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Len(t, groups, tt.wantCount)
			}
		})
	}
}

// TestGroupsClient_GetWithServer tests Get with a mock HTTP server
func TestGroupsClient_GetWithServer(t *testing.T) {
	tests := []struct {
		name           string
		groupID        string
		mockGroup      *Group
		mockStatusCode int
		wantErr        bool
		checkNotFound  bool
	}{
		{
			name:    "get existing group",
			groupID: "existing-id",
			mockGroup: &Group{
				ID:          ptr.String("existing-id"),
				Name:        ptr.String("Existing Group"),
				Path:        ptr.String("/Existing Group"),
				Description: ptr.String("A test group"),
			},
			mockStatusCode: http.StatusOK,
			wantErr:        false,
		},
		{
			name:           "group not found",
			groupID:        "missing-id",
			mockStatusCode: http.StatusNotFound,
			wantErr:        true,
			checkNotFound:  true,
		},
		{
			name:           "server error",
			groupID:        "error-id",
			mockStatusCode: http.StatusInternalServerError,
			wantErr:        true,
		},
		{
			name:           "unauthorized",
			groupID:        "unauth-id",
			mockStatusCode: http.StatusUnauthorized,
			wantErr:        true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, http.MethodGet, r.Method)
				assert.Contains(t, r.URL.Path, tt.groupID)

				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(tt.mockStatusCode)
				if tt.mockStatusCode == http.StatusOK && tt.mockGroup != nil {
					json.NewEncoder(w).Encode(tt.mockGroup)
				}
			}))
			defer server.Close()

			client := &Client{
				baseURL:  server.URL,
				realm:    "test-realm",
				pageSize: 50,
				resty:    newTestRestyClient(),
			}
			client.resty.SetBaseURL(server.URL)
			gc := &groupsClient{
				client: client,
			}

			ctx := context.Background()
			group, err := gc.Get(ctx, tt.groupID)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.checkNotFound {
					assert.Equal(t, ErrGroupNotFound, err)
				}
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, group)
				assert.Equal(t, *tt.mockGroup.ID, *group.ID)
			}
		})
	}
}

// TestGroupsClient_DeleteWithServer tests Delete with a mock HTTP server
func TestGroupsClient_DeleteWithServer(t *testing.T) {
	tests := []struct {
		name           string
		groupID        string
		mockStatusCode int
		wantErr        bool
	}{
		{
			name:           "successful deletion",
			groupID:        "group-to-delete",
			mockStatusCode: http.StatusNoContent,
			wantErr:        false,
		},
		{
			name:           "group not found",
			groupID:        "missing-group",
			mockStatusCode: http.StatusNotFound,
			wantErr:        true,
		},
		{
			name:           "forbidden",
			groupID:        "protected-group",
			mockStatusCode: http.StatusForbidden,
			wantErr:        true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, http.MethodDelete, r.Method)
				assert.Contains(t, r.URL.Path, tt.groupID)
				w.WriteHeader(tt.mockStatusCode)
			}))
			defer server.Close()

			client := &Client{
				baseURL:  server.URL,
				realm:    "test-realm",
				pageSize: 50,
				resty:    newTestRestyClient(),
			}
			client.resty.SetBaseURL(server.URL)
			gc := &groupsClient{
				client: client,
			}

			ctx := context.Background()
			err := gc.Delete(ctx, tt.groupID)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// TestGroupsClient_CountWithServer tests Count with a mock HTTP server
func TestGroupsClient_CountWithServer(t *testing.T) {
	tests := []struct {
		name           string
		search         *string
		top            *bool
		mockCount      int
		mockStatusCode int
		wantErr        bool
	}{
		{
			name:           "count all groups",
			search:         nil,
			top:            nil,
			mockCount:      42,
			mockStatusCode: http.StatusOK,
			wantErr:        false,
		},
		{
			name:           "count with search",
			search:         ptr.String("Test"),
			top:            nil,
			mockCount:      5,
			mockStatusCode: http.StatusOK,
			wantErr:        false,
		},
		{
			name:           "count top level only",
			search:         nil,
			top:            ptr.Bool(true),
			mockCount:      10,
			mockStatusCode: http.StatusOK,
			wantErr:        false,
		},
		{
			name:           "server error",
			search:         nil,
			top:            nil,
			mockStatusCode: http.StatusInternalServerError,
			wantErr:        true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, http.MethodGet, r.Method)
				assert.Contains(t, r.URL.Path, "/count")

				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(tt.mockStatusCode)
				if tt.mockStatusCode == http.StatusOK {
					json.NewEncoder(w).Encode(CountGroupResponse{Count: tt.mockCount})
				}
			}))
			defer server.Close()

			client := &Client{
				baseURL:  server.URL,
				realm:    "test-realm",
				pageSize: 50,
				resty:    newTestRestyClient(),
			}
			client.resty.SetBaseURL(server.URL)
			gc := &groupsClient{
				client: client,
			}

			ctx := context.Background()
			count, err := gc.Count(ctx, tt.search, tt.top)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.mockCount, count)
			}
		})
	}
}

// TestGroupsClient_ListSubGroupsWithServer tests ListSubGroups with a mock HTTP server
func TestGroupsClient_ListSubGroupsWithServer(t *testing.T) {
	tests := []struct {
		name           string
		groupID        string
		mockSubGroups  []*Group
		mockStatusCode int
		wantErr        bool
		wantCount      int
	}{
		{
			name:    "list subgroups",
			groupID: "parent-id",
			mockSubGroups: []*Group{
				{ID: ptr.String("sub1"), Name: ptr.String("Subgroup 1")},
				{ID: ptr.String("sub2"), Name: ptr.String("Subgroup 2")},
			},
			mockStatusCode: http.StatusOK,
			wantErr:        false,
			wantCount:      2,
		},
		{
			name:           "no subgroups",
			groupID:        "parent-id",
			mockSubGroups:  []*Group{},
			mockStatusCode: http.StatusOK,
			wantErr:        false,
			wantCount:      0,
		},
		{
			name:           "parent not found",
			groupID:        "missing-id",
			mockStatusCode: http.StatusNotFound,
			wantErr:        true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, http.MethodGet, r.Method)
				assert.Contains(t, r.URL.Path, tt.groupID)
				assert.Contains(t, r.URL.Path, "/children")

				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(tt.mockStatusCode)
				if tt.mockStatusCode == http.StatusOK {
					json.NewEncoder(w).Encode(tt.mockSubGroups)
				}
			}))
			defer server.Close()

			client := &Client{
				baseURL:  server.URL,
				realm:    "test-realm",
				pageSize: 50,
				resty:    newTestRestyClient(),
			}
			client.resty.SetBaseURL(server.URL)
			gc := &groupsClient{
				client: client,
			}

			ctx := context.Background()
			subGroups, err := gc.ListSubGroups(ctx, tt.groupID)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Len(t, subGroups, tt.wantCount)
			}
		})
	}
}

// TestGroupsClient_UpdateWithServer tests Update with a mock HTTP server
func TestGroupsClient_UpdateWithServer(t *testing.T) {
	tests := []struct {
		name           string
		group          Group
		mockStatusCode int
		wantErr        bool
	}{
		{
			name: "successful update",
			group: Group{
				ID:          ptr.String("group-1"),
				Name:        ptr.String("Updated Group"),
				Description: ptr.String("Updated description"),
			},
			mockStatusCode: http.StatusNoContent,
			wantErr:        false,
		},
		{
			name: "group not found",
			group: Group{
				ID:   ptr.String("missing-id"),
				Name: ptr.String("Updated Group"),
			},
			mockStatusCode: http.StatusNotFound,
			wantErr:        true,
		},
		{
			name: "conflict",
			group: Group{
				ID:   ptr.String("group-1"),
				Name: ptr.String("Duplicate Name"),
			},
			mockStatusCode: http.StatusConflict,
			wantErr:        true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, http.MethodPut, r.Method)
				assert.Contains(t, r.URL.Path, *tt.group.ID)

				var reqGroup Group
				err := json.NewDecoder(r.Body).Decode(&reqGroup)
				require.NoError(t, err)
				assert.Equal(t, *tt.group.Name, *reqGroup.Name)

				w.WriteHeader(tt.mockStatusCode)
			}))
			defer server.Close()

			client := &Client{
				baseURL:  server.URL,
				realm:    "test-realm",
				pageSize: 50,
				resty:    newTestRestyClient(),
			}
			client.resty.SetBaseURL(server.URL)
			gc := &groupsClient{
				client: client,
			}

			ctx := context.Background()
			err := gc.Update(ctx, tt.group)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// TestGroupsClient_ListMembersWithServer tests ListMembers with a mock HTTP server
func TestGroupsClient_ListMembersWithServer(t *testing.T) {
	tests := []struct {
		name           string
		groupID        string
		params         GroupMembersParams
		mockUsers      []*User
		mockStatusCode int
		wantErr        bool
		wantCount      int
	}{
		{
			name:    "list members",
			groupID: "group-1",
			params: GroupMembersParams{
				Max: ptr.Int(100),
			},
			mockUsers: []*User{
				{ID: ptr.String("u1"), Username: ptr.String("user1")},
				{ID: ptr.String("u2"), Username: ptr.String("user2")},
			},
			mockStatusCode: http.StatusOK,
			wantErr:        false,
			wantCount:      2,
		},
		{
			name:           "no members",
			groupID:        "empty-group",
			params:         GroupMembersParams{},
			mockUsers:      []*User{},
			mockStatusCode: http.StatusOK,
			wantErr:        false,
			wantCount:      0,
		},
		{
			name:           "group not found",
			groupID:        "missing-group",
			params:         GroupMembersParams{},
			mockStatusCode: http.StatusNotFound,
			wantErr:        true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, http.MethodGet, r.Method)
				assert.Contains(t, r.URL.Path, tt.groupID)
				assert.Contains(t, r.URL.Path, "/members")

				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(tt.mockStatusCode)
				if tt.mockStatusCode == http.StatusOK {
					json.NewEncoder(w).Encode(tt.mockUsers)
				}
			}))
			defer server.Close()

			client := &Client{
				baseURL:  server.URL,
				realm:    "test-realm",
				pageSize: 50,
				resty:    newTestRestyClient(),
			}
			client.resty.SetBaseURL(server.URL)
			gc := &groupsClient{
				client: client,
			}

			ctx := context.Background()
			users, err := gc.ListMembers(ctx, tt.groupID, tt.params)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Len(t, users, tt.wantCount)
			}
		})
	}
}

// TestGroupsClient_ManagementPermissionsWithServer tests permissions with a mock HTTP server
func TestGroupsClient_ManagementPermissionsWithServer(t *testing.T) {
	t.Run("get management permissions", func(t *testing.T) {
		groupID := "group-1"
		mockPerms := &ManagementPermissionReference{
			Enabled:  ptr.Bool(true),
			Resource: ptr.String("resource-id"),
		}

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, http.MethodGet, r.Method)
			assert.Contains(t, r.URL.Path, groupID)
			assert.Contains(t, r.URL.Path, "/management/permissions")

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(mockPerms)
		}))
		defer server.Close()

		client := &Client{
			baseURL:  server.URL,
			realm:    "test-realm",
			pageSize: 50,
			resty:    newTestRestyClient(),
		}
		client.resty.SetBaseURL(server.URL)
		gc := &groupsClient{
			client: client,
		}

		ctx := context.Background()
		perms, err := gc.GetManagementPermissions(ctx, groupID)

		assert.NoError(t, err)
		assert.NotNil(t, perms)
		assert.Equal(t, *mockPerms.Enabled, *perms.Enabled)
	})

	t.Run("update management permissions", func(t *testing.T) {
		groupID := "group-1"
		inputRef := ManagementPermissionReference{
			Enabled: ptr.Bool(true),
		}

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, http.MethodPut, r.Method)
			assert.Contains(t, r.URL.Path, groupID)
			assert.Contains(t, r.URL.Path, "/management/permissions")

			var reqRef ManagementPermissionReference
			err := json.NewDecoder(r.Body).Decode(&reqRef)
			require.NoError(t, err)
			assert.Equal(t, *inputRef.Enabled, *reqRef.Enabled)

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(inputRef)
		}))
		defer server.Close()

		client := &Client{
			baseURL:  server.URL,
			realm:    "test-realm",
			pageSize: 50,
			resty:    newTestRestyClient(),
		}
		client.resty.SetBaseURL(server.URL)
		gc := &groupsClient{
			client: client,
		}

		ctx := context.Background()
		result, err := gc.UpdateManagementPermissions(ctx, groupID, inputRef)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, *inputRef.Enabled, *result.Enabled)
	})
}

// TestGroupsClient_ListPaginatedWithServer tests ListPaginated with a mock HTTP server
func TestGroupsClient_ListPaginatedWithServer(t *testing.T) {
	tests := []struct {
		name                string
		search              *string
		briefRepresentation bool
		first               int
		max                 int
		mockGroups          []*Group
		mockStatusCode      int
		wantErr             bool
	}{
		{
			name:                "paginated list",
			search:              nil,
			briefRepresentation: true,
			first:               0,
			max:                 10,
			mockGroups: []*Group{
				{ID: ptr.String("g1"), Name: ptr.String("Group 1")},
			},
			mockStatusCode: http.StatusOK,
			wantErr:        false,
		},
		{
			name:                "second page",
			search:              nil,
			briefRepresentation: false,
			first:               10,
			max:                 10,
			mockGroups:          []*Group{},
			mockStatusCode:      http.StatusOK,
			wantErr:             false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, http.MethodGet, r.Method)
				assert.Equal(t, fmt.Sprintf("%d", tt.first), r.URL.Query().Get("first"))
				assert.Equal(t, fmt.Sprintf("%d", tt.max), r.URL.Query().Get("max"))

				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(tt.mockStatusCode)
				json.NewEncoder(w).Encode(tt.mockGroups)
			}))
			defer server.Close()

			client := &Client{
				baseURL:  server.URL,
				realm:    "test-realm",
				pageSize: 50,
				resty:    newTestRestyClient(),
			}
			client.resty.SetBaseURL(server.URL)
			gc := &groupsClient{
				client: client,
			}

			ctx := context.Background()
			groups, err := gc.ListPaginated(ctx, tt.search, tt.briefRepresentation, tt.first, tt.max)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Len(t, groups, len(tt.mockGroups))
			}
		})
	}
}

// TestGroupsClient_ListWithParamsWithServer tests ListWithParams with various scenarios
func TestGroupsClient_ListWithParamsWithServer(t *testing.T) {
	tests := []struct {
		name           string
		params         SearchGroupParams
		mockGroups     []*Group
		mockStatusCode int
		wantErr        bool
		verifyQuery    func(t *testing.T, r *http.Request)
	}{
		{
			name: "exact search",
			params: SearchGroupParams{
				Search: ptr.String("Exact Group"),
				Exact:  ptr.Bool(true),
			},
			mockGroups: []*Group{
				{ID: ptr.String("g1"), Name: ptr.String("Exact Group")},
			},
			mockStatusCode: http.StatusOK,
			wantErr:        false,
			verifyQuery: func(t *testing.T, r *http.Request) {
				assert.Equal(t, "Exact Group", r.URL.Query().Get("search"))
				assert.Equal(t, "true", r.URL.Query().Get("exact"))
			},
		},
		{
			name: "with populate hierarchy",
			params: SearchGroupParams{
				PopulateHierarchy: ptr.Bool(true),
				Search:            ptr.String("parent"),
			},
			mockGroups: []*Group{
				{
					ID:   ptr.String("p1"),
					Name: ptr.String("Parent"),
					SubGroups: &[]*Group{
						{ID: ptr.String("c1"), Name: ptr.String("Child")},
					},
				},
			},
			mockStatusCode: http.StatusOK,
			wantErr:        false,
			verifyQuery: func(t *testing.T, r *http.Request) {
				assert.Equal(t, "true", r.URL.Query().Get("populateHierarchy"))
			},
		},
		{
			name: "with q parameter",
			params: SearchGroupParams{
				Q: ptr.String("query string"),
			},
			mockGroups:     []*Group{},
			mockStatusCode: http.StatusOK,
			wantErr:        false,
			verifyQuery: func(t *testing.T, r *http.Request) {
				assert.Equal(t, "query string", r.URL.Query().Get("q"))
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, http.MethodGet, r.Method)
				if tt.verifyQuery != nil {
					tt.verifyQuery(t, r)
				}

				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(tt.mockStatusCode)
				json.NewEncoder(w).Encode(tt.mockGroups)
			}))
			defer server.Close()

			client := &Client{
				baseURL:  server.URL,
				realm:    "test-realm",
				pageSize: 50,
				resty:    newTestRestyClient(),
			}
			client.resty.SetBaseURL(server.URL)
			gc := &groupsClient{
				client: client,
			}

			ctx := context.Background()
			groups, err := gc.ListWithParams(ctx, tt.params)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, groups)
			}
		})
	}
}

// TestGroupsClient_ListWithSubGroupsWithServer tests the convenience method
func TestGroupsClient_ListWithSubGroupsWithServer(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method)

		// Verify required parameters are set
		assert.NotEmpty(t, r.URL.Query().Get("search"))
		assert.Equal(t, "true", r.URL.Query().Get("populateHierarchy"))

		mockGroups := []*Group{
			{
				ID:   ptr.String("p1"),
				Name: ptr.String("Parent"),
				SubGroups: &[]*Group{
					{ID: ptr.String("c1"), Name: ptr.String("Child 1")},
					{ID: ptr.String("c2"), Name: ptr.String("Child 2")},
				},
			},
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(mockGroups)
	}))
	defer server.Close()

	client := &Client{
		baseURL:  server.URL,
		realm:    "test-realm",
		pageSize: 50,
		resty:    newTestRestyClient(),
	}
	client.resty.SetBaseURL(server.URL)
	gc := &groupsClient{
		client: client,
	}

	ctx := context.Background()
	groups, err := gc.ListWithSubGroups(ctx, "parent", false, 0, 100)

	assert.NoError(t, err)
	assert.Len(t, groups, 1)
	assert.NotNil(t, groups[0].SubGroups)
	assert.Len(t, *groups[0].SubGroups, 2)
}

// TestGroupsClient_CreateSubGroupWithServer tests CreateSubGroup with a mock HTTP server
func TestGroupsClient_CreateSubGroupWithServer(t *testing.T) {
	tests := []struct {
		name           string
		parentID       string
		subGroupName   string
		attributes     map[string][]string
		mockStatusCode int
		mockLocation   string
		wantErr        bool
		wantID         string
	}{
		{
			name:           "create subgroup",
			parentID:       "parent-1",
			subGroupName:   "Child Group",
			attributes:     map[string][]string{"type": {"child"}},
			mockStatusCode: http.StatusCreated,
			mockLocation:   "/admin/realms/test-realm/groups/child-id",
			wantErr:        false,
			wantID:         "child-id",
		},
		{
			name:           "parent not found",
			parentID:       "missing-parent",
			subGroupName:   "Child Group",
			mockStatusCode: http.StatusNotFound,
			wantErr:        true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var serverURL string
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, http.MethodPost, r.Method)
				assert.Contains(t, r.URL.Path, tt.parentID)
				assert.Contains(t, r.URL.Path, "/children")

				if tt.mockLocation != "" {
					w.Header().Set("Location", serverURL+tt.mockLocation)
				}
				w.WriteHeader(tt.mockStatusCode)
			}))
			defer server.Close()
			serverURL = server.URL

			client := &Client{
				baseURL:  server.URL,
				realm:    "test-realm",
				pageSize: 50,
				resty:    newTestRestyClient(),
			}
			client.resty.SetBaseURL(server.URL)
			gc := &groupsClient{
				client: client,
			}

			ctx := context.Background()
			subGroupID, err := gc.CreateSubGroup(ctx, tt.parentID, tt.subGroupName, tt.attributes)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.wantID, subGroupID)
			}
		})
	}
}
