variable "name" {
  type = string
}

variable "sites" {
  type = map(object({
    domain = string
  }))
}

variable "certificate_arn_us_east_1" {
  type = string
}

variable "tags" {
  type    = map(string)
  default = {}
}
