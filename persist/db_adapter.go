package persist

import (
	"database/sql"
	_ "github.com/go-sql-driver/mysql"
	"github.com/hsluoyz/casbin"
)

// The database adapter for policy persistence, can load policy from database or save policy to database.
// For now, only MySQL is tested, but it should work for other RDBMS.
type DBAdapter struct {
	driverName     string
	dataSourceName string
	db             *sql.DB
}

// The constructor for DBAdapter.
func NewDBAdapter(driverName string, dataSourceName string) *DBAdapter {
	a := DBAdapter{}
	a.driverName = driverName
	a.dataSourceName = dataSourceName
	return &a
}

func (a *DBAdapter) open() {
	db, err := sql.Open(a.driverName, a.dataSourceName)
	if err != nil {
		panic(err)
	}

	_, err = db.Exec("CREATE DATABASE IF NOT EXISTS casbin")
	if err != nil {
		panic(err)
	}

	db, err = sql.Open("mysql", a.dataSourceName+"casbin")
	if err != nil {
		panic(err)
	}

	a.db = db

	a.createTable()
}

func (a *DBAdapter) close() {
	a.db.Close()
}

func (a *DBAdapter) createTable() {
	_, err := a.db.Exec("CREATE table IF NOT EXISTS policy (ptype VARCHAR(10), v1 VARCHAR(256), v2 VARCHAR(256), v3 VARCHAR(256), v4 VARCHAR(256))")
	if err != nil {
		panic(err)
	}
}

func (a *DBAdapter) dropTable() {
	_, err := a.db.Exec("DROP table policy")
	if err != nil {
		panic(err)
	}
}

// Load policy from database.
func (a *DBAdapter) LoadPolicy(model casbin.Model) {
	a.open()
	defer a.close()

	var (
		ptype string
		v1    string
		v2    string
		v3    string
		v4    string
	)

	rows, err := a.db.Query("select * from policy")
	if err != nil {
		panic(err)
	}
	defer rows.Close()
	for rows.Next() {
		err := rows.Scan(&ptype, &v1, &v2, &v3, &v4)
		if err != nil {
			panic(err)
		}

		line := ptype
		if v1 != "" {
			line += ", " + v1
		}
		if v2 != "" {
			line += ", " + v2
		}
		if v3 != "" {
			line += ", " + v3
		}
		if v4 != "" {
			line += ", " + v4
		}

		loadPolicyLine(line, model)
		// log.Println(ptype, v1, v2, v3, v4)
	}
	err = rows.Err()
	if err != nil {
		panic(err)
	}
}

func (a *DBAdapter) writeTableLine(ptype string, rule []string) {
	line := "'" + ptype + "'"
	for i := range rule {
		line += ",'" + rule[i] + "'"
	}
	for i := 0; i < 4-len(rule); i++ {
		line += ",''"
	}

	_, err := a.db.Exec("insert into policy values(" + line + ")")
	if err != nil {
		panic(err)
	}
}

// Save policy to database.
func (a *DBAdapter) SavePolicy(model casbin.Model) {
	a.open()
	defer a.close()

	a.dropTable()
	a.createTable()

	for ptype, ast := range model["p"] {
		for _, rule := range ast.Policy {
			a.writeTableLine(ptype, rule)
		}
	}

	for ptype, ast := range model["g"] {
		for _, rule := range ast.Policy {
			a.writeTableLine(ptype, rule)
		}
	}
}
