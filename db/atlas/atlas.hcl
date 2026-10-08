// Atlas project config. Desired-state schema.sql is regenerated from GORM models
// via `make atlas-schema` (AutoMigrate into a throwaway DB, then inspect).
env "local" {
  src = "file://schema.sql"
  // Ephemeral MySQL used by Atlas to replay/plan migrations.
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
