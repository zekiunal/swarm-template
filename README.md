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
$ swarm-template \
    -template_file="example/template.tmpl" \
    -target_file="example/target_file" \
    -cmd="/usr/sbin/nginx -s reload" \
```

#### Templating Language
Swarm Template consumes template files in the [Go Template][] format.

##### Additional Functions

###### `group`
Query Swarm API for all services in the group by "st.group" label. Services are queried using the following syntax:

```liquid
{{ range group . }}
    # ...
{{ end}}
```

###### `KeyBy`

Creates a map that groups services by tag

```liquid
{{$services := . }}

{{ range keyBy $services "example_value"}}
    # ...
{{end}}
```

###### `contains`
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

If you want to contribute to our project, please follow these guidelines:

1. Fork the repo
2. Choose the correct branch to base your contribution (see details bellow)
3. Implement and test your code
4. Submit a pull request

ToDo
------------
 * -cmd options multi command support
 * HealthCheck support
 
Thank You
------------
 
 * [AzMesai](http://azmesai.net/) Group Memebers
 
 