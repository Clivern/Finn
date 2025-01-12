#!/usr/bin/env bash
# Test the runner /_ping endpoint with curl.
# Usage: $0 <public_key.pem> [base_url] [runner_token]
# Example: $0 runner_pub.pem http://127.0.0.1:8000 a14f48d0-b338-4641-9ebb-f359ab7722ce
# The runner requires X-Runner-Token when provided.

set -e

PUBKEY_FILE="${1:?Usage: $0 <public_key.pem> [base_url] [runner_token]}"
BASE_URL="${2:-http://127.0.0.1:8000}"
RUNNER_TOKEN="${3:-}"

PAYLOAD=$(echo -n "PONG" | openssl pkeyutl -encrypt -pubin -inkey "$PUBKEY_FILE" \
  -pkeyopt rsa_padding_mode:oaep -pkeyopt rsa_oaep_md:sha256 | base64)

CURL_CMD=(curl -s -X POST "${BASE_URL}/_ping" -H "Content-Type: text/plain" -d "$PAYLOAD")
if [[ -n "$RUNNER_TOKEN" ]]; then
  CURL_CMD+=(-H "X-Runner-Token: $RUNNER_TOKEN")
fi

"${CURL_CMD[@]}" --verbose
echo
