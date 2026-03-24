#!/bin/bash
set -e

CLIENT_PATH="${CLIENT_PATH:-./order_service_client/bin/ordercli}"
SERVER_ADDR="${SERVER_ADDR:-localhost:50081}"
ENV_FILE="${ENV_FILE:-deploy/order_service/.env}"

JWT_SECRET="secret"

# Аргументы
MARKET_UUID="${1:-22222222-2222-2222-2222-222222222222}"
PRICE="${2:-10.12}"
QUANTITY="${3:-5}"
SIDE="${4:-buy}"
ORDER_TYPE="${5:-limit}"
USER_UUID="${6:-$(uuidgen)}"

b64url() {
    echo -n "$1" | base64 | tr -d '=' | tr '+/' '-_' | tr -d '\n'
}

HEADER='{"alg":"HS256","typ":"JWT"}'

# Текущий timestamp + 30 минут (1800 секунд)
EXP=$(($(date +%s) + 1800))
PAYLOAD="{\"sub\":\"${USER_UUID}\",\"exp\":${EXP},\"rls\":[\"TRADER\",\"MODER\"]}"

HEADER_B64=$(b64url "$HEADER")
PAYLOAD_B64=$(b64url "$PAYLOAD")

SIGNATURE=$(echo -n "${HEADER_B64}.${PAYLOAD_B64}" | openssl dgst -sha256 -hmac "$JWT_SECRET" -binary | base64 | tr -d '=' | tr '+/' '-_' | tr -d '\n')
JWT="${HEADER_B64}.${PAYLOAD_B64}.${SIGNATURE}"

echo "User UUID: $USER_UUID"
echo "JWT: $JWT"

$CLIENT_PATH create \
    --addr="$SERVER_ADDR" \
    --jwt="$JWT" \
    -m="$MARKET_UUID" \
    -p="$PRICE" \
    -q="$QUANTITY" \
    -s="$SIDE" \
    -t="$ORDER_TYPE"