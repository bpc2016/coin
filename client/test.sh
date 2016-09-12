#!/bin/bash
        for i in `seq 1 100`;
        do
                ./client -t 4 -u $i  &
        done     