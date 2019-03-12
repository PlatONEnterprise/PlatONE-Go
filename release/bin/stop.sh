#!/bin/bash

pids=`ps -ef | grep platon | grep -v grep | awk '{print $2}'`

for pid in $pids
do
    echo "Stopping platon["$pid"]..."
    kill $pid
done

if [ $pids"x" != "x" ]; then
    while true
    do
        pids=`ps -ef | grep platon | grep -v grep | awk '{print $2}'`
        if [ $pids"x" = "x" ]; then
            break
        else
            sleep 1
        fi
    done
    echo "Stop platon succ"
else
    echo "Not found platon precess"
fi
