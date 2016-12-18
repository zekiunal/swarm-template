package main

import (
	"github.com/docker/docker/api/types/swarm"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	"golang.org/x/net/context"
	"log"
	"os"
	"time"
	"text/template"
	"fmt"
	"os/exec"
	"encoding/json"

	"strconv"
	"sync"
)

type Service struct {
	Host                 string
	Services             map[string]SW
	ServiceLastCreatedAt time.Time
	DockerClient         *client.Client
}

type SW struct {
	ID        string
	Name      string                `json:",omitempty"`
	Labels    map[string]string     `json:",omitempty"`
	CreatedAt time.Time             `json:",omitempty"`
	UpdatedAt time.Time             `json:",omitempty"`
	Replicas  int                   `json:",omitempty"`
	Running   int                   `json:",omitempty"`
	Version   int			`json:",omitempty"`
	Changed   bool			`json:",omitempty"`
}

type ServiceCreator interface {
	GetServices() ([]SW, error)
	GetNewServices(services []SW) ([]SW, error)
	UpdateTargetFile(services []SW, all []SW, template_file string, target_file string) error
}

var debug = log.Println

func NewServiceFromEnv() *Service {

	host := "unix:///var/run/docker.sock"

	return NewService(host)
}

func NewService(host string) *Service {

	docker_client, err := client.NewClient(host, "v1.24", nil, map[string]string{"User-Agent":"engine-api-cli-1.0"})

	if err != nil {
		debug(err.Error())
	}

	return &Service{
		Host: host,
		Services:              map[string]SW{},
		DockerClient:          docker_client,
	}
}

func (service *Service) GetServices() ([]SW, error) {
	services, err := service.DockerClient.ServiceList(context.Background(), types.ServiceListOptions{})
	nodes, err := service.DockerClient.NodeList(context.Background(), types.NodeListOptions{})
	tasks, err := service.DockerClient.TaskList(context.Background(), types.TaskListOptions{})

	activeNodes := make(map[string]struct{})
	for _, n := range nodes {
		if n.Status.State != swarm.NodeStateDown {
			activeNodes[n.ID] = struct{}{}
		}
	}

	running := map[string]int{}
	tasksNoShutdown := map[string]int{}
	for _, task := range tasks {
		if task.DesiredState != swarm.TaskStateShutdown {
			tasksNoShutdown[task.ServiceID]++
		}

		if _, nodeActive := activeNodes[task.NodeID]; nodeActive && task.Status.State == swarm.TaskStateRunning {
			running[task.ServiceID]++
		}
	}

	new_services := []SW{}

	for _, s := range services {

		replicas, _ := strconv.Atoi(fmt.Sprintf("%d", *s.Spec.Mode.Replicated.Replicas));
		version, _:= strconv.Atoi(fmt.Sprintf("%d", s.Meta.Version.Index));
		running_service := running[s.ID];

		changed := false

		if(service.Services[s.Spec.Name].Running != running_service) {
			changed = true
		}

		k := SW{
			ID: s.ID,
			Name: s.Spec.Name,
			Labels: s.Spec.Labels,
			CreatedAt: s.Meta.CreatedAt,
			UpdatedAt: s.Meta.UpdatedAt,
			Replicas: replicas,
			Running: running[s.ID],
			Version: version,
			Changed: changed,
		}

		if (running_service > 0) {
			new_services = append(new_services, k)
		}

	}

	if err != nil {
		debug(err.Error())
		// @todo remove
		jsonString, _ := json.Marshal(services[0])
		fmt.Println(string(jsonString))
		return []SW{}, err
	}

	return new_services, nil
}

func (service *Service) GetNewServices(running_services []SW) ([]SW, error) {

	new_services := []SW{}

	tmpCreatedAt := service.ServiceLastCreatedAt

	for _, item := range running_services {
		_, ok := service.Services[item.Name]
		if tmpCreatedAt.Nanosecond() == 0 || (item.CreatedAt.After(tmpCreatedAt) || ((item.Changed) && !ok) ) {
			new_services = append(new_services, item)
			debug("Added   : " + item.Name )
			service.Services[item.Name] = item
			if service.ServiceLastCreatedAt.Before(item.CreatedAt) {
				service.ServiceLastCreatedAt = item.CreatedAt
			}
		}
	}

	return new_services, nil
}

func (service *Service) GetRemovedServices(services []SW) []string {
	tmpMap := make(map[string]SW)

	for k, v := range service.Services {
		tmpMap[k] = v
	}

	for _, v := range services {
		if (service.Services[v.Name].Replicas > 0)  {
			delete(tmpMap, v.Name)
		}
	}

	rs := []string{}
	for k, _ := range tmpMap {
		rs = append(rs, k)
	}
	return rs
}

func (m *Service) UpdateTargetFile(services []SW, all []SW, template_file string, target_file string, cmd string) error {
	if len(services) > 0 {
		f, _ := os.Create(target_file)
		t := template.New("template").Funcs(funcMap)
		t, _ = t.ParseFiles(template_file)
		t.ExecuteTemplate(f, "main", &all)
		f.Close()

		executeCMD(cmd)
	}

	return nil
}



func (m *Service) RemoveService(removed_services []string, all []SW, template_file string, target_file string, cmd string) error {
	for _, v := range removed_services {
		delete(m.Services, v)
		debug("Removed : " + v)
	}

	f, _ := os.Create(target_file)
	t := template.New("template").Funcs(funcMap)
	t, _ = t.ParseFiles(template_file)
	t.ExecuteTemplate(f, "main", &all)
	f.Close()

	executeCMD(cmd)

	return nil
}

func runCMD(cmd string, wg *sync.WaitGroup) {
	//// splitting head => g++ parts => rest of the command
	//parts := strings.Fields(cmd)
	//head := parts[0]
	//parts = parts[1:len(parts)]

	out, err := exec.Command("/bin/sh", "-c", cmd).Output()
	if err != nil {
		fmt.Printf("%s", err)
	}
	fmt.Printf("%s", out)

	wg.Done() // Need to signal to wait group that this go routine is done
}

func executeCMD(cmd string) {
	wg := new(sync.WaitGroup)
	wg.Add(1)

	x := []string{cmd}
	go runCMD(x[0], wg)
	wg.Wait()
}