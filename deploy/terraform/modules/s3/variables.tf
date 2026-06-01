variable "name" {
  type = string
}

variable "buckets" {
  type = map(object({
    versioning = bool
  }))
}

variable "kms_key_arn" {
  type = string
}

variable "tags" {
  type    = map(string)
  default = {}
}
