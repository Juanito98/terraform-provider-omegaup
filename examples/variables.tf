variable "groups" {
  type = list(object({
    alias       = string
    description = string
    members     = set(string)
  }))
}
