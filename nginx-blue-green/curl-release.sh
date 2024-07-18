#!/bin/bash

DURATION=${1-0s}
curl localhost:8080/?sleep=$DURATION

echo