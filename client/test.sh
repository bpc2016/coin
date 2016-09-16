#!/bin/bash

# with no parameters - start 3 clients each with 2 tosses per roll
# ./test.sh C  - start C clients .. 2 tosses
# ./test.shh C T  - start C clients ... T tosses 
# ./test.shh C T I - start C clients ... T tosses  ... server I

        if test $# -gt 0; then
                COUNTER=$1
        else
                COUNTER=3
        fi
        TOSSES=2
        shift
        if test $# -gt 0; then
                TOSSES=$1
        fi
        INDEX=0
        shift
        if test $# -gt 0; then
                INDEX=$1
        fi
        until [  $COUNTER -lt 1 ]; do
                ./client -t $TOSSES -u $COUNTER -s $INDEX &
                let COUNTER-=1
        done

        # for i in `seq 1 100`;
        # do
        #         ./client -t 4 -u $i  &
        # done     