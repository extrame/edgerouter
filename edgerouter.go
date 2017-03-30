package edgerouter

import (
	"flag"
	"fmt"
	"log"
	"reflect"

	"strings"

	"context"

	"github.com/BurntSushi/toml"
)

func Organize(name string, plugins ...interface{}) *EdgeRouter {
	er := new(EdgeRouter)
	er.Name = name
	er.Plugins = make(map[string]interface{})
	for _, plugin := range plugins {
		rp := reflect.ValueOf(plugin)
		if rp.Type().Kind() == reflect.Ptr {
			rp = rp.Elem()
		}
		path := rp.Type().PkgPath()
		paths := strings.Split(path, "/")
		er.Plugins[paths[len(paths)-1]] = plugin
	}
	fmt.Println("----")
	return er
}

type EdgeRouter struct {
	Name    string
	Plugins map[string]interface{}
	servers map[string][]Server
}

func (er *EdgeRouter) ConfigBy(filename string) error {
	var tmp = make(map[string]toml.Primitive)
	var err error
	var meta toml.MetaData
	if meta, err = toml.DecodeFile(filename, &tmp); err == nil {
		er.servers = make(map[string][]Server)
		for k, plugin := range er.Plugins {
			rp := reflect.ValueOf(plugin)
			for i := 0; i < rp.NumField(); i++ {
				fp := rp.Field(i)
				ft := fp.Type()
				if ft.Name() == "UdpServer" ||
					ft.Name() == "UdpSeeker" ||
					ft.Name() == "TcpServer" ||
					ft.Name() == "SerialTcpSeeker" ||
					ft.Name() == "ConcurrentTcpSeeker" {
					val := reflect.New(ft).Interface()
					if err := meta.PrimitiveDecode(tmp[k], val); err == nil {
						er.servers[k] = append(er.servers[k], val.(Server))
					} else {
						return err
					}
				}
			}
			val := reflect.New(rp.Type()).Interface()
			if err := meta.PrimitiveDecode(tmp[k], val); err != nil {
				fmt.Println(err)
			}
			er.Plugins[k] = val
		}
	}
	return err
	// return nil
}

func (er *EdgeRouter) Run() {
	cfgPath := flag.String("config", er.Name+".conf", "config file path for the edge router")
	flag.Parse()
	var err error
	ctx := context.Background()
	if err = er.ConfigBy(*cfgPath); err == nil {
		for name, servers := range er.servers {
			plugin := er.Plugins[name]
			for _, server := range servers {
				fmt.Println("start", server)
				if ctx, err = server.Run(ctx, plugin); err != nil || ctx == nil {
					log.Fatal(ctx, err)
				}
			}
		}
	} else {
		log.Fatal(err)
	}
	for {
		select {
		case <-ctx.Done():
			fmt.Println("-")
		}
	}
	if err != nil {
		fmt.Println("[ERROR]", err)
	}
}
