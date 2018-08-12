package main

import (
	"database/sql"
	"github.com/0xor1/panic"
	_ "github.com/go-sql-driver/mysql"
	"io/ioutil"
)

func main() {
	db, err := sql.Open("mysql", "root:dev@tcp(localhost:3306)/mysql?parseTime=true&loc=UTC&multiStatements=true")
	panic.IfNotNil(err)

	accountsSchema, err := ioutil.ReadFile("db/central/accounts.sql")
	panic.IfNotNil(err)

	pwdsSchema, err := ioutil.ReadFile("db/central/pwds.sql")
	panic.IfNotNil(err)

	treesSchema, err := ioutil.ReadFile("db/regional/trees.sql")
	panic.IfNotNil(err)

	_, err = db.Exec(string(accountsSchema))
	panic.IfNotNil(err)

	_, err = db.Exec(string(pwdsSchema))
	panic.IfNotNil(err)

	_, err = db.Exec(string(treesSchema))
	panic.IfNotNil(err)
}
