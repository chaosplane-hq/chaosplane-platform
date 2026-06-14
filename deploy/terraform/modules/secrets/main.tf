locals {
  all_secret_keys = nonsensitive(toset(concat(keys(var.secrets), tolist(var.generated_secrets))))
}

resource "aws_secretsmanager_secret" "this" {
  for_each = local.all_secret_keys

  name       = "${var.name}/${each.key}"
  kms_key_id = var.kms_key_arn
  tags       = var.tags
}

resource "random_password" "this" {
  for_each = var.generated_secrets

  length  = 48
  special = false
}

resource "aws_secretsmanager_secret_version" "this" {
  for_each = local.all_secret_keys

  secret_id = aws_secretsmanager_secret.this[each.key].id
  secret_string = contains(var.generated_secrets, each.key) ? (
    random_password.this[each.key].result
  ) : var.secrets[each.key]

  lifecycle {
    ignore_changes = [secret_string]
  }
}
