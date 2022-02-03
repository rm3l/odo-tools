#!/bin/sh

TEMPDIR=$(mktemp -d)
python graph.py
mv temp/graph.png $TEMPDIR

pushd $TEMPDIR
git clone git@gist.github.com:1ea5606207f6141af21c7c3b0d527635.git
pushd 1ea5606207f6141af21c7c3b0d527635
mv ../graph.png ./graph.png
git commit -m "updated graph" graph.png
git push origin master
popd
popd

rm -rf $TEMPDIR
