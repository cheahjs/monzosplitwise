#!/bin/bash

while true; do 
    echo "Running at $(date)."
    go run app/main.go
    sleep 300
done