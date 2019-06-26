#!/bin/bash

systemctl daemon-reload

if hash firewall-cmd 2>/dev/null; then
    firewall-cmd --permanent --add-service=eagle-tunnel-client
    firewall-cmd --permanent --add-service=eagle-tunnel-server
    firewall-cmd --reload
fi