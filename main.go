package main

import (
	"time"
	"flag"
)

var template_file string = "example/template.tmpl"
var target_file string = "example/template.cfg"
var interval int64 = 1
var cmd string = "true"
var version bool = false
var buildVersion string = "0.1.a1.007"

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
	interval := flag.Int64("interval", interval, "notify command interval (secs)")
	flag.BoolVar(&version, "version", false, "show version")

	flag.Usage = usage
	flag.Parse()

	if version {
		println("version : v" + buildVersion)
		return
	}

	debug("Starting Swarm Template")
	service := NewServiceFromEnv()

	for {
		services, _ := service.GetServices()

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