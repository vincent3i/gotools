package convertor

import (
	"encoding/binary"
	"fmt"
	log "github.com/vincent3i/gotools/logtool"
	"math"
	"reflect"
	"strconv"
	"strings"
	"time"
	"unsafe"
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

// @see http://studygolang.com/articles/2909
func BytesToStrUnsafe(rawBytes []byte) string {
	return *(*string)(unsafe.Pointer(&rawBytes))
}

// @see http://shinriyo.hateblo.jp/entry/2015/02/19/Go%E8%A8%80%E8%AA%9E%E3%81%AE%E5%B0%8F%E6%95%B0%E7%82%B9%E3%81%AE%E5%9B%9B%E6%8D%A8%E4%BA%94%E5%85%A5
func Round(f float64, places int) float64 {
	shift := math.Pow(10, float64(places))
	return math.Floor(f*shift+.5) / shift
}

func StrToFloat(value string) float64 {
	float, err := strconv.ParseFloat(value, 64)
	if err != nil {
		return 0
	}
	return float
}

// 先尝试转浮点，然后再转整型
func StrToFInt(value string) int {
	float := StrToFloat(value)
	return int(float)
}

func StrToFInt64(value string) int64 {
	float := StrToFloat(value)
	return int64(float)
}

func StrToInt(value string) int {
	number, err := strconv.Atoi(value)
	if err == nil {
		return number
	}
	return 0
}

// 取得整型值所表达的布尔类型
// 不同于golang提供的ParseBool
func StrToIntBool(value string) bool {
	// "" => false
	if len(value) <= 0 {
		return false
	}
	i := StrToFInt(value) // 先尝试尽量取得这个字符串表示的整型值
	if i <= 0 {
		return false
	}
	return true
}

func IntToTime(value int64) time.Time {
	if value <= 0 {
		return time.Time{}
	}
	return time.Unix(value, 0)
}

// 转型最好优先转型到最大的值，然后再往底缩进
// 更精确的做法，应该是根据位长，来做出适当的判断但过度优化，又不如直接用go提供一些方法
// 所以这个方法只是确保值的有效性转换，性能在能考虑的条件下，才考虑
func AnyToInt64(v interface{}) int64 {
	switch v.(type) {
	// #1
	// 这个nil必须保持，不然在检索结构的方法时，有可能会陷入死循环
	case nil:
		return 0
	// #2
	case bool:
		if v == true {
			return 1
		}
	// #3
	// 这玩意可真算不上优雅啊，go怎么就没有泛型呢？
	case int:
		if conv, ok := v.(int); ok {
			return int64(conv)
		}
	case int8:
		if conv, ok := v.(int8); ok {
			return int64(conv)
		}
	case int16:
		if conv, ok := v.(int16); ok {
			return int64(conv)
		}
	case int32:
		if conv, ok := v.(int32); ok {
			return int64(conv)
		}
	case int64:
		if conv, ok := v.(int64); ok {
			return int64(conv)
		}
	case uint:
		if conv, ok := v.(uint); ok {
			return int64(conv)
		}
	case uint8:
		if conv, ok := v.(uint8); ok {
			return int64(conv)
		}
	case uint16:
		if conv, ok := v.(uint16); ok {
			return int64(conv)
		}
	case uint32:
		if conv, ok := v.(uint32); ok {
			return int64(conv)
		}
	case uint64:
		if conv, ok := v.(uint64); ok {
			return int64(conv)
		} // 这里仍然是有问题
	// #4
	case float32:
		if conv, ok := v.(float32); ok {
			return int64(conv)
		}
	case float64:
		if conv, ok := v.(float64); ok {
			return int64(conv)
		}
	// #5
	case string:
		if conv, ok := v.(string); ok {
			return StrToFInt64(conv)
		}
	// #6
	case time.Time:
		if conv, ok := v.(time.Time); ok {
			return conv.Unix()
		}
	// #999
	default:
		// 数组、切片、Map转类型是什么类型呢？
		return AnyToInt64(CallAnyStructMethod(v, "Int"))
	}
	return 0
}

func AnyToInt(v interface{}) int {
	return int(AnyToInt64(v))
}

// 小位数的整型，还是经常会用到的
// 溢出值，以溢出的最大值处理，而不要作为负数处理
// 256 / 2
func AnyToInt8(v interface{}) int8 {
	iv := AnyToInt64(v)
	if iv > 127 { // 0 - 127
		return int8(127)
	} else if iv < -128 { // -128 - -1
		return int8(-128)
	}
	return int8(iv)
}

// 小位数的整型，还是经常会用到的
// 溢出值，以溢出的最大值处理，而不要作为负数处理
// 65536 / 2
func AnyToInt16(v interface{}) int16 {
	iv := AnyToInt64(v)
	if iv > 32767 { // 0 - 127
		return int16(32767)
	} else if iv < -32768 { // -128 - -1
		return int16(-32768)
	}
	return int16(iv)
}

func AnyToUInt64(v interface{}) uint64 {
	return uint64(AnyToInt64(v))
}

// 无负数的小位数整型，其实也是很常用到的
func AnyToUInt8(v interface{}) uint8 {
	iv := AnyToInt64(v)
	if iv > 255 { // 0 - 127
		return uint8(255)
	} else if iv < 0 { // -128 - -1
		return uint8(0)
	}
	return uint8(iv)
}

// 无负数的小位数整型，其实也是很常用到的，32的就自己手动转吧，uint32能表达的值意境非常大了
func AnyToUInt16(v interface{}) uint16 {
	iv := AnyToInt64(v)
	if iv > 65535 { // 0 - 127
		return uint16(65535)
	} else if iv < 0 { // -128 - -1
		return uint16(0)
	}
	return uint16(iv)
}

// 注意，所有其他的AnyTo转换，都不处理[]byte，因为实际上[]byte的情况会比较复杂，他可能包含了encode/gob的编码格式，也可能是json格式
// 也可能用户自己打包的，所以我们不做任何处理
// 但AnyToStr的话还是要处理，尝试最简单的转换
func AnyToStr(v interface{}) string {
	switch v.(type) {
	// #1
	// 这个nil必须保持，不然在检索结构的方法时，有可能会陷入死循环
	case nil:
		return ""
	// #2
	// 布尔类型，应该返回个啥呢？真头疼，暂时先返回一个1吧，总比返回了true好
	case bool:
		if v == true {
			return "1"
		}
	// #3
	// 这玩意可真算不上优雅啊，go怎么就没有泛型呢？
	case int:
		if conv, ok := v.(int); ok {
			return strconv.Itoa(conv)
		}
	case int8:
		if conv, ok := v.(int8); ok {
			return strconv.Itoa(int(conv))
		}
	case int16:
		if conv, ok := v.(int16); ok {
			return strconv.Itoa(int(conv))
		}
	case int32:
		if conv, ok := v.(int32); ok {
			return strconv.Itoa(int(conv))
		} // 32bit 64bit系统都能涵盖了这个值
	case int64:
		if conv, ok := v.(int64); ok {
			return fmt.Sprint(conv)
		}
	case uint:
		if conv, ok := v.(uint); ok {
			return fmt.Sprint(conv)
		}
	case uint8:
		if conv, ok := v.(uint8); ok {
			return strconv.Itoa(int(conv))
		}
	case uint16:
		if conv, ok := v.(uint16); ok {
			return strconv.Itoa(int(conv))
		}
	case uint32:
		if conv, ok := v.(uint32); ok {
			return fmt.Sprint(conv)
		} // 32无负数整型，转int就少了一截了
	case uint64:
		if conv, ok := v.(uint64); ok {
			return fmt.Sprint(conv)
		} // 64位无负数整型，就更加是少了一截了。
	// #4
	case float32:
		if conv, ok := v.(float32); ok {
			return strconv.FormatFloat(float64(conv), 'f', -1, 64)
		}
	case float64:
		if conv, ok := v.(float64); ok {
			return strconv.FormatFloat(conv, 'f', -1, 64)
		}
	// #5
	case []byte:
		if conv, ok := v.([]byte); ok {
			return string(conv)
		}
	case string:
		if conv, ok := v.(string); ok {
			return conv
		}
	// #6
	case time.Time:
		if conv, ok := v.(time.Time); ok {
			return conv.String()
		}
	// #999
	default:
		// 数组、切片、Map转类型是什么类型呢？
		return AnyToStr(CallAnyStructMethod(v, "String"))
	}
	return ""
}

func AnyToFloat(v interface{}) float64 {
	switch v.(type) {
	// #1
	// 这个nil必须保持，不然在检索结构的方法时，有可能会陷入死循环
	case nil:
		return 0
	// #2
	case bool:
		if v == true {
			return 1
		}
	// #3
	// 这玩意可真算不上优雅啊，go怎么就没有泛型呢？
	case int:
		if conv, ok := v.(int); ok {
			return float64(conv)
		}
	case int8:
		if conv, ok := v.(int8); ok {
			return float64(conv)
		}
	case int16:
		if conv, ok := v.(int16); ok {
			return float64(conv)
		}
	case int32:
		if conv, ok := v.(int32); ok {
			return float64(conv)
		}
	case int64:
		if conv, ok := v.(int64); ok {
			return float64(conv)
		}
	case uint:
		if conv, ok := v.(uint); ok {
			return float64(conv)
		}
	case uint8:
		if conv, ok := v.(uint8); ok {
			return float64(conv)
		}
	case uint16:
		if conv, ok := v.(uint16); ok {
			return float64(conv)
		}
	case uint32:
		if conv, ok := v.(uint32); ok {
			return float64(conv)
		}
	case uint64:
		if conv, ok := v.(uint64); ok {
			return float64(conv)
		} // 这里仍然是有问题
	// #4
	case float32:
		if conv, ok := v.(float32); ok {
			return float64(conv)
		}
	case float64:
		if conv, ok := v.(float64); ok {
			return float64(conv)
		}
	// #5
	case string:
		if conv, ok := v.(string); ok {
			return StrToFloat(conv)
		}
	// #6
	case time.Time:
		if conv, ok := v.(time.Time); ok {
			return float64(conv.Unix())
		}
	// #999
	default:
		// 数组、切片、Map转类型是什么类型呢？
		return AnyToFloat(CallAnyStructMethod(v, "Float"))
	}
	return 0
}

func AnyToRound(v interface{}, places int) float64 {
	return Round(AnyToFloat(v), places)
}

func CallAnyStructMethod(v interface{}, method string) interface{} {
	ref := reflect.ValueOf(v)
	refKind := ref.Kind()
	if refKind == reflect.Ptr {
		refKind = ref.Elem().Kind()
	}
	// 如果是结构的话，尝试检索一下他是否有Int、ToInt的函数
	if refKind == reflect.Struct {
		fn := ref.MethodByName(method)
		if fn.IsValid() {
			rs := fn.Call(nil)
			if len(rs) > 0 {
				return rs[0].Interface()
			}
		}
	}
	return nil
}

// Kind在reflect已经有比较明确的枚举，能比较方便的去进行比较
func KindOf(v interface{}) reflect.Kind {
	ref := reflect.ValueOf(v)
	refKind := ref.Kind()
	if refKind == reflect.Ptr {
		return ref.Elem().Kind()
	}
	return refKind
}

func ValueOf(v interface{}) reflect.Value {
	ref := reflect.ValueOf(v)
	refKind := ref.Kind()
	if refKind == reflect.Ptr {
		return ref.Elem()
	}
	return ref
}

func AnyToBool(v interface{}) bool {
	switch v.(type) {
	// #1
	// 这个nil必须保持，不然在检索结构的方法时，有可能会陷入死循环
	case nil:
		return false
	// #2
	case bool:
		if v == true {
			return true
		}
	// #3
	// 这玩意可真算不上优雅啊，go怎么就没有泛型呢？
	case int:
		if conv, ok := v.(int); ok {
			return conv > 0
		}
	case int8:
		if conv, ok := v.(int8); ok {
			return conv > 0
		}
	case int16:
		if conv, ok := v.(int16); ok {
			return conv > 0
		}
	case int32:
		if conv, ok := v.(int32); ok {
			return conv > 0
		}
	case int64:
		if conv, ok := v.(int64); ok {
			return conv > 0
		}
	case uint:
		if conv, ok := v.(uint); ok {
			return conv > 0
		}
	case uint8:
		if conv, ok := v.(uint8); ok {
			return conv > 0
		}
	case uint16:
		if conv, ok := v.(uint16); ok {
			return conv > 0
		}
	case uint32:
		if conv, ok := v.(uint32); ok {
			return conv > 0
		}
	case uint64:
		if conv, ok := v.(uint64); ok {
			return conv > 0
		}
	// #4
	case float32:
		if conv, ok := v.(float32); ok {
			return conv > 0
		}
	case float64:
		if conv, ok := v.(float64); ok {
			return conv > 0
		}
	// #5
	case string:
		if conv, ok := v.(string); ok {
			return len(conv) > 0
		}
	// #6
	case time.Time:
		if conv, ok := v.(time.Time); ok {
			return IsValidTime(conv)
		}
	// #999
	default:
		kind := KindOf(v)
		val := ValueOf(v)
		if kind == reflect.Slice || kind == reflect.Array || kind == reflect.Map {
			return val.Len() > 0
		} else if kind == reflect.Struct {
			// 结构真的没什么好判断的，只要不是无效的结构，就只好返回true了
			if val.IsValid() {
				return true
			}
		}
	}
	return false
}

// go的时间
// var t = time.Time {} => t.IsZero() => true
// var t = time.Unix(0, 0) => t.IsZero() => false
func IsValidTime(t time.Time) bool {
	return !t.IsZero() && t.Unix() > 0
}
