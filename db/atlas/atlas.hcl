// Atlas project config
// See `make atlas-schema`
env "local" {
  src = "file://schema.sql"
  dev = "docker+maria://_/mariadb:10.11/dev"

  migration {
    dir    = "file://migrations"
    format = golang-migrate
  }

  format {
    migrate {
      diff = "{{ sql . \"  \" }}"
    }
  }
}
