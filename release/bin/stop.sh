#!/bin/bash

pids=`ps -ef | grep platone | grep -v grep | awk '{print $2}'`

for pid in $pids
do
    echo "Stopping platone["$pid"]..."
    kill $pid
done

if [ $pids"x" != "x" ]; then
    while true
    do
        pids=`ps -ef | grep platone | grep -v grep | awk '{print $2}'`
        if [ $pids"x" = "x" ]; then
            break
        else
            sleep 1
        fi
    done
    echo "Stop platone succ"
else
    echo "Not found platone precess"
fi
