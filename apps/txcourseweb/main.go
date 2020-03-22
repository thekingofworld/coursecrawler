package main

import (
	"coursecrawler/txcourseweb"
	"log"
)

func init() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
}

func main() {
	tw, err := txcourseweb.NewTxCourseWeb()
	if err != nil {
		log.Println(err)
		return
	}
	tw.Run()
}
