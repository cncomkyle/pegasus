#!/bin/bash

master=$(curl -s -X GET http://127.0.0.1:10086/master)
curl -s -X POST http://${master}/project?proj=Lianjia-Crawler
for i in $(seq 600); do
    echo "======================================="
    curl -s -X GET http://${master}/project/status
    #curl -s -X GET http://${master}/rate
    sleep 1
done
