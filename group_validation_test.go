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
	"testing"

	"github.com/stretchr/testify/assert"
	"go.companyinfo.dev/ptr"
)

func TestGroupsClient_UpdateValidation(t *testing.T) {
	client := &Client{}
	gc := &groupsClient{client: client}
	ctx := context.Background()

	tests := []struct {
		name    string
		group   Group
		wantErr bool
	}{
		{
			name: "missing group ID",
			group: Group{
				Name: ptr.String("Test Group"),
			},
			wantErr: true,
		},
		{
			name: "nil group ID",
			group: Group{
				ID:   nil,
				Name: ptr.String("Test Group"),
			},
			wantErr: true,
		},
		{
			name: "empty group ID",
			group: Group{
				ID:   ptr.String(""),
				Name: ptr.String("Test Group"),
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := gc.Update(ctx, tt.group)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), "ID of the group is required")
			} else {
				// These tests validate input, actual HTTP call would fail
				// We're only testing validation logic here
			}
		})
	}
}

func TestGroupsClient_CreateSubGroupValidation(t *testing.T) {
	client := &Client{}
	gc := &groupsClient{client: client}
	ctx := context.Background()

	tests := []struct {
		name      string
		groupID   string
		subName   string
		wantErr   bool
		errString string
	}{
		{
			name:      "empty parent group ID",
			groupID:   "",
			subName:   "Subgroup",
			wantErr:   true,
			errString: "groupID parameter cannot be empty",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := gc.CreateSubGroup(ctx, tt.groupID, tt.subName, nil)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errString)
			}
		})
	}
}

func TestGroupsClient_DeleteValidation(t *testing.T) {
	client := &Client{}
	gc := &groupsClient{client: client}
	ctx := context.Background()

	tests := []struct {
		name      string
		groupID   string
		wantErr   bool
		errString string
	}{
		{
			name:      "empty group ID",
			groupID:   "",
			wantErr:   true,
			errString: "groupID parameter cannot be empty",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := gc.Delete(ctx, tt.groupID)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errString)
			}
		})
	}
}

func TestGroupsClient_ListSubGroupsPaginatedValidation(t *testing.T) {
	client := &Client{}
	gc := &groupsClient{client: client}
	ctx := context.Background()

	tests := []struct {
		name      string
		groupID   string
		params    SubGroupSearchParams
		wantErr   bool
		errString string
	}{
		{
			name:      "empty group ID",
			groupID:   "",
			params:    SubGroupSearchParams{},
			wantErr:   true,
			errString: "groupID parameter cannot be empty",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := gc.ListSubGroupsPaginated(ctx, tt.groupID, tt.params)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errString)
			}
		})
	}
}

func TestGroupsClient_GetByAttributeValidation(t *testing.T) {
	client := &Client{pageSize: 50}
	gc := &groupsClient{client: client}
	ctx := context.Background()

	tests := []struct {
		name      string
		attribute *GroupAttribute
		wantErr   bool
		errString string
	}{
		{
			name:      "nil attribute",
			attribute: nil,
			wantErr:   true,
			errString: "attribute parameter cannot be nil",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := gc.GetByAttribute(ctx, tt.attribute)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errString)
			}
		})
	}
}

func TestGroupsClient_ListMembersValidation(t *testing.T) {
	client := &Client{}
	gc := &groupsClient{client: client}
	ctx := context.Background()

	tests := []struct {
		name      string
		groupID   string
		params    GroupMembersParams
		wantErr   bool
		errString string
	}{
		{
			name:      "empty group ID",
			groupID:   "",
			params:    GroupMembersParams{},
			wantErr:   true,
			errString: "groupID parameter cannot be empty",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := gc.ListMembers(ctx, tt.groupID, tt.params)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errString)
			}
		})
	}
}

func TestGroupsClient_GetManagementPermissionsValidation(t *testing.T) {
	client := &Client{}
	gc := &groupsClient{client: client}
	ctx := context.Background()

	tests := []struct {
		name      string
		groupID   string
		wantErr   bool
		errString string
	}{
		{
			name:      "empty group ID",
			groupID:   "",
			wantErr:   true,
			errString: "groupID parameter cannot be empty",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := gc.GetManagementPermissions(ctx, tt.groupID)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errString)
			}
		})
	}
}

func TestGroupsClient_UpdateManagementPermissionsValidation(t *testing.T) {
	client := &Client{}
	gc := &groupsClient{client: client}
	ctx := context.Background()

	tests := []struct {
		name      string
		groupID   string
		ref       ManagementPermissionReference
		wantErr   bool
		errString string
	}{
		{
			name:      "empty group ID",
			groupID:   "",
			ref:       ManagementPermissionReference{},
			wantErr:   true,
			errString: "groupID parameter cannot be empty",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := gc.UpdateManagementPermissions(ctx, tt.groupID, tt.ref)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errString)
			}
		})
	}
}
