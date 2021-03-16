#!/bin/bash
base_dir=$(dirname ${BASH_SOURCE})
cd $base_dir
sdk_dir=$(pwd)
version=$(cat  ../../VERSION)
git_version=$(git rev-parse --short HEAD)
sdk_version="${version}-${git_version}"
cat $sdk_dir/setup.py | sed -e  "s@unknown@$version@g" | sed -e "s@Arena@Arena($sdk_version)@g"  > /tmp/setup.py
cd $sdk_dir
python3 /tmp/setup.py bdist_wheel
