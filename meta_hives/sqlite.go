package meta_hives

import (
	"database/sql"

	_ "github.com/mattn/go-sqlite3"

	"github.com/google/uuid"
)

type SqliteMetaHive struct {
	db *sql.DB
}

func NewSqlite(fileName string) (*SqliteMetaHive, error) {
	db, err := sql.Open("sqlite3", fileName)
	if err != nil {
		return nil, err
	}

	db.Exec("DROP TABLE tags; DROP TABLE entries;")

	_, err = db.Exec("CREATE TABLE tags (name TEXT NOT NULL PRIMARY KEY, id TEXT NOT NULL)")
	if err != nil {
		return nil, err
	}

	_, err = db.Exec("CREATE TABLE entries (id TEXT NOT NULL PRIMARY KEY, base_id TEXT NOT NULL)")
	if err != nil {
		return nil, err
	}

	return &SqliteMetaHive{
		db: db,
	}, nil
}

func (p *SqliteMetaHive) Tags() ([]Tag, error) {
	rows, err := p.db.Query("SELECT name, id FROM tags")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tags []Tag
	for rows.Next() {
		var id uuid.UUID
		var name string
		err = rows.Scan(&id, &name)
		if err != nil {
			return nil, err
		}

		tags = append(tags, Tag{
			Name: name,
			Id:   id,
		})
	}

	return tags, nil
}

func (p *SqliteMetaHive) FindTagByName(name string) (*Tag, error) {
	rows, err := p.db.Query("SELECT id FROM tags WHERE name=?", &name)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	if rows.Next() {
		var id uuid.UUID
		err = rows.Scan(&id)
		if err != nil {
			return nil, err
		}
		return &Tag{
			Name: name,
			Id:   id,
		}, nil
	}

	return nil, nil
}

func (p *SqliteMetaHive) UpdateTag(name string, newId uuid.UUID) error {
	_, err := p.db.Exec("INSERT INTO tags (name, id) VALUES (?,?) ON CONFLICT(name) DO UPDATE SET id=excluded.id", &name, &newId)
	if err != nil {
		return err
	}
	return nil
}

func (p *SqliteMetaHive) FindEntry(id uuid.UUID) (uuid.UUID, error) {
	rows, err := p.db.Query("SELECT base_id FROM entries WHERE id=?", &id)
	if err != nil {
		return uuid.Nil, err
	}
	defer rows.Close()

	if !rows.Next() {
		return uuid.Nil, nil
	}

	var base_id uuid.UUID
	err = rows.Scan(&base_id)
	if err != nil {
		return uuid.Nil, err
	}
	return base_id, nil
}

func (p *SqliteMetaHive) AddEntry(id uuid.UUID, baseId uuid.UUID) error {
	_, err := p.db.Exec("INSERT INTO entries (id, base_id) VALUES (?,?)", &id, &baseId)
	if err != nil {
		return err
	}
	return nil
}

func (p *SqliteMetaHive) Close() {
	p.db.Close()
}
