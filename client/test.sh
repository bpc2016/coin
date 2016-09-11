#!/bin/bash
        for i in `seq 11 40`;
        do
                ./client $i  &
        done    