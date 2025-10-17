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

// Package main demonstrates basic usage of the Keycloak client library.
// This example shows how to create a client, list groups, and retrieve group details.
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
	keycloakURL := os.Getenv("KEYCLOAK_URL")
	if keycloakURL == "" {
		keycloakURL = "https://keycloak.example.com"
	}

	realm := os.Getenv("KEYCLOAK_REALM")
	if realm == "" {
		realm = "master"
	}

	clientID := os.Getenv("KEYCLOAK_CLIENT_ID")
	if clientID == "" {
		clientID = "admin-cli"
	}

	clientSecret := os.Getenv("KEYCLOAK_CLIENT_SECRET")
	if clientSecret == "" {
		log.Fatal("KEYCLOAK_CLIENT_SECRET environment variable is required")
	}

	// Create client configuration
	config := keycloak.Config{
		URL:          keycloakURL,
		Realm:        realm,
		ClientID:     clientID,
		ClientSecret: clientSecret,
	}

	// Create new Keycloak client
	_, err := keycloak.New(ctx, config)
	if err != nil {
		log.Fatalf("Failed to create Keycloak client: %v", err)
	}

	fmt.Println("Successfully connected to Keycloak!")
	fmt.Printf("URL: %s\n", keycloakURL)
	fmt.Printf("Realm: %s\n", realm)
	fmt.Printf("Client ID: %s\n", clientID)
}
