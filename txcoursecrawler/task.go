package txcoursecrawler

const (
	SysTask = iota //系统课任务
	SpeTask        //专题课任务
)

//任务
type Task struct {
	Grade   int64 //年级id
	Subject int64 //科目id
	Type    int32 //任务类型
}
