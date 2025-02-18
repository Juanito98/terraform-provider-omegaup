// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package mocks

import (
	"encoding/json"
	"fmt"
	"net/http"
	"terraform-provider-omegaup/internal/apiclient"
)

func groupHandler(state state, payload []byte, w http.ResponseWriter, r *http.Request) {
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
			identities := make([]apiclient.GroupIdentity, 0, len(data.Members))
			for key := range data.Members {
				identities = append(identities, apiclient.GroupIdentity{Username: key}) // Extract keys into slice
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
}
