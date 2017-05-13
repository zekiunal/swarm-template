package main

import (
	"flag"
	"time"
)

var template_file 	string = "example/template.tmpl"
var target_file 	string = "example/template.cfg"
var cmd 		string = "true"
var interval int64 = 1
var version bool = false
var buildVersion string = "0.2.001"

// tcp://127.0.0.1:2375
var host string = "unix:///var/run/docker.sock"

func usage() {
	println(`Usage: swarm-template [options]

Generate files from docker swarm api

Options:`)

	flag.PrintDefaults()

	println(`For more information, see https://github.com/zekiunal/swarm-template`)
}

func main() {
	template_file := flag.String("template_file", template_file, "path to a template to generate")
	target_file := flag.String("target_file", target_file, "path to a write the template.")
	cmd := flag.String("cmd", cmd, "run command after template is regenerated (e.g `restart xyz`)")
	host := flag.String("host", host, "swarm manager address.")
	interval := flag.Int64("interval", interval, "notify command interval (secs)")
	flag.BoolVar(&version, "version", false, "show version")

	flag.Usage = usage
	flag.Parse()

	if version {
		println("version : v" + buildVersion)
		return
	}

	debug("Starting Swarm Template")
	service := NewServiceFromEnv(*host)

	for {
		services, _ := service.GetServices();

		new_services, _ := service.GetNewServices(services)
		if len(new_services) > 0 {
			service.UpdateTargetFile(new_services, services, *template_file, *target_file, *cmd)
		}

		removed_service := service.GetRemovedServices(services)
		if len(removed_service) > 0 {
			service.RemoveService(removed_service, services, *template_file, *target_file, *cmd)
		}

		time.Sleep(time.Second * time.Duration(*interval))
	}

}