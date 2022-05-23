#!/usr/bin/env bash


usage() {
    echo "Usage : $0 [build|run|tail|dev]"
}

docker_run(){
    
    source local/localenv

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

}

docker_tail() {
    docker exec -it vpn supervisorctl tail -f ovpnproxymonitor
}

build() {
  # Create Container
  TIMESTAMP=$(date +"%y%m%d")
  docker build -t chrislrx/ovpnproxy:latest  -t chrislrx/ovpnproxy:$TIMESTAMP .

}

dev_run() {
  PORT=8081 \
  OVPN_USER=user \
  OVPN_PWD=pass \
  DRYRUN=true \
  go run src/*.go
}


MAIN_ARG=${1:-run}

case "$MAIN_ARG" in
  run)
    docker_run
    ;;
  tail)
    docker_tail
    ;;
  build)
    build
    ;;
  dev)
    dev_run
    ;;
  *)
    usage
    ;;
esac

