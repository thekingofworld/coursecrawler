package repository

import (
	"database/sql"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"log"
	"os"
	"sync"
	"time"
)

const CrawlHistoryTableName = "crawl_history"
const Source = 1

var defaultRepository *repository
var singleInstance sync.Once
var (
	userName = os.Getenv("MYSQLUser")
	password = os.Getenv("MYSQLPass")
	ip       = os.Getenv("MYSQLAddr")
	port     = os.Getenv("MYSQLPort")
	dbName   = os.Getenv("MYSQLDBName")
)

type repository struct {
	db *sql.DB
}

//课程
type CourseRecord struct {
	CourseID  int64  //课程id
	GradeID   int64  //年级id
	SubjectID int64  //科目id
	Name      string //课程名称
	PreAmount int64  //原价
	AfAmount  int64  //折后价
	TeList    string //授课老师
	TuList    string //辅导老师
}

//抓取历史
type CrawlHistory struct {
	StartDate int64 //数据最早可查时间
	EndDate   int64 //数据最晚可查时间
}

//统计信息
type SubjectStat struct {
	GradeId     int64 //年级id
	SubjectId   int64 //科目id
	CourseCount int64 //课程数量
}

func RepoInstance() *repository {
	singleInstance.Do(func() {
		defaultRepository = &repository{}
		conn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s",
			userName, password, ip, port, defaultRepository.getDBName())
		DB, err := sql.Open("mysql", conn)
		if err != nil {
			log.Println("mysql connect failed:", err)
			return
		}
		defaultRepository.db = DB
	})
	return defaultRepository
}

//获取抓取历史
func (rp *repository) GetCrawlHistory() (*CrawlHistory, error) {
	crawlHistory := &CrawlHistory{}
	row := rp.db.QueryRow("select start_time, end_time from "+
		CrawlHistoryTableName+" where source=?", Source)
	if err := row.Scan(&crawlHistory.StartDate, &crawlHistory.EndDate); err != nil {
		log.Println("scan err: ", err)
		return nil, err
	}
	return crawlHistory, nil
}

//获取指定日期的课程统计数据
func (rp *repository) GetSubjects(dateStr string) ([]*SubjectStat, error) {
	var subjects []*SubjectStat
	tableName, err := rp.getTableNameByDateStr(dateStr)
	if err != nil {
		return nil, err
	}
	rows, err := rp.db.Query("SELECT grade_id,subject_id,COUNT(course_id) AS CourseCount FROM " +
		tableName + " GROUP BY grade_id,subject_id ORDER BY grade_id,subject_id")

	defer func() {
		if rows != nil {
			rows.Close()
		}
	}()
	if err != nil {
		log.Println("query failed: ", err)
		return nil, err
	}
	for rows.Next() {
		subject := &SubjectStat{}
		err = rows.Scan(&subject.GradeId, &subject.SubjectId, &subject.CourseCount)
		if err != nil {
			log.Println("scan failed: ", err)
			return nil, err
		}
		subjects = append(subjects, subject)
	}
	return subjects, nil
}

//获取指定日期、指定年级、科目对应的课程数据
func (rp *repository) GetAllCourses(gradeId, subjectId int64, dateStr string) ([]*CourseRecord, error) {
	var courses []*CourseRecord
	tableName, err := rp.getTableNameByDateStr(dateStr)
	if err != nil {
		return nil, err
	}
	rows, err := rp.db.Query("SELECT course_id, name, pre_amount, "+
		"af_amount, te_list, tu_list FROM "+tableName+
		" WHERE grade_id = ? AND subject_id = ?", gradeId, subjectId)

	defer func() {
		if rows != nil {
			rows.Close()
		}
	}()
	if err != nil {
		log.Println("query failed: ", err)
		return nil, err
	}
	for rows.Next() {
		course := &CourseRecord{}
		err = rows.Scan(&course.CourseID, &course.Name, &course.PreAmount,
			&course.AfAmount, &course.TeList, &course.TuList,
		)
		if err != nil {
			log.Println("scan failed: ", err)
			return nil, err
		}
		courses = append(courses, course)
	}
	return courses, nil
}

func (rp *repository) UpsertCourse(course *CourseRecord) {
	if course == nil {
		return
	}
	table := rp.getCurTableName()
	upSql := fmt.Sprintf("INSERT INTO %s(course_id, grade_id, subject_id, name, pre_amount, af_amount, te_list, tu_list) "+
		"values(?, ?, ?, ?, ?, ?, ?, ?) ON DUPLICATE KEY UPDATE "+
		"name = ?, pre_amount = ?, af_amount = ?, te_list = ?, tu_list = ?", table)
	_, err := rp.db.Exec(upSql, course.CourseID, course.GradeID, course.SubjectID,
		course.Name, course.PreAmount, course.AfAmount, course.TeList, course.TuList,
		course.Name, course.PreAmount, course.AfAmount, course.TeList, course.TuList,
	)
	if err != nil {
		log.Printf("upsert course failed,err: %+v", err)
		return
	}
}

func (rp *repository) CreateTable() error {
	sql := `CREATE TABLE IF NOT EXISTS ` + rp.getCurTableName() + `(
		id INT PRIMARY KEY AUTO_INCREMENT NOT NULL,
		course_id INT,
		grade_id INT,
		subject_id INT,
		name VARCHAR(50),
		pre_amount INT,
		af_amount INT,
		te_list VARCHAR(255),
		tu_list VARCHAR(255),
		UNIQUE (course_id)
	)ENGINE=InnoDB DEFAULT CHARSET=utf8;`

	if _, err := rp.db.Exec(sql); err != nil {
		log.Println("create table failed:", err)
		return err
	}
	return nil
}

func (rp *repository) UpdateHistory() error {
	sql := `CREATE TABLE IF NOT EXISTS crawl_history(
		id INT PRIMARY KEY AUTO_INCREMENT NOT NULL,
		source INT,
		start_time INT,
		end_time INT,
		UNIQUE (source)
	)ENGINE=InnoDB DEFAULT CHARSET=utf8;`

	if _, err := rp.db.Exec(sql); err != nil {
		log.Println("create table failed:", err)
		return err
	}

	sql = "INSERT INTO crawl_history(source, start_time, end_time) " +
		"values(?, ?, ?) ON DUPLICATE KEY UPDATE " +
		"end_time = ?"
	t := time.Now().Unix()
	if _, err := rp.db.Exec(sql, Source, t, t, t); err != nil {
		log.Println("update history failed:", err)
		return err
	}
	return nil
}

func (rp *repository) getDBName() string {
	return dbName
}

func (rp *repository) getCurTableName() string {
	t := time.Now()
	return fmt.Sprintf("course_%s", t.Format("20060102"))
}

func (rp *repository) getTableNameByDateStr(dateStr string) (string, error) {
	t, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("course_%s", t.Format("20060102")), nil
}
