# Iptv-Proxy

Simple iptv proxy that converts m3u lists to its own internal urls.  
It follows redirections and accept self signed certificates.

Client devices can connect to this app instead of original provider.  

## Typical usecases

* you dont want to reconfigure client devices when you change iptv provider
* you want to connect to iptv provider via vpn but you cant or dont want to setup vpn connection for client devices 

## Accessing playlist from client device

You need to run this application with proper configuration passed by environment varialbles or config file ```iptvproxy_config.yaml```.

If you want to setup list that will be available under /list/exmaple?token=123 url you need to set LIST_example=https://example-playlist/playlist.m3u8 and TOKEN_example=123  
or you can create entry in ```iptvproxy_config.yaml``` :   
```
lists:
  exmaple: #can be accesed by /list/example?token=123
    token: 123
    url: https://example-playlist/playlist.m3u8
```

You can create as many LIST_ & TOKEN_ variables as you need.  
  
If you want to override other settings from config file by envirenment variable you can access nested values by ```_```.  
For example to override encryption key you can set ```APP_ENCRYPTIONKEY=other_key```



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
      LIST_example: https://example.playlist.com/list
      TOKEN_example: 123
      APP_URL: http://127.0.0.1:1338
      APP_ENCRYPTIONKEY: your_encryption_passphrase
```

### Behind NordVPN docker-compose.yml sample

```
version: "3.4"
services:
  vpn:
    container_name: vpn
    image: azinchen/nordvpn:latest #See: https://github.com/azinchen/nordvpn
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
      LIST_example: https://example.playlist.com/list
      TOKEN_example: 123
      APP_URL: http://127.0.0.1:1338 #this should be your host ip or domain
      APP_ENCRYPTIONKEY: your_encryption_passphrase
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
  -e LIST_example=https://example.playlist.com/list \
  -e TOKEN_example=123 \
  -e APP_ENCRYPTIONKEY=your_encryption_passphrase \
  --name iptvproxy go-iptv-proxy
```

