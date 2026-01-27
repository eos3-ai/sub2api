# 开发票功能实施方案（可落地 MVP + 可扩展）

**项目：** Sub2API  
**文档类型：** Feature Design / Implementation Plan  
**更新时间：** 2026-01-26  
**目标读者：** 产品/后端/前端/运维/管理员  

---

## 0. 背景 & 现状

目前系统已有支付/充值订单体系：

- 用户端：`/api/v1/payment/orders`（下单、查看自己的订单）
- 管理端：`/api/v1/admin/payment/orders`（查询所有订单）、`/api/v1/admin/payment/orders/export`
- 订单核心字段（`payment_orders` 表）：
  - `status`：`pending|paid|failed|expired|cancelled|refunded`（已存在常量 `PaymentStatusPaid` 等）
  - `amount_cny`：用户实际支付的人民币金额（用于对账/支付金额展示）
  - `total_usd`：到账额度（USD，余额加值）
  - `provider/channel`：支付提供方/渠道

**新增需求：** 支持用户对已支付订单申请开票，管理员审核/开具并回传电子发票（PDF/URL），并可邮件通知用户。

---

## 1. 目标（MVP）

### 1.1 用户侧目标

- 用户可以看到“可开票订单/金额”
- 用户可以提交“开票申请”，填写抬头/税号/接收邮箱等信息
- 用户可以查看开票申请状态与发票下载链接（若已开具）
- 用户填写的开票信息会保存为默认信息，便于下次快速复用

### 1.2 管理员侧目标

- 管理员可以查看所有开票申请，按状态/用户/时间筛选
- 管理员可以审核（通过/驳回）申请
- 管理员可以“标记已开票”，录入发票号码/开票日期/发票 PDF 链接，并触发邮件通知

---

## 2. 非目标（MVP 不做）

- 不做自动对接税盘/第三方开票平台（后续迭代支持）
- 不做“部分开票/拆分开票/红冲”完整闭环（后续迭代支持）
- 不做“税务合规全链路自动化”能力（如红冲/作废/回调对账等，后续迭代支持）
- 不做复杂的商品明细多行（MVP 先按“技术服务费/充值服务”一行汇总）

---

## 3. 关键口径（务必先统一）

### 3.1 哪些订单可开票？

**默认仅允许：**

- `payment_orders.status == paid`
- `payment_orders.amount_cny > 0`（实际支付金额为 0 的订单不允许开票）
- 订单未被开票占用（避免重复开票）

**通常不允许/没有意义的：**

- `admin_recharge`（后台调账）——常见 `amount_cny=0`
- `activity_recharge`（活动赠送/邀请奖励）——`amount_cny=0`

> 说明：当前系统的 `order_type` 是由 `provider` 派生出来的展示字段（在线/后台/活动）。是否允许开票最好按 `amount_cny` 与 `status` 做硬约束，避免逻辑漂移。

### 3.2 发票金额（CNY）怎么计算？

- 开票金额（CNY） = 所选订单的 `Σ amount_cny`
- 充值额度（USD）可作为辅助展示：`Σ total_usd`

### 3.3 一笔订单能否多次开票？

**MVP 建议：不允许。**  
一个 `payment_order` 只能被一个开票申请引用（最终只开一次票）。

实现上用数据库唯一约束来保证，避免并发重复提交导致“重复开票”。

---

## 4. 业务流程（MVP）

### 4.1 用户端流程

1) 进入在线充值页 `/payment`，在“我的订单”区域右侧（刷新按钮旁）点击“开发票”
2) 弹出“开发票”弹窗/页面，系统列出“可开票订单（已支付且未开票）”，支持多选
3) 填写发票信息（默认自动回填用户已保存的开票信息）：
   - 发票类型：普票/专票（MVP 支持两者）
   - 抬头类型：个人/企业
   - 发票抬头（必填）
   - 纳税人识别号（企业必填，个人可选）
   - （专票必填）地址、电话、开户行、银行账号
   - 接收邮箱（必填，用于接收电子发票）
   - 备注（选填）
4) 提交申请 → 状态变为 `submitted`（待审核），并将本次填写的发票信息保存为“默认开票信息”（用于下次自动回填）
5) 后续可查看状态：
   - `approved`：审核通过，等待开票
   - `issued`：已开票，可下载/查看
   - `rejected`：驳回（展示原因）
   - `cancelled`：用户撤销（仅限审核前）

### 4.2 管理端流程

1) 管理员进入“开票管理”页
2) 查看申请列表 + 详情（订单列表、金额、用户信息、发票抬头等）
3) 审核：
   - 通过 → `approved`
   - 驳回（填写原因）→ `rejected`
4) 开票完成后：
   - 录入发票信息：发票号码/开票日期/发票 PDF 下载链接
   - 标记 `issued`
   - 触发邮件通知用户

---

## 5. 数据模型设计（SQL Migration 方案）

本项目已有 `backend/migrations/*.sql`，建议用 SQL Migration 增量添加以下表。

### 5.1 `invoice_requests`（开票申请主表）

建议字段（示例）：

- `id BIGSERIAL PRIMARY KEY`
- `invoice_request_no VARCHAR(64) UNIQUE NOT NULL`（内部单号，便于追踪）
- `user_id BIGINT NOT NULL`
- `status VARCHAR(32) NOT NULL`  
  建议枚举：`draft|submitted|approved|rejected|issued|cancelled`
- `invoice_type VARCHAR(16) NOT NULL`（发票类型：`normal|special`，分别对应普票/专票）
- `invoice_title VARCHAR(255) NOT NULL`（抬头）
- `buyer_type VARCHAR(16) NOT NULL`（`personal|company`）
- `tax_no VARCHAR(64) DEFAULT ''`（企业税号）
- `buyer_address VARCHAR(255) DEFAULT ''`（专票必填）
- `buyer_phone VARCHAR(32) DEFAULT ''`（专票必填）
- `buyer_bank_name VARCHAR(128) DEFAULT ''`（专票必填）
- `buyer_bank_account VARCHAR(64) DEFAULT ''`（专票必填）
- `receiver_email VARCHAR(255) NOT NULL`
- `receiver_phone VARCHAR(32) DEFAULT ''`
- `invoice_item_name VARCHAR(255) DEFAULT ''`（开票内容/商品名称，默认“技术服务费”，支持后台配置）
- `remark TEXT DEFAULT ''`（用户备注，可选）
- 金额快照：
  - `amount_cny_total NUMERIC(18,2) NOT NULL DEFAULT 0`
  - `total_usd_total NUMERIC(18,8) NOT NULL DEFAULT 0`（可选）
- 审核/开票信息：
  - `reviewed_by BIGINT NULL`（管理员ID，可选）
  - `reviewed_at TIMESTAMP NULL`
  - `reject_reason TEXT DEFAULT ''`
  - `issued_by BIGINT NULL`
  - `issued_at TIMESTAMP NULL`
  - `invoice_number VARCHAR(64) DEFAULT ''`（发票号码）
  - `invoice_date DATE NULL`
  - `invoice_pdf_url TEXT DEFAULT ''`（PDF/下载链接）
- `created_at TIMESTAMP NOT NULL DEFAULT NOW()`
- `updated_at TIMESTAMP NOT NULL DEFAULT NOW()`

索引建议：

- `(user_id, created_at DESC)`
- `(status, created_at DESC)`

### 5.2 `invoice_order_items`（申请-订单明细表）

用于把多个订单归集到一个开票申请，并做“去重开票”的强约束。

- `id BIGSERIAL PRIMARY KEY`
- `invoice_request_id BIGINT NOT NULL REFERENCES invoice_requests(id) ON DELETE CASCADE`
- `payment_order_id BIGINT NOT NULL`（或存 `order_no`，建议存 `id` 更稳）
- `order_no VARCHAR(64) NOT NULL`（冗余便于展示）
- 金额快照：
  - `amount_cny NUMERIC(18,2) NOT NULL`
  - `total_usd NUMERIC(18,8) NOT NULL`
- `created_at TIMESTAMP NOT NULL DEFAULT NOW()`

关键约束：

- `UNIQUE(payment_order_id) WHERE active=true`（推荐）：保证同一时刻一笔订单只会被“有效申请”占用；当申请被用户取消或管理员驳回时，将 `active=false` 释放订单，使其可再次申请开票
- 索引：`(invoice_request_id)`

### 5.3 `invoice_profiles`（用户默认开票信息，便于复用）

用于保存用户常用的开票信息（抬头/税号/邮箱等），让用户下次“开发票”时默认回填。

> MVP 建议：一个用户仅维护 1 份默认信息（`UNIQUE(user_id)`）；后续再扩展为多抬头管理（多条 profile + is_default）。

建议字段（示例）：

- `id BIGSERIAL PRIMARY KEY`
- `user_id BIGINT NOT NULL UNIQUE`
- `invoice_type VARCHAR(16) NOT NULL`（发票类型：`normal|special`）
- `buyer_type VARCHAR(16) NOT NULL`（`personal|company`）
- `invoice_title VARCHAR(255) NOT NULL`
- `tax_no VARCHAR(64) DEFAULT ''`
- `buyer_address VARCHAR(255) DEFAULT ''`
- `buyer_phone VARCHAR(32) DEFAULT ''`
- `buyer_bank_name VARCHAR(128) DEFAULT ''`
- `buyer_bank_account VARCHAR(64) DEFAULT ''`
- `receiver_email VARCHAR(255) NOT NULL`
- `receiver_phone VARCHAR(32) DEFAULT ''`
- `remark TEXT DEFAULT ''`
- `created_at TIMESTAMP NOT NULL DEFAULT NOW()`
- `updated_at TIMESTAMP NOT NULL DEFAULT NOW()`

索引建议：

- `(user_id)`（唯一索引已覆盖）

---

## 6. 后端实现方案（Handlers → Service → Repo）

### 6.1 用户 API（建议）

> 路由风格与现有一致：`/api/v1/...`

1) `GET /api/v1/invoices/eligible-orders`
- 返回当前用户可开票订单列表（`paid` 且未被 `invoice_order_items` 引用）
- 支持时间筛选：`from/to`（RFC3339）

2) `POST /api/v1/invoices`
- 创建开票申请
- body 示例：
```json
{
  "order_nos": ["PO2026...001", "PO2026...002"],
  "invoice_type": "special",
  "buyer_type": "company",
  "invoice_title": "某某科技有限公司",
  "tax_no": "9131XXXXXXXXXXXXXX",
  "buyer_address": "上海市xx区xx路xx号",
  "buyer_phone": "021-12345678",
  "buyer_bank_name": "xx银行xx支行",
  "buyer_bank_account": "6222************",
  "receiver_email": "finance@example.com",
  "receiver_phone": "13800000000",
  "invoice_item_name": "技术服务费",
  "remark": ""
}
```
- 关键校验：
  - 订单属于当前用户
  - `order_nos` 最多 5 笔（防止合并过多导致请求过大/对账困难）
  - `status==paid` 且 `amount_cny>0`
  - 未被开票占用（依赖唯一约束 + 事务处理）
  - `invoice_type` 校验：
    - `normal`：普票（允许个人/企业）
    - `special`：专票（仅允许企业，且地址/电话/开户行/账号必填）
- 保存默认开票信息：
  - 成功创建开票申请后，将本次发票信息 **Upsert** 到 `invoice_profiles`（让下次可直接复用）

3) `GET /api/v1/invoices`
- 用户查看自己的开票申请列表（分页）

4) `GET /api/v1/invoices/:id`
- 申请详情 + 关联订单列表 + 发票链接

5) `POST /api/v1/invoices/:id/cancel`
- 用户撤销（仅允许 `submitted` 状态）

6) `GET /api/v1/invoices/profile`
- 获取用户默认开票信息（用于前端自动回填）

7) `PUT /api/v1/invoices/profile`
- 更新用户默认开票信息（允许用户先维护资料再开发票）

### 6.2 管理员 API（建议）

1) `GET /api/v1/admin/invoices`
- 列表 + 筛选（status/user_email/date range）

2) `GET /api/v1/admin/invoices/:id`
- 详情

3) `POST /api/v1/admin/invoices/:id/approve`
- 审核通过（记录 `reviewed_by/reviewed_at`）

4) `POST /api/v1/admin/invoices/:id/reject`
- 驳回（记录原因）

5) `POST /api/v1/admin/invoices/:id/issue`
- 标记已开票并写入：
  - `invoice_number`
  - `invoice_date`
  - `invoice_pdf_url`
  - `issued_by/issued_at`
- 成功后调用 `EmailService` 给 `receiver_email` 发通知

> 备注：本项目已具备 SMTP 邮件发送能力（`EmailService.SendEmail`），可复用。

---

## 7. 前端实现方案（Vue）

### 7.1 用户侧页面建议

入口调整（按需求落实）：

- 在线充值页 `/payment` 的“我的订单”区域右侧：在“刷新”按钮旁新增“开发票”按钮  
  - 参考现有页面结构：`frontend/src/views/user/PaymentView.vue` 的 Orders 标题行

交互建议（MVP）：

- 点击“开发票”后弹出 Dialog/Drawer：
  - 列出可开票订单（默认勾选：当前列表里 `paid` 且未开票的订单）
  - 发票信息表单默认从 `GET /api/v1/invoices/profile` 自动回填
  - 提交成功后：
    - 调用 `POST /api/v1/invoices` 创建申请
    - 后端自动 Upsert `invoice_profiles`，前端无需额外保存动作

新增页面：`/invoices`（“我的发票/开票申请”）用于查看历史与下载

- 顶部：可开票金额汇总（从 eligible-orders 计算）
- 中部：可开票订单列表（多选）
- 表单：抬头、税号、邮箱、备注
- 提交后：跳转到详情页或在列表里展示状态

补充入口（可选）：

- 用户菜单里增加“发票”（进入 `/invoices`）

### 7.2 管理员侧页面建议

新增页面：`/admin/invoices`

功能：

- 列表（状态、用户邮箱、金额、创建时间）
- 筛选（状态/用户/时间）
- 操作按钮（通过/驳回/标记已开票）
- 详情弹窗展示：
  - 订单明细（order_no、amount_cny、total_usd、paid_at）
  - 发票信息（抬头、税号、邮箱）

---

## 8. 配置与依赖（运维可控）

### 8.1 功能开关

- 建议新增配置：`INVOICE_ENABLED=true/false`
- 若关闭：前端隐藏入口；后端接口返回 404/503（与 Payment 模块一致的风格）

### 8.2 开票方信息（销售方）

建议在系统设置里增加“开票信息”配置（可由管理员维护）：

- 销售方名称、税号、地址电话、开户行账号（后续对接开票平台会用）
- 默认商品名称（如：技术服务费/充值服务）
- 税率（可选）

### 8.3 发票文件存储

MVP 推荐：**存 URL，不直接存文件内容**

- 管理员在开票平台下载 PDF → 上传到对象存储（S3/OSS/COS） → 回填 URL
- 系统仅保存 `invoice_pdf_url` 并邮件通知用户

后续可选：实现受控上传接口（限制大小 + 鉴权 + 防盗链）

---

## 9. 安全与一致性（关键点）

### 9.1 权限控制

- 用户接口：只能访问自己的发票申请/订单
- 管理接口：仅管理员可访问（复用现有 Admin JWT/Auth 中间件）

### 9.2 防重复开票

必须使用 **数据库唯一约束** + **事务**：

- `invoice_order_items.payment_order_id UNIQUE`
- 创建申请时：
  1) 校验订单属于用户且 `paid`
  2) 插入 `invoice_requests`
  3) 批量插入 `invoice_order_items`  
  若遇到唯一冲突 → 返回 “订单已被开票/已在申请中”

### 9.3 数据敏感性

税号/地址等属于敏感信息：

- 管理端展示可完整展示
- 日志中避免打印完整税号（必要时脱敏）
- 如需更严格合规，可做字段加密（后续迭代）

---

## 10. 测试与验收标准（建议）

### 10.1 后端

- 单测：
  - 仅 `paid` 且 `amount_cny>0` 可创建申请
  - 同一订单重复提交会被唯一约束拦截
  - 状态流转合法性（submitted→approved→issued；submitted→rejected；submitted→cancelled）
- 集成测试（可选）：创建订单 → 标记 paid → 申请开票 → admin 开票 → 用户可见

### 10.2 前端

- 表单校验：企业税号必填、邮箱格式校验
- 多选订单金额合计正确
- 状态展示正确（含驳回原因、下载链接）

### 10.3 验收口径

- 任何时刻，同一笔 `payment_order` 只能对应一张发票（或一个申请）
- 用户只能看到自己的申请
- 管理员能按筛选查看并导出
- 生成发票后用户能收到邮件（SMTP 配置存在时）

---

## 11. 迭代路线图（建议）

### Phase 0（需求口径确认，0.5~1 天）

- 明确开票范围：仅 `paid` 且 `amount_cny>0` 是否为最终口径
- 明确是否允许多订单合并开票（本方案默认允许）
- 明确退款/撤销/驳回等状态流转与对账口径
- 明确抬头类型字段（个人/企业）与企业必填字段（税号等）

### Phase 1（后端 & 数据层 MVP，2~4 天）

- 新增数据表：`invoice_requests`、`invoice_order_items`、`invoice_profiles`
- 实现用户端开票 API：可开票订单、创建申请（事务 + 防重复开票）、申请列表/详情/撤销
- 实现默认开票信息 API：`GET/PUT /api/v1/invoices/profile`
- 实现管理端开票 API：申请列表/详情/审核通过/驳回/标记已开票
- 支持邮件通知（复用现有 SMTP `EmailService`，可配置开关）

### Phase 2（用户前端 MVP，2~4 天）

- `/payment` 页“我的订单”标题行：在“刷新”按钮旁新增“开发票”按钮入口
- “开发票”弹窗/抽屉：订单多选 + 发票信息表单
- 自动回填默认开票信息（从 `/api/v1/invoices/profile` 获取）
- 提交后展示结果，并提供入口跳转到“我的发票/申请”

### Phase 3（管理员前端 MVP，2~4 天）

- 新增 `/admin/invoices`：列表 + 筛选 + 详情
- 审核通过/驳回操作、开票信息录入（发票号码/日期/PDF 链接）
- 支持导出（可选：CSV）

### Phase 4（增强/自动化，2~4 周）

- 对接第三方电子发票平台（诺诺/百望/航信等其一）
- 自动开票、自动回传 PDF、失败重试、回调签名校验

### Phase 5（复杂场景，视需求）

- 退款/红冲流程
- 订单拆分/部分开票
- 发票抬头管理（用户保存多个抬头）
- 发票导出/对账报表

---

## 12. Phase 0 待确认问题（建议产品先拍板）

下面问题会直接影响 Phase 1 的数据库结构、接口契约与 UI 交互，请在开始开发前确认。

### 12.1 必选决策

1) **发票类型**
   - 选项：`电子普票` / `专票` / `两者都支持`
   - 推荐（MVP）：`电子普票`

2) **金额口径**
   - 问题：开票金额是否严格等于 `Σ amount_cny`（实际支付人民币）？
   - 选项：`是（推荐）` / `否（按其它规则）`
   - 推荐（MVP）：`是`

3) **合并规则**
   - 问题：是否允许多笔订单合并开一张票？
   - 选项：`允许（推荐）` / `不允许（每单一票）`
   - 推荐（MVP）：`允许`
   - 需确认：单次合并最多允许几笔订单（推荐 20 笔，防止超大请求）

4) **保存/复用开票信息**
   - 问题：用户提交开票申请后，是否自动将本次信息保存为默认开票信息（`invoice_profiles`）？
   - 选项：`自动保存（推荐）` / `用户手动保存` / `不保存`
   - 推荐（MVP）：`自动保存`
   - 需确认：是否要支持“多抬头管理”（多个 profile + 默认项）
     - 推荐（MVP）：先仅支持 1 份默认信息（`UNIQUE(user_id)`）

5) **状态流转与可撤销/可修改**
   - 问题：用户提交后是否允许修改抬头/税号/邮箱？
   - 选项：`不允许（推荐：撤销重提）` / `允许（需更复杂的审计与版本）`
   - 推荐（MVP）：`不允许`
   - 问题：用户是否允许撤销？
     - 推荐（MVP）：仅允许在 `submitted` 状态撤销（审核后不可撤销）

6) **退款/反向操作策略**
   - 问题：开票前如何处理退款订单？
     - 推荐（MVP）：`refunded` 订单不可开票（不可出现在 eligible-orders）
   - 问题：已开票后发生退款怎么办？
     - 推荐（MVP）：系统标记“需人工处理”，红冲/作废流程放到 Phase 5

### 12.2 可选决策（不影响核心，但建议提前确定）

1) **开票内容（商品/服务名称）**
   - 推荐：系统设置可配置一个默认值（如：`技术服务费` / `信息技术服务费`）

2) **收票邮箱规则**
   - 是否允许填写非登录邮箱？
   - 推荐：允许（财务/公司统一邮箱常见）

3) **最小/最大开票金额**
   - 推荐：最小金额可设为 1 元；最大金额按业务决定（或不限制）

### 12.3 你可以直接按这个模板回复（Phase 0 产出）

```text
1) 发票类型：电子普票 / 专票 / 都支持
2) 金额口径：Σ amount_cny（是/否）
3) 合并开票：允许/不允许；合并上限：__ 笔
4) 保存默认信息：自动保存/手动保存/不保存；多抬头：要/不要
5) 提交后修改：允许/不允许；用户撤销：允许（仅 submitted）/不允许
6) 退款策略：开票前退款不可开票（是/否）；开票后退款：__（人工处理/支持红冲/其它）
7) 默认开票内容：__
8) 收票邮箱：允许非登录邮箱（是/否）
```

### 12.4 Phase 0 决策结果（已确认）

> 以下为当前已确认的 Phase 0 结论，将作为 Phase 1 开发的输入。

1) 发票类型：**都支持（普票 + 专票）**
2) 金额口径：**开票金额 = Σ amount_cny（是）**
3) 合并开票：**允许**；合并上限：**5 笔**
4) 保存默认信息：**自动保存**；多抬头：**不要（仅 1 份默认信息，UNIQUE(user_id)）**
5) 提交后修改：**不允许**；用户撤销：**允许（仅 submitted）**
6) 退款策略：开票前退款不可开票：**是**；开票后退款：**人工处理**
7) 默认开票内容：**技术服务费**（默认值；支持后台配置覆盖）
8) 收票邮箱：允许非登录邮箱：**是**
