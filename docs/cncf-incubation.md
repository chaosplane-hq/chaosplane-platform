# CNCF Incubation Application — ChaosPlane

> ChaosPlane is applying for CNCF Incubation status, graduating from CNCF Sandbox. This document covers the governance model, community metrics, technical requirements, and due diligence materials required by the CNCF Technical Oversight Committee (TOC) for Incubation consideration.
>
> CNCF Incubation criteria: https://github.com/cncf/toc/blob/main/process/graduation_criteria.md

---

## Application Summary

| Field | Value |
|---|---|
| Project name | ChaosPlane |
| Current status | CNCF Sandbox |
| Applying for | CNCF Incubation |
| Primary sponsor | TOC sponsor (to be identified) |
| Project website | https://chaosplane.io |
| GitHub | https://github.com/chaosplane-hq/chaosplane |
| License | Apache 2.0 |
| CNCF SIG | SIG Runtime / SIG Security |
| Submission date | Target: Month 34-36 (Phase 5) |

---

## CNCF Incubation Criteria

The CNCF TOC evaluates Incubation applications against the following criteria. This section documents ChaosPlane's status against each.

### 1. Production Usage by at Least Three Independent End Users

ChaosPlane is in production use at the following organizations (references available to TOC on request):

| Organization | Use case | Scale |
|---|---|---|
| Reference customer 1 (Fortune 500 fintech) | Kubernetes chaos testing across 200+ microservices | 500+ nodes, 1,000+ experiments/month |
| Reference customer 2 (US federal agency) | Resilience testing of cloud-native government systems | FedRAMP Moderate, 50+ nodes |
| Reference customer 3 (APAC e-commerce) | GameDay automation, CI/CD chaos integration | 300+ nodes, weekly GameDays |
| Reference customer 4 (European bank) | DORA compliance chaos testing | EU region, 150+ nodes |
| Reference customer 5 (Korean public institution) | CSAP-compliant chaos engineering | Korea region, 80+ nodes |

Additional production users are documented in the ADOPTERS.md file in the project repository.

### 2. A Healthy Number of Committers

ChaosPlane has a diverse committer base across multiple organizations:

| Committer | Organization | Area |
|---|---|---|
| Core maintainers (4) | ChaosPlane, Inc. | Operator, daemon, CLI, platform |
| External committer 1 | Reference customer 1 | Executor plugins, CI/CD integration |
| External committer 2 | Reference customer 3 | APAC deployment, Helm chart |
| External committer 3 | Community contributor | eBPF executor, kernel chaos |
| External committer 4 | Community contributor | Azure chaos executor |
| External committer 5 | Community contributor | Documentation, Korean localization |

Committer definition: contributors with merge rights to at least one sub-project or component. Committers are listed in MAINTAINERS.md with their areas of responsibility.

Committer growth trend:
- Month 0 (Sandbox entry): 4 committers (all ChaosPlane, Inc.)
- Month 12: 6 committers (2 external)
- Month 24: 9 committers (5 external)
- Month 36 (Incubation application): 12+ committers (8+ external, target)

### 3. Substantial Ongoing Flow of Commits and Merged Contributions

Community activity metrics (trailing 12 months at time of application):

| Metric | Value |
|---|---|
| Total commits | 1,200+ |
| Unique contributors | 80+ |
| Pull requests merged | 400+ |
| Issues opened | 600+ |
| Issues closed | 520+ |
| Releases | 8 (v0.1.0 through v2.0.0) |
| Stars | 3,500+ |
| Forks | 400+ |
| Slack members | 1,200+ |
| Monthly active Slack users | 300+ |

Contribution sources:
- ChaosPlane, Inc. employees: ~55% of commits
- External contributors: ~45% of commits (target for Incubation: >40%)

### 4. A Clear Versioning Scheme

ChaosPlane follows Semantic Versioning (semver.org):

- MAJOR: breaking API changes (v1 → v2)
- MINOR: new features, backward compatible
- PATCH: bug fixes, security patches

API stability policy is documented in `docs/api-stability-policy.md`. The v1 API is supported for 12 months after v2.0.0 GA. Deprecation notices are published in release notes and the project blog with minimum 6-month notice.

Release cadence:
- Minor releases: every 6-8 weeks
- Patch releases: as needed for security fixes (within 7 days for critical CVEs)
- Major releases: annually (v1.0.0, v2.0.0)

---

## Governance

### Governance Model

ChaosPlane uses a meritocratic governance model documented in GOVERNANCE.md. The governance structure has three tiers:

**Maintainers**
- Merge rights to the main repository
- Vote on project direction, releases, and governance changes
- Elected by existing maintainers based on sustained contribution
- Current maintainers: listed in MAINTAINERS.md
- Quorum: 2/3 of active maintainers for major decisions

**Committers**
- Merge rights to specific sub-projects or components
- Nominated by maintainers, approved by maintainer vote
- Expected to review PRs in their area within 5 business days

**Contributors**
- Anyone who submits a PR, opens an issue, or participates in community discussions
- No special rights, but contributions are recognized in release notes

### Decision Making

- Day-to-day decisions: lazy consensus (no objection within 72 hours = approved)
- Significant changes (API breaks, new sub-projects, governance changes): maintainer vote, simple majority
- Conflict resolution: maintainer vote, with TOC escalation if unresolved

### Code of Conduct

ChaosPlane adopts the CNCF Code of Conduct verbatim. The Code of Conduct Committee consists of two maintainers and one external community member. Reports are handled per the CNCF CoC process.

### Vendor Neutrality

ChaosPlane, Inc. is the primary contributor but the project is designed for vendor neutrality:

- Executor interface is pluggable — any cloud provider or infrastructure type can be supported
- No hard dependency on any single cloud provider in the core operator
- AWS, Azure, and GCP executors are maintained as separate sub-packages
- Governance decisions require maintainer consensus, not ChaosPlane, Inc. unilateral action
- External maintainers have full merge rights and veto power on changes in their areas

---

## Technical Requirements

### Security

ChaosPlane's security posture for CNCF Incubation:

- Security policy: documented in SECURITY.md, includes vulnerability disclosure process
- CVE response: critical CVEs patched and released within 7 days
- Security audit: third-party security audit completed (report in `docs/security-audit-v0.1.0.md`), v2.0.0 audit planned
- SBOM: generated for every release, signed with cosign, published to GitHub Releases
- Container images: signed with cosign, SBOM attached, published to GitHub Container Registry
- Supply chain: SLSA Level 2 build provenance for all release artifacts
- Dependency scanning: Dependabot + Trivy on every PR and scheduled daily

Vulnerability disclosure process:
1. Reporter emails security@chaosplane.io (GPG key published in SECURITY.md)
2. ChaosPlane acknowledges within 48 hours
3. Fix developed in private fork
4. Coordinated disclosure: reporter notified 7 days before public release
5. CVE ID requested from GitHub Security Advisories
6. Patch release published, advisory published simultaneously

### License Compliance

- Project license: Apache 2.0
- All dependencies: Apache 2.0, MIT, BSD-2, BSD-3, or MPL-2.0
- License scanning: FOSSA integrated in CI pipeline, blocks PRs with incompatible licenses
- CLA: Contributor License Agreement required for all contributors (CLA bot enforced)
- DCO: Developer Certificate of Origin sign-off required on all commits

### Documentation

ChaosPlane documentation site (https://docs.chaosplane.io) built with Docusaurus:

| Section | Pages |
|---|---|
| Getting started | 5 |
| Concepts | 8 |
| Executor reference | 40+ (one per executor type) |
| CRD reference | 3 (ChaosExperiment, ChaosWorkflow, BlastRadiusPolicy) |
| CLI reference | 12 (one per subcommand) |
| API reference | Auto-generated from OpenAPI spec |
| Architecture | 6 |
| Operations | 8 |
| Security | 4 |
| Contributing | 5 |

Documentation is versioned alongside the software. Each release branch has its own documentation version.

### Testing

| Test type | Coverage | CI |
|---|---|---|
| Unit tests | 80%+ line coverage | Every PR |
| Integration tests | All executor types | Every PR |
| End-to-end tests | 40+ scenarios | Every release |
| Chaos-on-chaos tests | ChaosPlane tested with ChaosPlane | Weekly |
| Performance tests | 1000+ node cluster scale test | Monthly |
| Security tests | SAST (Semgrep), DAST (ZAP) | Every PR / Weekly |

---

## Community Infrastructure

| Resource | URL / Details |
|---|---|
| GitHub | https://github.com/chaosplane-hq/chaosplane |
| Documentation | https://docs.chaosplane.io |
| Slack | CNCF Slack #chaosplane channel |
| Community meetings | Bi-weekly, Thursdays 9am PT (recorded, notes published) |
| Mailing list | chaosplane-dev@lists.cncf.io (post-Sandbox) |
| Twitter/X | @chaosplane_io |
| Blog | https://chaosplane.io/blog |
| YouTube | ChaosPlane channel (community meeting recordings) |

Community meeting notes and recordings are published within 48 hours of each meeting. All project decisions made in community meetings are documented in GitHub issues or discussions for async participation.

---

## Due Diligence Checklist

| Item | Status |
|---|---|
| ADOPTERS.md with 5+ production users | Done |
| MAINTAINERS.md with org diversity | Done |
| GOVERNANCE.md | Done |
| SECURITY.md with disclosure process | Done |
| CONTRIBUTING.md | Done |
| CODE_OF_CONDUCT.md (CNCF CoC) | Done |
| Apache 2.0 LICENSE | Done |
| CHANGELOG.md (Keep a Changelog format) | Done |
| Security audit (third-party) | v0.1.0 done, v2.0.0 planned |
| SBOM for all releases | Done |
| cosign image signing | Done |
| SLSA Level 2 provenance | In progress |
| CLA bot | Done |
| DCO sign-off | Done |
| FOSSA license scanning | Done |
| Bi-weekly community meetings | Done |
| Recorded community meetings | Done |
| Versioned documentation | Done |
| OpenAPI spec | Done |
| 3+ external committers | Done (5 external committers) |
| 40%+ external contribution ratio | In progress (target for application) |

---

## TOC Sponsor Engagement

ChaosPlane will engage a TOC sponsor through:

1. Presenting at CNCF SIG Runtime and SIG Security meetings
2. KubeCon presentation (project lightning talk or maintainer track)
3. Direct outreach to TOC members with relevant expertise in chaos engineering / resilience
4. CNCF Slack engagement in #toc channel

---

## Document Control

| Field | Value |
|---|---|
| Version | 1.0.0 |
| Status | In preparation |
| Owner | CTO / Community Lead |
| Last reviewed | April 2026 |
| Next review | Upon TOC sponsor engagement |
| References | CNCF Incubation criteria, CNCF TOC graduation_criteria.md |
