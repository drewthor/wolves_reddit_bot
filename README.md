# wolves_reddit_bot
Reddit bot for /r/Timberwolves

# Getting Started
Create a file named reddit_config.json as a copy of dummy_reddit_config.json and fill in the fields with the info found from your reddit bot account's api access information page

Create a file named gcp.config as a copy of gcp_TEMPLATE.config and fill in the fields with the info of your google cloud functions, their topics, and the name of the entry point's function in this pacakage (unless you rename it the entry point is Receive)

Download gcloud cli tool and configure/login to your Google Cloud Platform account.

# Deploying and testing
I have the cloud function set up to run using Google Cloud Scheduler in Google Cloud Platform; works well and stays within the free tier as well.

You can run the program locally building and running cmd/main.go

Deploy using the scripts found in bin/

Test either locally or utilizing bin/test.sh after deploying bin/deploy_test.sh
