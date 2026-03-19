# Merge Report: `v0.1.103` into `zyp-dev`

## 1. Merge Context

- Merge date: 2026-03-19 (UTC)
- Target branch: `zyp-dev`
- Source tag: `v0.1.103` (`3cedfcd827809cb9d76196ee67dcc3480b5100b7`)
- Target HEAD before merge: `6facef2858e477e9017e24c2080ee792ea2f88ed`
- Merge-base: `7be5e1734c1c0882c851f9fd1c2ee70cc62b7d29`
- Divergence (`git rev-list --left-right --count HEAD...v0.1.103`):
  - Current branch unique commits: `287`
  - Tag unique commits: `649`

## 2. Safety and Resolution Strategy

- Backup branch created before merge: `backup/pre-merge-v0.1.103-20260319`
- Merge command: `git merge --no-commit --no-ff v0.1.103`
- Conflict handling principle: **do not overwrite current branch changes**.
- Actual conflict resolution:
  - All `UU` (both modified) files: resolved with `ours` (keep `zyp-dev` content)
  - `DU` (deleted by us, modified by tag) file: keep deletion on current branch

## 3. Conflict Points

Total conflicts: `55`

- `UU` (both modified): `54`
- `DU` (deleted by current branch, modified by tag): `1`

### 3.1 Infra / Build / Deploy Conflicts

- `UU` `.gitignore`
- `UU` `backend/cmd/jwtgen/main.go`
- `UU` `backend/cmd/server/wire.go`
- `UU` `backend/cmd/server/wire_gen.go`
- `UU` `backend/go.sum`
- `UU` `deploy/config.example.yaml`
- `UU` `deploy/docker-compose.yml`
- `DU` `tools/check_pnpm_audit_exceptions.py`

### 3.2 Backend Business / Routing / Service Conflicts

- `UU` `backend/internal/config/config.go`
- `UU` `backend/internal/handler/admin/setting_handler.go`
- `UU` `backend/internal/handler/auth_handler.go`
- `UU` `backend/internal/handler/dto/settings.go`
- `UU` `backend/internal/handler/openai_gateway_handler.go`
- `UU` `backend/internal/handler/setting_handler.go`
- `UU` `backend/internal/repository/account_repo.go`
- `UU` `backend/internal/repository/usage_log_repo.go`
- `UU` `backend/internal/server/middleware/admin_auth_test.go`
- `UU` `backend/internal/server/routes/admin.go`
- `UU` `backend/internal/server/routes/gateway.go`
- `UU` `backend/internal/service/account_test_service.go`
- `UU` `backend/internal/service/account_usage_service.go`
- `UU` `backend/internal/service/admin_service.go`
- `UU` `backend/internal/service/auth_service.go`
- `UU` `backend/internal/service/billing_service.go`
- `UU` `backend/internal/service/dashboard_service.go`
- `UU` `backend/internal/service/domain_constants.go`
- `UU` `backend/internal/service/ratelimit_service.go`
- `UU` `backend/internal/service/setting_service.go`
- `UU` `backend/internal/service/token_refresh_service.go`
- `UU` `backend/internal/service/wire.go`
- `UU` `backend/internal/web/embed_on.go`
- `UU` `backend/resources/model-pricing/model_prices_and_context_window.json`

### 3.3 Frontend UI / State / i18n Conflicts

- `UU` `frontend/src/App.vue`
- `UU` `frontend/src/api/admin/dashboard.ts`
- `UU` `frontend/src/api/admin/settings.ts`
- `UU` `frontend/src/components/account/AccountUsageCell.vue`
- `UU` `frontend/src/components/account/BulkEditAccountModal.vue`
- `UU` `frontend/src/components/account/EditAccountModal.vue`
- `UU` `frontend/src/components/admin/account/AccountBulkActionsBar.vue`
- `UU` `frontend/src/components/admin/account/AccountTableFilters.vue`
- `UU` `frontend/src/components/common/AnnouncementBell.vue`
- `UU` `frontend/src/components/layout/AppHeader.vue`
- `UU` `frontend/src/components/layout/AppSidebar.vue`
- `UU` `frontend/src/i18n/locales/en.ts`
- `UU` `frontend/src/i18n/locales/zh.ts`
- `UU` `frontend/src/main.ts`
- `UU` `frontend/src/stores/app.ts`
- `UU` `frontend/src/types/index.ts`
- `UU` `frontend/src/views/admin/AccountsView.vue`
- `UU` `frontend/src/views/admin/DashboardView.vue`
- `UU` `frontend/src/views/admin/GroupsView.vue`
- `UU` `frontend/src/views/admin/SettingsView.vue`
- `UU` `frontend/src/views/admin/UsageView.vue`
- `UU` `frontend/src/views/auth/RegisterView.vue`
- `UU` `frontend/src/views/user/UsageView.vue`

## 4. Why These Conflicts Happened

- Long-lived branch divergence: both sides evolved significantly from merge-base.
- Cross-cutting feature overlap: both sides changed shared hotspots (`service`, `route`, `dashboard`, `usage`, `settings`).
- Lock/config churn: dependency and deployment files (`go.sum`, `docker-compose`, config templates) are naturally conflict-prone.
- Frontend shared shell changes: app bootstrap, layout, store, i18n and admin pages were modified by both sides.
- Lifecycle mismatch (`DU`): current branch removed a tooling script while tag side kept updating it.

## 5. Recommended Conflict Resolution Plan

Current merge resolution already guarantees no overwrite of current-branch logic on conflicted files. To safely absorb important upstream behavior from `v0.1.103`, apply selective follow-up merges:

1. Feature re-apply by topic (recommended)
- Topic A: usage/upstream-model tracking
- Topic B: admin settings and dashboard changes
- Topic C: deploy/runtime adjustments
- For each topic, inspect tag diff and cherry-pick only validated hunks into `zyp-dev`.

2. High-risk file review priority
- Backend: `backend/internal/repository/usage_log_repo.go`, `backend/internal/service/dashboard_service.go`, `backend/internal/service/setting_service.go`, `backend/internal/service/ratelimit_service.go`
- Frontend: `frontend/src/views/admin/UsageView.vue`, `frontend/src/views/admin/DashboardView.vue`, `frontend/src/api/admin/dashboard.ts`, `frontend/src/api/admin/settings.ts`
- Deploy: `deploy/docker-compose.yml`, `deploy/config.example.yaml`

3. Validation checklist after merge commit
- Backend unit/integration tests
- Frontend unit tests and build
- End-to-end smoke for admin dashboard/usage/settings
- Config and docker compose startup check

## 6. Practical Commands for Follow-up (Optional)

- Compare a conflicted file against tag version:
  - `git diff HEAD v0.1.103 -- <file>`
- Restore only one file from tag into working tree for manual merge:
  - `git checkout v0.1.103 -- <file>`
- Review merge resolution for conflicted files:
  - `git log --oneline --merges -n 5`

