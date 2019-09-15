#!/bin/bash

. gcp.config

gcloud functions deploy $testCloudFunction --env-vars-file .env.yaml --entry-point $testCloudFunctionEntryPoint --runtime go111 --trigger-topic=$testCloudFunctionTopic
