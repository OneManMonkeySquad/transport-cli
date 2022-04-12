DROP TABLE IF EXISTS tags;
DROP TABLE IF EXISTS entries;

CREATE TABLE tags (name varchar(36) NOT NULL PRIMARY KEY, id varchar(36) NOT NULL);
CREATE TABLE entries (id varchar(36) NOT NULL PRIMARY KEY, base_id varchar(36) NOT NULL);