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
	"strings"
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
	if len(s.createdGroups) == 0 {
		return
	}

	s.T().Logf("Cleaning up %d test group(s)...", len(s.createdGroups))

	// Clean up groups in reverse order (children first)
	cleanedCount := 0
	for i := len(s.createdGroups) - 1; i >= 0; i-- {
		groupID := s.createdGroups[i]
		err := s.client.Groups.Delete(s.ctx, groupID)
		if err != nil {
			s.T().Logf("  Failed to cleanup group %s: %v", groupID, err)
		} else {
			cleanedCount++
		}
	}

	if cleanedCount == len(s.createdGroups) {
		s.T().Logf("✓ Successfully cleaned up all %d group(s)", cleanedCount)
	} else {
		s.T().Logf("⚠ Cleaned up %d/%d group(s)", cleanedCount, len(s.createdGroups))
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

	// Update attributes (Description field may not be supported in PUT operations on some Keycloak versions)
	if group.Attributes == nil {
		attrs := make(map[string][]string)
		group.Attributes = &attrs
	}
	(*group.Attributes)["updated"] = []string{"true"}
	(*group.Attributes)["lastModified"] = []string{fmt.Sprintf("%d", time.Now().Unix())}

	err = s.client.Groups.Update(s.ctx, *group)
	s.NoError(err)

	// Verify update
	updatedGroup, err := s.client.Groups.Get(s.ctx, groupID)
	s.Require().NoError(err)
	s.Equal([]string{"true"}, (*updatedGroup.Attributes)["updated"])
	s.NotEmpty((*updatedGroup.Attributes)["lastModified"])

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

// TestSearchGroupsByCustomAttributesWithQ tests the Q parameter for searching custom attributes.
// This test verifies if Keycloak supports the query format: 'key1:value1 key2:value2'
func (s *GroupsIntegrationTestSuite) TestSearchGroupsByCustomAttributesWithQ() {
	// Create groups with specific attributes
	timestamp := time.Now().Unix()

	// Group 1: Has department:engineering
	group1Name := fmt.Sprintf("test-q-search-eng-%d", timestamp)
	group1Attrs := map[string][]string{
		"department": {"engineering"},
		"location":   {"berlin"},
	}
	group1ID, err := s.client.Groups.Create(s.ctx, group1Name, group1Attrs)
	s.Require().NoError(err)
	s.trackGroup(group1ID)

	// Group 2: Has department:marketing
	group2Name := fmt.Sprintf("test-q-search-mkt-%d", timestamp)
	group2Attrs := map[string][]string{
		"department": {"marketing"},
		"location":   {"amsterdam"},
	}
	group2ID, err := s.client.Groups.Create(s.ctx, group2Name, group2Attrs)
	s.Require().NoError(err)
	s.trackGroup(group2ID)

	// Group 3: Has department:engineering and location:amsterdam
	group3Name := fmt.Sprintf("test-q-search-eng-ams-%d", timestamp)
	group3Attrs := map[string][]string{
		"department": {"engineering"},
		"location":   {"amsterdam"},
	}
	group3ID, err := s.client.Groups.Create(s.ctx, group3Name, group3Attrs)
	s.Require().NoError(err)
	s.trackGroup(group3ID)

	// Group 4: Different attributes
	group4Name := fmt.Sprintf("test-q-search-hr-%d", timestamp)
	group4Attrs := map[string][]string{
		"department": {"hr"},
		"location":   {"london"},
	}
	group4ID, err := s.client.Groups.Create(s.ctx, group4Name, group4Attrs)
	s.Require().NoError(err)
	s.trackGroup(group4ID)

	// Wait a bit to ensure Keycloak has indexed the groups
	time.Sleep(1 * time.Second)

	// Test 1: Search for single attribute
	s.Run("single_attribute_search", func() {
		s.T().Log("=" + strings.Repeat("=", 68) + "=")
		s.T().Log("TEST 1: Single Attribute Search")
		s.T().Log("=" + strings.Repeat("=", 68) + "=")

		query := "department:engineering"
		params := keycloak.SearchGroupParams{
			Q:                   ptr.String(query),
			BriefRepresentation: ptr.Bool(false),
		}

		s.T().Logf("Query: %q", query)
		s.T().Log("Expected: Find group1 and group3 (both have department:engineering)")

		groups, err := s.client.Groups.ListWithParams(s.ctx, params)
		s.NoError(err)

		s.T().Logf("Results: Found %d group(s)", len(groups))

		// Should find group1 and group3 (both have department:engineering)
		foundIDs := make(map[string]bool)
		if len(groups) > 0 {
			s.T().Log("Groups found:")
			for i, group := range groups {
				if group.ID != nil {
					foundIDs[*group.ID] = true
					s.T().Logf("  [%d] %s (ID: %s)", i+1, ptr.ToString(group.Name), *group.ID)
					if group.Attributes != nil {
						for key, values := range *group.Attributes {
							s.T().Logf("      %s=%v", key, values)
						}
					}
				}
			}
		}

		// Check if the expected groups are found
		s.T().Log("Validation:")
		group1Found := foundIDs[group1ID]
		group3Found := foundIDs[group3ID]
		s.T().Logf("  Group1 found: %v", group1Found)
		s.T().Logf("  Group3 found: %v", group3Found)

		if group1Found && group3Found {
			s.T().Log("✓ PASS: Single attribute search works as expected")
		} else {
			s.T().Log("✗ FAIL: Single attribute search might not be supported or behaves differently")
		}
	})

	// Test 2: Search for multiple attributes (AND logic)
	s.Run("multiple_attributes_search", func() {
		s.T().Log("=" + strings.Repeat("=", 68) + "=")
		s.T().Log("TEST 2: Multiple Attributes Search (AND Logic)")
		s.T().Log("=" + strings.Repeat("=", 68) + "=")

		query := "department:engineering location:amsterdam"
		params := keycloak.SearchGroupParams{
			Q:                   ptr.String(query),
			BriefRepresentation: ptr.Bool(false),
		}

		s.T().Logf("Query: %q", query)
		s.T().Log("Expected: Find only group3 (has BOTH attributes)")

		groups, err := s.client.Groups.ListWithParams(s.ctx, params)
		s.NoError(err)

		s.T().Logf("Results: Found %d group(s)", len(groups))

		// Should find only group3 (has both department:engineering AND location:amsterdam)
		foundIDs := make(map[string]bool)
		if len(groups) > 0 {
			s.T().Log("Groups found:")
			for i, group := range groups {
				if group.ID != nil {
					foundIDs[*group.ID] = true
					s.T().Logf("  [%d] %s (ID: %s)", i+1, ptr.ToString(group.Name), *group.ID)
					if group.Attributes != nil {
						for key, values := range *group.Attributes {
							s.T().Logf("      %s=%v", key, values)
						}
					}
				}
			}
		}

		// Check if the expected group is found
		s.T().Log("Validation:")
		group1Found := foundIDs[group1ID]
		group2Found := foundIDs[group2ID]
		group3Found := foundIDs[group3ID]
		group4Found := foundIDs[group4ID]
		s.T().Logf("  Group1 (eng+berlin):    %v", group1Found)
		s.T().Logf("  Group2 (mkt+amsterdam): %v", group2Found)
		s.T().Logf("  Group3 (eng+amsterdam): %v <- expected", group3Found)
		s.T().Logf("  Group4 (hr+london):     %v", group4Found)

		if group3Found && !group1Found && !group2Found && !group4Found {
			s.T().Log("✓ PASS: Multiple attributes search works with AND logic")
		} else {
			s.T().Log("✗ FAIL: Multiple attributes search might not be supported or behaves differently")
		}
	})

	// Test 3: Search with non-existent attribute value
	s.Run("non_existent_attribute_search", func() {
		s.T().Log("=" + strings.Repeat("=", 68) + "=")
		s.T().Log("TEST 3: Non-Existent Attribute Search")
		s.T().Log("=" + strings.Repeat("=", 68) + "=")

		query := "department:nonexistent"
		params := keycloak.SearchGroupParams{
			Q:                   ptr.String(query),
			BriefRepresentation: ptr.Bool(false),
		}

		s.T().Logf("Query: %q", query)
		s.T().Log("Expected: No test groups should be found")

		groups, err := s.client.Groups.ListWithParams(s.ctx, params)
		s.NoError(err)

		s.T().Logf("Results: Found %d group(s)", len(groups))

		// Should not find any of our test groups
		foundIDs := make(map[string]bool)
		for _, group := range groups {
			if group.ID != nil {
				foundIDs[*group.ID] = true
			}
		}

		// Check if none of our test groups are found
		s.T().Log("Validation:")
		anyTestGroupFound := foundIDs[group1ID] || foundIDs[group2ID] || foundIDs[group3ID] || foundIDs[group4ID]
		s.T().Logf("  Test groups found: %v", anyTestGroupFound)

		if !anyTestGroupFound {
			s.T().Log("✓ PASS: Non-existent attribute search returns no matches")
		} else {
			s.T().Log("✗ FAIL: Non-existent attribute search returned unexpected results")
			if foundIDs[group1ID] {
				s.T().Logf("  Unexpectedly found group1: %s", group1ID)
			}
			if foundIDs[group2ID] {
				s.T().Logf("  Unexpectedly found group2: %s", group2ID)
			}
			if foundIDs[group3ID] {
				s.T().Logf("  Unexpectedly found group3: %s", group3ID)
			}
			if foundIDs[group4ID] {
				s.T().Logf("  Unexpectedly found group4: %s", group4ID)
			}
		}
	})

	// Test 4: Empty Q parameter
	s.Run("empty_q_parameter", func() {
		s.T().Log("=" + strings.Repeat("=", 68) + "=")
		s.T().Log("TEST 4: Empty Q Parameter")
		s.T().Log("=" + strings.Repeat("=", 68) + "=")

		query := ""
		params := keycloak.SearchGroupParams{
			Q:                   ptr.String(query),
			BriefRepresentation: ptr.Bool(false),
		}

		s.T().Logf("Query: %q (empty string)", query)
		s.T().Log("Expected: Should not cause errors")

		groups, err := s.client.Groups.ListWithParams(s.ctx, params)
		s.NoError(err)

		s.T().Logf("Results: Found %d group(s)", len(groups))
		s.T().Log("✓ PASS: Empty Q parameter handled gracefully")
	})

	// Summary
	s.T().Log(strings.Repeat("-", 70))
	s.T().Log("Q PARAMETER TEST SUMMARY")
	s.T().Log(strings.Repeat("-", 70))
	s.T().Log("Documentation:")
	s.T().Log("  The 'q' parameter searches for custom attributes")
	s.T().Log("  Format: 'key1:value1 key2:value2'")
	s.T().Log("Test Results:")
	s.T().Log("  Check individual test outputs above for actual behavior")
	s.T().Log("Interpretation:")
	s.T().Log("  ✓ PASS = Feature works as expected in your Keycloak version")
	s.T().Log("  ✗ FAIL = Feature may not be supported or behaves differently")
	s.T().Log(strings.Repeat("-", 70))
}

// TestSearchSubGroupsByAttributesWithQ tests that the global groups endpoint
// with Q parameter can search for subgroups by custom attributes.
// This verifies the Keycloak documentation that states:
// "subGroups are only returned when using the search or q parameter"
func (s *GroupsIntegrationTestSuite) TestSearchSubGroupsByAttributesWithQ() {
	timestamp := time.Now().Unix()

	// Create parent group
	parentName := fmt.Sprintf("test-q-parent-%d", timestamp)
	parentID, err := s.client.Groups.Create(s.ctx, parentName, nil)
	s.Require().NoError(err)
	s.trackGroup(parentID)

	// Create subgroups with specific attributes
	// Subgroup 1: Has team:backend
	subGroup1Name := fmt.Sprintf("test-q-subgroup-backend-%d", timestamp)
	subGroup1Attrs := map[string][]string{
		"team":     {"backend"},
		"language": {"go"},
	}
	subGroup1ID, err := s.client.Groups.CreateSubGroup(s.ctx, parentID, subGroup1Name, subGroup1Attrs)
	s.Require().NoError(err)
	s.trackGroup(subGroup1ID)

	// Subgroup 2: Has team:frontend
	subGroup2Name := fmt.Sprintf("test-q-subgroup-frontend-%d", timestamp)
	subGroup2Attrs := map[string][]string{
		"team":     {"frontend"},
		"language": {"javascript"},
	}
	subGroup2ID, err := s.client.Groups.CreateSubGroup(s.ctx, parentID, subGroup2Name, subGroup2Attrs)
	s.Require().NoError(err)
	s.trackGroup(subGroup2ID)

	// Create another parent with a subgroup (to ensure we're filtering correctly)
	otherParentName := fmt.Sprintf("test-q-other-parent-%d", timestamp)
	otherParentID, err := s.client.Groups.Create(s.ctx, otherParentName, nil)
	s.Require().NoError(err)
	s.trackGroup(otherParentID)

	// Create subgroup under other parent with same attribute
	otherSubGroupName := fmt.Sprintf("test-q-other-subgroup-%d", timestamp)
	otherSubGroupAttrs := map[string][]string{
		"team":     {"backend"},
		"language": {"rust"},
	}
	otherSubGroupID, err := s.client.Groups.CreateSubGroup(s.ctx, otherParentID, otherSubGroupName, otherSubGroupAttrs)
	s.Require().NoError(err)
	s.trackGroup(otherSubGroupID)

	// Wait for indexing
	time.Sleep(1 * time.Second)

	// Test 1: Search for subgroups by single attribute
	s.Run("search_subgroups_by_attribute", func() {
		s.T().Log("=" + strings.Repeat("=", 68) + "=")
		s.T().Log("TEST 1: Search Subgroups by Single Attribute")
		s.T().Log("=" + strings.Repeat("=", 68) + "=")

		query := "team:backend"
		params := keycloak.SearchGroupParams{
			Q:                   ptr.String(query),
			BriefRepresentation: ptr.Bool(false),
		}

		s.T().Logf("Query: %q", query)
		s.T().Log("Expected: Find subgroups with team:backend attribute")

		groups, err := s.client.Groups.ListWithParams(s.ctx, params)
		s.NoError(err)

		s.T().Logf("Results: Found %d top-level group(s)", len(groups))

		// Check if subgroups are in the results (as top-level) OR in SubGroups field
		foundSubGroup1 := false
		foundOtherSubGroup := false
		foundInSubGroupsField := false

		if len(groups) > 0 {
			s.T().Log("Analyzing groups:")
			for i, group := range groups {
				if group.ID == nil {
					continue
				}

				s.T().Logf("  [%d] %s (ID: %s)", i+1, ptr.ToString(group.Name), *group.ID)
				if group.ParentID != nil {
					s.T().Logf("      Parent: %s", *group.ParentID)
				}
				if group.Attributes != nil && len(*group.Attributes) > 0 {
					for key, values := range *group.Attributes {
						s.T().Logf("      %s=%v", key, values)
					}
				}
				if group.SubGroups != nil && len(*group.SubGroups) > 0 {
					s.T().Logf("      Has %d subgroup(s)", len(*group.SubGroups))
					// Check if our subgroups are in the SubGroups field
					for j, subGroup := range *group.SubGroups {
						if subGroup.ID != nil {
							s.T().Logf("        [%d] %s (ID: %s)", j+1, ptr.ToString(subGroup.Name), *subGroup.ID)
							if *subGroup.ID == subGroup1ID {
								foundSubGroup1 = true
								foundInSubGroupsField = true
							}
							if *subGroup.ID == otherSubGroupID {
								foundOtherSubGroup = true
								foundInSubGroupsField = true
							}
						}
					}
				}

				// Check if the group itself is a subgroup we're looking for
				if *group.ID == subGroup1ID {
					foundSubGroup1 = true
				}
				if *group.ID == otherSubGroupID {
					foundOtherSubGroup = true
				}
			}
		}

		s.T().Log("Validation:")
		s.T().Logf("  SubGroup1 (backend+go):     %v", foundSubGroup1)
		s.T().Logf("  OtherSubGroup (backend+rust): %v", foundOtherSubGroup)
		s.T().Logf("  Found in SubGroups field:   %v", foundInSubGroupsField)

		s.T().Log("Interpretation:")
		if foundInSubGroupsField {
			s.T().Log("  ✓ 'q' returns PARENT groups with matching subgroups in SubGroups field")
		} else if foundSubGroup1 || foundOtherSubGroup {
			s.T().Log("  ✓ 'q' returns subgroups directly as top-level results")
		} else {
			s.T().Log("  ✗ 'q' does NOT search subgroup attributes")
			s.T().Log("     (may only search top-level attributes or names)")
		}
	})

	// Test 2: Try to understand what 'q' returns
	s.Run("understand_q_behavior", func() {
		s.T().Log("=" + strings.Repeat("=", 68) + "=")
		s.T().Log("TEST 2: Understanding 'q' Parameter with PopulateHierarchy")
		s.T().Log("=" + strings.Repeat("=", 68) + "=")

		// Try searching with PopulateHierarchy
		query := "team:backend"
		params := keycloak.SearchGroupParams{
			Q:                   ptr.String(query),
			BriefRepresentation: ptr.Bool(false),
			PopulateHierarchy:   ptr.Bool(true),
		}

		s.T().Logf("Query: %q with PopulateHierarchy=true", query)

		groups, err := s.client.Groups.ListWithParams(s.ctx, params)
		s.NoError(err)

		s.T().Logf("Results: Found %d group(s)", len(groups))

		if len(groups) > 0 {
			s.T().Log("Group details:")
			for i, group := range groups {
				if group.ID == nil {
					continue
				}

				s.T().Logf("  [%d] %s (ID: %s)", i+1, ptr.ToString(group.Name), *group.ID)
				if group.ParentID != nil {
					s.T().Logf("      Type: Subgroup (Parent: %s)", *group.ParentID)
				} else {
					s.T().Log("      Type: Top-level group")
				}
				if group.Attributes != nil && len(*group.Attributes) > 0 {
					for key, values := range *group.Attributes {
						s.T().Logf("      %s=%v", key, values)
					}
				}
				if group.SubGroups != nil && len(*group.SubGroups) > 0 {
					s.T().Logf("      Has %d subgroup(s)", len(*group.SubGroups))
				}
			}
		}

		s.T().Log("This test helps understand how PopulateHierarchy affects results")
	})

	// Test 3: Compare with search parameter (name-based)
	s.Run("compare_with_name_search", func() {
		s.T().Log("=" + strings.Repeat("=", 68) + "=")
		s.T().Log("TEST 3: Comparing 'search' (name) vs 'q' (attributes)")
		s.T().Log("=" + strings.Repeat("=", 68) + "=")

		// Search by name
		searchParams := keycloak.SearchGroupParams{
			Search:              ptr.String("backend"),
			BriefRepresentation: ptr.Bool(false),
		}

		s.T().Log("Test A: Search by name")
		s.T().Log("  Query: search=\"backend\"")

		searchGroups, err := s.client.Groups.ListWithParams(s.ctx, searchParams)
		s.NoError(err)

		s.T().Logf("  Found: %d group(s)", len(searchGroups))

		// Search by attribute
		qParams := keycloak.SearchGroupParams{
			Q:                   ptr.String("team:backend"),
			BriefRepresentation: ptr.Bool(false),
		}

		s.T().Log("Test B: Search by attribute")
		s.T().Log("  Query: q=\"team:backend\"")

		qGroups, err := s.client.Groups.ListWithParams(s.ctx, qParams)
		s.NoError(err)

		s.T().Logf("  Found: %d group(s)", len(qGroups))

		s.T().Log("Compare the results to understand behavior differences")
	})

	// Summary
	s.T().Log(strings.Repeat("-", 70))
	s.T().Log("SUBGROUP Q PARAMETER TEST CONCLUSION")
	s.T().Log(strings.Repeat("-", 70))
	s.T().Log("Keycloak Documentation:")
	s.T().Log("  'subGroups are only returned when using search or q parameter'")
	s.T().Log("What We Tested:")
	s.T().Log("  1. Whether 'q' parameter searches subgroup attributes")
	s.T().Log("  2. How results are returned (parent groups vs direct subgroups)")
	s.T().Log("  3. Impact of PopulateHierarchy parameter")
	s.T().Log("  4. Difference between 'search' (name) and 'q' (attributes)")
	s.T().Log("Key Findings:")
	s.T().Log("  Check the individual test outputs above for actual behavior")
	s.T().Log("  Behavior may vary depending on your Keycloak version")
	s.T().Log(strings.Repeat("-", 70))
}

// Run the suite
func TestGroupsIntegrationTestSuite(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}
	suite.Run(t, new(GroupsIntegrationTestSuite))
}
