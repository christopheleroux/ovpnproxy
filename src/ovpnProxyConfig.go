package main

import (
	"io/ioutil"
	"log"
	"os"
	"strconv"
)

type OvpnProxyConfig struct {
	ovpnUser       string
	ovpnPassword   string
	ovpnFile       string
	ovpnPoolDir    string
	port           int
	lan            string
	vpnIfaceName   string
	actionfileUp   string
	actionfileDown string
	dryRun         bool
}

const (
	// Mandatory parameters
	ENV_OVPN_USER = "OVPN_USER"
	ENV_OVPN_PWD  = "OVPN_PWD"

	//Optional parameters with default values
	ENV_OVPN_POOL_DIR = "OVPN_POOL_DIR"
	ENV_PORT          = "PORT"
	ENV_LAN           = "LAN"
	ENV_DRYRUN        = "DRYRUN"

	DEFAULT_OVPN_POOL_DIR = "/opt/ovpn-pool"
	DEFAULT_PORT          = 80
	DEFAULT_LAN           = "192.168.0.1"
	DEFAULT_DRYRUN        = false

	//
	DEFAULT_VPN_INTERFACE_NAME = "tun0"
	DEFAULT_ACTIONFILE_VPNUP   = "/opt/vpn_up_actions.list"
	DEFAULT_ACTIONFILE_VPNDOWN = "/opt/vpn_down_actions.list"

	//Optional parameters without default values
	ENV_OVPN_CONF = "OVPN_CONF"

	// Constants
	OVPN_CONF_SYMLINK   = "/tmp/active.ovpn"
	OVPN_AUTH           = "/tmp/ovpn_auth.conf"
	OVPN_UNDEFINED_CONF = "undefined"
	CMD_OPENVPN_START   = "supervisorctl start openvpn"
	CMD_OPENVPN_KILL    = "killall openvpn"
	CMD_FIREWALL_START  = "/opt/firewall.sh start"
	CMD_FIREWALL_STOP   = "/opt/firewall.sh stop"
)

func loadConfiguration() *OvpnProxyConfig {

	// Set Optional attributes with default values
	defaultConf := OvpnProxyConfig{
		port:           DEFAULT_PORT,
		ovpnPoolDir:    DEFAULT_OVPN_POOL_DIR,
		ovpnFile:       OVPN_UNDEFINED_CONF,
		lan:            DEFAULT_LAN,
		vpnIfaceName:   DEFAULT_VPN_INTERFACE_NAME,
		actionfileUp:   DEFAULT_ACTIONFILE_VPNUP,
		actionfileDown: DEFAULT_ACTIONFILE_VPNDOWN,
		dryRun:         DEFAULT_DRYRUN,
	}
	conf := &OvpnProxyConfig{}

	// Set mandatory variables
	conf.ovpnUser = os.Getenv(ENV_OVPN_USER)
	conf.ovpnPassword = os.Getenv(ENV_OVPN_PWD)

	// Set Optional variables
	envOvpnPoolDir, envOvpnPoolDirPresent := os.LookupEnv(ENV_OVPN_POOL_DIR)
	if envOvpnPoolDirPresent {
		conf.ovpnPoolDir = envOvpnPoolDir
	} else {
		conf.ovpnPoolDir = defaultConf.ovpnPoolDir
	}
	envPort, envPortPresent := os.LookupEnv(ENV_PORT)
	if envPortPresent {
		conf.port, _ = strconv.Atoi(envPort)
	} else {
		conf.port = defaultConf.port
	}

	envLan, envLanPresent := os.LookupEnv(ENV_LAN)
	if envLanPresent {
		conf.lan = envLan
	} else {
		conf.lan = defaultConf.lan
	}

	envDryRun, envDryRunPresent := os.LookupEnv(ENV_DRYRUN)
	if envDryRunPresent {
		conf.dryRun, _ = strconv.ParseBool(envDryRun)
	} else {
		conf.dryRun = defaultConf.dryRun
	}

	//Optional parameters without default values

	envOvpnFile, envOvpnFilePresent := os.LookupEnv(ENV_OVPN_CONF)
	if envOvpnFilePresent {
		conf.ovpnFile = envOvpnFile
	} else {
		conf.ovpnFile = defaultConf.ovpnFile
	}

	conf.vpnIfaceName = DEFAULT_VPN_INTERFACE_NAME
	conf.actionfileUp = DEFAULT_ACTIONFILE_VPNUP
	conf.actionfileDown = DEFAULT_ACTIONFILE_VPNDOWN

	return conf
}

func initConfig() {
	log.Println("=== Initialize configuration ===")

	log.Println("Write ovpn crendentials file")
	crendentials := []byte(conf.ovpnUser + "\n" + conf.ovpnPassword)
	err := ioutil.WriteFile(OVPN_AUTH, crendentials, 0600)
	if err != nil {
		panic(err)
	}

	// Apply network configuration
	if !conf.dryRun {
		defaultRouteIp = getDefaultRouteIP()
		log.Println("Initial default route via " + defaultRouteIp)
		//TODO extract command as const + error handling
		execCmd("route add -net " + conf.lan + " netmask 255.255.255.0 gw " + defaultRouteIp)
	}

	if !conf.dryRun {
		log.Println("Write DNS entries")
		writeDNSEntries()
	}

	var ovpnConfSet = setOvpnConfig()

	if !conf.dryRun && ovpnConfSet {
		log.Println("Run vpn service")
		execCmd(CMD_OPENVPN_START)
	}
}

func getCurrentConfig() string {
	link, _ := os.Readlink(OVPN_CONF_SYMLINK)
	return link
}

func setOvpnConfig() bool {
	// List availble config files
	files, err := ioutil.ReadDir(conf.ovpnPoolDir)
	if err != nil {
		log.Println("Error loading openvpn config files pool : " + err.Error())
		return false
	}

	if conf.ovpnFile == OVPN_UNDEFINED_CONF {
		conf.ovpnFile = files[0].Name()
		log.Println("No ovpn conf provided : selecting first one in pool - " + conf.ovpnFile)
	} else {
		// check if provided config exists
		var confFileExists = false
		for i := range files {
			if files[i].Name() == conf.ovpnFile {
				log.Println("Openvpn config file found " + conf.ovpnFile)
				confFileExists = true
			}
		}
		if !confFileExists {
			log.Println("Openvpn config file NOT found in Pool " + conf.ovpnFile)
			return false
		}
	}

	log.Println("Set openvpn configuration > " + conf.ovpnFile)

	// Delete symlink if already exists

	os.Symlink(conf.ovpnPoolDir+"/"+conf.ovpnFile, OVPN_CONF_SYMLINK)
	return true
}

func listAvailableConfigs() []string {
	files, err := ioutil.ReadDir(conf.ovpnPoolDir)
	if err != nil {
		panic(err)
	}
	filenames := make([]string, len(files))
	for i := range files {
		filenames[i] = files[i].Name()
	}
	return filenames
}
