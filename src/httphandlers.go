package main

import (
	"html/template"
	"net/http"
	"strconv"
)

type Page struct {
	Title       string
	Status      string
	Ip          IpInfo
	Location    Location
	ProcessList []Process
}

type Location struct {
	Configured string
	Available  []string
}

func viewHandler(w http.ResponseWriter, r *http.Request) {
	ipInfo := getMyIp()

	p := Page{"OVPN Proxy Monitor",
		strconv.FormatBool(netUp(conf.vpnIfaceName)),
		ipInfo,
		Location{
			getCurrentConfig(),
			listAvailableConfigs(),
		},
		listProcess(),
	}
	t := template.New("Ovpn Proxy Monitor Front")
	t = template.Must(t.ParseFiles("template/main.tmpl"))
	e := t.ExecuteTemplate(w, "main", p)
	if e != nil {
		panic(e)
	}

}
