job "lure-updater" {
  region      = "global"
  datacenters = ["dc1"]
  type        = "service"

  group "lure-updater" {
    network {
      port "webhook" {
        to = 8080
      }
    }

    volume "lure-updater-data" {
      type      = "host"
      source    = "lure-updater-data"
      read_only = false
    }

    task "lure-updater" {
      driver = "docker"

      volume_mount {
        volume      = "lure-updater-data"
        destination = "/etc/lure-updater"
        read_only   = false
      }

      env {
        GIT_REPO_DIR             = "/etc/lure-updater/repo"
        GIT_REPO_URL             = "https://github.com/Elara6331/lure-repo.git"
        GIT_CREDENTIALS_USERNAME = "lure-repo-bot"
        GIT_CREDENTIALS_PASSWORD = "${GITHUB_PASSWORD}"
        GIT_COMMIT_NAME          = "lure-repo-bot"
        GIT_COMMIT_EMAIL         = "lure@elara.ws"
        WEBHOOK_PASSWORD_HASH    = "${PASSWORD_HASH}"

        // Hack to force Nomad to re-deploy the service
        // instead of ignoring it
        COMMIT_SHA = "${DRONE_COMMIT_SHA}"
      }

      config {
        image   = "alpine:latest"
        command = "/opt/lure-updater/lure-updater"
        args    = ["-DE"]
        ports   = ["webhook"]
        volumes = ["local/lure-updater/:/opt/lure-updater:ro"]
      }

      artifact {
        source      = "https://api.minio.elara.ws/lure-updater/lure-updater-$${attr.cpu.arch}.tar.gz"
        destination = "local/lure-updater"
      }

      service {
        name = "site"
        port = "webhook"

        tags = [
          "traefik.enable=true",
          "traefik.http.routers.lure-updater.rule=Host(`updater.lure.elara.ws`)",
          "traefik.http.routers.lure-updater.tls.certResolver=letsencrypt",
        ]
      }
    }
  }
}

