#!/bin/bash


if ! [ -d /var/lib/backup-repo/ ]; then
    mkdir /var/lib/backup-repo
fi

if [ -f "/etc/systemd/system/backup-repo.service" ]; then
    systemctl stop backup-repo
    systemctl disable backup-repo
    systemctl daemon-reload
fi
