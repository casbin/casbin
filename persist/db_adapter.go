// Copyright 2017 The casbin Authors. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package persist

import (
	"database/sql"
	"github.com/casbin/casbin/model"
	_ "github.com/go-sql-driver/mysql" // This is for MySQL initialization.
)

// DBAdapter represents the database adapter for policy persistence, can load policy from database or save policy to database.
// For now, only MySQL is tested, but it should work for other RDBMS.
type DBAdapter struct {
	driverName     string
	dataSourceName string
	db             *sql.DB
}

// NewDBAdapter is the constructor for DBAdapter.
func NewDBAdapter(driverName string, dataSourceName string) *DBAdapter {
	a := DBAdapter{}
	a.driverName = driverName
	a.dataSourceName = dataSourceName
	return &a
}

func (a *DBAdapter) open() error {
	db, err := sql.Open(a.driverName, a.dataSourceName)
	if err != nil {
		return err
	}

	_, err = db.Exec("CREATE DATABASE IF NOT EXISTS casbin")
	if err != nil {
		return err
	}

	db, err = sql.Open("mysql", a.dataSourceName+"casbin")
	if err != nil {
		return err
	}

	a.db = db

	return a.createTable()
}

func (a *DBAdapter) close() {
	a.db.Close()
}

func (a *DBAdapter) createTable() error {
	_, err := a.db.Exec("CREATE table IF NOT EXISTS policy (ptype VARCHAR(10), v1 VARCHAR(256), v2 VARCHAR(256), v3 VARCHAR(256), v4 VARCHAR(256))")
	if err != nil {
		return err
	}

	return nil
}

func (a *DBAdapter) dropTable() error {
	_, err := a.db.Exec("DROP table policy")
	if err != nil {
		return err
	}

	return nil
}

// LoadPolicy loads policy from database.
func (a *DBAdapter) LoadPolicy(model model.Model) error {
	err := a.open()
	if err != nil {
		return err
	}

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
		return err
	}
	defer rows.Close()
	for rows.Next() {
		err := rows.Scan(&ptype, &v1, &v2, &v3, &v4)
		if err != nil {
			return err
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
		return err
	}

	return nil
}

func (a *DBAdapter) writeTableLine(ptype string, rule []string) error {
	line := "'" + ptype + "'"
	for i := range rule {
		line += ",'" + rule[i] + "'"
	}
	for i := 0; i < 4-len(rule); i++ {
		line += ",''"
	}

	_, err := a.db.Exec("insert into policy values(" + line + ")")
	if err != nil {
		return err
	}

	return nil
}

// SavePolicy saves policy to database.
func (a *DBAdapter) SavePolicy(model model.Model) error {
	err := a.open()
	if err != nil {
		return err
	}

	defer a.close()

	err = a.dropTable()
	if err != nil {
		return err
	}
	err = a.createTable()
	if err != nil {
		return err
	}

	for ptype, ast := range model["p"] {
		for _, rule := range ast.Policy {
			err = a.writeTableLine(ptype, rule)
			if err != nil {
				return err
			}
		}
	}

	for ptype, ast := range model["g"] {
		for _, rule := range ast.Policy {
			err = a.writeTableLine(ptype, rule)
			if err != nil {
				return err
			}
		}
	}

	return nil
}
