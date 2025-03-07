---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "omegaup_group_member Resource - omegaup"
subcategory: ""
description: |-
  Creates a new user permission for a group.
---

# omegaup_group_member (Resource)

Creates a new user permission for a group.

## Example Usage

```terraform
resource "omegaup_group_member" "member" {
  group_alias = "alias"
  username    = "user"
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `group_alias` (String) The alias used to identify the group.
- `username` (String) OmegaUp username to add to the group.

## Import

Import is supported using the following syntax:

```shell
terraform import omegaup_group_member.member alias,user
```
