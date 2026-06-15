# ChaosPlane Infrastructure Deployment

## Prerequisites

- AWS CLI configured with profile `chaosplane`
- Terraform >= 1.15.0
- kubectl
- helm
- Domain `chaosplane.dev` owned (registrar access needed for NS delegation)

## Initial Deployment

### 1. Bootstrap Terraform State

```bash
cd deploy/terraform
./init.sh chaosplane
```

### 2. Configure Variables

```bash
cd envs/prod
cp terraform.tfvars.example terraform.tfvars
# Edit terraform.tfvars with your values
```

### 3. Terraform Apply

```bash
terraform init
terraform plan
terraform apply
```

What gets created:

- VPC, public/private subnets, route tables, NAT gateways, and security groups
- EKS cluster, managed node groups, IAM roles, and service account roles
- ECR repositories for `chaosplane/api` and `chaosplane/web`
- RDS PostgreSQL instance, RDS Proxy, and related IAM auth configuration
- ElastiCache Redis cluster for application cache/session storage
- AWS Secrets Manager secrets for application runtime configuration
- Route 53 hosted zone and DNS records for `chaosplane.dev`
- AWS Load Balancer Controller prerequisites and ingress-related IAM resources
- GitHub Actions OIDC identity provider for short-lived AWS credentials

### 4. NS Delegation (one-time, manual)

```bash
terraform output zone_name_servers
```

Go to your domain registrar and set the NS records to these 4 values. DNS propagation takes up to 48 hours.

### 5. Populate Secrets (one-time)

```bash
aws secretsmanager put-secret-value --secret-id chaosplane/prod/jwt-secret --secret-string "<generate-random-32-char>" --profile chaosplane
aws secretsmanager put-secret-value --secret-id chaosplane/prod/csrf-secret --secret-string "<generate-random-32-char>" --profile chaosplane
```

Note: `redis-url` is automatically populated by Terraform from the ElastiCache endpoint. `db-url` is handled via RDS Proxy IAM auth, no password needed in secrets.

### 6. Configure kubeconfig

```bash
aws eks update-kubeconfig --name chaosplane-prod --region ap-northeast-2 --profile chaosplane
```

### 7. Bootstrap ArgoCD

ArgoCD is installed declaratively via Terraform (`module.argocd`, a `helm_release`
of the `argo-cd` chart pinned to a specific version) — not via raw `kubectl apply`.
It is created automatically during `terraform apply`. To (re)apply just ArgoCD:

```bash
cd deploy/terraform/envs/prod
terraform apply -target=module.argocd

# Wait for ArgoCD to be ready
kubectl wait --for=condition=available deployment/argocd-server -n argocd --timeout=300s

# Get initial admin password (rotate / delete after first login)
kubectl -n argocd get secret argocd-initial-admin-secret -o jsonpath="{.data.password}" | base64 -d
```

The chart version is pinned in `deploy/terraform/modules/argocd/variables.tf`
(`chart_version`) for reproducibility. ArgoCD pods run on the system node group
(tolerations + nodeSelector are set in the module).

### 8. Apply ArgoCD Applications

After the repo credential is in place (see "ArgoCD → Private Repo Access" below),
apply the root Applications:

```bash
kubectl apply -f deploy/argocd/
```

This triggers automatic sync of all workloads.

## ArgoCD → Private Repo Access (Fine-grained PAT)

ArgoCD needs read access to `chaosplane-platform` (private repo) to pull K8s manifests.
We use a **fine-grained Personal Access Token** scoped to a single repository with
read-only contents. The token is injected directly into the cluster as a Secret via
stdin and is **never written to a file or committed to the repo**.

### 1. Create the fine-grained PAT (browser only — GitHub has no API for this)

1. Go to https://github.com/settings/personal-access-tokens/new
2. Settings:
   - Token name: `chaosplane-argocd-repo`
   - Resource owner: `chaosplane-hq`
   - Expiration: 90 days (rotate before expiry — see rotation below)
   - Repository access: **Only select repositories** → `chaosplane-hq/chaosplane-platform`
   - Repository permissions → **Contents: Read-only** (this is the only permission needed)
3. Generate token and copy it (starts with `github_pat_...`).

### 2. Inject the credential into ArgoCD (stdin, never saved to disk)

```bash
# Paste the token when prompted; it is read into a shell variable and never written to a file.
read -rs GH_PAT
kubectl create secret generic chaosplane-platform-repo -n argocd \
  --from-literal=type=git \
  --from-literal=url=https://github.com/chaosplane-hq/chaosplane-platform.git \
  --from-literal=username=chaosplane-argocd \
  --from-literal=password="$GH_PAT" \
  --dry-run=client -o yaml \
  | kubectl label -f - --local -o yaml argocd.argoproj.io/secret-type=repository \
  | kubectl apply -f -
unset GH_PAT
```

> Security notes:
> - The token value lives only in cluster Secret storage (etcd, encrypted via the EKS KMS key) — never in a file, never in git.
> - Use a fine-grained PAT (single repo, Contents read-only), not a classic token or the broad `gh` OAuth token. The latter has `repo`, `delete_repo`, `workflow` scopes and violates least privilege.
> - For a temporary bring-up you may inject `$(gh auth token)` instead, but **replace it with the fine-grained PAT before going live** — the OAuth token is over-privileged.

### 3. Rotate the token

Before expiry, create a new fine-grained PAT and re-run step 2 (it overwrites the
existing Secret). ArgoCD picks up the new credential on its next repo poll. Then
delete the old PAT from GitHub.

### Alternative: GitHub App

If you prefer a GitHub App (no expiry, finer audit trail), create one with
`Contents: Read-only`, install it on `chaosplane-platform`, and inject a
`repo-creds` Secret with `githubAppID` / `githubAppInstallationID` /
`githubAppPrivateKey`. The PAT approach above is simpler and is the supported default.

## GitHub Actions Setup (OIDC → ECR Push)

GitHub Actions needs to push Docker images to ECR without long-lived credentials.

### 1. Create OIDC Identity Provider (already done by Terraform)

Terraform creates `aws_iam_openid_connect_provider` for `token.actions.githubusercontent.com`.

### 2. Create IAM Role for GitHub Actions

Create a role with trust policy allowing the chaosplane-platform repo:

```bash
aws iam create-role --role-name github-actions-ecr --assume-role-policy-document '{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Principal": {
        "Federated": "arn:aws:iam::<ACCOUNT_ID>:oidc-provider/token.actions.githubusercontent.com"
      },
      "Action": "sts:AssumeRoleWithWebIdentity",
      "Condition": {
        "StringEquals": {
          "token.actions.githubusercontent.com:aud": "sts.amazonaws.com"
        },
        "StringLike": {
          "token.actions.githubusercontent.com:sub": "repo:chaosplane-hq/chaosplane-platform:*"
        }
      }
    }
  ]
}' --profile chaosplane
```

### 3. Attach ECR Push Policy

```bash
aws iam put-role-policy --role-name github-actions-ecr --policy-name ecr-push --policy-document '{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Action": [
        "ecr:GetAuthorizationToken"
      ],
      "Resource": "*"
    },
    {
      "Effect": "Allow",
      "Action": [
        "ecr:BatchCheckLayerAvailability",
        "ecr:GetDownloadUrlForLayer",
        "ecr:BatchGetImage",
        "ecr:PutImage",
        "ecr:InitiateLayerUpload",
        "ecr:UploadLayerPart",
        "ecr:CompleteLayerUpload"
      ],
      "Resource": [
        "arn:aws:ecr:ap-northeast-2:<ACCOUNT_ID>:repository/chaosplane/api",
        "arn:aws:ecr:ap-northeast-2:<ACCOUNT_ID>:repository/chaosplane/web"
      ]
    }
  ]
}' --profile chaosplane
```

### 4. Configure GitHub Repository Variable

Go to https://github.com/chaosplane-hq/chaosplane-platform/settings/variables/actions

Add variable:

- Name: `AWS_DEPLOY_ROLE_ARN`
- Value: `arn:aws:iam::<ACCOUNT_ID>:role/github-actions-ecr`

### 5. Test

Push to main. The deploy workflow should:

1. Assume the IAM role via OIDC
2. Login to ECR
3. Build and push api + web images
4. Update kustomization.yaml image tag
5. ArgoCD detects change → deploys

## Day-to-Day Deployment

```bash
git push origin main
```

That's it. GitHub Actions builds images, updates manifests, ArgoCD deploys.

## Rollback

### App rollback (git revert)

```bash
git revert HEAD
git push origin main
# ArgoCD auto-syncs to previous state
```

### ArgoCD manual rollback

```bash
argocd app rollback chaosplane-system <previous-revision>
```

### RDS Point-in-Time Recovery

```bash
aws rds restore-db-instance-to-point-in-time \
  --source-db-instance-identifier chaosplane-prod \
  --target-db-instance-identifier chaosplane-prod-restored \
  --restore-time "2026-06-01T12:00:00Z" \
  --profile chaosplane
```

## Architecture

See `.sisyphus/plans/infra-aws-eks.md` for the full architecture document.

## Edge / TLS Topology

Traffic flows: **CloudFront (HTTPS) → ALB (HTTP :80) → go-api / web pods**.

- The ALB is HTTP-only because the custom-domain ACM certificate cannot be
  issued until the `chaosplane.dev` nameservers are delegated to Route 53.
- CloudFront terminates TLS. Two distributions exist (`api`, `app`), each
  injecting an `X-Site` header (`api` / `app`) so the single ALB routes by
  header, plus a shared `X-Origin-Verify` secret header.
- A WAF rule on the ALB blocks any request lacking the correct
  `X-Origin-Verify` value, so the public ALB cannot be reached directly
  (bypassing CloudFront returns 403). The secret lives only in Terraform
  state / WAF, never in git.

Until NS delegation, CloudFront serves on its default `*.cloudfront.net`
domains. Current distributions:

```
api → https://d24111gb9yg8wo.cloudfront.net
app → https://d1gbvkrts3frks.cloudfront.net
```

## Post-NS-Delegation: Custom Domain + HTTPS Everywhere

These steps require the `chaosplane.dev` nameservers to point at Route 53.

### 1. Delegate nameservers

`chaosplane.dev` currently resolves to Cloudflare. Point its NS records at the
Route 53 hosted zone (created by `module.dns`):

```
ns-63.awsdns-07.com
ns-1260.awsdns-29.org
ns-666.awsdns-19.net
ns-1944.awsdns-51.co.uk
```

Verify propagation: `dig +short NS chaosplane.dev` should return the AWS NS set.

### 2. Issue the ACM certificate

Flip `wait_for_validation` to `true` so Terraform blocks until the DNS-validated
cert is ISSUED (validation records live in the new Route 53 zone):

```bash
cd deploy/terraform/envs/prod
terraform apply -var='dns_wait_for_validation=true'   # or set in the dns module call
```

The us-east-1 certificate ARN is exposed as `module.dns.certificate_arn_us_east_1`
(CloudFront requires the cert in us-east-1).

### 3. Attach custom domains to CloudFront

Wire the domains and cert into the `cdn_alb` module call in `main.tf`:

```hcl
module "cdn_alb" {
  # ...
  acm_certificate_arn = module.dns.certificate_arn_us_east_1
  sites = {
    api = { origin_host = "api.chaosplane.dev", domain = "api.chaosplane.dev" }
    app = { origin_host = "app.chaosplane.dev", domain = "app.chaosplane.dev" }
  }
}
```

`terraform apply` then adds the aliases + SNI viewer cert to each distribution.

### 4. Route 53 alias records

Add A/AAAA alias records pointing `api.chaosplane.dev` / `app.chaosplane.dev`
at their CloudFront distributions (in the `dns` module or via the
`cdn_alb` outputs `domain_names` / `distribution_ids`).

### 5. (Optional) HTTPS origin

Once a regional ACM cert is attached to the ALB and a 443 listener is added,
flip the CloudFront origin `origin_protocol_policy` to `https-only`. This is a
distribution-only change that deploys in minutes with no downtime.

## Rotating the ArgoCD repo token

The ArgoCD repo credential is currently a temporary `gh` OAuth token
(over-privileged, `gho_...`). Replace it with a fine-grained PAT
(repo-scoped, Contents read-only) per the "ArgoCD → Private Repo Access"
section above. The injection command overwrites the existing
`chaosplane-platform-repo` Secret; the token value never touches git.
