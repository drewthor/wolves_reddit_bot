#!/bin/bash

. gcp.config

gcloud functions deploy $prodCloudFunction --env-vars-file .env --entry-point $prodCloudFunctionEntryPoint --runtime go111 --trigger-topic=$prodCloudFunctionTopic
