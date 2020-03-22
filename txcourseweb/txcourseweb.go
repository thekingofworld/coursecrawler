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
	SubjectId   int64
	Name        string
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
			SubjectId:   s.SubjectId,
			Name:        subjectIdToName[s.SubjectId],
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
	if err == nil && len(queryForm["subjectId"]) > 0 && len(queryForm["history"]) > 0 {
		subjectId, err := strconv.ParseInt(queryForm["subjectId"][0], 10, 64)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		history := queryForm["history"][0]
		courseRecords, err := repository.RepoInstance().GetAllCourses(subjectId, history)
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
