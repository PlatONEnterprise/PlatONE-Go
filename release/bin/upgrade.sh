#!/bin/sh

if [ -L ../../platone_release ]; then
    rm ../../platone_release
fi

path=`echo $(dirname $(pwd))`
echo "New PlatONE release path: " $path
ln -s $path ../../platone_release