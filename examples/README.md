# Examples

This directory contains standalone examples demonstrating how to use the `go.companyinfo.dev/keycloak` library.

## Prerequisites

Before running these examples, you need:

1. A running Keycloak instance
2. Valid credentials (client ID, client secret, username, password)
3. Appropriate permissions to perform the operations

## Configuration

All examples use environment variables for configuration:

```bash
export KEYCLOAK_URL="https://your-keycloak-instance.com"
export KEYCLOAK_REALM="your-realm"
export KEYCLOAK_CLIENT_ID="your-client-id"
export KEYCLOAK_CLIENT_SECRET="your-client-secret"
export KEYCLOAK_USERNAME="your-username"
export KEYCLOAK_PASSWORD="your-password"
```

## Available Examples

### Basic Example

Demonstrates basic client initialization and connection to Keycloak.

```bash
cd examples/basic
go run main.go
```

### Groups Example

Shows how to list groups, get group details, and work with subgroups.

```bash
cd examples/groups
go run main.go
```

### Advanced Example

Demonstrates advanced features including:

- Custom timeouts
- Searching for groups with filters
- Creating new groups (set `CREATE_GROUP=true`)
- Working with group attributes
- Recursive subgroup handling

```bash
cd examples/advanced
go run main.go

# To enable group creation:
CREATE_GROUP=true go run main.go
```

## Running Examples with Go Modules

These examples use the parent module, so they work out of the box:

```bash
cd examples/basic
go run main.go
```

If you want to test against a local version of the library during development, the examples automatically use the parent module via the `go.mod` replace directive.

## Error Handling

All examples include proper error handling and logging. If you encounter errors:

1. Verify your Keycloak instance is accessible
2. Check your credentials are correct
3. Ensure your user/client has appropriate permissions
4. Review the error messages for specific issues

## Contributing

When adding new examples:

1. Create a new directory under `examples/`
2. Include a descriptive `main.go` file
3. Use environment variables for configuration
4. Add proper error handling and logging
5. Include comments explaining the example's purpose
6. Update this README with the new example
