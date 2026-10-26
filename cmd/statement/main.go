package main

import (
	"fmt"
	"log"
	"statement-export"
)

func main() {
	report := statement.Report{Educator: "Mina Chen", Courses: []statement.Course{{Name: "Intro to APIs", Learners: 24, Deadline: "2026-11-14"}}}
	pdfURL, err := statement.Run("Northstar Academy", report)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(pdfURL)
}
