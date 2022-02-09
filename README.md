# Meetin

Loads events from your calendar and draw a countdown on LEDs.


## Setup

### 1. Google Cloud OAuth client id

Create `credentials.json` to be able to fetch events from Google Calendar.
Follow https://developers.google.com/workspace/guides/create-credentials to
create a OAuth client ID file. Rename the file named similar to
`code_secret_client_<LONGSTRING>.apps.googleusercontent.com.json` as
`credentials.json` in the current directory.

* Create (or reuse) a Google Cloud project: https://console.cloud.google.com/
* Create OAuth consent: https://console.cloud.google.com/apis/credentials/consent
* Create OAuth client ID: https://console.cloud.google.com/apis/credentials
* Enable Calendar API:
  https://console.cloud.google.com/apis/library/calendar-json.googleapis.com


### 2.Home Assistant api key

Create `api.key` with a Long-Lived Access Token to access Home Assistant. You
can create one from the Web UI by visiting `http://IP_ADDRESS:8123/profile`


### 3. Esphome

Flash a esp8266 or esp32 with [esphome.io](https://esphome.io) with a light
configuration similar to what [meetin.yaml](meetin.yaml) defines.


## Usage

```
go install github.com/maruel/meetin
meetin -hahost <HA_HOST>:8123 -cid <CALENDAR_ID> -path <PATH_TO_SECRET_FILES>
```

Installing permanently when you login:

```
cp rsc/meetin.service to $HOME/.config/systemd/user/
# Replace HA_HOST and CALENDAR_ID as relevant.
vim $HOME/.config/systemd/user/meetin.service
systemctl --user daemon-reload
systemctl --user enable meetin
systemctl --user start meetin
journalctl --user -u meetin -f
```
