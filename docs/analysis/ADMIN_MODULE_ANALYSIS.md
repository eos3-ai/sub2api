# Sub2API Admin Module - Comprehensive Technical Analysis

**Project:** Sub2API
**Analysis Date:** 2026-01-08
**Version:** Current (zyp-dev branch)
**Document Type:** Technical Architecture Analysis

---

## Executive Summary

The Admin Module is a comprehensive administrative interface for managing the Sub2API platform. It provides full control over users, AI platform accounts (Claude/OpenAI/Gemini/Antigravity), API keys, subscriptions, system settings, and usage analytics. The module follows a clean three-tier architecture: Backend Handlers → Service Layer → Database Layer, with a Vue.js frontend consuming RESTful APIs.

**Key Metrics:**
- **Backend Handlers:** 16 handler files
- **Frontend Views:** 10 admin views
- **API Endpoints:** 100+ admin endpoints
- **Supported Platforms:** Anthropic (Claude), OpenAI, Google (Gemini), Antigravity
- **Authentication:** JWT-based admin role verification

---

## Table of Contents

1. [Architecture Overview](#1-architecture-overview)
2. [Backend Analysis](#2-backend-analysis)
3. [Frontend Analysis](#3-frontend-analysis)
4. [Feature Mapping](#4-feature-mapping)
5. [Data Flow Analysis](#5-data-flow-analysis)
6. [API Endpoint Documentation](#6-api-endpoint-documentation)
7. [Security & Authorization](#7-security--authorization)
8. [Key Findings & Recommendations](#8-key-findings--recommendations)

---

## 1. Architecture Overview

### 1.1 System Architecture

```
┌─────────────────────────────────────────────────────────────────┐
│                        Frontend Layer                            │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐          │
│  │ Vue.js Views │  │  Components  │  │  API Client  │          │
│  └──────────────┘  └──────────────┘  └──────────────┘          │
└─────────────────────────────┬───────────────────────────────────┘
                              │ HTTP/REST (JSON)
                              ↓
┌─────────────────────────────────────────────────────────────────┐
│                        Backend Layer                             │
│  ┌──────────────────────────────────────────────────────────┐  │
│  │ Router (Gin) + Middleware (JWT + Admin Auth)            │  │
│  └────────────────────────┬─────────────────────────────────┘  │
│                           ↓                                     │
│  ┌─────────────────────────────────────────────────────────┐  │
│  │  Admin Handlers (16 modules)                            │  │
│  │  - AccountHandler    - UserHandler                      │  │
│  │  - DashboardHandler  - GroupHandler                     │  │
│  │  - SettingHandler    - ProxyHandler                     │  │
│  │  - SubscriptionHandler - RedeemHandler                  │  │
│  │  - UsageHandler      - PaymentOrdersHandler            │  │
│  │  - OAuth Handlers (Claude/OpenAI/Gemini/Antigravity)   │  │
│  └────────────────────────┬─────────────────────────────────┘  │
│                           ↓                                     │
│  ┌─────────────────────────────────────────────────────────┐  │
│  │  Service Layer (Business Logic)                         │  │
│  │  - AdminService     - DashboardService                  │  │
│  │  - OAuthService     - SubscriptionService               │  │
│  │  - UsageService     - SettingService                    │  │
│  └────────────────────────┬─────────────────────────────────┘  │
│                           ↓                                     │
│  ┌─────────────────────────────────────────────────────────┐  │
│  │  Repository Layer (Data Access)                         │  │
│  │  - PostgreSQL/MySQL Database                            │  │
│  └─────────────────────────────────────────────────────────┘  │
└─────────────────────────────────────────────────────────────────┘
```

### 1.2 Module Organization

**Backend Structure:**
```
backend/internal/handler/admin/
├── account_handler.go           # AI account management
├── user_handler.go              # User management
├── dashboard_handler.go         # Statistics & metrics
├── group_handler.go             # API key group management
├── setting_handler.go           # System settings
├── proxy_handler.go             # HTTP/SOCKS proxy management
├── subscription_handler.go      # User subscription management
├── redeem_handler.go            # Redeem code management
├── usage_handler.go             # Usage logs & statistics
├── payment_orders_handler.go    # Payment order management
├── openai_oauth_handler.go      # OpenAI OAuth integration
├── gemini_oauth_handler.go      # Google Gemini OAuth
├── antigravity_oauth_handler.go # Antigravity OAuth
├── system_handler.go            # System updates & maintenance
├── user_attribute_handler.go    # Custom user attributes
└── csv_sanitize.go              # CSV export utilities
```

**Frontend Structure:**
```
frontend/src/
├── views/admin/
│   ├── DashboardView.vue        # Main dashboard with charts
│   ├── UsersView.vue            # User management table
│   ├── AccountsView.vue         # AI account management
│   ├── GroupsView.vue           # Group configuration
│   ├── SettingsView.vue         # System settings panel
│   ├── ProxiesView.vue          # Proxy configuration
│   ├── SubscriptionsView.vue    # Subscription management
│   ├── RedeemView.vue           # Redeem code generation
│   ├── UsageView.vue            # Usage logs viewer
│   └── PaymentOrdersView.vue    # Payment history
├── components/admin/
│   ├── account/                 # Account-specific components
│   ├── usage/                   # Usage-specific components
│   └── user/                    # User-specific components
└── api/admin/
    ├── accounts.ts              # Account API client
    ├── users.ts                 # User API client
    ├── dashboard.ts             # Dashboard API client
    ├── groups.ts                # Group API client
    ├── settings.ts              # Settings API client
    └── [10 more API modules]
```

---

## 2. Backend Analysis

### 2.1 Handler Modules

#### 2.1.1 AccountHandler (`account_handler.go`)

**Purpose:** Manages AI platform accounts (Claude, OpenAI, Gemini, Antigravity) including creation, OAuth flows, credential refresh, and monitoring.

**Key Dependencies:**
- `AdminService`: Core CRUD operations
- `OAuthService`: Claude OAuth (full + setup-token)
- `OpenAIOAuthService`: OpenAI OAuth integration
- `GeminiOAuthService`: Google Gemini OAuth
- `AntigravityOAuthService`: Antigravity platform OAuth
- `RateLimitService`: Rate limit status management
- `AccountUsageService`: Usage statistics
- `AccountTestService`: Connectivity testing (SSE streaming)
- `ConcurrencyService`: Real-time concurrency tracking
- `CRSSyncService`: Sync accounts from claude-relay-service

**Core Operations:**

1. **CRUD Operations:**
   - `List()`: Paginated account listing with filters (platform, type, status, search) + real-time concurrency counts
   - `GetByID()`: Retrieve account details
   - `Create()`: Create new account with OAuth or API key
   - `Update()`: Update account credentials, proxy, concurrency, priority, groups
   - `Delete()`: Remove account
   - `BulkUpdate()`: Batch update multiple accounts

2. **OAuth Management:**
   - `GenerateAuthURL()`: Generate OAuth authorization URL for Claude (full scope)
   - `GenerateSetupTokenURL()`: Generate OAuth URL for setup token (inference-only)
   - `ExchangeCode()`: Exchange authorization code for tokens
   - `CookieAuth()`: Cookie-based OAuth (sessionKey auto-auth)
   - `Refresh()`: Refresh OAuth tokens (Claude, OpenAI, Gemini, Antigravity)

3. **Account Operations:**
   - `Test()`: Test account connectivity with SSE streaming results
   - `GetStats()`: Get usage statistics (requests, tokens, cost) over N days
   - `GetUsage()`: Get current 5h/7d usage windows
   - `GetTodayStats()`: Get today's statistics
   - `ClearError()`: Reset account error status
   - `ClearRateLimit()`: Clear rate limit status
   - `SetSchedulable()`: Toggle account participation in scheduling
   - `GetAvailableModels()`: Get list of available models for account

4. **Advanced Features:**
   - `SyncFromCRS()`: Sync accounts from claude-relay-service (CRS)
   - `BatchCreate()`: Batch create multiple accounts
   - `BatchUpdateCredentials()`: Batch update credentials fields (account_uuid, org_uuid, etc.)
   - `RefreshTier()`: Refresh Google One tier for Gemini accounts
   - `BatchRefreshTier()`: Batch refresh tier for multiple Gemini accounts
   - `GetTempUnschedulable()`: Get temporary unschedulable status
   - `ClearTempUnschedulable()`: Clear temporary unschedulable status

**Data Structures:**
```go
type AccountWithConcurrency struct {
    *dto.Account
    CurrentConcurrency int `json:"current_concurrency"`
}

type CreateAccountRequest struct {
    Name        string
    Platform    string // anthropic|openai|gemini|antigravity
    Type        string // oauth|setup-token|apikey
    Credentials map[string]any
    Extra       map[string]any
    ProxyID     *int64
    Concurrency int
    Priority    int
    GroupIDs    []int64
    ConfirmMixedChannelRisk *bool
}
```

---

#### 2.1.2 UserHandler (`user_handler.go`)

**Purpose:** Complete user lifecycle management including registration, balance operations, custom attributes, and data export.

**Key Operations:**

1. **CRUD Operations:**
   - `List()`: Paginated user listing with filters (status, role, search, custom attributes)
   - `Export()`: CSV export of filtered users
   - `GetByID()`: Retrieve user details
   - `Create()`: Create new user with initial balance and concurrency
   - `Update()`: Update user profile, balance, concurrency, allowed groups
   - `Delete()`: Remove user

2. **Balance Management:**
   - `UpdateBalance()`: Set, add, or subtract balance with audit notes

3. **User Analytics:**
   - `GetUserAPIKeys()`: List user's API keys
   - `GetUserUsage()`: Get usage statistics for specific period

**Advanced Features:**
- **Custom Attribute Filtering**: Supports `attr[{id}]=value` query parameters for filtering by custom user attributes
- **CSV Sanitization**: Uses `sanitizeCSVCell()` to prevent CSV injection attacks

**Data Structures:**
```go
type CreateUserRequest struct {
    Email         string
    Password      string
    Username      string
    Balance       float64
    Concurrency   int
    AllowedGroups []int64
}

type UpdateBalanceRequest struct {
    Balance   float64
    Operation string // set|add|subtract
    Notes     string
}
```

---

#### 2.1.3 DashboardHandler (`dashboard_handler.go`)

**Purpose:** Provides comprehensive system-wide statistics and analytics for monitoring platform health and usage trends.

**Key Services:**
- `DashboardService`: Aggregates data from multiple sources

**Core Endpoints:**

1. **`GetStats()`**: Overall system statistics
   - User metrics (total, today new, active)
   - API key metrics (total, active)
   - Account metrics (total, normal, error, rate-limited, overload)
   - Token usage (cumulative and today): input, output, cache creation, cache read
   - Cost metrics (standard cost vs actual cost)
   - Performance metrics (average duration, RPM, TPM)
   - System uptime

2. **`GetUsageTrend()`**: Time-series usage data
   - Parameters: start_date, end_date, granularity (day/hour), user_id, api_key_id
   - Returns trend data with requests, tokens, cost per time point

3. **`GetModelStats()`**: Model-level usage breakdown
   - Parameters: start_date, end_date, user_id, api_key_id
   - Returns per-model statistics (requests, tokens, cost)

4. **`GetAPIKeyUsageTrend()`**: Top API key usage trends
   - Parameters: start_date, end_date, granularity, limit (default 5)
   - Returns top N API keys with usage trends

5. **`GetUserUsageTrend()`**: Top user usage trends
   - Parameters: start_date, end_date, granularity, limit (default 12)
   - Returns top N users with usage trends

6. **`GetBatchUsersUsage()`**: Batch fetch usage stats for multiple users
   - Input: Array of user IDs
   - Output: Map of user_id → today_actual_cost, total_actual_cost

7. **`GetBatchAPIKeysUsage()`**: Batch fetch usage stats for multiple API keys
   - Input: Array of API key IDs
   - Output: Map of api_key_id → today_actual_cost, total_actual_cost

**Timezone Handling:**
- Uses user's timezone from query parameter `timezone`
- Parses dates in user's local timezone
- Ensures accurate date range filtering across different timezones

---

#### 2.1.4 GroupHandler (`group_handler.go`)

**Purpose:** Manages API key groups which define pricing, rate limits, and platform access for API keys.

**Core Features:**

1. **CRUD Operations:**
   - `List()`: Paginated group listing with filters (platform, status, is_exclusive)
   - `GetAll()`: Get all active groups (no pagination)
   - `GetByPlatform()`: Filter groups by platform
   - `GetByID()`: Retrieve group details
   - `Create()`: Create new group
   - `Update()`: Update group configuration
   - `Delete()`: Remove group

2. **Group Configuration:**
   - Platform selection (anthropic, openai, gemini, antigravity)
   - Rate multiplier (pricing modifier)
   - Exclusive mode (dedicated accounts)
   - Subscription type (standard vs subscription)
   - Usage limits (daily, weekly, monthly USD caps)
   - Image generation pricing (1K, 2K, 4K for Gemini/Antigravity)

3. **Group Analytics:**
   - `GetStats()`: Total/active API keys, requests, cost
   - `GetGroupAPIKeys()`: List API keys in group

**Data Structures:**
```go
type CreateGroupRequest struct {
    Name             string
    Description      string
    Platform         string  // anthropic|openai|gemini|antigravity
    RateMultiplier   float64
    IsExclusive      bool
    SubscriptionType string  // standard|subscription
    DailyLimitUSD    *float64
    WeeklyLimitUSD   *float64
    MonthlyLimitUSD  *float64
    ImagePrice1K     *float64 // Gemini/Antigravity image pricing
    ImagePrice2K     *float64
    ImagePrice4K     *float64
}
```

---

#### 2.1.5 SettingHandler (`setting_handler.go`)

**Purpose:** Manages system-wide configuration including registration, email, security, and OEM settings.

**Configuration Categories:**

1. **Registration Settings:**
   - `registration_enabled`: Enable/disable new user registration
   - `email_verify_enabled`: Require email verification

2. **Email Service (SMTP):**
   - Connection details (host, port, username, password)
   - Sender configuration (from email, from name)
   - TLS encryption
   - `TestSmtpConnection()`: Test SMTP connectivity
   - `SendTestEmail()`: Send test email to verify configuration

3. **Security (Cloudflare Turnstile):**
   - Turnstile CAPTCHA integration
   - Site key (public)
   - Secret key (private, validated before saving)
   - Automatic validation on key change to prevent lockouts

4. **OEM/Branding:**
   - Site name, logo, subtitle
   - API base URL
   - Contact information
   - Documentation URL

5. **Default User Settings:**
   - Default balance for new users
   - Default concurrency limit

6. **Admin API Key:**
   - `GetAdminApiKey()`: Get masked admin API key status
   - `RegenerateAdminApiKey()`: Generate new admin API key (shown once)
   - `DeleteAdminApiKey()`: Revoke admin API key

**Security Features:**
- Passwords are never returned in GET requests (only `*_configured` flags)
- Turnstile secret key is validated before accepting changes
- Audit logging for all settings changes (logs changed fields)

---

#### 2.1.6 ProxyHandler (`proxy_handler.go`)

**Purpose:** Manages HTTP/SOCKS proxies used by AI accounts for geo-restriction bypass and IP rotation.

**Core Operations:**

1. **CRUD Operations:**
   - `List()`: Paginated proxy listing with filters (protocol, status, search)
   - `GetAll()`: Get all active proxies (optionally with account counts)
   - `GetByID()`: Retrieve proxy details
   - `Create()`: Add new proxy
   - `Update()`: Update proxy configuration
   - `Delete()`: Remove proxy

2. **Proxy Management:**
   - Protocols: http, https, socks5, socks5h
   - Authentication (username/password)
   - `Test()`: Test proxy connectivity
   - `GetStats()`: Proxy usage statistics
   - `GetProxyAccounts()`: List accounts using this proxy

3. **Batch Operations:**
   - `BatchCreate()`: Bulk import proxies with duplicate detection

**Data Structures:**
```go
type CreateProxyRequest struct {
    Name     string
    Protocol string // http|https|socks5|socks5h
    Host     string
    Port     int    // 1-65535
    Username string
    Password string
}
```

---

#### 2.1.7 SubscriptionHandler (`subscription_handler.go`)

**Purpose:** Manages user subscriptions to exclusive groups with validity periods and usage tracking.

**Core Operations:**

1. **Subscription Management:**
   - `List()`: Paginated subscriptions with filters (user_id, group_id, status)
   - `GetByID()`: Get subscription details
   - `Assign()`: Assign subscription to user (admin action)
   - `BulkAssign()`: Assign subscription to multiple users
   - `Extend()`: Extend subscription validity
   - `Revoke()`: Cancel subscription

2. **Usage Tracking:**
   - `GetProgress()`: Get subscription usage progress vs limits

3. **Related Endpoints:**
   - `ListByGroup()`: Get all subscriptions for a group
   - `ListByUser()`: Get all subscriptions for a user

**Data Structures:**
```go
type AssignSubscriptionRequest struct {
    UserID       int64
    GroupID      int64
    ValidityDays int    // Max 36500 (100 years)
    Notes        string
}

type BulkAssignSubscriptionRequest struct {
    UserIDs      []int64
    GroupID      int64
    ValidityDays int
    Notes        string
}
```

---

#### 2.1.8 RedeemHandler (`redeem_handler.go`)

**Purpose:** Generate and manage redeem codes for balance, concurrency, or subscription giveaways.

**Core Operations:**

1. **Code Management:**
   - `List()`: Paginated codes with filters (type, status, search)
   - `GetByID()`: Get code details
   - `Generate()`: Generate N codes at once
   - `Delete()`: Delete single code
   - `BatchDelete()`: Delete multiple codes
   - `Expire()`: Mark code as expired

2. **Code Types:**
   - `balance`: Add credits to user balance
   - `concurrency`: Increase user concurrency limit
   - `subscription`: Grant subscription to a group

3. **Export & Statistics:**
   - `Export()`: Export codes to CSV
   - `GetStats()`: Code usage statistics (total, active, used, expired)

**Data Structures:**
```go
type GenerateRedeemCodesRequest struct {
    Count        int     // 1-100 per batch
    Type         string  // balance|concurrency|subscription
    Value        float64
    GroupID      *int64  // Required for subscription type
    ValidityDays int     // For subscription codes, max 36500
}
```

---

#### 2.1.9 UsageHandler (`usage_handler.go`)

**Purpose:** Query and analyze detailed usage logs with advanced filtering capabilities.

**Core Operations:**

1. **Usage Logs:**
   - `List()`: Paginated usage records with comprehensive filters:
     - user_id, api_key_id, account_id, group_id
     - model (model name filter)
     - stream (true/false)
     - billing_type (0=standard, 1=subscription)
     - start_date, end_date (with timezone support)

2. **Statistics:**
   - `Stats()`: Aggregate statistics with same filters as List
     - Total requests, tokens, cost
     - Grouped by time period (today, week, month, custom range)

3. **Search Helpers:**
   - `SearchUsers()`: Search users by email keyword (autocomplete)
   - `SearchAPIKeys()`: Search API keys by user or keyword

**Advanced Features:**
- Timezone-aware date filtering
- Complex multi-dimensional filtering
- CSV export support via parent handlers

---

#### 2.1.10 PaymentOrdersHandler (`payment_orders_handler.go`)

**Purpose:** View and export payment orders from various payment providers (zpay, stripe, admin, activity).

**Core Operations:**

1. **Order Management:**
   - `List()`: Paginated orders with filters:
     - provider (zpay, stripe, admin, activity)
     - status (pending, paid, failed)
     - user_email
     - date range (from, to)

2. **Export:**
   - `Export()`: CSV export of filtered orders with user email resolution

**Features:**
- Automatic user email resolution from user IDs
- Order type classification (online_recharge, admin_recharge, activity_recharge)
- CSV sanitization for security

---

#### 2.1.11 OAuth Handlers

**Three separate OAuth handlers for different AI platforms:**

1. **OpenAIOAuthHandler (`openai_oauth_handler.go`):**
   - `GenerateAuthURL()`: Generate OpenAI OAuth URL
   - `ExchangeCode()`: Exchange code for OpenAI tokens
   - `RefreshToken()`: Refresh expired OpenAI tokens
   - `RefreshAccountToken()`: Refresh specific account's token
   - `CreateAccountFromOAuth()`: Create account from OAuth tokens

2. **GeminiOAuthHandler (`gemini_oauth_handler.go`):**
   - `GenerateAuthURL()`: Generate Google OAuth URL for Gemini
   - `ExchangeCode()`: Exchange code for Google tokens
   - `GetCapabilities()`: Get Gemini API capabilities

3. **AntigravityOAuthHandler (`antigravity_oauth_handler.go`):**
   - `GenerateAuthURL()`: Generate Antigravity OAuth URL
   - `ExchangeCode()`: Exchange code for Antigravity tokens

**Common OAuth Flow:**
```
1. Frontend → GenerateAuthURL() → Get auth_url + session_id
2. User authorizes in popup → Get authorization code
3. Frontend → ExchangeCode(session_id, code) → Get tokens
4. Frontend → CreateAccount(tokens) → Account created
```

---

#### 2.1.12 SystemHandler (`system_handler.go`)

**Purpose:** System updates, version management, and service control.

**Operations:**
- `GetVersion()`: Get current system version
- `CheckUpdates()`: Check for available updates
- `PerformUpdate()`: Execute system update
- `Rollback()`: Rollback to previous version
- `RestartService()`: Restart the service

---

#### 2.1.13 UserAttributeHandler (`user_attribute_handler.go`)

**Purpose:** Manage custom user attributes for extended user profiling.

**Operations:**
- `ListDefinitions()`: List all attribute definitions
- `CreateDefinition()`: Create new attribute
- `UpdateDefinition()`: Update attribute definition
- `DeleteDefinition()`: Remove attribute
- `ReorderDefinitions()`: Change display order
- `GetUserAttributes()`: Get attribute values for specific user
- `UpdateUserAttributes()`: Update user's attribute values
- `GetBatchUserAttributes()`: Batch fetch attributes for multiple users

---

### 2.2 Service Layer Interactions

The handlers delegate business logic to service layer:

```
AccountHandler
    ├─→ AdminService (CRUD operations)
    ├─→ OAuthService (Claude OAuth)
    ├─→ OpenAIOAuthService (OpenAI OAuth)
    ├─→ GeminiOAuthService (Gemini OAuth)
    ├─→ AntigravityOAuthService (Antigravity OAuth)
    ├─→ RateLimitService (rate limit management)
    ├─→ AccountUsageService (usage statistics)
    ├─→ AccountTestService (connectivity testing)
    ├─→ ConcurrencyService (real-time concurrency)
    └─→ CRSSyncService (CRS synchronization)

UserHandler
    └─→ AdminService (user CRUD, balance operations)

DashboardHandler
    └─→ DashboardService (statistics aggregation)

SubscriptionHandler
    └─→ SubscriptionService (subscription lifecycle)

SettingHandler
    ├─→ SettingService (settings persistence)
    ├─→ EmailService (SMTP operations)
    └─→ TurnstileService (CAPTCHA validation)

UsageHandler
    ├─→ UsageService (usage log queries)
    ├─→ APIKeyService (API key search)
    └─→ AdminService (user search)
```

---

### 2.3 Authentication & Authorization

**Middleware Stack:**
```go
admin := v1.Group("/admin")
admin.Use(gin.HandlerFunc(adminAuth))
```

**AdminAuthMiddleware** validates:
1. JWT token presence and validity
2. User role = "admin"
3. Account status = "active"

**Protection Level:**
- All `/api/v1/admin/*` endpoints require admin authentication
- No public admin endpoints
- Session-based authentication via JWT stored in cookies or headers

---

## 3. Frontend Analysis

### 3.1 View Components

#### 3.1.1 DashboardView.vue

**Purpose:** Main admin landing page with real-time system metrics and charts.

**Key Features:**
- Overview cards (users, API keys, accounts, usage statistics)
- Usage trend charts (daily/hourly granularity)
- Model distribution pie chart
- Top API keys and users rankings
- Real-time updates (RPM, TPM)

**Data Sources:**
- `/admin/dashboard/stats` - Overall statistics
- `/admin/dashboard/trend` - Time-series data
- `/admin/dashboard/models` - Model breakdown
- `/admin/dashboard/api-keys-trend` - Top API keys
- `/admin/dashboard/users-trend` - Top users

---

#### 3.1.2 UsersView.vue

**Purpose:** Complete user management interface with advanced filtering.

**Key Features:**
- User table with sorting and pagination
- Advanced filters:
  - Status (active/disabled)
  - Role (admin/user)
  - Search (email/username)
  - Custom attributes (dynamic filters)
- User actions:
  - Create user modal
  - Edit user (balance, concurrency, allowed groups)
  - Delete user
  - View API keys
  - View subscriptions
  - View usage statistics
- CSV export functionality
- Batch operations (status toggle)

**Components:**
- `UserCreateModal.vue` - User creation form
- `UserEditModal.vue` - User editing form
- `UserBalanceModal.vue` - Balance adjustment
- `UserAllowedGroupsModal.vue` - Group access management
- `UserApiKeysModal.vue` - API key listing

---

#### 3.1.3 AccountsView.vue

**Purpose:** AI platform account management with OAuth integration.

**Key Features:**
- Account table with real-time concurrency counts
- Platform filters (Anthropic, OpenAI, Gemini, Antigravity)
- Type filters (OAuth, Setup-Token, API Key)
- Status filters (Active, Inactive, Error, Rate-Limited)
- Account actions:
  - Create account (OAuth wizard or manual)
  - Edit credentials, proxy, concurrency, priority, groups
  - Test connectivity (SSE streaming results)
  - Refresh credentials (OAuth token refresh)
  - View statistics (requests, tokens, cost charts)
  - Clear error/rate limit status
  - Toggle schedulable status
  - Batch operations (bulk update, delete)
- OAuth flows:
  - Generate auth URL → Open popup → Exchange code
  - Cookie-based auto-auth (sessionKey)
- CRS sync (import accounts from claude-relay-service)
- Google One tier refresh (for Gemini accounts)

**Components:**
- `AccountActionMenu.vue` - Per-account action menu
- `AccountBulkActionsBar.vue` - Batch operation toolbar
- `AccountStatsModal.vue` - Usage statistics with charts
- `AccountTestModal.vue` - Connectivity test results (SSE)
- `ReAuthAccountModal.vue` - OAuth re-authentication
- `AccountTableActions.vue` - Table action buttons
- `AccountTableFilters.vue` - Advanced filtering

---

#### 3.1.4 GroupsView.vue

**Purpose:** API key group configuration for pricing and access control.

**Key Features:**
- Group table with platform icons
- Create/edit group modal:
  - Platform selection
  - Rate multiplier (pricing)
  - Exclusive mode toggle
  - Subscription type
  - Usage limits (daily, weekly, monthly)
  - Image generation pricing (Gemini/Antigravity)
- Group actions:
  - View API keys in group
  - View subscriptions
  - Toggle status
  - Delete group
- Platform-specific configuration (image pricing for Gemini/Antigravity)

---

#### 3.1.5 SettingsView.vue

**Purpose:** System-wide configuration panel with live testing.

**Tabs:**
1. **Registration Settings:**
   - Enable/disable registration
   - Email verification toggle
   - Default balance and concurrency

2. **Email Service (SMTP):**
   - SMTP server configuration
   - Test connection button (instant feedback)
   - Send test email (to verify end-to-end)

3. **Security (Turnstile):**
   - Cloudflare Turnstile integration
   - Site key and secret key
   - Live validation on save

4. **OEM/Branding:**
   - Site name, logo, subtitle
   - API base URL
   - Contact info and documentation URL

5. **Admin API Key:**
   - View masked key
   - Regenerate (shows full key once)
   - Delete key

**Validation Features:**
- Real-time SMTP connection testing
- Turnstile secret key validation before saving
- Password fields support "leave blank to keep current"

---

#### 3.1.6 ProxiesView.vue

**Purpose:** HTTP/SOCKS proxy management for account geo-unblocking.

**Key Features:**
- Proxy table with protocol, host, port
- Create/edit proxy modal
- Batch import from text/CSV
- Test connectivity
- View accounts using proxy
- Duplicate detection on batch import

---

#### 3.1.7 SubscriptionsView.vue

**Purpose:** User subscription management for exclusive groups.

**Key Features:**
- Subscription table with user/group info
- Assign subscription modal (single or bulk)
- Extend validity
- Revoke subscription
- View usage progress (vs limits)
- Filters: user, group, status

---

#### 3.1.8 RedeemView.vue

**Purpose:** Redeem code generation and management.

**Key Features:**
- Code table with type, value, status
- Generate codes modal:
  - Type selection (balance, concurrency, subscription)
  - Quantity (1-100)
  - Value/validity
  - Group selection (for subscriptions)
- Export to CSV
- Batch delete
- Expire code
- Statistics dashboard

---

#### 3.1.9 UsageView.vue

**Purpose:** Advanced usage log viewer with multi-dimensional filtering.

**Key Features:**
- Usage log table with detailed token breakdown
- Comprehensive filters:
  - User (autocomplete search)
  - API key (autocomplete search)
  - Account (dropdown)
  - Group (dropdown)
  - Model (text input)
  - Stream (true/false)
  - Billing type (standard/subscription)
  - Date range (with timezone awareness)
- Statistics cards (total requests, tokens, cost)
- Real-time filter application
- Export to CSV

**Components:**
- `UsageFilters.vue` - Advanced filter panel
- `UsageStatsCards.vue` - Summary statistics
- `UsageTable.vue` - Usage log table
- `UsageExportProgress.vue` - CSV export progress

---

#### 3.1.10 PaymentOrdersView.vue

**Purpose:** Payment order history and export.

**Key Features:**
- Order table with user email, amount, status
- Filters: provider, user email, status, date range
- Order type classification (online/admin/activity)
- Export to CSV

---

### 3.2 API Client Layer

All frontend API calls go through TypeScript API client modules in `/frontend/src/api/admin/`:

**API Client Pattern:**
```typescript
import { apiClient } from '../client'

export async function list(
  page: number = 1,
  pageSize: number = 20,
  filters?: {...}
): Promise<PaginatedResponse<T>> {
  const { data } = await apiClient.get<PaginatedResponse<T>>('/admin/endpoint', {
    params: { page, page_size: pageSize, ...filters }
  })
  return data
}
```

**Key Modules:**
- `accounts.ts` - 20+ account management functions
- `users.ts` - User CRUD, balance, API keys
- `dashboard.ts` - Statistics and trend queries
- `groups.ts` - Group CRUD and configuration
- `settings.ts` - System settings, SMTP, Turnstile
- `proxies.ts` - Proxy management
- `subscriptions.ts` - Subscription operations
- `redeem.ts` - Redeem code generation
- `usage.ts` - Usage log queries
- `paymentOrders.ts` - Payment order listing

**Common Patterns:**
- Pagination support (page, page_size)
- Filter serialization (query parameters)
- TypeScript type safety
- Error handling via axios interceptors
- AbortSignal support for request cancellation

---

## 4. Feature Mapping

### 4.1 Backend → Frontend Mapping

| Feature | Backend Handler | Backend Endpoint | Frontend View | Frontend Components |
|---------|----------------|------------------|---------------|---------------------|
| **Dashboard** | DashboardHandler | GET /admin/dashboard/stats | DashboardView.vue | StatCards, TrendChart, ModelChart |
| **User Management** | UserHandler | GET/POST/PUT/DELETE /admin/users | UsersView.vue | UserCreateModal, UserEditModal, UserBalanceModal |
| **Account Management** | AccountHandler | GET/POST/PUT/DELETE /admin/accounts | AccountsView.vue | AccountActionMenu, AccountStatsModal, AccountTestModal |
| **OAuth Flows** | OAuthHandler, OpenAIOAuthHandler, etc. | POST /admin/accounts/generate-auth-url | AccountsView.vue | OAuth wizard modals |
| **Group Configuration** | GroupHandler | GET/POST/PUT/DELETE /admin/groups | GroupsView.vue | GroupEditModal |
| **System Settings** | SettingHandler | GET/PUT /admin/settings | SettingsView.vue | SMTP test, Turnstile config |
| **Proxy Management** | ProxyHandler | GET/POST/PUT/DELETE /admin/proxies | ProxiesView.vue | ProxyEditModal |
| **Subscriptions** | SubscriptionHandler | GET/POST/DELETE /admin/subscriptions | SubscriptionsView.vue | AssignModal, ExtendModal |
| **Redeem Codes** | RedeemHandler | POST /admin/redeem-codes/generate | RedeemView.vue | GenerateModal |
| **Usage Logs** | UsageHandler | GET /admin/usage | UsageView.vue | UsageFilters, UsageTable |
| **Payment Orders** | PaymentOrdersHandler | GET /admin/payment/orders | PaymentOrdersView.vue | OrderTable |

### 4.2 Data Flow Examples

#### Example 1: Account Creation with OAuth

```
1. User clicks "Create Account" → Opens modal
2. User selects "Claude OAuth" → Frontend calls:
   POST /admin/accounts/generate-auth-url { proxy_id: 1 }
3. Backend returns:
   { auth_url: "https://claude.ai/oauth/authorize?...", session_id: "abc123" }
4. Frontend opens popup with auth_url
5. User authorizes → Popup URL contains code=xyz789
6. Frontend calls:
   POST /admin/accounts/exchange-code { session_id: "abc123", code: "xyz789" }
7. Backend exchanges code for tokens, returns:
   { access_token, refresh_token, expires_at, ... }
8. Frontend calls:
   POST /admin/accounts {
     name: "My Claude Account",
     platform: "anthropic",
     type: "oauth",
     credentials: { access_token, refresh_token, ... },
     proxy_id: 1,
     group_ids: [1, 2]
   }
9. Backend creates account, returns account object
10. Frontend closes modal, refreshes account list
```

#### Example 2: Usage Statistics Query

```
1. Admin opens Usage page
2. Admin applies filters:
   - User: john@example.com (autocomplete)
   - Date range: 2026-01-01 to 2026-01-07
   - Model: claude-sonnet-4
3. Frontend calls:
   GET /admin/usage?user_id=123&start_date=2026-01-01&end_date=2026-01-07&model=claude-sonnet-4&page=1&page_size=20
4. Backend queries usage_logs table with filters, returns:
   {
     data: [{ id, user_id, api_key_id, model, tokens, cost, ... }, ...],
     pagination: { total, page, pages, ... }
   }
5. Frontend displays table with usage records
6. Admin clicks "Export CSV"
7. Frontend calls same endpoint with responseType: 'blob'
8. Backend generates CSV, returns as download
9. Browser prompts "Save As" dialog
```

#### Example 3: Batch Account Update

```
1. Admin selects 5 accounts in AccountsView
2. Admin clicks "Bulk Edit" → Opens modal
3. Admin updates: priority = 10, concurrency = 3
4. Frontend calls:
   POST /admin/accounts/bulk-update {
     account_ids: [1, 2, 3, 4, 5],
     priority: 10,
     concurrency: 3
   }
5. Backend iterates accounts, updates each:
   - Validates changes
   - Updates database
   - Returns results: { success: 5, failed: 0, results: [...] }
6. Frontend shows toast: "Updated 5 accounts"
7. Frontend refreshes account list
```

---

## 5. Data Flow Analysis

### 5.1 Typical Request Flow

```
┌──────────────┐
│   Browser    │
│  (Vue.js)    │
└──────┬───────┘
       │ 1. User Action (e.g., "Create User")
       │
       ↓
┌──────────────────────┐
│   API Client         │
│  (TypeScript)        │
│  users.ts:create()   │
└──────┬───────────────┘
       │ 2. HTTP POST /api/v1/admin/users
       │    { email, password, balance, ... }
       │
       ↓
┌──────────────────────────────────────┐
│   Backend Router (Gin)               │
│   + Middleware Stack:                │
│     - Logger                         │
│     - CORS                           │
│     - Security Headers               │
│     - JWT Auth                       │
│     - Admin Role Check               │
└──────┬───────────────────────────────┘
       │ 3. Route to handler
       │
       ↓
┌──────────────────────────────────────┐
│   UserHandler.Create()               │
│   - Validate request                 │
│   - Call service layer               │
└──────┬───────────────────────────────┘
       │ 4. Business logic
       │
       ↓
┌──────────────────────────────────────┐
│   AdminService.CreateUser()          │
│   - Hash password                    │
│   - Set default values               │
│   - Call repository                  │
└──────┬───────────────────────────────┘
       │ 5. Database operation
       │
       ↓
┌──────────────────────────────────────┐
│   UserRepository                     │
│   - INSERT INTO users ...            │
│   - Return user entity               │
└──────┬───────────────────────────────┘
       │ 6. Return user
       │
       ↓
┌──────────────────────────────────────┐
│   UserHandler.Create()               │
│   - Convert to DTO                   │
│   - Return JSON response             │
└──────┬───────────────────────────────┘
       │ 7. HTTP 200 OK
       │    { id, email, balance, ... }
       │
       ↓
┌──────────────────────┐
│   API Client         │
│   - Return typed data│
└──────┬───────────────┘
       │ 8. Update Vue state
       │
       ↓
┌──────────────┐
│   Browser    │
│   - Show toast: "User created" │
│   - Refresh user list │
└──────────────┘
```

### 5.2 OAuth Flow Sequence

```
User Browser                Frontend                Backend (Sub2API)         Claude OAuth Server
     │                          │                           │                         │
     │ 1. Click "Add Claude"    │                           │                         │
     │─────────────────────────>│                           │                         │
     │                          │ 2. POST /admin/accounts/  │                         │
     │                          │    generate-auth-url      │                         │
     │                          │──────────────────────────>│                         │
     │                          │                           │ 3. Build OAuth URL      │
     │                          │                           │    with callback        │
     │                          │<──────────────────────────│                         │
     │                          │ { auth_url, session_id }  │                         │
     │                          │                           │                         │
     │ 4. Open popup with       │                           │                         │
     │    auth_url              │                           │                         │
     │─────────────────────────────────────────────────────────────────────────────>│
     │                          │                           │                         │
     │                          │                           │                         │ 5. User authorizes
     │<─────────────────────────────────────────────────────────────────────────────│
     │ 6. Redirect to callback  │                           │                         │
     │    ?code=xyz789          │                           │                         │
     │─────────────────────────>│                           │                         │
     │                          │ 7. POST /admin/accounts/  │                         │
     │                          │    exchange-code          │                         │
     │                          │    { session_id, code }   │                         │
     │                          │──────────────────────────>│                         │
     │                          │                           │ 8. POST /oauth/token    │
     │                          │                           │    { code, ... }        │
     │                          │                           │────────────────────────>│
     │                          │                           │<────────────────────────│
     │                          │                           │ { access_token, ... }   │
     │                          │<──────────────────────────│                         │
     │                          │ { access_token,           │                         │
     │                          │   refresh_token, ... }    │                         │
     │                          │                           │                         │
     │ 9. Show "Save Account"   │                           │                         │
     │    modal with token info │                           │                         │
     │<─────────────────────────│                           │                         │
     │                          │                           │                         │
     │ 10. Click "Save"         │                           │                         │
     │─────────────────────────>│                           │                         │
     │                          │ 11. POST /admin/accounts  │                         │
     │                          │     { credentials: {...}} │                         │
     │                          │──────────────────────────>│                         │
     │                          │                           │ 12. INSERT account      │
     │                          │<──────────────────────────│                         │
     │                          │ { id, name, status, ... } │                         │
     │<─────────────────────────│                           │                         │
     │ "Account created!"       │                           │                         │
```

### 5.3 Real-time Concurrency Tracking

```
┌─────────────────────────────────────────────────────────────────┐
│  Request Flow (API Gateway)                                     │
│  ┌───────────────────────────────────────────────────────────┐ │
│  │ 1. User request → Acquire account from pool                │ │
│  │ 2. ConcurrencyService.Acquire(accountID)                   │ │
│  │    → Redis INCR account:{id}:concurrency                   │ │
│  │ 3. Process request                                         │ │
│  │ 4. ConcurrencyService.Release(accountID)                   │ │
│  │    → Redis DECR account:{id}:concurrency                   │ │
│  └───────────────────────────────────────────────────────────┘ │
└─────────────────────────────────────────────────────────────────┘
                            ↓
┌─────────────────────────────────────────────────────────────────┐
│  Admin Dashboard Display                                        │
│  ┌───────────────────────────────────────────────────────────┐ │
│  │ 1. GET /admin/accounts → AccountHandler.List()            │ │
│  │ 2. Fetch account IDs: [1, 2, 3, ...]                      │ │
│  │ 3. ConcurrencyService.GetAccountConcurrencyBatch(ids)     │ │
│  │    → Redis MGET account:1:concurrency,                    │ │
│  │                account:2:concurrency, ...                  │ │
│  │ 4. Build response:                                         │ │
│  │    [                                                       │ │
│  │      { id: 1, name: "Account 1",                          │ │
│  │        current_concurrency: 5, max_concurrency: 10 },     │ │
│  │      { id: 2, name: "Account 2",                          │ │
│  │        current_concurrency: 2, max_concurrency: 5 },      │ │
│  │      ...                                                   │ │
│  │    ]                                                       │ │
│  └───────────────────────────────────────────────────────────┘ │
└─────────────────────────────────────────────────────────────────┘
```

---

## 6. API Endpoint Documentation

### 6.1 Dashboard Endpoints

| Method | Path | Description | Query Params | Response |
|--------|------|-------------|--------------|----------|
| GET | `/admin/dashboard/stats` | Get overall system statistics | - | Dashboard stats (users, keys, accounts, tokens, cost, performance) |
| GET | `/admin/dashboard/realtime` | Get real-time metrics | - | Active requests, RPM, avg response time, error rate |
| GET | `/admin/dashboard/trend` | Get usage trend | start_date, end_date, granularity, user_id, api_key_id | Time-series data points |
| GET | `/admin/dashboard/models` | Get model statistics | start_date, end_date, user_id, api_key_id | Per-model usage breakdown |
| GET | `/admin/dashboard/api-keys-trend` | Get top API key usage | start_date, end_date, granularity, limit | Top N API keys with trends |
| GET | `/admin/dashboard/users-trend` | Get top user usage | start_date, end_date, granularity, limit | Top N users with trends |
| POST | `/admin/dashboard/users-usage` | Batch user usage stats | { user_ids } | Map of user_id → stats |
| POST | `/admin/dashboard/api-keys-usage` | Batch API key usage stats | { api_key_ids } | Map of api_key_id → stats |

### 6.2 User Management Endpoints

| Method | Path | Description | Query Params | Request Body | Response |
|--------|------|-------------|--------------|--------------|----------|
| GET | `/admin/users` | List users | page, page_size, status, role, search, attr[id] | - | Paginated user list |
| GET | `/admin/users/export` | Export users CSV | status, role, search, attr[id] | - | CSV file |
| GET | `/admin/users/:id` | Get user by ID | - | - | User object |
| POST | `/admin/users` | Create user | - | { email, password, balance, concurrency, allowed_groups } | Created user |
| PUT | `/admin/users/:id` | Update user | - | { email, username, balance, concurrency, status, allowed_groups } | Updated user |
| DELETE | `/admin/users/:id` | Delete user | - | - | Success message |
| POST | `/admin/users/:id/balance` | Update balance | - | { balance, operation, notes } | Updated user |
| GET | `/admin/users/:id/api-keys` | Get user's API keys | page, page_size | - | Paginated API keys |
| GET | `/admin/users/:id/usage` | Get user usage | period | - | Usage statistics |
| GET | `/admin/users/:id/attributes` | Get user attributes | - | - | Attribute values |
| PUT | `/admin/users/:id/attributes` | Update user attributes | - | { attributes } | Updated attributes |

### 6.3 Account Management Endpoints

| Method | Path | Description | Request Body | Response |
|--------|------|-------------|--------------|----------|
| GET | `/admin/accounts` | List accounts (with concurrency) | - | Paginated accounts |
| GET | `/admin/accounts/:id` | Get account by ID | - | Account object |
| POST | `/admin/accounts` | Create account | { name, platform, type, credentials, proxy_id, concurrency, priority, group_ids } | Created account |
| PUT | `/admin/accounts/:id` | Update account | { name, credentials, proxy_id, concurrency, priority, status, group_ids } | Updated account |
| DELETE | `/admin/accounts/:id` | Delete account | - | Success message |
| POST | `/admin/accounts/:id/test` | Test connectivity (SSE) | { model_id } | SSE stream of test results |
| POST | `/admin/accounts/:id/refresh` | Refresh OAuth tokens | - | Updated account |
| GET | `/admin/accounts/:id/stats` | Get usage statistics | ?days=30 | Statistics with history |
| POST | `/admin/accounts/:id/clear-error` | Clear error status | - | Updated account |
| GET | `/admin/accounts/:id/usage` | Get 5h/7d usage windows | - | Usage info |
| GET | `/admin/accounts/:id/today-stats` | Get today's stats | - | Today's statistics |
| POST | `/admin/accounts/:id/clear-rate-limit` | Clear rate limit | - | Success message |
| GET | `/admin/accounts/:id/temp-unschedulable` | Get temp unschedulable status | - | Status object |
| DELETE | `/admin/accounts/:id/temp-unschedulable` | Clear temp unschedulable | - | Success message |
| POST | `/admin/accounts/:id/schedulable` | Set schedulable | { schedulable } | Updated account |
| GET | `/admin/accounts/:id/models` | Get available models | - | Model array |
| POST | `/admin/accounts/batch` | Batch create accounts | { accounts } | Batch results |
| POST | `/admin/accounts/batch-update-credentials` | Batch update credentials | { account_ids, field, value } | Batch results |
| POST | `/admin/accounts/bulk-update` | Bulk update accounts | { account_ids, ...updates } | Batch results |
| POST | `/admin/accounts/sync/crs` | Sync from CRS | { base_url, username, password, sync_proxies } | Sync results |
| POST | `/admin/accounts/:id/refresh-tier` | Refresh Google One tier | - | Tier info |
| POST | `/admin/accounts/batch-refresh-tier` | Batch refresh tier | { account_ids } | Batch results |

### 6.4 OAuth Endpoints (Claude)

| Method | Path | Description | Request Body | Response |
|--------|------|-------------|--------------|----------|
| POST | `/admin/accounts/generate-auth-url` | Generate OAuth URL (full scope) | { proxy_id } | { auth_url, session_id } |
| POST | `/admin/accounts/generate-setup-token-url` | Generate OAuth URL (inference-only) | { proxy_id } | { auth_url, session_id } |
| POST | `/admin/accounts/exchange-code` | Exchange code for tokens | { session_id, code, proxy_id } | Token info |
| POST | `/admin/accounts/exchange-setup-token-code` | Exchange setup token code | { session_id, code, proxy_id } | Token info |
| POST | `/admin/accounts/cookie-auth` | Cookie-based auth (full scope) | { code, proxy_id } | Token info |
| POST | `/admin/accounts/setup-token-cookie-auth` | Cookie-based auth (setup token) | { code, proxy_id } | Token info |

### 6.5 Group Management Endpoints

| Method | Path | Description | Query Params | Request Body | Response |
|--------|------|-------------|--------------|--------------|----------|
| GET | `/admin/groups` | List groups | page, page_size, platform, status, is_exclusive | - | Paginated groups |
| GET | `/admin/groups/all` | Get all active groups | platform | - | Group array |
| GET | `/admin/groups/:id` | Get group by ID | - | - | Group object |
| POST | `/admin/groups` | Create group | - | { name, platform, rate_multiplier, is_exclusive, subscription_type, limits, image_prices } | Created group |
| PUT | `/admin/groups/:id` | Update group | - | { name, rate_multiplier, limits, image_prices, status } | Updated group |
| DELETE | `/admin/groups/:id` | Delete group | - | - | Success message |
| GET | `/admin/groups/:id/stats` | Get group stats | - | - | Statistics |
| GET | `/admin/groups/:id/api-keys` | Get group API keys | page, page_size | - | Paginated API keys |

### 6.6 Settings Endpoints

| Method | Path | Description | Request Body | Response |
|--------|------|-------------|--------------|----------|
| GET | `/admin/settings` | Get all settings | - | Settings object |
| PUT | `/admin/settings` | Update settings | { registration_enabled, email_verify_enabled, smtp_*, turnstile_*, oem_*, defaults } | Updated settings |
| POST | `/admin/settings/test-smtp` | Test SMTP connection | { smtp_host, smtp_port, smtp_username, smtp_password, smtp_use_tls } | Success/error |
| POST | `/admin/settings/send-test-email` | Send test email | { email, smtp_* } | Success/error |
| GET | `/admin/settings/admin-api-key` | Get admin API key status | - | { exists, masked_key } |
| POST | `/admin/settings/admin-api-key/regenerate` | Regenerate admin API key | - | { key } (full key, shown once) |
| DELETE | `/admin/settings/admin-api-key` | Delete admin API key | - | Success message |

### 6.7 Proxy Management Endpoints

| Method | Path | Description | Query Params | Request Body | Response |
|--------|------|-------------|--------------|--------------|----------|
| GET | `/admin/proxies` | List proxies | page, page_size, protocol, status, search | - | Paginated proxies |
| GET | `/admin/proxies/all` | Get all active proxies | with_count | - | Proxy array (optionally with account counts) |
| GET | `/admin/proxies/:id` | Get proxy by ID | - | - | Proxy object |
| POST | `/admin/proxies` | Create proxy | - | { name, protocol, host, port, username, password } | Created proxy |
| PUT | `/admin/proxies/:id` | Update proxy | - | { name, protocol, host, port, username, password, status } | Updated proxy |
| DELETE | `/admin/proxies/:id` | Delete proxy | - | - | Success message |
| POST | `/admin/proxies/:id/test` | Test proxy | - | - | Test result |
| GET | `/admin/proxies/:id/stats` | Get proxy stats | - | - | Statistics |
| GET | `/admin/proxies/:id/accounts` | Get proxy accounts | page, page_size | - | Paginated accounts |
| POST | `/admin/proxies/batch` | Batch create proxies | - | { proxies } | { created, skipped } |

### 6.8 Subscription Endpoints

| Method | Path | Description | Query Params | Request Body | Response |
|--------|------|-------------|--------------|--------------|----------|
| GET | `/admin/subscriptions` | List subscriptions | page, page_size, user_id, group_id, status | - | Paginated subscriptions |
| GET | `/admin/subscriptions/:id` | Get subscription by ID | - | - | Subscription object |
| GET | `/admin/subscriptions/:id/progress` | Get usage progress | - | - | Progress stats |
| POST | `/admin/subscriptions/assign` | Assign subscription | - | { user_id, group_id, validity_days, notes } | Created subscription |
| POST | `/admin/subscriptions/bulk-assign` | Bulk assign | - | { user_ids, group_id, validity_days, notes } | Bulk results |
| POST | `/admin/subscriptions/:id/extend` | Extend validity | - | { days } | Updated subscription |
| DELETE | `/admin/subscriptions/:id` | Revoke subscription | - | - | Success message |
| GET | `/admin/groups/:id/subscriptions` | List group subscriptions | page, page_size | - | Paginated subscriptions |
| GET | `/admin/users/:id/subscriptions` | List user subscriptions | - | - | Subscription array |

### 6.9 Redeem Code Endpoints

| Method | Path | Description | Query Params | Request Body | Response |
|--------|------|-------------|--------------|--------------|----------|
| GET | `/admin/redeem-codes` | List codes | page, page_size, type, status, search | - | Paginated codes |
| GET | `/admin/redeem-codes/:id` | Get code by ID | - | - | Code object |
| POST | `/admin/redeem-codes/generate` | Generate codes | - | { count, type, value, group_id, validity_days } | Generated codes |
| DELETE | `/admin/redeem-codes/:id` | Delete code | - | - | Success message |
| POST | `/admin/redeem-codes/batch-delete` | Batch delete | - | { ids } | { deleted } |
| POST | `/admin/redeem-codes/:id/expire` | Expire code | - | - | Updated code |
| GET | `/admin/redeem-codes/stats` | Get statistics | - | - | Code statistics |
| GET | `/admin/redeem-codes/export` | Export to CSV | type, status | - | CSV file |

### 6.10 Usage Log Endpoints

| Method | Path | Description | Query Params | Response |
|--------|------|-------------|--------------|----------|
| GET | `/admin/usage` | List usage logs | page, page_size, user_id, api_key_id, account_id, group_id, model, stream, billing_type, start_date, end_date, timezone | Paginated usage logs |
| GET | `/admin/usage/stats` | Get usage statistics | user_id, api_key_id, account_id, group_id, model, stream, billing_type, period, start_date, end_date, timezone | Aggregate statistics |
| GET | `/admin/usage/search-users` | Search users | q (keyword) | Simple user array |
| GET | `/admin/usage/search-api-keys` | Search API keys | user_id, q (keyword) | Simple API key array |

### 6.11 Payment Order Endpoints

| Method | Path | Description | Query Params | Response |
|--------|------|-------------|--------------|----------|
| GET | `/admin/payment/orders` | List payment orders | page, page_size, provider, user_email, status, from, to | Paginated orders |
| GET | `/admin/payment/orders/export` | Export orders CSV | provider, user_email, status, from, to | CSV file |

---

## 7. Security & Authorization

### 7.1 Authentication Flow

```
1. Admin logs in via /api/v1/auth/login
2. Backend validates credentials, checks role = "admin"
3. Backend generates JWT token with claims: { user_id, email, role }
4. Frontend stores JWT in localStorage/sessionStorage
5. For all admin API calls:
   - Frontend adds Authorization: Bearer <token> header
   - Backend middleware validates token
   - Backend middleware checks role = "admin"
   - Backend middleware checks account status = "active"
6. If validation fails → 401 Unauthorized
```

### 7.2 Authorization Levels

| Role | Access Level |
|------|--------------|
| **admin** | Full access to all `/api/v1/admin/*` endpoints |
| **user** | No access to admin endpoints (401 Unauthorized) |

### 7.3 Security Best Practices Observed

1. **Password Handling:**
   - Never return passwords in GET responses
   - Only show `*_configured` flags for sensitive fields
   - Hash passwords before storage (bcrypt)

2. **Token Security:**
   - OAuth tokens stored encrypted in database
   - Refresh tokens used for long-lived sessions
   - Token expiration enforced

3. **CSRF Protection:**
   - CORS middleware with whitelisted origins
   - SameSite cookie attribute
   - CSRF token validation for state-changing operations

4. **Input Validation:**
   - Gin binding validation (`binding:"required,email"`)
   - Custom validation for complex rules
   - SQL injection prevention (ORM parameterized queries)

5. **CSV Injection Prevention:**
   - `sanitizeCSVCell()` escapes dangerous characters
   - Prevents formula injection in Excel

6. **Rate Limiting:**
   - Account-level rate limit tracking
   - Temporary unschedulable mechanism
   - Admin can clear rate limits

7. **Audit Logging:**
   - Settings changes logged with user_id and changed fields
   - Critical operations tracked

---

## 8. Key Findings & Recommendations

### 8.1 Strengths

1. **Well-Structured Architecture:**
   - Clean separation of concerns (Handlers → Services → Repositories)
   - Modular design with focused handlers
   - Consistent RESTful API design

2. **Comprehensive Feature Set:**
   - Complete CRUD for all entities
   - Advanced filtering and search
   - Batch operations for efficiency
   - Real-time metrics and concurrency tracking

3. **Multi-Platform OAuth Support:**
   - Supports 4 AI platforms (Claude, OpenAI, Gemini, Antigravity)
   - Flexible OAuth flows (code exchange, cookie auth, setup token)
   - Automatic token refresh

4. **Strong Type Safety:**
   - TypeScript frontend API client
   - Go structs with validation tags
   - DTO pattern for data transfer

5. **User Experience:**
   - SSE streaming for real-time feedback (account testing)
   - Autocomplete search for users/API keys
   - CSV export for data portability
   - Timezone-aware date filtering

6. **Security:**
   - JWT authentication
   - Role-based access control
   - CSRF protection
   - CSV injection prevention
   - Audit logging

### 8.2 Areas for Improvement

1. **Error Handling Consistency:**
   - Some handlers use `response.ErrorFrom(c, err)`, others use `response.InternalError(c, msg)`
   - Recommendation: Standardize error response format across all handlers

2. **API Documentation:**
   - No OpenAPI/Swagger specification
   - Recommendation: Generate OpenAPI spec from code or add Swagger annotations

3. **Testing Coverage:**
   - No unit tests visible in handler files
   - Recommendation: Add handler unit tests with mocked services
   - Example:
     ```go
     func TestAccountHandler_List(t *testing.T) {
         // Mock AdminService
         // Call handler
         // Assert response
     }
     ```

4. **Pagination Standardization:**
   - Most endpoints use `page, page_size`
   - Some use `pagination.PaginationParams`, others use `response.ParsePagination()`
   - Recommendation: Standardize pagination parsing

5. **Batch Operation Error Reporting:**
   - Batch operations return `{ success, failed, results }` but frontend doesn't always show detailed errors
   - Recommendation: Improve error display in bulk operation modals

6. **Real-time Updates:**
   - Dashboard stats are polled, not pushed
   - Recommendation: Consider WebSocket for real-time dashboard updates

7. **Caching Strategy:**
   - No evidence of Redis caching for frequently accessed data
   - Recommendation: Cache system settings, group list, proxy list

8. **Database Query Optimization:**
   - Account list fetches concurrency counts in separate batch query (N+1 issue)
   - Recommendation: Use JOIN or CTE to fetch in single query

9. **Frontend Performance:**
   - Large tables may cause performance issues
   - Recommendation: Implement virtual scrolling for tables with 1000+ rows

10. **Monitoring & Alerting:**
    - No evidence of Prometheus metrics export
    - Recommendation: Add metrics endpoints for monitoring:
      - `/metrics` - Prometheus metrics
      - `/health` - Health check endpoint
      - `/ready` - Readiness probe

### 8.3 Security Recommendations

1. **API Rate Limiting:**
   - Add rate limiting to admin endpoints (e.g., 100 req/min per admin)
   - Prevent abuse from compromised admin accounts

2. **Two-Factor Authentication (2FA):**
   - Add optional 2FA for admin accounts
   - Use TOTP (Time-based One-Time Password)

3. **Session Management:**
   - Add "Active Sessions" view for admins
   - Allow revoking sessions remotely

4. **Sensitive Data Masking:**
   - Mask more sensitive fields in logs (email, IP address)
   - Add PII redaction to audit logs

5. **OAuth Token Storage:**
   - Ensure tokens are encrypted at rest
   - Rotate encryption keys periodically

### 8.4 Scalability Recommendations

1. **Database Indexing:**
   - Ensure indexes on frequently queried columns:
     - `users.email`, `users.status`, `users.role`
     - `accounts.platform`, `accounts.status`, `accounts.type`
     - `usage_logs.created_at`, `usage_logs.user_id`, `usage_logs.api_key_id`

2. **Query Optimization:**
   - Use database query analyzer to identify slow queries
   - Add covering indexes for common filter combinations

3. **Batch Processing:**
   - Use goroutines for parallel batch operations (already implemented in some handlers)
   - Add progress tracking for long-running batch jobs

4. **Export Performance:**
   - Stream large CSV exports instead of loading all data into memory
   - Add pagination to export endpoints (e.g., export in chunks of 10,000 records)

5. **Concurrency Control:**
   - Already using Redis for concurrency tracking (good!)
   - Consider adding Redis Cluster for high availability

### 8.5 Feature Recommendations

1. **Account Health Monitoring:**
   - Add automated account health checks (scheduled background jobs)
   - Alert admins when accounts go offline
   - Auto-refresh tokens before expiration

2. **Usage Anomaly Detection:**
   - Flag unusual usage patterns (sudden spikes, unexpected models)
   - Send email alerts to admins

3. **Advanced Analytics:**
   - Add cost forecasting (predict monthly costs)
   - Add user segmentation (high-value users, inactive users)
   - Add retention analysis

4. **Audit Trail:**
   - Add comprehensive audit log viewer in frontend
   - Track all admin actions (create, update, delete)
   - Export audit logs for compliance

5. **Backup & Restore:**
   - Add database backup scheduler
   - Add restore functionality from backup
   - Export/import account configurations

6. **API Key Templates:**
   - Add "API Key Templates" for quick creation
   - Pre-configure common settings (groups, limits, models)

7. **Scheduled Maintenance:**
   - Add maintenance mode toggle
   - Display banner to users during maintenance
   - Graceful shutdown of active requests

---

## Appendix

### A. File Structure Summary

**Backend:**
```
backend/internal/handler/admin/
├── account_handler.go           (1,225 lines) - AI account management
├── user_handler.go              (342 lines)   - User management
├── dashboard_handler.go         (305 lines)   - Statistics & metrics
├── group_handler.go             (260 lines)   - Group management
├── setting_handler.go           (469 lines)   - System settings
├── proxy_handler.go             (324 lines)   - Proxy management
├── subscription_handler.go      (279 lines)   - Subscription management
├── redeem_handler.go            (239 lines)   - Redeem code management
├── usage_handler.go             (347 lines)   - Usage logs
├── payment_orders_handler.go    (262 lines)   - Payment orders
├── openai_oauth_handler.go      (~300 lines)  - OpenAI OAuth
├── gemini_oauth_handler.go      (~250 lines)  - Gemini OAuth
├── antigravity_oauth_handler.go (~200 lines)  - Antigravity OAuth
├── system_handler.go            (~150 lines)  - System maintenance
├── user_attribute_handler.go    (~200 lines)  - Custom attributes
└── csv_sanitize.go              (~50 lines)   - CSV utilities

Total: ~5,000+ lines of backend code
```

**Frontend:**
```
frontend/src/views/admin/
├── DashboardView.vue
├── UsersView.vue
├── AccountsView.vue
├── GroupsView.vue
├── SettingsView.vue
├── ProxiesView.vue
├── SubscriptionsView.vue
├── RedeemView.vue
├── UsageView.vue
└── PaymentOrdersView.vue

frontend/src/components/admin/
├── account/ (7 components)
├── usage/ (4 components)
└── user/ (5 components)

frontend/src/api/admin/
├── accounts.ts (373 lines)
├── users.ts (209 lines)
├── dashboard.ts (200 lines)
├── groups.ts (169 lines)
├── settings.ts (175 lines)
└── [10 more modules]

Total: ~5,000+ lines of frontend code
```

### B. Technology Stack

**Backend:**
- Language: Go 1.21+
- Framework: Gin (HTTP router)
- Database: PostgreSQL/MySQL (via GORM)
- Cache: Redis (concurrency tracking)
- Authentication: JWT
- OAuth: Custom implementation (Claude, OpenAI, Gemini, Antigravity)

**Frontend:**
- Framework: Vue.js 3 (Composition API)
- Language: TypeScript
- UI Library: Element Plus / Custom components
- HTTP Client: Axios
- State Management: Pinia (assumed, not visible in read files)
- Charts: ECharts / Chart.js (assumed)
- Build Tool: Vite

**Infrastructure:**
- Deployment: Docker (docker-compose.yml)
- Reverse Proxy: Nginx (optional)
- Monitoring: Prometheus (recommended)
- Logging: Structured logging (JSON)

### C. Glossary

| Term | Definition |
|------|------------|
| **Account** | An AI platform account (Claude, OpenAI, Gemini, Antigravity) used to proxy user requests |
| **API Key** | A key issued to users for accessing the Sub2API gateway |
| **Group** | A collection of accounts with shared pricing and rate limits |
| **Subscription** | User's access to an exclusive group for a limited time |
| **OAuth** | Open Authorization protocol for third-party authentication |
| **Setup Token** | Claude OAuth token with inference-only scope (no org management) |
| **CRS** | Claude Relay Service - external service for account management |
| **Redeem Code** | One-time code for balance/concurrency/subscription |
| **Proxy** | HTTP/SOCKS proxy used by accounts for geo-unblocking |
| **Concurrency** | Maximum simultaneous requests allowed |
| **Rate Limit** | API rate limit imposed by AI platform |
| **Temp Unschedulable** | Temporary exclusion from account pool (due to rate limit or error) |
| **Schedulable** | Whether account participates in request scheduling |
| **OEM** | Original Equipment Manufacturer (branding customization) |
| **Turnstile** | Cloudflare CAPTCHA service |

---

## Document Control

| Version | Date | Author | Changes |
|---------|------|--------|---------|
| 1.0 | 2026-01-08 | Claude Code (Analysis Agent) | Initial comprehensive analysis |

**End of Document**
