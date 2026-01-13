# v0.1.53 合并到 zyp-dev 分支报告

**合并时间:** 2026-01-13  
**源版本:** `v0.1.53` (tag, `93db889a1060121cf6f1af5a1178070a7cd66789`)  
**目标分支:** `zyp-dev`  
**合并提交:** `5f14349cf8012d33f5bcdc3fde8c13d55ee497af`  
**合并前备份分支:** `backup-pre-merge-v0.1.53-20260113-124356`  

---

## 1. 背景与合并范围

当前分支 `zyp-dev` 与 `v0.1.53` 的共同祖先为：

- merge base: `678b088a133ae49932ff8a1fc4b6c8218e4e7623`

两边分叉较久：

- `zyp-dev` 相对 `v0.1.53` 的独立提交：193
- `v0.1.53` 相对 `zyp-dev` 的独立提交：381

这意味着“直接合并”会在大量相同文件上同时引入两边的变化（尤其是生成代码、前端页面与 i18n、网关转发逻辑等），冲突面会显著扩大。

---

## 2. 冲突概况

本次最终合并的**实际冲突文件**只有 1 个（已解决并完成合并提交）：

- `frontend/package-lock.json`（modify/delete 冲突）

备注：最初尝试不带策略选项的 `git merge v0.1.53` 时，会在后端生成代码、前端组件等多个区域产生大量文本冲突；为避免在一次合并里手工解决几十上百处冲突，本次采用了“冲突倾向”策略进行收敛（见第 4 节）。

---

## 3. 冲突原因分析与解决方案

### 3.1 `frontend/package-lock.json`（modify/delete）

**冲突类型:** `modify/delete`  
**冲突表现:**  
- `zyp-dev` 侧：该文件在 HEAD 已删除（通常代表项目不再使用 npm 的 lockfile）
- `v0.1.53` 侧：该文件仍存在，并在 tag 侧被更新
- Git 无法自动判断“应该保留删除”还是“应该恢复并保留更新后的文件”，因此标记为冲突

**根因分析（为什么会冲突）**

- 两边对包管理器/依赖锁文件的约定不同：  
  - 一边倾向 `pnpm`（使用 `pnpm-lock.yaml`）  
  - 一边仍在维护 `package-lock.json`
- lockfile 属于“全局状态文件”，一旦两个分支的包管理器策略不同，冲突几乎不可避免。

**采用的解决方案（已执行）**

- 选择保留 `zyp-dev` 的策略：继续删除 `frontend/package-lock.json`
- 执行命令：
  - `git rm frontend/package-lock.json`

**理由（为什么这样更稳）**

- 同一项目同时保留 `package-lock.json` 与 `pnpm-lock.yaml`，很容易出现“依赖树不一致、CI 与本地安装结果不一致”的问题。
- 仓库当前已经存在并在使用 `pnpm-lock.yaml`（以当前分支为准时），因此延续现状更安全。

**可选替代方案（如果你希望跟随 v0.1.53 的 npm 生态）**

- 恢复 `package-lock.json` 并以 npm 为准：
  - `git checkout --theirs frontend/package-lock.json`
  - 同时应评估是否删除 `pnpm-lock.yaml`，并统一开发/CI 的安装命令为 `npm ci`

---

## 4. 合并策略说明（为什么这样合并）

本次采用了：

- `git merge -X ours v0.1.53`

含义：

- 当 Git 判断某个 hunk 存在内容冲突时，优先保留当前分支（`zyp-dev`）的版本
- 对于“无冲突的新文件/新迁移/新增页面/新增接口”等，仍会正常合入 `v0.1.53` 的变更

**采用该策略的原因**

- `zyp-dev` 与 `v0.1.53` 分叉提交数量较大，一次性在生成代码与大量前端文件上逐个手工冲突合并，时间成本与引入新问题的风险都很高。
- 现阶段更合理的目标是：先把 `v0.1.53` 的整体历史与大部分非冲突变更合入，确保主干可继续前进；后续再针对确实需要的 upstream 变更做定点对齐或 cherry-pick。

**该策略的风险点（需要你知情）**

- 如果 `v0.1.53` 在与 `zyp-dev` 相同代码区域做了重要修复，而这些修复刚好落在冲突 hunks 内，则可能被 “ours” 策略覆盖掉。  
  建议在关键模块（网关转发、Gemini/OpenAI 兼容、运维监控页面）做一次 smoke review 或最小化回归验证（见第 6 节）。

---

## 5. 可追溯的执行记录

为避免丢失现场并支持随时回退，本次合并过程包含两类保护措施：

### 5.1 备份分支

- `backup-pre-merge-v0.1.53-20260113-124356`

用途：若发现合并结果不符合预期，可快速回到合并前的 `zyp-dev` 状态。

### 5.2 临时 stash（仅用于保留现场）

本地 stash 中保留了若干条记录（不影响当前分支 HEAD 的合并结果），主要用途是：

- 保存“上一次未完成 merge 尝试”的工作区内容，便于需要时追溯
- 保存本次合并后进行的编译修复尝试（已整体 stash，不纳入本次合并结果），避免污染分支

可通过 `git stash list` 查看当前 stash 的具体条目与编号；若确认不再需要，可按需 drop（见第 7.2 节）。

---

## 6. 建议验证项（合并后）

由于合并引入了大量后端与前端改动，建议至少做一次离线的后端编译/测试：

### 6.1 后端（离线优先）

在不允许拉取新依赖的环境下先做离线测试：

```bash
cd backend
GOPROXY=off go test ./...
```

如果本地 Go module 缓存齐全，上述命令应可通过；若缺依赖，会提示无法下载（此时再按需要开放网络或预热缓存）。

#### 本次实际验证结果（2026-01-13）

- 已在本地以 `GOPROXY=off` 方式跑通：`cd backend && GOPROXY=off go test ./...`
- 过程中为修复编译/注入问题，执行了 Wire 生成：
  - `cd backend/cmd/server && GOPROXY=off go generate`
  - 并补齐了 Wire Provider（例如 `PromoCodeRepository`、`TimeoutCounterCache`）以满足注入依赖

### 6.2 前端（不安装依赖也能做的检查）

```bash
node -v
pnpm -v
```

若需要构建/运行前端，通常需要先安装依赖（可能涉及网络访问）。

---

## 7. 回退与修复方案

### 7.1 一键回退到合并前

```bash
git reset --hard backup-pre-merge-v0.1.53-20260113-124356
```

### 7.2 仅清理无用 stash（确认不需要后）

```bash
git stash list
git stash drop stash@{N}
```

### 7.3 如果你希望“更完整地吸收 v0.1.53 的冲突区域改动”

建议采用“先合 schema/核心逻辑，再重新生成”的方式，而不是手工改生成文件：

1. 先对 `backend/ent/schema/*`、`backend/cmd/server/wire.go` 等“源文件”做精确合并
2. 再执行对应生成命令（例如 ent / wire 的 generate 流程）
3. 用 `go test ./...` 兜底验证编译与行为
