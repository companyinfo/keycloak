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

// Package keycloak_test provides example usage of the keycloak package.
// These examples demonstrate common use cases and best practices for interacting
// with the Keycloak Admin API using this client library.
package keycloak_test

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"go.companyinfo.dev/keycloak"
)

// ExampleNew demonstrates how to create a new Keycloak client.
func ExampleNew() {
	ctx := context.Background()

	config := keycloak.Config{
		URL:          os.Getenv("KEYCLOAK_URL"),
		Realm:        os.Getenv("KEYCLOAK_REALM"),
		ClientID:     os.Getenv("KEYCLOAK_CLIENT_ID"),
		ClientSecret: os.Getenv("KEYCLOAK_CLIENT_SECRET"),
	}

	client, err := keycloak.New(ctx, config)
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}

	// Client is ready to use
	_ = client
}

// ExampleNew_withOptions demonstrates creating a client with functional options.
func ExampleNew_withOptions() {
	ctx := context.Background()

	config := keycloak.Config{
		URL:          os.Getenv("KEYCLOAK_URL"),
		Realm:        os.Getenv("KEYCLOAK_REALM"),
		ClientID:     os.Getenv("KEYCLOAK_CLIENT_ID"),
		ClientSecret: os.Getenv("KEYCLOAK_CLIENT_SECRET"),
	}

	client, err := keycloak.New(ctx, config,
		keycloak.WithPageSize(100),
		keycloak.WithTimeout(30*time.Second),
		keycloak.WithRetry(3, 5*time.Second, 30*time.Second),
	)
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}

	// Client is ready to use with custom configuration
	_ = client
}

// ExampleGroupsClient_Create demonstrates creating a group.
func ExampleGroupsClient_Create() {
	ctx := context.Background()
	client, _ := keycloak.New(ctx, keycloak.Config{
		URL:          os.Getenv("KEYCLOAK_URL"),
		Realm:        os.Getenv("KEYCLOAK_REALM"),
		ClientID:     os.Getenv("KEYCLOAK_CLIENT_ID"),
		ClientSecret: os.Getenv("KEYCLOAK_CLIENT_SECRET"),
	})

	attributes := map[string][]string{
		"description": {"Customer group"},
		"type":        {"organization"},
	}

	groupID, err := client.Groups.Create(ctx, "ACME Corp", attributes)
	if err != nil {
		log.Fatalf("Failed to create group: %v", err)
	}

	fmt.Printf("Created group with ID: %s\n", groupID)
}

// ExampleGroupsClient_List demonstrates listing groups.
func ExampleGroupsClient_List() {
	ctx := context.Background()
	client, _ := keycloak.New(ctx, keycloak.Config{
		URL:          os.Getenv("KEYCLOAK_URL"),
		Realm:        os.Getenv("KEYCLOAK_REALM"),
		ClientID:     os.Getenv("KEYCLOAK_CLIENT_ID"),
		ClientSecret: os.Getenv("KEYCLOAK_CLIENT_SECRET"),
	})

	// List all groups with full details
	groups, err := client.Groups.List(ctx, nil, false)
	if err != nil {
		log.Fatalf("Failed to list groups: %v", err)
	}

	for _, group := range groups {
		fmt.Printf("Group: %s (ID: %s)\n", *group.Name, *group.ID)

		// Show additional fields from GroupRepresentation
		if group.Description != nil {
			fmt.Printf("  Description: %s\n", *group.Description)
		}
		if group.Path != nil {
			fmt.Printf("  Path: %s\n", *group.Path)
		}
		if group.ParentID != nil {
			fmt.Printf("  Parent ID: %s\n", *group.ParentID)
		}
		if group.SubGroupCount != nil {
			fmt.Printf("  Subgroup Count: %d\n", *group.SubGroupCount)
		}
	}
}

// ExampleGroupsClient_ListWithParams demonstrates listing groups with advanced query parameters.
func ExampleGroupsClient_ListWithParams() {
	ctx := context.Background()
	client, _ := keycloak.New(ctx, keycloak.Config{
		URL:          "https://keycloak.example.com",
		Realm:        "my-realm",
		ClientID:     "admin-cli",
		ClientSecret: "secret",
	})

	// Example 1: Search with exact match and populate hierarchy
	searchTerm := "ACME Corp"
	params := keycloak.SearchGroupParams{
		Search:            &searchTerm,
		Exact:             keycloak.BoolP(true),
		PopulateHierarchy: keycloak.BoolP(true),
		SubGroupsCount:    keycloak.BoolP(true),
	}

	groups, err := client.Groups.ListWithParams(ctx, params)
	if err != nil {
		log.Fatalf("Failed to list groups: %v", err)
	}

	for _, group := range groups {
		fmt.Printf("Group: %s (ID: %s)\n", *group.Name, *group.ID)
		if group.SubGroups != nil {
			fmt.Printf("  Has %d subgroups\n", len(*group.SubGroups))
		}
	}

	// Example 2: Paginated search with brief representation
	briefParams := keycloak.SearchGroupParams{
		BriefRepresentation: keycloak.BoolP(true),
		First:               keycloak.IntP(0),
		Max:                 keycloak.IntP(20),
		Q:                   keycloak.StringP("organization"),
	}

	briefGroups, err := client.Groups.ListWithParams(ctx, briefParams)
	if err != nil {
		log.Fatalf("Failed to list groups: %v", err)
	}

	fmt.Printf("Found %d groups matching query\n", len(briefGroups))

	// Example 3: Get all top-level groups without subgroups (no search/q parameter)
	topLevelParams := keycloak.SearchGroupParams{
		BriefRepresentation: keycloak.BoolP(false),
	}

	topLevelGroups, err := client.Groups.ListWithParams(ctx, topLevelParams)
	if err != nil {
		log.Fatalf("Failed to list top-level groups: %v", err)
	}

	fmt.Printf("Retrieved %d top-level groups\n", len(topLevelGroups))
}

// ExampleGroupsClient_GetByAttribute demonstrates searching for groups by attribute.
func ExampleGroupsClient_GetByAttribute() {
	ctx := context.Background()
	client, _ := keycloak.New(ctx, keycloak.Config{
		URL:          os.Getenv("KEYCLOAK_URL"),
		Realm:        os.Getenv("KEYCLOAK_REALM"),
		ClientID:     os.Getenv("KEYCLOAK_CLIENT_ID"),
		ClientSecret: os.Getenv("KEYCLOAK_CLIENT_SECRET"),
	})

	attribute := &keycloak.GroupAttribute{
		Key:   "salesforceID",
		Value: "SF-12345",
	}

	group, err := client.Groups.GetByAttribute(ctx, attribute)
	if err == keycloak.ErrGroupNotFound {
		log.Println("Group not found")
		return
	}
	if err != nil {
		log.Fatalf("Error: %v", err)
	}

	fmt.Printf("Found group: %s\n", *group.Name)
}

// ExampleGroupsClient_CreateSubGroup demonstrates creating subgroups.
func ExampleGroupsClient_CreateSubGroup() {
	ctx := context.Background()
	client, _ := keycloak.New(ctx, keycloak.Config{
		URL:          os.Getenv("KEYCLOAK_URL"),
		Realm:        "my-realm",
		ClientID:     "admin-cli",
		ClientSecret: "secret",
	})

	parentGroupID := "parent-group-id"

	attributes := map[string][]string{
		"type":          {"account"},
		"pricingPlanID": {"premium"},
	}

	subGroupID, err := client.Groups.CreateSubGroup(ctx, parentGroupID, "Premium Account", attributes)
	if err != nil {
		log.Fatalf("Failed to create subgroup: %v", err)
	}

	fmt.Printf("Created subgroup with ID: %s\n", subGroupID)
}

// ExampleGroupsClient_Update demonstrates updating a group.
func ExampleGroupsClient_Update() {
	ctx := context.Background()
	client, _ := keycloak.New(ctx, keycloak.Config{
		URL:          os.Getenv("KEYCLOAK_URL"),
		Realm:        os.Getenv("KEYCLOAK_REALM"),
		ClientID:     os.Getenv("KEYCLOAK_CLIENT_ID"),
		ClientSecret: os.Getenv("KEYCLOAK_CLIENT_SECRET"),
	})

	groupID := "group-id"

	// Get the group
	group, err := client.Groups.Get(ctx, groupID)
	if err != nil {
		log.Fatalf("Failed to get group: %v", err)
	}

	// Update description and attributes
	group.Description = keycloak.StringP("Updated description for the group")

	if group.Attributes == nil {
		attrs := make(map[string][]string)
		group.Attributes = &attrs
	}
	(*group.Attributes)["updated"] = []string{"true"}
	(*group.Attributes)["lastModified"] = []string{"2025-01-01"}

	// Update the group
	err = client.Groups.Update(ctx, *group)
	if err != nil {
		log.Fatalf("Failed to update group: %v", err)
	}

	fmt.Println("Group updated successfully")
}

// ExampleGroupsClient_Delete demonstrates deleting a group.
func ExampleGroupsClient_Delete() {
	ctx := context.Background()
	client, _ := keycloak.New(ctx, keycloak.Config{
		URL:          os.Getenv("KEYCLOAK_URL"),
		Realm:        os.Getenv("KEYCLOAK_REALM"),
		ClientID:     os.Getenv("KEYCLOAK_CLIENT_ID"),
		ClientSecret: os.Getenv("KEYCLOAK_CLIENT_SECRET"),
	})

	groupID := "group-to-delete"

	err := client.Groups.Delete(ctx, groupID)
	if err != nil {
		log.Fatalf("Failed to delete group: %v", err)
	}

	fmt.Println("Group deleted successfully")
}

// ExampleGroupsClient_ListSubGroupsPaginated demonstrates listing subgroups with pagination and search.
func ExampleGroupsClient_ListSubGroupsPaginated() {
	ctx := context.Background()
	client, _ := keycloak.New(ctx, keycloak.Config{
		URL:          os.Getenv("KEYCLOAK_URL"),
		Realm:        os.Getenv("KEYCLOAK_REALM"),
		ClientID:     os.Getenv("KEYCLOAK_CLIENT_ID"),
		ClientSecret: os.Getenv("KEYCLOAK_CLIENT_SECRET"),
	})

	parentGroupID := "parent-group-id"

	// Example 1: Basic pagination
	params := keycloak.SubGroupSearchParams{
		First: keycloak.IntP(0),
		Max:   keycloak.IntP(10),
	}

	subGroups, err := client.Groups.ListSubGroupsPaginated(ctx, parentGroupID, params)
	if err != nil {
		log.Fatalf("Failed to list subgroups: %v", err)
	}

	for _, subGroup := range subGroups {
		fmt.Printf("Subgroup: %s (ID: %s)\n", *subGroup.Name, *subGroup.ID)
	}

	// Example 2: Search with exact match
	searchTerm := "Premium Account"
	exactMatch := true
	searchParams := keycloak.SubGroupSearchParams{
		Search:              &searchTerm,
		Exact:               &exactMatch,
		BriefRepresentation: keycloak.BoolP(false),
		SubGroupsCount:      keycloak.BoolP(true),
	}

	results, err := client.Groups.ListSubGroupsPaginated(ctx, parentGroupID, searchParams)
	if err != nil {
		log.Fatalf("Failed to search subgroups: %v", err)
	}

	fmt.Printf("Found %d subgroups matching search\n", len(results))

	// Example 3: Brief representation with pagination
	briefParams := keycloak.SubGroupSearchParams{
		BriefRepresentation: keycloak.BoolP(true),
		First:               keycloak.IntP(0),
		Max:                 keycloak.IntP(20),
	}

	briefResults, err := client.Groups.ListSubGroupsPaginated(ctx, parentGroupID, briefParams)
	if err != nil {
		log.Fatalf("Failed to list subgroups: %v", err)
	}

	fmt.Printf("Retrieved %d subgroups with brief representation\n", len(briefResults))
}

// ExampleGroupsClient_ListMembers demonstrates listing members (users) of a group.
func ExampleGroupsClient_ListMembers() {
	ctx := context.Background()
	client, _ := keycloak.New(ctx, keycloak.Config{
		URL:          os.Getenv("KEYCLOAK_URL"),
		Realm:        os.Getenv("KEYCLOAK_REALM"),
		ClientID:     os.Getenv("KEYCLOAK_CLIENT_ID"),
		ClientSecret: os.Getenv("KEYCLOAK_CLIENT_SECRET"),
	})

	groupID := "group-id"

	// Example 1: Get all members with full details
	params := keycloak.GroupMembersParams{
		BriefRepresentation: keycloak.BoolP(false),
		First:               keycloak.IntP(0),
		Max:                 keycloak.IntP(100),
	}

	members, err := client.Groups.ListMembers(ctx, groupID, params)
	if err != nil {
		log.Fatalf("Failed to list group members: %v", err)
	}

	for _, user := range members {
		fmt.Printf("User: %s (ID: %s)\n", keycloak.PString(user.Username), keycloak.PString(user.ID))
		if user.Email != nil {
			fmt.Printf("  Email: %s (Verified: %v)\n", *user.Email, user.EmailVerified != nil && *user.EmailVerified)
		}
		if user.FirstName != nil || user.LastName != nil {
			fmt.Printf("  Name: %s %s\n", keycloak.PString(user.FirstName), keycloak.PString(user.LastName))
		}
		if user.Enabled != nil {
			fmt.Printf("  Enabled: %v\n", *user.Enabled)
		}
		if user.Origin != nil {
			fmt.Printf("  Origin: %s\n", *user.Origin)
		}
		if user.FederatedIdentities != nil && len(*user.FederatedIdentities) > 0 {
			fmt.Printf("  Federated Identities: %d\n", len(*user.FederatedIdentities))
		}
	}

	// Example 2: Get members with brief representation (faster)
	briefParams := keycloak.GroupMembersParams{
		BriefRepresentation: keycloak.BoolP(true),
		Max:                 keycloak.IntP(50),
	}

	briefMembers, err := client.Groups.ListMembers(ctx, groupID, briefParams)
	if err != nil {
		log.Fatalf("Failed to list group members: %v", err)
	}

	fmt.Printf("Found %d members in group\n", len(briefMembers))
}

// ExampleGroupsClient_GetManagementPermissions demonstrates getting management permissions for a group.
func ExampleGroupsClient_GetManagementPermissions() {
	ctx := context.Background()
	client, _ := keycloak.New(ctx, keycloak.Config{
		URL:          os.Getenv("KEYCLOAK_URL"),
		Realm:        os.Getenv("KEYCLOAK_REALM"),
		ClientID:     os.Getenv("KEYCLOAK_CLIENT_ID"),
		ClientSecret: os.Getenv("KEYCLOAK_CLIENT_SECRET"),
	})

	groupID := "group-id"

	permissions, err := client.Groups.GetManagementPermissions(ctx, groupID)
	if err != nil {
		log.Fatalf("Failed to get management permissions: %v", err)
	}

	if permissions.Enabled != nil && *permissions.Enabled {
		fmt.Println("Management permissions are enabled")
		if permissions.Resource != nil {
			fmt.Printf("Resource: %s\n", *permissions.Resource)
		}
	} else {
		fmt.Println("Management permissions are disabled")
	}
}

// ExampleGroupsClient_UpdateManagementPermissions demonstrates enabling management permissions for a group.
func ExampleGroupsClient_UpdateManagementPermissions() {
	ctx := context.Background()
	client, _ := keycloak.New(ctx, keycloak.Config{
		URL:          os.Getenv("KEYCLOAK_URL"),
		Realm:        os.Getenv("KEYCLOAK_REALM"),
		ClientID:     os.Getenv("KEYCLOAK_CLIENT_ID"),
		ClientSecret: os.Getenv("KEYCLOAK_CLIENT_SECRET"),
	})

	groupID := "group-id"

	// Enable management permissions
	ref := keycloak.ManagementPermissionReference{
		Enabled: keycloak.BoolP(true),
	}

	result, err := client.Groups.UpdateManagementPermissions(ctx, groupID, ref)
	if err != nil {
		log.Fatalf("Failed to update management permissions: %v", err)
	}

	if result.Enabled != nil && *result.Enabled {
		fmt.Println("Management permissions enabled successfully")
	}

	// Disable management permissions
	disableRef := keycloak.ManagementPermissionReference{
		Enabled: keycloak.BoolP(false),
	}

	result, err = client.Groups.UpdateManagementPermissions(ctx, groupID, disableRef)
	if err != nil {
		log.Fatalf("Failed to update management permissions: %v", err)
	}

	fmt.Println("Management permissions disabled successfully")
}
