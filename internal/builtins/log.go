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
	"strings"

	"go.elara.ws/logger"
	"go.elara.ws/logger/log"
	"go.starlark.net/starlark"
	"go.starlark.net/starlarkstruct"
)

func logModule(name string) *starlarkstruct.Module {
	return &starlarkstruct.Module{
		Name: "log",
		Members: starlark.StringDict{
			"debug": logDebug(name),
			"info":  logInfo(name),
			"warn":  logWarn(name),
			"error": logError(name),
		},
	}
}

func logDebug(name string) *starlark.Builtin {
	return starlark.NewBuiltin("log.debug", func(thread *starlark.Thread, b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
		return starlark.None, doLogEvent(name, "log.debug", log.Debug, thread, args, kwargs)
	})
}

func logInfo(name string) *starlark.Builtin {
	return starlark.NewBuiltin("log.info", func(thread *starlark.Thread, b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
		return starlark.None, doLogEvent(name, "log.info", log.Info, thread, args, kwargs)
	})
}

func logWarn(name string) *starlark.Builtin {
	return starlark.NewBuiltin("log.warn", func(thread *starlark.Thread, b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
		return starlark.None, doLogEvent(name, "log.warn", log.Warn, thread, args, kwargs)
	})
}

func logError(name string) *starlark.Builtin {
	return starlark.NewBuiltin("log.error", func(thread *starlark.Thread, b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
		return starlark.None, doLogEvent(name, "log.error", log.Error, thread, args, kwargs)
	})
}

func doLogEvent(pluginName, fnName string, eventFn func(msg string) logger.LogBuilder, thread *starlark.Thread, args starlark.Tuple, kwargs []starlark.Tuple) error {
	var msg string
	var fields *starlark.Dict
	err := starlark.UnpackArgs(fnName, args, kwargs, "msg", &msg, "fields??", &fields)
	if err != nil {
		return err
	}

	evt := eventFn("[" + pluginName + "] " + msg)

	if fields != nil {
		for _, key := range fields.Keys() {
			val, _, err := fields.Get(key)
			if err != nil {
				return err
			}
			keyStr := strings.Trim(key.String(), `"`)
			evt = evt.Stringer(keyStr, val)
		}
	}

	if fnName == "log.debug" {
		evt = evt.Stringer("pos", thread.CallFrame(1).Pos)
	}

	evt.Send()

	return nil
}
