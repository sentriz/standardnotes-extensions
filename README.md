### standardnotes extensions server

- 30+ auto updating extensions for your standardnotes server
- use activation code `https://extensions.your.domain/index.json`
- pure go, no git subprocess

### docker example

```yaml
services:
  extensions:
    build: path/to/this/repo
    environment:
    - SN_EXTS_LISTEN_ADDR=:80
    - SN_EXTS_REPOS_DIR=/repos
    - SN_EXTS_DEFINITIONS_DIR=/definitions
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

![](https://i.imgur.com/EWQvpVR.png)
![](https://i.imgur.com/ZqmkzEW.png)
