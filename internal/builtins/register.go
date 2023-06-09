package builtins

import (
	"net/http"

	"go.elara.ws/lure-updater/internal/config"
	"go.etcd.io/bbolt"
	"go.starlark.net/starlark"
	"go.starlark.net/starlarkjson"
)

type Options struct {
	Name   string
	DB     *bbolt.DB
	Config *config.Config
	Mux    *http.ServeMux
}

func Register(sd starlark.StringDict, opts *Options) {
	sd["run_every"] = starlark.NewBuiltin("run_every", runEvery)
	sd["sleep"] = starlark.NewBuiltin("sleep", sleep)
	sd["http"] = httpModule
	sd["regex"] = regexModule
	sd["store"] = storeModule(opts.DB, opts.Name)
	sd["updater"] = updaterModule(opts.Config)
	sd["log"] = logModule(opts.Name)
	sd["json"] = starlarkjson.Module
	sd["register_webhook"] = registerWebhook(opts.Mux, opts.Config, opts.Name)
}
