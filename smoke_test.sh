#!/usr/bin/env bash
#
# smoke_test.sh
# One-click smoke tests for Tasker backend.
#
# Usage:
#   HOST=http://127.0.0.1 PORT=3000 ./smoke_test.sh
# Environment variables:
#   BUILD_AND_RUN (default "true") - if "true", script will `go build` and run binary from project root.
#   APP_BINARY (optional) - if set, script will run that binary instead of building.
#   APP_ARGS (optional) - arguments passed to binary.
#   APP_WORKDIR (optional) - directory where go build/run is executed (default: current dir)
#

set -u

HOST="${HOST:-http://127.0.0.1}"
PORT="${PORT:-3000}"
BASE_URL="${HOST%/}:${PORT}"
BUILD_AND_RUN="${BUILD_AND_RUN:-true}"
APP_BINARY="${APP_BINARY:-}"
APP_ARGS="${APP_ARGS:-}"
APP_WORKDIR="${APP_WORKDIR:-.}"

TMPDIR="$(mktemp -d)"
OWNER_COOKIE="$TMPDIR/owner_cookies.txt"
USER_COOKIE="$TMPDIR/user_cookies.txt"
LOGFILE="$TMPDIR/smoke.log"
SERVER_PID_FILE="$TMPDIR/server.pid"

# credentials used by tests (non-sensitive, for testing only)
OWNER_LOGIN="test_owner"
OWNER_PASS="owner_pass_123"
OWNER_NAME="Owner"
OWNER_SURNAME="One"
USER_LOGIN="test_user"
USER_PASS="user_pass_123"
USER_NAME="User"
USER_SURNAME="Two"

# helper: log
log() { echo "[$(date +'%H:%M:%S')] $*" | tee -a "$LOGFILE"; }

# JSON field extractor: prefer jq, fallback to python
json_get() {
  local key="$1"
  if command -v jq >/dev/null 2>&1; then
    jq -r "$key" 2>/dev/null
  else
    python -c "import sys, json
obj=json.load(sys.stdin)
def get(o,k):
    for part in k.strip().split('.'):
        if part=='':
            continue
        if part.endswith(']'):
            # support simple array access like .items[0]
            name, idx = part[:-1].split('[')
            o=o.get(name,None)
            if o is None:
                return ''
            o=o[int(idx)]
        else:
            if isinstance(o, dict):
                o=o.get(part,None)
            else:
                return ''
        if o is None:
            return ''
    return o
v=get(obj,'$key')
if v is None:
    print('')
else:
    print(v)"
  fi
}

# check required tools
for cmd in curl go; do
  if ! command -v "$cmd" >/dev/null 2>&1; then
    if [ "$cmd" = "jq" ]; then
      continue
    fi
    if [ "$BUILD_AND_RUN" = "true" ] && [ "$cmd" = "go" ]; then
      log "ERROR: 'go' not found but BUILD_AND_RUN=true. Install Go or set BUILD_AND_RUN=false and provide APP_BINARY."
      exit 1
    fi
  fi
done

# build and run app (optional)
start_server() {
  if [ -n "$APP_BINARY" ]; then
    log "Starting provided binary: $APP_BINARY"
    "$APP_BINARY" $APP_ARGS >/dev/null 2>&1 &
    echo $! > "$SERVER_PID_FILE"
    return 0
  fi

  if [ "$BUILD_AND_RUN" != "true" ]; then
    log "BUILD_AND_RUN != true and no APP_BINARY provided — assuming server already running."
    return 0
  fi

  log "Building the project in $APP_WORKDIR..."
  pushd "$APP_WORKDIR" >/dev/null || return 1
  BIN_PATH="$TMPDIR/tasker_bin"
  if ! go build -o "$BIN_PATH" . 2>&1 | tee -a "$LOGFILE"; then
    log "Build failed. See $LOGFILE"
    popd >/dev/null
    return 1
  fi

  log "Starting server: $BIN_PATH $APP_ARGS"
  nohup "$BIN_PATH" $APP_ARGS >/dev/null 2>&1 &
  echo $! > "$SERVER_PID_FILE"
  popd >/dev/null
  return 0
}

stop_server() {
  if [ -f "$SERVER_PID_FILE" ]; then
    pid=$(cat "$SERVER_PID_FILE")
    if kill -0 "$pid" >/dev/null 2>&1; then
      log "Stopping server (pid $pid)..."
      kill "$pid" || true
      sleep 1
      kill -9 "$pid" >/dev/null 2>&1 || true
    fi
    rm -f "$SERVER_PID_FILE"
  fi
}

wait_for_health() {
  local max=30
  local i=0
  log "Waiting for $BASE_URL/health ..."
  while [ $i -lt $max ]; do
    code=$(curl -s -o /dev/null -w "%{http_code}" "$BASE_URL/health" || echo "000")
    if [ "$code" = "200" ]; then
      body=$(curl -s "$BASE_URL/health" || echo "")
      if [ "$body" = "OK" ] || [ "$code" = "200" ]; then
        log "Health ok"
        return 0
      fi
    fi
    i=$((i+1))
    sleep 1
  done
  log "Health check failed after ${max}s"
  return 1
}

report_summary() {
  echo
  log "=== SUMMARY LOG ==="
  sed -n '1,200p' "$LOGFILE"
  echo "=== END SUMMARY ==="
}

# cleanup on exit
trap 'stop_server; report_summary; rm -rf "$TMPDIR"' EXIT

# start
log "SMOKE TEST START. BASE_URL=$BASE_URL"
start_server || { log "Failed to start server"; exit 1; }

if ! wait_for_health; then
  log "Server did not become healthy; aborting tests."
  exit 2
fi

pass=0
fail=0

# helper for curl POST JSON and capture response+status
do_post() {
  local url="$1"; shift
  local data="$1"; shift
  local cookies="$1"
  local out
  if [ -n "$cookies" ]; then
    out=$(curl -s -w "\n%{http_code}" -X POST "$url" -H "Content-Type: application/json" -d "$data" -c "$cookies" -b "$cookies")
  else
    out=$(curl -s -w "\n%{http_code}" -X POST "$url" -H "Content-Type: application/json" -d "$data")
  fi
  echo "$out"
}

do_get() {
  local url="$1"; shift
  local cookies="$1"
  if [ -n "$cookies" ]; then
    curl -s -w "\n%{http_code}" -b "$cookies" -c "$cookies" "$url"
  else
    curl -s -w "\n%{http_code}" "$url"
  fi
}

# 1) Register owner
log "1) Registering owner..."
owner_reg_payload=$(jq -n --arg n "$OWNER_NAME" --arg s "$OWNER_SURNAME" --arg l "$OWNER_LOGIN" --arg p "$OWNER_PASS" --arg r "spaceOwner" '{name:$n,surname:$s,login:$l,password:$p,role:$r,spaceName:"Smoke Test Space"}' 2>/dev/null || \
  printf '{"name":"%s","surname":"%s","login":"%s","password":"%s","role":"spaceOwner","spaceName":"Smoke Test Space"}' "$OWNER_NAME" "$OWNER_SURNAME" "$OWNER_LOGIN" "$OWNER_PASS")
owner_resp=$(do_post "$BASE_URL/api/register" "$owner_reg_payload" "")
owner_body=$(echo "$owner_resp" | sed '$d')   # all but last line
owner_code=$(echo "$owner_resp" | tail -n1)
log "Register owner HTTP:$owner_code body: $owner_body"
if [ "$owner_code" = "201" ] || [ "$owner_code" = "200" ]; then
  owner_id=$(echo "$owner_body" | json_get '.id')
  log "Owner id = $owner_id"
  pass=$((pass+1))
else
  log "Owner registration FAILED"
  fail=$((fail+1))
fi

# 2) Register user
log "2) Registering plain user..."
user_reg_payload=$(jq -n --arg n "$USER_NAME" --arg s "$USER_SURNAME" --arg l "$USER_LOGIN" --arg p "$USER_PASS" --arg r "user" '{name:$n,surname:$s,login:$l,password:$p,role:$r}' 2>/dev/null || \
  printf '{"name":"%s","surname":"%s","login":"%s","password":"%s","role":"user"}' "$USER_NAME" "$USER_SURNAME" "$USER_LOGIN" "$USER_PASS")
user_resp=$(do_post "$BASE_URL/api/register" "$user_reg_payload" "")
user_body=$(echo "$user_resp" | sed '$d')
user_code=$(echo "$user_resp" | tail -n1)
log "Register user HTTP:$user_code body: $user_body"
if [ "$user_code" = "201" ] || [ "$user_code" = "200" ]; then
  user_id=$(echo "$user_body" | json_get '.id')
  log "User id = $user_id"
  pass=$((pass+1))
else
  log "User registration FAILED"
  fail=$((fail+1))
fi

# 3) Login owner (store cookies)
log "3) Login owner (store cookies)..."
owner_login_payload=$(jq -n --arg l "$OWNER_LOGIN" --arg p "$OWNER_PASS" '{login:$l,password:$p}' 2>/dev/null || printf '{"login":"%s","password":"%s"}' "$OWNER_LOGIN" "$OWNER_PASS")
owner_login_resp=$(do_post "$BASE_URL/api/login" "$owner_login_payload" "$OWNER_COOKIE")
owner_login_body=$(echo "$owner_login_resp" | sed '$d')
owner_login_code=$(echo "$owner_login_resp" | tail -n1)
log "Owner login HTTP:$owner_login_code"
if [ "$owner_login_code" = "200" ]; then
  pass=$((pass+1))
else
  log "Owner login FAILED"
  fail=$((fail+1))
fi

# 4) Create a space via API (so we can get space ID)
log "4) Owner creates a new space to obtain spaceID..."
create_space_payload=$(jq -n --arg n "smoke-space-$(date +%s)" '{name:$n}' 2>/dev/null || printf '{"name":"smoke-space-%s"}' "$(date +%s)")
space_resp=$(do_post "$BASE_URL/spaces/" "$create_space_payload" "$OWNER_COOKIE")
space_body=$(echo "$space_resp" | sed '$d')
space_code=$(echo "$space_resp" | tail -n1)
log "Create space HTTP:$space_code body: $space_body"
if [ "$space_code" = "201" ] || [ "$space_code" = "200" ]; then
  space_id=$(echo "$space_body" | json_get '.id')
  log "Created space id = $space_id"
  pass=$((pass+1))
else
  log "Create space FAILED"
  fail=$((fail+1))
fi

# 5) Owner gets space token
log "5) Owner requests space token..."
token_resp=$(do_get "$BASE_URL/spaces/$space_id/token" "$OWNER_COOKIE")
token_body=$(echo "$token_resp" | sed '$d')
token_code=$(echo "$token_resp" | tail -n1)
log "Get token HTTP:$token_code body:$token_body"
if [ "$token_code" = "200" ]; then
  token_val=$(echo "$token_body" | json_get '.token')
  log "Token = $token_val"
  pass=$((pass+1))
else
  log "Get token FAILED"
  fail=$((fail+1))
fi

# 6) Owner creates dashboard in space
log "6) Owner creates dashboard in space..."
dash_payload=$(jq -n --arg name "Sprint Board" '{name:$name}' 2>/dev/null || printf '{"name":"Sprint Board"}')
dash_resp=$(do_post "$BASE_URL/spaces/$space_id/dashboards" "$dash_payload" "$OWNER_COOKIE")
dash_body=$(echo "$dash_resp" | sed '$d')
dash_code=$(echo "$dash_resp" | tail -n1)
log "Create dashboard HTTP:$dash_code body:$dash_body"
if [ "$dash_code" = "201" ] || [ "$dash_code" = "200" ]; then
  dash_id=$(echo "$dash_body" | json_get '.id')
  log "Dashboard id = $dash_id"
  pass=$((pass+1))
else
  log "Create dashboard FAILED"
  fail=$((fail+1))
fi

# 7) Login plain user (store cookies)
log "7) Login plain user (store cookies)..."
user_login_payload=$(jq -n --arg l "$USER_LOGIN" --arg p "$USER_PASS" '{login:$l,password:$p}' 2>/dev/null || printf '{"login":"%s","password":"%s"}' "$USER_LOGIN" "$USER_PASS")
user_login_resp=$(do_post "$BASE_URL/api/login" "$user_login_payload" "$USER_COOKIE")
user_login_body=$(echo "$user_login_resp" | sed '$d')
user_login_code=$(echo "$user_login_resp" | tail -n1)
log "User login HTTP:$user_login_code"
if [ "$user_login_code" = "200" ]; then
  pass=$((pass+1))
else
  log "User login FAILED"
  fail=$((fail+1))
fi

# 8) User joins by token
log "8) Plain user joins space by token..."
join_payload=$(jq -n --arg t "$token_val" '{token:$t}' 2>/dev/null || printf '{"token":"%s"}' "$token_val")
join_resp=$(do_post "$BASE_URL/spaces/join" "$join_payload" "$USER_COOKIE")
join_body=$(echo "$join_resp" | sed '$d')
join_code=$(echo "$join_resp" | tail -n1)
log "Join HTTP:$join_code body:$join_body"
if [ "$join_code" = "204" ] || [ "$join_code" = "200" ]; then
  log "User joined OK"
  pass=$((pass+1))
else
  log "User join FAILED"
  fail=$((fail+1))
fi

# 9) Owner adds a custom role to space
log "9) Owner adds custom role 'viewer' to space..."
addrole_payload='{"name":"viewer"}'
addrole_resp=$(do_post "$BASE_URL/spaces/$space_id/roles" "$addrole_payload" "$OWNER_COOKIE")
addrole_body=$(echo "$addrole_resp" | sed '$d')
addrole_code=$(echo "$addrole_resp" | tail -n1)
log "Add role HTTP:$addrole_code"
if [ "$addrole_code" = "201" ] || [ "$addrole_code" = "200" ] || [ "$addrole_code" = "204" ]; then
  pass=$((pass+1))
else
  log "Add role FAILED"
  fail=$((fail+1))
fi

# 10) Owner removes 'viewer' role
log "10) Owner removes role 'viewer'..."
delrole_resp=$(curl -s -w "\n%{http_code}" -X DELETE -b "$OWNER_COOKIE" -c "$OWNER_COOKIE" "$BASE_URL/spaces/$space_id/roles/viewer")
delrole_body=$(echo "$delrole_resp" | sed '$d')
delrole_code=$(echo "$delrole_resp" | tail -n1)
log "Delete role HTTP:$delrole_code"
if [ "$delrole_code" = "204" ] || [ "$delrole_code" = "200" ]; then
  pass=$((pass+1))
else
  log "Delete role FAILED"
  fail=$((fail+1))
fi

# 11) Owner removes the plain user (member) from space
log "11) Owner removes user (member) from space..."
delmem_resp=$(curl -s -w "\n%{http_code}" -X DELETE -b "$OWNER_COOKIE" -c "$OWNER_COOKIE" "$BASE_URL/spaces/$space_id/members/$user_id")
delmem_body=$(echo "$delmem_resp" | sed '$d')
delmem_code=$(echo "$delmem_resp" | tail -n1)
log "Delete member HTTP:$delmem_code"
if [ "$delmem_code" = "204" ] || [ "$delmem_code" = "200" ]; then
  pass=$((pass+1))
else
  log "Delete member FAILED"
  fail=$((fail+1))
fi

# 12) Owner deletes created dashboard
log "12) Owner deletes created dashboard..."
deldash_resp=$(curl -s -w "\n%{http_code}" -X DELETE -b "$OWNER_COOKIE" -c "$OWNER_COOKIE" "$BASE_URL/spaces/$space_id/dashboards/$dash_id")
deldash_body=$(echo "$deldash_resp" | sed '$d')
deldash_code=$(echo "$deldash_resp" | tail -n1)
log "Delete dashboard HTTP:$deldash_code"
if [ "$deldash_code" = "204" ] || [ "$deldash_code" = "200" ]; then
  pass=$((pass+1))
else
  log "Delete dashboard FAILED"
  fail=$((fail+1))
fi

# Finish
log "Tests completed. Passed: $pass  Failed: $fail"
if [ "$fail" -gt 0 ]; then
  exit 3
else
  exit 0
fi
