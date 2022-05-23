package main

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"os/exec"
	"regexp"
	"strings"
	"time"
)

//TODO improve by service test, remote network acces, local ip ?
func netUp(netIface string) bool {
	l, err := net.Interfaces()
	if err != nil {
		panic(err)
	}
	for _, f := range l {
		if netIface == f.Name {
			return true
		}
	}
	return false
}

// TODO error handling
func execCmd(cmd string) ([]byte, error) {
	// splitting head => g++ parts => rest of the command
	parts := strings.Fields(cmd)
	head := parts[0]
	parts = parts[1:len(parts)]
	return exec.Command(head, parts...).Output()
}

// TODO improve DNS accessibility and parameters - ensure accesible through vpn network
func writeDNSEntries() {
	dnsEntries := []byte("nameserver 8.8.8.8\nnameserver 8.8.4.4")
	err := ioutil.WriteFile("/etc/resolv.conf", dnsEntries, 0644)
	if err != nil {
		log.Println("Error writing DNS entries in resolv.conf file - " + err.Error())
	}

}

func getDefaultRouteIP() string {

	//Extract initial default route ip
	cmdExtractRoute := exec.Command("/sbin/ip", "route")
	var out bytes.Buffer
	cmdExtractRoute.Stdout = &out
	cmdExtractRoute.Run()
	re := regexp.MustCompile(".*default via ([0-9]+\\.[0-9]+\\.[0-9]+\\.[0-9]+).*")
	match := re.FindStringSubmatch(out.String())
	defaultRouteIp = match[1]

	return defaultRouteIp
}

type IpInfo struct {
	Status      string
	Country     string
	CountryCode string
	Region      string
	RegionName  string
	City        string
	Zip         string
	Lat         string
	Lon         string
	Timezone    string
	Isp         string
	Org         string
	As          string
	Query       string
}

const (
	URL_IP_SERVICE = "http://ip-api.com/json"
)

func getMyIp() IpInfo {
	client := http.Client{
		Timeout: 5 * time.Second,
	}
	resp, err := client.Get(URL_IP_SERVICE)
	if err != nil {
		log.Println("Error fetching ip info")
	}
	defer resp.Body.Close()
	body, _ := ioutil.ReadAll(resp.Body)
	var ipInfo IpInfo
	json.Unmarshal([]byte(body), &ipInfo)
	return ipInfo
}

type Process struct {
	Name   string
	Status string
}

const (
	CMD_SUPERVISORCTL_STATUS = "supervisorctl status"
)

func listProcess() []Process {
	out, _ := execCmd(CMD_SUPERVISORCTL_STATUS)
	lines := regexp.MustCompile("\r?\n").Split(string(out), -1)
	var processes []Process
	for _, content := range lines {
		items := strings.Fields(content)
		if len(items) > 0 {
			processes = append(processes, Process{items[0], items[1]})
		}
	}
	return processes
}
