#!/bin/sh
SPECS=pkg/avs/specs
MANIFEST=manifest

DIR=`pwd`
cd $SPECS

if [ -f $MANIFEST ]; then
    rm -f $MANIFEST
fi

touch $MANIFEST
for file in `find ./ -type f ! -name $MANIFEST | sed 's|^./||' | sort`; do
    sha1sum $file >> $MANIFEST
done

cd $DIR
