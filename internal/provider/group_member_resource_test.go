// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
)

func TestAccGroupMemberResource(t *testing.T) {
	mockServer := newMockServer()
	defer mockServer.Close()

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: provider_config(mockServer.URL) + testAccGroupMemberResourceConfig("admins", "test"),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"omegaup_group_member.member",
						tfjsonpath.New("group_alias"),
						knownvalue.StringExact("admins"),
					),
					statecheck.ExpectKnownValue(
						"omegaup_group_member.member",
						tfjsonpath.New("username"),
						knownvalue.StringExact("test"),
					),
				},
			},
			// ImportState testing
			{
				ResourceName:                         "omegaup_group_member.member",
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateId:                        "admins,test",
				ImportStateVerifyIdentifierAttribute: "group_alias",
			},
			// Update and Read testing
			{
				Config: provider_config(mockServer.URL) + testAccGroupMemberResourceConfig("admins", "other"),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"omegaup_group_member.member",
						tfjsonpath.New("group_alias"),
						knownvalue.StringExact("admins"),
					),
					statecheck.ExpectKnownValue(
						"omegaup_group_member.member",
						tfjsonpath.New("username"),
						knownvalue.StringExact("other"),
					),
				},
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}

func testAccGroupMemberResourceConfig(alias string, member string) string {
	return fmt.Sprintf(`
resource "omegaup_group" "group" {
  alias = %[1]q
  description = "description"
}
resource "omegaup_group_member" "member" {
	group_alias = omegaup_group.group.alias
	username = %[2]q
}
`, alias, member)
}
