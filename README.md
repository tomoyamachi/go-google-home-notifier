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
3. Copy a authorization code from an URL parameter: `http://localhost/?state=state-token&code=<authz-code-is-here>&scope=https://www.googleapis.com/auth/calendar.readonly`
4. Input authorization code to terminal.
5. Create or modify `tokens.json`

## Run as daemon

```
$ chmod +x notify
$ mv notify /usr/local/sbin/google-home-notifier

$ mkdir /etc/google-home-notifier
$ mv google-credentialfile.json /etc/google-home-notifier/credentials.json
$ google-home-notifier calendar add-token --path /etc/google-home-notifier/


$ mv /path/to/google-home-notifier.service /usr/lib/systemd/system/google-home-notifier.service
# Regist new service to systemd
$ systemctl start google-home-notifier.service

# Check systemd status
systemctl status google-home-notifier.service
systemctl status google-home-notifier.service
● google-home-notifier.service - Send notifications to Google Home
   Loaded: loaded (/lib/systemd/system/google-home-notifier.service; disabled; vendor preset: enabled)
   Active: active (running) since Fri 2023-01-20 13:54:28 JST; 13s ago
 Main PID: 9005 (google-home-not)
    Tasks: 10 (limit: 4915)
   CGroup: /system.slice/google-home-notifier.service
           └─9005 /usr/local/sbin/google-home-notifier daemon --path /etc/google-home-notifier/ --locale ja

Jan 20 13:54:28 tomoya-rasp systemd[1]: Started Send notifications to Google Home.
Jan 20 13:54:28 tomoya-rasp google-home-notifier[9005]: 2023/01/20 13:54:28 commands.go:58: Start daemon.
Jan 20 13:54:28 tomoya-rasp google-home-notifier[9005]: 2023/01/20 13:54:28 server.go:42: server start on port: 8000
```
