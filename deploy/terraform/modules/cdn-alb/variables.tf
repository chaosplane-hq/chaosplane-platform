variable "name" {
  type = string
}

variable "alb_name" {
  type = string
}

variable "origin_verify_secret" {
  type      = string
  sensitive = true
}
