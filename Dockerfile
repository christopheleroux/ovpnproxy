# syntax=docker/dockerfile:1
FROM golang:1.18 AS build
WORKDIR /src/
COPY src/*.go ./src/
COPY go.mod ./go.mod
RUN go get all  
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o OvpnProxyMonitor ./src/*.go

FROM debian:11
MAINTAINER christophe Le Roux <christopheleroux1@yahoo.fr>

# Install packages
RUN apt-get update && \
    apt-get -y install pwgen python supervisor net-tools curl \
    fping openvpn wget \
    iptables psmisc procps \
    unzip \
    dante-server \
    && rm -rf /var/lib/apt/lists/*

# Create vpn config dir
RUN mkdir -p /opt/ovpn-pool
# openvpn running scripts
COPY scripts/vpn_up_actions.list /opt/vpn_up_actions.list
COPY scripts/vpn_down_actions.list /opt/vpn_down_actions.list
COPY scripts/vpn_run.sh /opt/vpn_run.sh


# Install firewall script
COPY scripts/firewall.sh /opt/firewall.sh
RUN chmod u+x /opt/firewall.sh

# danted socks proxy config
COPY config/danted.conf /etc/danted.conf

# Config supervisor
RUN mkdir -p /var/log/supervisor
COPY config/supervisord.conf /etc/supervisor/conf.d/supervisord.conf

# Copy go bin ovpnproxymonitor & templates
COPY --from=build /src/OvpnProxyMonitor /opt/OvpnProxyMonitor
COPY template/* ./opt/template/
RUN chmod u+x /opt/OvpnProxyMonitor

WORKDIR /opt/
VOLUME /opt/ovpn-pool

EXPOSE 80 1080
ENTRYPOINT ["/usr/bin/supervisord", "-n", "-c", "/etc/supervisor/conf.d/supervisord.conf"]
