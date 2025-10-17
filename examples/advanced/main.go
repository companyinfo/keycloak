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

// Package main demonstrates advanced usage of the Keycloak client library.
// This example shows advanced client configuration including custom timeouts,
// retry policies, debugging, and comprehensive group management operations.
package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"go.companyinfo.dev/keycloak"
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

	// Create client with custom timeout
	client, err := keycloak.New(ctx, config,
		keycloak.WithTimeout(30*time.Second),
		keycloak.WithPageSize(100),
	)
	if err != nil {
		log.Fatalf("Failed to create Keycloak client: %v", err)
	}

	// Example: Search for groups with specific criteria
	fmt.Println("Searching for groups...")
	searchTerm := "example"
	params := keycloak.SearchGroupParams{
		Search: keycloak.StringP(searchTerm),
		Max:    keycloak.IntP(5),
	}

	groups, err := client.Groups.ListWithParams(ctx, params)
	if err != nil {
		log.Fatalf("Failed to search groups: %v", err)
	}

	fmt.Printf("Found %d groups matching '%s':\n", len(groups), searchTerm)
	for _, group := range groups {
		displayGroup(group, 0)
	}

	// Example: Create a new group
	if os.Getenv("CREATE_GROUP") == "true" {
		fmt.Println("\nCreating a new group...")
		groupName := "example-group-" + time.Now().Format("20060102-150405")
		attributes := map[string][]string{
			"description": {"Created by example program"},
			"environment": {"development"},
		}

		createdGroupID, err := client.Groups.Create(ctx, groupName, attributes)
		if err != nil {
			log.Printf("Failed to create group: %v", err)
		} else {
			fmt.Printf("Successfully created group with ID: %s\n", createdGroupID)

			// Get the created group
			createdGroup, err := client.Groups.Get(ctx, createdGroupID)
			if err != nil {
				log.Printf("Failed to get created group: %v", err)
			} else {
				fmt.Println("Created group details:")
				displayGroup(createdGroup, 0)
			}
		}
	}

	// Example: Working with subgroups
	fmt.Println("\nListing groups with subgroups:")
	allGroups, err := client.Groups.ListPaginated(ctx, nil, false, 0, 20)
	if err != nil {
		log.Fatalf("Failed to get groups: %v", err)
	}

	for _, group := range allGroups {
		if group.SubGroups != nil && len(*group.SubGroups) > 0 {
			displayGroup(group, 0)
		}
	}
}

func displayGroup(group *keycloak.Group, indent int) {
	prefix := ""
	for i := 0; i < indent; i++ {
		prefix += "  "
	}

	fmt.Printf("%s- %s (ID: %s)\n", prefix, keycloak.PString(group.Name), keycloak.PString(group.ID))

	if group.Path != nil {
		fmt.Printf("%s  Path: %s\n", prefix, keycloak.PString(group.Path))
	}

	if group.Attributes != nil && len(*group.Attributes) > 0 {
		fmt.Printf("%s  Attributes:\n", prefix)
		for key, values := range *group.Attributes {
			fmt.Printf("%s    %s: %v\n", prefix, key, values)
		}
	}

	if group.SubGroups != nil && len(*group.SubGroups) > 0 {
		fmt.Printf("%s  Subgroups:\n", prefix)
		for _, subgroup := range *group.SubGroups {
			displayGroup(subgroup, indent+2)
		}
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
