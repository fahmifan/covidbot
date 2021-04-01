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
It is using ssh to copy the binary and restart the systemd. 
To deploy run:
```
make deploy host=example.com
```

## TODO
- [x] add webhook
- [x] add sync updates to telegram
- [ ] add more features
    - filter statuses by date