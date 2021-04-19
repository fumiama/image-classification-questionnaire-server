#!/usr/bin/env bash

while :
do
    if [ $(pgrep -f "sudo nohup ./server.py") ];then
        sleep 1
    else
        sudo nohup ./server.py ./users/ ./imgs/ ./pwd.txt 1000 2>&1 > ./log.txt &
    fi
done