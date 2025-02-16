// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"terraform-provider-omegaup/internal/apiclient"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
)

type mockGroup struct {
	Group   *apiclient.Group
	Members map[string]struct{}
}

type state struct {
	MockGroups map[string]mockGroup
}

func newMockServer() *httptest.Server {
	state := state{MockGroups: make(map[string]mockGroup)}
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		err := r.ParseMultipartForm(10 << 20)
		if err != nil {
			http.Error(w, "Error decoding body", http.StatusInternalServerError)
			return
		}
		// Convert form values to JSON
		formData := make(map[string]string)
		for key, values := range r.Form {
			if len(values) > 0 {
				formData[key] = values[0]
			}
		}
		payload, err := json.Marshal(formData)
		if err != nil {
			http.Error(w, "Error encoding form data to JSON", http.StatusBadRequest)
			return
		}

		if r.URL.Path == "/api/group/create" {
			var req *apiclient.GroupCreateRequest
			if err := json.Unmarshal(payload, &req); err != nil {
				http.Error(w, "Error decoding form data to JSON", http.StatusBadRequest)
				return
			}
			state.MockGroups[req.Alias] = mockGroup{
				Group:   (*apiclient.Group)(req),
				Members: make(map[string]struct{}),
			}
			w.WriteHeader(http.StatusOK)
			return
		}

		if r.URL.Path == "/api/group/update" {
			var req *apiclient.GroupUpdateRequest
			if err := json.Unmarshal(payload, &req); err != nil {
				http.Error(w, "Error decoding form data to JSON", http.StatusBadRequest)
				return
			}
			entry := state.MockGroups[req.Alias]
			entry.Group = (*apiclient.Group)(req)
			state.MockGroups[req.Alias] = entry
			w.WriteHeader(http.StatusOK)
			return
		}

		if r.URL.Path == "/api/group/details" {
			var req *apiclient.GroupDetailsRequest
			if err := json.Unmarshal(payload, &req); err != nil {
				http.Error(w, "Error decoding form data to JSON", http.StatusBadRequest)
				return
			}
			if data, exists := state.MockGroups[req.GroupAlias]; exists {
				res, err := json.Marshal(&apiclient.GroupDetailsResponse{
					Group: apiclient.GroupDetailsResponseGroup{
						Alias:       data.Group.Alias,
						Description: data.Group.Description,
						Name:        data.Group.Name},
				})
				if err != nil {
					http.Error(w, "Marshalling response", http.StatusInternalServerError)
				}
				if _, err = w.Write(res); err != nil {
					http.Error(w, "Writing response", http.StatusInternalServerError)
				}
				w.WriteHeader(http.StatusOK)
				return
			} else {
				http.Error(w, fmt.Sprintf("Group %s does not exists", req.GroupAlias), http.StatusNotFound)
				return
			}
		}

		if r.URL.Path == "/api/group/addUser" {
			var req *apiclient.GroupAddUserRequest
			if err := json.Unmarshal(payload, &req); err != nil {
				http.Error(w, "Error decoding form data to JSON", http.StatusBadRequest)
				return
			}
			if data, exists := state.MockGroups[req.GroupAlias]; exists {
				data.Members[req.UsernameOrEmail] = struct{}{}
				state.MockGroups[req.GroupAlias] = data
				w.WriteHeader(http.StatusOK)
				return
			} else {
				http.Error(w, fmt.Sprintf("Group %s does not exists", req.GroupAlias), http.StatusNotFound)
				return
			}
		}

		if r.URL.Path == "/api/group/removeUser" {
			var req *apiclient.GroupRemoveUserRequest
			if err := json.Unmarshal(payload, &req); err != nil {
				http.Error(w, "Error decoding form data to JSON", http.StatusBadRequest)
				return
			}
			if data, exists := state.MockGroups[req.GroupAlias]; exists {
				delete(data.Members, req.UsernameOrEmail)
				state.MockGroups[req.GroupAlias] = data
				w.WriteHeader(http.StatusOK)
				return
			} else {
				http.Error(w, fmt.Sprintf("Group %s does not exists", req.GroupAlias), http.StatusNotFound)
				return
			}
		}

		if r.URL.Path == "/api/group/members" {
			var req *apiclient.GroupMembersRequest
			if err := json.Unmarshal(payload, &req); err != nil {
				http.Error(w, "Error decoding form data to JSON", http.StatusBadRequest)
				return
			}
			if data, exists := state.MockGroups[req.GroupAlias]; exists {
				identities := make([]apiclient.Identity, 0, len(data.Members))
				for key := range data.Members {
					identities = append(identities, apiclient.Identity{Username: key}) // Extract keys into slice
				}
				res, err := json.Marshal(&apiclient.GroupMembersResponse{
					Identities: identities,
				})
				if err != nil {
					http.Error(w, "Marshalling response", http.StatusInternalServerError)
				}
				if _, err = w.Write(res); err != nil {
					http.Error(w, "Writing response", http.StatusInternalServerError)
				}
				w.WriteHeader(http.StatusOK)
				return
			} else {
				http.Error(w, fmt.Sprintf("Group %s does not exists", req.GroupAlias), http.StatusNotFound)
				return
			}
		}
		http.Error(w, "Not implemented", http.StatusNotImplemented)
	}))
}

func TestAccGroupResource(t *testing.T) {
	mockServer := newMockServer()
	defer mockServer.Close()

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: provider_config(mockServer.URL) + testAccGroupResourceConfig("admins", "description"),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"omegaup_group.test",
						tfjsonpath.New("alias"),
						knownvalue.StringExact("admins"),
					),
					statecheck.ExpectKnownValue(
						"omegaup_group.test",
						tfjsonpath.New("description"),
						knownvalue.StringExact("description"),
					),
					statecheck.ExpectKnownValue(
						"omegaup_group.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact("admins"),
					),
				},
			},
			// ImportState testing
			{
				ResourceName:                         "omegaup_group.test",
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateId:                        "admins",
				ImportStateVerifyIdentifierAttribute: "alias",
			},
			// Update and Read testing
			{
				Config: provider_config(mockServer.URL) + testAccGroupResourceConfig("admins", "description changed"),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"omegaup_group.test",
						tfjsonpath.New("alias"),
						knownvalue.StringExact("admins"),
					),
					statecheck.ExpectKnownValue(
						"omegaup_group.test",
						tfjsonpath.New("description"),
						knownvalue.StringExact("description changed"),
					),

					statecheck.ExpectKnownValue(
						"omegaup_group.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact("admins"),
					),
				},
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}

func testAccGroupResourceConfig(alias string, description string) string {
	return fmt.Sprintf(`
resource "omegaup_group" "test" {
  alias = %[1]q
  description = %[2]q
}
`, alias, description)
}
