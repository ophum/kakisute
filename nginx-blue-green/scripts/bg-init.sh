#!/bin/bash

CONFD=/etc/nginx/conf.d
SITES_AVAILABLE=/etc/nginx/sites-available

if [ -e "$CONFD/default.conf" ]; then
    rm $CONFD/default.conf

    ln -s $SITES_AVAILABLE/blue.conf $CONFD/blue.conf
    ln -s $SITES_AVAILABLE/green-preview.conf $CONFD/green-preview.conf
    nginx -s reload
fi