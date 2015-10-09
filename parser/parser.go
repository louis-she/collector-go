package parser

import (
	"fmt"
	"strings"
	"time"
)

type Sets struct {
	Test string
}

func (s Sets) Timefmt(path string) string {
	now := time.Now()
	path = strings.Replace(path, "%Y", fmt.Sprintf("%d", now.Year()), -1)
	path = strings.Replace(path, "%M", fmt.Sprintf("%d", now.Month()), -1)
	path = strings.Replace(path, "%zM", fmt.Sprintf("%02d", now.Month()), -1)
	path = strings.Replace(path, "%D", fmt.Sprintf("%d", now.Day()), -1)
	path = strings.Replace(path, "%zD", fmt.Sprintf("%02d", now.Day()), -1)
	path = strings.Replace(path, "%H", fmt.Sprintf("%d", now.Hour()), -1)
	path = strings.Replace(path, "%zH", fmt.Sprintf("%02d", now.Hour()), -1)
	path = strings.Replace(path, "%m", fmt.Sprintf("%d", now.Minute()), -1)
	path = strings.Replace(path, "%zm", fmt.Sprintf("%02d", now.Minute()), -1)
	path = strings.Replace(path, "%S", fmt.Sprintf("%d", now.Second()), -1)
	return path
}
