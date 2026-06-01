variable "name" {
  type = string
}

variable "cluster_name" {
  type = string
}

variable "rds_proxy_resource_id" {
  type = string
}

variable "s3_bucket_arns" {
  type = list(string)
}

variable "tags" {
  type    = map(string)
  default = {}
}
