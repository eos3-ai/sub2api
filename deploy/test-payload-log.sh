#!/usr/bin/env bash

set -euo pipefail

usage() {
  cat <<'EOF'
测试 Sub2API 正文采集：发起真实网关请求，并按 request_id 查询采集正文。

必填环境变量：
  ADMIN_API_KEY    管理员 API Key，格式为 admin-...
  GATEWAY_API_KEY  普通用户网关 API Key，通常为 sk-...
  MODEL            要测试的模型名称

可选环境变量：
  BASE_URL         Sub2API 地址，默认 http://172.16.13.87:8011
  PROTOCOL         openai 或 anthropic，默认 openai
  PROMPT           测试提示词，默认“请只回复：正文采集测试成功”
  MAX_ATTEMPTS     查询采集结果的最大次数，默认 15
  POLL_INTERVAL    每次查询的等待秒数，默认 1
  REQUEST_TIMEOUT  网关请求超时秒数，默认 300

示例：
  ADMIN_API_KEY='admin-...' \
  GATEWAY_API_KEY='sk-...' \
  MODEL='gpt-5.6-sol' \
  ./deploy/test-payload-log.sh

  PROTOCOL='anthropic' \
  ADMIN_API_KEY='admin-...' \
  GATEWAY_API_KEY='sk-...' \
  MODEL='claude-haiku-4-5-20251001' \
  ./deploy/test-payload-log.sh
EOF
}

die() {
  printf '错误：%s\n' "$*" >&2
  exit 1
}

require_command() {
  command -v "$1" >/dev/null 2>&1 || die "缺少命令：$1"
}

pretty_print_file() {
  local file="$1"
  if ! jq . "$file" 2>/dev/null; then
    cat "$file"
    printf '\n'
  fi
}

if [[ "${1:-}" == "-h" || "${1:-}" == "--help" ]]; then
  usage
  exit 0
fi

require_command curl
require_command jq

: "${ADMIN_API_KEY:?请设置 ADMIN_API_KEY=admin-...}"
: "${GATEWAY_API_KEY:?请设置 GATEWAY_API_KEY=sk-...}"
: "${MODEL:?请设置 MODEL=要测试的模型名称}"

BASE_URL="${BASE_URL:-http://172.16.13.87:8011}"
BASE_URL="${BASE_URL%/}"
PROTOCOL="${PROTOCOL:-openai}"
PROMPT="${PROMPT:-请只回复：正文采集测试成功}"
MAX_ATTEMPTS="${MAX_ATTEMPTS:-15}"
POLL_INTERVAL="${POLL_INTERVAL:-1}"
REQUEST_TIMEOUT="${REQUEST_TIMEOUT:-300}"

[[ "$ADMIN_API_KEY" == admin-* ]] || die "ADMIN_API_KEY 应以 admin- 开头"
[[ "$PROTOCOL" == "openai" || "$PROTOCOL" == "anthropic" ]] || \
  die "PROTOCOL 只能是 openai 或 anthropic"
[[ "$MAX_ATTEMPTS" =~ ^[1-9][0-9]*$ ]] || die "MAX_ATTEMPTS 必须是正整数"
[[ "$REQUEST_TIMEOUT" =~ ^[1-9][0-9]*$ ]] || die "REQUEST_TIMEOUT 必须是正整数"

task_tmp_dir="$(mktemp -d "${TMPDIR:-/tmp}/sub2api-payload-test.XXXXXX")"
trap 'rm -rf "$task_tmp_dir"' EXIT

settings_file="$task_tmp_dir/settings.json"
gateway_request_file="$task_tmp_dir/gateway-request.json"
gateway_response_file="$task_tmp_dir/gateway-response.txt"
gateway_headers_file="$task_tmp_dir/gateway-response-headers.txt"
payload_detail_file="$task_tmp_dir/payload-detail.json"

printf '1. 检查正文采集状态：%s\n' "$BASE_URL"
settings_http_code="$(
  curl -sS \
    --connect-timeout 10 \
    --max-time 30 \
    -o "$settings_file" \
    -w '%{http_code}' \
    -H "Authorization: Bearer ${ADMIN_API_KEY}" \
    "$BASE_URL/api/v1/admin/payload-logs/settings"
)"

if [[ "$settings_http_code" != "200" ]]; then
  pretty_print_file "$settings_file"
  die "采集状态接口返回 HTTP ${settings_http_code}"
fi

pretty_print_file "$settings_file"
jq -e '.code == 0' "$settings_file" >/dev/null || die "采集状态接口返回失败"
jq -e '.data.enabled == true' "$settings_file" >/dev/null || \
  die "运行时正文采集开关未开启"
jq -e '.data.store_available == true' "$settings_file" >/dev/null || \
  die "ClickHouse 正文存储不可用：$(jq -r '.data.store_error // "未知错误"' "$settings_file")"

printf '\n2. 发起真实网关请求\n'
printf '   protocol=%s model=%s\n' "$PROTOCOL" "$MODEL"

if [[ "$PROTOCOL" == "openai" ]]; then
  jq -n \
    --arg model "$MODEL" \
    --arg prompt "$PROMPT" \
    '{model: $model, messages: [{role: "user", content: $prompt}]}' \
    >"$gateway_request_file"

  gateway_http_code="$(
    curl -sS \
      --connect-timeout 10 \
      --max-time "$REQUEST_TIMEOUT" \
      -D "$gateway_headers_file" \
      -o "$gateway_response_file" \
      -w '%{http_code}' \
      -H "Authorization: Bearer ${GATEWAY_API_KEY}" \
      -H 'Content-Type: application/json' \
      --data-binary "@$gateway_request_file" \
      "$BASE_URL/v1/chat/completions"
  )"
else
  jq -n \
    --arg model "$MODEL" \
    --arg prompt "$PROMPT" \
    '{model: $model, max_tokens: 64, messages: [{role: "user", content: $prompt}]}' \
    >"$gateway_request_file"

  gateway_http_code="$(
    curl -sS \
      --connect-timeout 10 \
      --max-time "$REQUEST_TIMEOUT" \
      -D "$gateway_headers_file" \
      -o "$gateway_response_file" \
      -w '%{http_code}' \
      -H "x-api-key: ${GATEWAY_API_KEY}" \
      -H 'anthropic-version: 2023-06-01' \
      -H 'Content-Type: application/json' \
      --data-binary "@$gateway_request_file" \
      "$BASE_URL/v1/messages"
  )"
fi

request_id="$(
  awk '
    tolower($1) == "x-client-request-id:" {
      gsub(/\r/, "", $2)
      id = $2
    }
    END { print id }
  ' "$gateway_headers_file"
)"

printf '   网关响应 HTTP %s\n' "$gateway_http_code"
[[ -n "$request_id" ]] || {
  pretty_print_file "$gateway_response_file"
  die "网关响应中没有 X-Client-Request-ID，无法精确查询采集记录"
}
printf '   服务端采集 request_id=%s\n' "$request_id"
pretty_print_file "$gateway_response_file"

printf '\n3. 轮询采集详情（最长约 %s 次 × %s 秒）\n' \
  "$MAX_ATTEMPTS" "$POLL_INTERVAL"

found=false
for ((attempt = 1; attempt <= MAX_ATTEMPTS; attempt++)); do
  detail_http_code="$(
    curl -sS \
      --connect-timeout 10 \
      --max-time 30 \
      -o "$payload_detail_file" \
      -w '%{http_code}' \
      -H "Authorization: Bearer ${ADMIN_API_KEY}" \
      "$BASE_URL/api/v1/admin/payload-logs/$request_id"
  )"

  if [[ "$detail_http_code" == "200" ]] && \
    jq -e '.code == 0 and (.data | type == "array") and (.data | length > 0)' \
      "$payload_detail_file" >/dev/null 2>&1; then
    found=true
    printf '   第 %s 次查询已找到采集记录。\n' "$attempt"
    break
  fi

  if [[ "$detail_http_code" != "200" ]]; then
    printf '   第 %s 次查询返回 HTTP %s。\n' "$attempt" "$detail_http_code" >&2
  else
    printf '   第 %s 次查询暂未入库。\n' "$attempt"
  fi
  sleep "$POLL_INTERVAL"
done

if [[ "$found" != "true" ]]; then
  printf '\n最后一次详情接口响应：\n' >&2
  pretty_print_file "$payload_detail_file" >&2
  die "在等待时间内未找到 request_id=${request_id} 的采集记录"
fi

printf '\n4. 采集到的完整正文\n'
jq '
  def parse_body:
    if type == "string" then
      try fromjson catch .
    else
      .
    end;
  .data[] |
  {
    id,
    request_id,
    created_at,
    username,
    api_key_id,
    api_key_name,
    account_id,
    model_name,
    method,
    path,
    status_code,
    is_stream,
    request_bytes,
    response_bytes,
    request_body: (.request_body | parse_body),
    response_body: (.response_body | parse_body)
  }
' "$payload_detail_file"

printf '\n测试通过：网关请求已发起，且正文采集记录可通过管理员 API Key 查询。\n'
printf 'request_id=%s gateway_http_code=%s\n' "$request_id" "$gateway_http_code"

if [[ ! "$gateway_http_code" =~ ^2 ]]; then
  printf '提示：网关请求不是 2xx，但其错误响应正文已被采集；请检查上面的网关响应。\n'
fi
