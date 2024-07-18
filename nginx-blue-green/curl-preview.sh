#!/bin/bash

DURATION=${1-0s}
curl -H "Host: preview" localhost:8080/?sleep=$DURATION

echo