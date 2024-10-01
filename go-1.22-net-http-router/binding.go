package main

import (
	"encoding/json"
	"errors"
	"net/http"
	"reflect"
	"strconv"
)

// bodyからjsonを構造体にbindingする
func bindJSON(r *http.Request, v any) error {
	return json.NewDecoder(r.Body).Decode(v)
}

// uriから構造体にbindingする(intとstringのみ対応)
func bindURI(r *http.Request, obj any) error {
	if reflect.TypeOf(obj).Kind() != reflect.Ptr {
		return errors.New("not pointer")
	}
	t := reflect.TypeOf(obj).Elem()

	if t.Kind() != reflect.Struct {
		return errors.New("not struct")
	}

	v := reflect.ValueOf(obj).Elem()

	if !v.CanSet() {
		return errors.New("cannot set")
	}

	for i := 0; i < t.NumField(); i++ {
		tt := t.Field(i)
		name, ok := tt.Tag.Lookup("uri")
		if !ok {
			continue
		}
		vv := r.PathValue(name)

		tv := v.FieldByName(tt.Name)
		switch tt.Type.Kind() {
		case reflect.Int:
			n, err := strconv.ParseInt(vv, 10, 64)
			if err != nil {
				return err
			}
			tv.SetInt(n)
		case reflect.String:
			tv.SetString(vv)
		}
	}
	return nil
}
