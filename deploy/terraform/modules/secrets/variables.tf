variable "name" {
  type = string
}

variable "secrets" {
  type = map(string)
}

variable "kms_key_arn" {
  type = string
}

variable "tags" {
  type    = map(string)
  default = {}
}
