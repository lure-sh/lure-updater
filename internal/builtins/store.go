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
	"go.elara.ws/logger/log"
	"go.etcd.io/bbolt"
	"go.starlark.net/starlark"
	"go.starlark.net/starlarkstruct"
)

func storeModule(db *bbolt.DB, bucketName string) *starlarkstruct.Module {
	return &starlarkstruct.Module{
		Name: "store",
		Members: starlark.StringDict{
			"set":    storeSet(db, bucketName),
			"get":    storeGet(db, bucketName),
			"delete": storeDelete(db, bucketName),
		},
	}
}

func storeSet(db *bbolt.DB, bucketName string) *starlark.Builtin {
	return starlark.NewBuiltin("store.set", func(thread *starlark.Thread, b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
		var key, value string
		err := starlark.UnpackArgs("store.set", args, kwargs, "key", &key, "value", &value)
		if err != nil {
			return nil, err
		}

		err = db.Update(func(tx *bbolt.Tx) error {
			bucket, err := tx.CreateBucketIfNotExists([]byte(bucketName))
			if err != nil {
				return err
			}
			err = bucket.Put([]byte(key), []byte(value))
			if err != nil {
				return err
			}
			return nil
		})
		if err != nil {
			return nil, err
		}

		log.Debug("Set value").Str("bucket", bucketName).Str("key", key).Str("value", value).Stringer("pos", thread.CallFrame(1).Pos).Send()
		return starlark.None, nil
	})
}

func storeGet(db *bbolt.DB, bucketName string) *starlark.Builtin {
	return starlark.NewBuiltin("store.get", func(thread *starlark.Thread, b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
		var key string
		err := starlark.UnpackArgs("store.get", args, kwargs, "key", &key)
		if err != nil {
			return nil, err
		}

		var value string
		err = db.Update(func(tx *bbolt.Tx) error {
			bucket, err := tx.CreateBucketIfNotExists([]byte(bucketName))
			if err != nil {
				return err
			}
			data := bucket.Get([]byte(key))
			value = string(data)
			return nil
		})
		if err != nil {
			return nil, err
		}

		log.Debug("Retrieved value").Str("bucket", bucketName).Str("key", key).Str("value", value).Stringer("pos", thread.CallFrame(1).Pos).Send()
		return starlark.String(value), nil
	})
}

func storeDelete(db *bbolt.DB, bucketName string) *starlark.Builtin {
	return starlark.NewBuiltin("store.delete", func(thread *starlark.Thread, b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
		var key string
		err := starlark.UnpackArgs("store.delete", args, kwargs, "key", &key)
		if err != nil {
			return nil, err
		}

		err = db.Update(func(tx *bbolt.Tx) error {
			bucket, err := tx.CreateBucketIfNotExists([]byte(bucketName))
			if err != nil {
				return err
			}
			return bucket.Delete([]byte(key))
		})
		if err != nil {
			return nil, err
		}

		log.Debug("Deleted value").Str("bucket", bucketName).Str("key", key).Stringer("pos", thread.CallFrame(1).Pos).Send()
		return starlark.None, nil
	})
}
