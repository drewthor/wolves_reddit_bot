#!/bin/bash

. gcp.config

gcloud functions call $testCloudFunction --data '{"topic":"$testCloudFunctionTopic","message":""}'
sleep 3s
gcloud functions logs read $testCloudFunction
