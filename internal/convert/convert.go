package convert

import (
	"errors"
	"fmt"
	"reflect"

	"go.starlark.net/starlark"
)

var ErrInvalidType = errors.New("unknown type")

func Convert(v any) (starlark.Value, error) {
	if v == nil {
		return starlark.None, nil
	}
	val := reflect.ValueOf(v)
	kind := val.Kind()
	for kind == reflect.Pointer || kind == reflect.Interface {
		val = val.Elem()
	}
	return convert(val)
}

func convert(val reflect.Value) (starlark.Value, error) {
	switch val.Kind() {
	case reflect.Interface:
		return convert(val.Elem())
	case reflect.Int, reflect.Int64, reflect.Int32, reflect.Int16, reflect.Int8:
		return starlark.MakeInt64(val.Int()), nil
	case reflect.Uint, reflect.Uint64, reflect.Uint32, reflect.Uint16, reflect.Uint8:
		return starlark.MakeUint64(val.Uint()), nil
	case reflect.Float64, reflect.Float32:
		return starlark.Float(val.Float()), nil
	case reflect.Bool:
		return starlark.Bool(val.Bool()), nil
	case reflect.String:
		return starlark.String(val.String()), nil
	case reflect.Slice, reflect.Array:
		return convertSlice(val)
	case reflect.Map:
		return convertMap(val)
	default:
		return nil, fmt.Errorf("%w: %s", ErrInvalidType, val.Type())
	}
}

func convertSlice(val reflect.Value) (starlark.Value, error) {
	// Detect byte slice
	if val.Type().Elem().Kind() == reflect.Uint8 {
		return starlark.Bytes(val.Bytes()), nil
	}

	elems := make([]starlark.Value, val.Len())

	for i := 0; i < val.Len(); i++ {
		elem, err := convert(val.Index(i))
		if err != nil {
			return nil, err
		}
		elems[i] = elem
	}

	return starlark.NewList(elems), nil
}

func convertMap(val reflect.Value) (starlark.Value, error) {
	dict := starlark.NewDict(val.Len())
	iter := val.MapRange()
	for iter.Next() {
		k, err := convert(iter.Key())
		if err != nil {
			return nil, err
		}

		v, err := convert(iter.Value())
		if err != nil {
			return nil, err
		}

		err = dict.SetKey(k, v)
		if err != nil {
			return nil, err
		}
	}
	return dict, nil
}
