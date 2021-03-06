#!/bin/bash
# Copyright (C) 2009 all rights reserved 
# GNU GENERAL PUBLIC LICENSE VERSION 3.0
# Author bjarneh@ifi.uio.no
#
# Bash command completion file for godag
#
# Add this file somewhere where it gets sourced.
# If you have sudo power it can be dropped into
# /etc/bash_completion.d/. If not it can be sourced
# by one of your startup scripts (.bashrc .profile ...)

_gd(){

    local cur prev opts gd_long_opts gd_short_opts gd_short_explain
    # long options
    gd_long_opts="--help --version --list --print --sort --output --static --arch --dryrun --clean --dot --test --benchmarks --match --verbose --test-bin --fmt --rew-rule --tab --tabwidth --no-comments --external --quiet --lib --main --backend"
    # short options + explain
    gd_short_explain="-h[--help] -v[--version] -l[--list] -p[--print] -s[--sort] -o[--output] -S[--static] -a[--arch] -d[--dryrun] -c[--clean] -I -t[--test] -b[--bench] -m[--match] -V[--verbose] -f[--fmt] -e[--external] -q[--quiet] -L[--lib] -M[--main] -B[--backend]"
    # short options
    gd_short_opts="-h -v -l -p -s -o -S -a -d -c -I -t -b -m -V -f -q -B"
    gd_special="clean test"


    COMPREPLY=()
    
    prev="${COMP_WORDS[COMP_CWORD-1]}"
    cur="${COMP_WORDS[COMP_CWORD]}"
    
    if [[ "${cur}" == --* ]]; then
        COMPREPLY=( $(compgen -W "${gd_long_opts}" -- "${cur}") )
        return 0
    fi

    if [[ "${cur}" == -* ]]; then
        COMPREPLY=( $(compgen -W "${gd_short_opts}" -- "${cur}") )
        if [ "${#COMPREPLY[@]}" -gt 1 ]; then
            COMPREPLY=( $(compgen -W "${gd_short_explain}" -- "${cur}") )
        fi
        return 0
    fi
    if [[ "${cur}" == c* || "${cur}" == t* ]]; then
        COMPREPLY=( $(compgen -W "${gd_special}" -- "${cur}") )
    fi
    if [[ "${prev}" == -* ]]; then
        case "${prev}" in
            '-a' | '-arch' | '--arch' | '-arch=' | '--arch=')
                COMPREPLY=( $(compgen -W "arm 386 amd64" -- "${cur}") )
                return 0
                ;;
        esac
    fi
    if [[ "${prev}" == -* ]]; then
        case "${prev}" in
            '-B' | '-B=' | '-backend' | '--backend' | '-backend=' | '--backend=')
                COMPREPLY=( $(compgen -W "gc gccgo express" -- "${cur}") )
                return 0
                ;;
        esac
    fi
}

## directories only -d was a bit to strict
complete -o default -F _gd gd
