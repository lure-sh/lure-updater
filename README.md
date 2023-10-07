# LURE Updater

Modular bot that automatically checks for upstream updates and pushes new packages to [lure-repo](https://github.com/lure-sh/lure-repo).

---

### How it works

Since LURE is meant to be able to install many different types of packages, this bot accepts [plugins](https://gitea.elara.ws/lure/updater-plugins) in the form of [Starlark](https://github.com/bazelbuild/starlark) files rather than hardcoding each package. These plugins can schedule functions to be run at certain intervals, or when a webhook is received, and they have access to persistent key/value storage to keep track of information. This allows plugins to use many different ways to detect upstream updates.

For example, the plugin for `discord-bin` repeatedly polls discord's API every hour for the current latest download link. It puts the link in persistent storage, and if it has changed since last time, it parses the URL to extract the version number, and uses that to update the build script for `discord-bin`.

Another example is the plugin for `lure-bin`, which accepts a webhook from GoReleaser. When it receives the webhook, it parses the JSON body and gets the download URL, which it uses to download the checksum file, and uses the information inside that to update the build script for `lure-bin`.

---

### Configuration

There's an example config file in the `lure-updater.example.toml` file. Edit that to fit your needs and put it at `/etc/lure-updater/config.toml`. You can change the location of the config file using the `--config` or `-c` flag.
