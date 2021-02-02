package models

import (
	"encoding/json"
	"fmt"
	"strconv"
)

// Question is a manager for accessing questions in the database
type Question struct {
	id int64
}

// NewQuestion creates a new question, saves it to the database, and returns the newly created question
func NewQuestion(statement, answer string, choices []string) error {
	exists, err := client.HExists("question:by-statement", statement).Result()
	if err != nil {
		return err
	}
	if exists {
		return ErrQuestionDuplicate
	}

	id, err := client.Incr("question:next-id").Result()
	if err != nil {
		return err
	}

	choicesBin, err := json.Marshal(choices)
	if err != nil {
		return err
	}

	key := fmt.Sprintf("question:%d", id)
	pipe := client.Pipeline()
	pipe.HSet(key, "id", id)
	pipe.HSet(key, "statement", statement)
	pipe.HSet(key, "answer", answer)
	pipe.HSet(key, "choices", choicesBin)
	pipe.HSet("question:by-statement", statement, id)
	pipe.LPush("questions", id)
	_, err = pipe.Exec()
	if err != nil {
		return err
	}

	return nil
}

// GetQuestionID QuestionID getter
func (q *Question) GetQuestionID() int64 { return q.id }

// GetStatement Statement getter
func (q *Question) GetStatement() (string, error) {
	key := fmt.Sprintf("question:%d", q.id)
	return client.HGet(key, "statement").Result()
}

// GetAnswer Answer getter
func (q *Question) GetAnswer() (string, error) {
	key := fmt.Sprintf("question:%d", q.id)
	return client.HGet(key, "answer").Result()
}

// GetChoices Choices getter
func (q *Question) GetChoices() ([]string, error) {
	key := fmt.Sprintf("question:%d", q.id)
	b, err := client.HGet(key, "choices").Bytes()
	if err != nil {
		return nil, err
	}

	var choices []string
	if err := json.Unmarshal(b, &choices); err != nil {
		return nil, err
	}

	return choices, nil
}

func queryQuestions(key string) ([]*Question, error) {
	questionIDs, err := client.LRange(key, 0, -1).Result()
	if err != nil {
		return nil, err
	}

	questions := make([]*Question, len(questionIDs))
	for i, val := range questionIDs {
		id, err := strconv.ParseInt(val, 10, 64)
		if err != nil {
			return nil, err
		}
		questions[i] = &Question{id: id}
	}
	return questions, nil
}

// GetAllQuestions All Updates getter
func GetAllQuestions() ([]*Question, error) {
	return queryQuestions("questions")
}

// GetQuestions gets all updates related to the user
func GetQuestions(userID int64) ([]*Question, error) {
	key := fmt.Sprintf("user:%d:questions", userID)
	return queryQuestions(key)
}

// QuestionT is a unit for the exam
type QuestionT struct {
	Statement string   `csv:"statement" json:"statement"`
	Answer    string   `csv:"answer" json:"answer"`
	Choices   []string `csv:"choices" json:"choices"`
}

// GetQuestion retrieves a whole struct of question from the database
func GetQuestion(q *Question) (*QuestionT, error) {
	s, err := q.GetStatement()
	if err != nil {
		return nil, err
	}

	c, err := q.GetChoices()
	if err != nil {
		return nil, err
	}

	a, err := q.GetAnswer()
	if err != nil {
		return nil, err
	}

	return &QuestionT{
		Statement: s,
		Choices:   c,
		Answer:    a,
	}, nil
}
