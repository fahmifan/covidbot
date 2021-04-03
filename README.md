# Covidbot

Crawl data from Pikobar API, currently only crawl daily cases.

Commands:
```
parse             parse json output from crawl
crawl             crawl into pikobar api and output a json
crawl everyday    craw every day
help              Help about any command
```

[deployr](https://github.com/skx/deployr) is used to deploy the binary to server. 
It uses ssh to copy the binary and restart the systemd.

To deploy using `deployr` follow these instructions:
- create new user `deployr`
- add your local machine public ssh key to user `deployr`
- login as `deployr`
- create a systemd unit file in `/home/deployr/.config/systemd/user/covidbot.service`
    - check `example` folder for references
- if using systemd unit file from exmaple, then create a folder in `/home/deployr/admin`
- create env file `touch /home/deployr/admin/.env`
    - reference: `.env.example`
- enable the user systemd to run on boot: 
    - login as `root` then run `loginctl enable-linger deployr`
- reload the current user systemd `systemd --user daemon-reload`
- enable the service `systemd --user enable covidbot.service`
- from your local machine, run `make deploy host=example.com`
    - `example.com` should be your server host or IP
- check if the service is running well `journalctl --user -f -u covidbot.service`

## TODO
- [x] add webhook
- [x] add sync updates to telegram
- [ ] add more features
    - filter statuses by date