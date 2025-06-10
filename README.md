# thermofridge-api

### Installation

- Add custom systemctl service in `/etc/systemd/system/thermofridge-api.service`:

  ```
  [Unit]
  Description=Thermofridge API
  After=network.target docker.service monitoring.service mosquitto.service
  Requires=docker.service monitoring.service mosquitto.service

  [Service]
  Type=simple
  User=USER_NAME_HERE
  WorkingDirectory=/path/to/thermofridge-api
  ExecStart=docker compose up
  ExecStop=docker compose down
  Restart=on-failure

  [Install]
  WantedBy=multi-user.target
  ```

  > Remember to replace `/path/to` with an actual path to the project

  Then enable the service to run on startup:

  ```
  sudo systemctl enable thermofridge-api.service
  ```
