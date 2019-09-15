#!/bin/bash

. gcp.config

gcloud functions call $testCloudFunction --data '{"topic":"$testCloudFunctionTopic","message":""}'
gcloud functions logs read $testCloudFunction
