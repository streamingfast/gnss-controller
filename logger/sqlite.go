package logger

import (
	"database/sql"
	"fmt"

	//_ "github.com/mattn/go-sqlite3"
	_ "modernc.org/sqlite"
)

const create string = `
  CREATE TABLE IF NOT EXISTS gnss (
  	id INTEGER NOT NULL PRIMARY KEY,
  	time DATETIME NOT NULL,
  	system_time DATETIME NOT NULL,
	fix TEXT NOT NULL,
	Eph INTEGER NOT NULL,
	Sep INTEGER NOT NULL,
	latitude REAL NOT NULL,
	longitude REAL NOT NULL,
	altitude REAL NOT NULL,
	heading REAL NOT NULL,
	speed REAL NOT NULL,
	gdop REAL NOT NULL,
	hdop REAL NOT NULL,
	pdop REAL NOT NULL,
	tdop REAL NOT NULL,
	vdop REAL NOT NULL,
	xdop REAL NOT NULL,
	ydop REAL NOT NULL,
	seen INTEGER NOT NULL,
	used INTEGER NOT NULL
  );`

const insert string = `
INSERT INTO gnss VALUES(NULL,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?);
`

type Sqlite struct {
	db   *sql.DB
	file string
}

func NewSqlite(file string) *Sqlite {
	return &Sqlite{
		file: file,
	}
}

func (s *Sqlite) Init() error {
	db, err := sql.Open("sqlite", s.file)
	if err != nil {
		return fmt.Errorf("opening database: %s", err.Error())
	}

	if _, err := db.Exec(create); err != nil {
		return fmt.Errorf("creating table: %s", err.Error())
	}
	s.db = db
	return nil
}

func (s *Sqlite) Log(data *Data) error {
	if s.db == nil {
		return fmt.Errorf("database not initialized")
	}

	_, err := s.db.Exec(
		insert,
		data.Timestamp,
		data.SystemTime,
		data.Fix,
		data.Eph,
		data.Sep,
		data.Latitude,
		data.Longitude,
		data.Altitude,
		data.Heading,
		data.Speed,
		data.Dop.GDop,
		data.Dop.HDop,
		data.Dop.PDop,
		data.Dop.TDop,
		data.Dop.VDop,
		data.Dop.XDop,
		data.Dop.YDop,
		data.Satellites.Seen,
		data.Satellites.Used,
	)
	if err != nil {
		return err
	}

	return nil
}