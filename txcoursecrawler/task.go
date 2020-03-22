package txcoursecrawler

const (
	SysTask = iota
	SpeTask
)

type Task struct {
	Grade   int64
	Subject int64
	Type    int32
}
