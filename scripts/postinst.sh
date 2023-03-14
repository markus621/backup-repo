#!/bin/bash


if [ -f "/etc/systemd/system/backup-repo.service" ]; then
    systemctl start backup-repo
    systemctl enable backup-repo
    systemctl daemon-reload
fi
