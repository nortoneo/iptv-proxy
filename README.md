# Snippets

## Standalone docker-compose.yml

```
version: "3.4"
services:
  iptv-proxy:
    container_name: iptvproxy
    build:
      context: ./
      dockerfile: Dockerfile
    ports:
      - "80:1338"
    environment:
      LIST_sample: https://sample.playlist.com/list
      APP_URL: http://127.0.0.1:80
      ENCRYPTION_KEY: your_encryption_passphrase
```


## Docker cli build and run

```
docker build -t go-iptv-proxy .
```

```
docker run --rm -it -p 1338:1338 -e LIST_sample=https://google.com --name iptvproxy go-iptv-proxy
```