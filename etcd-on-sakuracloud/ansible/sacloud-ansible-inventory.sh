#!/bin/bash

if [ ! -e sacloud_inventory.py ]; then
    curl -LOs https://raw.githubusercontent.com/sakura-internet/sacloud-ansible-inventory/refs/heads/master/sacloud_inventory.py
    chmod +x sacloud_inventory.py
fi

./sacloud_inventory.py