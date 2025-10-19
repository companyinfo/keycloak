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
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"go.companyinfo.dev/ptr"
)

// BenchmarkGroupsClient_List benchmarks listing groups without search
func BenchmarkGroupsClient_List(b *testing.B) {
	mockGroups := make([]*Group, 50)
	for i := 0; i < 50; i++ {
		mockGroups[i] = &Group{
			ID:   ptr.String(fmt.Sprintf("group-%d", i)),
			Name: ptr.String(fmt.Sprintf("Group %d", i)),
		}
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(mockGroups)
	}))
	defer server.Close()

	client := &Client{
		baseURL:  server.URL,
		realm:    "test-realm",
		pageSize: 50,
		resty:    newTestRestyClient(),
	}
	client.resty.SetBaseURL(server.URL)
	gc := &groupsClient{
		client: client,
	}

	ctx := context.Background()

	b.ResetTimer()
	for b.Loop() {
		_, err := gc.List(ctx, nil, false)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkGroupsClient_Get benchmarks getting a single group by ID
func BenchmarkGroupsClient_Get(b *testing.B) {
	mockGroup := &Group{
		ID:          ptr.String("test-group-id"),
		Name:        ptr.String("Test Group"),
		Path:        ptr.String("/Test Group"),
		Description: ptr.String("A test group"),
		Attributes: &map[string][]string{
			"type": {"test"},
		},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(mockGroup)
	}))
	defer server.Close()

	client := &Client{
		baseURL:  server.URL,
		realm:    "test-realm",
		pageSize: 50,
		resty:    newTestRestyClient(),
	}
	client.resty.SetBaseURL(server.URL)
	gc := &groupsClient{
		client: client,
	}

	ctx := context.Background()

	b.ResetTimer()
	for b.Loop() {
		_, err := gc.Get(ctx, "test-group-id")
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkGroupsClient_Create benchmarks group creation
func BenchmarkGroupsClient_Create(b *testing.B) {
	var serverURL string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Location", fmt.Sprintf("%s/admin/realms/test-realm/groups/new-id", serverURL))
		w.WriteHeader(http.StatusCreated)
	}))
	defer server.Close()
	serverURL = server.URL

	client := &Client{
		baseURL:  serverURL,
		realm:    "test-realm",
		pageSize: 50,
		resty:    newTestRestyClient(),
	}
	client.resty.SetHostURL(serverURL)
	gc := &groupsClient{
		client: client,
	}

	ctx := context.Background()
	attributes := map[string][]string{"type": {"test"}}

	b.ResetTimer()
	for b.Loop() {
		_, err := gc.Create(ctx, "Test Group", attributes)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkGroupsClient_GetByAttribute benchmarks attribute-based search
func BenchmarkGroupsClient_GetByAttribute(b *testing.B) {
	// Create groups with different attributes
	mockGroups := make([]*Group, 50)
	for i := 0; i < 50; i++ {
		mockGroups[i] = &Group{
			ID:   ptr.String(fmt.Sprintf("group-%d", i)),
			Name: ptr.String(fmt.Sprintf("Group %d", i)),
			Attributes: &map[string][]string{
				"customID": {fmt.Sprintf("id-%d", i)},
			},
		}
	}

	// Add target group at position 25 (middle of page)
	mockGroups[25] = &Group{
		ID:   ptr.String("target-group"),
		Name: ptr.String("Target Group"),
		Attributes: &map[string][]string{
			"customID": {"target-12345"},
		},
	}

	var serverURL string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(mockGroups)
	}))
	defer server.Close()
	serverURL = server.URL

	client := &Client{
		baseURL:  serverURL,
		realm:    "test-realm",
		pageSize: 50,
		resty:    newTestRestyClient(),
	}
	client.resty.SetHostURL(serverURL)
	gc := &groupsClient{
		client: client,
	}

	ctx := context.Background()
	attribute := &GroupAttribute{
		Key:   "customID",
		Value: "target-12345",
	}

	b.ResetTimer()
	for b.Loop() {
		_, err := gc.GetByAttribute(ctx, attribute)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkFindGroupByAttribute benchmarks the helper function for finding groups
func BenchmarkFindGroupByAttribute(b *testing.B) {
	// Create test data
	groups := make([]*Group, 100)
	for i := 0; i < 100; i++ {
		groups[i] = &Group{
			ID:   ptr.String(fmt.Sprintf("group-%d", i)),
			Name: ptr.String(fmt.Sprintf("Group %d", i)),
			Attributes: &map[string][]string{
				"customID": {fmt.Sprintf("id-%d", i)},
			},
		}
	}

	attribute := GroupAttribute{
		Key:   "customID",
		Value: "id-50", // Middle of the list
	}

	b.ResetTimer()
	for b.Loop() {
		_, found := findGroupByAttribute(groups, attribute)
		if !found {
			b.Fatal("should find group")
		}
	}
}

// BenchmarkMapper benchmarks the mapper utility function
func BenchmarkMapper(b *testing.B) {
	params := SearchGroupParams{
		Search:              ptr.String("test"),
		BriefRepresentation: ptr.Bool(true),
		PopulateHierarchy:   ptr.Bool(true),
		Exact:               ptr.Bool(false),
		First:               ptr.Int(0),
		Max:                 ptr.Int(50),
		SubGroupsCount:      ptr.Bool(true),
	}

	b.ResetTimer()
	for b.Loop() {
		_, err := mapper(params)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkGetID benchmarks extracting ID from Location header
func BenchmarkGetID(b *testing.B) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Location", "https://keycloak.example.com/admin/realms/test-realm/groups/test-group-id-123-456-789")
		w.WriteHeader(http.StatusCreated)
	}))
	defer server.Close()

	restyClient := newTestRestyClient()

	b.ResetTimer()
	for b.Loop() {
		resp, err := restyClient.R().Post(server.URL)
		if err != nil {
			b.Fatal(err)
		}
		_ = getID(resp)
	}
}

// BenchmarkGroupsClient_ListPaginated benchmarks paginated group listing
func BenchmarkGroupsClient_ListPaginated(b *testing.B) {
	mockGroups := make([]*Group, 10)
	for i := 0; i < 10; i++ {
		mockGroups[i] = &Group{
			ID:   ptr.String(fmt.Sprintf("group-%d", i)),
			Name: ptr.String(fmt.Sprintf("Group %d", i)),
		}
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(mockGroups)
	}))
	defer server.Close()

	client := &Client{
		baseURL:  server.URL,
		realm:    "test-realm",
		pageSize: 50,
		resty:    newTestRestyClient(),
	}
	client.resty.SetBaseURL(server.URL)
	gc := &groupsClient{
		client: client,
	}

	ctx := context.Background()

	b.ResetTimer()
	for b.Loop() {
		_, err := gc.ListPaginated(ctx, nil, false, 0, 10)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkGroupsClient_Count benchmarks group counting
func BenchmarkGroupsClient_Count(b *testing.B) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(CountGroupResponse{Count: 42})
	}))
	defer server.Close()

	client := &Client{
		baseURL:  server.URL,
		realm:    "test-realm",
		pageSize: 50,
		resty:    newTestRestyClient(),
	}
	client.resty.SetBaseURL(server.URL)
	gc := &groupsClient{
		client: client,
	}

	ctx := context.Background()

	b.ResetTimer()
	for b.Loop() {
		_, err := gc.Count(ctx, nil, nil)
		if err != nil {
			b.Fatal(err)
		}
	}
}
