package handler

import (
	"../line"
	"database/sql"
	"fmt"
	"log"
	"reflect"
	"strconv"
	"strings"
	"time"
)

import _ "../github.com/go-sql-driver/mysql"

type Sets struct{}

// print it out
func (s Sets) Println(l string) string {
	fmt.Println(l)
	return l
}

// access log parse and save to db
// this is the first class function
func (s Sets) DfApacheAccesslogExtimePathCodeAverage() func(l string) string {
	counter := make(map[string]int)
	sumextime := make(map[string]int)
	times := 10
	threshold := 100000 // 0.1 seconds

	return func(l string) string {
		res := line.PickColumn(reflect.ValueOf(l).String(), " ", 5, 7, 3)
		if len(res) != 3 {
			log.Println("DfApacheAccesslogExtimePathCodeAverage found unrecognized line %s", l)
			return ""
		}
		extime, _ := strconv.Atoi(res[0])
		reqtime := res[2][1:]

		if extime < threshold {
			return ""
		}
		path := res[1]
		pmodule := strings.Split(path, "/")
		if len(pmodule) < 2 {
			return ""
		}
		module := pmodule[1]
		counter[path] += 1
		sumextime[path] += extime
		if counter[path] == times {
			counter[path] = 0
			average := sumextime[path] / times
			sumextime[path] = 0
			db, err := sql.Open("mysql", "root:akaqa123@tcp(10.100.56.32:6001)/monitor")
			defer db.Close()
			if err != nil {
				log.Println(err)
				return ""
			}
			_, err = db.Exec("insert into human_apache_timespan(time, module, path, reqtime, ctime) value (?, ?, ?, ?, FROM_UNIXTIME(?))", average, module, path, reqtime, time.Now().Unix())
			if err != nil {
				log.Println(err)
				return ""
			}
		}
		return ""
	}
}

// python ouput stream not running
func (s Sets) OuputStreamNotRunning() {
	dfAlert("Python Output Process may not running")
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
		log.Println(err)
		return ""
	}
	_, err = db.Exec("insert into accesslog(code, time, path) value (?, ?, ?)", code, extime, path)
	if err != nil {
		log.Println(err)
		return ""
	}
	return ""
}

// add a msg to human_alert table to send
// alert email and sms
func dfAlert(msg string) {
	db, err := sql.Open("mysql", "root:akaqa123@tcp(10.100.56.32:6001)/monitor")
	defer db.Close()
	if err != nil {
		log.Println(err)
	}
	_, err = db.Exec("insert into human_alert(msg, ctime) value (?, FROM_UNIXTIME(?))", msg, time.Now().Unix())
	if err != nil {
		log.Println(err)
	}
	return
}
