// schema.hcl — MySQL (InnoDB, utf8mb4)
schema "db_name" {
  name    = "db_name"
  charset = "utf8mb4"
  collate = "utf8mb4_0900_ai_ci"
}

/* -----------------------------
   users / organizations / projects / tasks
   many-to-many: org_members
-------------------------------- */

table "users" {
  schema = schema.db_name

  column "id" {
    type           = bigint
    unsigned       = true
    nullable       = false
    auto_increment = true
  }
  column "email" {
    type     = varchar(255)
    nullable = false
  }
  column "username" {
    type     = varchar(50)
    nullable = false
  }
  column "status" {
    // ENUM 예시
    type     = enum("active", "suspended")
    nullable = false
    default  = "active"
  }
  column "created_at" {
    type     = timestamp
    nullable = false
    default  = sql("CURRENT_TIMESTAMP")
  }
  column "updated_at" {
    type     = timestamp
    nullable = false
    default  = sql("CURRENT_TIMESTAMP")
    // MySQL의 ON UPDATE CURRENT_TIMESTAMP는 Atlas에서 on_update 속성으로 분리 지원
    on_update = sql("CURRENT_TIMESTAMP")
  }

  primary_key {
    columns = [column.id]
  }

  index "ux_users_email" {
    unique  = true
    columns = [column.email]
  }

  index "ux_users_username" {
    unique  = true
    columns = [column.username]
  }
}

table "organizations" {
  schema = schema.db_name

  column "id" {
    type           = bigint
    unsigned       = true
    nullable       = false
    auto_increment = true
  }
  column "name" {
    type     = varchar(120)
    nullable = false
  }
  column "owner_id" {
    type     = bigint
    unsigned = true
    nullable = false
  }
  column "created_at" {
    type     = timestamp
    nullable = false
    default  = sql("CURRENT_TIMESTAMP")
  }

  primary_key {
    columns = [column.id]
  }

  index "ux_orgs_name" {
    unique  = true
    columns = [column.name]
  }

  foreign_key "fk_org_owner" {
    columns     = [column.owner_id]
    ref_columns = [table.users.column.id]
    on_delete   = RESTRICT
    on_update   = CASCADE
  }
}

table "org_members" {
  schema = schema.db_name

  column "org_id" {
    type     = bigint
    unsigned = true
    nullable = false
  }
  column "user_id" {
    type     = bigint
    unsigned = true
    nullable = false
  }
  column "role" {
    type     = enum("owner", "admin", "member", "viewer")
    nullable = false
    default  = "member"
  }
  column "created_at" {
    type     = timestamp
    nullable = false
    default  = sql("CURRENT_TIMESTAMP")
  }

  primary_key {
    columns = [column.org_id, column.user_id]
  }

  index "ux_member_unique" {
    unique  = true
    columns = [column.org_id, column.user_id]
  }

  foreign_key "fk_member_org" {
    columns     = [column.org_id]
    ref_columns = [table.organizations.column.id]
    on_delete   = CASCADE
    on_update   = CASCADE
  }

  foreign_key "fk_member_user" {
    columns     = [column.user_id]
    ref_columns = [table.users.column.id]
    on_delete   = CASCADE
    on_update   = CASCADE
  }
}

table "projects" {
  schema = schema.db_name

  column "id" {
    type           = bigint
    unsigned       = true
    nullable       = false
    auto_increment = true
  }
  column "org_id" {
    type     = bigint
    unsigned = true
    nullable = false
  }
  column "name" {
    type     = varchar(200)
    nullable = false
  }
  column "slug" {
    type     = varchar(80)
    nullable = false
  }
  column "status" {
    type     = enum("active", "archived")
    nullable = false
    default  = "active"
  }
  column "created_at" {
    type     = timestamp
    nullable = false
    default  = sql("CURRENT_TIMESTAMP")
  }
  column "updated_at" {
    type      = timestamp
    nullable  = false
    default   = sql("CURRENT_TIMESTAMP")
    on_update = sql("CURRENT_TIMESTAMP")
  }

  primary_key {
    columns = [column.id]
  }

  index "ux_projects_org_slug" {
    unique  = true
    columns = [column.org_id, column.slug]
  }

  index "ix_projects_org" {
    columns = [column.org_id]
  }

  foreign_key "fk_projects_org" {
    columns     = [column.org_id]
    ref_columns = [table.organizations.column.id]
    on_delete   = CASCADE
    on_update   = CASCADE
  }
}

table "tasks" {
  schema = schema.db_name

  column "id" {
    type           = bigint
    unsigned       = true
    nullable       = false
    auto_increment = true
  }
  column "project_id" {
    type     = bigint
    unsigned = true
    nullable = false
  }
  column "assignee_id" {
    type     = bigint
    unsigned = true
    nullable = false
  }
  column "title" {
    type     = varchar(200)
    nullable = false
  }
  column "description" {
    type     = text
    nullable = true
  }
  column "priority" {
    type     = enum("low", "medium", "high", "urgent")
    nullable = false
    default  = "medium"
  }
  column "status" {
    type     = enum("todo", "doing", "done", "blocked")
    nullable = false
    default  = "todo"
  }
  column "due_date" {
    type     = date
    nullable = true
  }
  column "created_at" {
    type     = timestamp
    nullable = false
    default  = sql("CURRENT_TIMESTAMP")
  }
  column "updated_at" {
    type      = timestamp
    nullable  = false
    default   = sql("CURRENT_TIMESTAMP")
    on_update = sql("CURRENT_TIMESTAMP")
  }

  primary_key {
    columns = [column.id]
  }

  index "ix_tasks_project" {
    columns = [column.project_id]
  }

  index "ix_tasks_status" {
    columns = [column.status]
  }

  foreign_key "fk_tasks_project" {
    columns     = [column.project_id]
    ref_columns = [table.projects.column.id]
    on_delete   = CASCADE
    on_update   = CASCADE
  }
}