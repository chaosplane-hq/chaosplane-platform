# ChaosPlane v1 API Stability Policy

> This policy defines how ChaosPlane manages backward compatibility, versioning, deprecation, and breaking changes for the v1 API. It applies to all consumers of the ChaosPlane REST API, WebSocket API, and the Go, TypeScript, and Python SDKs.

---

## Guiding Principle

Once an API is marked stable, you can trust it. We won't break your integration without warning, a migration path, and enough time to adapt.

---

## Versioning

ChaosPlane follows [Semantic Versioning 2.0.0](https://semver.org) for the platform and SDKs.

```
MAJOR.MINOR.PATCH
```

| Component | Meaning |
|---|---|
| MAJOR | Breaking change. Incremented when backward compatibility is broken. |
| MINOR | New functionality added in a backward-compatible way. |
| PATCH | Backward-compatible bug fixes and security patches. |

### API Version vs. Platform Version

The API version (`v1`, `v2`, ...) is separate from the platform release version (`1.0.0`, `1.1.0`, ...).

- `v1` API is introduced with ChaosPlane 1.0.0 GA.
- The `v1` API remains stable for the entire v1 lifecycle, regardless of platform minor/patch releases.
- A new API major version (`v2`) will only be introduced when breaking changes are unavoidable and cannot be handled via additive changes.

### URL Structure

```
https://api.chaosplane.io/v1/{resource}
```

All v1 endpoints are prefixed with `/v1/`. When v2 is introduced, both versions will be served simultaneously during the overlap period.

---

## What "Stable" Means

A `v1` endpoint is stable once it ships in a GA release. Stable means:

- The endpoint path will not change.
- Required request fields will not be removed or renamed.
- Response fields will not be removed or renamed.
- Response field types will not change (e.g., a string will not become an integer).
- HTTP status codes for documented scenarios will not change.
- Enum values that are documented will not be removed.
- Authentication and authorization behavior will not change in a breaking way.

---

## What Is NOT a Breaking Change

The following changes are considered backward-compatible and may be made at any time without a deprecation notice:

- Adding new optional request fields
- Adding new response fields (clients must ignore unknown fields)
- Adding new enum values to existing fields (clients must handle unknown enum values gracefully)
- Adding new endpoints
- Adding new HTTP methods to existing resources
- Changing error message text (not error codes)
- Improving performance or reducing latency
- Adding new optional query parameters
- Expanding rate limits
- Bug fixes that make behavior match documented behavior

**SDK consumers:** always code defensively. Deserialize responses into structs that tolerate unknown fields. Do not rely on the exact order of fields in JSON responses.

---

## What IS a Breaking Change

The following changes are breaking and require a deprecation period before removal:

- Removing an endpoint
- Removing a required or optional request field
- Removing a response field
- Renaming an endpoint, field, or parameter
- Changing the type of a field (e.g., string to integer, object to array)
- Changing the meaning of a field value
- Removing a documented enum value
- Changing authentication requirements in a way that breaks existing integrations
- Reducing rate limits below current documented values
- Changing pagination behavior in a way that breaks existing cursor-based iteration

---

## Deprecation Policy

### Notice Period

| Change type | Minimum deprecation notice |
|---|---|
| Endpoint removal | 12 months |
| Field removal or rename | 12 months |
| Enum value removal | 12 months |
| API major version sunset | 24 months after successor GA |

These are minimums. We will extend notice periods when the impact is broad or migration is complex.

### How Deprecation Is Communicated

When something is deprecated:

1. The API reference documentation is updated with a `Deprecated` label and the planned removal date.
2. A deprecation notice is published in the changelog.
3. Deprecated endpoints return a `Deprecation` response header:
   ```
   Deprecation: true
   Sunset: Sat, 01 Apr 2028 00:00:00 GMT
   Link: <https://docs.chaosplane.io/migration/v1-to-v2>; rel="successor-version"
   ```
4. Enterprise customers receive direct notification via their Customer Success Manager.
5. SDK releases mark deprecated methods with language-native deprecation annotations (`@deprecated` in TypeScript, `// Deprecated:` in Go, `warnings.warn` in Python).

### Deprecation Does Not Mean Removal

A deprecated endpoint continues to work until the sunset date. Deprecation is a signal to migrate, not an immediate break.

---

## Breaking Change Process

When a breaking change is unavoidable:

1. A new API version (`v2`) is introduced alongside `v1`. Both are served simultaneously.
2. The breaking change ships only in `v2`. `v1` is unchanged.
3. `v1` enters a deprecation period of at least 24 months from the `v2` GA date.
4. Migration guide published at `docs.chaosplane.io/migration/v1-to-v2`.
5. SDKs ship a new major version with the breaking change. The previous major version receives security patches only during the overlap period.
6. `v1` is sunset on the announced date. After sunset, `v1` endpoints return `410 Gone`.

We will not introduce breaking changes within a stable API version. If a security vulnerability requires an immediate breaking fix, we will notify affected customers directly and provide the shortest feasible migration window.

---

## Preview and Beta Endpoints

Some endpoints ship as `preview` or `beta` before reaching stable status.

| Stability label | Meaning |
|---|---|
| `preview` | Early access. May change or be removed without notice. Not recommended for production. |
| `beta` | Functionally complete. Breaking changes possible but will be communicated with at least 30 days notice. |
| `stable` | Full stability guarantees as described in this policy. |

Preview and beta endpoints are marked in the API reference and return a header:

```
X-ChaosPlane-Stability: preview
```

or

```
X-ChaosPlane-Stability: beta
```

Stable endpoints do not return this header.

---

## SDK Versioning

The Go, TypeScript, and Python SDKs follow the same semantic versioning rules.

- SDK major version aligns with API major version (SDK v1.x.x wraps API v1, SDK v2.x.x wraps API v2).
- SDK minor versions may add new convenience methods, helpers, or support for new optional API fields.
- SDK patch versions contain bug fixes and security patches.
- Deprecated SDK methods are annotated and documented. They are removed only in the next major SDK version.

### SDK Support Window

| SDK version | Support status |
|---|---|
| Latest major | Active: new features, bug fixes, security patches |
| Previous major | Maintenance: security patches only, for 24 months after successor GA |
| Older | End of life: no updates |

---

## Changelog and Release Notes

Every release includes a changelog entry categorized as:

- `added` — new endpoints, fields, or SDK methods
- `changed` — non-breaking behavior changes
- `deprecated` — items entering the deprecation period
- `removed` — items removed after deprecation period
- `fixed` — bug fixes
- `security` — security patches

The changelog is published at `docs.chaosplane.io/changelog` and as GitHub releases.

---

## Rate Limits

Rate limits are documented per endpoint in the API reference. We will not reduce rate limits below documented values without a 90-day notice period. Rate limit increases are non-breaking and may happen at any time.

Current default rate limits:

| Tier | API requests | Experiment executions |
|---|---|---|
| Free | 100 req/min | 10/month |
| Team | 1,000 req/min | Unlimited |
| Business | 2,000 req/min | Unlimited |
| Enterprise | Custom (negotiated) | Unlimited |
| Government | Custom (negotiated) | Unlimited |

---

## Pagination

All list endpoints use cursor-based pagination. The cursor format is opaque — do not parse or construct cursors manually. Cursors are valid for 24 hours.

```json
{
  "data": [...],
  "pagination": {
    "next_cursor": "eyJpZCI6MTIzfQ==",
    "has_more": true
  }
}
```

The cursor format may change between minor versions, but existing cursors will remain valid for their 24-hour window. New cursors issued after a format change will use the new format.

---

## Error Responses

Error responses follow a stable structure:

```json
{
  "error": {
    "code": "experiment_not_found",
    "message": "Experiment with ID exp_abc123 was not found.",
    "request_id": "req_xyz789"
  }
}
```

- `error.code` is a stable machine-readable string. Do not parse `error.message` programmatically.
- New error codes may be added at any time (non-breaking).
- Existing error codes will not be removed or renamed without a deprecation period.
- HTTP status codes for documented error scenarios are stable.

---

## Feedback and Exceptions

If a change we've classified as non-breaking causes a real problem for your integration, reach out. We take compatibility seriously and will work with you on a solution.

Enterprise customers can raise API stability concerns directly with their Customer Success Manager. All customers can open an issue at github.com/chaosplane/chaosplane or email api-feedback@chaosplane.io.

---

## Document Control

| Field | Value |
|---|---|
| Version | 1.0.0 |
| Status | Active |
| Owner | Platform Engineering |
| Effective date | ChaosPlane v1.0.0 GA |
| Last reviewed | April 2026 |
| Next review | April 2027 |
