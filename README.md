Swarm Template
===============

The daemon swarm-template queries a Docker Swarm API and updates any number of specified templates on the file system.
swarm-template can optionally run arbitrary commands when the update process completes.

Installation
------------

You can download a released `swarm-template` artifact from
 [the Swarm Template release page](https://github.com/zekiunal/swarm-template/releases/). If you wish to compile from source,
please see the instructions in the [Contributing](#contributing) section.

Usage
-----

### Command Line
The CLI interface supports all of the options detailed above.

#### Help

```shell
# ./swarm-template -h

Usage: swarm-template [options]

Generate files from docker swarm api

Options:
  -cmd restart xyz
        run command after template is regenerated (e.g restart xyz) (default "true")
  -host string
        swarm manager address. (default "unix:///var/run/docker.sock")
  -interval int
        notify command interval (secs) (default 1)
  -target_file string
        path to a write the template. (default "example/template.cfg")
  -template_file string
        path to a template to generate (default "example/template.tmpl")
  -version
        show version
For more information, see https://github.com/zekiunal/swarm-template

```

#### Run 

```shell
$ ./swarm-template \
    -host="tcp://0.0.0.0:2375"
    -template_file="example/template.tmpl" \
    -target_file="example/target_file" \
    -cmd="/usr/sbin/nginx -s reload" \
```

### Example

#### 1 - Create example services

```shell
docker service rm api-example-com_v1-000 api-example-com_v1-001 anotherapi-example-com_v1-0000 anotherapi-example-com_v1-0001 anotherapi-example-com_v1-0002 logs-example-com static-example-com www-example-com

docker service create --name www-example-com --label st.tags="backend,development,public" --label st.version="1.000" --label st.group=www.example.com alpine ping docker.com

docker service create --name static-example-com --label st.tags="static,development,public" --label st.version="1.000" --label st.group=static.example.com alpine ping docker.com   
 
docker service create --name logs-example-com --label st.tags="backend,development,internal" --label st.version="1.000" --label st.group=logs.example.com alpine ping docker.com

docker service create --name api-example-com_v1-000 --label st.tags="backend,development,public,internal,api" --label st.version="1.000" --label st.group=api.example.com  alpine ping docker.com
                        
docker service create --name api-example-com_v1-001 --label st.tags="backend,development,public,internal,api" --label st.version="1.001" --label st.group=api.example.com alpine ping docker.com
 
docker service create --name anotherapi-example-com_v1-0000 --label st.tags="backend,development,public,internal,api" --label st.version="1.000" --label st.group=anotherapi.example.com alpine ping docker.com
                        
docker service create --name anotherapi-example-com_v1-0001 --label st.tags="backend,development,public,internal,api" --label st.version="1.001" --label st.group=anotherapi.example.com alpine ping docker.com

docker service create --name anotherapi-example-com_v1-0002 --label st.tags="backend,development,public,internal,api" --label st.version="1.002" --label st.group=anotherapi.example.com alpine ping docker.com
```

#### 2 - Run swarm-template

```shell
go get -d -v -t && go build -v  && ./swarm-template -template_file="example/template.tmpl" -target_file="example/nginx.conf"
```

#### Results

```shell
# ./swarm-template -template_file="example/template.tmpl" -target_file="example/nginx.conf"
_/zeki/swarm-template
2016/12/20 12:20:47 Starting Swarm Template
2016/12/20 12:20:47 Added   : api-example-com_v1-001
2016/12/20 12:20:47 Added   : www-example-com
2016/12/20 12:20:47 Added   : anotherapi-example-com_v1-0000
2016/12/20 12:20:47 Added   : anotherapi-example-com_v1-0001
2016/12/20 12:20:47 Added   : anotherapi-example-com_v1-0002
2016/12/20 12:20:47 Added   : logs-example-com
2016/12/20 12:20:47 Added   : api-example-com_v1-000
2016/12/20 12:20:47 Added   : static-example-com

```

##### template file

```
{{ define "main" }}
########################################################################################################################
# List Backend UpStreams
########################################################################################################################
{{ range $service := . }}{{$service_name:=.Name}}{{ range $label := .Labels }}{{ if and ( . | contains "backend") ( . | contains "development") }}
upstream {{$service_name}} {
    server {{$service_name}}:9000 weight=100 max_fails=5000 fail_timeout=5;
}{{ end }}{{ end }}{{ end }}

########################################################################################################################
# List External Services
########################################################################################################################
{{$services := .}}{{ range group . }}{{$service_name:=.Name}}{{$domain := index .Labels "st.group"}}{{ range $label := .Labels }}{{ if and ( . | contains "backend") ( . | contains "development") }}
server {
    listen 80;
    server_name {{$domain}};
    ..
    ...

    location @rewrite3 {
        rewrite ^({{ range keyBy $services $domain}}/v{{index .Labels "st.version"}}|{{end}})/(.*)$ $1/index.php?_url=/$2;
    }
    {{ range keyBy $services $domain}}{{$version := index .Labels "st.version"}}
    location /v{{$version}} {
        root /www/http/public/;
        try_files $uri $uri/ @rewrite3;
        location ~ ^/v{{$version}}(.+\.php)$ {
            fastcgi_pass {{$service_name}};
            ..
            ...
            fastcgi_param VERSION v{{$version}};
        }
    }{{end}}
}{{ end }}{{end}}{{end}}
########################################################################################################################
# List Internal Setup
########################################################################################################################
server {
    listen       80  default_server;
    server_name  _;
    ..
    ...

    location @rewrite2 {
        rewrite ^({{ range $service := . }}{{$service_name:=.Name}}{{ range $label := .Labels }}{{ if and ( . | contains "development") ( . | contains "backend") ( . | contains "internal") }}{{$service_parse := $service_name | split "_"}}{{range $key, $value := $service_parse}}{{if eq $key 1}}{{$value_v := $value | replaceAll "-" "."}}/{{$value_v}}{{else}}/{{$value}}{{end}}{{end}}|{{end}}{{end}}{{end}})/(.*)$ $1/index.php?_url=/$2;
    }
    {{ range $service := . }}{{$service_name:=.Name}}{{ range $label := .Labels }}{{ if and ( . | contains "development") ( . | contains "backend") ( . | contains "internal") }}{{$service_parse := $service_name | split "_"}}
    location {{range $key, $value := $service_parse}}{{if eq $key 1}}{{$value_v := $value | replaceAll "-" "."}}/{{$value_v}}{{else}}/{{$value}}{{end}}{{end}} {
        root /www/http/public/;
        try_files $uri $uri/ @rewrite2;
        location ~ ^{{$service_parse := $service_name | split "_"}}{{range $key, $value := $service_parse}}{{if eq $key 1}}{{$value_v := $value | replaceAll "-" "."}}/{{$value_v}}{{else}}/{{$value}}{{end}}{{end}}(.+\.php)$ {
            fastcgi_pass {{$service_name}};
            ..
            ...
            fastcgi_param VERSION {{$service_parse := $service_name | split "_"}}{{range $key, $value := $service_parse}}{{if eq $key 1}}{{$value_v := $value | replaceAll "-" "."}}{{$value_v}}{{end}}{{end}};
        }
    }{{end}}{{end}}{{end}}
}
{{end}}
```

##### generated target file

```
########################################################################################################################
# List Backend UpStreams
########################################################################################################################
upstream anotherapi-example-com_v1-0000 {
    server anotherapi-example-com_v1-0000:9000 weight=100 max_fails=5000 fail_timeout=5;
}
upstream api-example-com_v1-001 {
    server api-example-com_v1-001:9000 weight=100 max_fails=5000 fail_timeout=5;
}
upstream www-example-com {
    server www-example-com:9000 weight=100 max_fails=5000 fail_timeout=5;
}
upstream api-example-com_v1-000 {
    server api-example-com_v1-000:9000 weight=100 max_fails=5000 fail_timeout=5;
}
upstream logs-example-com {
    server logs-example-com:9000 weight=100 max_fails=5000 fail_timeout=5;
}
upstream anotherapi-example-com_v1-0001 {
    server anotherapi-example-com_v1-0001:9000 weight=100 max_fails=5000 fail_timeout=5;
}
upstream anotherapi-example-com_v1-0002 {
    server anotherapi-example-com_v1-0002:9000 weight=100 max_fails=5000 fail_timeout=5;
}

########################################################################################################################
# List External Services
########################################################################################################################
server {
    listen 80;
    server_name anotherapi.example.com;
    ..
    ...

    location @rewrite3 {
        rewrite ^(/v1.000|/v1.001|/v1.002|)/(.*)$ $1/index.php?_url=/$2;
    }
    
    location /v1.000 {
        root /www/http/public/;
        try_files $uri $uri/ @rewrite3;
        location ~ ^/v1.000(.+\.php)$ {
            fastcgi_pass anotherapi-example-com_v1-0000;
            ..
            ...
            fastcgi_param VERSION v1.000;
        }
    }
    location /v1.001 {
        root /www/http/public/;
        try_files $uri $uri/ @rewrite3;
        location ~ ^/v1.001(.+\.php)$ {
            fastcgi_pass anotherapi-example-com_v1-0000;
            ..
            ...
            fastcgi_param VERSION v1.001;
        }
    }
    location /v1.002 {
        root /www/http/public/;
        try_files $uri $uri/ @rewrite3;
        location ~ ^/v1.002(.+\.php)$ {
            fastcgi_pass anotherapi-example-com_v1-0000;
            ..
            ...
            fastcgi_param VERSION v1.002;
        }
    }
}

server {
    listen 80;
    server_name api.example.com;
    ..
    ...

    location @rewrite3 {
        rewrite ^(/v1.001|/v1.000|)/(.*)$ $1/index.php?_url=/$2;
    }
    
    location /v1.001 {
        root /www/http/public/;
        try_files $uri $uri/ @rewrite3;
        location ~ ^/v1.001(.+\.php)$ {
            fastcgi_pass api-example-com_v1-001;
            ..
            ...
            fastcgi_param VERSION v1.001;
        }
    }
    location /v1.000 {
        root /www/http/public/;
        try_files $uri $uri/ @rewrite3;
        location ~ ^/v1.000(.+\.php)$ {
            fastcgi_pass api-example-com_v1-001;
            ..
            ...
            fastcgi_param VERSION v1.000;
        }
    }
}

server {
    listen 80;
    server_name www.example.com;
    ..
    ...

    location @rewrite3 {
        rewrite ^(/v1.000|)/(.*)$ $1/index.php?_url=/$2;
    }
    
    location /v1.000 {
        root /www/http/public/;
        try_files $uri $uri/ @rewrite3;
        location ~ ^/v1.000(.+\.php)$ {
            fastcgi_pass www-example-com;
            ..
            ...
            fastcgi_param VERSION v1.000;
        }
    }
}

server {
    listen 80;
    server_name logs.example.com;
    ..
    ...

    location @rewrite3 {
        rewrite ^(/v1.000|)/(.*)$ $1/index.php?_url=/$2;
    }
    
    location /v1.000 {
        root /www/http/public/;
        try_files $uri $uri/ @rewrite3;
        location ~ ^/v1.000(.+\.php)$ {
            fastcgi_pass logs-example-com;
            ..
            ...
            fastcgi_param VERSION v1.000;
        }
    }
}
########################################################################################################################
# List Internal Setup
########################################################################################################################
server {
    listen       80  default_server;
    server_name  _;
    ..
    ...

    location @rewrite2 {
        rewrite ^(/anotherapi-example-com/v1.0000|/api-example-com/v1.001|/api-example-com/v1.000|/logs-example-com|/anotherapi-example-com/v1.0001|/anotherapi-example-com/v1.0002|)/(.*)$ $1/index.php?_url=/$2;
    }
    
    location /anotherapi-example-com/v1.0000 {
        root /www/http/public/;
        try_files $uri $uri/ @rewrite2;
        location ~ ^/anotherapi-example-com/v1.0000(.+\.php)$ {
            fastcgi_pass anotherapi-example-com_v1-0000;
            ..
            ...
            fastcgi_param VERSION v1.0000;
        }
    }
    location /api-example-com/v1.001 {
        root /www/http/public/;
        try_files $uri $uri/ @rewrite2;
        location ~ ^/api-example-com/v1.001(.+\.php)$ {
            fastcgi_pass api-example-com_v1-001;
            ..
            ...
            fastcgi_param VERSION v1.001;
        }
    }
    location /api-example-com/v1.000 {
        root /www/http/public/;
        try_files $uri $uri/ @rewrite2;
        location ~ ^/api-example-com/v1.000(.+\.php)$ {
            fastcgi_pass api-example-com_v1-000;
            ..
            ...
            fastcgi_param VERSION v1.000;
        }
    }
    location /logs-example-com {
        root /www/http/public/;
        try_files $uri $uri/ @rewrite2;
        location ~ ^/logs-example-com(.+\.php)$ {
            fastcgi_pass logs-example-com;
            ..
            ...
            fastcgi_param VERSION ;
        }
    }
    location /anotherapi-example-com/v1.0001 {
        root /www/http/public/;
        try_files $uri $uri/ @rewrite2;
        location ~ ^/anotherapi-example-com/v1.0001(.+\.php)$ {
            fastcgi_pass anotherapi-example-com_v1-0001;
            ..
            ...
            fastcgi_param VERSION v1.0001;
        }
    }
    location /anotherapi-example-com/v1.0002 {
        root /www/http/public/;
        try_files $uri $uri/ @rewrite2;
        location ~ ^/anotherapi-example-com/v1.0002(.+\.php)$ {
            fastcgi_pass anotherapi-example-com_v1-0002;
            ..
            ...
            fastcgi_param VERSION v1.0002;
        }
    }
}

```

### Templating Language

Swarm Template consumes template files in the [Go Template][] format.

#### Additional Functions

##### `group`
Query Swarm API for all services in the group by "st.group" label. Services are queried using the following syntax:

```liquid
{{ range group . }}
    # ...
{{ end}}
```

##### `KeyBy`

Creates a map that groups services by tag

```liquid
{{$services := . }}

{{ range keyBy $services "example_value"}}
    # ...
{{end}}
```

##### `contains`
Determines if a needle is within an iterable element.

```liquid
{{ range .Labels }}
    {{ if and ( . | contains "backend") ( . | contains "production") }}
        # ...
    {{ end }}
{{ end }}
```

##### `replaceAll`
Takes the argument as a string and replaces all occurrences of the given string with the given string.

```liquid
{{"foo.bar" | replaceAll "." "_"}}
```

This function can be chained with other functions as well:

```liquid
{{$variable := .Example | replaceAll "-" "."}}
```

##### `split`
Splits the given string on the provided separator:

```liquid
{{"application_version" | split "_"}}
```

Contributing
------------

### Build 

```
docker run --rm --name go-build -v "$PWD":"/go/src/swarm-template" -w "/go/src/swarm-template"  golang sh -c " go get; go build -v"
```

If you want to contribute to our project, please follow these guidelines:

1. Fork the repo
2. Choose the correct branch to base your contribution (see details bellow)
3. Implement and test your code
4. Submit a pull request

ToDo
------------
 * -cmd options multi command support
 
Thank You
------------
 
 * [AzMesai](http://azmesai.net/) Group Members
 
 
