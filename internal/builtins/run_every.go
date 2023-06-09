package builtins

import (
	"sync"
	"time"

	"go.elara.ws/logger/log"
	"go.starlark.net/starlark"
	"go.starlark.net/starlarkstruct"
)

var (
	tickerMtx   = &sync.Mutex{}
	tickerCount = 0
	tickers     = map[int]*time.Ticker{}
)

func runEvery(thread *starlark.Thread, b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	var every string
	var fn *starlark.Function
	err := starlark.UnpackArgs("run_every", args, kwargs, "every", &every, "function", &fn)
	if err != nil {
		return nil, err
	}

	d, err := time.ParseDuration(every)
	if err != nil {
		return nil, err
	}

	tickerMtx.Lock()
	t := time.NewTicker(d)
	handle := tickerCount
	tickers[handle] = t
	tickerCount++
	tickerMtx.Unlock()
	log.Debug("Created new ticker").Int("handle", handle).Str("duration", every).Stringer("pos", thread.CallFrame(1).Pos).Send()

	go func() {
		for range t.C {
			log.Debug("Calling scheduled function").Str("name", fn.Name()).Stringer("pos", fn.Position()).Send()
			_, err := starlark.Call(thread, fn, nil, nil)
			if err != nil {
				log.Warn("Error while executing scheduled function").Str("name", fn.Name()).Stringer("pos", fn.Position()).Err(err).Send()
			}
		}
	}()

	return newTickerHandle(handle), nil
}

func sleep(thread *starlark.Thread, b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	var duration string
	err := starlark.UnpackArgs("sleep", args, kwargs, "duration", &duration)
	if err != nil {
		return nil, err
	}

	d, err := time.ParseDuration(duration)
	if err != nil {
		return nil, err
	}

	log.Debug("Sleeping").Str("duration", duration).Stringer("pos", thread.CallFrame(1).Pos).Send()
	time.Sleep(d)
	return starlark.None, nil
}

func stopTicker(handle int) *starlark.Builtin {
	return starlark.NewBuiltin("stop", func(thread *starlark.Thread, b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
		tickerMtx.Lock()
		tickers[handle].Stop()
		delete(tickers, handle)
		tickerMtx.Unlock()
		log.Debug("Stopped ticker").Int("handle", handle).Stringer("pos", thread.CallFrame(1).Pos).Send()
		return starlark.None, nil
	})
}

func newTickerHandle(handle int) starlark.Value {
	return starlarkstruct.FromStringDict(starlarkstruct.Default, starlark.StringDict{
		"stop": stopTicker(handle),
	})
}
