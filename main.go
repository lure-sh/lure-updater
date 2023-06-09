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

package main

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/go-git/go-git/v5"
	"github.com/pelletier/go-toml/v2"
	"github.com/spf13/pflag"
	"go.elara.ws/logger"
	"go.elara.ws/logger/log"
	"go.elara.ws/lure-updater/internal/builtins"
	"go.elara.ws/lure-updater/internal/config"
	"go.etcd.io/bbolt"
	"go.starlark.net/starlark"
	"golang.org/x/crypto/bcrypt"
	"golang.org/x/term"
)

func init() {
	log.Logger = logger.NewPretty(os.Stderr)
	log.Logger.SetLevel(logger.LogLevelDebug)
}

func main() {
	configPath := pflag.StringP("config", "c", "/etc/lure-updater/config.toml", "Path to config file")
	dbPath := pflag.StringP("database", "d", "/etc/lure-updater/db", "Path to database file")
	pluginDir := pflag.StringP("plugin-dir", "p", "/etc/lure-updater/plugins", "Path to plugin directory")
	serverAddr := pflag.StringP("address", "a", ":8080", "Webhook server address")
	genHash := pflag.BoolP("gen-hash", "g", false, "Generate a password hash for webhooks")
	pflag.Parse()

	if *genHash {
		fmt.Print("Password: ")
		pwd, err := term.ReadPassword(int(os.Stdin.Fd()))
		if err != nil {
			log.Fatal("Error reading password").Err(err).Send()
		}
		hash, err := bcrypt.GenerateFromPassword(pwd, bcrypt.DefaultCost)
		if err != nil {
			log.Fatal("Error hashing password").Err(err).Send()
		}
		fmt.Printf("\n%s\n", hash)
		return
	}

	db, err := bbolt.Open(*dbPath, 0o644, nil)
	if err != nil {
		log.Fatal("Error opening database").Err(err).Send()
	}

	cfg := &config.Config{}
	fl, err := os.Open(*configPath)
	if err != nil {
		log.Fatal("Error opening config file").Err(err).Send()
	}
	err = toml.NewDecoder(fl).Decode(cfg)
	if err != nil {
		log.Fatal("Error decoding config file").Err(err).Send()
	}

	if _, err := os.Stat(cfg.Git.RepoDir); os.IsNotExist(err) {
		err = os.MkdirAll(cfg.Git.RepoDir, 0o755)
		if err != nil {
			log.Fatal("Error creating repository directory").Err(err).Send()
		}

		_, err := git.PlainClone(cfg.Git.RepoDir, false, &git.CloneOptions{
			URL:      cfg.Git.RepoURL,
			Progress: os.Stderr,
		})
		if err != nil {
			log.Fatal("Error cloning repository").Err(err).Send()
		}
	} else if err != nil {
		log.Fatal("Cannot stat configured repo directory").Err(err).Send()
	}

	starFiles, err := filepath.Glob(filepath.Join(*pluginDir, "*.star"))
	if err != nil {
		log.Fatal("Error finding plugin files").Err(err).Send()
	}

	if len(starFiles) == 0 {
		log.Fatal("No plugins found. At least one plugin is required.").Send()
	}

	mux := http.NewServeMux()

	for _, starFile := range starFiles {
		pluginName := filepath.Base(strings.TrimSuffix(starFile, ".star"))
		thread := &starlark.Thread{Name: pluginName}

		predeclared := starlark.StringDict{}
		builtins.Register(predeclared, &builtins.Options{
			Name:   pluginName,
			Config: cfg,
			DB:     db,
			Mux:    mux,
		})

		_, err = starlark.ExecFile(thread, starFile, nil, predeclared)
		if err != nil {
			log.Fatal("Error executing starlark file").Str("file", starFile).Err(err).Send()
		}

		log.Info("Initialized plugin").Str("name", pluginName).Send()
	}

	log.Info("Starting HTTP server").Str("addr", *serverAddr).Send()
	http.ListenAndServe(*serverAddr, mux)
}
