variable "name" {
  type = string
}

variable "secrets" {
  type = map(string)
}

variable "generated_secrets" {
  type    = set(string)
  default = []
}

variable "kms_key_arn" {
  type = string
}

variable "tags" {
  type    = map(string)
  default = {}
}
