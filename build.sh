#!/bin/bash
# Copyright (C) 2009 all rights reserved 
# GNU GENERAL PUBLIC LICENSE VERSION 3.0
# Author bjarneh@ifi.uio.no

COMPILER=""
LINKY=""
HERE=$(dirname $(readlink -f $0))
IDIR=$HERE/src


function build(){
    echo "build"
    cd src/utilz && $COMPILER walker.go || exit 1
    $COMPILER handy.go || exit 1
    $COMPILER stringset.go || exit 1
    cd $HERE/src/parse/ && $COMPILER option.go gopt.go || exit 1
    cd $HERE/src/cmplr && $COMPILER -I $IDIR dag.go || exit 1
    $COMPILER -I $IDIR compiler.go || exit 1
    cd $HERE/src/start && $COMPILER -I $IDIR main.go || exit 1
    cd $HERE && $LINKY -o gd src/start/main.? || exit 1
}

function clean(){
    echo "clean"
    cd $HERE
    rm -rf src/utilz/walker.?
    rm -rf src/utilz/stringset.?
    rm -rf src/utilz/handy.?
    rm -rf src/cmplr/dag.?
    rm -rf src/cmplr/compiler.?
    rm -rf src/parse/gopt.?
    rm -rf src/parse/option.?
    rm -rf src/start/main.?
    rm -rf gd
}

function phelp(){
cat <<EOH

compile go source

targets:

  clean
  build (default)

EOH
}

function die(){
    echo "variable: $1 not set"
    exit 1
}


# main
{
[ "$GOROOT" ] || die "GOROOT"
[ "$GOARCH" ] || die "GOARCH"
[ "$GOOS" ]   || die "GOOS"

case "$GOARCH" in
    '386')
    COMPILER="8g"
    LINKY="8l"
    ;;
    'arm')
    COMPILER="5g"
    LINKY="5l"
    ;;
    'arm64')
    COMPILER="6g"
    LINKY="6l"
    ;;
    *)
    echo "architecture not: 'arm64' '386' 'arm'"
    echo "architecture was ${GOARC}"
    exit 1
    ;;
esac


case "$1" in
     'help' | '-h' | '--help' | '-help')
      phelp
      ;;
      'clean' | 'c' | '-c' | '--clean' | '-clean')
      time clean
      ;;
      *)
      time build
      ;;
esac
}
