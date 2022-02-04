#!/bin/sh

TEMPDIR=$(mktemp -d)
go run main.go > $TEMPDIR/test-stats.md


cd  $TEMPDIR
echo -n ${G_TOKEN} | gh auth login --with-token
git clone git@gist.github.com:102d08389c61e7f7293570bf38205d3b.git
cd 102d08389c61e7f7293570bf38205d3b
mv ../test-stats.md ./test-stats.md
git commit -m "updated `date`" test-stats.md
git push origin master