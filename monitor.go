//logs collector
package main

import (
	"./handler"
	"./parser"
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"reflect"
	"strings"
	"time"
)

// receiver of the json config
type ConfigEntity struct {
	Path         []string
	PathParser   []string
	HandlerChain [][]string
	Timespan     int64
}

//receiver of the json config
type Configuration struct {
	Entity []ConfigEntity
}

// like tail -f
type Tail func()

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

// Supporting handlers and parsers
var supportHandlers handler.Sets
var supportParsers parser.Sets

// apply json config to Configuration type
func applyConfig(conf string, c *Configuration) error {
	file, err := os.Open(conf)
	if err != nil {
		return err
	}
	decoder := json.NewDecoder(file)
	err = decoder.Decode(c)
	return err
}

func genEntity(file string, e ConfigEntity, me *MonitorEntity) {
	var logmsg string
	me.path = file
	// apply file name parsers
	for _, parser := range e.PathParser {
		method := reflect.ValueOf(supportParsers).MethodByName(parser)
		if method.IsValid() == false {
			// parser function is not valid
			logmsg = fmt.Sprintf("parser %s is not valid function", parser)
			log.Println(logmsg)
			continue
		}
		me.parser = append(me.parser, method)
	}

	// apply file line handler chain
	for _, chain := range e.HandlerChain {
		tmpChain := make([]reflect.Value, 0)
		for _, fHandler := range chain {
			spt := strings.Split(fHandler, "(")
			fHandler = spt[0]
			method := reflect.ValueOf(supportHandlers).MethodByName(fHandler)
			if len(spt) == 2 && spt[1] == ")" {
				// fHandler is a first class function,
				// should call it first
				method = method.Call([]reflect.Value{})[0]
			}
			if method.IsValid() == false {
				// line handler function is not valid
				logmsg = fmt.Sprintf("handler %s is not valid function", fHandler)
				log.Println(logmsg)
				continue
			}
			tmpChain = append(tmpChain, method)
		}
		me.handler = append(me.handler, tmpChain)
	}
	me.tail = genTailFunc(me)
}

func main() {
	// Read the config file
	c := Configuration{}
	err := applyConfig("conf.json", &c)
	if err != nil {
		log.Fatal(err)
		return
	}

	// MonitorEntities hold all the monitor
	// entities, a file to be monitored count
	// a entity
	var monitorEntities []MonitorEntity

	supportHandlers = handler.Sets{}
	supportParsers = parser.Sets{}

	for _, e := range c.Entity {
		// A monitor is just for a file, it may be
		// confused as in the config file where a
		// entity seems like to associate with several
		// files. But in the real an entity is just for
		// one file, this is why the me := MonitorEntity
		// is not called here.
		for _, p := range e.Path {
			me := MonitorEntity{}
			genEntity(p, e, &me)
			monitorEntities = append(monitorEntities, me)
		}
	}

	// Dead loop to monitor all the files
	for {
		time.Sleep(1000000000)
		now := time.Now().Unix()
		for _, entity := range monitorEntities {
			if now-entity.lasexec < entity.span || entity.running == true {
				// not this entity by now
				continue
			}
			entity.tail()
		}
	}
}

// First class function to create custom
// tail function for each entity
func genTailFunc(me *MonitorEntity) Tail {
	var lastPos int64
	lastPos = 0
	return func() {
		// generate file path by parser chain
		me.running = true
		path := me.path
		for _, p := range me.parser {
			res := p.Call([]reflect.Value{reflect.ValueOf(path)})
			path = res[0].String()
		}

		// open the file
		file, err := os.Open(path)
		defer file.Close()
		if err != nil {
			log.Println(err)
			return
		}

		// get size of the file
		info, err := file.Stat()
		if err != nil {
			log.Println(err)
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
			log.Println(err)
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
				result := reflect.ValueOf(line)
				for _, handler := range chain {
					res := handler.Call([]reflect.Value{result})
					result = res[0]
				}
			}
		}

		me.lasexec = time.Now().Unix()
		me.running = false
	}
}
