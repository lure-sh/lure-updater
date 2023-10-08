package builtins

import (
	"bufio"
	"encoding/json"
	"errors"
	"io"
	"strings"

	"github.com/vmihailenco/msgpack/v5"
	"lure.sh/lure-updater/internal/convert"
	"go.starlark.net/starlark"
	"go.starlark.net/starlarkstruct"
)

type starlarkReader struct {
	closeFunc func() error
	br        *bufio.Reader
	*starlarkstruct.Struct
}

func newStarlarkReader(r io.Reader) starlarkReader {
	sr := starlarkReader{br: bufio.NewReader(r)}

	if rc, ok := r.(io.ReadCloser); ok {
		sr.closeFunc = rc.Close
	}

	sr.Struct = starlarkstruct.FromStringDict(starlark.String("regex"), starlark.StringDict{
		"read":              starlark.NewBuiltin("reader.read", sr.read),
		"peek":              starlark.NewBuiltin("reader.peek", sr.peek),
		"discard":           starlark.NewBuiltin("reader.discard", sr.discard),
		"read_string":       starlark.NewBuiltin("reader.read_string", sr.readString),
		"read_until":        starlark.NewBuiltin("reader.read_until", sr.readUntil),
		"read_string_until": starlark.NewBuiltin("reader.read_string_until", sr.readStringUntil),
		"read_all":          starlark.NewBuiltin("reader.read_all", sr.readAll),
		"read_all_string":   starlark.NewBuiltin("reader.read_all_string", sr.readAllString),
		"read_json":         starlark.NewBuiltin("reader.read_json", sr.readJSON),
		"read_msgpack":      starlark.NewBuiltin("reader.read_msgpack", sr.readMsgpack),
		"close":             starlark.NewBuiltin("reader.close", sr.closeReader),
	})

	return sr
}

func (sr starlarkReader) read(thread *starlark.Thread, b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	var n int
	err := starlark.UnpackArgs("reader.read", args, kwargs, "n", &n)
	if err != nil {
		return nil, err
	}

	buf := make([]byte, n)
	_, err = io.ReadFull(sr.br, buf)
	if err != nil {
		return nil, err
	}

	return starlark.Bytes(buf), nil
}

func (sr starlarkReader) readString(thread *starlark.Thread, b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	var n int
	err := starlark.UnpackArgs("reader.read_string", args, kwargs, "n", &n)
	if err != nil {
		return nil, err
	}

	buf := make([]byte, n)
	_, err = io.ReadFull(sr.br, buf)
	if err != nil {
		return nil, err
	}

	return starlark.String(buf), nil
}

func (sr starlarkReader) readUntil(thread *starlark.Thread, b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	var delimiter string
	err := starlark.UnpackArgs("reader.read_until", args, kwargs, "delimiter", &delimiter)
	if err != nil {
		return nil, err
	}

	buf, err := sr.br.ReadBytes(delimiter[0])
	if err != nil {
		return nil, err
	}

	return starlark.Bytes(buf), nil
}

func (sr starlarkReader) readStringUntil(thread *starlark.Thread, b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	var delimiter string
	err := starlark.UnpackArgs("reader.read_string_until", args, kwargs, "delimiter", &delimiter)
	if err != nil {
		return nil, err
	}

	buf, err := sr.br.ReadString(delimiter[0])
	if err != nil {
		return nil, err
	}

	return starlark.String(buf), nil
}

func (sr starlarkReader) peek(thread *starlark.Thread, b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	var n int
	err := starlark.UnpackArgs("reader.peek", args, kwargs, "n", &n)
	if err != nil {
		return nil, err
	}

	buf, err := sr.br.Peek(n)
	if err != nil {
		return nil, err
	}

	return starlark.Bytes(buf), nil
}

func (sr starlarkReader) discard(thread *starlark.Thread, b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	var n int
	err := starlark.UnpackArgs("reader.discard", args, kwargs, "n", &n)
	if err != nil {
		return nil, err
	}

	dn, err := sr.br.Discard(n)
	if err != nil {
		return nil, err
	}

	return starlark.MakeInt(dn), nil
}

func (sr starlarkReader) readAll(thread *starlark.Thread, b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	var limit int64 = 102400
	err := starlark.UnpackArgs("reader.read_all", args, kwargs, "limit??", &limit)
	if err != nil {
		return nil, err
	}

	var r io.Reader = sr.br
	if limit > 0 {
		r = io.LimitReader(sr.br, limit)
	}

	buf, err := io.ReadAll(r)
	if err != nil {
		return nil, err
	}

	return starlark.Bytes(buf), nil
}

func (sr starlarkReader) readAllString(thread *starlark.Thread, b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	var limit int64 = 102400
	err := starlark.UnpackArgs("reader.read_all_string", args, kwargs, "limit??", &limit)
	if err != nil {
		return nil, err
	}

	var r io.Reader = sr.br
	if limit > 0 {
		r = io.LimitReader(sr.br, limit)
	}

	buf, err := io.ReadAll(r)
	if err != nil {
		return nil, err
	}

	return starlark.String(buf), nil
}

func (sr starlarkReader) readJSON(thread *starlark.Thread, b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	var v any
	err := json.NewDecoder(sr.br).Decode(&v)
	if err != nil {
		return nil, err
	}
	return convert.Convert(v)
}

func (sr starlarkReader) readMsgpack(thread *starlark.Thread, b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	var v any
	err := msgpack.NewDecoder(sr.br).Decode(&v)
	if err != nil {
		return nil, err
	}
	return convert.Convert(v)
}

func (sr starlarkReader) closeReader(thread *starlark.Thread, b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	if sr.closeFunc != nil {
		err := sr.closeFunc()
		if err != nil {
			return nil, err
		}
	}
	return starlark.None, nil
}

// Read implements the io.ReadCloser interface
func (sr starlarkReader) Read(b []byte) (int, error) {
	return sr.br.Read(b)
}

// Close implements the io.ReadCloser interface
func (sr starlarkReader) Close() error {
	if sr.closeFunc != nil {
		return sr.closeFunc()
	}
	return nil
}

type readerValue struct {
	io.ReadCloser
}

func (rv *readerValue) Unpack(v starlark.Value) error {
	switch val := v.(type) {
	case starlark.String:
		rv.ReadCloser = io.NopCloser(strings.NewReader(string(val)))
	case starlark.Bytes:
		rv.ReadCloser = io.NopCloser(strings.NewReader(string(val)))
	case starlarkReader:
		rv.ReadCloser = val
	}

	if rv.ReadCloser == nil {
		return errors.New("invalid type for reader")
	}

	return nil
}
