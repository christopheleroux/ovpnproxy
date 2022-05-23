package main

import (
	"bufio"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"
)

var defaultRouteIp string

var conf OvpnProxyConfig = *loadConfiguration()

func main() {
	log.Println("Start Ovpn Proxy Monitor")
	initConfig()

	if !conf.dryRun {
		go infiniteCheck()
	}

	http.HandleFunc("/", viewHandler)
	var httpServeErr = http.ListenAndServe(":"+strconv.Itoa(conf.port), nil)
	if httpServeErr != nil {
		log.Println("Http server starting error" + httpServeErr.Error())
	}
}

func infiniteCheck() {
	var stateVpnUp = false
	for {
		newStateVpn := netUp(conf.vpnIfaceName)

		if !stateVpnUp && newStateVpn {
			log.Println("VPN goes up")
			//Start firewall
			execCmd(CMD_FIREWALL_START)
			execActionsList(conf.actionfileUp)
			stateVpnUp = true
		} else if stateVpnUp && !newStateVpn {
			log.Println("VPN goes down")
			//Reset firewall
			execCmd(CMD_FIREWALL_STOP)
			execActionsList(conf.actionfileDown)
			//try to reset vpn
			log.Println("Kill vpn service")
			execCmd(CMD_OPENVPN_KILL)
			time.Sleep(10 * time.Second)
			log.Println("ReRun vpn service")
			execCmd(CMD_OPENVPN_START)
			stateVpnUp = false
		}
		time.Sleep(1 * time.Second)
	}
}

// TODO error handling + exclude empty lines
func execActionsList(actionFile string) {
	file, _ := os.Open(actionFile)
	defer file.Close()
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		cmd := scanner.Text()
		log.Println("Run " + cmd)
		execCmd(cmd)
	}

}
