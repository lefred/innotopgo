#!/bin/bash

make genall
version=$(cat innotop/innotop.go | grep version | grep ':=' | cut -d'"' -f2)
cd build
for i in $(ls | grep -v 'tar.gz')
do
  tar czvf "$i-$version.tar.gz" $i
done
cd -

