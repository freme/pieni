#!/bin/bash

URL=""
if [ "" = "$1" ];then
    docker run --detach --name pieni --publish 3000:3001 pieni
    URL="http://localhost:3000"   
else
    URL=$1
fi

LIST="1 2 3 4 5 6 7 8 9 10 11 12 13 14 15 20"

if [ ! -d benchmark ]; then
    mkdir benchmark
fi

cd benchmark

for SIZE in $LIST; do
	if [ ! -f ${SIZE}k.txt ]; then
		base64 /dev/urandom | head -c $(($SIZE*1024)) >${SIZE}k.txt
	fi
done

du -hs *.txt

for SIZE in $LIST; do
	echo "Upload ${SIZE}k.txt"
	curl --user user:user -T ${SIZE}k.txt ${URL}/benchmark/${SIZE}k.txt
done

cd ..

if [ "" = "$1" ];then
    docker stop pieni
    docker rm pieni
fi
