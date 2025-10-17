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

package keycloak

// User represents a Keycloak user with all their properties.
// Returned by the group members endpoint and other user-related endpoints.
// This struct maps to Keycloak's UserRepresentation.
type User struct {
	ID                         *string              `json:"id,omitempty"`                         // Unique identifier for the user
	Username                   *string              `json:"username,omitempty"`                   // Username (login name)
	FirstName                  *string              `json:"firstName,omitempty"`                  // First name
	LastName                   *string              `json:"lastName,omitempty"`                   // Last name
	Email                      *string              `json:"email,omitempty"`                      // Email address
	EmailVerified              *bool                `json:"emailVerified,omitempty"`              // Whether email is verified
	Attributes                 *map[string][]string `json:"attributes,omitempty"`                 // Custom user attributes
	UserProfileMetadata        *UserProfileMetadata `json:"userProfileMetadata,omitempty"`        // User profile metadata
	Enabled                    *bool                `json:"enabled,omitempty"`                    // Whether user account is enabled
	Self                       *string              `json:"self,omitempty"`                       // Self reference URL
	Origin                     *string              `json:"origin,omitempty"`                     // Origin of the user
	CreatedTimestamp           *int64               `json:"createdTimestamp,omitempty"`           // Unix timestamp of creation (milliseconds)
	Totp                       *bool                `json:"totp,omitempty"`                       // Whether TOTP is configured
	FederationLink             *string              `json:"federationLink,omitempty"`             // Link to federated identity provider
	ServiceAccountClientID     *string              `json:"serviceAccountClientId,omitempty"`     // Client ID if this is a service account
	Credentials                *[]Credential        `json:"credentials,omitempty"`                // User credentials
	DisableableCredentialTypes *[]string            `json:"disableableCredentialTypes,omitempty"` // Credential types that can be disabled (Set)
	RequiredActions            *[]string            `json:"requiredActions,omitempty"`            // Required actions for the user
	FederatedIdentities        *[]FederatedIdentity `json:"federatedIdentities,omitempty"`        // Federated identity providers
	RealmRoles                 *[]string            `json:"realmRoles,omitempty"`                 // Realm-level roles
	ClientRoles                *map[string][]string `json:"clientRoles,omitempty"`                // Client-specific roles
	ClientConsents             *[]UserConsent       `json:"clientConsents,omitempty"`             // Client consents
	NotBefore                  *int32               `json:"notBefore,omitempty"`                  // Not valid before timestamp (seconds)
	ApplicationRoles           *map[string][]string `json:"applicationRoles,omitempty"`           // Application-specific roles
	SocialLinks                *[]SocialLink        `json:"socialLinks,omitempty"`                // Social links (deprecated, use federatedIdentities)
	Groups                     *[]string            `json:"groups,omitempty"`                     // Groups the user belongs to
	Access                     *map[string]bool     `json:"access,omitempty"`                     // Access permissions
}

// UserProfileMetadata represents metadata about a user's profile.
type UserProfileMetadata struct {
	Attributes *[]UserProfileAttributeMetadata      `json:"attributes,omitempty"` // Attribute metadata
	Groups     *[]UserProfileAttributeGroupMetadata `json:"groups,omitempty"`     // Group metadata
}

// UserProfileAttributeMetadata represents metadata for a user profile attribute.
type UserProfileAttributeMetadata struct {
	Name         *string                 `json:"name,omitempty"`         // Attribute name
	DisplayName  *string                 `json:"displayName,omitempty"`  // Display name
	Required     *bool                   `json:"required,omitempty"`     // Whether attribute is required
	ReadOnly     *bool                   `json:"readOnly,omitempty"`     // Whether attribute is read-only
	Annotations  *map[string]interface{} `json:"annotations,omitempty"`  // Annotations
	Validators   *map[string]interface{} `json:"validators,omitempty"`   // Validators configuration (Map of map)
	Group        *string                 `json:"group,omitempty"`        // Group this attribute belongs to
	Multivalued  *bool                   `json:"multivalued,omitempty"`  // Whether attribute can have multiple values
	DefaultValue *string                 `json:"defaultValue,omitempty"` // Default value for the attribute
}

// UserProfileAttributeGroupMetadata represents metadata for a user profile attribute group.
type UserProfileAttributeGroupMetadata struct {
	Name               *string                 `json:"name,omitempty"`               // Group name
	DisplayHeader      *string                 `json:"displayHeader,omitempty"`      // Display header
	DisplayDescription *string                 `json:"displayDescription,omitempty"` // Display description
	Annotations        *map[string]interface{} `json:"annotations,omitempty"`        // Annotations
}

// Credential represents a user credential in Keycloak.
type Credential struct {
	ID                *string                 `json:"id,omitempty"`                // Credential ID
	Type              *string                 `json:"type,omitempty"`              // Credential type (e.g., "password", "otp")
	UserLabel         *string                 `json:"userLabel,omitempty"`         // User-defined label
	CreatedDate       *int64                  `json:"createdDate,omitempty"`       // Creation timestamp (milliseconds)
	SecretData        *string                 `json:"secretData,omitempty"`        // Secret data (encrypted)
	CredentialData    *string                 `json:"credentialData,omitempty"`    // Credential data
	Priority          *int32                  `json:"priority,omitempty"`          // Priority
	Value             *string                 `json:"value,omitempty"`             // Credential value (for setting)
	Temporary         *bool                   `json:"temporary,omitempty"`         // Whether credential is temporary
	Device            *string                 `json:"device,omitempty"`            // Device identifier
	HashedSaltedValue *string                 `json:"hashedSaltedValue,omitempty"` // Hashed and salted credential value
	Salt              *string                 `json:"salt,omitempty"`              // Salt used for hashing
	HashIterations    *int32                  `json:"hashIterations,omitempty"`    // Number of hash iterations
	Counter           *int32                  `json:"counter,omitempty"`           // Counter (for HOTP)
	Algorithm         *string                 `json:"algorithm,omitempty"`         // Hash algorithm
	Digits            *int32                  `json:"digits,omitempty"`            // Number of digits (for OTP)
	Period            *int32                  `json:"period,omitempty"`            // Period in seconds (for TOTP)
	Config            *map[string]interface{} `json:"config,omitempty"`            // Configuration map
	FederationLink    *string                 `json:"federationLink,omitempty"`    // Federation link
}

// FederatedIdentity represents a federated identity link for a user.
type FederatedIdentity struct {
	IdentityProvider *string `json:"identityProvider,omitempty"` // Identity provider ID
	UserID           *string `json:"userId,omitempty"`           // User ID in the external provider
	UserName         *string `json:"userName,omitempty"`         // Username in the external provider
}

// UserConsent represents a user's consent to a client.
type UserConsent struct {
	ClientID            *string   `json:"clientId,omitempty"`            // Client ID
	GrantedClientScopes *[]string `json:"grantedClientScopes,omitempty"` // Granted client scopes
	CreatedDate         *int64    `json:"createdDate,omitempty"`         // Creation timestamp (milliseconds)
	LastUpdatedDate     *int64    `json:"lastUpdatedDate,omitempty"`     // Last update timestamp (milliseconds)
	GrantedRealmRoles   *[]string `json:"grantedRealmRoles,omitempty"`   // Granted realm roles
}

// SocialLink represents a social link (deprecated).
// Use FederatedIdentity instead.
type SocialLink struct {
	SocialProvider *string `json:"socialProvider,omitempty"` // Social provider name
	SocialUserID   *string `json:"socialUserId,omitempty"`   // User ID in the social provider
	SocialUsername *string `json:"socialUsername,omitempty"` // Username in the social provider
}
