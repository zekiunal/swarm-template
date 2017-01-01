package main

import (
	"text/template"
	"strings"
	"reflect"
	"os"
)

var funcMap = template.FuncMap{
	"contains": contains,
	"replaceAll":  replaceAll,
	"split":  split,
	"group":  group,
	"keyBy":  keyBy,
	"env": getEnv,
}

// https://github.com/kelseyhightower/confd/blob/master/resource/template/template_funcs.go#L45
// Getenv retrieves the value of the environment variable named by the key.
// It returns the value, which will the default value if the variable is not present.
// If no default value was given - returns "".
func getEnv(key string, v ...string) string {
	defaultValue := ""
	if len(v) > 0 {
		defaultValue = v[0]
	}

	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}

func replaceAll(f, t, s string) (string, error) {
	return strings.Replace(s, f, t, -1), nil
}

func contains(v, l interface{}) (bool, error) {
	return in(l, v)
}

func in(l, v interface{}) (bool, error) {
	lv := reflect.ValueOf(l)
	vv := reflect.ValueOf(v)

	switch lv.Kind() {
	case reflect.Array, reflect.Slice:
		var interfaceSlice []interface{}
		if reflect.TypeOf(l).Elem().Kind() == reflect.Interface {
			interfaceSlice = l.([]interface{})
		}

		for i := 0; i < lv.Len(); i++ {
			var lvv reflect.Value
			if interfaceSlice != nil {
				lvv = reflect.ValueOf(interfaceSlice[i])
			} else {
				lvv = lv.Index(i)
			}

			switch lvv.Kind() {
			case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
				switch vv.Kind() {
				case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
					if vv.Int() == lvv.Int() {
						return true, nil
					}
				}
			case reflect.Float32, reflect.Float64:
				switch vv.Kind() {
				case reflect.Float32, reflect.Float64:
					if vv.Float() == lvv.Float() {
						return true, nil
					}
				}
			case reflect.String:
				if vv.Type() == lvv.Type() && vv.String() == lvv.String() {
					return true, nil
				}
			}
		}
	case reflect.String:
		if vv.Type() == lv.Type() && strings.Contains(lv.String(), vv.String()) {
			return true, nil
		}
	}

	return false, nil
}

func split(sep, s string) ([]string, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return []string{}, nil
	}
	return strings.Split(s, sep), nil
}

func group(value []SW) []SW {
	groups := []SW{}
	seen := map[string]SW{}
	for _, s := range value {
		for range s.Labels {
			if _, ok := seen[s.Labels["st.group"]]; !ok {
				groups = append(groups, s)
				seen[s.Labels["st.group"]] = s
			}
		}
	}
	return groups
}

func keyBy(value []SW, key string) []SW {
	groups := []SW{}
	for _, s := range value {
		for _, k := range s.Labels {
			if (k == key) {
				groups = append(groups, s)
			}
		}
	}


	return groups
}