# Quay Store Query Logs

This application queries Quay usage logs for a repository and stores them in object storage. This is because quay.io maintains the logs for a limited period of time. Those that want access to logs over a longer period need to query for them and store them.

This application is designed to be run at regular intervals (e.g., once per day) to get previous logs and export them. This is useful to run as something like a Kubernetes CronJob.

Note, this application was written for the Helm project. The current storage location is a development environment while the org goes through a governance transition as part of joining the CNCF.