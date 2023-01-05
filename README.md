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
chmod +x google-home-notifier
mv google-home-notifier /usr/local/sbin/

mv /path/to/google-home-notifier.service /usr/lib/systemd/system/google-home-notifier.service

# Regist new service to systemd
systemctl daemon-reload
systemctl start google-home-notifier.service

# Check systemd status
systemctl status git-daemon
● git-daemon.service - Git Daemon for Malware
   Loaded: loaded (/usr/lib/systemd/system/git-daemon.service; disabled; vendor preset: disabled)
   Active: active (running) since Tue 2021-08-17 14:23:58 UTC; 5s ago
  Process: 23292 ExecStop=/bin/kill -KILL $MAINPID (code=exited, status=1/FAILURE)
 Main PID: 23347 (git-daemon)
    Tasks: 4
   Memory: 1.1M
   CGroup: /system.slice/git-daemon.service
           └─23347 /usr/local/sbin/git-daemon --file /var/git-daemon/config.yml --log-file /var/log/git-daemon.log
```
