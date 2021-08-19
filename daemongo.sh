#!/usr/bin/env bash

while :
do
    if ![ $(pgrep -f "./server") ];then
        ~/run_icqs.sh
    fi
    sleep 10
done
