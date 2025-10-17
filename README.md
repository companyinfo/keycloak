# gokeycloak

[![Go Reference](https://pkg.go.dev/badge/go.companyinfo.dev/keycloak.svg)](https://pkg.go.dev/go.companyinfo.dev/keycloak)
[![Go Report Card](https://goreportcard.com/badge/go.companyinfo.dev/keycloak)](https://goreportcard.com/report/go.companyinfo.dev/keycloak)
[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](LICENSE)

A production-ready, idiomatic Go client for the Keycloak Admin API.

## Overview

GoKeycloak provides a clean, type-safe interface for interacting with Keycloak's Admin API. Built for production use, it handles authentication, automatic token refresh, retry logic, and provides comprehensive resource management capabilities.

**Why GoKeycloak?**

- **Production Ready**: Battle-tested with automatic token management, retry logic, and comprehensive error handling
- **Type Safe**: Strongly typed models prevent runtime errors and improve code maintainability
- **Modern Go Patterns**: Context support, functional options, and resource-based client design
- **Well Tested**: Extensive test coverage with unit, mock, and integration tests
- **Easy to Use**: Simple API with sensible defaults, yet flexible for advanced use cases

## Table of Contents

- [Features](#features)
- [Requirements](#requirements)
- [Installation](#installation)
- [Quick Start](#quick-start)
- [Usage](#usage)
  - [Basic Setup](#basic-setup)
  - [Advanced Configuration](#advanced-configuration-with-options)
  - [Group Operations](#creating-a-group)
  - [Real-World Examples](#real-world-examples)
- [API Reference](#api-reference)
- [Error Handling](#error-handling)
- [Best Practices](#best-practices)
- [Troubleshooting](#troubleshooting)
- [Testing](#testing)
- [Contributing](#contributing)
- [FAQ](#faq)
- [License](#license)

## Features

- **Authentication**: OAuth2 client credentials flow with automatic token management and refresh
- **Group Management**: Complete CRUD operations for groups and subgroups with attribute support
- **Pagination**: Built-in support for paginated requests with configurable page sizes
- **Type Safety**: Strongly typed models and interfaces prevent runtime errors
- **Context Support**: All operations accept context for cancellation, timeout control, and request tracing
- **Retry Logic**: Configurable exponential backoff for handling transient failures
- **Flexible Configuration**: Functional options pattern for easy customization
- **Production Ready**: Debug logging, custom headers, proxy support, and comprehensive error handling

## Requirements

- **Go**: 1.24 or later
- **Keycloak**: 26.x or later
- **Client Credentials**: A Keycloak client with appropriate permissions:
  - Service accounts enabled
  - At minimum: `view-users`, `manage-users`, `manage-groups` roles
  - For full admin operations: `realm-admin` role

## Installation

```bash
go get go.companyinfo.dev/keycloak
```

## Quick Start

Get up and running in 60 seconds:

```go
package main

import (
    "context"
    "fmt"
    "log"
    "time"
    
    "go.companyinfo.dev/keycloak"
)

func main() {
    ctx := context.Background()
    
    // 1. Create client with production-ready defaults
    client, err := gokeycloak.New(ctx, gokeycloak.Config{
        URL:          "https://keycloak.example.com",
        Realm:        "my-realm",
        ClientID:     "admin-cli",
        ClientSecret: "your-client-secret",
    },
        gokeycloak.WithTimeout(30*time.Second),                    // Prevent hanging
        gokeycloak.WithRetry(3, 1*time.Second, 10*time.Second),   // Handle transient failures
    )
    if err != nil {
        log.Fatalf("Failed to initialize client: %v", err)
    }
    
    // 2. Create a group
    groupID, err := client.Groups.Create(ctx, "Engineering", map[string][]string{
        "department": {"engineering"},
        "location":   {"remote"},
    })
    if err != nil {
        log.Fatalf("Failed to create group: %v", err)
    }
    
    fmt.Printf("✓ Created group: %s\n", groupID)
    
    // 3. List groups
    groups, err := client.Groups.List(ctx, nil, false)
    if err != nil {
        log.Fatalf("Failed to list groups: %v", err)
    }
    
    fmt.Printf("✓ Found %d groups\n", len(groups))
}
```

That's it! Continue reading for advanced features and best practices.

## Usage

### Basic Setup

```go
package main

import (
    "context"
    "log"
    
    "go.companyinfo.dev/keycloak"
)

func main() {
    ctx := context.Background()
    
    config := gokeycloak.Config{
        URL:          "https://keycloak.example.com",
        Realm:        "my-realm",
        ClientID:     "admin-cli",
        ClientSecret: "your-client-secret",
    }
    
    client, err := gokeycloak.New(ctx, config)
    if err != nil {
        log.Fatalf("Failed to create client: %v", err)
    }
    
    // Access resource-specific clients
    // client.Groups - for group operations
    // Future: client.Users, client.Roles, client.Organizations, etc.
}
```

### Advanced Configuration with Options

The client supports functional options for flexible configuration:

```go
import (
    "context"
    "log"
    "time"
    
    "go.companyinfo.dev/keycloak"
)

func main() {
    ctx := context.Background()
    
    config := gokeycloak.Config{
        URL:          "https://keycloak.example.com",
        Realm:        "my-realm",
        ClientID:     "admin-cli",
        ClientSecret: "your-client-secret",
    }
    
    client, err := gokeycloak.New(ctx, config,
        gokeycloak.WithPageSize(100),                                  // Custom page size
        gokeycloak.WithTimeout(30*time.Second),                        // Request timeout
        gokeycloak.WithRetry(3, 5*time.Second, 30*time.Second),       // Retry configuration
        gokeycloak.WithDebug(true),                                    // Enable debug logging
        gokeycloak.WithUserAgent("my-app/1.0"),                        // Custom User-Agent
        gokeycloak.WithHeaders(map[string]string{                      // Custom headers
            "X-Request-ID": "12345",
        }),
    )
    if err != nil {
        log.Fatalf("Failed to create client: %v", err)
    }
    
    // Client is ready with custom configuration
}
```

#### Available Options

- **`WithPageSize(size int)`** - Set default page size for paginated requests (default: 50)
- **`WithTimeout(timeout time.Duration)`** - Set request timeout for all API calls
- **`WithRetry(count int, waitTime, maxWaitTime time.Duration)`** - Configure retry behavior
- **`WithDebug(debug bool)`** - Enable debug logging for requests and responses
- **`WithHeaders(headers map[string]string)`** - Add custom headers to all requests
- **`WithUserAgent(userAgent string)`** - Set custom User-Agent header
- **`WithProxy(proxyURL string)`** - Set proxy URL for all requests
- **`WithHTTPClient(httpClient *http.Client)`** - Use custom HTTP client (advanced)

### Creating a Group

```go
attributes := map[string][]string{
    "description": {"My group description"},
    "type":        {"organization"},
}

groupID, err := client.Groups.Create(ctx, "My Group", attributes)
if err != nil {
    log.Fatalf("Failed to create group: %v", err)
}

log.Printf("Created group with ID: %s", groupID)
```

### Getting Groups

```go
// Get all groups
groups, err := client.Groups.List(ctx, nil, false)
if err != nil {
    log.Fatalf("Failed to get groups: %v", err)
}

// Get groups with search
searchTerm := "My Group"
groups, err := client.Groups.List(ctx, &searchTerm, false)
if err != nil {
    log.Fatalf("Failed to search groups: %v", err)
}

// Get groups with pagination
groups, err := client.Groups.ListPaginated(ctx, nil, false, 0, 10)
if err != nil {
    log.Fatalf("Failed to get paginated groups: %v", err)
}

// Get group by ID
group, err := client.Groups.Get(ctx, groupID)
if err != nil {
    log.Fatalf("Failed to get group: %v", err)
}

// Count groups
count, err := client.Groups.Count(ctx, nil, nil)
if err != nil {
    log.Fatalf("Failed to count groups: %v", err)
}
log.Printf("Total groups: %d", count)
```

### Working with Group Attributes

```go
// Get group by specific attribute
attribute := &gokeycloak.GroupAttribute{
    Key:   "salesforceID",
    Value: "SF-12345",
}

group, err := client.Groups.GetByAttribute(ctx, attribute)
if err != nil {
    log.Fatalf("Failed to find group: %v", err)
}
```

### Managing Subgroups

```go
// Create a subgroup
subGroupID, err := client.Groups.CreateSubGroup(ctx, parentGroupID, "Sub Group", attributes)
if err != nil {
    log.Fatalf("Failed to create subgroup: %v", err)
}

// Get subgroups
subGroups, err := client.Groups.ListSubGroups(ctx, parentGroupID)
if err != nil {
    log.Fatalf("Failed to get subgroups: %v", err)
}

// Count subgroups
count, err := client.Groups.CountSubGroups(ctx, parentGroupID)
if err != nil {
    log.Fatalf("Failed to count subgroups: %v", err)
}

// Get subgroup by ID
subGroup, err := client.Groups.GetSubGroupByID(parentGroup, subGroupID)
if err != nil {
    log.Fatalf("Failed to find subgroup: %v", err)
}
```

### Updating and Deleting Groups

```go
// Update a group
group, err := client.Groups.Get(ctx, groupID)
if err != nil {
    log.Fatalf("Failed to get group: %v", err)
}

// Modify group attributes
(*group.Attributes)["updated"] = []string{"true"}

err = client.Groups.Update(ctx, *group)
if err != nil {
    log.Fatalf("Failed to update group: %v", err)
}

// Delete a group
err = client.Groups.Delete(ctx, groupID)
if err != nil {
    log.Fatalf("Failed to delete group: %v", err)
}
```

### Real-World Examples

#### Example 1: Sync External System with Keycloak Groups

```go
// Sync departments from your HR system to Keycloak
func syncDepartments(ctx context.Context, client *gokeycloak.Client, departments []Department) error {
    for _, dept := range departments {
        // Try to find existing group by external ID
        attr := &gokeycloak.GroupAttribute{
            Key:   "externalID",
            Value: dept.ExternalID,
        }
        
        group, err := client.Groups.GetByAttribute(ctx, attr)
        if err == gokeycloak.ErrGroupNotFound {
            // Create new group
            attributes := map[string][]string{
                "externalID":  {dept.ExternalID},
                "syncedAt":    {time.Now().Format(time.RFC3339)},
                "description": {dept.Description},
            }
            
            _, err := client.Groups.Create(ctx, dept.Name, attributes)
            if err != nil {
                return fmt.Errorf("create group %s: %w", dept.Name, err)
            }
            log.Printf("Created group: %s", dept.Name)
        } else if err != nil {
            return fmt.Errorf("lookup group %s: %w", dept.Name, err)
        } else {
            // Update existing group
            (*group.Attributes)["syncedAt"] = []string{time.Now().Format(time.RFC3339)}
            (*group.Attributes)["description"] = []string{dept.Description}
            
            if err := client.Groups.Update(ctx, *group); err != nil {
                return fmt.Errorf("update group %s: %w", dept.Name, err)
            }
            log.Printf("Updated group: %s", dept.Name)
        }
    }
    return nil
}
```

#### Example 2: Bulk Operations with Proper Error Handling

```go
// Create multiple groups with rollback on failure
func createOrganizationStructure(ctx context.Context, client *gokeycloak.Client) error {
    var createdGroups []string
    defer func() {
        if err := recover(); err != nil {
            // Cleanup on panic
            for _, groupID := range createdGroups {
                _ = client.Groups.Delete(ctx, groupID)
            }
        }
    }()
    
    // Create parent organization
    orgID, err := client.Groups.Create(ctx, "Acme Corp", map[string][]string{
        "type": {"organization"},
    })
    if err != nil {
        return fmt.Errorf("create organization: %w", err)
    }
    createdGroups = append(createdGroups, orgID)
    
    // Create departments
    departments := []string{"Engineering", "Sales", "Marketing"}
    for _, dept := range departments {
        deptID, err := client.Groups.CreateSubGroup(ctx, orgID, dept, map[string][]string{
            "type": {"department"},
        })
        if err != nil {
            // Rollback all created groups
            for _, id := range createdGroups {
                _ = client.Groups.Delete(ctx, id)
            }
            return fmt.Errorf("create department %s: %w", dept, err)
        }
        createdGroups = append(createdGroups, deptID)
    }
    
    log.Printf("Successfully created organization with %d departments", len(departments))
    return nil
}
```

#### Example 3: Context with Timeout and Cancellation

```go
// Graceful shutdown with context cancellation
func processGroupsWithCancellation(ctx context.Context, client *gokeycloak.Client) error {
    // Create a context with timeout
    ctx, cancel := context.WithTimeout(ctx, 5*time.Minute)
    defer cancel()
    
    // Listen for interrupt signal
    sigChan := make(chan os.Signal, 1)
    signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
    
    go func() {
        <-sigChan
        log.Println("Interrupt received, cancelling operations...")
        cancel()
    }()
    
    // Process groups page by page
    first := 0
    max := 100
    
    for {
        select {
        case <-ctx.Done():
            return ctx.Err()
        default:
            groups, err := client.Groups.ListPaginated(ctx, nil, false, first, max)
            if err != nil {
                return fmt.Errorf("list groups (page %d): %w", first/max, err)
            }
            
            if len(groups) == 0 {
                break // No more groups
            }
            
            // Process this batch
            for _, group := range groups {
                if err := processGroup(ctx, client, group); err != nil {
                    return err
                }
            }
            
            first += max
        }
    }
    
    return nil
}
```

#### Example 4: Production Service with Dependency Injection

```go
// Service struct for dependency injection
type KeycloakService struct {
    client *gokeycloak.Client
    logger *slog.Logger
}

func NewKeycloakService(ctx context.Context, cfg Config, logger *slog.Logger) (*KeycloakService, error) {
    client, err := gokeycloak.New(ctx, gokeycloak.Config{
        URL:          cfg.KeycloakURL,
        Realm:        cfg.Realm,
        ClientID:     cfg.ClientID,
        ClientSecret: cfg.ClientSecret,
    },
        gokeycloak.WithTimeout(30*time.Second),
        gokeycloak.WithRetry(3, 1*time.Second, 10*time.Second),
        gokeycloak.WithUserAgent(fmt.Sprintf("myapp/%s", cfg.Version)),
    )
    if err != nil {
        return nil, fmt.Errorf("initialize keycloak client: %w", err)
    }
    
    return &KeycloakService{
        client: client,
        logger: logger,
    }, nil
}

func (s *KeycloakService) GetOrCreateGroup(ctx context.Context, name string) (*gokeycloak.Group, error) {
    s.logger.InfoContext(ctx, "looking up group", "name", name)
    
    // Try to find by name
    groups, err := s.client.Groups.List(ctx, &name, false)
    if err != nil {
        return nil, fmt.Errorf("search groups: %w", err)
    }
    
    for _, group := range groups {
        if *group.Name == name {
            s.logger.InfoContext(ctx, "found existing group", "id", *group.ID)
            return group, nil
        }
    }
    
    // Create if not found
    groupID, err := s.client.Groups.Create(ctx, name, nil)
    if err != nil {
        return nil, fmt.Errorf("create group: %w", err)
    }
    
    s.logger.InfoContext(ctx, "created new group", "id", groupID)
    return s.client.Groups.Get(ctx, groupID)
}
```

## API Reference

### Client Structure

The `Client` struct provides access to resource-specific clients:

```go
type Client struct {
    Groups GroupsClient  // Group management operations
    // Future: Users, Roles, Organizations, etc.
}
```

### GroupsClient Interface

The `GroupsClient` provides methods for managing Keycloak groups:

#### Group Operations

- `Create(ctx, name, attributes) (string, error)` - Create a new group
- `Update(ctx, group) error` - Update an existing group
- `Delete(ctx, groupID) error` - Delete a group
- `Get(ctx, groupID) (*Group, error)` - Get group by ID
- `List(ctx, search, briefRepresentation) ([]*Group, error)` - List all groups
- `ListPaginated(ctx, search, briefRepresentation, first, max) ([]*Group, error)` - Get paginated groups
- `ListWithSubGroups(ctx, searchQuery, briefRepresentation, first, max) ([]*Group, error)` - List groups with subgroups included
- `ListWithParams(ctx, params) ([]*Group, error)` - List groups with full parameter control
- `Count(ctx, search, top) (int, error)` - Get total count of groups
- `GetByAttribute(ctx, attribute) (*Group, error)` - Find group by attribute

#### Subgroup Operations

- `CreateSubGroup(ctx, groupID, name, attributes) (string, error)` - Create a subgroup
- `ListSubGroups(ctx, groupID) ([]*Group, error)` - Get all subgroups
- `ListSubGroupsPaginated(ctx, groupID, params) ([]*Group, error)` - Get paginated subgroups with search
- `CountSubGroups(ctx, groupID) (int, error)` - Get count of subgroups
- `GetSubGroupByID(group, subGroupID) (*Group, error)` - Find subgroup by ID
- `GetSubGroupByAttribute(group, attribute) (*Group, error)` - Find subgroup by attribute

#### Important: Working with Subgroups

**Keycloak API Behavior**: Due to how Keycloak's REST API works, the `SubGroups` field is only populated in group responses when a `search` or `q` query parameter is provided. This is a limitation of Keycloak's API, not this library.

**Two Approaches to Fetch Subgroups**:

1. **Use `ListWithSubGroups()` (Recommended for hierarchies)**:

   ```go
   // Fetches groups with their subgroups included in the response
   groups, err := client.Groups.ListWithSubGroups(ctx, "search-term", false, 0, 100)
   for _, group := range groups {
       if group.SubGroups != nil {
           for _, subgroup := range *group.SubGroups {
               fmt.Printf("Subgroup: %s\n", *subgroup.Name)
           }
       }
   }
   ```

2. **Use `ListSubGroups()` (Explicit subgroup fetch)**:

   ```go
   // First get parent groups
   groups, err := client.Groups.List(ctx, nil, false)
   
   // Then explicitly fetch subgroups for each parent
   for _, group := range groups {
       subgroups, err := client.Groups.ListSubGroups(ctx, *group.ID)
       // Process subgroups...
   }
   ```

**Note**: The `Get()` method does NOT populate the `SubGroups` field. Use `ListSubGroups()` if you need to fetch children of a specific group.

## Models

### Group

```go
type Group struct {
    ID          *string
    Name        *string
    Path        *string
    SubGroups   *[]*Group
    Attributes  *map[string][]string
    Access      *map[string]bool
    ClientRoles *map[string][]string
    RealmRoles  *[]string
}
```

### GroupAttribute

```go
type GroupAttribute struct {
    Key   string
    Value string
}
```

## Error Handling

The package uses standard Go error handling with sentinel errors for common cases:

### Sentinel Errors

```go
import "go.companyinfo.dev/keycloak"

group, err := client.Groups.GetByAttribute(ctx, attribute)
if err == gokeycloak.ErrGroupNotFound {
    log.Println("Group not found")
} else if err != nil {
    log.Fatalf("Unexpected error: %v", err)
}
```

### HTTP Error Handling

```go
// Check for specific HTTP status codes
if err != nil {
    if strings.Contains(err.Error(), "401") {
        // Authentication failed - check credentials
        log.Fatal("Authentication failed. Check your client credentials.")
    } else if strings.Contains(err.Error(), "403") {
        // Permission denied - check client roles
        log.Fatal("Permission denied. Ensure client has required roles.")
    } else if strings.Contains(err.Error(), "409") {
        // Conflict - resource already exists
        log.Println("Resource already exists")
    }
    return err
}
```

### Best Error Handling Practices

```go
// Always wrap errors with context
if err != nil {
    return fmt.Errorf("failed to create group %s: %w", groupName, err)
}

// Use context for timeout control
ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
defer cancel()

group, err := client.Groups.Get(ctx, groupID)
if err != nil {
    if ctx.Err() == context.DeadlineExceeded {
        return fmt.Errorf("operation timed out: %w", err)
    }
    return err
}
```

## Best Practices

### 1. Always Use Context with Timeout

```go
// ✅ Good: Prevents hanging requests
ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
defer cancel()

// ❌ Bad: Can hang indefinitely
ctx := context.Background()
```

### 2. Configure Retry Logic for Production

```go
// ✅ Good: Handles transient failures
client, err := gokeycloak.New(ctx, config,
    gokeycloak.WithRetry(3, 1*time.Second, 10*time.Second),
    gokeycloak.WithTimeout(30*time.Second),
)

// ❌ Bad: No retry, fails on first network hiccup
client, err := gokeycloak.New(ctx, config)
```

### 3. Use Pagination for Large Datasets

```go
// ✅ Good: Memory efficient, handles any size
first := 0
max := 100
for {
    groups, err := client.Groups.ListPaginated(ctx, nil, false, first, max)
    if err != nil || len(groups) == 0 {
        break
    }
    processGroups(groups)
    first += max
}

// ❌ Bad: Loads everything into memory
groups, err := client.Groups.List(ctx, nil, false)
```

### 4. Store Secrets Securely

```go
// ✅ Good: Load from environment or secret manager
config := gokeycloak.Config{
    URL:          os.Getenv("KEYCLOAK_URL"),
    Realm:        os.Getenv("KEYCLOAK_REALM"),
    ClientID:     os.Getenv("KEYCLOAK_CLIENT_ID"),
    ClientSecret: os.Getenv("KEYCLOAK_CLIENT_SECRET"),
}

// ❌ Bad: Hardcoded credentials
config := gokeycloak.Config{
    ClientSecret: "my-secret-12345", // Never do this!
}
```

### 5. Use Structured Logging

```go
// ✅ Good: Structured logs with context
logger.InfoContext(ctx, "creating group",
    "name", groupName,
    "attributes", attributes,
)

// ❌ Bad: Unstructured logs
log.Printf("Creating group %s with attributes %v", groupName, attributes)
```

### 6. Handle Idempotency

```go
// ✅ Good: Check before create
func ensureGroupExists(ctx context.Context, client *gokeycloak.Client, name string) (string, error) {
    // Try to find existing
    groups, err := client.Groups.List(ctx, &name, false)
    if err != nil {
        return "", err
    }
    
    for _, g := range groups {
        if *g.Name == name {
            return *g.ID, nil
        }
    }
    
    // Create if not found
    return client.Groups.Create(ctx, name, nil)
}
```

### 7. Add Request Tracing

```go
// ✅ Good: Add tracing headers
client, err := gokeycloak.New(ctx, config,
    gokeycloak.WithHeaders(map[string]string{
        "X-Request-ID": generateRequestID(),
        "X-Service":    "my-service",
    }),
)
```

### 8. Test with Mocks

```go
// ✅ Good: Use interface for testing
type GroupManager interface {
    Create(ctx context.Context, name string, attrs map[string][]string) (string, error)
    Get(ctx context.Context, id string) (*gokeycloak.Group, error)
}

// Your service depends on interface, not concrete implementation
type MyService struct {
    groups GroupManager
}
```

## Troubleshooting

### Common Issues and Solutions

#### Authentication Failures (401 Unauthorized)

**Problem**: Client cannot authenticate with Keycloak.

**Solutions**:

1. **Verify credentials**:

   ```bash
   # Test with curl
   curl -X POST "https://keycloak.example.com/realms/my-realm/protocol/openid-connect/token" \
     -d "client_id=admin-cli" \
     -d "client_secret=your-secret" \
     -d "grant_type=client_credentials"
   ```

2. **Check client configuration**:

   - Client authentication: Must be ON
   - Service accounts enabled: Must be ON
   - Access Type: confidential

3. **Verify realm name**: Case-sensitive!

#### Permission Denied (403 Forbidden)

**Problem**: Client authenticated but lacks permissions.

**Solutions**:

1. **Check service account roles**:
   - Go to Clients → your client → Service accounts roles
   - Assign necessary roles: `manage-groups`, `view-users`, etc.
   - For full admin: assign `realm-admin` role

2. **Check fine-grained permissions**: Some operations require specific permissions

#### Connection Timeouts

**Problem**: Requests hang or timeout.

**Solutions**:

1. **Add timeout to context**:

   ```go
   ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
   defer cancel()
   ```

2. **Configure client timeout**:

   ```go
   client, err := gokeycloak.New(ctx, config,
       gokeycloak.WithTimeout(30*time.Second),
   )
   ```

3. **Check network connectivity**:

   ```bash
   curl -v https://keycloak.example.com
   ```

#### Certificate Errors (TLS/SSL)

**Problem**: `x509: certificate signed by unknown authority`

**Solutions**:

1. **For development only** - Skip verification (NOT FOR PRODUCTION):

   ```go
   httpClient := &http.Client{
       Transport: &http.Transport{
           TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
       },
   }
   client, err := gokeycloak.New(ctx, config,
       gokeycloak.WithHTTPClient(httpClient),
   )
   ```

2. **For production** - Add CA certificate:

   ```go
   caCert, _ := os.ReadFile("/path/to/ca.crt")
   caCertPool := x509.NewCertPool()
   caCertPool.AppendCertsFromPEM(caCert)
   
   httpClient := &http.Client{
       Transport: &http.Transport{
           TLSClientConfig: &tls.Config{RootCAs: caCertPool},
       },
   }
   ```

#### Group Not Found Errors

**Problem**: Cannot find groups that exist in Keycloak.

**Solutions**:

1. **Check search is case-sensitive**: Use exact name
2. **Use briefRepresentation=false**: Gets full group details including attributes
3. **Check group path**: Groups might be subgroups

#### Rate Limiting (429 Too Many Requests)

**Problem**: Too many requests to Keycloak.

**Solutions**:

1. **Enable retry with backoff**:

   ```go
   client, err := gokeycloak.New(ctx, config,
       gokeycloak.WithRetry(5, 2*time.Second, 30*time.Second),
   )
   ```

2. **Implement request throttling**: Use `time.Ticker` or rate limiter
3. **Batch operations**: Use pagination instead of individual requests

#### Memory Issues with Large Datasets

**Problem**: Application runs out of memory when fetching many groups.

**Solution**: Always use pagination:

```go
// ✅ Good: Process in chunks
first := 0
max := 100
for {
    groups, err := client.Groups.ListPaginated(ctx, nil, false, first, max)
    if err != nil || len(groups) == 0 {
        break
    }
    processGroups(groups) // Process and discard
    first += max
}
```

### Debug Mode

Enable debug logging to see all requests and responses:

```go
client, err := gokeycloak.New(ctx, config,
    gokeycloak.WithDebug(true),
)
```

## FAQ

### General Questions

**Q: Is this library production-ready?**  
A: Yes. It includes automatic token refresh, retry logic, comprehensive error handling, and has been battle-tested in production environments.

**Q: What versions of Keycloak are supported?**  
A: Keycloak 26.x and later. The library is tested against Keycloak 26.

**Q: Does it support Keycloak 25 or earlier?**  
A: It may work, but it's not officially tested.

**Q: Can I use this with Red Hat SSO?**  
A: Yes, Red Hat SSO is based on Keycloak, so this library should work.

### Authentication

**Q: What authentication methods are supported?**  
A: Currently only OAuth2 client credentials flow (service accounts). This is the recommended method for server-to-server communication.

**Q: Can I use username/password authentication?**  
A: Not currently. Client credentials flow is more secure for automated processes.

**Q: How often do tokens refresh?**  
A: Tokens are automatically refreshed before expiration. You don't need to handle this.

### Feature Support

**Q: Does this support user management?**  
A: Not yet. Currently only group management is implemented. User management is planned for a future release.

**Q: Can I manage roles?**  
A: Not yet. Role management is planned for a future release.

**Q: What about realm management?**  
A: Not currently. The library focuses on resource management within a realm.

**Q: Can I create custom attributes?**  
A: Yes! Groups support arbitrary attributes as `map[string][]string`.

### Performance

**Q: How many requests per second can it handle?**  
A: This depends on your Keycloak instance. The library includes retry logic and supports concurrent requests.

**Q: Should I create one client per request or reuse it?**  
A: **Reuse the client**. Create one client at application startup and reuse it. The client maintains connection pools and authentication state.

**Q: Does it support connection pooling?**  
A: Yes, through the underlying `http.Client`. You can customize this with `WithHTTPClient()`.

### Testing and Development

**Q: How do I test code that uses this library?**  
A: The library uses interfaces (`GroupsClient`, etc.) that you can mock. See the test files for examples.

**Q: Can I run tests without a real Keycloak instance?**  
A: Yes. Unit tests and mock suite tests don't require Keycloak. Only integration tests need a real instance.

**Q: Should I test against production Keycloak?**  
A: **Never!** Always use a dedicated test realm. Integration tests can create/delete resources.

### Errors and Edge Cases

**Q: What happens if Keycloak is down?**  
A: Requests will fail after the configured timeout and retry attempts. Use appropriate error handling and monitoring.

**Q: Are operations atomic?**  
A: No. Keycloak API calls are independent. If you need transaction-like behavior, implement compensating operations (see Example 2 above).

**Q: What if I create duplicate groups?**  
A: Keycloak allows groups with the same name. Use attributes (like `externalID`) to enforce uniqueness in your application.

### Configuration Options

**Q: How do I use a proxy?**  
A: Use `WithProxy(proxyURL)` option when creating the client.

**Q: Can I customize HTTP headers?**  
A: Yes, use `WithHeaders(map[string]string{...})` for headers on all requests.

**Q: What's the default timeout?**  
A: There's no default timeout. Always set one with `WithTimeout()` or use context timeout.

## Architecture

The package follows a resource-based client design pattern:

- **Main Client**: Entry point that holds resource-specific clients
- **Resource Clients**: Focused interfaces for each Keycloak resource (Groups, Users, Roles, etc.)
- **Shared State**: Authentication and configuration shared across all resource clients

This design makes it easy to add new Keycloak resources without bloating a single interface, and allows for better organization and testability.

## Configuration

### Required Configuration

The `Config` struct requires the following fields:

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `URL` | string | ✅ | Keycloak server URL (e.g., `https://keycloak.example.com`) |
| `Realm` | string | ✅ | Keycloak realm name |
| `ClientID` | string | ✅ | OAuth2 client ID |
| `ClientSecret` | string | ✅ | OAuth2 client secret |

**Example**:

```go
config := gokeycloak.Config{
    URL:          "https://keycloak.example.com",
    Realm:        "production",
    ClientID:     "backend-service",
    ClientSecret: os.Getenv("KEYCLOAK_CLIENT_SECRET"),
}
```

### Environment-Based Configuration

Best practice: Load configuration from environment variables:

```go
import (
    "os"
    "log"
)

func loadConfig() gokeycloak.Config {
    config := gokeycloak.Config{
        URL:          getEnv("KEYCLOAK_URL", ""),
        Realm:        getEnv("KEYCLOAK_REALM", ""),
        ClientID:     getEnv("KEYCLOAK_CLIENT_ID", ""),
        ClientSecret: getEnv("KEYCLOAK_CLIENT_SECRET", ""),
    }
    
    // Validate required fields
    if config.URL == "" || config.Realm == "" || 
       config.ClientID == "" || config.ClientSecret == "" {
        log.Fatal("Missing required Keycloak configuration")
    }
    
    return config
}

func getEnv(key, fallback string) string {
    if value := os.Getenv(key); value != "" {
        return value
    }
    return fallback
}
```

**Environment Variables**:

```bash
export KEYCLOAK_URL="https://keycloak.example.com"
export KEYCLOAK_REALM="production"
export KEYCLOAK_CLIENT_ID="backend-service"
export KEYCLOAK_CLIENT_SECRET="your-secret-here"
```

### Optional Configuration

Use functional options to customize client behavior:

| Option | Description | Default | Recommendation |
|--------|-------------|---------|----------------|
| `WithPageSize(size int)` | Default page size for pagination | 50 | Use 100-500 for batch operations |
| `WithTimeout(duration)` | Request timeout for API calls | No timeout | **Always set** (e.g., 30s) |
| `WithRetry(count, wait, maxWait)` | Retry behavior with exponential backoff | No retry | Use 3-5 retries for production |
| `WithDebug(bool)` | Enable debug logging | false | Only in development |
| `WithHeaders(map[string]string)` | Add custom headers | None | Use for tracing/correlation IDs |
| `WithUserAgent(string)` | Set custom User-Agent | "" | Include app name/version |
| `WithProxy(proxyURL)` | Configure HTTP proxy | None | As needed for your network |
| `WithHTTPClient(*http.Client)` | Use custom HTTP client | Default | For advanced scenarios only |

### Recommended Production Configuration

```go
client, err := gokeycloak.New(ctx, config,
    // Essential for production
    gokeycloak.WithTimeout(30*time.Second),                    // Prevent hanging
    gokeycloak.WithRetry(3, 1*time.Second, 10*time.Second),   // Handle transient failures
    
    // Recommended for operations and debugging
    gokeycloak.WithUserAgent(fmt.Sprintf("myapp/%s", version)), // Identify your app
    gokeycloak.WithHeaders(map[string]string{
        "X-Service": "backend-api",                             // For tracking
    }),
    
    // Optional based on your needs
    gokeycloak.WithPageSize(100),                              // For bulk operations
)
```

### Advanced HTTP Client Configuration

For custom TLS, proxy, or connection pooling:

```go
import (
    "crypto/tls"
    "net/http"
    "time"
)

// Custom HTTP client with connection pooling
httpClient := &http.Client{
    Transport: &http.Transport{
        MaxIdleConns:        100,
        MaxIdleConnsPerHost: 10,
        IdleConnTimeout:     90 * time.Second,
        TLSClientConfig: &tls.Config{
            MinVersion: tls.VersionTLS12,
        },
    },
    Timeout: 30 * time.Second,
}

client, err := gokeycloak.New(ctx, config,
    gokeycloak.WithHTTPClient(httpClient),
)
```

See the [Advanced Configuration](#advanced-configuration-with-options) section for more usage examples.

## License

This project is licensed under the Apache License 2.0 - see the [LICENSE](LICENSE) file for details.

### What This Means

- ✅ **Commercial use** - Use in commercial products
- ✅ **Modification** - Modify the source code
- ✅ **Distribution** - Distribute the library
- ✅ **Patent use** - Patent grant included
- ⚠️ **Trademark** - No trademark rights granted
- ⚠️ **Liability** - No warranty provided
- ⚠️ **State changes** - Must document modifications

## Testing

The project includes comprehensive test coverage with multiple test types:

### Test Structure

- **Unit Tests** (`*_test.go`) - Test individual components without external dependencies
- **Mock Suite Tests** (`groups_mock_suite_test.go`) - Test API operations using HTTP mocks
- **Integration Tests** (`groups_integration_suite_test.go`) - Test against a real Keycloak instance

### Running Tests

#### Run Unit Tests Only

```bash
# Fast unit tests with mocks
go test -v ./...

# Or with short flag to skip integration tests
go test -v -short ./...
```

#### Run Mock Suite Tests

```bash
# Run just the mock suite
go test -v -run TestGroupsMockSuite ./...
```

#### Run Integration Tests

Integration tests require a running Keycloak instance. Set up your environment first:

```bash
# Copy the example environment file
cp .env.example .env

# Edit .env with your Keycloak credentials
# Then run integration tests
go test -v -tags=integration ./...
```

**Required environment variables for integration tests:**

- `KEYCLOAK_URL` - Keycloak server URL (e.g., `https://keycloak.example.com`)
- `KEYCLOAK_REALM` - Test realm name (use a dedicated test realm!)
- `KEYCLOAK_CLIENT_ID` - Client ID with admin privileges
- `KEYCLOAK_CLIENT_SECRET` - Client secret

**Important:** Always use a dedicated test realm, never run integration tests against production!

#### Run All Tests

```bash
# Run all tests including integration tests
go test -v -tags=integration ./...
```

### Test Coverage

Generate and view test coverage:

```bash
# Generate coverage report
go test -v -cover -coverprofile=coverage.out ./...

# View coverage in browser
go tool cover -html=coverage.out

# Or view in terminal
go tool cover -func=coverage.out
```

### Setting Up Keycloak for Integration Tests

1. **Run Keycloak locally** (using Docker):

   ```bash
   docker run -d \
     --name keycloak-test \
     -p 8080:8080 \
     -e KEYCLOAK_ADMIN=admin \
     -e KEYCLOAK_ADMIN_PASSWORD=admin \
     quay.io/keycloak/keycloak:latest \
     start-dev
   ```

2. **Create a test realm:**
   - Access Keycloak Admin Console at <http://localhost:8080>
   - Create a new realm (e.g., `test-realm`)

3. **Create a client for testing:**
   - Go to Clients → Create
   - Client ID: `admin-cli` (or custom name)
   - Client authentication: ON
   - Service accounts enabled: ON
   - Save

4. **Assign admin roles:**
   - Go to Clients → your client → Service accounts roles
   - Assign Client role: `realm-admin` (or at minimum `manage-groups`)

5. **Get credentials:**
   - Go to Clients → your client → Credentials
   - Copy the client secret

6. **Update .env file:**

   ```env
   KEYCLOAK_URL=http://localhost:8080
   KEYCLOAK_REALM=test-realm
   KEYCLOAK_CLIENT_ID=admin-cli
   KEYCLOAK_CLIENT_SECRET=your-copied-secret
   ```

### Writing Tests

The project uses [testify/suite](https://github.com/stretchr/testify#suite-package) for organized test suites and [testify/assert](https://github.com/stretchr/testify#assert-package) for readable assertions.

#### Example Unit Test

```go
func TestWithPageSize(t *testing.T) {
    tests := []struct {
        name      string
        size      int
        wantErr   bool
        wantValue int
    }{
        {
            name:      "valid page size",
            size:      100,
            wantErr:   false,
            wantValue: 100,
        },
        {
            name:    "invalid page size",
            size:    -1,
            wantErr: true,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            client := &Client{pageSize: defaultSize}
            err := WithPageSize(tt.size)(client)

            if tt.wantErr {
                assert.Error(t, err)
            } else {
                assert.NoError(t, err)
                assert.Equal(t, tt.wantValue, client.pageSize)
            }
        })
    }
}
```

#### Example Mock Suite Test

```go
func (s *GroupsMockSuite) TestGetGroupSuccess() {
    groupID := "test-group-id"
    expectedGroup := &Group{
        ID:   StringP(groupID),
        Name: StringP("Test Group"),
    }

    path := fmt.Sprintf("/admin/realms/%s/groups/%s", s.mockRealm, groupID)
    s.mockJSONResponse(http.MethodGet, path, http.StatusOK, expectedGroup)

    group, err := s.client.Groups.Get(s.ctx, groupID)
    
    s.NoError(err)
    s.NotNil(group)
    s.Equal(*expectedGroup.ID, *group.ID)
}
```

#### Example Integration Test

```go
func (s *GroupsIntegrationTestSuite) TestGroupLifecycle() {
    // Create
    groupID, err := s.client.Groups.Create(s.ctx, "Test Group", nil)
    s.Require().NoError(err)
    s.trackGroup(groupID) // Auto-cleanup

    // Read
    group, err := s.client.Groups.Get(s.ctx, groupID)
    s.NoError(err)
    s.Equal("Test Group", *group.Name)

    // Update
    group.Description = gokeycloak.StringP("Updated")
    err = s.client.Groups.Update(s.ctx, *group)
    s.NoError(err)

    // Delete
    err = s.client.Groups.Delete(s.ctx, groupID)
    s.NoError(err)
}
```

### Continuous Integration

The tests are designed to run in CI/CD pipelines:

```yaml
# Example GitHub Actions workflow
name: Tests
on: [push, pull_request]

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - uses: actions/setup-go@v4
        with:
          go-version: '1.24'
      
      # Run unit and mock tests
      - name: Unit Tests
        run: go test -v -short -cover ./...
      
      # Optional: Run integration tests if Keycloak is available
      - name: Integration Tests
        env:
          KEYCLOAK_URL: ${{ secrets.KEYCLOAK_URL }}
          KEYCLOAK_REALM: ${{ secrets.KEYCLOAK_REALM }}
          KEYCLOAK_CLIENT_ID: ${{ secrets.KEYCLOAK_CLIENT_ID }}
          KEYCLOAK_CLIENT_SECRET: ${{ secrets.KEYCLOAK_CLIENT_SECRET }}
        run: go test -v -tags=integration ./...
        if: env.KEYCLOAK_URL != ''
```

## Contributing

Contributions are welcome! We appreciate your help in making this library better.

### How to Contribute

1. **Fork the repository**
2. **Create a feature branch**: `git checkout -b feature/your-feature-name`
3. **Make your changes** with clear, focused commits
4. **Add tests** for new functionality
5. **Run tests**: `go test -v ./...`
6. **Update documentation** if needed
7. **Submit a pull request**

### Code Standards

- Follow [Effective Go](https://golang.org/doc/effective_go) guidelines
- Use `gofmt` to format your code
- Add meaningful comments for exported functions
- Keep functions small and focused
- Write tests for all new functionality
- Maintain backward compatibility when possible

### Testing Requirements

All contributions must include appropriate tests:

- **Unit tests** for new functions
- **Mock tests** for API interactions
- **Integration tests** for critical paths (when applicable)

Run all tests before submitting:

```bash
# Unit and mock tests
go test -v -short ./...

# With coverage
go test -v -cover -coverprofile=coverage.out ./...

# Integration tests (requires Keycloak)
go test -v -tags=integration ./...
```

### What We're Looking For

**High Priority**:

- Bug fixes with test cases
- Performance improvements
- Documentation improvements
- Additional resource clients (Users, Roles, etc.)

**Welcome**:

- New features with clear use cases
- Code quality improvements
- Example applications

**Please Discuss First**:

- Breaking API changes
- Major architectural changes
- Large new features

### Reporting Issues

When reporting bugs, please include:

1. **Go version**: `go version`
2. **Keycloak version**
3. **Library version**
4. **Minimal reproduction code**
5. **Expected vs actual behavior**
6. **Error messages** (with stack traces if applicable)

### Questions and Support

- **Documentation issues**: Open an issue with the "documentation" label
- **Usage questions**: Check the [FAQ](#faq) first, then open a discussion
- **Bug reports**: Open an issue with reproduction steps
- **Feature requests**: Open an issue describing the use case

### Development Setup

1. Clone the repository:

   ```bash
   git clone https://github.com/companyinfo/gokeycloak.git
   cd gokeycloak
   ```

2. Install dependencies:

   ```bash
   go mod download
   ```

3. Run tests:

   ```bash
   go test -v ./...
   ```

4. (Optional) Set up Keycloak for integration tests:

   ```bash
   # See Testing section for detailed setup
   cp .env.example .env
   # Edit .env with your test Keycloak credentials
   ```

### Code Review Process

1. All submissions require review
2. We aim to review PRs within 3-5 business days
3. Address feedback promptly
4. Once approved, maintainers will merge

Thank you for contributing! 🙏

## Acknowledgments

Built with:

- [go-resty](https://github.com/go-resty/resty) - HTTP client library
- [testify](https://github.com/stretchr/testify) - Testing toolkit

Inspired by the Keycloak community and Go best practices.

---

**Maintained by** [CompanyInfo](https://github.com/companyinfo)

**Found this useful?** Give it a star ⭐ to show your support!
