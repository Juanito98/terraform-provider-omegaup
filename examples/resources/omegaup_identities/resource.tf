resource "omegaup_identities" "identity" {
  group_alias = "group"
  identities = [
    {
      username    = "group:user1"
      name        = "Name"
      gender      = "other"
      password    = "password1"
      school_name = "OFMI"
      country_id  = "MX"
      state_id    = "MEX"
    },
  ]
}
