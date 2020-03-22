package txcoursecrawler

type Options struct {
	MaxSysCrawlers int64
	MaxSpeCrawlers int64
}

func NewOptions() *Options {
	return &Options{}
}
