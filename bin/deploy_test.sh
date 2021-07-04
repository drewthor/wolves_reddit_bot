#!/bin/bash

. gcp.config

gcloud functions deploy $testCloudFunction --env-vars-file .env --entry-point $testCloudFunctionEntryPoint --runtime go111 --trigger-topic=$testCloudFunctionTopic
