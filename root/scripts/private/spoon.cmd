#!/usr/bin/env bash

po=$1
se=$2
ch=$3
us=$4

echo "$se 1 PRIVMSG $us :th3r3 1s n0 sp0on..." | nc 127.0.0.1 $po
