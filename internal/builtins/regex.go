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
	"sync"

	"go.elara.ws/pcre"
	"go.starlark.net/starlark"
	"go.starlark.net/starlarkstruct"
)

var (
	cacheMtx   = &sync.Mutex{}
	regexCache = map[string]*pcre.Regexp{}
)

var regexModule = &starlarkstruct.Module{
	Name: "regex",
	Members: starlark.StringDict{
		"compile":      starlark.NewBuiltin("regex.compile", regexCompile),
		"compile_glob": starlark.NewBuiltin("regex.compile_glob", regexCompileGlob),
	},
}

func regexCompile(thread *starlark.Thread, b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	var regexStr string
	err := starlark.UnpackArgs("regex.compile", args, kwargs, "regex", &regexStr)
	if err != nil {
		return nil, err
	}

	cacheMtx.Lock()
	regex, ok := regexCache[regexStr]
	if !ok {
		regex, err = pcre.Compile(regexStr)
		if err != nil {
			return nil, err
		}
		regexCache[regexStr] = regex
	}
	cacheMtx.Unlock()

	return starlarkRegex(regex), nil
}

func regexCompileGlob(thread *starlark.Thread, b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	var globStr string
	err := starlark.UnpackArgs("regex.compile_glob", args, kwargs, "glob", &globStr)
	if err != nil {
		return nil, err
	}

	cacheMtx.Lock()
	regex, ok := regexCache[globStr]
	if !ok {
		regex, err = pcre.CompileGlob(globStr)
		if err != nil {
			return nil, err
		}
		regexCache[globStr] = regex
	}
	cacheMtx.Unlock()

	return starlarkRegex(regex), nil
}

func starlarkRegex(regex *pcre.Regexp) *starlarkstruct.Struct {
	return starlarkstruct.FromStringDict(starlarkstruct.Default, starlark.StringDict{
		"find_all": findAll(regex),
		"find_one": findOne(regex),
		"matches":  matches(regex),
	})
}

func findAll(regex *pcre.Regexp) *starlark.Builtin {
	return starlark.NewBuiltin("regex.regexp.find_all", func(thread *starlark.Thread, fn *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
		var in string
		err := starlark.UnpackArgs("regex.compile", args, kwargs, "in", &in)
		if err != nil {
			return nil, err
		}

		matches := regex.FindAllStringSubmatch(in, -1)
		return matchesToStarlark2D(matches), nil
	})
}

func findOne(regex *pcre.Regexp) *starlark.Builtin {
	return starlark.NewBuiltin("regex.regexp.find_one", func(thread *starlark.Thread, fn *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
		var in string
		err := starlark.UnpackArgs("regex.compile", args, kwargs, "in", &in)
		if err != nil {
			return nil, err
		}

		match := regex.FindStringSubmatch(in)
		return matchesToStarlark1D(match), nil
	})
}

func matches(regex *pcre.Regexp) *starlark.Builtin {
	return starlark.NewBuiltin("regex.regexp.find_one", func(thread *starlark.Thread, fn *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
		var in string
		err := starlark.UnpackArgs("regex.compile", args, kwargs, "in", &in)
		if err != nil {
			return nil, err
		}

		found := regex.MatchString(in)
		return starlark.Bool(found), nil
	})
}

func matchesToStarlark2D(matches [][]string) *starlark.List {
	outer := make([]starlark.Value, len(matches))
	for i, match := range matches {
		outer[i] = matchesToStarlark1D(match)
	}
	return starlark.NewList(outer)
}

func matchesToStarlark1D(match []string) *starlark.List {
	list := make([]starlark.Value, len(match))
	for j, val := range match {
		list[j] = starlark.String(val)
	}
	return starlark.NewList(list)
}
