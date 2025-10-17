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

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/go-resty/resty/v2"
)

var (
	// ErrGroupNotFound is returned when a requested group cannot be found.
	ErrGroupNotFound = errors.New("group not found")
)

// GroupsClient provides methods for managing Keycloak groups.
// It handles group CRUD operations, subgroup management, and group searches.
type GroupsClient interface {
	// Create creates a new group in Keycloak with the specified name and attributes.
	// Returns the newly created group's ID.
	Create(ctx context.Context, name string, attributes map[string][]string) (string, error)

	// Update updates an existing group with the provided group data.
	Update(ctx context.Context, updatedGroup Group) error

	// Delete deletes a group by its ID.
	Delete(ctx context.Context, groupID string) error

	// List retrieves all groups matching the optional search criteria.
	// If briefRepresentation is true, returns groups without detailed attributes.
	List(ctx context.Context, search *string, briefRepresentation bool) ([]*Group, error)

	// ListWithParams retrieves groups with full control over all query parameters.
	// This provides access to all Keycloak API parameters including exact matching,
	// hierarchy population, and subgroup counts.
	ListWithParams(ctx context.Context, params SearchGroupParams) ([]*Group, error)

	// ListWithSubGroups retrieves groups including their subgroup hierarchies.
	// This is a convenience method that automatically sets the Search parameter,
	// which is required by Keycloak's API to include subgroups in the response.
	// Use searchQuery to filter groups (use empty string "" or a broad term to match all groups).
	ListWithSubGroups(ctx context.Context, searchQuery string, briefRepresentation bool, first, max int) ([]*Group, error)

	// Count returns the total count of groups matching the search criteria.
	Count(ctx context.Context, search *string, top *bool) (int, error)

	// ListPaginated retrieves a paginated list of groups.
	// Parameters first and max control pagination (offset and limit).
	ListPaginated(ctx context.Context, search *string, briefRepresentation bool, first, max int) ([]*Group, error)

	// Get retrieves a single group by its ID.
	Get(ctx context.Context, groupID string) (*Group, error)

	// GetByAttribute searches for a group with the specified attribute key-value pair.
	// Returns ErrGroupNotFound if no matching group is found.
	GetByAttribute(ctx context.Context, attribute *GroupAttribute) (*Group, error)

	// ListSubGroups retrieves all direct child groups of the specified parent group.
	ListSubGroups(ctx context.Context, groupID string) ([]*Group, error)

	// CountSubGroups returns the count of direct child groups.
	CountSubGroups(ctx context.Context, groupID string) (int, error)

	// ListSubGroupsPaginated retrieves a paginated list of subgroups with optional search filtering.
	// Uses the /groups/{group-id}/children endpoint for server-side pagination and filtering.
	ListSubGroupsPaginated(ctx context.Context, groupID string, params SubGroupSearchParams) ([]*Group, error)

	// CreateSubGroup creates a new subgroup under the specified parent group.
	// Returns the newly created subgroup's ID.
	CreateSubGroup(ctx context.Context, groupID, name string, attributes map[string][]string) (string, error)

	// GetSubGroupByAttribute searches for a subgroup with the specified attribute within a parent group.
	GetSubGroupByAttribute(group Group, attribute GroupAttribute) (*Group, error)

	// GetSubGroupByID finds a subgroup by its ID within a parent group's children.
	GetSubGroupByID(group Group, subGroupID string) (*Group, error)

	// ListMembers retrieves the users that are members of the specified group.
	// Returns a filtered stream of users according to the query parameters.
	ListMembers(ctx context.Context, groupID string, params GroupMembersParams) ([]*User, error)

	// GetManagementPermissions returns whether client Authorization permissions have been initialized
	// for this group and provides a reference.
	GetManagementPermissions(ctx context.Context, groupID string) (*ManagementPermissionReference, error)

	// UpdateManagementPermissions enables or disables client Authorization permissions for this group
	// and returns the updated permission reference.
	UpdateManagementPermissions(ctx context.Context, groupID string, ref ManagementPermissionReference) (*ManagementPermissionReference, error)
}

// groupsClient implements the GroupsClient interface.
type groupsClient struct {
	client          *Client
	authAdminRealms string
}

// newGroupsClient creates a new GroupsClient implementation.
func newGroupsClient(client *Client, authAdminRealms string) GroupsClient {
	return &groupsClient{
		client:          client,
		authAdminRealms: authAdminRealms,
	}
}

// Create creates a new group in Keycloak with the specified name and attributes.
func (g *groupsClient) Create(ctx context.Context, name string, attributes map[string][]string) (string, error) {
	group := Group{
		Name:       &name,
		Attributes: &attributes,
	}

	resp, err := g.getRequest(ctx).SetBody(group).Post(g.getAdminRealmURL("groups"))
	if err != nil {
		return "", fmt.Errorf("unable to create group: %w", err)
	}
	if !resp.IsSuccess() {
		return "", fmt.Errorf("unable to create group: %v", resp.Error())
	}

	return getID(resp), nil
}

// Update updates an existing group with the provided group data.
func (g *groupsClient) Update(ctx context.Context, group Group) error {
	if NilOrEmpty(group.ID) {
		return fmt.Errorf("the ID of the group is required")
	}

	resp, err := g.getRequest(ctx).SetBody(group).Put(g.getAdminRealmURL("groups", *group.ID))
	if err != nil {
		return fmt.Errorf("unable to update group: %w", err)
	}
	if !resp.IsSuccess() {
		return fmt.Errorf("unable to update group: %v", resp.Error())
	}

	return nil
}

// List retrieves all groups matching the optional search criteria.
func (g *groupsClient) List(ctx context.Context, search *string, briefRepresentation bool) ([]*Group, error) {
	return g.list(ctx, SearchGroupParams{
		Search:              search,
		BriefRepresentation: &briefRepresentation,
	})
}

// ListPaginated retrieves a paginated list of groups.
func (g *groupsClient) ListPaginated(ctx context.Context, search *string, briefRepresentation bool, first, max int) ([]*Group, error) {
	return g.list(ctx, SearchGroupParams{
		Search:              search,
		BriefRepresentation: &briefRepresentation,
		First:               &first,
		Max:                 &max,
	})
}

// ListWithParams retrieves groups with full control over all query parameters.
func (g *groupsClient) ListWithParams(ctx context.Context, params SearchGroupParams) ([]*Group, error) {
	return g.list(ctx, params)
}

// ListWithSubGroups retrieves groups including their subgroup hierarchies.
// This is a convenience method that automatically sets the Search parameter,
// which is required by Keycloak's API to include subgroups in the response.
//
// Note: Due to Keycloak API behavior, subgroups are only returned when a search
// parameter is provided. This method uses the provided searchQuery to enable
// subgroup population.
//
// Parameters:
//   - searchQuery: Search term to filter groups (use empty string "" or a broad term to match all groups)
//   - briefRepresentation: If true, return groups without detailed attributes
//   - first: Pagination offset
//   - max: Maximum number of results
//
// Returns groups matching the search with their SubGroups field populated.
func (g *groupsClient) ListWithSubGroups(ctx context.Context, searchQuery string, briefRepresentation bool, first, max int) ([]*Group, error) {
	populateHierarchy := true
	return g.list(ctx, SearchGroupParams{
		Search:              &searchQuery,
		BriefRepresentation: &briefRepresentation,
		PopulateHierarchy:   &populateHierarchy,
		First:               &first,
		Max:                 &max,
	})
}

// list is an internal method that handles group listing with all optional parameters.
func (g *groupsClient) list(ctx context.Context, params SearchGroupParams) ([]*Group, error) {
	var result []*Group

	queryParams, err := mapper(params)
	if err != nil {
		return nil, fmt.Errorf("failed to initiate search parameters of groups: %w", err)
	}

	resp, err := g.getRequest(ctx).SetResult(&result).SetQueryParams(queryParams).Get(g.getAdminRealmURL("groups"))
	if err != nil {
		return nil, fmt.Errorf("unable to list groups: %w", err)
	}

	if !resp.IsSuccess() {
		return nil, fmt.Errorf("unable to list groups: %v", resp.Error())
	}

	return result, nil
}

// Count returns the total count of groups matching the search criteria.
func (g *groupsClient) Count(ctx context.Context, search *string, top *bool) (int, error) {
	var result CountGroupResponse

	queryParams, err := mapper(CountGroupParams{
		Search: search, // name search
		Top:    top,
	})
	if err != nil {
		return 0, fmt.Errorf("failed to initiate search parameters of groups: %w", err)
	}

	resp, err := g.getRequest(ctx).SetResult(&result).SetQueryParams(queryParams).Get(g.getAdminRealmURL("groups", "count"))
	if err != nil {
		return 0, fmt.Errorf("unable to count groups: %w", err)
	}

	if !resp.IsSuccess() {
		return 0, fmt.Errorf("unable to count groups: %v", resp.Error())
	}

	return result.Count, nil
}

// Get retrieves a single group by its ID.
func (g *groupsClient) Get(ctx context.Context, groupID string) (*Group, error) {
	var result Group

	resp, err := g.getRequest(ctx).SetResult(&result).Get(g.getAdminRealmURL("groups", groupID))
	if err != nil {
		return nil, fmt.Errorf("unable to get group: %w", err)
	}

	if !resp.IsSuccess() {
		// Return sentinel error for 404 Not Found
		if resp.StatusCode() == 404 {
			return nil, ErrGroupNotFound
		}
		return nil, fmt.Errorf("unable to get group: %v", resp.Error())
	}

	return &result, nil
}

// GetByAttribute searches for a group with the specified attribute key-value pair.
func (g *groupsClient) GetByAttribute(ctx context.Context, attribute *GroupAttribute) (*Group, error) {
	if attribute == nil {
		return nil, errors.New("attributes map is empty")
	}

	currentPage := 0

	var (
		groups []*Group
		err    error
	)

	for {
		groups, err = g.ListPaginated(ctx, nil, false, currentPage*g.client.pageSize, g.client.pageSize)
		if err != nil {
			return nil, err
		}

		// iterate result and look for the Reference
		if group, ok := searchInGroupsAttributes(groups, *attribute); ok {
			return group, nil
		}

		if len(groups) < g.client.pageSize {
			// last page, finish search
			return nil, ErrGroupNotFound
		}

		currentPage++
	}
}

// GetSubGroupByID finds a subgroup by its ID within a parent group's children.
func (g *groupsClient) GetSubGroupByID(group Group, subGroupID string) (*Group, error) {
	for _, subGroup := range *group.SubGroups {
		if subGroup != nil && subGroup.ID != nil && *subGroup.ID == subGroupID {
			return subGroup, nil
		}
	}

	return nil, ErrGroupNotFound
}

// CreateSubGroup creates a new subgroup under the specified parent group.
func (g *groupsClient) CreateSubGroup(ctx context.Context, groupID, name string, attributes map[string][]string) (string, error) {
	if groupID == "" {
		return "", errors.New("groupID parameter cannot be empty")
	}

	group := Group{
		Name:       &name,
		Attributes: &attributes,
	}

	resp, err := g.getRequest(ctx).SetBody(group).Post(g.getAdminRealmURL("groups", groupID, "children"))
	if err != nil {
		return "", fmt.Errorf("unable to create sub-group: %w", err)
	}
	if !resp.IsSuccess() {
		return "", fmt.Errorf("unable to create sub-group: %v", resp.Error())
	}

	return getID(resp), nil
}

// ListSubGroups retrieves all direct child groups of the specified parent group.
func (g *groupsClient) ListSubGroups(ctx context.Context, groupID string) ([]*Group, error) {
	var result []*Group

	resp, err := g.getRequest(ctx).SetResult(&result).Get(g.getAdminRealmURL("groups", groupID, "children"))
	if err != nil {
		return nil, fmt.Errorf("unable to list groups: %w", err)
	}
	if !resp.IsSuccess() {
		return nil, fmt.Errorf("unable to list groups: %v", resp.Error())
	}

	return result, nil
}

// ListSubGroupsPaginated retrieves a paginated list of subgroups.
// Uses the /groups/{group-id}/children endpoint which supports server-side pagination,
// search filtering, and other query parameters.
func (g *groupsClient) ListSubGroupsPaginated(ctx context.Context, groupID string, params SubGroupSearchParams) ([]*Group, error) {
	if groupID == "" {
		return nil, fmt.Errorf("groupID parameter cannot be empty")
	}

	var result []*Group

	queryParams, err := mapper(params)
	if err != nil {
		return nil, fmt.Errorf("failed to initiate search parameters for sub-groups: %w", err)
	}

	resp, err := g.getRequest(ctx).SetResult(&result).SetQueryParams(queryParams).Get(g.getAdminRealmURL("groups", groupID, "children"))
	if err != nil {
		return nil, fmt.Errorf("unable to list sub-groups: %w", err)
	}

	if !resp.IsSuccess() {
		return nil, fmt.Errorf("unable to list sub-groups: %v", resp.Error())
	}

	return result, nil
}

// CountSubGroups returns the count of direct child groups.
func (g *groupsClient) CountSubGroups(ctx context.Context, groupID string) (int, error) {
	var result Group
	resp, err := g.getRequest(ctx).SetResult(&result).Get(g.getAdminRealmURL("groups", groupID))
	if err != nil {
		return 0, fmt.Errorf("unable to list sub-groups: %w", err)
	}

	if !resp.IsSuccess() {
		return 0, fmt.Errorf("unable to list sub-groups: %v", resp.Error())
	}

	if result.SubGroups == nil {
		return 0, nil
	}

	return len(*result.SubGroups), nil
}

// GetSubGroupByAttribute searches for a subgroup with the specified attribute within a parent group.
func (g *groupsClient) GetSubGroupByAttribute(group Group, attribute GroupAttribute) (*Group, error) {
	if group.SubGroups == nil {
		return nil, fmt.Errorf("unable to find subgroup")
	}

	subGroup, found := findGroupByAttribute(*group.SubGroups, attribute)
	if !found {
		return nil, fmt.Errorf("unable to find subgroup")
	}

	return subGroup, nil
}

// Delete deletes a group by its ID.
func (g *groupsClient) Delete(ctx context.Context, groupID string) error {
	if groupID == "" {
		return fmt.Errorf("groupID parameter cannot be empty")
	}

	resp, err := g.getRequest(ctx).Delete(g.getAdminRealmURL("groups", groupID))
	if err != nil {
		return fmt.Errorf("unable to delete group: %w", err)
	}

	if !resp.IsSuccess() {
		return fmt.Errorf("unable to delete group: %v", resp.Error())
	}

	return nil
}

// getRequest creates an HTTP request with error handling configured.
func (g *groupsClient) getRequest(ctx context.Context) *resty.Request {
	var err HTTPErrorResponse
	return g.client.resty.R().SetContext(ctx).SetError(&err)
}

// getAdminRealmURL constructs the full URL for Keycloak Admin API group endpoints.
func (g *groupsClient) getAdminRealmURL(path ...string) string {
	allPaths := append([]string{g.client.baseURL, g.authAdminRealms, g.client.realm}, path...)
	return makeURL(allPaths...)
}

// searchInGroupsAttributes searches for a group with a specific attribute in a list of groups.
func searchInGroupsAttributes(groups []*Group, attribute GroupAttribute) (*Group, bool) {
	return findGroupByAttribute(groups, attribute)
}

// findGroupByAttribute is a helper function that searches for a group with a specific attribute
// in a slice of groups. It returns the matching group and a boolean indicating if found.
func findGroupByAttribute(groups []*Group, attribute GroupAttribute) (*Group, bool) {
	for _, group := range groups {
		if group == nil || group.Attributes == nil {
			continue
		}

		groupAttributes := *group.Attributes

		if value, ok := groupAttributes[attribute.Key]; ok {
			if len(value) != 1 {
				return nil, false
			}
			if value[0] == attribute.Value {
				return group, true
			}
		}
	}

	return nil, false
}

// getID extracts the resource ID from the Location header in the HTTP response.
func getID(resp *resty.Response) string {
	header := resp.Header().Get("Location")
	splitPath := strings.Split(header, urlSeparator)
	return splitPath[len(splitPath)-1]
}

// ListMembers retrieves the users that are members of the specified group.
func (g *groupsClient) ListMembers(ctx context.Context, groupID string, params GroupMembersParams) ([]*User, error) {
	if groupID == "" {
		return nil, fmt.Errorf("groupID parameter cannot be empty")
	}

	var result []*User

	queryParams, err := mapper(params)
	if err != nil {
		return nil, fmt.Errorf("failed to initiate search parameters for group members: %w", err)
	}

	resp, err := g.getRequest(ctx).SetResult(&result).SetQueryParams(queryParams).Get(g.getAdminRealmURL("groups", groupID, "members"))
	if err != nil {
		return nil, fmt.Errorf("unable to list group members: %w", err)
	}

	if !resp.IsSuccess() {
		return nil, fmt.Errorf("unable to list group members: %v", resp.Error())
	}

	return result, nil
}

// GetManagementPermissions returns whether client Authorization permissions have been initialized.
func (g *groupsClient) GetManagementPermissions(ctx context.Context, groupID string) (*ManagementPermissionReference, error) {
	if groupID == "" {
		return nil, fmt.Errorf("groupID parameter cannot be empty")
	}

	var result ManagementPermissionReference

	resp, err := g.getRequest(ctx).SetResult(&result).Get(g.getAdminRealmURL("groups", groupID, "management", "permissions"))
	if err != nil {
		return nil, fmt.Errorf("unable to get management permissions: %w", err)
	}

	if !resp.IsSuccess() {
		return nil, fmt.Errorf("unable to get management permissions: %v", resp.Error())
	}

	return &result, nil
}

// UpdateManagementPermissions enables or disables client Authorization permissions for the group.
func (g *groupsClient) UpdateManagementPermissions(ctx context.Context, groupID string, ref ManagementPermissionReference) (*ManagementPermissionReference, error) {
	if groupID == "" {
		return nil, fmt.Errorf("groupID parameter cannot be empty")
	}

	var result ManagementPermissionReference

	resp, err := g.getRequest(ctx).SetBody(ref).SetResult(&result).Put(g.getAdminRealmURL("groups", groupID, "management", "permissions"))
	if err != nil {
		return nil, fmt.Errorf("unable to update management permissions: %w", err)
	}

	if !resp.IsSuccess() {
		return nil, fmt.Errorf("unable to update management permissions: %v", resp.Error())
	}

	return &result, nil
}
