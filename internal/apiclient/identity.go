package apiclient

import "encoding/json"

type Identity struct {
	GroupAlias string `json:"group_alias"`
	Username   string `json:"username"`
	Name       string `json:"name"`
	Gender     string `json:"gender"`
	Password   string `json:"password"`
	SchoolName string `json:"school_name"`
	CountryId  string `json:"country_id"`
	StateId    string `json:"state_id"`
}

func (c *Client) IdentityCreate(req *IdentityCreateRequest) (*IdentityCreateResponse, error) {
	var res *IdentityCreateResponse
	bytes, err := c.query("/api/identity/create", req)
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal(bytes, &res); err != nil {
		return nil, err
	}

	return res, nil
}

type IdentityCreateRequest Identity

type IdentityCreateResponse struct {
	Username string `json:"username"`
}

func (c *Client) IdentityUpdate(req *IdentityUpdateRequest) error {
	_, err := c.query("/api/identity/update", req)
	return err
}

type IdentityUpdateRequest struct {
	GroupAlias       string `json:"group_alias"`
	OriginalUsername string `json:"original_username"`
	Username         string `json:"username"`
	Name             string `json:"name"`
	Gender           string `json:"gender"`
	SchoolName       string `json:"school_name"`
	CountryId        string `json:"country_id"`
	StateId          string `json:"state_id"`
}

func (c *Client) IdentityChangePassword(req *IdentityChangePasswordRequest) error {
	_, err := c.query("/api/identity/changePassword", req)
	return err
}

type IdentityChangePasswordRequest struct {
	GroupAlias string `json:"group_alias"`
	Username   string `json:"username"`
	Password   string `json:"password"`
}

func (c *Client) IdentityBulkCreate(req *IdentityBulkCreateRequest) error {
	identitiesJson, err := json.Marshal(req.Identities)
	if err != nil {
		return err
	}

	_, err = c.query("/api/identity/bulkCreate", map[string]string{
		"group_alias": req.GroupAlias,
		"identities":  string(identitiesJson),
	})
	return err
}

type IdentityBulkCreateRequest struct {
	GroupAlias string
	Identities []Identity
}
