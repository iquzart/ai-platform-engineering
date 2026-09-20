#!/usr/bin/env bash
# Validate the local chat API's semantic-cache miss and store path.
set -euo pipefail

API_URL="http://localhost:8080"
REQUEST_TIMEOUT=120
TMP_DIR=""

log() {
  local level=$1 event=$2 message=${3:-}
  jq -cn \
    --arg timestamp "$(date -u +%Y-%m-%dT%H:%M:%SZ)" \
    --arg level "$level" \
    --arg event "$event" \
    --arg message "$message" \
    '{timestamp: $timestamp, level: $level, event: $event, message: $message}'
}

info() { log INFO "$@"; }
fail() { log FAIL "$@" >&2; exit 1; }

cleanup() {
  [[ -n "$TMP_DIR" ]] && rm -rf -- "$TMP_DIR"
}

usage() {
  cat <<'EOF'
Usage: ./validate-semantic-cache.sh "CHAT MESSAGE"

Validates a cache miss followed by an identical-request cache hit for the
supplied chat message against the local API at http://localhost:8080. A
semantic cache can match a differently worded request, so run this against a
fresh cache when validating the miss-and-store path.
EOF
}

require_command() {
  command -v "$1" >/dev/null 2>&1 || fail dependency_missing "Required command not found: $1"
}

curl_get() {
  local path=$1 output=$2 status
  if ! status=$(curl --silent --show-error --location --output "$output" --write-out '%{http_code}' \
      --connect-timeout 3 --max-time "$REQUEST_TIMEOUT" "${API_URL}${path}"); then
    fail request_failed "Request to ${path} could not be completed. Ensure the local stack is running."
  fi
  [[ "$status" == "200" ]] || fail unexpected_http_status "${path} returned HTTP ${status}; expected 200."
}

assert_json_status() {
  local file=$1 expected=$2 endpoint=$3
  jq -e --arg expected "$expected" \
    'if .status == $expected then true else error("unexpected status: " + (.status | tojson)) end' \
    "$file" >/dev/null 2>&1 || {
      printf '%s did not return valid JSON with status=%s\n' "$endpoint" "$expected"
      return 1
    }
}

metric_value() {
  local metric=$1 file=$2
  awk -v metric="$metric" '$1 == metric { print $2; found = 1; exit } END { exit !found }' "$file"
}

assert_metric_delta() {
  local metric=$1 before=$2 after=$3 minimum=$4
  awk -v metric="$metric" -v before="$before" -v after="$after" -v minimum="$minimum" '
    BEGIN {
      number = "^[-+]?[0-9]*([.][0-9]+)?([eE][-+]?[0-9]+)?$"
      if (before !~ number || after !~ number || minimum !~ number) {
        printf "%s contains a non-numeric metric value\n", metric
        exit 1
      }
      delta = after - before
      if (delta < minimum) {
        printf "%s increased by %g; expected at least %g\n", metric, delta, minimum
        exit 1
      }
      printf "%s delta=%g\n", metric, delta
    }
  '
}

run_check() {
  local description=$1 output
  shift
  if ! output=$("$@" 2>&1); then
    fail check_failed "${description}: ${output}"
  fi
  [[ -z "$output" ]] || info check_result "$output"
}

log_json_payload() {
  local event=$1 attempt=$2 payload_name=$3 file=$4
  jq -c \
    --arg timestamp "$(date -u +%Y-%m-%dT%H:%M:%SZ)" \
    --arg event "$event" \
    --argjson attempt "$attempt" \
    --arg payload_name "$payload_name" \
    '{timestamp: $timestamp, level: "INFO", event: $event, attempt: $attempt} + {($payload_name): .}' \
    "$file"
}

log_chat_response() {
  local attempt=$1 response=$2
  if jq -e . "$response" >/dev/null 2>&1; then
    log_json_payload chat_response "$attempt" response "$response"
  else
    jq -cn \
      --arg timestamp "$(date -u +%Y-%m-%dT%H:%M:%SZ)" \
      --argjson attempt "$attempt" \
      --rawfile response "$response" \
      '{timestamp: $timestamp, level: "INFO", event: "chat_response", attempt: $attempt, response_raw: $response}'
  fi
}

post_chat() {
  local request=$1 response=$2 attempt=$3 status
  log_json_payload chat_request "$attempt" request "$request"
  if ! status=$(curl --silent --show-error --location --output "$response" --write-out '%{http_code}' \
      --connect-timeout 3 --max-time "$REQUEST_TIMEOUT" --request POST "${API_URL}/chat" \
      --header 'Content-Type: application/json' --data-binary "@${request}"); then
    fail request_failed "POST /chat could not be completed. Check the configured OpenAI-compatible provider, Redis, and the chat-app logs."
  fi
  [[ "$status" == "200" ]] || fail unexpected_http_status "POST /chat returned HTTP ${status}; expected 200."
  log_chat_response "$attempt" "$response"
}

assert_cache_response() {
  jq -e -s '
    def valid_response($number):
      if (.answer | type) != "string" then
        error("chat response " + ($number | tostring) + " has no non-empty answer")
      elif (.answer | test("\\S") | not) then
        error("chat response " + ($number | tostring) + " has no non-empty answer")
      elif (.cache_hit | type) != "boolean" then
        error("chat response " + ($number | tostring) + " has a non-boolean cache_hit")
      elif (.similarity | type) != "number" then
        error("chat response " + ($number | tostring) + " has a non-numeric similarity")
      else true
      end;
    .[0] as $response |
    ($response | valid_response(1)) |
    "response valid: cache_hit=\($response.cache_hit); similarity=\($response.similarity)"
  ' "$1"
}

assert_cache_hit() {
  local file=$1 expected=$2
  jq -e --argjson expected "$expected" '
    if .cache_hit == $expected then true
    else error("cache_hit was \(.cache_hit); expected \($expected)")
    end
  ' "$file"
}

assert_same_answer() {
  jq -e -s '
    if .[0].answer == .[1].answer then true
    else error("cache-hit response answer differs from the cache-miss response answer")
    end
  ' "$1" "$2"
}

main() {
  [[ $# -eq 1 ]] || { usage >&2; fail invalid_arguments "Provide exactly one chat message."; }
  [[ "$1" =~ [^[:space:]] ]] || fail invalid_arguments "Chat message must contain non-whitespace text."
  local chat_message=$1

  if ! command -v jq >/dev/null 2>&1; then
    printf '{"timestamp":"%s","level":"FAIL","event":"dependency_missing","message":"Required command not found: jq"}\n' "$(date -u +%Y-%m-%dT%H:%M:%SZ)" >&2
    exit 1
  fi
  require_command curl
  require_command awk

  TMP_DIR=$(mktemp -d "${TMPDIR:-/tmp}/semantic-cache-validation.XXXXXX")
  trap cleanup EXIT
  info validation_started "Validating semantic cache against ${API_URL} (request timeout: ${REQUEST_TIMEOUT}s)"

  curl_get /system/health/live "${TMP_DIR}/live.json"
  run_check "Liveness precheck failed" assert_json_status "${TMP_DIR}/live.json" live /system/health/live
  info liveness_precheck_passed "Liveness precheck passed"
  curl_get /system/health/ready "${TMP_DIR}/ready.json"
  run_check "Readiness precheck failed" assert_json_status "${TMP_DIR}/ready.json" ready /system/health/ready
  info readiness_precheck_passed "Readiness precheck passed"
  curl_get /system/metrics "${TMP_DIR}/metrics-before.txt"

  local request_id query before_requests before_hits before_misses after_requests after_hits after_misses
  request_id="$(date -u +%Y%m%dT%H%M%SZ)-$$-$RANDOM"
  query="${chat_message} [semantic-cache-validation: ${request_id}]"
  jq -n --arg query "$query" '{query: $query}' >"${TMP_DIR}/request.json"

  if ! before_requests=$(metric_value semantic_cache_requests_total "${TMP_DIR}/metrics-before.txt"); then fail metric_missing "Unable to read semantic_cache_requests_total before validation"; fi
  if ! before_hits=$(metric_value semantic_cache_hits_total "${TMP_DIR}/metrics-before.txt"); then fail metric_missing "Unable to read semantic_cache_hits_total before validation"; fi
  if ! before_misses=$(metric_value semantic_cache_misses_total "${TMP_DIR}/metrics-before.txt"); then fail metric_missing "Unable to read semantic_cache_misses_total before validation"; fi

  post_chat "${TMP_DIR}/request.json" "${TMP_DIR}/first.json" 1
  run_check "Semantic-cache response validation failed" assert_cache_response "${TMP_DIR}/first.json"
  if ! jq -e '.cache_hit == false' "${TMP_DIR}/first.json" >/dev/null; then
    fail semantic_neighbor_hit "Initial request matched an existing semantic-cache entry. A unique identifier does not guarantee a miss with nearest-neighbor matching; use a fresh local cache to validate the miss-and-store path."
  fi
  run_check "Cache-miss response validation failed" assert_cache_hit "${TMP_DIR}/first.json" false
  post_chat "${TMP_DIR}/request.json" "${TMP_DIR}/repeat.json" 2
  run_check "Repeated response validation failed" assert_cache_response "${TMP_DIR}/repeat.json"
  run_check "Repeated request did not hit the semantic cache" assert_cache_hit "${TMP_DIR}/repeat.json" true
  run_check "Repeated request returned a different answer" assert_same_answer "${TMP_DIR}/first.json" "${TMP_DIR}/repeat.json"

  curl_get /system/metrics "${TMP_DIR}/metrics-after.txt"
  if ! after_requests=$(metric_value semantic_cache_requests_total "${TMP_DIR}/metrics-after.txt"); then fail metric_missing "Unable to read semantic_cache_requests_total after validation"; fi
  if ! after_hits=$(metric_value semantic_cache_hits_total "${TMP_DIR}/metrics-after.txt"); then fail metric_missing "Unable to read semantic_cache_hits_total after validation"; fi
  if ! after_misses=$(metric_value semantic_cache_misses_total "${TMP_DIR}/metrics-after.txt"); then fail metric_missing "Unable to read semantic_cache_misses_total after validation"; fi
  run_check "Request metric validation failed" assert_metric_delta semantic_cache_requests_total "$before_requests" "$after_requests" 2
  run_check "Hit metric validation failed" assert_metric_delta semantic_cache_hits_total "$before_hits" "$after_hits" 1
  run_check "Miss metric validation failed" assert_metric_delta semantic_cache_misses_total "$before_misses" "$after_misses" 1

  info validation_passed "PASS: initial request missed and was stored; its identical repeat was a semantic-cache hit"
}

main "$@"
