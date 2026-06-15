variable "project" {
  type    = string
  default = "chaosplane"
}

variable "environment" {
  type    = string
  default = "prod"
}

variable "region" {
  type    = string
  default = "ap-northeast-2"
}

variable "aws_profile" {
  type    = string
  default = "chaosplane"
}

variable "enable_edge" {
  type    = bool
  default = true
}
