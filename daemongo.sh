#!/usr/bin/env bash

while :
do
    if [ $(pgrep -f "./server") ];then
        sleep 10
    else
        ~/run_icqs.sh
    fi
done