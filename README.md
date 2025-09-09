# GCP to Discord: Google Cloud Alerting to Discord Webhook

⚙️ A simple, yet powerful, Google Cloud Function written in Go that acts as a proxy to transform and forward Google Cloud Platform (GCP) alerting notifications to Discord via webhooks. This solution enables real-time monitoring and incident response by centralizing your GCP alerts directly within your team's Discord channels.

_This project has been tested with the Go 1.23.3 runtimes._

![Notification in Discord](screenshot.png "Notification in Discord")

## Getting Started

### Prerequisites

- Ensure you have `gcloud` installed:
  - MacOS: `brew cask install google-cloud-sdk`
  - Others: <https://cloud.google.com/sdk/gcloud>
- Ensure you have authenticated with Google Cloud: `gcloud init`
- (Optional) Set your current working project: `gcloud config set project <project>`

### Discord Webhook Setup

1. In your Discord server, go to **Server Settings** > **Integrations** > **Webhooks**.
2. Click **New Webhook** and configure its name and channel.
3. Copy the generated **Webhook URL**. This will be used as `DISCORD_WEBHOOK_URL` in your `env.yaml`.

### Deployment

1. Clone / download a copy of this repository
2. Copy `env.sample.yaml` to `env.yaml` and modify the environment variables declared in `env.yaml`.
3. The Google Cloud Function relies on environment variables defined in `env.yaml` (copied from `env.sample.yaml`).
   1. **`AUTH_TOKEN`**: A secret token used to authenticate incoming webhook requests from GCP. This should be a strong, randomly generated string
   2. **`DISCORD_WEBHOOK_URL`**: The full URL of your Discord channel's webhook, obtained during the Discord Webhook Setup.
4. Run `./deploy.sh`

### GCP Alerting Setup

1. Navigate to [GCP Monitoring Alerting](https://console.cloud.google.com/monitoring/alerting/notifications) in your Google Cloud Console.
2. Create a new Notification Channel of type "Webhook".
3. For the "Endpoint URL", use the URL of your deployed Google Cloud Function, appending `?auth_token=<YOUR_AUTH_TOKEN>` to the end. For example: `https://<REGION>-<PROJECT_ID>.cloudfunctions.net/gcp-to-discord?auth_token=YOUR_AUTH_TOKEN`.
4. Ensure the `AUTH_TOKEN` matches the one you set in `env.yaml`.

### Local Testing

**Local Function Invocation**: You can use the `functions-framework-go` to run the function locally.

    ```bash
    go run github.com/GoogleCloudPlatform/functions-framework-go/cmd/functions-framework --target GcpToDiscord --port 8080
    ```

    Then, you can send a POST request to `http://localhost:8080` with a sample GCP alert payload.

## License

This project is licensed under the Apache License 2.0 - see the [LICENSE](LICENSE) file for details.
