# thermostat-api

### Installation

Add custom systemctl service in `/etc/systemd/system/thermostat-api.service`:

```
[Unit]
Description=Thermostat API
After=network.target docker.service monitoring.service mqtt.service
Requires=docker.service monitoring.service mqtt.service

[Service]
Type=simple
User=USER_NAME_HERE
WorkingDirectory=/path/to/thermostat-api
ExecStart=docker compose up
ExecStop=docker compose down
Restart=on-failure

[Install]
WantedBy=multi-user.target
```

> Remember to replace `/path/to` with an actual path to the project

Then enable the service to run on startup:

```
sudo systemctl enable thermostat-api.service
```
