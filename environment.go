package environment

import (
	"embed"
	"fmt"
	"gopkg.in/yaml.v2"
	"os"
	"reflect"
	"regexp"
	"strconv"
	"strings"
)

//go:embed application.yml
var embedIn embed.FS

var embedOut embed.FS

type Env struct{}

var envStruct = &Env{}
var env map[string]string

func GetEnv() *Env {
	return envStruct
}

var isEnv = regexp.MustCompile(`\$\{(.*?)\}`)

func Get(name string) string {
	return envStruct.GetString(name)
}
func GetEmbed() embed.FS {
	return envStruct.GetEmbed()
}
func GetString(name string) string {
	return envStruct.GetString(name)
}
func GetInt(name string) int {
	return envStruct.GetInt(name)
}
func GetBool(name string) bool {
	return envStruct.GetBool(name)
}

func (e *Env) Get(name string) string {
	return e.GetString(name)
}
func (e *Env) GetEmbed() embed.FS {
	return embedOut
}

func (e *Env) GetString(name string) string {
	strVal, ok := env[name]
	if !ok {
		return ""
	}
	if isEnv.MatchString(strVal) {
		find := isEnv.FindStringSubmatch(strVal)
		sub := strings.SplitN(find[1], ":", 2)
		envOs := os.Getenv(sub[0])
		if len(envOs) > 0 {
			return envOs
		}
		if len(sub) > 1 {
			return sub[1]
		}
	}
	return strings.TrimSpace(strVal)
}
func (e *Env) GetInt(name string) int {
	strVal := e.GetString(name)
	if len(strVal) < 1 {
		return 0
	}
	intVal, _ := strconv.Atoi(strVal)
	return intVal
}
func (e *Env) GetBool(name string) bool {
	return e.GetString(name) == "true"
}

func InitEnv(emb embed.FS) {
	embedOut = emb
	initEnv()
}

func initEnv() {
	file, err := embedIn.ReadFile("application.yml")
	if err != nil {
		return
	}
	env = convertYamlToProp(file)
	file, err = embedOut.ReadFile("resources/application.yml")
	if err != nil {
		return
	}
	subenv := convertYamlToProp(file)
	for key, value := range subenv {
		env[key] = value
	}
}

func convertYamlToProp(file []byte) map[string]string {
	result := make(map[string]string)
	var local map[string]interface{}
	err := yaml.Unmarshal(file, &local)
	if err != nil {
		return nil
	}
	for key, val := range local {
		subMap := getFromMap(key, val, reflect.TypeOf(val))
		for k, v := range subMap {
			result[k] = v
		}
	}
	return result
}

func getFromMap(parent string, val interface{}, typeReflect reflect.Type) map[string]string {
	result := make(map[string]string)
	switch typeReflect.Kind() {
	case reflect.String:
		result[parent] = val.(string)
	case reflect.Bool:
		if val.(bool) {
			result[parent] = "true"
		} else {
			result[parent] = "false"
		}
	case reflect.Int:
		result[parent] = strconv.Itoa(val.(int))
	case reflect.Int64:
		result[parent] = strconv.FormatInt(val.(int64), 10)
	case reflect.Uint64:
		result[parent] = strconv.FormatUint(val.(uint64), 10)
	case reflect.Float64:
		result[parent] = strconv.FormatFloat(val.(float64), 'f', -1, 64)
	case reflect.Float32:
		result[parent] = strconv.FormatFloat(float64(val.(float32)), 'f', -1, 64)
	case reflect.Map:
		mapObj := val.(map[interface{}]interface{})
		for kk, vv := range mapObj {
			kkStr := kk.(string)
			subMap := getFromMap(parent+"."+kkStr, vv, reflect.TypeOf(vv))
			for k, v := range subMap {
				result[k] = v
			}
		}
	default:
		fmt.Println(val)
	}
	return result
}
