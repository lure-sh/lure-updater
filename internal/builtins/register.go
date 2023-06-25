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
	sd["utils"] = utilsModule
	sd["html"] = htmlModule
	sd["register_webhook"] = registerWebhook(opts.Mux, opts.Config, opts.Name)
}
