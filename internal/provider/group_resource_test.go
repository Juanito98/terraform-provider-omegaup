// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"fmt"
	"terraform-provider-omegaup/internal/mocks"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
)

func TestAccGroupResource(t *testing.T) {
	mockServer := mocks.NewMockServer()
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
