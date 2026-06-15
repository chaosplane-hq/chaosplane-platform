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

variable "sites" {
  type = map(object({
    origin_host = string
    domain      = optional(string)
  }))
}

variable "acm_certificate_arn" {
  type    = string
  default = ""
}
