package main

import (
	"coursecrawler/txcourseweb"
	"fmt"
)

func main() {
	tw, err := txcourseweb.NewTxCourseWeb()
	if err != nil {
		fmt.Println(err)
		return
	}
	tw.Run()
}
