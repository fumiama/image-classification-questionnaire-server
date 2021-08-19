#!/usr/bin/env bash

while :
do
    if [ -z $(pgrep -f "sudo ./server.py") ];then
        sudo ./server.py -d ./users/ ./imgs/ ./pwd.txt 1000 2>&1 > ./log.txt &
    fi
    sleep 10
done
