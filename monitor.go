//Monitor logs in go
package main

import (
	"./ffmt"
	"./handler"
	"./parser"
	//"bufio"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"reflect"
	//"time"
)

type ConfFile struct {
	Path         []string
	PathParser   []string
	HandlerChain [][]string
	Timespan     int64
}

type Configuration struct {
	Entity []ConfFile
}

type Tail func(string)

type Suit struct {
	path    string
	parser  []reflect.Value
	handler [][]reflect.Value
	tail    Tail
}

// function that will handle a single line
type Handler func(string)

type Parser func(string)

// type Case struct {
// 	tailFunc Tail
// }

//MonitorEntity stands for a group of files which are needed by a same
//rule to monitor
type MonitorEntity struct {
	//monitor path
	path string
	//file name parser
	parser []reflect.Value
	//file lines handler chains
	handler [][]reflect.Value
	//tail function
	tail Tail
	//monitor span in seconds
	span int64
	//last execute time
	lasexec int64
	//is the monitor entity is running
	running bool
}

func main() {
	var logmsg string
	fname := ffmt.Fname{Path: "access.log.%H"}
	fname = fname.Parse(ffmt.TimefmtParser)
	fmt.Println(fname.Path)

	file, _ := os.Open("conf.json")
	decoder := json.NewDecoder(file)
	c := Configuration{}
	err := decoder.Decode(&c)
	if err != nil {
		fmt.Println(err)
		return
	}

	supportHandlers := handler.Sets{}
	supportParsers := parser.Sets{}

	// configure monitor entities
	var monitorEntities []MonitorEntity

	for _, e := range c.Entity {
		var me MonitorEntity
		// a monitor is just for a file
		for _, p := range e.Path {
			// apply file name parsers
			for _, parser := range e.PathParser {
				method := reflect.ValueOf(supportParsers).MethodByName(parser)
				if method.IsValid() == false {
					// parser function is not valid
					logmsg = fmt.Sprintf("parser %s is not valid function", parser)
					log.Fatal(logmsg)
					continue
				}
				append(me.parser, method)
			}

			// apply file line handler chain
			for _, chain := range e.HandlerChain {
				tmpChain := make([]reflect.Value, 5, 5)
				for _, fHandler := range chain {
					method := reflect.ValueOf(supportHandlers).MethodByName(fHandler)
					if method.IsValid() == false {
						// line handler function is not valid
						logmsg = fmt.Sprintf("handler %s is not valid function", fHandler)
						log.Fatal(logmsg)
						continue
					}
					append(tmpChain, method)
				}
				me.tail = append(me.handler, tmpChain)
			}
		}
		me.tail = genTailFunc(me)
		append(monitorEntities, me)
	}

	fmt.Println(monitorEntities)
}

func genTailFunc(me MonitorEntity) Tail {
	lastPos := 0
	return func() {
		// generate file path by parser chain
		me.running = true
		path := me.path
		for _, p := range me.parser {
			res := me.parser.Call([]reflect.Value{reflect.ValueOf(path)})
			path = res[0].String()
		}

		// open the file
		file, err := os.Open(path)
		if err != nil {
			log.Fatal(err)
			return nil
		}

		// get size of the file
		info, err := file.Stat()
		if err != nil {
			log.Fatal(err)
			return
		}
		currentSize := info.Size()

		// current size of the file is smaller than the
		// last position, file may rotate, read from the
		// 0 position.
		if currentSize < lastPos {
			lastPos = 0
		}

		// seek to the last position before read it
		_, err = file.Seek(lastPos, 0)
		if err != nil {
			log.Fatal(err)
			return
		}

		// update the lastPosition by size of the file
		lastPos = currentSize

		// tail the file line by line
		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			line := scanner.Text()
			// loop the chain slice to gen the results
			for _, chain := range me.handler {
				//loop the chain to call every handler in that chain
				//cache the result of every chain to let be reuseable
				result := line
				for _, handler := range chain {
					res := handler.Call([]reflect.Value{reflect.ValueOf(result)})
					result = res[0].String()
				}
			}
		}

		me.lasexec = time.Now().Unix()
		me.running = false
	}
}

// func genWorkSlice(c Configuration) []Case {
// 	var ret []Case
// 	for _, v := range c.Entity {
// 		var h Handler
// 		if v.Function == "Println" {
// 			h = func(line string) {
// 				fmt.Println(line)
// 			}
// 		} else {
// 			h = func(line string) {
// 				nt := time.Now()
// 				currentTime := fmt.Sprintf("%d:%d", nt.Hour(), nt.Minute())
// 				fmt.Println("[", currentTime, "]", " default handler: ", line)
// 			}
// 		}
// 		tail := tailTheFile(v.Path, 0, h)
//
// 		ca := Case{tailFunc: tail}
// 		ret = append(ret, ca)
// 	}
// 	return ret
// }

// generate a function to tail for a file
// func tailTheFile(path string, pos int64, h Handler) Tail {
// 	lastPos := pos
// 	file, err := os.Open(path)
// 	if err != nil {
// 		log.Fatal(err)
// 		return nil
// 	}
// 	return func() {
// 		info, err := file.Stat()
// 		if err != nil {
// 			log.Fatal(err)
// 			return
// 		}
//
// 		currentSize := info.Size()
//
// 		// current size of the file is smaller than the
// 		// last position, file may rotate, read from the
// 		// 0 position.
// 		if currentSize < lastPos {
// 			lastPos = 0
// 		}
//
// 		_, err = file.Seek(lastPos, 0)
// 		if err != nil {
// 			log.Fatal(err)
// 			return
// 		}
//
// 		lastPos = currentSize
// 		scanner := bufio.NewScanner(file)
// 		for scanner.Scan() {
// 			line := scanner.Text()
// 			h(line)
// 		}
// 	}
// }
