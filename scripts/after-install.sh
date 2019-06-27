#!/bin/bash

systemctl daemon-reload

if hash firewall-cmd 2>/dev/null; then
    firewall-cmd --zone=public --add-port=8080/tcp --permanent
    firewall-cmd --reload
fi
