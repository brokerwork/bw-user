#!/bin/bash

ProcessName="bw-user"
ProcessCount=`ps -ef|grep ${ProcessName} | grep -v log | grep -v grep | wc -l`
if [ ${ProcessCount} -lt 1 ]; then
	echo "Begin Start ${ProcessName} "
	DIR="$( cd "$( dirname "$0"  )" && pwd  )"
	cd ${DIR}
	nohup ./${ProcessName} &
	
	sleep 1
	ProcessCount=`ps -ef|grep ${ProcessName} | grep -v log | grep -v grep | wc -l`
	if [ ${ProcessCount} -lt 1 ]; then
		echo "Start ${ProcessName} Failed"
		exit 1
	fi
else
	echo "${ProcessName} have exist"
fi

exit 0
