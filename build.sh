#!/bin/bash
# Copyright (C) 2009 all rights reserved 
# GNU GENERAL PUBLIC LICENSE VERSION 3.0
# Author bjarneh@ifi.uio.no

COMPILER=""
LINKY=""
D=`dirname "$0"`
B=`basename "$0"`
FULL="`cd \"$D\" 2>/dev/null && pwd || echo \"$D\"`/$B"
HERE=$(dirname "$FULL")
IDIR=$HERE/src
CPROOT=`date +"tmp-pkgroot-%s"`
SRCROOT="$GOROOT/src/pkg"

# array to store packages which are pure go
declare -a package;

# this is done statically for now, no grepping
# to figure out which packges are actually pure go..
package=(
'archive'
'asn1'
'bufio'
'cmath'
'compress'
'container'
'ebnf'
'encoding'
## 'exec'
'expvar'
'flag'
'fmt'
'go'
'hash'
'http'
'image'
'io'
'json'
'log'
'mime'
'netchan'
'nntp'
'once'
'patch'
'path'
'rand'
'reflect'
'regexp'
'rpc'
'scanner'
'sort'
'strconv'
'strings'
'syslog'
'tabwriter'
'template'
'testing'
'unicode'
'unsafe'
'utf16'
'utf8'
'websocket'
'xml'
)


function build(){
    echo "build"
    cd src/utilz && $COMPILER walker.go || exit 1
    $COMPILER handy.go || exit 1
    $COMPILER stringset.go || exit 1
    $COMPILER stringbuffer.go || exit 1
    cd $HERE/src/parse && $COMPILER -o gopt.$OBJ option.go gopt.go || exit 1
    cd $HERE/src/cmplr && $COMPILER -I $IDIR dag.go || exit 1
    $COMPILER -I $IDIR compiler.go || exit 1
    cd $HERE/src/start && $COMPILER -I $IDIR main.go || exit 1
    cd $HERE && $LINKY -o gd -L src src/start/main.? || exit 1
}

function clean(){
    echo "clean"
    cd $HERE
    rm -rf src/utilz/walker.?
    rm -rf src/utilz/stringset.?
    rm -rf src/utilz/stringbuffer.?
    rm -rf src/utilz/utilz_test.?
    rm -rf src/utilz/handy.?
    rm -rf src/cmplr/dag.?
    rm -rf src/cmplr/compiler.?
    rm -rf src/parse/gopt.?
    rm -rf src/parse/gopt_test.?
    rm -rf src/parse/option.?
    rm -rf src/start/main.?
    rm -rf gd
    rm -rf "$HOME/bin/gd"
    rm -rf "$GOBIN/gd"
}

function phelp(){
cat <<EOH

build.sh - utility script for godag

targets:

  help    : print this menu and exit
  clean   : rm *.[865a] from src + rm gd \$HOME/bin/gd
  build   : compile source code in ./src
  install : build + mv gd \$HOME/bin  (default)
  cproot  : copy modified (pure go) part of \$GOROOT/src/pkg

EOH
}

function die(){
    echo "variable: $1 not set"
    exit 1
}


# recursively copy all the $GOROOT/src/pkg to $CPROOT,
# with a *.go filter, any test that includes testdata will fail.
# NOTE main packages are also removed, these are used for testing
# and since too many of these end up in the same name-space, they
# are all removed here..
function recursive_copy {

    mkdir "$2"

    for i in $(ls "$1");
    do
        if [ -f "$1/$i" ]; then
            case "$i" in *.go)
                grep "^package main$" -q "$1/$i" || cp "$1/$i" "$2/$i"
            esac
        fi

        if [ -d "$1/$i" ]; then
            if [ ! "$i" == "testdata" ];then
                recursive_copy "$1/$i" "$2/$i"
            fi
        fi
    done

    return 1
}


# move all go packages up one level, and give them
# a fitting header based on directory..
function up_one_level {

    for element in $(ls $1);
    do
        if [ -f "$1/$element" ]; then
            mv "$1/$element" "${1}/${2}_${element}"
            mv "${1}/${2}_${element}" "${1}/.."
        fi

        if [ -d "$1/$element" ]; then
            up_one_level "$1/$element" "$element"
        fi

    done

    return 1
}

function cproot {

    mkdir "$CPROOT";
    echo "cp *.go: \$GOROOT/src/pkg  ->  $CPROOT"
    echo "this may take some time..."

    for p in "${package[@]}";
    do
        recursive_copy "$SRCROOT/$p" "$CPROOT/$p"
    done

    up_one_level "$CPROOT" "$CPROOT"

    # delete empty directories from $CPROOT
    find -depth -type d -empty -exec rmdir {} \;

    exit 0
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
	OBJ="8"
    ;;
    'arm')
    COMPILER="5g"
    LINKY="5l"
	OBJ="5"
    ;;
    'amd64')
    COMPILER="6g"
    LINKY="6l"
	OBJ="6"
    ;;
    *)
    echo "architecture not: 'amd64' '386' 'arm'"
    echo "architecture was ${GOARC}"
    exit 1
    ;;
esac


case "$1" in
     'help' | '-h' | '--help' | '-help')
      phelp
      ;;
      'cproot' | '--cproot' | '-cproot')
      time cproot
      ;;
      'clean' | 'c' | '-c' | '--clean' | '-clean')
      time clean
      ;;
      'build' | 'b' | '-b' | '--build' | '-build')
      time build
      ;;
      *)
      time build
      if [ -d "${HOME}/bin" ]; then
          cd "$HERE"
          mv gd "$HOME/bin"
      else
          if [ -d "$GOBIN" ]; then
              cd "$HERE"
              mv gd "$GOBIN"
          else
              echo -e "\n[ERROR] ${HOME}/bin: not a directory"
              echo -e "[ERROR] \$GOBIN: not set\n"
          fi
      fi
      ;;
esac
}
