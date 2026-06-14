variable "domain" {
  type = string
}

variable "tags" {
  type    = map(string)
  default = {}
}

variable "wait_for_validation" {
  type    = bool
  default = false # flip to true after NS delegation completes
}
