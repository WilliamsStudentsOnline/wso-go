// Atlas project config
// See `make atlas-schema`
env "local" {
  src = "file://schema.sql"
  // Throwaway MariaDB for replay/planning
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
