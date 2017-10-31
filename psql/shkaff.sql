CREATE SCHEMA IF NOT EXISTS shkaff;

CREATE TABLE IF NOT EXISTS shkaff.users (
  user_id SMALLINT NOT NULL,
  login VARCHAR(16) NULL,
  password VARCHAR(32) NULL,
  api_token VARCHAR(32) NULL,
  first_name VARCHAR(32) NULL,
  last_name VARCHAR(32) NULL,
  is_active SMALLINT NULL,
  is_admin SMALLINT NULL,
  CONSTRAINT users_id_UNIQUE UNIQUE  (user_id),
  PRIMARY KEY (user_id),
  CONSTRAINT login_UNIQUE UNIQUE  (login),
  CONSTRAINT api_token_UNIQUE UNIQUE  (api_token));

CREATE SEQUENCE shkaff.types_seq;

CREATE TABLE IF NOT EXISTS shkaff.types (
  type_id SMALLINT NOT NULL DEFAULT NEXTVAL ('shkaff.types_seq'),
  type VARCHAR(16) NULL,
  cmd_cli VARCHAR(16) NULL,
  cmd_dump VARCHAR(16) NULL,
  cmd_restore VARCHAR(16) NULL,
  PRIMARY KEY (type_id),
  CONSTRAINT type_id_UNIQUE UNIQUE  (type_id));

CREATE SEQUENCE shkaff.db_settings_seq;

CREATE TABLE IF NOT EXISTS shkaff.db_settings (
  db_id int8 NOT NULL DEFAULT NEXTVAL ('shkaff.db_settings_seq'),
  custom_name VARCHAR(32) NULL,
  server_name VARCHAR(32) NULL,
  host VARCHAR(40) NULL,
  port SMALLINT NULL,
  user_id SMALLINT NULL,
  is_active BOOLEAN false,
  type SMALLINT NOT NULL,
  PRIMARY KEY (db_id, type),
  CONSTRAINT db_id_UNIQUE UNIQUE  (db_id),
  CONSTRAINT db_name_UNIQUE UNIQUE  (custom_name)
 ,
  CONSTRAINT fk_db_settings_types1
    FOREIGN KEY (type)
    REFERENCES shkaff.types (type_id)
    ON DELETE NO ACTION
    ON UPDATE NO ACTION);
 
 CREATE INDEX fk_db_settings_types1_idx ON shkaff.db_settings (type);

CREATE SEQUENCE shkaff.tasks_seq;

CREATE TABLE IF NOT EXISTS shkaff.tasks (
  task_id SMALLINT NOT NULL DEFAULT NEXTVAL ('shkaff.tasks_seq'),
  task_name VARCHAR(32) NULL,
  verb SMALLINT NULL,
  start_time TIMESTAMP(0) NULL,
  is_active SMALLINT NULL,
  thread_count SMALLINT NULL,
  ipv6 BOOLEAN false,
  database VARCHAR(40) NULL,
  sheet VARCHAR(40) NULL,
  gzip BOOLEAN false,
  db_settings_id int8 NOT NULL,
  db_settings_type SMALLINT NOT NULL,
  PRIMARY KEY (task_id),
  CONSTRAINT task_id_UNIQUE UNIQUE  (task_id)
 ,
  CONSTRAINT fk_tasks_db_settings1
    FOREIGN KEY (db_settings_id , db_settings_type)
    REFERENCES shkaff.db_settings (db_id , type)
    ON DELETE NO ACTION
    ON UPDATE NO ACTION);
 
 CREATE INDEX fk_tasks_db_settings1_idx ON shkaff.tasks (db_settings_id, db_settings_type);

CREATE TABLE IF NOT EXISTS shkaff.users_has_db_settings (
  users_user_id SMALLINT NOT NULL,
  db_settings_db_id int8 NOT NULL,
  PRIMARY KEY (users_user_id, db_settings_db_id)
 ,
  CONSTRAINT fk_users_has_db_settings_users
    FOREIGN KEY (users_user_id)
    REFERENCES shkaff.users (user_id)
    ON DELETE NO ACTION
    ON UPDATE NO ACTION,
  CONSTRAINT fk_users_has_db_settings_db_settings1
    FOREIGN KEY (db_settings_db_id)
    REFERENCES shkaff.db_settings (db_id)
    ON DELETE NO ACTION
    ON UPDATE NO ACTION);
 
 CREATE INDEX fk_users_has_db_settings_db_settings1_idx ON shkaff.users_has_db_settings (db_settings_db_id);
 CREATE INDEX fk_users_has_db_settings_users_idx ON shkaff.users_has_db_settings (users_user_id);
