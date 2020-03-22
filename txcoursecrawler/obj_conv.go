package txcoursecrawler

import (
	"coursecrawler/internal/repository"
	"strings"
)

var DefaultCourseConv = courseConv{}

type courseConv struct {
}

func (cc courseConv) convertCourseToRecord(gradeId, subjectId int64, course *Course) *repository.CourseRecord {
	if course == nil {
		return nil
	}
	courseRecord := &repository.CourseRecord{
		CourseID:  course.CID,
		GradeID:   gradeId,
		SubjectID: subjectId,
		Name:      course.Name,
		PreAmount: course.PreAmount,
		AfAmount:  course.AfAmount,
		TeList:    "",
		TuList:    "",
	}
	var tmpTeList []string
	var tmpTuList []string
	for _, t := range course.TeList {
		tmpTeList = append(tmpTeList, t.Name)
	}
	courseRecord.TeList = strings.Join(tmpTeList, ",")
	if course.ClassInfo != nil {
		for _, tu := range course.ClassInfo.TuList {
			tmpTuList = append(tmpTuList, tu.Name)
		}
		courseRecord.TuList = strings.Join(tmpTuList, ",")
	}
	return courseRecord
}

func (cc courseConv) convertCourseSliceToRecords(gradeId, subjectId int64, courses []*Course) []*repository.CourseRecord {
	var data []*repository.CourseRecord
	for _, course := range courses {
		courseRecord := cc.convertCourseToRecord(gradeId, subjectId, course)
		if courseRecord != nil {
			data = append(data, courseRecord)
		}
	}
	return data
}
