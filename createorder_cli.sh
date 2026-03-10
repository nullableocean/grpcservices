#!/bin/bash

CLIENT_BIN="./orderserviceclient/bin/ordercli"
USER_LOG="userservice/logs/logs.log"
MARKET_LOG="spotinstrument/logs/logs.log"

if [ ! -f "$CLIENT_BIN" ]; then
    cd ./orderserviceclient && make build && cd ..
fi

if [ ! -f "$USER_LOG" ]; then
    echo "Файл логов пользователя не найден: $USER_LOG"
    exit 1
fi

if [ ! -f "$MARKET_LOG" ]; then
    echo "Файл логов рынков не найден: $MARKET_LOG"
    exit 1
fi

# UUID пользователя из логов
USER_UUID=$(grep -o '"UUID":"[^"]*"' "$USER_LOG" | tail -4 | head -1 | sed 's/"UUID":"//;s/"//')

# UUID маркета из логов
MARKET_UUID=$(grep -o '"UUID":"[^"]*"' "$MARKET_LOG" | tail -1 | sed 's/"UUID":"//;s/"//')

if [ -z "$USER_UUID" ]; then
    echo "Не удалось найти UUID пользователя в $USER_LOG"
    exit 1
fi

if [ -z "$MARKET_UUID" ]; then
    echo "Не удалось найти UUID рынка в $MARKET_LOG"
    exit 1
fi

# вызываем клиент для создания оредров
echo "./orderserviceclient/bin/ordercli create -a localhost:8091 -u "$USER_UUID" -m "$MARKET_UUID" -p 55.12344555 -q 10 -t buy"
./orderserviceclient/bin/ordercli create -a localhost:8091 -u "$USER_UUID" -m "$MARKET_UUID" -p 55.12344555 -q 10 -t buy
