package txcoursecrawler

import (
	"coursecrawler/internal/repository"
	"fmt"
	"sync"
)

type SysCrawler struct {
	SysCrawlerPool chan chan Task
	TaskChannel    chan Task
	quit           chan bool
	wg             sync.WaitGroup
}

type SysCoursePkgListResp struct {
	RetCode int64                 `json:"retcode"`
	Result  *SysCoursePkgListData `json:"result"`
	Msg     string                `json:"msg"`
}

type SysCoursePkgListData struct {
	RetCode          int64           `json:"retcode"`
	SysCoursePkgList []*SysCoursePkg `json:"sys_course_pkg_list"`
}

type SysCoursePkg struct {
	SubjectPackageId string `json:"subject_package_id"`
}

type CoursePkgInfoResp struct {
	RetCode int64              `json:"retcode"`
	Result  *CoursePkgInfoData `json:"result"`
	Msg     string             `json:"msg"`
}

type CoursePkgInfoData struct {
	RetCode int64     `json:"retcode"`
	Courses []*Course `json:"courses"`
}

type Course struct {
	CID       int64      `json:"cid"`
	Name      string     `json:"name"`
	PreAmount int64      `json:"pre_amount"`
	AfAmount  int64      `json:"af_amount"`
	TeList    []*Teacher `json:"te_list"`
	ClassInfo *Class     `json:"class_info"`
}

type Teacher struct {
	Name string `json:"name"`
}

type Class struct {
	ClassID int64    `json:"class_id"`
	TuList  []*Tutor `json:"tu_list"`
}

type Tutor struct {
	Name string `json:"name"`
}

//系统课爬虫
func NewSysCrawler(sysCrawlerPool chan chan Task) *SysCrawler {
	c := &SysCrawler{
		SysCrawlerPool: sysCrawlerPool,
		TaskChannel:    make(chan Task),
		quit:           make(chan bool),
	}
	return c
}

func (sc *SysCrawler) Start() {
	go func() {
		for {
			sc.SysCrawlerPool <- sc.TaskChannel
			select {
			case task := <-sc.TaskChannel:
				sc.wg.Add(1)
				sc.handleTask(task)
				sc.wg.Done()
			case <-sc.quit:
				return
			}
		}
	}()
}

func (sc *SysCrawler) handleTask(task Task) {
	sysCoursePkgListData, err := sc.getSysCoursePkgList(task.Grade, task.Subject)
	if err != nil {
		fmt.Println(err)
		return
	}
	if sysCoursePkgListData == nil || sysCoursePkgListData.RetCode != 0 {
		fmt.Println("err sysCoursePkgListData: ", sysCoursePkgListData)
		return
	}
	var tmpCourses []*Course
	for _, sysCoursePkg := range sysCoursePkgListData.SysCoursePkgList {
		coursePkgInfoData, err := sc.getCoursePkgInfo(sysCoursePkg.SubjectPackageId)
		if err != nil {
			fmt.Println(err)
			continue
		}
		if coursePkgInfoData == nil || coursePkgInfoData.RetCode != 0 {
			fmt.Println("err getCoursePkgInfo: ", coursePkgInfoData)
			continue
		}
		tmpCourses = append(tmpCourses, coursePkgInfoData.Courses...)
	}
	records := DefaultCourseConv.convertCourseSliceToRecords(task.Grade, task.Subject, tmpCourses)
	for _, record := range records {
		repository.RepoInstance().UpsertCourse(record)
	}
}

func (sc *SysCrawler) getSysCoursePkgList(grade int64, subject int64) (*SysCoursePkgListData, error) {
	var sysCoursePkgListResp SysCoursePkgListResp
	url := fmt.Sprintf("https://fudao.qq.com/cgi-proxy/course/discover_subject?"+
		"client=4&platform=3&version=30&grade=%d&subject=%d&"+
		"showid=0&page=1&size=10&t=0.7507805918494652", grade, subject)
	err := apiRequest(url, nil, &sysCoursePkgListResp)
	if err != nil {
		return nil, err
	}
	if sysCoursePkgListResp.RetCode != 0 {
		return nil, fmt.Errorf("get getSysCoursePkgList err: %s", sysCoursePkgListResp.Msg)
	}
	return sysCoursePkgListResp.Result, nil
}

func (sc *SysCrawler) getCoursePkgInfo(subjectPkgId string) (*CoursePkgInfoData, error) {
	var coursePkgInfoResp CoursePkgInfoResp
	url := fmt.Sprintf("https://fudao.qq.com/cgi-proxy/course/get_course_package_info?"+
		"client=4&platform=3&version=30&"+
		"subject_package_id=%s&t=0.8022188257626726", subjectPkgId)
	err := apiRequest(url, nil, &coursePkgInfoResp)
	if err != nil {
		return nil, err
	}
	if coursePkgInfoResp.RetCode != 0 {
		return nil, fmt.Errorf("get coursePkgInfo err: %s", coursePkgInfoResp.Msg)
	}
	return coursePkgInfoResp.Result, nil
}

func (sc *SysCrawler) Stop() {
	sc.wg.Wait()
	sc.quit <- true
}
