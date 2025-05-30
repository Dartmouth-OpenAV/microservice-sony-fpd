#!/bin/bash

# Set your environment variables before running
MICROSERVICE_URL="your.microservice.url"
DEVICE_FQDN="your.device.fqdn"

echo "Running Sony FPD Microservice Tests..."

# GET requests
curl -X GET "http://$MICROSERVICE_URL/$DEVICE_FQDN/power"
sleep 1

curl -X GET "http://$MICROSERVICE_URL/$DEVICE_FQDN/videoroute"
sleep 1

curl -X GET "http://$MICROSERVICE_URL/$DEVICE_FQDN/volume"
sleep 1

curl -X GET "http://$MICROSERVICE_URL/$DEVICE_FQDN/audiomute"
sleep 1

# PUT requests
curl -X PUT "http://$MICROSERVICE_URL/$DEVICE_FQDN/power" -H "Content-Type: application/json" -d '"on"'
sleep 10

curl -X PUT "http://$MICROSERVICE_URL/$DEVICE_FQDN/videoroute/1" -H "Content-Type: application/json" -d '"1"'
sleep 1

curl -X PUT "http://$MICROSERVICE_URL/$DEVICE_FQDN/volume/1" -H "Content-Type: application/json" -d '"30"'
sleep 1

curl -X PUT "http://$MICROSERVICE_URL/$DEVICE_FQDN/audiomute/1" -H "Content-Type: application/json" -d '"false"'
sleep 1

echo "Tests complete."
