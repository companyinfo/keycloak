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

// Package keycloak provides an idiomatic Go client for the Keycloak Admin REST API.
//
// This package offers a clean, type-safe interface for managing Keycloak resources
// with comprehensive error handling, automatic OAuth2 token management, and support
// for all major group management operations.
//
// # Features
//
//   - Group and subgroup management (create, read, update, delete, search)
//   - OAuth2 authentication with automatic token refresh
//   - Paginated list operations with customizable page sizes
//   - Advanced search with filtering and exact matching
//   - Comprehensive error handling with detailed error responses
//   - Configurable timeouts, retries, and debugging
//   - Group member management
//   - Management permissions control
//   - Type-safe API with pointer-based optional fields
//
// # Installation
//
//	go get go.companyinfo.dev/keycloak
//
// # Quick Start
//
// Create a client with minimal configuration:
//
//	client, err := keycloak.New(ctx, keycloak.Config{
//	    URL:          "https://keycloak.example.com",
//	    Realm:        "my-realm",
//	    ClientID:     "admin-cli",
//	    ClientSecret: "secret",
//	})
//	if err != nil {
//	    log.Fatal(err)
//	}
//
// # Advanced Configuration
//
// Configure the client with custom options:
//
//	client, err := keycloak.New(ctx, config,
//	    keycloak.WithPageSize(100),
//	    keycloak.WithTimeout(30*time.Second),
//	    keycloak.WithRetry(3, 5*time.Second, 30*time.Second),
//	    keycloak.WithDebug(true),
//	    keycloak.WithUserAgent("my-app/1.0"),
//	)
//
// # Working with Groups
//
// Create a group:
//
//	attributes := map[string][]string{
//	    "department": {"engineering"},
//	    "location":   {"remote"},
//	}
//	groupID, err := client.Groups.Create(ctx, "Engineering", attributes)
//
// List groups:
//
//	groups, err := client.Groups.List(ctx, nil, false)
//	for _, group := range groups {
//	    fmt.Printf("Group: %s (ID: %s)\n", *group.Name, *group.ID)
//	}
//
// Search groups with parameters:
//
//	params := keycloak.SearchGroupParams{
//	    Search:              keycloak.StringP("Engineering"),
//	    Exact:               keycloak.BoolP(true),
//	    BriefRepresentation: keycloak.BoolP(false),
//	    First:               keycloak.IntP(0),
//	    Max:                 keycloak.IntP(50),
//	}
//	groups, err := client.Groups.ListWithParams(ctx, params)
//
// Get a group by ID:
//
//	group, err := client.Groups.Get(ctx, groupID)
//
// Search by attribute:
//
//	attr := &keycloak.GroupAttribute{
//	    Key:   "department",
//	    Value: "engineering",
//	}
//	group, err := client.Groups.GetByAttribute(ctx, attr)
//
// Update a group:
//
//	group.Description = keycloak.StringP("Updated description")
//	err = client.Groups.Update(ctx, *group)
//
// Delete a group:
//
//	err = client.Groups.Delete(ctx, groupID)
//
// # Working with Subgroups
//
// Create a subgroup:
//
//	subGroupID, err := client.Groups.CreateSubGroup(ctx, parentGroupID, "Team A", attributes)
//
// List subgroups:
//
//	subGroups, err := client.Groups.ListSubGroups(ctx, parentGroupID)
//
// List subgroups with pagination:
//
//	params := keycloak.SubGroupSearchParams{
//	    Search: keycloak.StringP("Team"),
//	    First:  keycloak.IntP(0),
//	    Max:    keycloak.IntP(20),
//	}
//	subGroups, err := client.Groups.ListSubGroupsPaginated(ctx, parentGroupID, params)
//
// Get subgroup by attribute:
//
//	attr := keycloak.GroupAttribute{Key: "team", Value: "alpha"}
//	subGroup, err := client.Groups.GetSubGroupByAttribute(*parentGroup, attr)
//
// # Working with Group Members
//
// List group members:
//
//	params := keycloak.GroupMembersParams{
//	    First: keycloak.IntP(0),
//	    Max:   keycloak.IntP(100),
//	}
//	members, err := client.Groups.ListMembers(ctx, groupID, params)
//	for _, user := range members {
//	    fmt.Printf("User: %s (%s)\n", *user.Username, *user.Email)
//	}
//
// # Pagination
//
// The client supports both automatic and manual pagination:
//
//	// Paginated list with offset and limit
//	groups, err := client.Groups.ListPaginated(ctx, nil, false, 0, 50)
//
//	// Count total groups
//	count, err := client.Groups.Count(ctx, nil, nil)
//
//	// Manual pagination loop
//	pageSize := 50
//	for page := 0; ; page++ {
//	    groups, err := client.Groups.ListPaginated(ctx, nil, false, page*pageSize, pageSize)
//	    if err != nil {
//	        return err
//	    }
//	    if len(groups) == 0 {
//	        break
//	    }
//	    // Process groups...
//	}
//
// # Error Handling
//
// The package provides detailed error information:
//
//	group, err := client.Groups.GetByAttribute(ctx, attr)
//	if err != nil {
//	    if errors.Is(err, keycloak.ErrGroupNotFound) {
//	        // Handle not found case
//	    } else {
//	        // Handle other errors
//	    }
//	}
//
// # Management Permissions
//
// Control group management permissions:
//
//	// Get current permissions
//	perms, err := client.Groups.GetManagementPermissions(ctx, groupID)
//
//	// Enable permissions
//	perms.Enabled = keycloak.BoolP(true)
//	updated, err := client.Groups.UpdateManagementPermissions(ctx, groupID, *perms)
//
// # Helper Functions
//
// The package provides pointer helper functions for working with optional fields:
//
//	str := keycloak.StringP("value")      // Create *string
//	i := keycloak.IntP(42)                // Create *int
//	i32 := keycloak.Int32P(42)            // Create *int32
//	i64 := keycloak.Int64P(42)            // Create *int64
//	b := keycloak.BoolP(true)             // Create *bool
//	value := keycloak.PString(str)        // Dereference safely
//	empty := keycloak.NilOrEmpty(str)     // Check if nil or empty
//
// # Testing
//
// The package includes comprehensive test suites:
//
//   - Unit tests with HTTP mocks (no external dependencies)
//   - Integration tests against real Keycloak instances (requires integration build tag)
//   - Example tests demonstrating common use cases
//
// Run unit tests:
//
//	go test ./...
//
// Run integration tests:
//
//	go test -tags=integration ./...
//
// # Thread Safety
//
// The Client is safe for concurrent use by multiple goroutines. The underlying
// HTTP client handles connection pooling and the OAuth2 token source is thread-safe.
//
// # Best Practices
//
//   - Always pass context for timeout and cancellation control
//   - Use pointer helper functions (StringP, IntP, BoolP) for optional fields
//   - Check for ErrGroupNotFound when searching by attributes
//   - Set appropriate page sizes for large datasets
//   - Enable retry for production environments
//   - Use brief representation when detailed attributes aren't needed
//   - Clean up test groups in integration tests
//
// # Examples
//
// For complete working examples, see the examples directory:
//
//   - examples/basic - Simple group operations
//   - examples/advanced - Advanced configuration and operations
//   - examples/groups - Comprehensive group management
//
// # API Compatibility
//
// This client is compatible with Keycloak 20.0 and later versions.
// It uses the Keycloak Admin REST API endpoints under /admin/realms/{realm}.
//
// # License
//
// Copyright 2025 Company.info B.V.
// Licensed under the Apache License, Version 2.0
package keycloak
