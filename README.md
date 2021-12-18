### standardnotes extensions server

- 30+ auto updating extensions for your standardnotes server
- pure go, no git subprocess

### usage
1. setup docker as described below
1. navigate to standardnotes settings -> General -> Advanced Settings -> Install Custom Extension
1. take the link for the extension you want from the web UI

### docker example

```yaml
services:
  extensions:
    build: path/to/this/repo
    environment:
    - SN_EXTS_LISTEN_ADDR=:80
    - SN_EXTS_REPOS_DIR=/repos
    - SN_EXTS_BASE_URL=https://extensions.your.domain
    - SN_EXTS_UPDATE_INTERVAL_MINS=4320 # 3 days
    expose:
    - 80
    volumes:
    - ./extensions_repos:/repos
  web:
    ...
  db:
    ...
```

### screenshots

standardnotes settings page
![image](https://user-images.githubusercontent.com/6832539/146651248-0b2f4cc0-7b96-49cf-80d9-ec4d60d0001a.png)

web UI for this repo
![image](https://user-images.githubusercontent.com/6832539/146651232-a2877284-2f79-40aa-88f8-1e39f0cd10d6.png)

