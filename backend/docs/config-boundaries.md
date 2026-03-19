# Configuration System Boundaries

Use this table to decide where a new configuration belongs.

## Quick Rule

- If the value is returned to the mini-program client, use `mini_program_config`.
- If the value is a backend credential, integration setting, or secret, use `system_setting`.
- Never store secrets in `mini_program_config`.

## Decision Table

| Question | If yes, use | Why |
|---|---|---|
| Will this value be returned to the mini-program client? | `mini_program_config` | It is part of client-visible bootstrap/config data. |
| Is this a secret, credential, token, or private integration parameter? | `system_setting` | Supports encrypted storage for secrets. |
| Is this used by backend runtime integrations rather than client UI behavior? | `system_setting` | Backend-owned operational config belongs here. |
| Does this need scope-based overrides like global/station/agent? | `mini_program_config` | It already supports `namespace + scopeType + scopeCode`. |
| Does this control mini-program UI, feature flags, labels, or display behavior? | `mini_program_config` | It is public product configuration. |
| Does this need audit-style revision history for backend settings changes? | `system_setting` | It has revision records. |
| Does this need hot reload into backend runtime without restart? | `system_setting` | Current runtime refresh flow is built around it. |

## System Comparison

### `mini_program_config`

- Purpose: Public mini-program bootstrap and client behavior config
- Audience: Mini-program client and admin editors
- Access:
  - Public read: `/api/miniapp/config/bootstrap`
  - Public read: `/api/v1/miniapp/config/bootstrap`
  - Admin manage: `/api/admin/mini-program/configs`
  - Admin manage: `/api/v1/admin/mini-program/configs`
- Shape: `namespace + configKey + scopeType + scopeCode`
- Security: No encryption; treat as non-secret
- Best for:
  - Feature flags exposed to client
  - UI text or labels
  - Recharge switch, support channels, app display settings
- Examples:
  - `general.app_name`
  - `feature.recharge_enabled`
  - `general.support_wechat`

### `system_setting`

- Purpose: Backend runtime settings and integration credentials
- Audience: Backend services and admin operators
- Access:
  - Admin read/write: `/api/admin/system-settings/wechat`
  - Admin read/write: `/api/v1/admin/system-settings/wechat`
- Shape: Flat category/key records such as `wechat.app_id`
- Security: Supports secret storage and masking
- Runtime: Used for backend hot-reload/runtime refresh
- Best for:
  - Third-party credentials
  - Backend integration toggles/settings
  - Private operational parameters
- Examples:
  - `wechat.app_id`
  - `wechat.app_secret`

## Common Examples

| Example | Use |
|---|---|
| Mini-program app name shown in UI | `mini_program_config` |
| Whether recharge page is visible | `mini_program_config` |
| Customer support WeChat ID shown to users | `mini_program_config` |
| WeChat API AppID used by backend login flow | `system_setting` |
| WeChat API AppSecret | `system_setting` |
| Any token, key, or credential | `system_setting` |

## If You Picked Wrong

- Secret found in `mini_program_config`: move it to `system_setting` immediately and rotate the secret if it may have been exposed.
- Backend-only value stored in `mini_program_config`: migrate it to `system_setting` and keep only any safe public derivative in `mini_program_config`.
- Client-visible value stored in `system_setting`: move it to `mini_program_config` if the mini-program needs to read it directly.

## Non-Goals

- `mini_program_config` is not a secret store.
- `system_setting` is not a general-purpose client bootstrap catalog.

## Current Status: Client Bootstrap

**Status**: Backend-ready, client integration not enabled yet

The `mini_program_config` system and `/api/miniapp/config/bootstrap` endpoint are fully implemented on the backend, but the mini-program client does not currently fetch this configuration.

**Why not enabled**:
- No real runtime consumers yet (app_name, support_wechat not wired to UI)
- Static config in `miniprogram/config/index.js` is sufficient for current needs
- API_BASE_URL must remain local (needed before bootstrap can be called)

**When to enable**:
- When a mini-program feature needs dynamic runtime configuration
- First candidate: `feature.recharge_enabled` to show/hide recharge entry
- Second candidate: `general.support_wechat` when support contact UI is added

**How to enable** (future):
1. Fetch asynchronously after app launch (not blocking)
2. Keep local defaults in code
3. Load cached config from storage on startup
4. Background refresh from `/api/v1/miniapp/config/bootstrap`
5. Merge server values over defaults

**What should NOT move to bootstrap**:
- `API_BASE_URL` (needed before bootstrap call)
- `SITES` (deployment-specific)
- Enum constants and color values (static code assets)
