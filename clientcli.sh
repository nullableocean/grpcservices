#!/bin/bash
set -e

CLIENT_PATH="${CLIENT_PATH:-./order_service_client/bin/ordercli}"
SERVER_ADDR="${SERVER_ADDR:-localhost:50081}"
JWT_SECRET="${JWT_SECRET:-secret}"


generate_jwt() {
    local user_uuid="${1:-$(uuidgen)}"
    local header='{"alg":"HS256","typ":"JWT"}'
    local exp=$(($(date +%s) + 1800))
    local payload="{\"sub\":\"${user_uuid}\",\"uuid\":\"${user_uuid}\",\"exp\":${exp},\"roles\":[\"TRADER\",\"MODER\"]}"

    b64url() {
        echo -n "$1" | base64 | tr -d '=' | tr '+/' '-_' | tr -d '\n'
    }

    local header_b64=$(b64url "$header")
    local payload_b64=$(b64url "$payload")
    local signature=$(echo -n "${header_b64}.${payload_b64}" | openssl dgst -sha256 -hmac "$JWT_SECRET" -binary | base64 | tr -d '=' | tr '+/' '-_' | tr -d '\n')
    local jwt="${header_b64}.${payload_b64}.${signature}"

    JWT="$jwt"
    USER_UUID="$user_uuid"
}

if [ $# -lt 1 ]; then
    echo "Usage: $0 {create|list} [args...]"
    exit 1
fi

COMMAND="$1"
shift

case "$COMMAND" in
    create)
        # Ожидаем: MARKET_UUID PRICE QUANTITY SIDE ORDER_TYPE [USER_UUID]
        USER_UUID_ARG="${1:-$(uuidgen)}"
        MARKET_UUID="${2:-22222222-2222-2222-2222-222222222222}"
        PRICE="${3:-10.12}"
        QUANTITY="${4:-5}"
        SIDE="${5:-buy}"
        ORDER_TYPE="${6:-limit}"

        # Генерируем JWT (если USER_UUID_ARG не пуст – используем его)
        generate_jwt "$USER_UUID_ARG"

        echo "User UUID: $USER_UUID"
        echo "JWT: $JWT"
        echo ""

        # Выполняем команду create
        exec $CLIENT_PATH create \
            --addr="$SERVER_ADDR" \
            --jwt="$JWT" \
            -m="$MARKET_UUID" \
            -p="$PRICE" \
            -q="$QUANTITY" \
            -s="$SIDE" \
            -t="$ORDER_TYPE" \
            -u="$USER_UUID"
        ;;

    list)
        # Ожидаем: [USER_UUID] [PAGE_SIZE] [PAGE_TOKEN]
        USER_UUID_ARG="${1:-}"
        PAGE_SIZE="${2:-10}"
        PAGE_TOKEN="${3:-}"

        if [ -z "$USER_UUID_ARG" ]; then
            echo "empty user uuid"
            exit 1
        fi

        generate_jwt "$USER_UUID_ARG"

        echo "User UUID: $USER_UUID"
        echo ""
        
        # Флаги для list
        LIST_FLAGS="--addr=$SERVER_ADDR --jwt=$JWT -u=$USER_UUID"
        if [ -n "$PAGE_TOKEN" ]; then
            LIST_FLAGS="$LIST_FLAGS -t=$PAGE_TOKEN"
        fi
        if [ -n "$PAGE_SIZE" ]; then
            LIST_FLAGS="$LIST_FLAGS -l=$PAGE_SIZE"
        fi

        # Выполняем команду list
        exec $CLIENT_PATH list $LIST_FLAGS
        ;;

    *)
        echo "Unknown command: $COMMAND. Use 'create' or 'list'."
        exit 1
        ;;
esac