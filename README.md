# Openvpn Proxy

OpenVpn proxy is buildkit to run openvpn proxy docker container.


## Features 

- Golang process monitor `OvpnProxyMonitor` : Circuit break on vpn state.
- Openvpn
- Preconfigured Socks proxy powered by `danted`
- Networking configuration : firewall, routing

## Process monitor `OvpnProxyMonitor`

`OvpnProxyMonitor` ensures the traffic is always routed through vpn. Services are started or stopped according to the vpn connection state.

A basic http front end provides an overview on current connection and services state.

OpenVpn candidate configuration files are provided to the container through the `$OVPN_POOL` parameter.


## Disclaimer

Openvpn Proxy must be ran with docker elevated privileges (`--privileges`) : **DO NOT USE IT FOR PRODUCTION** 

Openvpn Proxy wraps multiple processes in a single docker image. This is convenient but it's also a **bad practice : DO NOT USE IT FOR PRODUCTION** 

Go code quality is basic and has to be improved. 

# Run container

- Standalone

```
docker run --name vpn --rm  \
    --privileged \
    --sysctl net.ipv6.conf.all.disable_ipv6=0 \
    -e OVPN_CONF=$OVPN_CONF \
    -e OVPN_USER=$OVPN_USER \
    -e OVPN_PWD=$OVPN_PWD \
    -e LAN=$LAN \
    -v $OVPN_POOL:/opt/ovpn-pool \
    -p $HTTP_PORT:80 \
    -p $SOCKS_PORT:1080 \
    chrislrx/ovpnproxy:latest
```

- run.sh wrapper

First, create and populate env vars file `local/localenv`
Then `./run.sh`


# Options reference

|Parameter|Mandatory|Comment|
|-|-|-|
|OVPN_CONF||if not setted, first file in $OVPN_POOL selected as config|
|OVPN_USER|X|Open vpn login|
|OVPN_PWD|X|Open vpn password|
|OVPN_POOL_DIR||openvpn pool directory (*! runtime context*)|
|LAN|X|Lan ip|
|PORT| |Ovpn proxy monitor http port - *default 80*|
|HTTP_PORT|X|docker mapped http port - access to Ovpn Proxy Monitor|
|SOCKS_PORT|X|docker mapped socks port - access to danted socks proxy|
|DRYRUN| |Dry run mode - *default false*|


# Backlog

- UI - js static app : configuration selection and reload
- Network management improve : DNS settings, firewall
- Use golang port for supervisor :  https://github.com/ochinchina/supervisord
- Multiarch build : arm
- VPN selection policies : rotation
- UI - traffic stats