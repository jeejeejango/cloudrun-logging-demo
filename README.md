# Cloud Run Logging Demo with Pub/Sub

## Create a Pub/Sub topic

This service is triggered by messages published to a Pub/Sub topic, so you will
need to create a topic in Pub/Sub.

```shell
gcloud pubsub topics create demo-logging-topic
```

You can use `demo-logging-topic` or replace with a topic name unique in your
Cloud Project.

## Shipping the code

Shipping code consists of three steps: building a container image with Cloud Build,
uploading the container image to Container Registry, and deploying the container
image to Cloud Run.

### Building the code

```shell
gcloud builds submit --tag gcr.io/<PROJECT_ID>/demo-cloudrun-logging
```

Where `<PROJECT_ID>` is your Cloud project ID, and `demo-cloudrun-logging` is the image name.

Upon success, you should see a SUCCESS message containing the ID, creation time,
and image name. The image is stored in Container Registry and can be re-used if
desired.

### Deploy your application

```shell
gcloud run deploy demo-cloudrun-logger --image gcr.io/<PROJECT_ID>/demo-cloudrun-logging  \
 --no-allow-unauthenticated \
 --update-env-vars PROJECT_ID=<PROJECT_ID>
```

Replace `<PROJECT_ID>` with your Cloud project ID. `demo-cloudrun-logging` is the image name and
`demo-cloudrun-logger` is the name of the service. Notice that the container image is deployed to the
service and region that you configured previously under Setting up gcloud.

The `--no-allow-unauthenticated` flag restricts unauthenticated access to the service. By
keeping the service private you can rely on Cloud Run's automatic Pub/Sub integration to
authenticate requests.

Wait until the deployment is complete: this can take about half a minute. On success, the command
line displays the service URL. This URL is used to configure a Pub/Sub subscription.

## Integrating with Pub/Sub

To integrate the service with Pub/Sub:

Create or select a service account to represent the Pub/Sub subscription identity.

```shell
gcloud iam service-accounts create demo-cloudrun-logging-invoker \
    --display-name "Demo Cloud Run Logging Pub/Sub Invoker"
```

You can use `demo-cloudrun-logging-invoker` or replace with a name unique within your Cloud project.

Create a Pub/Sub subscription with the service account:

Give the invoker service account permission to invoke your `demo-cloudrun-logger` service

```shell
gcloud run services add-iam-policy-binding demo-cloudrun-logger \
--member=serviceAccount:demo-cloudrun-logging-invoker@<PROJECT_ID>.iam.gserviceaccount.com \
--role=roles/run.invoker
```

Replace `<PROJECT_ID>` with your Cloud project ID. It can take several minutes for the IAM changes
to propagate. In the meantime, you might see HTTP 403 errors in the service logs.

Allow Pub/Sub to create authentication tokens in your project:

```shell
gcloud projects add-iam-policy-binding <PROJECT_ID> \
   --member=serviceAccount:service-<PROJECT_NUMBER>@gcp-sa-pubsub.iam.gserviceaccount.com \
   --role=roles/iam.serviceAccountTokenCreator
```

Replace:

* `<PROJECT_ID>` with your Google Cloud project ID.
* `<PROJECT_NUMBER>` with your Google Cloud project number.

Create a Pub/Sub subscription with the service account:

```shell
gcloud pubsub subscriptions create demo-logging-topic-sub --topic demo-logging-topic \
--ack-deadline=600 \
--push-endpoint=<SERVICE-URL>/ \
--push-auth-service-account=demo-cloudrun-logging-invoker@<PROJECT_ID>.iam.gserviceaccount.com
```

Replace:

* `demo-logging-topic` with the topic you previously created.
* `<SERVICE-URL>` with the HTTPS URL provided on deploying the service. This 
   URL works even if you have also added a domain mapping.
* `<PROJECT_ID>` with your Cloud project ID.

The `--push-auth-service-account` flag activates the Pub/Sub push functionality for 
Authentication and authorization.

Your Cloud Run service domain is automatically registered for use with Pub/Sub 
subscriptions.

For Cloud Run only, there is a built-in authentication check that the token is 
valid and an authorization check that the service account has permission to invoke 
the Cloud Run service.

## Trying it out
Configure Logstash pipeline to send the log message to the Pub/Sub topic:
```
    google_pubsub {
        project_id => "<PROJECT_ID>"
        topic => "demo-logging-topic"
    }
```
Replace `<PROJECT_ID>` with your Cloud project ID. This sample is using 
[ADC](https://cloud.google.com/docs/authentication/provide-credentials-adc) and it's
safer than using JSON API key file.
