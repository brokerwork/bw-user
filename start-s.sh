#!/bin/bash
name=$1
env "consul.endpoint=$CONSUL_IP:8500"  "service.report_ip=$NODE_IP" ./$name app.toml