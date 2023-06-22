/*
 * LURE Updater - Automated updater bot for LURE packages
 * Copyright (C) 2023 Elara Musayelyan
 *
 * This program is free software: you can redistribute it and/or modify
 * it under the terms of the GNU General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * This program is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU General Public License for more details.
 *
 * You should have received a copy of the GNU General Public License
 * along with this program.  If not, see <http://www.gnu.org/licenses/>.
 */

package builtins

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"

	"go.elara.ws/logger/log"
	"go.elara.ws/lure-updater/internal/config"
	"go.starlark.net/starlark"
	"go.starlark.net/starlarkstruct"
	"golang.org/x/crypto/bcrypt"
)

const maxBodySize = 16384

var (
	ErrInvalidBodyType   = errors.New("invalid body type")
	ErrInvalidHdrKeyType = errors.New("invalid header key type")
	ErrInvalidHdrVal     = errors.New("invalid header value type")
	ErrInvalidType       = errors.New("invalid type")
	ErrInsecureWebhook   = errors.New("secure webhook missing authorization")
	ErrIncorrectPassword = errors.New("incorrect password")
)

var httpModule = &starlarkstruct.Module{
	Name: "http",
	Members: starlark.StringDict{
		"get":  starlark.NewBuiltin("http.get", httpGet),
		"post": starlark.NewBuiltin("http.post", httpPost),
		"put":  starlark.NewBuiltin("http.put", httpPut),
		"head": starlark.NewBuiltin("http.head", httpHead),
	},
}

func httpGet(thread *starlark.Thread, b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	return makeRequest("http.get", http.MethodGet, args, kwargs, thread)
}

func httpPost(thread *starlark.Thread, b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	return makeRequest("http.post", http.MethodPost, args, kwargs, thread)
}

func httpPut(thread *starlark.Thread, b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	return makeRequest("http.put", http.MethodPut, args, kwargs, thread)
}

func httpHead(thread *starlark.Thread, b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	return makeRequest("http.head", http.MethodHead, args, kwargs, thread)
}

type starlarkBodyReader struct {
	io.Reader
}

func (sbr *starlarkBodyReader) Unpack(v starlark.Value) error {
	switch v := v.(type) {
	case starlark.String:
		sbr.Reader = strings.NewReader(string(v))
	case starlark.Bytes:
		sbr.Reader = strings.NewReader(string(v))
	default:
		return fmt.Errorf("%w: %s", ErrInvalidBodyType, v.Type())
	}
	return nil
}

func newBodyReader() *starlarkBodyReader {
	return &starlarkBodyReader{
		Reader: bytes.NewReader(nil),
	}
}

type starlarkHeaders struct {
	http.Header
}

func (sh *starlarkHeaders) Unpack(v starlark.Value) error {
	dict, ok := v.(*starlark.Dict)
	if !ok {
		return fmt.Errorf("%w: %s", ErrInvalidType, v.Type())
	}

	sh.Header = make(http.Header, dict.Len())
	for _, key := range dict.Keys() {
		keyStr, ok := key.(starlark.String)
		if !ok {
			return fmt.Errorf("%w: %s", ErrInvalidHdrKeyType, key.Type())
		}

		val, _, _ := dict.Get(key)
		list, ok := val.(*starlark.List)
		if !ok {
			return fmt.Errorf("%w: %s", ErrInvalidHdrVal, val.Type())
		}

		hdrVals := make([]string, list.Len())
		for i := 0; i < list.Len(); i++ {
			hdrVal, ok := list.Index(i).(starlark.String)
			if !ok {
				return fmt.Errorf("%w: %s", ErrInvalidHdrVal, list.Index(i).Type())
			}

			hdrVals[i] = string(hdrVal)
		}

		sh.Header[string(keyStr)] = hdrVals
	}

	return nil
}

func makeRequest(name, method string, args starlark.Tuple, kwargs []starlark.Tuple, thread *starlark.Thread) (starlark.Value, error) {
	var (
		url      string
		redirect = true
		headers  = &starlarkHeaders{}
		body     = newBodyReader()
	)
	err := starlark.UnpackArgs(name, args, kwargs, "url", &url, "redirect??", &redirect, "headers??", headers, "body??", body)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return nil, err
	}
	req.Header = headers.Header

	client := http.DefaultClient
	if !redirect {
		client = &http.Client{
			CheckRedirect: func(req *http.Request, via []*http.Request) error {
				return http.ErrUseLastResponse
			},
		}
	}

	log.Debug("Making HTTP request").Str("url", url).Str("method", req.Method).Bool("redirect", redirect).Stringer("pos", thread.CallFrame(1).Pos).Send()

	res, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	log.Debug("Got HTTP response").Str("host", res.Request.URL.Host).Int("code", res.StatusCode).Stringer("pos", thread.CallFrame(1).Pos).Send()

	return starlarkResponse(res), nil
}

func starlarkResponse(res *http.Response) *starlarkstruct.Struct {
	return starlarkstruct.FromStringDict(starlarkstruct.Default, starlark.StringDict{
		"code":    starlark.MakeInt(res.StatusCode),
		"headers": starlarkStringSliceMap(res.Header),
		"body":    newStarlarkReader(res.Body),
	})
}

func starlarkRequest(req *http.Request) *starlarkstruct.Struct {
	return starlarkstruct.FromStringDict(starlarkstruct.Default, starlark.StringDict{
		"method":      starlark.String(req.Method),
		"remote_addr": starlark.String(req.RemoteAddr),
		"headers":     starlarkStringSliceMap(req.Header),
		"query":       starlarkStringSliceMap(req.URL.Query()),
		"body":        newStarlarkReader(req.Body),
	})
}

func starlarkStringSliceMap(ssm map[string][]string) *starlark.Dict {
	dict := starlark.NewDict(len(ssm))
	for key, vals := range ssm {
		sVals := make([]starlark.Value, len(vals))
		for i, val := range vals {
			sVals[i] = starlark.String(val)
		}
		dict.SetKey(starlark.String(key), starlark.NewList(sVals))
	}
	return dict
}

func registerWebhook(mux *http.ServeMux, cfg *config.Config, pluginName string) *starlark.Builtin {
	return starlark.NewBuiltin("register_webhook", func(thread *starlark.Thread, b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
		var fn *starlark.Function
		secure := true
		err := starlark.UnpackArgs("register_webhook", args, kwargs, "function", &fn, "secure??", &secure)
		if err != nil {
			return nil, err
		}

		if !secure {
			log.Warn("Plugin is registering an insecure webhook").Str("plugin", pluginName).Send()
		}

		path := "/webhook/" + pluginName + "/" + fn.Name()
		mux.HandleFunc(path, webhookHandler(pluginName, secure, cfg, thread, fn))
		log.Debug("Registered webhook").Str("path", path).Str("function", fn.Name()).Stringer("pos", thread.CallFrame(1).Pos).Send()
		return starlark.None, nil
	})
}

func webhookHandler(pluginName string, secure bool, cfg *config.Config, thread *starlark.Thread, fn *starlark.Function) http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		defer req.Body.Close()

		res.Header().Add("X-Updater-Plugin", pluginName)

		if secure {
			err := verifySecure(cfg.Webhook.PasswordHash, pluginName, req)
			if err != nil {
				log.Error("Error verifying webhook").Err(err).Send()
				res.WriteHeader(http.StatusForbidden)
				_, _ = io.WriteString(res, err.Error())
				return
			}
		}

		log.Debug("Calling webhook function").Str("name", fn.Name()).Stringer("pos", fn.Position()).Send()
		val, err := starlark.Call(thread, fn, starlark.Tuple{starlarkRequest(req)}, nil)
		if err != nil {
			log.Error("Error while executing webhook").Err(err).Stringer("pos", fn.Position()).Send()
			res.WriteHeader(http.StatusInternalServerError)
			_, _ = io.WriteString(res, err.Error())
			return
		}

		switch val := val.(type) {
		case starlark.NoneType:
			res.WriteHeader(http.StatusOK)
		case starlark.Int:
			var code int
			err = starlark.AsInt(val, &code)
			if err == nil {
				res.WriteHeader(code)
			} else {
				res.WriteHeader(http.StatusOK)
			}
		case starlark.String, starlark.Bytes:
			body := newBodyReader()
			err = body.Unpack(val)
			if err != nil {
				log.Error("Error unpacking returned body").Err(err).Send()
				return
			}
			_, err = io.Copy(res, body)
			if err != nil {
				log.Error("Error writing body").Err(err).Send()
				return
			}
		case *starlark.Dict:
			code := http.StatusOK
			codeVal, ok, _ := val.Get(starlark.String("code"))
			if ok {
				err = starlark.AsInt(codeVal, &code)
				if err != nil {
					log.Error("Error decoding returned status code").Err(err).Send()
					return
				}
				res.WriteHeader(code)
			}

			body := newBodyReader()
			bodyVal, ok, _ := val.Get(starlark.String("body"))
			if ok {
				err = body.Unpack(bodyVal)
				if err != nil {
					log.Error("Error unpacking returned body").Err(err).Send()
					return
				}
				_, err = io.Copy(res, body)
				if err != nil {
					log.Error("Error writing body").Err(err).Send()
					return
				}
			}
		}
	}
}

func verifySecure(pwdHash, pluginName string, req *http.Request) error {
	var pwd []byte
	if _, pwdStr, ok := req.BasicAuth(); ok {
		pwdStr = strings.TrimSpace(pwdStr)
		pwd = []byte(pwdStr)
	} else if hdrStr := req.Header.Get("Authorization"); hdrStr != "" {
		hdrStr = strings.TrimPrefix(hdrStr, "Bearer")
		hdrStr = strings.TrimSpace(hdrStr)
		pwd = []byte(hdrStr)
	} else {
		log.Warn("Insecure webhook request").
			Str("from", req.RemoteAddr).
			Str("plugin", pluginName).
			Send()
		return ErrInsecureWebhook
	}

	if err := bcrypt.CompareHashAndPassword([]byte(pwdHash), pwd); err != nil {
		return ErrIncorrectPassword
	}

	return nil
}
