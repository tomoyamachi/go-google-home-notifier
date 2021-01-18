# Google Home Notifier

Send notifications to Google Home.

- Run a notification listener server
- Notify next schedule on Google Calendars (Support multi accounts)
- Fetch schedules from Google Calendars (Support multi accounts)

## Usage

### Send notifications from CLI

```
$ notify notify --locale en --message "Sample notification"
```

### Run a simple notification server

```
$ notify server --port 8000
```

You can send notification to Google Home devices by `curl -X POST -d "Sample Message" localhost:8000`.

## Daemon mode

Daemon mode provides following feature:

- Run a notification listener server
- Notify next schedule on Google Calendars (Support multi accounts)

You can run daemon after regists google accounts:

```
notify daemon 
```

### Regists a Google account to CLI tools

#### 1. Enable the API and create your OAuth client

1. Go to [this page](https://developers.google.com/calendar/quickstart/go) and click `Enable the Google Calendar API`
2. `Configure your OAuth client` > `Desktop app` > `CREATE` 
3. Download a client configuration file to same directory as `credentials.json`.

#### 2. Register a Google account token

1. Run `notify calendar add-token` then show OAuth URL.
2. Go to the link and authorize.
3. Input authorization code to terminal.
4. Create or modify `tokens.json`