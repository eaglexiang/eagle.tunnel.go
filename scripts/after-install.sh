#!/bin/bash

systemctl daemon-reload
firewall-cmd --add-port=8080/tcp --permanent
firewall-cmd --reload