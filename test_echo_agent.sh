#!/bin/bash
echo "Echo agent started (PID: $$)"
while read -r line; do
    echo "ECHO: $line"
done
