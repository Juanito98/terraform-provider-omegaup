// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package mocks

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"terraform-provider-omegaup/internal/apiclient"
)

type mockGroup struct {
	Group   *apiclient.Group
	Members map[string]struct{}
}

type state struct {
	MockGroups map[string]mockGroup
}

func NewMockServer() *httptest.Server {
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

		if strings.HasPrefix(r.URL.Path, "/api/group/") {
			groupHandler(state, payload, w, r)
			return
		}

		http.Error(w, "Not implemented", http.StatusNotImplemented)
	}))
}
