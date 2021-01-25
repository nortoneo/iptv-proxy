# Iptv-Proxy

Simple iptv proxy that converts m3u lists to its own internal urls.

Client devices can connect to this app instead of original provider.

## Typical usecases

* you dont want to reconfigure client devices when you change iptv provider
* you want to connect to iptv provider via vpn but you cant or dont want to setup vpn connection for client devices 

## Accessing playlist from  client device

You need to run this application with proper configuration passed by environment varialbles listed bellow.

You can create as many LIST_ variables as you need. 
For examle if you create variable ```LIST_myiptv``` with value of your provider url - you can access this list by you app url ```/list/myiptv```

For additional security or https support you can use nginx as reverse proxy.

## Environment variables that should be set

| Name | Default value | Description |
| :----: | --- | --- |
| `APP_URL` | http://localhost:1338 | Application url without trailing slash |
| `LISTEN_ADDRESS` | :1338 | Listen address for app http server |
| `LIST_yourlist` | | Url to original m3u list that will be accessible under /list/yourlist |
| `ENCRYPTION_KEY` |  | Passphrase used to encode original host |

## Environment variables that are optional

| Name | Default value | Description |
| :----: | --- | --- |
| `C_DIAL_TIMEOUT` | 60 | HTTP client dial timeout (seconds) |
| `C_DIAL_KEEPALIVE` | 300 | HTTP client keep-alive timeout (seconds) |
| `C_TLS_HANDSHAKE_TIMEOUT` | 30 | HTTP client TLS handshake timeout (seconds) |
| `C_RESPONSE_HEADER_TIMEOUT` | 30 | HTTP client response header timeout (seconds) |
| `C_EXPECT_CONTINUE_TIMEOUT` | 5 | HTTP expect continue timeout (seconds) |
| `C_TIMEOUT` | 300 | HTTP client request timeout (seconds) |
| `S_WRITE_TIMEOUT` | 300 | HTTP server write timeout (seconds) |
| `S_READ_TIMEOUT` | 300 | HTTP server read timeout (seconds) |
| `S_IDLE_TIMEOUT` | 300 | HTTP server keep-alive timeout (seconds) |


## Snippets

### Standalone docker-compose.yml sample

```
version: "3.4"
services:
  iptv-proxy:
    container_name: iptvproxy
    build:
      context: ./
      dockerfile: Dockerfile
    ports:
      - "1338:1338"
    environment:
      LIST_sample: https://sample.playlist.com/list
      APP_URL: http://127.0.0.1:1338
      ENCRYPTION_KEY: your_encryption_passphrase
```

### Behind NordVPN docker-compose.yml sample

```
version: "3.4"
services:
  vpn:
    container_name: vpn
    image: azinchen/nordvpn:latest # See: https://github.com/azinchen/nordvpn
    cap_add:
      - net_admin
    devices:
      - /dev/net/tun
    environment:
      - USER=${NORD_VPN_USER}
      - PASS=${NORD_VPN_PASS}
      - COUNTRY=${NORD_VPN_CONNECT}
      - CATEGORY=${NORD_VPN_CATEGORY}
      - NETWORK=${NORD_VPN_NETWORK}
      - TZ=${TIMEZONE}
      - CHECK_CONNECTION_CRON=*/1 * * * * 
      - CHECK_CONNECTION_URL=${VPN_CONNECTION_HEALTHCHECK_URL}
      - CHECK_CONNECTION_ATTEMPTS=2
      - OPENVPN_OPTS=--pull-filter ignore "ping-restart" --ping-exit 180
    ports:
      - 1338:1338 #iptv-proxy container
    dns:
      - 8.8.8.8
      - 8.8.4.4
    restart: unless-stopped

  iptv-proxy:
    container_name: iptv-proxy
    build:
      context: ./
      dockerfile: Dockerfile
    network_mode: service:vpn
    environment:
      LIST_sample: https://sample.playlist.com/list
      APP_URL: http://127.0.0.1:1338 #this should be your host ip or domain
      ENCRYPTION_KEY: your_encryption_passphrase
    restart: unless-stopped
```

### Docker cli build and run

```
docker build -t go-iptv-proxy .
```

```
docker run --rm \
  -p 1338:1338 \
  -e APP_URL=http://127.0.0.1:1338 \
  -e LIST_sample=https://sample.playlist.com/list \
  -e ENCRYPTION_KEY=your_encryption_passphrase \
  --name iptvproxy go-iptv-proxy
```

