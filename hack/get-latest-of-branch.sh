#!/usr/bin/env bash

# Get branch to work on
DEFAULT_BRANCH=runai-master
if [ $# -eq 0 ] 
  then
    echo "Using default branch $DEFAULT_BRANCH"
    BRANCH=$DEFAULT_BRANCH
else
    BRANCH=$1
fi

# Create temp folder
TMP_FOLDER=/tmp/runai-cli
if [ -d $TMP_FOLDER ] 
  then
    rm -rf $TMP_FOLDER
fi
mkdir $TMP_FOLDER

echo "Getting latest revision for branch $BRANCH"
gsutil cp gs://cli-artifacts/branch-versions/$BRANCH $TMP_FOLDER
REVISION=$(cat $TMP_FOLDER/$BRANCH)
echo "Downloading revision $REVISION"
ARCHIVE_VERSION=darwin-amd64.tar.gz
gsutil cp gs://cli-artifacts/$REVISION/$ARCHIVE_VERSION $TMP_FOLDER
echo "Unarchiving version"
INSTALL_FOLDER=$TMP_FOLDER/install
mkdir $INSTALL_FOLDER
tar -C $INSTALL_FOLDER -zxvf $TMP_FOLDER/$ARCHIVE_VERSION
echo "Installing version"
$INSTALL_FOLDER/install-runai.sh
echo "Installation complete"