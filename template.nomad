job "lure-updater" {
  region      = "global"
  datacenters = ["dc1"]
  type        = "service"

  group "site" {
    network {
      port "webhook" {
        to = 8080
      }
    }

    task "lure-updater" {
      driver = "docker"

      env {
        GIT_REPO_DIR="/etc/lure-updater/repo"
        GIT_REPO_URL="https://github.com/Elara6331/lure-repo.git"
        GIT_CREDENTIALS_USERNAME="lure-repo-bot"
        GIT_CREDENTIALS_PASSWORD="${GITHUB_PASSWORD}"
        GIT_COMMIT_NAME="lure-repo-bot"
        GIT_COMMIT_EMAIL="lure@elara.ws"
        WEBHOOK_PASSWORD_HASH="${PASSWORD_HASH}"
        
        // Hack to force Nomad to re-deploy the service
        // instead of ignoring it
        COMMIT_SHA = "${DRONE_COMMIT_SHA}"
      }

      config {
        image   = "alpine:latest"
        ports   = ["webhook"]
        volumes = ["local/:/opt/lure-updater:ro"]
        command = "/opt/lure-updater/lure-updater"
        args    = ["-E"]
    }

      artifact {
        source      = "https://api.minio.elara.ws/site/site.tar.gz"
        destination = "local/site"
      }

      service {
        name = "site"
        port = "webhook"

        tags = [
          "traefik.enable=true",
          "traefik.http.routers.site.rule=Host(`updater.lure.elara.ws`)",
          "traefik.http.routers.site.tls.certResolver=letsencrypt",
        ]
      }
    }
  }
}

