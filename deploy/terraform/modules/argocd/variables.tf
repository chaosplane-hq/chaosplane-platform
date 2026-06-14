variable "chart_version" {
  type    = string
  default = "9.5.17"
}

variable "tags" {
  type    = map(string)
  default = {}
}
