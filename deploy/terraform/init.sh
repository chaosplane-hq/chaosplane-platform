#!/usr/bin/env bash
set -euo pipefail

PROJECT="chaosplane"
ENVIRONMENT="prod"
REGION="ap-northeast-2"
BUCKET_NAME="${PROJECT}-${ENVIRONMENT}-terraform-state"
PROFILE="${1:-}"
PROFILE_FLAG=""

if [[ -n "${PROFILE}" ]]; then
  PROFILE_FLAG="--profile ${PROFILE}"
fi

echo "=== ChaosPlane Terraform State Bootstrap ==="
echo "Bucket: ${BUCKET_NAME}"
echo "Region: ${REGION}"
echo "Profile: ${PROFILE:-default}"
echo ""

if aws s3api head-bucket --bucket "${BUCKET_NAME}" ${PROFILE_FLAG} 2>/dev/null; then
  echo "Bucket ${BUCKET_NAME} already exists. Skipping creation."
else
  echo "Creating S3 bucket: ${BUCKET_NAME}"
  aws s3api create-bucket \
    --bucket "${BUCKET_NAME}" \
    --region "${REGION}" \
    --create-bucket-configuration LocationConstraint="${REGION}" \
    ${PROFILE_FLAG}

  echo "Enabling versioning..."
  aws s3api put-bucket-versioning \
    --bucket "${BUCKET_NAME}" \
    --versioning-configuration Status=Enabled \
    ${PROFILE_FLAG}

  echo "Enabling server-side encryption (AES256)..."
  aws s3api put-bucket-encryption \
    --bucket "${BUCKET_NAME}" \
    --server-side-encryption-configuration '{
      "Rules": [
        {
          "ApplyServerSideEncryptionByDefault": {
            "SSEAlgorithm": "AES256"
          },
          "BucketKeyEnabled": true
        }
      ]
    }' \
    ${PROFILE_FLAG}

  echo "Blocking public access..."
  aws s3api put-public-access-block \
    --bucket "${BUCKET_NAME}" \
    --public-access-block-configuration '{
      "BlockPublicAcls": true,
      "IgnorePublicAcls": true,
      "BlockPublicPolicy": true,
      "RestrictPublicBuckets": true
    }' \
    ${PROFILE_FLAG}

  echo "Adding lifecycle rule for noncurrent versions (30 days)..."
  aws s3api put-bucket-lifecycle-configuration \
    --bucket "${BUCKET_NAME}" \
    --lifecycle-configuration '{
      "Rules": [
        {
          "ID": "expire-noncurrent-versions",
          "Status": "Enabled",
          "Filter": {},
          "NoncurrentVersionExpiration": {
            "NoncurrentDays": 30
          }
        }
      ]
    }' \
    ${PROFILE_FLAG}

  echo ""
  echo "=== Bucket created successfully ==="
fi

echo ""
echo "Usage: ./init.sh [aws-profile]"
echo "  e.g. ./init.sh chaosplane"
echo ""
echo "Add this to your backend.tf:"
echo ""
echo 'terraform {'
echo '  backend "s3" {'
echo "    bucket       = \"${BUCKET_NAME}\""
echo '    key          = "prod/terraform.tfstate"'
echo "    region       = \"${REGION}\""
echo '    encrypt      = true'
echo '    use_lockfile = true'
if [[ -n "${PROFILE}" ]]; then
  echo "    profile      = \"${PROFILE}\""
fi
echo '  }'
echo '}'
