package main

import (
	"coursecrawler/txcoursecrawler"
	"flag"
	"log"
	"os"
	"time"
)

func init() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
}

func main() {
	t := time.Now()
	opts := txcoursecrawler.NewOptions()
	flagSet := flag.NewFlagSet("txcoursecrawler", flag.ExitOnError)
	flagSet.Int64Var(&opts.MaxSysCrawlers, "maxsyscrawlers", 1, "系统课爬虫最大并行数量")
	flagSet.Int64Var(&opts.MaxSpeCrawlers, "maxspecrawlers", 1, "专题课爬虫最大并行数量")
	flagSet.Parse(os.Args[1:])
	tc, err := txcoursecrawler.NewTxCourseCrawler(opts)
	if err != nil {
		log.Println(err)
		return
	}
	err = tc.Run()
	if err != nil {
		log.Println(err)
		return
	}
	log.Println(time.Since(t))
}
