resource "omegaup_group" "groups" {
  for_each    = tomap({ for _, group in var.groups : group.alias => group })
  alias       = each.key
  description = each.value.description
}

resource "omegaup_group_member" "members" {
  for_each = merge(
    [for _, group in var.groups : tomap({
      for _, member in group.members :
      "${group.alias}_${member}" => {
        alias    = group.alias
        username = member
      }
      })
    ]...
  )

  group_alias = each.value.alias
  username    = each.value.username
}
