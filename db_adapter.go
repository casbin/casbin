package casbin

import (
	"database/sql"
	_ "github.com/go-sql-driver/mysql"
)

type dbAdapter struct {
	driverName     string
	dataSourceName string
	db             *sql.DB
}

func newDbAdapter(driverName string, dataSourceName string) *dbAdapter {
	a := dbAdapter{}
	a.driverName = driverName
	a.dataSourceName = dataSourceName
	return &a
}

func (a *dbAdapter) open() {
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

func (a *dbAdapter) close() {
	a.db.Close()
}

func (a *dbAdapter) createTable() {
	_, err := a.db.Exec("CREATE table IF NOT EXISTS policy (ptype VARCHAR(10), v1 VARCHAR(256), v2 VARCHAR(256), v3 VARCHAR(256), v4 VARCHAR(256))")
	if err != nil {
		panic(err)
	}
}

func (a *dbAdapter) dropTable() {
	_, err := a.db.Exec("DROP table policy")
	if err != nil {
		panic(err)
	}
}

func (a *dbAdapter) loadPolicy(model Model) {
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

func (a *dbAdapter) writeTableLine(ptype string, rule []string) {
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

func (a *dbAdapter) savePolicy(model Model) {
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
