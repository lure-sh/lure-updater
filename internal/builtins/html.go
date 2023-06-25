package builtins

import (
	"io"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"go.starlark.net/starlark"
	"go.starlark.net/starlarkstruct"
)

var (
	_ starlark.Iterable  = (*starlarkSelection)(nil)
	_ starlark.Sliceable = (*starlarkSelection)(nil)
	_ starlark.Sequence  = (*starlarkSelection)(nil)
	_ starlark.Value     = (*starlarkSelection)(nil)
)

var htmlModule = &starlarkstruct.Module{
	Name: "html",
	Members: starlark.StringDict{
		"parse": starlark.NewBuiltin("html.parse", htmlParse),
	},
}

func htmlParse(thread *starlark.Thread, fn *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	var val starlark.Value
	err := starlark.UnpackArgs("html.selection.find", args, kwargs, "from", &val)
	if err != nil {
		return nil, err
	}

	var r io.ReadCloser
	switch val := val.(type) {
	case starlark.String:
		r = io.NopCloser(strings.NewReader(string(val)))
	case starlark.Bytes:
		r = io.NopCloser(strings.NewReader(string(val)))
	case starlarkReader:
		r = val
	}
	defer r.Close()

	doc, err := goquery.NewDocumentFromReader(r)
	if err != nil {
		return nil, err
	}

	return newStarlarkSelection(doc.Selection), nil
}

type starlarkSelection struct {
	sel *goquery.Selection
	*starlarkstruct.Struct
}

func newStarlarkSelection(sel *goquery.Selection) starlarkSelection {
	ss := starlarkSelection{sel: sel}
	ss.Struct = starlarkstruct.FromStringDict(starlark.String("html.selection"), starlark.StringDict{
		"text":           starlark.NewBuiltin("html.selection.text", ss.text),
		"html":           starlark.NewBuiltin("html.selection.html", ss.html),
		"children":       starlark.NewBuiltin("html.selection.children", ss.children),
		"parent":         starlark.NewBuiltin("html.selection.parent", ss.parent),
		"find":           starlark.NewBuiltin("html.selection.find", ss.find),
		"add":            starlark.NewBuiltin("html.selection.add", ss.add),
		"attr":           starlark.NewBuiltin("html.selection.attr", ss.attr),
		"has_class":      starlark.NewBuiltin("html.selection.has_class", ss.hasClass),
		"index_selector": starlark.NewBuiltin("html.selection.index_selector", ss.indexSelector),
		"and_self":       starlark.NewBuiltin("html.selection.and_self", ss.andSelf),
		"first":          starlark.NewBuiltin("html.selection.first", ss.first),
		"last":           starlark.NewBuiltin("html.selection.last", ss.last),
		"next":           starlark.NewBuiltin("html.selection.last", ss.next),
		"next_all":       starlark.NewBuiltin("html.selection.next_all", ss.nextAll),
		"next_until":     starlark.NewBuiltin("html.selection.next_until", ss.nextUntil),
		"prev":           starlark.NewBuiltin("html.selection.prev", ss.prev),
		"prev_all":       starlark.NewBuiltin("html.selection.prev_all", ss.prevAll),
		"prev_until":     starlark.NewBuiltin("html.selection.prev_until", ss.prevUntil),
	})
	return ss
}

func (ss starlarkSelection) Truth() starlark.Bool {
	return len(ss.sel.Nodes) > 0
}

func (ss starlarkSelection) Len() int {
	return ss.sel.Length()
}

func (ss starlarkSelection) Index(i int) starlark.Value {
	return newStarlarkSelection(ss.sel.Slice(i, i+1))
}

func (ss starlarkSelection) Slice(start, end, _ int) starlark.Value {
	return newStarlarkSelection(ss.sel.Slice(start, end))
}

func (ss starlarkSelection) Iterate() starlark.Iterator {
	return newSelectionIterator(ss.sel)
}

func (ss starlarkSelection) text(thread *starlark.Thread, fn *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	return starlark.String(ss.sel.Text()), nil
}

func (ss starlarkSelection) html(thread *starlark.Thread, fn *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	s, err := ss.sel.Html()
	return starlark.String(s), err
}

func (ss starlarkSelection) children(thread *starlark.Thread, fn *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	return newStarlarkSelection(ss.sel.Children()), nil
}

func (ss starlarkSelection) parent(thread *starlark.Thread, fn *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	return newStarlarkSelection(ss.sel.Parent()), nil
}

func (ss starlarkSelection) find(thread *starlark.Thread, fn *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	var selector string
	err := starlark.UnpackArgs("html.selection.find", args, kwargs, "selector", &selector)
	if err != nil {
		return nil, err
	}

	return newStarlarkSelection(ss.sel.Find(selector)), nil
}

func (ss starlarkSelection) add(thread *starlark.Thread, fn *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	var selector string
	err := starlark.UnpackArgs("html.selection.add", args, kwargs, "selector", &selector)
	if err != nil {
		return nil, err
	}
	return newStarlarkSelection(ss.sel.Add(selector)), nil
}

func (ss starlarkSelection) indexSelector(thread *starlark.Thread, fn *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	var selector string
	err := starlark.UnpackArgs("html.selection.index_selector", args, kwargs, "selector", &selector)
	if err != nil {
		return nil, err
	}
	return starlark.MakeInt(ss.sel.IndexSelector(selector)), nil
}

func (ss starlarkSelection) attr(thread *starlark.Thread, fn *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	var name, def string
	err := starlark.UnpackArgs("html.selection.find", args, kwargs, "name", &name, "default??", &def)
	if err != nil {
		return nil, err
	}
	return starlark.String(ss.sel.AttrOr(name, def)), nil
}

func (ss starlarkSelection) hasClass(thread *starlark.Thread, fn *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	var name string
	err := starlark.UnpackArgs("html.selection.has_class", args, kwargs, "name", &name)
	if err != nil {
		return nil, err
	}
	return starlark.Bool(ss.sel.HasClass(name)), nil
}

func (ss starlarkSelection) andSelf(thread *starlark.Thread, fn *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	return newStarlarkSelection(ss.sel.AndSelf()), nil
}

func (ss starlarkSelection) first(thread *starlark.Thread, fn *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	return newStarlarkSelection(ss.sel.First()), nil
}

func (ss starlarkSelection) last(thread *starlark.Thread, fn *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	return newStarlarkSelection(ss.sel.Last()), nil
}

func (ss starlarkSelection) next(thread *starlark.Thread, fn *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	return newStarlarkSelection(ss.sel.Next()), nil
}

func (ss starlarkSelection) nextAll(thread *starlark.Thread, fn *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	return newStarlarkSelection(ss.sel.NextAll()), nil
}

func (ss starlarkSelection) nextUntil(thread *starlark.Thread, fn *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	var selector string
	err := starlark.UnpackArgs("html.selection.next_until", args, kwargs, "selector", &selector)
	if err != nil {
		return nil, err
	}
	return newStarlarkSelection(ss.sel.NextUntil(selector)), nil
}

func (ss starlarkSelection) prev(thread *starlark.Thread, fn *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	return newStarlarkSelection(ss.sel.Prev()), nil
}

func (ss starlarkSelection) prevAll(thread *starlark.Thread, fn *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	return newStarlarkSelection(ss.sel.PrevAll()), nil
}

func (ss starlarkSelection) prevUntil(thread *starlark.Thread, fn *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	var selector string
	err := starlark.UnpackArgs("html.selection.prev_until", args, kwargs, "selector", &selector)
	if err != nil {
		return nil, err
	}
	return newStarlarkSelection(ss.sel.PrevUntil(selector)), nil
}

type starlarkSelectionIterator struct {
	sel   *goquery.Selection
	index int
}

func newSelectionIterator(sel *goquery.Selection) *starlarkSelectionIterator {
	return &starlarkSelectionIterator{sel: sel}
}

func (ssi *starlarkSelectionIterator) Next(v *starlark.Value) bool {
	if ssi.index == ssi.sel.Length() {
		return false
	}
	*v = newStarlarkSelection(ssi.sel.Slice(ssi.index, ssi.index+1))
	ssi.index++
	return true
}

func (ssi *starlarkSelectionIterator) Done() {}
