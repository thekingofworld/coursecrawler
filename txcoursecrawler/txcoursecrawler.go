package txcoursecrawler

import (
	"coursecrawler/internal/repository"
	"fmt"
)

//爬虫实例
type TxCourseCrawler struct {
	opts *Options
}

type GradeSubjectsResp struct {
	RetCode int64             `json:"retcode"`
	Result  *GradeSubjectData `json:"result"`
	Msg     string            `json:"msg"`
}

type GradeSubjectData struct {
	GradeSubjects []*GradeSubject `json:"grade_subjects"`
}

type GradeSubject struct {
	Grade   int64   `json:"grade"`
	Subject []int64 `json:"subject"`
}

func NewTxCourseCrawler(opts *Options) (*TxCourseCrawler, error) {
	c := &TxCourseCrawler{
		opts: opts,
	}
	return c, nil
}

func (tc *TxCourseCrawler) Run() error {
	if err := repository.RepoInstance().CreateTable(); err != nil {
		return err
	}
	gradeSubjectData, err := tc.getGradeSubjectData()
	if err != nil {
		return err
	}
	taskQueue := make(chan Task)
	dispatcher := NewDispatcher(tc.opts.MaxSysCrawlers, tc.opts.MaxSpeCrawlers, taskQueue)
	dispatcher.Run()
	for _, gb := range gradeSubjectData.GradeSubjects {
		for _, sb := range gb.Subject {
			sysTask := Task{
				Grade:   gb.Grade,
				Subject: sb,
				Type:    SysTask,
			}
			speTask := Task{
				Grade:   gb.Grade,
				Subject: sb,
				Type:    SpeTask,
			}
			taskQueue <- sysTask
			taskQueue <- speTask
		}
	}
	dispatcher.Stop()
	if err := repository.RepoInstance().UpdateHistory(); err != nil {
		return err
	}
	return nil
}

func (tc *TxCourseCrawler) getGradeSubjectData() (*GradeSubjectData, error) {
	var gradeSubjectResp GradeSubjectsResp
	url := "https://fudao.qq.com/cgi-proxy/course/grade_subject?t=0.08175389807474542"
	err := apiRequest(url, nil, &gradeSubjectResp)
	if err != nil {
		return nil, err
	}
	if gradeSubjectResp.RetCode != 0 {
		return nil, fmt.Errorf("get gradeSubjects err: %s", gradeSubjectResp.Msg)
	}
	return gradeSubjectResp.Result, nil
}
