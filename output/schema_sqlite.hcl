table "user" {
  schema = schema.main
  column "id" {
    null = true
    type = text
  }
  column "name" {
    null = true
    type = text
  }
  column "seq" {
    null = true
    type = integer
  }
  primary_key {
    columns = [column.seq]
  }
}
schema "main" {
}
