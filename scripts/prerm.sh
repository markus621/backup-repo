#!/bin/bash


if [ -f "/etc/systemd/system/backup-repo.service" ]; then
    systemctl stop backup-repo
    systemctl disable backup-repo
    systemctl daemon-reload
fi
