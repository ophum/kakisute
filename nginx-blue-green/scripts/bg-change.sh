#!/bin/bash


CONFD=/etc/nginx/conf.d
SITES_AVAILABLE=/etc/nginx/sites-available

function change() {
    echo "$1 -> $2"
    rm $CONFD/$1.conf
    rm $CONFD/$2-preview.conf

    ln -s $SITES_AVAILABLE/$2.conf $CONFD/$2.conf
    ln -s $SITES_AVAILABLE/$1-preview.conf $CONFD/$1-preview.conf

    nginx -s reload
}

current=$(bash /etc/nginx/scripts/bg-current.sh)
if [ $current == "blue" ]; then
    change blue green
else
    change green blue
fi
