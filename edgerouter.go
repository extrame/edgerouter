package edgerouter

import (
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"reflect"

	"strings"

	"context"

	"github.com/BurntSushi/toml"
	"github.com/golang/glog"
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
	return er
}

type EdgeRouter struct {
	Name       string
	Plugins    map[string]interface{}
	components map[string]Component
	configged  bool
}

func (er *EdgeRouter) ConfigBy(filename string) error {
	bs, err := ioutil.ReadFile(filename)
	if err == nil {
		return er.ConfigByString(string(bs))
	}
	return err
}

func (er *EdgeRouter) ConfigByString(str string) error {
	var tmp = make(map[string]toml.Primitive)
	var err error
	var meta toml.MetaData

	if len(er.Plugins) == 0 {
		return errors.New("no plugin for edge router")
	}

	if meta, err = toml.Decode(str, &tmp); err == nil {
		er.components = make(map[string]Component)
		for k, plugin := range er.Plugins {
			glog.Infoln("start parse plugin", k)
			rp := reflect.ValueOf(plugin)
			var com Component
			for i := 0; i < rp.NumField(); i++ {
				fp := rp.Field(i)
				ft := fp.Type()
				val := reflect.New(ft).Interface()
				if err := meta.PrimitiveDecode(tmp[k], val); err == nil {
					if vc, ok := val.(Controller); ok {
						com.Ctrl = vc
					}
					if vs, ok := val.(Server); ok {
						com.Server = vs
					}
					if vt, ok := val.(Transport); ok {
						com.Trans = vt
					}
				}
				initCheck(val)
			}
			val := reflect.New(rp.Type()).Interface()
			if err = meta.PrimitiveDecode(tmp[k], val); err != nil {
				goto errHandle
			}
			com.Ctrl.SetTransport(com.Trans)
			if err = com.Ctrl.SetHandler(val); err != nil {
				goto errHandle
			}
			com.Trans.SetController(com.Ctrl)
			er.components[k] = com
			er.Plugins[k] = val
		}
	}
	er.configged = true
	return err
errHandle:
	glog.Errorln(err)
	return err
}

func (er *EdgeRouter) Run() {
	if !er.configged {
		cfgPath := flag.String("config", er.Name+".conf", "config file path for the edge router")
		flag.Parse()
		var err error
		if err = er.ConfigBy(*cfgPath); err != nil {
			glog.Errorln(err)
		} else {
			er.configged = true
		}
	}

	if er.configged {
		glog.Infoln("run edge router")

		ctx := context.Background()

		for name, component := range er.components {
			glog.Infof("run server (%s) AS (%s)", name, component)
			if component.Server != nil {
				go component.Server.Run()
			}
			go component.Ctrl.Run()
			// plugin := er.Plugins[name]
			// for _, server := range servers {
			// 	fmt.Println("start", server)
			// 	if ctx, err = server.Run(ctx, plugin); err != nil || ctx == nil {
			// 		log.Fatal(ctx, err)
			// 	}
			// }
		}
		for {
			select {
			case <-ctx.Done():
				fmt.Println("-")
			}
		}
	} else {
		glog.Errorln("not configged correctly")
	}
}

func initCheck(val interface{}) {
	if iv, ok := val.(Initer); ok {
		if err := iv.Init(); err != nil {
			glog.Errorf("%T(%v) need init,occur error(%s)", val, val, err)
		}
	}
}
