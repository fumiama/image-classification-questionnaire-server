#!/usr/bin/env bash

while :
do
    if [ -z $(pgrep -f "./server") ];then
        ~/run_icqs.sh
    fi
    sleep 10
done
