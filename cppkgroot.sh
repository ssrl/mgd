#!/bin/bash
# Copyright (C) 2009 all rights reserved 
# GNU GENERAL PUBLIC LICENSE VERSION 3.0
# Author bjarneh@ifi.uio.no

function phelp(){
cat <<EOH

create a copy of the standard library
with some modifications to be able to 
compile it without makefiles.

EOH
}


case $1 in
    '-h' | '-help' | '--help' | 'help')
    phelp
    exit 0
esac


# this is the name of our copy of the
# pure go source root
CPROOT=`date +"tmp-pkgroot-%s"`
mkdir "$CPROOT";


if [ ! "$GOROOT" ]; then
    echo "Missing \$GOROOT variable, will die now.."
    exit 1
fi

# $GOROOT
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
'crypto'
'ebnf'
'encoding'
'exec'
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
            case $i in *.go)
                grep "^package main$" -q "$1/$i" || cp "$1/$i" "$2/$i"
            esac
        fi

        if [ -d "$1/$i" ]; then
            recursive_copy "$1/$i" "$2/$i"
        fi
    done

    return 1
}



for p in "${package[@]}";
do
    recursive_copy "$SRCROOT/$p" "$CPROOT/$p"
done


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


up_one_level "$CPROOT" "$CPROOT"


# delete empty directories from $CPROOT
find -depth -type d -empty -exec rmdir {} \;
