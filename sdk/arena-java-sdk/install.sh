#!/bin/bash
set -e
arena_pkg=$1
if [[ $arena_pkg == "" ]];then
	echo "Error: not set arena install package,please set it"
	exit 1
fi
tar -xf $arena_pkg -C /tmp
bash /tmp/arena-installer/install.sh
if ! java -version &> /dev/null;then
	echo "Error: failed to execute 'java -version',please make sure jdk is installed."
	exit 2
fi
mkdir -pv ~/.jars
rm -rf ~/.jars/arena-java-sdk.jar
cp out/artifacts/arena_java_sdk/arena-java-sdk.jar ~/.jars
if ! grep 'CLASSPATH.*/.jars/arena-java-sdk.jar' ~/.bashrc &> /dev/null;then
	if [[ $CLASSPATH == "" ]];then
		echo 'export CLASSPATH=.:$JAVA_HOME/lib/dt.jar:$JAVA_HOME/lib/tools.jar:~/.jars/arena-java-sdk.jar' >> ~/.bashrc
	else
		echo 'export CLASSPATH=$CLASSPATH:~/.jars/arena-java-sdk.jar' >> ~/.bashrc
	fi
fi
source ~/.bashrc
