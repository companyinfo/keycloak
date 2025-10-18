//go:build integration

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

// Package keycloak_test provides integration tests for the keycloak package.
// These tests require a running Keycloak instance and are only executed when
// the integration build tag is specified.
package keycloak_test

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
	"go.companyinfo.dev/keycloak"
	"go.companyinfo.dev/ptr"
)

// GroupsIntegrationTestSuite tests Groups operations against a real Keycloak instance.
// Run with: go test -v -tags=integration ./...
//
// Required environment variables:
//   - KEYCLOAK_URL: Keycloak server URL (e.g., https://keycloak.example.com)
//   - KEYCLOAK_REALM: Realm name for testing
//   - KEYCLOAK_CLIENT_ID: Client ID with admin privileges
//   - KEYCLOAK_CLIENT_SECRET: Client secret
type GroupsIntegrationTestSuite struct {
	suite.Suite
	ctx    context.Context
	client *keycloak.Client
	realm  string

	// Track created resources for cleanup
	createdGroups []string
}

// SetupSuite runs once before all tests - creates authenticated client
func (s *GroupsIntegrationTestSuite) SetupSuite() {
	s.ctx = context.Background()

	// Get configuration from environment
	url := os.Getenv("KEYCLOAK_URL")
	realm := os.Getenv("KEYCLOAK_REALM")
	clientID := os.Getenv("KEYCLOAK_CLIENT_ID")
	clientSecret := os.Getenv("KEYCLOAK_CLIENT_SECRET")

	// Validate environment variables
	s.Require().NotEmpty(url, "KEYCLOAK_URL environment variable is required")
	s.Require().NotEmpty(realm, "KEYCLOAK_REALM environment variable is required")
	s.Require().NotEmpty(clientID, "KEYCLOAK_CLIENT_ID environment variable is required")
	s.Require().NotEmpty(clientSecret, "KEYCLOAK_CLIENT_SECRET environment variable is required")

	s.realm = realm

	// Create client with options
	client, err := keycloak.New(s.ctx, keycloak.Config{
		URL:          url,
		Realm:        realm,
		ClientID:     clientID,
		ClientSecret: clientSecret,
	},
		keycloak.WithTimeout(30*time.Second),
		keycloak.WithRetry(3, 2*time.Second, 10*time.Second),
	)

	s.Require().NoError(err, "Failed to create Keycloak client")
	s.Require().NotNil(client, "Client should not be nil")

	s.client = client
}

// SetupTest runs before each test - resets tracking
func (s *GroupsIntegrationTestSuite) SetupTest() {
	s.createdGroups = []string{}
}

// TearDownTest runs after each test - cleans up created resources
func (s *GroupsIntegrationTestSuite) TearDownTest() {
	// Clean up groups in reverse order (children first)
	for i := len(s.createdGroups) - 1; i >= 0; i-- {
		groupID := s.createdGroups[i]
		err := s.client.Groups.Delete(s.ctx, groupID)
		if err != nil {
			s.T().Logf("Warning: Failed to cleanup group %s: %v", groupID, err)
		}
	}
}

// trackGroup adds a group ID to cleanup list
func (s *GroupsIntegrationTestSuite) trackGroup(groupID string) {
	s.createdGroups = append(s.createdGroups, groupID)
}

// TestGroupLifecycle tests complete CRUD operations for a group
func (s *GroupsIntegrationTestSuite) TestGroupLifecycle() {
	// Create
	groupName := fmt.Sprintf("test-group-%d", time.Now().Unix())
	attributes := map[string][]string{
		"description": {"Integration test group"},
		"type":        {"test"},
	}

	groupID, err := s.client.Groups.Create(s.ctx, groupName, attributes)
	s.Require().NoError(err)
	s.Require().NotEmpty(groupID)
	s.trackGroup(groupID)

	// Read
	group, err := s.client.Groups.Get(s.ctx, groupID)
	s.Require().NoError(err)
	s.Require().NotNil(group)
	s.Equal(groupName, *group.Name)
	s.NotNil(group.Attributes)
	s.Equal(attributes["type"], (*group.Attributes)["type"])

	// Update
	newDescription := "Updated description"
	group.Description = ptr.String(newDescription)
	(*group.Attributes)["updated"] = []string{"true"}

	err = s.client.Groups.Update(s.ctx, *group)
	s.NoError(err)

	// Verify update
	updatedGroup, err := s.client.Groups.Get(s.ctx, groupID)
	s.Require().NoError(err)
	s.Equal(newDescription, *updatedGroup.Description)
	s.Equal([]string{"true"}, (*updatedGroup.Attributes)["updated"])

	// Delete
	err = s.client.Groups.Delete(s.ctx, groupID)
	s.NoError(err)

	// Verify deletion
	_, err = s.client.Groups.Get(s.ctx, groupID)
	s.Error(err)
	s.Equal(keycloak.ErrGroupNotFound, err)

	// Remove from tracking since already deleted
	s.createdGroups = s.createdGroups[:len(s.createdGroups)-1]
}

// TestCreateGroup tests group creation
func (s *GroupsIntegrationTestSuite) TestCreateGroup() {
	groupName := fmt.Sprintf("test-create-%d", time.Now().Unix())

	groupID, err := s.client.Groups.Create(s.ctx, groupName, nil)
	s.Require().NoError(err)
	s.NotEmpty(groupID)
	s.trackGroup(groupID)

	// Verify group was created
	group, err := s.client.Groups.Get(s.ctx, groupID)
	s.NoError(err)
	s.Equal(groupName, *group.Name)
}

// TestListGroups tests listing groups
func (s *GroupsIntegrationTestSuite) TestListGroups() {
	// Create test groups
	baseGroup := fmt.Sprintf("test-list-%d", time.Now().Unix())
	groupID1, err := s.client.Groups.Create(s.ctx, baseGroup+"-1", nil)
	s.Require().NoError(err)
	s.trackGroup(groupID1)

	groupID2, err := s.client.Groups.Create(s.ctx, baseGroup+"-2", nil)
	s.Require().NoError(err)
	s.trackGroup(groupID2)

	// List all groups
	groups, err := s.client.Groups.List(s.ctx, nil, false)
	s.NoError(err)
	s.NotEmpty(groups)

	// Verify our groups are in the list
	foundCount := 0
	for _, group := range groups {
		if *group.ID == groupID1 || *group.ID == groupID2 {
			foundCount++
		}
	}
	s.Equal(2, foundCount, "Should find both created groups")
}

// TestListWithParams tests advanced group search
func (s *GroupsIntegrationTestSuite) TestListWithParams() {
	// Create a uniquely named group
	uniqueName := fmt.Sprintf("test-search-unique-%d", time.Now().Unix())
	groupID, err := s.client.Groups.Create(s.ctx, uniqueName, nil)
	s.Require().NoError(err)
	s.trackGroup(groupID)

	// Search with exact match
	params := keycloak.SearchGroupParams{
		Search: &uniqueName,
		Exact:  ptr.Bool(true),
	}

	groups, err := s.client.Groups.ListWithParams(s.ctx, params)
	s.NoError(err)
	s.Len(groups, 1, "Should find exactly one group with exact search")
	s.Equal(uniqueName, *groups[0].Name)
}

// TestGetByAttribute tests finding groups by attribute
func (s *GroupsIntegrationTestSuite) TestGetByAttribute() {
	// Create group with unique attribute
	uniqueValue := fmt.Sprintf("test-attr-%d", time.Now().Unix())
	attributes := map[string][]string{
		"testID": {uniqueValue},
	}

	groupName := fmt.Sprintf("test-attribute-%d", time.Now().Unix())
	groupID, err := s.client.Groups.Create(s.ctx, groupName, attributes)
	s.Require().NoError(err)
	s.trackGroup(groupID)

	// Search by attribute
	attr := &keycloak.GroupAttribute{
		Key:   "testID",
		Value: uniqueValue,
	}

	group, err := s.client.Groups.GetByAttribute(s.ctx, attr)
	s.NoError(err)
	s.NotNil(group)
	s.Equal(groupID, *group.ID)
}

// TestSubGroups tests subgroup operations
func (s *GroupsIntegrationTestSuite) TestSubGroups() {
	// Create parent group
	parentName := fmt.Sprintf("test-parent-%d", time.Now().Unix())
	parentID, err := s.client.Groups.Create(s.ctx, parentName, nil)
	s.Require().NoError(err)
	s.trackGroup(parentID)

	// Create subgroup
	subGroupName := fmt.Sprintf("test-sub-%d", time.Now().Unix())
	subGroupID, err := s.client.Groups.CreateSubGroup(s.ctx, parentID, subGroupName, nil)
	s.Require().NoError(err)
	s.NotEmpty(subGroupID)
	s.trackGroup(subGroupID) // Track for cleanup (will be cleaned up before parent)

	// Verify subgroup
	subGroup, err := s.client.Groups.Get(s.ctx, subGroupID)
	s.NoError(err)
	s.Equal(subGroupName, *subGroup.Name)
	s.NotNil(subGroup.ParentID)
	s.Equal(parentID, *subGroup.ParentID)

	// List subgroups
	params := keycloak.SubGroupSearchParams{
		BriefRepresentation: ptr.Bool(false),
	}

	subGroups, err := s.client.Groups.ListSubGroupsPaginated(s.ctx, parentID, params)
	s.NoError(err)
	s.Len(subGroups, 1)
	s.Equal(subGroupName, *subGroups[0].Name)
}

// TestGroupCount tests counting groups
func (s *GroupsIntegrationTestSuite) TestGroupCount() {
	count, err := s.client.Groups.Count(s.ctx, nil, nil)
	s.NoError(err)
	s.GreaterOrEqual(count, 0)
}

// TestManagementPermissions tests management permissions
func (s *GroupsIntegrationTestSuite) TestManagementPermissions() {
	// Create group
	groupName := fmt.Sprintf("test-permissions-%d", time.Now().Unix())
	groupID, err := s.client.Groups.Create(s.ctx, groupName, nil)
	s.Require().NoError(err)
	s.trackGroup(groupID)

	// Get current permissions
	permissions, err := s.client.Groups.GetManagementPermissions(s.ctx, groupID)

	// Skip test if management permissions feature is not enabled in Keycloak
	if err != nil {
		s.T().Skipf("Management permissions feature not enabled in Keycloak: %v", err)
		return
	}

	s.NotNil(permissions)

	// Enable permissions if not already enabled
	if permissions.Enabled == nil || !*permissions.Enabled {
		ref := keycloak.ManagementPermissionReference{
			Enabled: ptr.Bool(true),
		}

		result, err := s.client.Groups.UpdateManagementPermissions(s.ctx, groupID, ref)
		s.NoError(err)
		s.NotNil(result)
		s.NotNil(result.Enabled)
		s.True(*result.Enabled)
	}

	// Verify permissions are enabled
	updatedPermissions, err := s.client.Groups.GetManagementPermissions(s.ctx, groupID)
	s.NoError(err)
	s.NotNil(updatedPermissions.Enabled)
	s.True(*updatedPermissions.Enabled)
}

// TestPaginatedListing tests pagination
func (s *GroupsIntegrationTestSuite) TestPaginatedListing() {
	// Create multiple groups
	baseGroup := fmt.Sprintf("test-page-%d", time.Now().Unix())
	for i := 0; i < 5; i++ {
		groupName := fmt.Sprintf("%s-%d", baseGroup, i)
		groupID, err := s.client.Groups.Create(s.ctx, groupName, nil)
		s.Require().NoError(err)
		s.trackGroup(groupID)
	}

	// List with pagination
	groups, err := s.client.Groups.ListPaginated(s.ctx, nil, true, 0, 3)
	s.NoError(err)
	s.NotEmpty(groups)

	// Note: May get more or fewer than 3 depending on other groups in realm
	// Just verify we got some groups and no error
}

// TestErrorHandling tests error scenarios
func (s *GroupsIntegrationTestSuite) TestErrorHandling() {
	// Try to get non-existent group
	_, err := s.client.Groups.Get(s.ctx, "non-existent-group-id-12345")
	s.Error(err)
	s.Equal(keycloak.ErrGroupNotFound, err)

	// Try to update non-existent group
	err = s.client.Groups.Update(s.ctx, keycloak.Group{
		ID:   ptr.String("non-existent-id"),
		Name: ptr.String("Test"),
	})
	s.Error(err)

	// Try to delete non-existent group
	err = s.client.Groups.Delete(s.ctx, "non-existent-group-id-12345")
	s.Error(err)

	// Try to create subgroup under non-existent parent
	_, err = s.client.Groups.CreateSubGroup(s.ctx, "non-existent-parent", "subgroup", nil)
	s.Error(err)
}

// TestComplexHierarchy tests multi-level group hierarchy
func (s *GroupsIntegrationTestSuite) TestComplexHierarchy() {
	// Create parent
	parentName := fmt.Sprintf("test-hierarchy-%d", time.Now().Unix())
	parentID, err := s.client.Groups.Create(s.ctx, parentName, nil)
	s.Require().NoError(err)
	s.trackGroup(parentID)

	// Create first level subgroups
	subGroup1Name := fmt.Sprintf("%s-sub1", parentName)
	subGroup1ID, err := s.client.Groups.CreateSubGroup(s.ctx, parentID, subGroup1Name, nil)
	s.Require().NoError(err)
	s.trackGroup(subGroup1ID)

	subGroup2Name := fmt.Sprintf("%s-sub2", parentName)
	subGroup2ID, err := s.client.Groups.CreateSubGroup(s.ctx, parentID, subGroup2Name, nil)
	s.Require().NoError(err)
	s.trackGroup(subGroup2ID)

	// Create nested subgroup (child of subGroup1)
	nestedName := fmt.Sprintf("%s-nested", parentName)
	nestedID, err := s.client.Groups.CreateSubGroup(s.ctx, subGroup1ID, nestedName, nil)
	s.Require().NoError(err)
	s.trackGroup(nestedID)

	// Verify hierarchy
	parent, err := s.client.Groups.Get(s.ctx, parentID)
	s.NoError(err)
	s.NotNil(parent.SubGroups)

	// List subgroups of parent
	subGroups, err := s.client.Groups.ListSubGroupsPaginated(s.ctx, parentID, keycloak.SubGroupSearchParams{})
	s.NoError(err)
	s.GreaterOrEqual(len(subGroups), 2, "Parent should have at least 2 direct subgroups")

	// List subgroups of first subgroup
	nestedSubGroups, err := s.client.Groups.ListSubGroupsPaginated(s.ctx, subGroup1ID, keycloak.SubGroupSearchParams{})
	s.NoError(err)
	s.GreaterOrEqual(len(nestedSubGroups), 1, "First subgroup should have at least 1 subgroup")
}

// Run the suite
func TestGroupsIntegrationTestSuite(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}
	suite.Run(t, new(GroupsIntegrationTestSuite))
}
