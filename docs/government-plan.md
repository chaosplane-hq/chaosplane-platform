# ChaosPlane Government Plan — Features and Architecture

> This document describes the ChaosPlane Government Plan: a dedicated offering for US federal agencies, DoD contractors, and Korean public institutions. It covers air-gap deployment, CAC/PIV authentication, DISA STIG compliance, and CUI marking. The Government Plan is planned for launch in Phase 4 (Month 18-20).
>
> Features described here are in active development or preparation. This document does not represent a generally available product offering at the time of writing.

---

## Overview

The ChaosPlane Government Plan is a hardened, compliance-ready edition of the ChaosPlane platform designed for environments with strict security, connectivity, and regulatory requirements. It builds on the commercial Enterprise Plan and adds:

- Air-gap deployment for disconnected or classified-adjacent environments
- CAC/PIV smart card authentication via agency identity providers
- DISA STIG compliance for all platform components
- CUI marking and handling system
- FedRAMP Moderate authorization (in preparation)
- CMMC Level 2 support
- Dedicated AWS GovCloud (us-gov-west-1) infrastructure
- Korean public sector deployment on NHN Cloud / NCP (CSAP 표준등급)

---

## Air-Gap Deployment

### What Air-Gap Means for ChaosPlane

An air-gap deployment runs ChaosPlane in an environment with no internet connectivity. This is required for:

- Classified-adjacent networks (SIPRNet-adjacent, IL4/IL5 environments)
- Sensitive compartmented information facilities (SCIFs)
- Industrial control system (ICS) environments
- Korean public institution networks with strict egress controls
- On-premises deployments where outbound internet access is prohibited by policy

### Air-Gap Bundle

ChaosPlane provides a self-contained air-gap bundle that includes everything needed to run the platform with zero external dependencies:

```
chaosplane-airgap-v{VERSION}/
├── images/
│   ├── chaosplane-api.tar.gz          # API server container image
│   ├── chaosplane-agent.tar.gz        # Chaos agent container image
│   ├── chaosplane-web.tar.gz          # Web frontend container image
│   ├── chaosplane-workflow.tar.gz     # Workflow engine container image
│   ├── chaosplane-identity.tar.gz     # Identity service container image
│   ├── chaosplane-auditlog.tar.gz     # Audit log service container image
│   ├── postgres-15.tar.gz             # PostgreSQL database image
│   ├── redis-7.tar.gz                 # Redis image
│   └── cert-manager.tar.gz            # Certificate manager image
├── helm/
│   ├── chaosplane/                    # Main Helm chart
│   ├── chaosplane-agent/              # Agent Helm chart
│   └── values-airgap.yaml             # Air-gap specific values
├── manifests/
│   ├── namespace.yaml
│   ├── rbac.yaml
│   └── network-policies.yaml
├── scripts/
│   ├── load-images.sh                 # Load all images into local registry
│   ├── preflight-check.sh             # Verify prerequisites
│   └── install.sh                     # Full installation script
├── docs/
│   ├── installation-guide.md
│   ├── upgrade-guide.md
│   └── troubleshooting.md
└── checksums.sha256                   # SHA-256 checksums for all files
```

The bundle is signed with ChaosPlane's GPG key. Customers verify the signature before installation.

### Air-Gap Architecture

In an air-gap deployment, all external dependencies are eliminated:

| Dependency | Commercial SaaS | Air-Gap Replacement |
|---|---|---|
| Container registry | AWS ECR / Docker Hub | Local registry (Harbor or plain registry) loaded from bundle |
| TLS certificates | Let's Encrypt (ACME) | Customer-provided certificates or internal CA |
| License validation | chaosplane.dev/license API | Offline license file (signed JWT, validated locally) |
| AI Assistant LLM | OpenAI / Anthropic API | Optional: customer-provided local LLM endpoint, or AI features disabled |
| Email notifications | SendGrid | Customer-provided SMTP relay, or notifications disabled |
| Telemetry / analytics | chaosplane.dev telemetry | Disabled entirely |
| Software updates | Pull from registry | Manual bundle update process |
| Time sync | AWS Time Sync / NTP | Customer-provided NTP server |

### Air-Gap Installation Process

1. Download the air-gap bundle on an internet-connected machine (or receive via approved transfer media)
2. Verify GPG signature and SHA-256 checksums
3. Transfer bundle to the air-gapped environment via approved media (USB, CD, secure file transfer)
4. Run `preflight-check.sh` to verify Kubernetes cluster prerequisites
5. Run `load-images.sh` to load all container images into the local registry
6. Configure `values-airgap.yaml` with environment-specific settings (domain, TLS certs, SMTP, license file)
7. Run `install.sh` to deploy via Helm
8. Verify installation with built-in health checks

### Air-Gap Upgrade Process

Upgrades follow the same process as installation. The upgrade bundle includes only changed images to minimize transfer size. The upgrade script performs a rolling update with automatic rollback on failure.

### Air-Gap Constraints

- AI Assistant features require a customer-provided local LLM endpoint. If no endpoint is configured, AI features are disabled gracefully with a clear UI message.
- Software updates must be applied manually. ChaosPlane provides security advisories via a separate channel (email, customer portal) for air-gapped customers.
- License files are valid for 12 months and must be renewed via the customer portal on an internet-connected machine.

---

## CAC/PIV Authentication

### Overview

Common Access Card (CAC) and Personal Identity Verification (PIV) cards are US government smart cards used for authentication. ChaosPlane Government Edition supports CAC/PIV authentication via SAML SSO integration with the agency's identity provider.

ChaosPlane does not directly read CAC/PIV cards. Instead, the agency's IdP (typically Active Directory Federation Services with smart card authentication, or a PIV-enabled identity provider) handles the card authentication and issues a SAML assertion to ChaosPlane.

### Supported Integration Patterns

| Pattern | Description | Configuration |
|---|---|---|
| ADFS + Smart Card | Agency uses ADFS with smart card authentication. ADFS issues SAML 2.0 assertion to ChaosPlane. | Standard SAML SSO configuration with ADFS as IdP |
| Okta + PIV | Agency uses Okta with PIV factor. Okta issues SAML 2.0 assertion. | Standard SAML SSO configuration with Okta as IdP |
| Azure AD + CAC | Agency uses Azure AD with certificate-based authentication (CBA). Azure AD issues SAML 2.0 assertion. | Standard SAML SSO configuration with Azure AD as IdP |
| Ping Identity + PIV | Agency uses Ping Identity with PIV. Ping issues SAML 2.0 assertion. | Standard SAML SSO configuration with Ping as IdP |

### SAML Assertion Requirements

ChaosPlane requires the following attributes in the SAML assertion for CAC/PIV users:

| Attribute | Required | Description |
|---|---|---|
| `NameID` | Yes | User's unique identifier (email or EDIPI) |
| `email` | Yes | User's email address |
| `displayName` | Recommended | User's display name |
| `groups` | Recommended | Group memberships for RBAC role mapping |
| `assuranceLevel` | Recommended | Authentication assurance level (e.g., LOA3 for PIV) |

### Authentication Assurance Level Enforcement

ChaosPlane Government Edition can be configured to require a minimum authentication assurance level. When configured:

- Users authenticating with username/password only are denied access
- Users authenticating with PIV/CAC (LOA3) are granted access
- The assurance level is extracted from the SAML assertion and stored in the session
- CUI-marked resources can be configured to require LOA3 assurance level

### CAC/PIV Configuration

CAC/PIV authentication is configured in the ChaosPlane admin console under Settings > Authentication > SAML SSO. The configuration includes:

- IdP metadata URL or XML upload
- Attribute mapping (NameID format, email attribute, groups attribute)
- Minimum assurance level requirement (optional)
- Just-in-time (JIT) provisioning settings
- Group-to-role mapping for automatic RBAC assignment

### Session Behavior with CAC/PIV

- Sessions established via CAC/PIV are tagged with the assurance level
- Session duration follows the agency IdP's session policy (ChaosPlane respects the SAML `SessionNotOnOrAfter`)
- Re-authentication is required when the session expires — users are redirected to the agency IdP
- WebSocket sessions re-authenticated every 5 minutes via server-initiated token revalidation

---

## DISA STIG Compliance

### Overview

Defense Information Systems Agency (DISA) Security Technical Implementation Guides (STIGs) define security configuration requirements for DoD systems. ChaosPlane Government Edition is prepared to comply with the following STIGs:

| STIG | Version | Applicability |
|---|---|---|
| Kubernetes STIG | V1R11 | ChaosPlane EKS deployment |
| Docker Enterprise STIG | V2R2 | Container runtime |
| Red Hat Enterprise Linux 8 STIG | V1R13 | Node OS (where applicable) |
| PostgreSQL 9.x STIG | V2R4 | Database (mapped to PostgreSQL 15) |
| Web Server STIG (Apache/Nginx) | V2R10 | Ingress controller |
| Application Security and Development STIG | V5R3 | ChaosPlane application |

### Kubernetes STIG Implementation

Key Kubernetes STIG controls implemented in ChaosPlane Government Edition:

| STIG ID | Requirement | ChaosPlane Implementation |
|---|---|---|
| V-242376 | Kubernetes API server must use TLS | API server TLS enforced. Client certificate authentication for internal components. |
| V-242377 | Kubernetes etcd must use TLS | etcd TLS enforced. Client certificate authentication required. |
| V-242381 | Kubernetes must enable RBAC | RBAC enabled on all EKS clusters. No ClusterAdmin bindings for application service accounts. |
| V-242383 | Kubernetes must restrict anonymous access | Anonymous authentication disabled on API server. |
| V-242386 | Kubernetes must use approved CNI plugin | Approved CNI plugin configured. Network policies enforced. |
| V-242390 | Kubernetes must have audit logging enabled | Kubernetes audit logging enabled. Logs shipped to CloudWatch (GovCloud). |
| V-242395 | Kubernetes must limit privilege escalation | `allowPrivilegeEscalation: false` set on all ChaosPlane containers. |
| V-242397 | Kubernetes must not run containers as root | All ChaosPlane containers run as non-root user (UID 1000+). |
| V-242400 | Kubernetes must use read-only root filesystem | Read-only root filesystem enforced on all ChaosPlane containers. Writable volumes mounted explicitly where needed. |
| V-242404 | Kubernetes must limit container capabilities | All unnecessary Linux capabilities dropped. Only required capabilities (e.g., NET_BIND_SERVICE where needed) explicitly added. |
| V-242406 | Kubernetes must use Pod Security Standards | Pod Security Standards enforced at namespace level (restricted profile for ChaosPlane namespaces). |

### Application STIG Implementation

Key Application Security and Development STIG controls:

| STIG ID | Requirement | ChaosPlane Implementation |
|---|---|---|
| V-222396 | Application must enforce approved authorizations | RBAC/ABAC enforced on all resources. Default-deny. |
| V-222402 | Application must use multifactor authentication | MFA enforced for all Government Edition users. CAC/PIV via agency IdP. |
| V-222404 | Application must implement DoD-approved encryption | FIPS 140-2 validated cryptography (BoringCrypto). TLS 1.2+ with FIPS cipher suites. |
| V-222418 | Application must produce audit records | Append-only audit logs with actor, timestamp, IP, resource, outcome. |
| V-222430 | Application must protect audit information | Audit logs append-only. Tamper-evident signing. Access restricted to authorized roles. |
| V-222550 | Application must not expose sensitive information in error messages | Error responses do not expose internal details. Stack traces not returned to clients. |
| V-222562 | Application must implement CSRF protections | CSRF tokens on all state-changing endpoints. SameSite=Strict cookie attribute. |
| V-222596 | Application must use approved TLS versions | TLS 1.2+ enforced. TLS 1.0 and 1.1 disabled. |

### STIG Compliance Scanning

ChaosPlane Government Edition includes STIG compliance scanning as part of the CI/CD pipeline:

- OpenSCAP scans run against container images before deployment
- Kubernetes STIG checks run via kube-bench on every cluster
- STIG findings categorized as CAT I (critical), CAT II (high), CAT III (medium)
- CAT I findings block deployment. CAT II findings require documented remediation plan within 30 days.
- STIG scan reports available to agency customers via the admin console

### STIG Deviation Process

Where a STIG requirement cannot be fully implemented due to technical constraints or operational necessity, ChaosPlane documents a formal deviation:

- Deviation ID and STIG control reference
- Technical justification for the deviation
- Compensating controls in place
- Risk acceptance by CISO
- Deviation reviewed annually

---

## CUI Marking System

### Overview

The CUI marking system in ChaosPlane Government Edition allows customers to identify, mark, and control access to resources that contain Controlled Unclassified Information. The system follows the National Archives CUI Registry and NARA 32 CFR Part 2002.

### CUI Categories Supported

| Category | Abbreviation | Description |
|---|---|---|
| Critical Infrastructure | CRIT | Information about critical infrastructure systems |
| Defense | DEF | Defense-related information |
| Intelligence | INTEL | Intelligence-related information |
| Law Enforcement | LAW | Law enforcement sensitive information |
| Privacy | PRVCY | Personally identifiable information under federal privacy law |
| Procurement and Acquisition | PROPIN | Procurement-sensitive information |
| Export Control | EXPT | Export-controlled technical data |

### Marking Application

CUI markings are applied at the resource level:

- **Experiment definitions**: Marked as CUI when the experiment targets CUI-bearing systems or includes CUI in its description or parameters
- **Topology data**: Marked as CUI when imported topology data contains CUI (e.g., system names, IP addresses of CUI systems)
- **Audit log exports**: Automatically marked as CUI if any CUI-marked resources are included in the export
- **Resilience reports**: Marked as CUI if the report covers CUI-marked experiments

Marking format follows the CUI standard: `CUI//[CATEGORY]` or `CUI//[CATEGORY]//[DISSEMINATION CONTROL]`

Examples:
- `CUI//DEF` — Defense CUI
- `CUI//CRIT//NOFORN` — Critical Infrastructure CUI, not releasable to foreign nationals

### CUI Banner Display

When a user accesses a CUI-marked resource, a banner is displayed at the top of the page:

```
┌─────────────────────────────────────────────────────────────────┐
│  CUI // CONTROLLED UNCLASSIFIED INFORMATION                     │
│  Category: DEFENSE  |  Handle per CUI policy                   │
└─────────────────────────────────────────────────────────────────┘
```

The banner persists for the duration of the session in which CUI-marked resources are accessed.

### CUI Access Controls

- CUI-marked resources are accessible only to users with the `cui-access` attribute in the ABAC policy
- `cui-access` attribute is granted by the organization administrator
- CUI access is logged with the `CUI_ACCESS` event type in the audit log
- CUI access can be further restricted by CUI category (e.g., `cui-access:def` for Defense CUI only)
- CUI-marked resources are excluded from standard data exports and API responses unless the requesting user has `cui-access`

### CUI Export Controls

Exporting CUI-marked data requires:

1. User must have `cui-access` attribute
2. Export request must include a justification (free text, logged)
3. Export is flagged as CUI in the export metadata
4. Export notification sent to the organization's CUI officer (configurable)
5. Export logged with `CUI_EXPORT` event type, including justification and recipient

### CUI Destruction

When a CUI-marked resource is deleted:

1. Deletion logged with `CUI_DESTRUCTION` event type
2. Standard deletion pipeline applies (30-day grace period for account deletion, immediate for resource deletion)
3. CUI destruction notification sent to the organization's CUI officer
4. Destruction record retained for 3 years (per CUI destruction requirements)

---

## Government Plan Tiers

| Feature | Enterprise | Government |
|---|---|---|
| All Enterprise features | ✅ | ✅ |
| AWS GovCloud deployment | — | ✅ |
| Air-gap deployment bundle | — | ✅ |
| CAC/PIV authentication (via SAML) | — | ✅ |
| DISA STIG compliance scanning | — | ✅ |
| CUI marking system | — | ✅ |
| FIPS 140-2 cryptography | — | ✅ |
| FedRAMP Moderate (in preparation) | — | ✅ |
| CMMC Level 2 support | — | ✅ |
| CSAP 표준등급 (Korea region) | — | ✅ |
| Dedicated GovCloud CI/CD | — | ✅ |
| Session recording for privileged access | — | ✅ |
| Tamper-evident audit log signing | — | ✅ |
| Offline license validation | — | ✅ |
| Dedicated government support SLA | — | ✅ |
| On-premises deployment option | — | ✅ (air-gap) |

---

## Government Plan Deployment Options

### Option 1: AWS GovCloud SaaS

ChaosPlane operates the platform in AWS GovCloud (us-gov-west-1). The customer accesses it as a SaaS service. This is the recommended option for agencies that can use cloud services.

- Fastest time to value
- ChaosPlane handles all infrastructure operations
- FedRAMP Moderate authorization covers this deployment
- Suitable for IL2/IL4 workloads

### Option 2: Customer-Managed GovCloud

ChaosPlane provides the Helm charts and deployment tooling. The customer deploys and operates ChaosPlane in their own AWS GovCloud account. ChaosPlane provides support and software updates.

- Customer retains full control of infrastructure
- Suitable for agencies with strict data sovereignty requirements
- Customer is responsible for FedRAMP boundary (inherits ChaosPlane's control documentation)

### Option 3: Air-Gap On-Premises

ChaosPlane provides the air-gap bundle. The customer deploys on their own on-premises Kubernetes cluster with no internet connectivity.

- No external dependencies
- Suitable for classified-adjacent environments, SCIFs, and air-gapped networks
- Customer is responsible for all infrastructure operations
- Software updates applied manually via new bundle releases

### Option 4: Korea Public Sector (CSAP)

ChaosPlane operates the platform on NHN Cloud or Naver Cloud Platform in the Korea region. Designed for Korean public institutions (공공기관) requiring CSAP 표준등급.

- All data remains in Korea
- CSAP 표준등급 certification in preparation
- Korean-language support
- PIPA compliance

---

## Support and SLA

| Metric | Government Plan |
|---|---|
| Uptime SLA | 99.9% |
| Support hours | 24/7 for P0/P1 |
| P0 response time | 15 minutes |
| P1 response time | 1 hour |
| P2 response time | 4 hours |
| Dedicated CSM | ✅ |
| Security incident notification | Within 1 hour |
| Patch notification (critical CVE) | Within 24 hours |
| Air-gap security advisory | Via secure email channel |

---

## Roadmap

| Feature | Target |
|---|---|
| Government Plan GA | Month 20 |
| FedRAMP Moderate ATO | Phase 4 (Month 22-24) |
| CMMC Level 2 certification | Phase 4 (Month 23-24) |
| CSAP 표준등급 | Phase 4 (Month 19-20) |
| DoD IL5 support | Phase 5 |
| CSAP 상등급 | Phase 5 |
| AWS Marketplace GovCloud listing | Phase 5 |

---

## Document Control

| Field | Value |
|---|---|
| Version | 1.0.0 |
| Status | Draft — Government Plan in development |
| Owner | CISO / Product |
| Last reviewed | April 2026 |
| Next review | October 2026 |
