package handler

import (
	//"bufio"
	"fmt"
	"io/ioutil"
)

type Sets struct{}

func (s Sets) WriteTmp(line string) string {
	file := "./tmpfile"
	ioutil.WriteFile(file, []byte(line), 0644)
	return ""
}

func (s Sets) Println(line string) string {
	fmt.Println(line)
	return ""
}
