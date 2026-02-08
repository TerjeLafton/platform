env "local" {
  src = [
    "apps/todo/db/schema.sql",
    "apps/id/db/schema.sql",
  ]

  url = "postgres://postgres:postgres@localhost:5432/platform?sslmode=disable"

  dev = "postgres://postgres:postgres@localhost:5432/atlas_dev?sslmode=disable"

  migration {
    dir = "file://migrations"
  }

  format {
    migrate {
      diff = "{{ sql . \"  \" }}"
    }
  }
}
