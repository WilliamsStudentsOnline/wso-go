// Atlas project config
// See `make atlas-schema`
env "local" {
  src = "file://schema.sql"
  // Throwaway MySQL for replay/planning (Atlas spins this up)
  dev = "docker://mysql/8/dev"

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
