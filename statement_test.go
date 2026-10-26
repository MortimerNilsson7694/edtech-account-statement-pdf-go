package statement

import "testing"

func TestBuildStatementRequiresUsefulCourseRows(t *testing.T) {
	tests := []struct {
		name    string
		report  Report
		wantErr bool
	}{
		{"complete delivery report", Report{Educator: "Mina", Courses: []Course{{Name: "APIs", Learners: 3, Deadline: "2026-11-14"}}}, false},
		{"empty learner cohort", Report{Educator: "Mina", Courses: []Course{{Name: "APIs", Learners: 0, Deadline: "2026-11-14"}}}, true},
		{"missing deadline", Report{Educator: "Mina", Courses: []Course{{Name: "APIs", Learners: 3}}}, true},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, err := BuildStatement("Academy", test.report, MeteredUsage{Total: 4})
			if (err != nil) != test.wantErr {
				t.Fatalf("error = %v, wantErr %v", err, test.wantErr)
			}
		})
	}
}
