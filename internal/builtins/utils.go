package builtins

import (
	"go.elara.ws/vercmp"
	"go.starlark.net/starlark"
	"go.starlark.net/starlarkstruct"
)

var utilsModule = &starlarkstruct.Module{
	Name: "utils",
	Members: starlark.StringDict{
		"ver_cmp": starlark.NewBuiltin("utils.ver_cmp", utilsVerCmp),
	},
}

func utilsVerCmp(thread *starlark.Thread, b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	var v1, v2 string
	err := starlark.UnpackArgs("utils.ver_cmp", args, kwargs, "v1", &v1, "v2", &v2)
	if err != nil {
		return nil, err
	}
	return starlark.MakeInt(vercmp.Compare(v1, v2)), nil
}
