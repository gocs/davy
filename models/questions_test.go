package models

import (
	"testing"
)

func TestLoadQuestions(t *testing.T) {
	NewRedisDB()
	t.Log(MigrateQuestions(client, "../private/questions.json"))
}

func TestDiff(t *testing.T) {
	n := []string{"1", "2", "3", "4", "5"}
	o := []string{"1", "2", "3"}
	expected := []string{"5", "4"}

	result := diff(n, o)
	for i := 0; i < len(expected); i++ {
		if expected[i] != result[i] {
			t.Errorf("not same: expected=%s, result=%s", expected[i], result[i])
		}
	}
}
