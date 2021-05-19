#!/usr/bin/env bash

while :
do
    if [ $(pgrep -f "sudo ./server.py") ];then
        sleep 5
    else
        sudo ./server.py -d ./users/ ./imgs/ ./pwd.txt 1000 2>&1 > ./log.txt &
    fi
done