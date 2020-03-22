package txcoursecrawler

type Options struct {
	MaxSysCrawlers int64 //最大并行系统课爬虫数量
	MaxSpeCrawlers int64 //最大并行专题课爬虫数量
}

func NewOptions() *Options {
	return &Options{}
}
