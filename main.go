package main

import (
	"fmt"

	"github.com/jmoiron/sqlx"
)

func main() {
	fmt.Println("vim-go")
	conStr := "user=brianspates dbname=postgres sslmode=disable"
	db := sqlx.MustConnect("postgres", conStr)
	m := make(map[string]Action)
	m["woot"] = SomeAction{}
	kway := Kway{1000, 0, true, nil, m}
	fin := kway.poll(db, 1)
	<-fin
}
