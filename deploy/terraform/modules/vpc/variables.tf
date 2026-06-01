variable "name" {
  type = string
}

variable "tags" {
  type    = map(string)
  default = {}
}

variable "flow_log_bucket_arn" {
  type    = string
  default = null
}
