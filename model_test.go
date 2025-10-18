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
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.companyinfo.dev/ptr"
)

func TestGroup_JSONMarshaling(t *testing.T) {
	tests := []struct {
		name     string
		group    Group
		wantJSON string
	}{
		{
			name: "minimal group",
			group: Group{
				ID:   ptr.String("group-1"),
				Name: ptr.String("Test Group"),
			},
			wantJSON: `{"id":"group-1","name":"Test Group"}`,
		},
		{
			name: "group with attributes",
			group: Group{
				ID:   ptr.String("group-1"),
				Name: ptr.String("Test Group"),
				Attributes: &map[string][]string{
					"type": {"customer"},
				},
			},
			wantJSON: `{"id":"group-1","name":"Test Group","attributes":{"type":["customer"]}}`,
		},
		{
			name: "group with empty attributes",
			group: Group{
				ID:         ptr.String("group-1"),
				Name:       ptr.String("Test Group"),
				Attributes: &map[string][]string{},
			},
			wantJSON: `{"id":"group-1","name":"Test Group","attributes":{}}`,
		},
		{
			name: "group with subgroups",
			group: Group{
				ID:   ptr.String("parent"),
				Name: ptr.String("Parent"),
				SubGroups: &[]*Group{
					{
						ID:   ptr.String("child-1"),
						Name: ptr.String("Child 1"),
					},
				},
			},
			wantJSON: `{"id":"parent","name":"Parent","subGroups":[{"id":"child-1","name":"Child 1"}]}`,
		},
		{
			name: "group with all fields",
			group: Group{
				ID:            ptr.String("group-1"),
				Name:          ptr.String("Test Group"),
				Description:   ptr.String("A test group"),
				Path:          ptr.String("/Test Group"),
				ParentID:      ptr.String("parent-id"),
				SubGroupCount: ptr.Int64(5),
				Attributes: &map[string][]string{
					"key": {"value"},
				},
				Access: &map[string]bool{
					"view":   true,
					"manage": false,
				},
				ClientRoles: &map[string][]string{
					"client1": {"role1"},
				},
				RealmRoles: &[]string{"admin"},
			},
			wantJSON: `{"id":"group-1","name":"Test Group","description":"A test group","path":"/Test Group","parentId":"parent-id","subGroupCount":5,"attributes":{"key":["value"]},"access":{"view":true,"manage":false},"clientRoles":{"client1":["role1"]},"realmRoles":["admin"]}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Marshal to JSON
			jsonBytes, err := json.Marshal(tt.group)
			require.NoError(t, err)
			assert.JSONEq(t, tt.wantJSON, string(jsonBytes))

			// Unmarshal back
			var unmarshaled Group
			err = json.Unmarshal(jsonBytes, &unmarshaled)
			require.NoError(t, err)

			// Verify fields match
			if tt.group.ID != nil {
				assert.Equal(t, *tt.group.ID, *unmarshaled.ID)
			}
			if tt.group.Name != nil {
				assert.Equal(t, *tt.group.Name, *unmarshaled.Name)
			}
		})
	}
}

func TestSearchGroupParams_Marshaling(t *testing.T) {
	tests := []struct {
		name   string
		params SearchGroupParams
	}{
		{
			name: "all parameters",
			params: SearchGroupParams{
				BriefRepresentation: ptr.Bool(true),
				PopulateHierarchy:   ptr.Bool(false),
				Exact:               ptr.Bool(true),
				First:               ptr.Int(0),
				Full:                ptr.Bool(false),
				Max:                 ptr.Int(50),
				Q:                   ptr.String("query"),
				Search:              ptr.String("search"),
				SubGroupsCount:      ptr.Bool(true),
			},
		},
		{
			name: "minimal parameters",
			params: SearchGroupParams{
				Search: ptr.String("test"),
			},
		},
		{
			name:   "empty parameters",
			params: SearchGroupParams{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := mapper(tt.params)
			require.NoError(t, err)
			assert.NotNil(t, result)

			// Verify expected fields are present
			if tt.params.Search != nil {
				assert.Contains(t, result, "search")
			}
			if tt.params.BriefRepresentation != nil {
				assert.Contains(t, result, "briefRepresentation")
			}
		})
	}
}

func TestSubGroupSearchParams_Marshaling(t *testing.T) {
	tests := []struct {
		name   string
		params SubGroupSearchParams
	}{
		{
			name: "all parameters",
			params: SubGroupSearchParams{
				BriefRepresentation: ptr.Bool(false),
				Exact:               ptr.Bool(true),
				First:               ptr.Int(10),
				Max:                 ptr.Int(20),
				Search:              ptr.String("sub"),
				SubGroupsCount:      ptr.Bool(false),
			},
		},
		{
			name:   "empty parameters",
			params: SubGroupSearchParams{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := mapper(tt.params)
			require.NoError(t, err)
			assert.NotNil(t, result)
		})
	}
}

func TestGroupMembersParams_Marshaling(t *testing.T) {
	tests := []struct {
		name   string
		params GroupMembersParams
	}{
		{
			name: "all parameters",
			params: GroupMembersParams{
				BriefRepresentation: ptr.Bool(true),
				First:               ptr.Int(0),
				Max:                 ptr.Int(100),
			},
		},
		{
			name:   "empty parameters",
			params: GroupMembersParams{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := mapper(tt.params)
			require.NoError(t, err)
			assert.NotNil(t, result)
		})
	}
}

func TestCountGroupParams_Marshaling(t *testing.T) {
	tests := []struct {
		name   string
		params CountGroupParams
	}{
		{
			name: "with search",
			params: CountGroupParams{
				Search: ptr.String("test"),
				Top:    ptr.Bool(true),
			},
		},
		{
			name:   "empty parameters",
			params: CountGroupParams{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := mapper(tt.params)
			require.NoError(t, err)
			assert.NotNil(t, result)
		})
	}
}

func TestManagementPermissionReference_JSONMarshaling(t *testing.T) {
	tests := []struct {
		name string
		ref  ManagementPermissionReference
	}{
		{
			name: "enabled permissions",
			ref: ManagementPermissionReference{
				Enabled:  ptr.Bool(true),
				Resource: ptr.String("resource-id"),
				ScopePermissions: &map[string]string{
					"view": "permission-id-1",
				},
			},
		},
		{
			name: "disabled permissions",
			ref: ManagementPermissionReference{
				Enabled: ptr.Bool(false),
			},
		},
		{
			name: "empty permissions",
			ref:  ManagementPermissionReference{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			jsonBytes, err := json.Marshal(tt.ref)
			require.NoError(t, err)

			var unmarshaled ManagementPermissionReference
			err = json.Unmarshal(jsonBytes, &unmarshaled)
			require.NoError(t, err)

			if tt.ref.Enabled != nil {
				assert.Equal(t, *tt.ref.Enabled, *unmarshaled.Enabled)
			}
		})
	}
}

func TestUser_JSONMarshaling(t *testing.T) {
	tests := []struct {
		name     string
		user     User
		wantJSON string
	}{
		{
			name: "minimal user",
			user: User{
				ID:       ptr.String("user-1"),
				Username: ptr.String("john.doe"),
			},
			wantJSON: `{"id":"user-1","username":"john.doe"}`,
		},
		{
			name: "user with email",
			user: User{
				ID:            ptr.String("user-1"),
				Username:      ptr.String("john.doe"),
				Email:         ptr.String("john@example.com"),
				EmailVerified: ptr.Bool(true),
			},
			wantJSON: `{"id":"user-1","username":"john.doe","email":"john@example.com","emailVerified":true}`,
		},
		{
			name: "user with attributes",
			user: User{
				ID:       ptr.String("user-1"),
				Username: ptr.String("john.doe"),
				Attributes: &map[string][]string{
					"department": {"engineering"},
				},
			},
			wantJSON: `{"id":"user-1","username":"john.doe","attributes":{"department":["engineering"]}}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			jsonBytes, err := json.Marshal(tt.user)
			require.NoError(t, err)
			assert.JSONEq(t, tt.wantJSON, string(jsonBytes))

			var unmarshaled User
			err = json.Unmarshal(jsonBytes, &unmarshaled)
			require.NoError(t, err)

			if tt.user.ID != nil {
				assert.Equal(t, *tt.user.ID, *unmarshaled.ID)
			}
		})
	}
}

func TestGroupAttribute_Struct(t *testing.T) {
	attr := GroupAttribute{
		Key:   "testKey",
		Value: "testValue",
	}

	assert.Equal(t, "testKey", attr.Key)
	assert.Equal(t, "testValue", attr.Value)

	// Test JSON marshaling
	jsonBytes, err := json.Marshal(attr)
	require.NoError(t, err)
	assert.Contains(t, string(jsonBytes), "testKey")
	assert.Contains(t, string(jsonBytes), "testValue")

	// Test JSON unmarshaling
	var unmarshaled GroupAttribute
	err = json.Unmarshal(jsonBytes, &unmarshaled)
	require.NoError(t, err)
	assert.Equal(t, attr.Key, unmarshaled.Key)
	assert.Equal(t, attr.Value, unmarshaled.Value)
}

func TestCountGroupResponse_JSONMarshaling(t *testing.T) {
	tests := []struct {
		name     string
		response CountGroupResponse
		wantJSON string
	}{
		{
			name:     "zero count",
			response: CountGroupResponse{Count: 0},
			wantJSON: `{"count":0}`,
		},
		{
			name:     "positive count",
			response: CountGroupResponse{Count: 42},
			wantJSON: `{"count":42}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			jsonBytes, err := json.Marshal(tt.response)
			require.NoError(t, err)
			assert.JSONEq(t, tt.wantJSON, string(jsonBytes))

			var unmarshaled CountGroupResponse
			err = json.Unmarshal(jsonBytes, &unmarshaled)
			require.NoError(t, err)
			assert.Equal(t, tt.response.Count, unmarshaled.Count)
		})
	}
}
