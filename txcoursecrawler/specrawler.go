package txcoursecrawler

import (
	"coursecrawler/internal/repository"
	"fmt"
	"sync"
)

type SpeCrawler struct {
	SpeCrawlerPool chan chan Task
	TaskChannel    chan Task
	quit           chan bool
	wg             sync.WaitGroup
}

type SpeCourseListResp struct {
	RetCode int64              `json:"retcode"`
	Result  *SpeCourseListData `json:"result"`
	Msg     string             `json:"msg"`
}

type SpeCourseListData struct {
	RetCode       int64          `json:"retcode"`
	SpeCourseList *SpeCourseList `json:"spe_course_list"`
}

type SpeCourseList struct {
	Page  int64     `json:"page"`
	Size  int64     `json:"size"`
	Total int64     `json:"total"`
	Data  []*Course `json:"data"`
}

//专题课爬虫
func NewSpeCrawler(speCrawlerPool chan chan Task) *SpeCrawler {
	c := &SpeCrawler{
		SpeCrawlerPool: speCrawlerPool,
		TaskChannel:    make(chan Task),
		quit:           make(chan bool),
	}
	return c
}

func (spc *SpeCrawler) Start() {
	go func() {
		for {
			spc.SpeCrawlerPool <- spc.TaskChannel
			select {
			case task := <-spc.TaskChannel:
				spc.wg.Add(1)
				spc.handleTask(task)
				spc.wg.Done()
			case <-spc.quit:
				return
			}
		}
	}()
}

func (spc *SpeCrawler) handleTask(task Task) {
	var tmpCourses []*Course
	var page int64 = 1
	var size int64 = 10
	for {
		speCourseListData, err := spc.getSpeCourseListData(task.Grade, task.Subject, page, size)
		if err != nil {
			fmt.Println(err)
			break
		}
		if speCourseListData == nil || speCourseListData.RetCode != 0 ||
			speCourseListData.SpeCourseList == nil {
			fmt.Println("err getSpeCourseListData: ", speCourseListData)
			break
		}
		tmpCourses = append(tmpCourses, speCourseListData.SpeCourseList.Data...)
		if speCourseListData.SpeCourseList.Total <= (page * size) {
			break
		}
		page++
	}
	records := DefaultCourseConv.convertCourseSliceToRecords(task.Grade, task.Subject, tmpCourses)
	for _, record := range records {
		repository.RepoInstance().UpsertCourse(record)
	}
}

func (spc *SpeCrawler) getSpeCourseListData(grade int64, subject int64, page int64, size int64) (*SpeCourseListData, error) {
	var speCourseListResp SpeCourseListResp
	url := fmt.Sprintf("https://fudao.qq.com/cgi-proxy/course/discover_subject?"+
		"client=4&platform=3&version=30&grade=%d&subject=%d&"+
		"showid=0&page=%d&size=%d&t=0.7507805918494652", grade, subject, page, size)
	err := apiRequest(url, nil, &speCourseListResp)
	if err != nil {
		return nil, err
	}
	if speCourseListResp.RetCode != 0 {
		return nil, fmt.Errorf("get getSysCoursePkgList err: %s", speCourseListResp.Msg)
	}
	return speCourseListResp.Result, nil
}

func (spc *SpeCrawler) Stop() {
	spc.wg.Wait()
	spc.quit <- true
}
