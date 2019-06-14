#!/bin/bash

#assumes that the gcloud function command name is wolves_reddit_bot with a schedule set up to trigger on "postGameThread"
gcloud functions deploy wolves_reddit_bot --env-vars-file .env.yaml --entry-point Receive --runtime go111 --trigger-topic=postGameThread
