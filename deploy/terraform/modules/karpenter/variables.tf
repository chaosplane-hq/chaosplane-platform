variable "name" {
  type = string
}

variable "cluster_name" {
  type = string
}

variable "cluster_endpoint" {
  type = string
}

variable "node_security_group_id" {
  type = string
}

variable "tags" {
  type    = map(string)
  default = {}
}
