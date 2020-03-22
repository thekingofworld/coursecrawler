package txcourseweb

import (
	"coursecrawler/internal/repository"
	"coursecrawler/internal/util"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"time"
)

var gradeIdToName = map[int64]string{
	5001: "高一",
	5002: "高二",
	5003: "高三",
	6001: "初一",
	6002: "初二",
	6003: "初三",
	7001: "一年级",
	7002: "二年级",
	7003: "三年级",
	7004: "四年级",
	7005: "五年级",
	7006: "六年级",
	8001: "小班",
	8002: "中班",
	8003: "大班",
}

var subjectIdToName = map[int64]string{
	6001: "语文",
	6002: "数学",
	6003: "化学",
	6004: "物理",
	6005: "英语",
	6006: "生物",
	6007: "政治",
	6008: "历史",
	6009: "地理",
	6010: "讲座",
	7057: "编程",
	7058: "科学",
}

var courseTemplate *template.Template

type TxCourseWeb struct {
}

type SubjectStats struct {
	CurDate   string
	StartDate string
	EndDate   string
	Subjects  []*SubjectStat
}

type SubjectStat struct {
	GradeId     int64
	SubjectId   int64
	GradeName   string
	SubjectName string
	CourseCount int64
}

type Course struct {
	CourseID  int64
	Name      string
	PreAmount int64
	AfAmount  int64
	TeList    string
	TuList    string
}

func NewTxCourseWeb() (*TxCourseWeb, error) {
	return &TxCourseWeb{}, nil
}

func (tcw *TxCourseWeb) Run() {
	var err error
	err = os.Chdir(fmt.Sprintf("%s%s..%s%s", util.GetAppPath(),
		string(os.PathSeparator), string(os.PathSeparator), "txcourseweb"))
	if err != nil {
		panic(err)
	}
	courseTemplate, err = template.New("index.html").Funcs(template.FuncMap{
		"divide": func(a, b int64) int64 {
			return a / b
		},
	}).ParseFiles("static/html/index.html", "static/html/subject.html")
	if err != nil {
		panic(err)
	}
	http.HandleFunc("/", indexHandler)
	http.HandleFunc("/subject", subjectHandler)
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
	dateStr := time.Now().Format("2006-01-02")
	crawlHistory, err := repository.RepoInstance().GetCrawlHistory()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	dateStr = time.Unix(crawlHistory.EndDate, 0).Format("2006-01-02")
	queryForm, err := url.ParseQuery(r.URL.RawQuery)
	if err == nil && len(queryForm["history"]) > 0 {
		if len(queryForm["history"][0]) > 0 {
			dateStr = queryForm["history"][0]
		}
	}
	stats := &SubjectStats{}
	stats.CurDate = dateStr
	stats.StartDate = time.Unix(crawlHistory.StartDate, 0).Format("2006-01-02")
	stats.EndDate = time.Unix(crawlHistory.EndDate, 0).Format("2006-01-02")
	data, err := repository.RepoInstance().GetSubjects(dateStr)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	for _, s := range data {
		stats.Subjects = append(stats.Subjects, &SubjectStat{
			GradeId:     s.GradeId,
			SubjectId:   s.SubjectId,
			GradeName:   gradeIdToName[s.GradeId],
			SubjectName: subjectIdToName[s.SubjectId],
			CourseCount: s.CourseCount,
		})
	}
	err = courseTemplate.ExecuteTemplate(w, "index.html", stats)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func subjectHandler(w http.ResponseWriter, r *http.Request) {
	var courses []*Course
	queryForm, err := url.ParseQuery(r.URL.RawQuery)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if len(queryForm["gradeId"]) > 0 && len(queryForm["subjectId"]) > 0 &&
		len(queryForm["history"]) > 0 {
		gradeId, err := strconv.ParseInt(queryForm["gradeId"][0], 10, 64)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		subjectId, err := strconv.ParseInt(queryForm["subjectId"][0], 10, 64)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		history := queryForm["history"][0]
		courseRecords, err := repository.RepoInstance().GetAllCourses(gradeId, subjectId, history)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		for _, c := range courseRecords {
			courses = append(courses, &Course{
				CourseID:  c.CourseID,
				Name:      c.Name,
				PreAmount: c.PreAmount,
				AfAmount:  c.AfAmount,
				TeList:    c.TeList,
				TuList:    c.TuList,
			})
		}
	}
	err = courseTemplate.ExecuteTemplate(w, "subject.html", courses)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
