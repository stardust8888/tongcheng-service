package jsoniter

import "github.com/json-iterator/go"

var cjson = jsoniter.ConfigCompatibleWithStandardLibrary

func Marshal(i any) ([]byte, error) {
	return cjson.Marshal(i)
}

func MarshalToString(i any) (string, error) {
	return cjson.MarshalToString(i)
}

func Unmarshal(data []byte, v any) error {
	return cjson.Unmarshal(data, v)
}

func UnmarshalFromString(str string, v any) error {
	return cjson.UnmarshalFromString(str, v)
}

func Get(data []byte, path ...any) any {
	return cjson.Get(data, path...)
}
