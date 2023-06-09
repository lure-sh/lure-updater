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
