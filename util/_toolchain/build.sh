#!/bin/bash

set -eo pipefail

version=1.6
shasum=5470eac05d273c74ff8bac7bef5bad0b5abbd1c4052efbdbc8db45332e836b0b
pkg="go${version}.linux-amd64"
go=go/bin/go

test -f ${go} && ${go} version | grep -q ${version} && exit

rm -rf go.tar.gz go

curl -o go.tar.gz -L "https://storage.googleapis.com/golang/${pkg}.tar.gz"
echo "${shasum}  go.tar.gz" | shasum -c -

tar xzf go.tar.gz
rm go.tar.gz
