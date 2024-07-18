#!/bin/bash

if [ -e "/etc/nginx/conf.d/blue.conf" ]; then
    echo "blue"
elif [ -e "/etc/nginx/conf.d/green.conf" ]; then
    echo "green"
else
    echo "unknown"
fi