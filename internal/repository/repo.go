package repository

import (
	"database/sql"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"os"
	"sync"
	"time"
)

const DBName = "tx_course"
const CrawlHistoryTableName = "crawl_history"
const Source = 1

var defaultRepository *repository
var singleInstance sync.Once
var (
	userName = os.Getenv("MYSQLUser")
	password = os.Getenv("MYSQLPass")
	ip       = os.Getenv("MYSQLAddr")
	port     = os.Getenv("MYSQLPort")
)

type repository struct {
	db *sql.DB
}

type CourseRecord struct {
	CourseID  int64
	GradeID   int64
	SubjectID int64
	Name      string
	PreAmount int64
	AfAmount  int64
	TeList    string
	TuList    string
}

type SubjectStats struct {
	History  *CrawlHistory
	Subjects []*SubjectStat
}

type CrawlHistory struct {
	StartDate int64
	EndDate   int64
}

type SubjectStat struct {
	SubjectId   int64
	CourseCount int64
}

func RepoInstance() *repository {
	singleInstance.Do(func() {
		defaultRepository = &repository{}
		conn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s",
			userName, password, ip, port, defaultRepository.getDBName())
		DB, err := sql.Open("mysql", conn)
		if err != nil {
			fmt.Println("mysql connect failed:", err)
			return
		}
		defaultRepository.db = DB
	})
	return defaultRepository
}

func (rp *repository) GetCrawlHistory() (*CrawlHistory, error) {
	crawlHistory := &CrawlHistory{}
	row := rp.db.QueryRow("select start_time, end_time from "+
		CrawlHistoryTableName+" where source=?", Source)
	if err := row.Scan(&crawlHistory.StartDate, &crawlHistory.EndDate); err != nil {
		fmt.Println("scan err: ", err)
		return nil, err
	}
	return crawlHistory, nil
}

func (rp *repository) GetSubjects(dateStr string) ([]*SubjectStat, error) {
	var subjects []*SubjectStat
	tableName, err := rp.getTableNameByDateStr(dateStr)
	if err != nil {
		return nil, err
	}
	rows, err := rp.db.Query("SELECT subject_id,COUNT(course_id) AS CourseCount FROM " +
		tableName + " GROUP BY subject_id ORDER BY subject_id")

	defer func() {
		if rows != nil {
			rows.Close()
		}
	}()
	if err != nil {
		fmt.Println("query failed: ", err)
		return nil, err
	}
	for rows.Next() {
		subject := &SubjectStat{}
		err = rows.Scan(&subject.SubjectId, &subject.CourseCount)
		if err != nil {
			fmt.Println("scan failed: ", err)
			return nil, err
		}
		subjects = append(subjects, subject)
	}
	return subjects, nil
}

func (rp *repository) GetAllCourses(subjectId int64, dateStr string) ([]*CourseRecord, error) {
	var courses []*CourseRecord
	tableName, err := rp.getTableNameByDateStr(dateStr)
	if err != nil {
		return nil, err
	}
	rows, err := rp.db.Query("SELECT course_id, name, pre_amount, "+
		"af_amount, te_list, tu_list FROM "+tableName+
		" WHERE subject_id = ?", subjectId)

	defer func() {
		if rows != nil {
			rows.Close()
		}
	}()
	if err != nil {
		fmt.Println("query failed: ", err)
		return nil, err
	}
	for rows.Next() {
		course := &CourseRecord{}
		err = rows.Scan(&course.CourseID, &course.Name, &course.PreAmount,
			&course.AfAmount, &course.TeList, &course.TuList,
		)
		if err != nil {
			fmt.Println("scan failed: ", err)
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
		fmt.Printf("upsert course failed,err: %+v", err)
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
		fmt.Println("create table failed:", err)
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
		fmt.Println("create table failed:", err)
		return err
	}

	sql = "INSERT INTO crawl_history(source, start_time, end_time) " +
		"values(?, ?, ?) ON DUPLICATE KEY UPDATE " +
		"end_time = ?"
	t := time.Now().Unix()
	if _, err := rp.db.Exec(sql, Source, t, t, t); err != nil {
		fmt.Println("update history failed:", err)
		return err
	}
	return nil
}

func (rp *repository) getDBName() string {
	return DBName
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
