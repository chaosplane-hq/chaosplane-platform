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

```bash
kubectl create namespace argocd
kubectl apply -n argocd -f https://raw.githubusercontent.com/argoproj/argo-cd/v2.14.0/manifests/install.yaml

# Wait for ArgoCD to be ready
kubectl wait --for=condition=available deployment/argocd-server -n argocd --timeout=300s

# Get initial admin password
kubectl -n argocd get secret argocd-initial-admin-secret -o jsonpath="{.data.password}" | base64 -d
```

### 8. Apply ArgoCD Applications

```bash
kubectl apply -f deploy/argocd/
```

This triggers automatic sync of all workloads.

## GitHub App Setup (ArgoCD → Private Repo Access)

ArgoCD needs to read `chaosplane-platform` (private repo) for K8s manifests.

### Create GitHub App

1. Go to https://github.com/organizations/chaosplane-hq/settings/apps/new
2. Settings:
   - Name: `chaosplane-argocd`
   - Homepage URL: `https://app.chaosplane.dev`
   - Uncheck "Webhook Active"
   - Permissions:
     - Repository permissions → Contents: Read-only
   - Where can this app be installed: Only on this account
3. Click "Create GitHub App"
4. Note the **App ID** (shown at top of app settings page)
5. Generate a **Private Key** (scroll down → "Generate a private key") — downloads a .pem file
6. Install the app:
   - Go to "Install App" tab
   - Install on `chaosplane-hq`
   - Select "Only select repositories" → choose `chaosplane-platform`
   - Note the **Installation ID** from the URL: `https://github.com/organizations/chaosplane-hq/settings/installations/<INSTALLATION_ID>`

### Configure ArgoCD with GitHub App

```bash
kubectl apply -f - <<EOF
apiVersion: v1
kind: Secret
metadata:
  name: github-app-creds
  namespace: argocd
  labels:
    argocd.argoproj.io/secret-type: repo-creds
stringData:
  url: https://github.com/chaosplane-hq/
  githubAppID: "<APP_ID>"
  githubAppInstallationID: "<INSTALLATION_ID>"
  githubAppPrivateKey: |
    <paste contents of .pem file>
EOF
```

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
