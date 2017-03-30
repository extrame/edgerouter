package edgerouter

import (
	"fmt"
	"log"
	"reflect"
	"strings"

	"strconv"

	"github.com/extrame/edgerouter/math"
	"github.com/golang/glog"
)

type BytesMessage struct {
	Message []byte
	To      string
	For     string
}

func NewBytesMessage(v interface{}, to string) *BytesMessage {
	msg := new(BytesMessage)
	msg.Message = marshall(v)
	msg.To = to
	if dv, ok := v.(Device); ok {
		msg.For = dv.DeviceID()
	} else {
		glog.Fatal("msg is not a Device interface")
	}
	return msg
}

func Unmarsall(bts []byte, v interface{}) error {
	rp := reflect.ValueOf(v)
	if rp.Type().Kind() == reflect.Ptr {
		rp = rp.Elem()
	}
	rt := rp.Type()
	if rt.Kind() == reflect.Struct {
		var offset = 0
		for i := 0; i < rt.NumField(); i++ {
			var btsSlice []byte
			fv := rp.Field(i)
			ft := rt.Field(i)
			tag := ft.Tag.Get("er")
			tags := strings.Split(tag, ",")
			for _, tag := range tags {
				kv := strings.Split(tag, "=")
				switch kv[0] {
				case "length":
					length, _ := strconv.Atoi(kv[1])
					if offset+length > len(bts) {
						log.Fatal("override array length", offset, length, len(bts))
					} else if offset+length == len(bts) {
						btsSlice = bts[offset:]
					} else {
						btsSlice = bts[offset : offset+length]
					}
					offset += length
				case "calculator":
					_, tokens := math.Lexer(kv[1])
					intSlice := make([]int, len(btsSlice))
					for k, v := range btsSlice {
						intSlice[k] = int(v)
					}
					fv.SetInt(int64(math.Parse(tokens).Val(intSlice)))
					continue
				default:
					var i int64
					var err error
					if i, err = strconv.ParseInt(kv[0], 0, 8); err != nil {
						i = int64(kv[0][0])
					}
					if i != int64(bts[offset]) {
						return fmt.Errorf("bit[%d] is not equal as expected, wanted (%d),but is (%d) in %v", offset, i, bts[offset], bts)
					} else {
						switch fv.Type().Kind() {
						case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
							fv.SetUint(uint64(i))
						case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
							fv.SetInt(i)
						}
					}
					offset++
				}
			}
		}
	}
	return nil
}

func marshall(v interface{}) []byte {
	var bts = make([]byte, 0)
	rp := reflect.ValueOf(v)
	if rp.Type().Kind() == reflect.Ptr {
		rp = rp.Elem()
	}
	rt := rp.Type()
	if rt.Kind() == reflect.Struct {
		for i := 0; i < rt.NumField(); i++ {
			fv := rp.Field(i)
			ft := rt.Field(i)
			tag := ft.Tag.Get("er")
			tags := strings.Split(tag, ",")
			handled := false
			val := fv.Interface()
			for _, tag := range tags {
				if tag != "" {
					kv := strings.Split(tag, "=")
					switch kv[0] {
					case "length":
						length, _ := strconv.Atoi(kv[1])
						switch ft.Type.Kind() {
						case reflect.Slice:
							val = fv.Slice(0, length).Interface()
						case reflect.String:
							val = fv.Slice(0, length).Interface()
						}
					default:
						if i, err := strconv.ParseInt(kv[0], 0, 8); err == nil {
							bts = append(bts, byte(i))
						} else {
							//try use string as ascii
							bts = append(bts, []byte(kv[0])...)
						}
						handled = true
					}
				}
			}
			if !handled {
				switch ft.Type.Kind() {
				case reflect.Slice:
					bts = append(bts, val.([]byte)...)
				case reflect.String:
					bts = append(bts, []byte(val.(string))...)
				}
			}
		}
	}
	return bts
}
