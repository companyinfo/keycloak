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

// Package main demonstrates comprehensive group management operations.
// This example shows how to create groups, manage subgroups, search by attributes,
// and handle group hierarchies using the Keycloak client library.
package main

import (
	"context"
	"fmt"
	"log"
	"os"

	keycloak "go.companyinfo.dev/keycloak"
)

func main() {
	ctx := context.Background()

	// Get configuration from environment variables
	clientSecret := os.Getenv("KEYCLOAK_CLIENT_SECRET")
	if clientSecret == "" {
		log.Fatal("KEYCLOAK_CLIENT_SECRET environment variable is required")
	}

	config := keycloak.Config{
		URL:          getEnv("KEYCLOAK_URL", "https://keycloak.example.com"),
		Realm:        getEnv("KEYCLOAK_REALM", "master"),
		ClientID:     getEnv("KEYCLOAK_CLIENT_ID", "admin-cli"),
		ClientSecret: clientSecret,
	}

	// Create new Keycloak client
	client, err := keycloak.New(ctx, config)
	if err != nil {
		log.Fatalf("Failed to create Keycloak client: %v", err)
	}

	// ============================================================================
	// APPROACH 1: List groups with subgroups included using ListWithSubGroups
	// ============================================================================
	// IMPORTANT: Keycloak's API only returns subgroups when a search parameter is provided.
	// The ListWithSubGroups convenience method handles this automatically.
	fmt.Println("Fetching groups with subgroups...")

	// Use a search term to filter groups (adjust based on your group names)
	// For broader results, use a common prefix or empty string ""
	searchQuery := "test"

	groups, err := client.Groups.ListWithSubGroups(ctx, searchQuery, false, 0, 100)
	if err != nil {
		log.Fatalf("Failed to get groups: %v", err)
	}

	fmt.Printf("Found %d groups:\n", len(groups))
	for i, group := range groups {
		fmt.Printf("%d. %s (ID: %s, Path: %s)\n",
			i+1,
			keycloak.PString(group.Name),
			keycloak.PString(group.ID),
			keycloak.PString(group.Path))

		// Print subgroups if any (when using search parameter)
		if group.SubGroups != nil && len(*group.SubGroups) > 0 {
			fmt.Printf("   Subgroups (%d):\n", len(*group.SubGroups))
			for _, subgroup := range *group.SubGroups {
				fmt.Printf("   - %s (ID: %s, Path: %s)\n",
					keycloak.PString(subgroup.Name),
					keycloak.PString(subgroup.ID),
					keycloak.PString(subgroup.Path))
			}
		}
	}

	// ============================================================================
	// APPROACH 2: Explicitly fetch subgroups using ListSubGroups
	// ============================================================================
	// If you need to fetch subgroups for a specific parent, use ListSubGroups
	if len(groups) > 0 {
		groupID := keycloak.PString(groups[0].ID)
		fmt.Printf("\n--- Alternative: Fetching subgroups explicitly for group %s ---\n", groupID)

		subGroups, err := client.Groups.ListSubGroups(ctx, groupID)
		if err != nil {
			log.Printf("Failed to list subgroups: %v", err)
		} else {
			fmt.Printf("Found %d direct subgroups:\n", len(subGroups))
			for i, sub := range subGroups {
				fmt.Printf("%d. %s (ID: %s, Path: %s)\n",
					i+1,
					keycloak.PString(sub.Name),
					keycloak.PString(sub.ID),
					keycloak.PString(sub.Path))
			}
		}
	}

	// ============================================================================
	// Get group details
	// ============================================================================
	// Note: Get() method does NOT include subgroups in the response
	// Use ListSubGroups() if you need to fetch children
	if len(groups) > 0 {
		groupID := keycloak.PString(groups[0].ID)
		fmt.Printf("\n--- Get group details for %s ---\n", groupID)

		group, err := client.Groups.Get(ctx, groupID)
		if err != nil {
			log.Printf("Failed to get group: %v", err)
		} else {
			fmt.Printf("Name: %s\n", keycloak.PString(group.Name))
			fmt.Printf("Path: %s\n", keycloak.PString(group.Path))
			if group.Attributes != nil && len(*group.Attributes) > 0 {
				fmt.Printf("Attributes: %+v\n", group.Attributes)
			}
		}
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
