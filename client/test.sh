#!/bin/bash
        for i in `seq 11 30`;
        do
                ./client -t 3 -u $i  &
        done     