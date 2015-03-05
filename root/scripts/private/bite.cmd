#!/usr/bin/env bash

po=$1
se=$2
ch=$3
us=$4
da=$5

size=$(echo "$RANDOM % 20 + 3" | bc)
bite="B"
while [ $size -gt 0 ]
do
    bite="${bite}="
    size=$(echo $size - 1 | bc)
done
bite="${bite}D"

echo "$se 1 PRIVMSG $us :${us}> $bite $da" | nc 127.0.0.1 $po
