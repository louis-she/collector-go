package handler

import (
	"../line"
	"database/sql"
	"fmt"
	"io/ioutil"
	"log"
	"reflect"
	"strconv"
)

import _ "../github.com/go-sql-driver/mysql"

type Sets struct{}

// pring it out
func (s Sets) Println(l string) string {
	fmt.Println(l)
	return l
}

// access log parse and save to db
// this is the first class function
func (s Sets) DfApacheAccesslogExtimePathCodeAverage() func(l string) string {
	counter := make(map[string]int)
	sumextime := make(map[string]int)
	times := 5
	return func(l string) string {
		res := line.PickColumn(reflect.ValueOf(l).String(), " ", 5, 7, 9)
		extime, _ := strconv.Atoi(res[0])
		path := res[1]
		counter[path] += 1
		sumextime[path] += extime
		if counter[path] == times {
			counter[path] = 0
			average := sumextime[path] / times
			sumextime[path] = 0
			db, err := sql.Open("mysql", "golang:golang@tcp(localhost:3306)/golang")
			defer db.Close()
			if err != nil {
				log.Fatal(err)
				return ""
			}
			_, err = db.Exec("insert into timespan(time, path) value (?, ?)", average, path)
			if err != nil {
				log.Fatal(err)
				return ""
			}
		}
		return ""
	}
}

// get extime path code from access log
func (s Sets) DfApacheAccesslogExtimePathCode(l string) string {
	res := line.PickColumn(reflect.ValueOf(l).String(), " ", 5, 7, 9)
	extime := res[0]
	path := res[1]
	code := res[2]

	db, err := sql.Open("mysql", "golang:golang@tcp(localhost:3306)/golang")
	defer db.Close()
	if err != nil {
		log.Fatal(err)
		return ""
	}
	_, err = db.Exec("insert into accesslog(code, time, path) value (?, ?, ?)", code, extime, path)
	if err != nil {
		log.Fatal(err)
		return ""
	}
	return ""
}
