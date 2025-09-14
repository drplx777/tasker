#!/usr/bin/env bash
# smoke_ci.sh
# One-shot CI smoke test for Tasker backend (assumes app already running, e.g. in Docker).
# Exits 0 if all steps passed, non-zero otherwise.
#
# Usage:
#   HOST=http://localhost PORT=3000 ./smoke_ci.sh
# Optional env:
#   OWNER_LOGIN, OWNER_PASS, USER_LOGIN, USER_PASS (defaults below)
set -euo pipefail

HOST="${HOST:-http://localhost}"
PORT="${PORT:-3000}"
BASE_URL="${HOST%/}:${PORT}"

OWNER_LOGIN="${OWNER_LOGIN:-ci_owner}"
OWNER_PASS="${OWNER_PASS:-owner_pass_123}"
OWNER_NAME="${OWNER_NAME:-CIOwner}"
OWNER_SURNAME="${OWNER_SURNAME:-Owner}"
USER_LOGIN="${USER_LOGIN:-ci_user}"
USER_PASS="${USER_PASS:-user_pass_123}"
USER_NAME="${USER_NAME:-CIUser}"
USER_SURNAME="${USER_SURNAME:-User}"

log() { printf '%s %s\n' "[$(date +'%H:%M:%S')]" "$*"; }

# helper: run curl POST and return combined (body+headers) + status code appended on last line
curl_post() {
  local url="$1"; local data="$2"; shift 2
  # we capture headers+body and append http code as last line
  curl -s -w "\n%{http_code}" -X POST "$url" -H "Content-Type: application/json" -d "$data" "$@"
}

curl_get() {
  local url="$1"; shift
  curl -s -w "\n%{http_code}" -X GET "$url" "$@"
}

# extract last line as status, rest as response
extract_status_and_body() {
  local resp="$1"
  status="$(printf "%s" "$resp" | tail -n1)"
  body="$(printf "%s" "$resp" | sed '$d')"
}

# extract api_token value from response (search Set-Cookie header)
extract_api_token_from_resp() {
  local resp="$1"
  # normalize CRLF to LF and find Set-Cookie lines, then extract api_token value before ';'
  token="$(printf "%s" "$resp" | tr '\r' '\n' | awk 'BEGIN{IGNORECASE=1} /Set-Cookie:/ {print;}' | grep -i 'api_token=' || true)"
  if [ -n "$token" ]; then
    # extract value after api_token=
    api_token_value="$(printf "%s" "$token" | sed -E 's/.*[Aa][Pp][Ii]_[Tt][Oo][Kk][Ee][Nn]=([^;]+).*/\1/;t;d' | head -n1 || true)"
    api_token_value="${api_token_value:-}"
  else
    api_token_value=""
  fi
}

# check health
log "Checking health at $BASE_URL/health"
hc="$(curl -s -o /dev/null -w "%{http_code}" "$BASE_URL/health" || echo "000")"
if [ "$hc" != "200" ]; then
  log "Health check failed: HTTP $hc. Aborting."
  exit 2
fi
log "Health ok"

pass=0
fail=0

# 1) Register owner
log "1) Register owner ${OWNER_LOGIN}"
payload_owner=$(cat <<EOF
{"name":"$OWNER_NAME","surname":"$OWNER_SURNAME","login":"$OWNER_LOGIN","password":"$OWNER_PASS","role":"spaceOwner","spaceName":"CI Space"}
EOF
)
resp="$(curl_post "$BASE_URL/api/register" "$payload_owner")"
extract_status_and_body "$resp"
log "Register owner HTTP:$status Body: $body"
if [ "$status" = "201" ] || [ "$status" = "200" ]; then
  owner_id="$(printf "%s" "$body" | sed -n 's/.*"id":[[:space:]]*\([0-9][0-9]*\).*/\1/p' | head -n1 || true)"
  log "Owner id: ${owner_id:-<not-found>}"
  pass=$((pass+1))
else
  log "Owner registration failed"
  fail=$((fail+1))
fi

# 2) Register user
log "2) Register user ${USER_LOGIN}"
payload_user=$(cat <<EOF
{"name":"$USER_NAME","surname":"$USER_SURNAME","login":"$USER_LOGIN","password":"$USER_PASS","role":"user"}
EOF
)
resp="$(curl_post "$BASE_URL/api/register" "$payload_user")"
extract_status_and_body "$resp"
log "Register user HTTP:$status Body: $body"
if [ "$status" = "201" ] || [ "$status" = "200" ]; then
  user_id="$(printf "%s" "$body" | sed -n 's/.*"id":[[:space:]]*\([0-9][0-9]*\).*/\1/p' | head -n1 || true)"
  log "User id: ${user_id:-<not-found>}"
  pass=$((pass+1))
else
  log "User registration failed"
  fail=$((fail+1))
fi

# 3) Login owner -> extract api_token
log "3) Login owner to get api_token"
payload_login_owner=$(cat <<EOF
{"login":"$OWNER_LOGIN","password":"$OWNER_PASS"}
EOF
)
resp="$(curl_post "$BASE_URL/api/login" "$payload_login_owner")"
extract_status_and_body "$resp"
log "Owner login HTTP:$status Body: (omitted)"
if [ "$status" = "200" ]; then
  extract_api_token_from_resp "$body"
  if [ -n "${api_token_value:-}" ]; then
    JWT="$api_token_value"
    log "Extracted api_token (len=$(printf "%d" ${#JWT}))"
    pass=$((pass+1))
  else
    log "Failed to extract api_token from login response headers"
    fail=$((fail+1))
  fi
else
  log "Owner login failed"
  fail=$((fail+1))
fi

# convenience header to always send jwt explicitly
COOKIE_HEADER="api_token=${JWT:-}"

# 4) Create space (owner)
log "4) Create space via POST /spaces/ (owner)"
payload_space='{"name":"ci-created-space"}'
resp="$(curl_post "$BASE_URL/spaces/" "$payload_space" -H "Cookie: ${COOKIE_HEADER}")"
extract_status_and_body "$resp"
log "Create space HTTP:$status Body: $body"
if [ "$status" = "201" ] || [ "$status" = "200" ]; then
  space_id="$(printf "%s" "$body" | sed -n 's/.*"id":[[:space:]]*"\([^"]\+\)".*/\1/p' | head -n1 || true)"
  log "space_id: ${space_id:-<not-found>}"
  pass=$((pass+1))
else
  log "Create space failed"
  fail=$((fail+1))
fi

# 5) Get space token (space id) — use owner JWT header
log "5) GET /spaces/{id}/token"
if [ -n "${space_id:-}" ]; then
  resp="$(curl_get "$BASE_URL/spaces/$space_id/token" -H "Cookie: ${COOKIE_HEADER}")"
  extract_status_and_body "$resp"
  log "Get token HTTP:$status Body: $body"
  if [ "$status" = "200" ]; then
    space_token="$(printf "%s" "$body" | sed -n 's/.*"token"[[:space:]]*:[[:space:]]*"\([^"]\+\)".*/\1/p' | head -n1 || true)"
    space_token="${space_token:-$space_id}" # fallback to id
    log "space_token: ${space_token}"
    pass=$((pass+1))
  else
    log "Get token failed"
    fail=$((fail+1))
  fi
else
  log "Skipping get token because space_id empty"
  fail=$((fail+1))
fi

# 6) Create dashboard in space
log "6) Create dashboard in space"
payload_db='{"name":"ci-dashboard"}'
resp="$(curl_post "$BASE_URL/spaces/$space_id/dashboards" "$payload_db" -H "Cookie: ${COOKIE_HEADER}")"
extract_status_and_body "$resp"
log "Create dashboard HTTP:$status Body: $body"
if [ "$status" = "201" ] || [ "$status" = "200" ]; then
  dash_id="$(printf "%s" "$body" | sed -n 's/.*"id"[[:space:]]*:[[:space:]]*"\?\([^",}]\+\)".*/\1/p' | head -n1 || true)"
  log "dash_id: ${dash_id:-<not-found>}"
  pass=$((pass+1))
else
  log "Create dashboard failed"
  fail=$((fail+1))
fi

# 7) Login plain user
log "7) Login plain user to get api_token"
payload_login_user=$(cat <<EOF
{"login":"$USER_LOGIN","password":"$USER_PASS"}
EOF
)
resp="$(curl_post "$BASE_URL/api/login" "$payload_login_user")"
extract_status_and_body "$resp"
log "User login HTTP:$status"
if [ "$status" = "200" ]; then
  extract_api_token_from_resp "$body"
  if [ -n "${api_token_value:-}" ]; then
    USER_JWT="$api_token_value"
    log "Extracted user api_token"
    pass=$((pass+1))
  else
    log "Failed to extract user api_token"
    fail=$((fail+1))
  fi
else
  log "User login failed"
  fail=$((fail+1))
fi

# 8) User joins space by token (uses space_token)
log "8) User joins by token"
join_payload="{\"token\":\"${space_token}\",\"role\":\"member\"}"
resp="$(curl_post "$BASE_URL/spaces/join" "$join_payload" -H "Cookie: api_token=${USER_JWT:-}")"
extract_status_and_body "$resp"
log "Join HTTP:$status Body: $body"
if [ "$status" = "204" ] || [ "$status" = "200" ]; then
  pass=$((pass+1))
else
  log "User join failed"
  fail=$((fail+1))
fi

# 9) Owner adds custom role
log "9) Owner adds role 'viewer'"
addrole_payload='{"name":"viewer"}'
resp="$(curl_post "$BASE_URL/spaces/$space_id/roles" "$addrole_payload" -H "Cookie: ${COOKIE_HEADER}")"
extract_status_and_body "$resp"
log "Add role HTTP:$status"
if [ "$status" = "201" ] || [ "$status" = "200" ] || [ "$status" = "204" ]; then
  pass=$((pass+1))
else
  log "Add role failed"
  fail=$((fail+1))
fi

# 10) Owner removes role
log "10) Owner deletes role 'viewer'"
delrole_resp="$(curl -s -w "\n%{http_code}" -X DELETE -H "Cookie: ${COOKIE_HEADER}" "$BASE_URL/spaces/$space_id/roles/viewer")"
delrole_status="$(printf "%s" "$delrole_resp" | tail -n1)"
log "Delete role HTTP:$delrole_status"
if [ "$delrole_status" = "204" ] || [ "$delrole_status" = "200" ]; then
  pass=$((pass+1))
else
  log "Delete role failed"
  fail=$((fail+1))
fi

# 11) Owner removes user from space
log "11) Owner deletes member (user_id: ${user_id:-})"
delmem_resp="$(curl -s -w "\n%{http_code}" -X DELETE -H "Cookie: ${COOKIE_HEADER}" "$BASE_URL/spaces/$space_id/members/${user_id:-}")"
delmem_status="$(printf "%s" "$delmem_resp" | tail -n1)"
log "Delete member HTTP:$delmem_status"
if [ "$delmem_status" = "204" ] || [ "$delmem_status" = "200" ]; then
  pass=$((pass+1))
else
  log "Delete member failed"
  fail=$((fail+1))
fi

# 12) Owner deletes dashboard
log "12) Owner deletes dashboard (dash_id: ${dash_id:-})"
deldash_resp="$(curl -s -w "\n%{http_code}" -X DELETE -H "Cookie: ${COOKIE_HEADER}" "$BASE_URL/spaces/$space_id/dashboards/${dash_id:-}")"
deldash_status="$(printf "%s" "$deldash_resp" | tail -n1)"
log "Delete dashboard HTTP:$deldash_status"
if [ "$deldash_status" = "204" ] || [ "$deldash_status" = "200" ]; then
  pass=$((pass+1))
else
  log "Delete dashboard failed"
  fail=$((fail+1))
fi

log "Smoke tests finished. Passed: $pass  Failed: $fail"
if [ "$fail" -gt 0 ]; then
  exit 3
fi
exit 0
