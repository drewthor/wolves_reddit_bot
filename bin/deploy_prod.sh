#!/bin/bash

. gcp.config

gcloud functions deploy $prodCloudFunction --env-vars-file .env.yaml --entry-point $prodCloudFunctionEntryPoint --runtime go111 --trigger-topic=$prodCloudFunctionTopic
