#!/usr/bin/env bash
# Setup Keycloak for Integration Testing
# 
# This script configures a Keycloak instance with a test realm and client
# for running integration tests.
#
# Prerequisites:
#   - Keycloak running on http://localhost:8080
#   - curl and jq installed
#
# Usage:
#   ./scripts/setup-keycloak.sh

set -euo pipefail

# Configuration
KEYCLOAK_URL="${KEYCLOAK_URL:-http://localhost:8080}"
KEYCLOAK_ADMIN_USER="${KEYCLOAK_ADMIN_USER:-admin}"
KEYCLOAK_ADMIN_PASSWORD="${KEYCLOAK_ADMIN_PASSWORD:-admin}"
TEST_REALM="${TEST_REALM:-test-realm}"
TEST_CLIENT_ID="${TEST_CLIENT_ID:-test-client}"
TEST_CLIENT_SECRET="${TEST_CLIENT_SECRET:-test-secret-12345}"

echo "🔧 Setting up Keycloak for integration testing..."
echo "   URL: $KEYCLOAK_URL"
echo "   Realm: $TEST_REALM"
echo "   Client: $TEST_CLIENT_ID"
echo ""

# Wait for Keycloak to be ready
echo "⏳ Waiting for Keycloak to be ready..."
max_attempts=60
attempt=0
while [ $attempt -lt $max_attempts ]; do
  if curl -sf "$KEYCLOAK_URL/realms/master" > /dev/null 2>&1; then
    echo "✓ Keycloak is ready!"
    break
  fi
  echo "   Waiting... (attempt $((attempt + 1))/$max_attempts)"
  sleep 2
  attempt=$((attempt + 1))
done

if [ $attempt -eq $max_attempts ]; then
  echo "❌ Keycloak failed to become ready in time"
  exit 1
fi

# Get admin access token
echo ""
echo "🔑 Obtaining admin access token..."
ADMIN_TOKEN=$(curl -s -X POST "$KEYCLOAK_URL/realms/master/protocol/openid-connect/token" \
  -H "Content-Type: application/x-www-form-urlencoded" \
  -d "username=$KEYCLOAK_ADMIN_USER" \
  -d "password=$KEYCLOAK_ADMIN_PASSWORD" \
  -d "grant_type=password" \
  -d "client_id=admin-cli" | jq -r '.access_token')

if [ "$ADMIN_TOKEN" == "null" ] || [ -z "$ADMIN_TOKEN" ]; then
  echo "❌ Failed to obtain admin token"
  echo "   Check admin credentials: $KEYCLOAK_ADMIN_USER / $KEYCLOAK_ADMIN_PASSWORD"
  exit 1
fi
echo "✓ Admin token obtained"

# Check if test realm already exists
echo ""
echo "🌐 Checking for existing test realm..."
EXISTING_REALM=$(curl -s -X GET "$KEYCLOAK_URL/admin/realms/$TEST_REALM" \
  -H "Authorization: Bearer $ADMIN_TOKEN" \
  -H "Accept: application/json" | jq -r '.realm // empty')

if [ "$EXISTING_REALM" == "$TEST_REALM" ]; then
  echo "⚠️  Test realm '$TEST_REALM' already exists"
  read -p "   Delete and recreate? (y/N): " -n 1 -r
  echo
  if [[ $REPLY =~ ^[Yy]$ ]]; then
    echo "   Deleting existing realm..."
    curl -s -X DELETE "$KEYCLOAK_URL/admin/realms/$TEST_REALM" \
      -H "Authorization: Bearer $ADMIN_TOKEN"
    echo "   ✓ Realm deleted"
  else
    echo "   Using existing realm"
    # Skip realm creation but continue with client setup
    SKIP_REALM_CREATION=true
  fi
fi

# Create test realm
if [ "$SKIP_REALM_CREATION" != "true" ]; then
  echo ""
  echo "🏗️  Creating test realm '$TEST_REALM'..."
  curl -s -X POST "$KEYCLOAK_URL/admin/realms" \
    -H "Authorization: Bearer $ADMIN_TOKEN" \
    -H "Content-Type: application/json" \
    -d "{
      \"realm\": \"$TEST_REALM\",
      \"enabled\": true,
      \"displayName\": \"Test Realm for Integration Testing\",
      \"loginWithEmailAllowed\": true,
      \"duplicateEmailsAllowed\": false,
      \"resetPasswordAllowed\": true,
      \"editUsernameAllowed\": false,
      \"bruteForceProtected\": false
    }"
  echo "✓ Test realm created"
fi

# Create or update test client
echo ""
echo "👤 Setting up test client '$TEST_CLIENT_ID'..."

# Check if client exists
EXISTING_CLIENT=$(curl -s -X GET "$KEYCLOAK_URL/admin/realms/$TEST_REALM/clients?clientId=$TEST_CLIENT_ID" \
  -H "Authorization: Bearer $ADMIN_TOKEN" | jq -r '.[0].id // empty')

if [ -n "$EXISTING_CLIENT" ]; then
  echo "⚠️  Client already exists, deleting..."
  curl -s -X DELETE "$KEYCLOAK_URL/admin/realms/$TEST_REALM/clients/$EXISTING_CLIENT" \
    -H "Authorization: Bearer $ADMIN_TOKEN"
fi

curl -s -X POST "$KEYCLOAK_URL/admin/realms/$TEST_REALM/clients" \
  -H "Authorization: Bearer $ADMIN_TOKEN" \
  -H "Content-Type: application/json" \
  -d "{
    \"clientId\": \"$TEST_CLIENT_ID\",
    \"enabled\": true,
    \"serviceAccountsEnabled\": true,
    \"publicClient\": false,
    \"protocol\": \"openid-connect\",
    \"secret\": \"$TEST_CLIENT_SECRET\",
    \"standardFlowEnabled\": false,
    \"directAccessGrantsEnabled\": false,
    \"implicitFlowEnabled\": false
  }"
echo "✓ Test client created"

# Get client UUID (needed for role assignment)
echo ""
echo "🔍 Getting client UUID..."
CLIENT_UUID=$(curl -s -X GET "$KEYCLOAK_URL/admin/realms/$TEST_REALM/clients?clientId=$TEST_CLIENT_ID" \
  -H "Authorization: Bearer $ADMIN_TOKEN" | jq -r '.[0].id')

if [ -z "$CLIENT_UUID" ] || [ "$CLIENT_UUID" == "null" ]; then
  echo "❌ Failed to get client UUID"
  exit 1
fi
echo "✓ Client UUID: $CLIENT_UUID"

# Get service account user
echo ""
echo "👥 Getting service account user..."
SERVICE_ACCOUNT_USER=$(curl -s -X GET "$KEYCLOAK_URL/admin/realms/$TEST_REALM/clients/$CLIENT_UUID/service-account-user" \
  -H "Authorization: Bearer $ADMIN_TOKEN" | jq -r '.id')

if [ -z "$SERVICE_ACCOUNT_USER" ] || [ "$SERVICE_ACCOUNT_USER" == "null" ]; then
  echo "❌ Failed to get service account user"
  exit 1
fi
echo "✓ Service account user ID: $SERVICE_ACCOUNT_USER"

# Get realm-management client
echo ""
echo "🔐 Assigning admin roles..."
REALM_MGMT_CLIENT=$(curl -s -X GET "$KEYCLOAK_URL/admin/realms/$TEST_REALM/clients?clientId=realm-management" \
  -H "Authorization: Bearer $ADMIN_TOKEN" | jq -r '.[0].id')

# Get all available roles
ALL_ROLES=$(curl -s -X GET "$KEYCLOAK_URL/admin/realms/$TEST_REALM/clients/$REALM_MGMT_CLIENT/roles" \
  -H "Authorization: Bearer $ADMIN_TOKEN" | jq -c 'map(select(.name | IN("manage-groups", "view-users", "manage-users", "query-groups")))')

# Assign all available roles at once
curl -s -X POST "$KEYCLOAK_URL/admin/realms/$TEST_REALM/users/$SERVICE_ACCOUNT_USER/role-mappings/clients/$REALM_MGMT_CLIENT" \
  -H "Authorization: Bearer $ADMIN_TOKEN" \
  -H "Content-Type: application/json" \
  -d "$ALL_ROLES" > /dev/null

echo "✓ Roles assigned"

# Test the configuration
echo ""
echo "🧪 Testing client credentials..."
TEST_TOKEN=$(curl -s -X POST "$KEYCLOAK_URL/realms/$TEST_REALM/protocol/openid-connect/token" \
  -H "Content-Type: application/x-www-form-urlencoded" \
  -d "client_id=$TEST_CLIENT_ID" \
  -d "client_secret=$TEST_CLIENT_SECRET" \
  -d "grant_type=client_credentials" | jq -r '.access_token')

if [ "$TEST_TOKEN" == "null" ] || [ -z "$TEST_TOKEN" ]; then
  echo "❌ Failed to obtain access token"
  exit 1
fi
echo "✓ Client credentials working"

# Create .env file for local testing
echo ""
echo "📝 Creating .env file for local testing..."
cat > .env << EOF
# Keycloak Integration Test Configuration
# Generated by setup-keycloak.sh on $(date)

KEYCLOAK_URL=$KEYCLOAK_URL
KEYCLOAK_REALM=$TEST_REALM
KEYCLOAK_CLIENT_ID=$TEST_CLIENT_ID
KEYCLOAK_CLIENT_SECRET=$TEST_CLIENT_SECRET
EOF
echo "✓ .env file created"

# Print summary
echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "✅ Keycloak setup completed successfully!"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""
echo "Configuration saved to .env:"
echo "  KEYCLOAK_URL=$KEYCLOAK_URL"
echo "  KEYCLOAK_REALM=$TEST_REALM"
echo "  KEYCLOAK_CLIENT_ID=$TEST_CLIENT_ID"
echo "  KEYCLOAK_CLIENT_SECRET=$TEST_CLIENT_SECRET"
echo ""
echo "Next steps:"
echo "  1. Load environment variables:"
echo "     source .env"
echo ""
echo "  2. Run integration tests:"
echo "     go test -v -tags=integration ./..."
echo ""
echo "  3. Or run with coverage:"
echo "     go test -v -tags=integration -coverprofile=coverage.txt ./..."
echo ""
echo "Keycloak Admin Console: $KEYCLOAK_URL"
echo "  Username: $KEYCLOAK_ADMIN_USER"
echo "  Password: $KEYCLOAK_ADMIN_PASSWORD"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"

