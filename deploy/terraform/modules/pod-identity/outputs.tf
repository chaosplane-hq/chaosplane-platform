output "role_arns" {
  value = {
    aws-load-balancer-controller = module.lb_controller.iam_role_arn
    external-secrets             = module.external_secrets.iam_role_arn
    cert-manager                 = module.cert_manager.iam_role_arn
    fluent-bit                   = module.fluent_bit.iam_role_arn
    chaosplane-api               = module.chaosplane_api.iam_role_arn
  }
}
