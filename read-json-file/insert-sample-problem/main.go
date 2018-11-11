package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"os"

	"github.com/go-pg/pg"
	"github.com/go-pg/pg/orm"
	"github.com/jinzhu/copier"
	"github.com/rs/xid"
)

type Problem struct {
	// Tên bảng
	TableName struct{} `json:"table_name" sql:"problem.problems"`
	// Mã problem (chuỗi ngẫu nhiên duy nhất)
	Id string `json:"id"`
	// Tiêu đề
	Title string `json:"title"`
	// Mô tả
	Description string `json:"description"`
	// Input
	InputFormat string `json:"input_format"`
	// Output
	OutputFormat string `json:"output_format"`
	// Ngôn ngữ lập trình hỗ trợ
	Languages []string `json:"languages" pg:",array"`
	// Có hỗ trợ tất cả các ngôn ngữ không
	SupportAllLanguages bool `json:"support_all_languages" sql:"default:false"`
	// Có phải problem mẫu không
	IsSample int32 `json:"is_sample" sql:"default:0"`
	// Tenant id
	TenantID string `json:"tenant_id"`
	// Người tạo
	CreatedBy string `json:"created_by"`
	// Testcases
	Testcases []string `json:"testcases" pg:",array"`
	// Trạng thái
	Status int32 `json:"status" sql:"default:0"`
}

type Testcase struct {
	// Tên bảng
	TableName struct{} `json:"table_name" sql:"problem.testcase"`
	// Mã testcase (chuỗi ngẫu nhiên duy nhất)
	Id string `json:"id"`
	// Stdin
	Stdin []string `json:"stdin" pg:",array"`
	// Stdout
	Stdout string `json:"stdout"`
	// Status
	Status int32 `json:"status" sql:"default:0"`
}

type SampleCode struct {
	// Tên bảng
	TableName struct{} `json:"table_name" sql:"problem.sample_code"`
	// Mã User (chuỗi ngẫu nhiên duy nhất)
	ProblemId string `json:"problem_id" sql:",pk"`
	// Source code
	SourceCode string `json:"source_code"`
	// Ngôn ngữ
	Language string `json:"language" sql:",pk"`
}

type CreateSampleProblem struct {
	Title        string                 `json:"title"`
	Description  string                 `json:"description"`
	InputFormat  string                 `json:"input_format"`
	OutputFormat string                 `json:"output_format"`
	Testcases    []CreatePublicTestcase `json:"testcases"`
	SampleCode   []CreateSampleCode     `json:"sample_code"`
}

type CreatePublicTestcase struct {
	Stdin  []string `json:"stdin"`
	Stdout string   `json:"stdout"`
}

type CreateSampleCode struct {
	// Source code
	SourceCode string `json:"source_code"`
	// Ngôn ngữ
	Language string `json:"language"`
}

func createTable(model interface{}, db *pg.DB) error {
	err := db.CreateTable(model, &orm.CreateTableOptions{
		Temp:          false,
		FKConstraints: true,
		IfNotExists:   true,
	})

	return err
}

func main() {
	db := pg.Connect(&pg.Options{
		Addr:     "192.168.1.37:5432",
		Database: "postgres",
		User:     "postgres",
		Password: "123456",
	})
	defer db.Close()

	// // Tạo schema theo tên service
	// _, err := db.Exec("CREATE SCHEMA IF NOT EXISTS problem;")
	// if err != nil {
	// 	panic(err)
	// }

	// // Tạo bảng
	// var problemT Problem
	// var testcase Testcase
	// var sampleCode SampleCode
	// err = createTable(&problemT, db)
	// if err != nil {
	// 	panic(err)
	// }
	// err = createTable(&testcase, db)
	// if err != nil {
	// 	panic(err)
	// }
	// err = createTable(&sampleCode, db)
	// if err != nil {
	// 	panic(err)
	// }

	err := CreateDefaultProblem(db)
	if err != nil {
		panic (err)
	}
}

func CreateDefaultProblem(db *pg.DB) error {
	files, err := ioutil.ReadDir("json/")
    if err != nil {
        return err
    }

    for _, f := range files {
		jsonFile, err := os.Open("json/" + f.Name())
		if err != nil {
			return err
		}

		byteValue, err := ioutil.ReadAll(jsonFile)
		if err != nil {
			return err
		}

		var data CreateSampleProblem
		json.Unmarshal(byteValue, &data)
		err = CreateSampleProblemFunc(db, data)
		if err !=nil {
			return err
		}

		jsonFile.Close()
	}
	
	return nil
}

func CreateSampleProblemFunc(db *pg.DB, data CreateSampleProblem) error {
	var problem Problem
	copier.Copy(&problem, &data)
	problem.Id = xid.New().String()

	// Lưu source code
	for _, code := range data.SampleCode {
		// TODO: kiểm tra ngôn ngữ có hợp lệ

		var sc SampleCode
		sc.Language = code.Language
		sc.ProblemId = problem.Id
		sc.SourceCode = code.SourceCode
		err := db.Insert(&sc)
		if err != nil {
			return err
		}

		problem.Languages = append(problem.Languages, code.Language)
	}

	// Tạo testcase
	for _, testcase := range data.Testcases {
		var tc Testcase
		tc.Id = xid.New().String()
		tc.Status = 1
		tc.Stdin = testcase.Stdin
		tc.Stdout = testcase.Stdout
		err := db.Insert(&tc)
		if err != nil {
			return err
		}
		problem.Testcases = append(problem.Testcases, tc.Id)
	}

	// Thêm sample problem
	problem.Status = 1
	problem.SupportAllLanguages = false
	problem.TenantID = "1"
	problem.IsSample = 1
	problem.CreatedBy = ""
	err := db.Insert(&problem)
	if err != nil {
		return err
	}

	log.Println("Thêm thành công")
	return nil
}