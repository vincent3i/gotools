package convertor

import (
	"encoding/binary"
	"fmt"
	log "github.com/vincent3i/gotools/logtool"
	"reflect"
	"strconv"
	"strings"
	"time"
)

func Int64ToBytes(i int64) []byte {
	var buf = make([]byte, 8)
	binary.BigEndian.PutUint64(buf, uint64(i))
	return buf
}

func BytesToInt64(buf []byte) int64 {
	return int64(binary.BigEndian.Uint64(buf))
}

// convert any numeric value to int64
func ToInt64(value interface{}) (d int64, err error) {
	val := reflect.ValueOf(value)
	switch value.(type) {
	case int, int8, int16, int32, int64:
		d = val.Int()
	case uint, uint8, uint16, uint32, uint64:
		d = int64(val.Uint())
	default:
		err = fmt.Errorf("ToInt64 need numeric not `%T`", value)
	}
	return
}

//see https://github.com/astaxie/gor/blob/master/map2struct.go
func ToIntWithDefault(v interface{}, defaultValue int) int {
	if v == nil {
		return defaultValue
	}
	if i, ok := v.(int); ok {
		return i
	}
	if i2, ok := v.(int64); ok {
		return int(i2)
	}
	str := fmt.Sprintf("%v", v)
	i, err := strconv.Atoi(str)
	if err != nil {
		return defaultValue
	}
	return i
}

//see https://github.com/astaxie/gor/blob/master/map2struct.go
func ToInt64WithDefault(v interface{}, defaultValue int64) int64 {
	if v == nil {
		return defaultValue
	}
	if i, ok := v.(int64); ok {
		return i
	}
	if i2, ok := v.(int); ok {
		return int64(i2)
	}
	str := fmt.Sprintf("%v", v)
	i, err := strconv.ParseInt(str, 0, 64)
	if err != nil {
		return defaultValue
	}
	return i
}

//convert struct to map
func Struct2Map(obj interface{}) map[string]interface{} {
	var data = make(map[string]interface{})
	if obj == nil {
		return data
	}

	t := reflect.TypeOf(obj)
	v := reflect.ValueOf(obj)

	for i := 0; i < t.NumField(); i++ {
		data[t.Field(i).Name] = v.Field(i).Interface()
	}
	return data
}

//convert map to struct
//struct in val must be a pointer
func Map2Struct(m map[string]interface{}, val reflect.Value) {
	if m == nil {
		return
	}

	if val.Kind() == reflect.Ptr && val.Elem().Kind() != reflect.Struct {
		log.Error("Unsupport type of %s, only *struct can be supported.", val.Elem().Kind())
		return
	}

	origin_val := val
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}
	for k, v := range m {
		field := val.FieldByName(strings.Title(k))
		if !field.IsValid() {
			continue
		}
		if !field.CanSet() {
			log.Debug("CanSet = false", k, v)
			continue
		}
		log.Debug("Found key=%s, value=%v, type [%s]", k, v, field.Type().String())
		switch field.Kind() {
		case reflect.String:
			//log.Println("Ready to SetString", v)
			if _str, ok := v.(string); ok {
				field.SetString(_str)
			} else {
				field.SetString(fmt.Sprint("%v", v))
			}
		case reflect.Int:
			fallthrough
		case reflect.Int64:
			field.SetInt(ToInt64WithDefault(v, 0))
		case reflect.Slice: // 字段总是slice
			if _strs, ok := v.([]string); ok {
				field.Set(reflect.ValueOf(_strs))
			} else if _slice, ok := v.([]interface{}); ok {
				strs := make([]string, len(_slice))
				for i, vz := range _slice {
					strs[i] = vz.(string)
				}
				field.Set(reflect.ValueOf(strs))
			} else {
				log.Debug("Only []string is supported yet")
			}
		case reflect.Map:
			field.Set(reflect.ValueOf(v))
		case reflect.Ptr:
		// No support yet
		case reflect.Struct:
			if field.Type().String() == "time.Time" {
				if t, ok := v.(time.Time); ok {
					field.Set(reflect.ValueOf(t))
					break
				}
			}
			v2, ok := v.(map[string]interface{})
			if !ok {
				log.Debug("Not a map[string]interface{} key=%s value=%v", k, v)
				return
			}

			Map2Struct(v2, field)
		default:
			field.Set(reflect.ValueOf(v))
		}
	}
	_ = origin_val
}
