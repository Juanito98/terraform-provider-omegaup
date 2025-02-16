// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package apiclient

import "encoding/json"

type Group struct {
	Alias       string `json:"alias"`
	Description string `json:"description"`
	Name        string `json:"name"`
}

func (c *Client) GroupCreate(req *GroupCreateRequest) error {
	_, err := c.query("/api/group/create", req)
	if err != nil {
		return err
	}
	return nil
}

type GroupCreateRequest Group

func (c *Client) GroupDetails(req *GroupDetailsRequest) (*GroupDetailsResponse, error) {
	var res *GroupDetailsResponse
	bytes, err := c.query("/api/group/details", req)
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal(bytes, &res); err != nil {
		return nil, err
	}

	return res, nil
}

type GroupDetailsRequest struct {
	GroupAlias string `json:"group_alias"`
}

type GroupDetailsResponseGroup struct {
	Alias       string `json:"alias"`
	Description string `json:"description"`
	Name        string `json:"name"`
	CreateTime  int    `json:"create_time"`
}
type GroupDetailsResponse struct {
	Group GroupDetailsResponseGroup `json:"group"`
}

func (c *Client) GroupUpdate(req *GroupUpdateRequest) error {
	_, err := c.query("/api/group/update", req)
	return err
}

type GroupUpdateRequest Group

func (c *Client) GroupAddUser(req *GroupAddUserRequest) error {
	_, err := c.query("/api/group/addUser", req)
	return err
}

type GroupAddUserRequest struct {
	GroupAlias      string `json:"group_alias"`
	UsernameOrEmail string `json:"usernameOrEmail"`
}

func (c *Client) GroupMembers(req *GroupMembersRequest) (*GroupMembersResponse, error) {
	var res *GroupMembersResponse
	bytes, err := c.query("/api/group/members", req)
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal(bytes, &res); err != nil {
		return nil, err
	}

	return res, nil
}

type GroupMembersRequest struct {
	GroupAlias string `json:"group_alias"`
}

type Identity struct {
	Username string `json:"username"`
}

type GroupMembersResponse struct {
	Identities []Identity `json:"identities"`
}

func (c *Client) GroupRemoveUser(req *GroupRemoveUserRequest) error {
	_, err := c.query("/api/group/removeUser", req)
	return err
}

type GroupRemoveUserRequest GroupAddUserRequest
