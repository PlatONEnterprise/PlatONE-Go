#!/bin/bash
list=`cat ../conf/node_list`
cd ../..
workdir=`pwd`
cd -
bindir=`pwd`
#workdir=`pwd`/../..

if [ "$1" = "stopall" ];
then
	for i in ${list[*]}
	do   
	  echo "run on $i"
	  ssh $i "pkill -9 platone"
	  ssh $i "ps -ef|grep platone"
	done
fi


if [ "$1" = "resetall" ];
then
	for i in ${list[*]}
	do   
	  echo "run on $i"
	  scp platone $i:${workdir}/.
	  ssh $i "cd ${workdir} && pkill -9 platone"
	  ssh $i "ps -ef|grep platone"
	  ssh $i "cd ${workdir} && ./reset.sh"
	done
fi


if [ "$1" = "rmtpslog" ];
then
	for i in ${list[*]}
	do   
	  ssh $i "rm ${workdir}/*.log"
	done
fi



if [ "$1" = "startall" ];
then
	m=0
	for i in ${list[*]}
	do   
	  echo "run on $i"
	  a=(${i//./ })
	  ssh $i bash ${workdir}/start.sh ${a[-1]}
          ((m=$m+1))
	done
fi

if [ "$1" = "stopfirewall" ];
then
	for i in ${list[*]}
	do   
	  ssh $i "sudo service firewalld stop"
	done
fi


if [ "$1" = "status" ];
then
	m=0
	for i in ${list[*]}
	do   
	  ssh $i "ps -ef|grep platone"
          ((m=$m+1))
	done
fi


if [ "$1" = "checkround" ];
then
	tailf node5.log |grep -i "roundChange round:"
fi


if [ "$1" = "upd" ];
then
	for i in ${list[*]}
	do
	  echo send to $i
	  scp platone $i:${workdir}/.
	done 
fi

if [ "$1" = "pkg" ];
then
	m=0
	for i in ${list[*]}
	do
	  [ $m -eq 0 ] && ((m=m+1)) && continue
	  echo send to $i
	  ssh $i "mkdir -p /home/wxuser/node_cluster"
	  ssh $i "rm -rf ${workdir}/*"
	  scp -r ${workdir}/* $i:/home/wxuser/node_cluster/
	done 
fi

if [ "$1" = "-f" ];
then
	for i in ${list[*]}
	do
	  echo send to $i
	  scp -r ${bindir}/$2 $i:${bindir}
	done 
fi

if [ "$1" = "cleanlog" ];
then
	for i in ${list[*]}
	do
 	  ssh $i 'for m in `find '${workdir}' -name "*platone.log"`;do rm $m ;done;'
          ssh $i 'for m in `find '${workdir}' -name "*wasm.log"`;do rm $m ;done;'
	done 
fi

if [ "$1" = "cleanall" ];
then
	for i in ${list[*]}
	do
 	  ssh $i 'cd '${bindir}' && bash clean.sh'
	done 
fi


