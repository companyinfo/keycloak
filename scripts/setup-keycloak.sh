#!/usr/bin/env bash
# ============================================================================
# Keycloak Setup Script for Integration Testing
# ============================================================================
#
# This script automates the complete setup of a Keycloak instance for
# integration testing. It creates a test realm, configures a service account
# client with appropriate permissions, and validates the configuration.
#
# Features:
#   - Beautiful CLI UX with Gum (spinners, styled output, structured logging)
#   - Robust HTTP health checking with wait4x
#   - Comprehensive error handling and validation
#   - Interactive or non-interactive modes
#   - Idempotent operations (safe to run multiple times)
#
# Prerequisites:
#   - Keycloak instance accessible at the configured URL
#   - Required tools: curl, jq, gum, wait4x
#   - Valid admin credentials for Keycloak
#
# Usage:
#   ./scripts/setup-keycloak.sh
#
# Environment Variables:
#   KEYCLOAK_URL              - Keycloak base URL (default: http://localhost:8080)
#   KEYCLOAK_ADMIN_USER       - Admin username (default: admin)
#   KEYCLOAK_ADMIN_PASSWORD   - Admin password (default: admin)
#   TEST_REALM                - Realm name (default: test-realm)
#   TEST_CLIENT_ID            - Client ID (default: test-client)
#   TEST_CLIENT_SECRET        - Client secret (default: test-secret-12345)
#   NON_INTERACTIVE           - Skip prompts (default: false)
#   SKIP_REALM_CREATION       - Skip realm creation (default: false)
#
# Script Steps:
#   1. Display styled header and configuration
#   2. Wait for Keycloak to become available (with wait4x)
#   3. Authenticate and obtain admin access token
#   4. Check if test realm already exists
#   5. Create or reuse test realm based on user preference
#   6. Create service account client with credentials
#   7. Retrieve client UUID for role assignment
#   8. Get service account user ID
#   9. Assign admin roles (manage-groups, view-users, etc.)
#   10. Test the configuration with client credentials flow
#   11. Create .env file with configuration
#   12. Display success summary with next steps
#
# Exit Codes:
#   0  - Success
#   1  - Configuration error or operation failure
#
# ============================================================================

set -euo pipefail

# ============================================================================
# Configuration - Load from environment variables or use defaults
# ============================================================================
KEYCLOAK_URL="${KEYCLOAK_URL:-http://localhost:8080}"
KEYCLOAK_ADMIN_USER="${KEYCLOAK_ADMIN_USER:-admin}"
KEYCLOAK_ADMIN_PASSWORD="${KEYCLOAK_ADMIN_PASSWORD:-admin}"
TEST_REALM="${TEST_REALM:-test-realm}"
TEST_CLIENT_ID="${TEST_CLIENT_ID:-test-client}"
TEST_CLIENT_SECRET="${TEST_CLIENT_SECRET:-test-secret-12345}"
SKIP_REALM_CREATION="${SKIP_REALM_CREATION:-false}"
NON_INTERACTIVE="${NON_INTERACTIVE:-false}"

# ============================================================================
# STEP 1: Display styled header and configuration
# ============================================================================
gum style \
  --foreground 212 \
  --border-foreground 212 \
  --border double \
  --align center \
  --width 60 \
  --margin "1 0" \
  --padding "1 2" \
  "Keycloak Setup" "Integration Testing Configuration"

gum log --structured --level info "Configuration" url="$KEYCLOAK_URL" realm="$TEST_REALM" client="$TEST_CLIENT_ID"
echo ""

# ============================================================================
# STEP 2: Wait for Keycloak to become available
# ============================================================================
# Uses wait4x for robust HTTP health checking with timeout and retry logic
gum log --level info "Waiting for Keycloak to be ready..."

if gum spin --spinner dot --title "Connecting to Keycloak..." -- \
  wait4x http "$KEYCLOAK_URL/realms/master" --timeout 2m --interval 2s; then
  gum log --level info "Keycloak is ready"
else
  gum log --level error "Keycloak failed to become ready in time"
  exit 1
fi

# ============================================================================
# STEP 3: Authenticate and obtain admin access token
# ============================================================================
# Uses password grant type with admin-cli client to get access token
echo ""
gum log --level info "Obtaining admin access token..."

ADMIN_TOKEN=$(gum spin --spinner dot --title "Authenticating..." -- \
  curl -s -X POST "$KEYCLOAK_URL/realms/master/protocol/openid-connect/token" \
  -H "Content-Type: application/x-www-form-urlencoded" \
  -d "username=$KEYCLOAK_ADMIN_USER" \
  -d "password=$KEYCLOAK_ADMIN_PASSWORD" \
  -d "grant_type=password" \
  -d "client_id=admin-cli" | jq -r '.access_token')

if [ "$ADMIN_TOKEN" == "null" ] || [ -z "$ADMIN_TOKEN" ]; then
  gum log --level error "Failed to obtain admin token" username="$KEYCLOAK_ADMIN_USER"
  gum log --level error "Check admin credentials"
  exit 1
fi
gum log --level info "Admin token obtained successfully"

# ============================================================================
# STEP 4: Check if test realm already exists
# ============================================================================
# Query Keycloak API - 404 is expected and normal when realm doesn't exist
echo ""
gum log --level info "Checking for existing test realm..."
# Note: Don't use -f flag here - 404 is expected when realm doesn't exist
EXISTING_REALM=$(curl -s -X GET "$KEYCLOAK_URL/admin/realms/$TEST_REALM" \
  -H "Authorization: Bearer $ADMIN_TOKEN" \
  -H "Accept: application/json" 2>/dev/null | jq -r '.realm // empty')

if [ "$EXISTING_REALM" == "$TEST_REALM" ]; then
  gum log --level warn "Test realm already exists" realm="$TEST_REALM"
  
  if [ "$NON_INTERACTIVE" == "true" ]; then
    gum log --level info "Running in non-interactive mode: using existing realm"
    SKIP_REALM_CREATION="true"
  else
    if gum confirm "Delete and recreate realm '$TEST_REALM'?"; then
      gum log --level info "Deleting existing realm..."
      if gum spin --spinner dot --title "Removing realm..." -- \
        curl -sf -X DELETE "$KEYCLOAK_URL/admin/realms/$TEST_REALM" \
        -H "Authorization: Bearer $ADMIN_TOKEN"; then
        gum log --level info "Realm deleted successfully"
      else
        gum log --level error "Failed to delete realm"
        exit 1
      fi
    else
      gum log --level info "Using existing realm"
      SKIP_REALM_CREATION="true"
    fi
  fi
fi

# ============================================================================
# STEP 5: Create test realm (if not skipped)
# ============================================================================
# Creates a new realm with settings appropriate for integration testing
if [ "$SKIP_REALM_CREATION" != "true" ]; then
  echo ""
  gum log --level info "Creating test realm" realm="$TEST_REALM"
  
  # Build realm configuration JSON
  REALM_JSON=$(jq -n \
    --arg realm "$TEST_REALM" \
    '{
      realm: $realm,
      enabled: true,
      displayName: "Test Realm for Integration Testing",
      loginWithEmailAllowed: true,
      duplicateEmailsAllowed: false,
      resetPasswordAllowed: true,
      editUsernameAllowed: false,
      bruteForceProtected: false
    }')
  
  if gum spin --spinner dot --title "Creating realm..." -- \
    curl -sf -X POST "$KEYCLOAK_URL/admin/realms" \
    -H "Authorization: Bearer $ADMIN_TOKEN" \
    -H "Content-Type: application/json" \
    -d "$REALM_JSON"; then
    gum log --level info "Test realm created successfully"
  else
    gum log --level error "Failed to create realm"
    exit 1
  fi
fi

# ============================================================================
# STEP 6: Create service account client
# ============================================================================
# Sets up a confidential client with service account enabled for API access
echo ""
gum log --level info "Setting up test client" client="$TEST_CLIENT_ID"

# Check if client already exists (empty result is normal if it doesn't)
# Note: Don't use -f flag here - empty result is expected when client doesn't exist
EXISTING_CLIENT=$(curl -s -X GET "$KEYCLOAK_URL/admin/realms/$TEST_REALM/clients?clientId=$TEST_CLIENT_ID" \
  -H "Authorization: Bearer $ADMIN_TOKEN" 2>/dev/null | jq -r '.[0].id // empty')

if [ -n "$EXISTING_CLIENT" ]; then
  gum log --level warn "Client already exists, deleting..."
  if gum spin --spinner dot --title "Removing old client..." -- \
    curl -sf -X DELETE "$KEYCLOAK_URL/admin/realms/$TEST_REALM/clients/$EXISTING_CLIENT" \
    -H "Authorization: Bearer $ADMIN_TOKEN"; then
    gum log --level debug "Old client deleted"
  else
    gum log --level error "Failed to delete old client"
    exit 1
  fi
fi

CLIENT_JSON=$(jq -n \
  --arg clientId "$TEST_CLIENT_ID" \
  --arg secret "$TEST_CLIENT_SECRET" \
  '{
    clientId: $clientId,
    enabled: true,
    serviceAccountsEnabled: true,
    publicClient: false,
    protocol: "openid-connect",
    secret: $secret,
    standardFlowEnabled: false,
    directAccessGrantsEnabled: false,
    implicitFlowEnabled: false
  }')

if gum spin --spinner dot --title "Creating client..." -- \
  curl -sf -X POST "$KEYCLOAK_URL/admin/realms/$TEST_REALM/clients" \
  -H "Authorization: Bearer $ADMIN_TOKEN" \
  -H "Content-Type: application/json" \
  -d "$CLIENT_JSON"; then
  gum log --level info "Test client created successfully"
else
  gum log --level error "Failed to create client"
  exit 1
fi

# ============================================================================
# STEP 7: Retrieve client UUID
# ============================================================================
# Keycloak uses internal UUIDs - we need this for subsequent operations
echo ""
gum log --level info "Retrieving client UUID..."
CLIENT_UUID=$(curl -sf -X GET "$KEYCLOAK_URL/admin/realms/$TEST_REALM/clients?clientId=$TEST_CLIENT_ID" \
  -H "Authorization: Bearer $ADMIN_TOKEN" | jq -r '.[0].id') || {
  gum log --level error "Failed to retrieve client UUID"
  exit 1
}

if [ -z "$CLIENT_UUID" ] || [ "$CLIENT_UUID" == "null" ]; then
  gum log --level error "Failed to get client UUID"
  exit 1
fi
gum log --level debug "Client UUID retrieved" uuid="$CLIENT_UUID"

# ============================================================================
# STEP 8: Get service account user ID
# ============================================================================
# Service account clients have an associated user for role assignments
echo ""
gum log --level info "Retrieving service account user..."
SERVICE_ACCOUNT_USER=$(curl -sf -X GET "$KEYCLOAK_URL/admin/realms/$TEST_REALM/clients/$CLIENT_UUID/service-account-user" \
  -H "Authorization: Bearer $ADMIN_TOKEN" | jq -r '.id') || {
  gum log --level error "Failed to retrieve service account user"
  exit 1
}

if [ -z "$SERVICE_ACCOUNT_USER" ] || [ "$SERVICE_ACCOUNT_USER" == "null" ]; then
  gum log --level error "Failed to get service account user"
  exit 1
fi
gum log --level debug "Service account user retrieved" user_id="$SERVICE_ACCOUNT_USER"

# ============================================================================
# STEP 9: Assign admin roles to service account
# ============================================================================
# Grants necessary permissions: manage-groups, view-users, manage-users, query-groups
echo ""
gum log --level info "Assigning admin roles..."
REALM_MGMT_CLIENT=$(curl -sf -X GET "$KEYCLOAK_URL/admin/realms/$TEST_REALM/clients?clientId=realm-management" \
  -H "Authorization: Bearer $ADMIN_TOKEN" | jq -r '.[0].id') || {
  gum log --level error "Failed to retrieve realm-management client"
  exit 1
}

if [ -z "$REALM_MGMT_CLIENT" ] || [ "$REALM_MGMT_CLIENT" == "null" ]; then
  gum log --level error "Invalid realm-management client ID"
  exit 1
fi

# Get all available roles
ALL_ROLES=$(curl -sf -X GET "$KEYCLOAK_URL/admin/realms/$TEST_REALM/clients/$REALM_MGMT_CLIENT/roles" \
  -H "Authorization: Bearer $ADMIN_TOKEN" | jq -c 'map(select(.name | IN("manage-groups", "view-users", "manage-users", "query-groups")))') || {
  gum log --level error "Failed to retrieve roles"
  exit 1
}

# Assign all available roles at once
if gum spin --spinner dot --title "Assigning roles..." -- \
  curl -sf -X POST "$KEYCLOAK_URL/admin/realms/$TEST_REALM/users/$SERVICE_ACCOUNT_USER/role-mappings/clients/$REALM_MGMT_CLIENT" \
  -H "Authorization: Bearer $ADMIN_TOKEN" \
  -H "Content-Type: application/json" \
  -d "$ALL_ROLES" > /dev/null; then
  gum log --level info "Roles assigned successfully"
else
  gum log --level error "Failed to assign roles"
  exit 1
fi

# ============================================================================
# STEP 10: Test the configuration
# ============================================================================
# Validates setup by obtaining a token using client credentials flow
echo ""
gum log --level info "Testing client credentials..."
TEST_TOKEN=$(gum spin --spinner dot --title "Validating configuration..." -- \
  curl -s -X POST "$KEYCLOAK_URL/realms/$TEST_REALM/protocol/openid-connect/token" \
  -H "Content-Type: application/x-www-form-urlencoded" \
  -d "client_id=$TEST_CLIENT_ID" \
  -d "client_secret=$TEST_CLIENT_SECRET" \
  -d "grant_type=client_credentials" | jq -r '.access_token')

if [ "$TEST_TOKEN" == "null" ] || [ -z "$TEST_TOKEN" ]; then
  gum log --level error "Failed to obtain access token"
  exit 1
fi
gum log --level info "Client credentials validated successfully"

# ============================================================================
# STEP 11: Create .env file with configuration
# ============================================================================
# Saves connection details for use in integration tests
echo ""
gum log --level info "Creating .env file for local testing..."
if cat > .env << EOF
# Keycloak Integration Test Configuration
# Generated by setup-keycloak.sh on $(date)

KEYCLOAK_URL=$KEYCLOAK_URL
KEYCLOAK_REALM=$TEST_REALM
KEYCLOAK_CLIENT_ID=$TEST_CLIENT_ID
KEYCLOAK_CLIENT_SECRET=$TEST_CLIENT_SECRET
EOF
then
  gum log --level info ".env file created successfully"
else
  gum log --level error "Failed to create .env file"
  exit 1
fi

# ============================================================================
# STEP 12: Display success summary with next steps
# ============================================================================
# Shows styled summary with configuration details and usage instructions
echo ""
gum style \
  --foreground 212 \
  --border-foreground 212 \
  --border double \
  --align center \
  --width 60 \
  --padding "1 2" \
  "Setup Completed Successfully!"

echo ""
gum format -- "## Configuration Saved to .env

\`\`\`bash
KEYCLOAK_URL=$KEYCLOAK_URL
KEYCLOAK_REALM=$TEST_REALM
KEYCLOAK_CLIENT_ID=$TEST_CLIENT_ID
KEYCLOAK_CLIENT_SECRET=$TEST_CLIENT_SECRET
\`\`\`

## Next Steps

1. **Load environment variables:**
   \`\`\`bash
   source .env
   \`\`\`

2. **Run integration tests:**
   \`\`\`bash
   go test -v -tags=integration ./...
   \`\`\`

3. **Or run with coverage:**
   \`\`\`bash
   go test -v -tags=integration -coverprofile=coverage.txt ./...
   \`\`\`

## Keycloak Admin Console

- **URL:** $KEYCLOAK_URL
- **Username:** $KEYCLOAK_ADMIN_USER
- **Password:** $KEYCLOAK_ADMIN_PASSWORD
"

echo ""
