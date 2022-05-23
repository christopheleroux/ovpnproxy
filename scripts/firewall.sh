#!/usr/bin/env bash

ipt="/sbin/iptables"
lan_interface="eth0"

start() {
    # iptables firewall for common LAMP servers.
    # This file should be located at /etc/firewall.bash, and is meant to work with
    # Jeff Geerling's firewall init script.

    # Remove all rules and chains.
    $ipt -F
    $ipt -X

    # Accept traffic from loopback interface (localhost).
    $ipt -A INPUT -i lo -j ACCEPT

    # Default : in and out connexions forbidden
    $ipt -t filter -P INPUT DROP
    $ipt -t filter -P FORWARD DROP
    $ipt -t filter -P OUTPUT DROP

    ## Allow established connexion input
    $ipt -A INPUT -m state --state ESTABLISHED,RELATED -j ACCEPT

    ## Allow established connexion output
    $ipt -A OUTPUT -m state --state ESTABLISHED,RELATED -j ACCEPT

    ##### IN lan 1080 - dante socks proxy
    $ipt -A INPUT -i $lan_interface -p tcp --dport 1080 -m state --state NEW,ESTABLISHED -j ACCEPT
    $ipt -A OUTPUT -o $lan_interface -p tcp --sport 1080 -m state --state ESTABLISHED,RELATED -j ACCEPT

    ##### IN lan 80
    $ipt -A INPUT -i $lan_interface -p tcp --dport 80 -m state --state NEW,ESTABLISHED -j ACCEPT
    $ipt -A OUTPUT -o $lan_interface -p tcp --sport 80 -m state --state ESTABLISHED,RELATED -j ACCEPT
    ##### IN lan 443
    $ipt -A INPUT -i $lan_interface -p tcp --dport 443 -m state --state NEW,ESTABLISHED -j ACCEPT
    $ipt -A OUTPUT -o $lan_interface -p tcp --sport 443 -m state --state ESTABLISHED,RELATED -j ACCEPT


    ##### OUT $lan_interface

    #dns
    $ipt -t filter -o $lan_interface -A OUTPUT -p tcp --dport 53 -j ACCEPT
    $ipt -t filter -o $lan_interface -A OUTPUT -p udp --dport 53 -j ACCEPT

    ##### IN lan
    $ipt -A INPUT -i $lan_interface -p tcp --dport 22 -m state --state NEW,ESTABLISHED -j ACCEPT

    ##### OUT lan
    #dns
    $ipt -t filter -o $lan_interface -A OUTPUT -p tcp --dport 53 -j ACCEPT
    $ipt -t filter -o $lan_interface -A OUTPUT -p udp --dport 53 -j ACCEPT

    ###### OUT TUN0
    $ipt -A OUTPUT -o tun0 -j ACCEPT
}

stop() {
    ## Failsafe - die if /sbin/iptables not found
    [ ! -x "$ipt" ] && { echo "$0: \"${ipt}\" command not found."; exit 1; }
    $ipt -P INPUT ACCEPT
    $ipt -P FORWARD ACCEPT
    $ipt -P OUTPUT ACCEPT
    $ipt -F
    $ipt -X
    $ipt -t nat -F
    $ipt -t nat -X
    $ipt -t mangle -F
    $ipt -t mangle -X
    $ipt -t raw -F
    $ipt -t raw -X
}


usage(){
    echo "$0 [start|stop]"
}


if [ "$1" == "stop" ];then
    stop
    exit 0
elif [ "$1" == "start" ];then
    start
    exit 0
fi

usage
