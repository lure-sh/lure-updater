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

package config

type Config struct {
	Git     Git     `toml:"git" envPrefix:"GIT_"`
	Webhook Webhook `toml:"webhook" envPrefix:"WEBHOOK_"`
}

type Git struct {
	RepoDir     string      `toml:"repoDir" env:"REPO_DIR"`
	RepoURL     string      `toml:"repoURL" env:"REPO_URL"`
	Commit      Commit      `toml:"commit" envPrefix:"COMMIT_"`
	Credentials Credentials `toml:"credentials" envPrefix:"CREDENTIALS_"`
}

type Credentials struct {
	Username string `toml:"username" env:"USERNAME"`
	Password string `toml:"password" env:"PASSWORD"`
}

type Commit struct {
	Name  string `toml:"name" env:"NAME"`
	Email string `toml:"email" env:"EMAIL"`
}

type Webhook struct {
	PasswordHash string `toml:"pwd_hash" env:"PASSWORD_HASH"`
}
