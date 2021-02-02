package models

import (
	"errors"
	"fmt"
	"math/rand"
	"strconv"
	"time"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

// UserQuestion is a manager for accessing users' current question status in the database
type UserQuestion struct {
	id int64
}

// newUserQuestion creates a new question for the user, saves it to the database, and returns the newly created question
func newUserQuestion(userID int64) (*UserQuestion, error) {
	exists, err := client.HExists("user-question:by-username", fmt.Sprint(userID)).Result()
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, errors.New("question already given to the current user")
	}
	id, err := client.Incr("user-question:next-id").Result()
	if err != nil {
		return nil, err
	}

	key := fmt.Sprintf("user-question:%d", id)
	pipe := client.Pipeline()
	pipe.HSet(key, "id", id)
	pipe.HSet(key, "user_id", userID)
	pipe.HSet(key, "question_id", 1)
	pipe.HSet(key, "points", 0)
	pipe.HSet(key, "rank", 0)
	pipe.LPush(fmt.Sprintf("user:%d:questions", userID), 1)
	pipe.LPush(fmt.Sprintf("user:%d:user-question", userID), id)
	_, err = pipe.Exec()
	if err != nil {
		return nil, err
	}

	return &UserQuestion{id: id}, nil
}

// GetUser User getter
func (uq *UserQuestion) GetUser() (*User, error) {
	key := fmt.Sprintf("user-question:%d", uq.id)
	id, err := client.HGet(key, "user_id").Int64()
	if err != nil {
		return nil, err
	}

	return &User{id: id}, nil
}

// GetQuestion Question getter
func (uq *UserQuestion) GetQuestion() (*Question, error) {
	key := fmt.Sprintf("user-question:%d", uq.id)
	id, err := client.HGet(key, "question_id").Int64()
	if err != nil {
		return nil, err
	}

	return &Question{id: id}, nil
}

// GetUserQuestions Questions by the user getter
func (uq *UserQuestion) GetUserQuestions() ([]*Question, error) {

	userID, err := uq.GetUser()
	if err != nil {
		return nil, err
	}

	// get questions given to the user
	key := fmt.Sprintf("user-question:%d:questions", userID)
	return queryQuestions(key)
}

// GetUserQuestionsLen Questions len getter
func GetUserQuestionsLen() (int64, error) {
	return client.LLen("questions").Result()
}

// GetUnansweredQuestions gets all questions user haven't answered
func (uq *UserQuestion) GetUnansweredQuestions() ([]*Question, error) {
	userID, err := uq.GetUser()
	if err != nil {
		return nil, err
	}

	key := fmt.Sprintf("user:%d:questions", userID)
	userQuestionIDs, err := client.LRange(key, 0, -1).Result()
	if err != nil {
		return nil, err
	}

	questionIDs, err := client.LRange("questions", 0, -1).Result()
	if err != nil {
		return nil, err
	}

	ids := diff(questionIDs, userQuestionIDs)

	questions := make([]*Question, len(ids))
	for i, val := range ids {
		id, err := strconv.ParseInt(val, 10, 64)
		if err != nil {
			return nil, err
		}
		questions[i] = &Question{id: id}
	}
	return questions, nil
}

// diff gets the list of items m that are not in m
// see test
func diff(minu, subt []string) []string {
	s := minu[:]
	for i := 0; i < len(minu); i++ {
		for j := 0; j < len(subt); j++ {
			if minu[i] == subt[j] {
				s = remove(s, i)
				break
			}
		}
	}
	return s
}

func remove(s []string, i int) []string {
	s[i] = s[len(s)-1]
	return s[:len(s)-1]
}

// RegisterQuestion Question setter
func (uq *UserQuestion) RegisterQuestion(questionID int64) error {
	key := fmt.Sprintf("user-question:%d", uq.id)
	pipe := client.Pipeline()
	pipe.HSet(key, "question_id", questionID)
	pipe.LPush("questions", questionID)
	_, err := pipe.Exec()
	return err
}

// GetPoints Points getter
func (uq *UserQuestion) GetPoints() (int64, error) {
	key := fmt.Sprintf("user-question:%d", uq.id)
	points, err := client.HGet(key, "points").Int64()
	if err != nil {
		return 0, err
	}

	return points, nil
}

// AddPoints Points setter
func (uq *UserQuestion) AddPoints(amount int64) error {
	key := fmt.Sprintf("user-question:%d", uq.id)

	currPts, err := uq.GetPoints()
	if err != nil {
		return err
	}

	pipe := client.Pipeline()
	pipe.HSet(key, "points", currPts+amount)
	_, err = pipe.Exec()
	return err
}

// GetUserQuestion gets UserQuestion using user id
func GetUserQuestion(userID int64) (*UserQuestion, error) {
	key := fmt.Sprintf("user:%d:user-question", userID)

	userQuestionIDs, err := client.LRange(key, 0, 0).Result()
	if err != nil {
		return nil, err
	}

	if len(userQuestionIDs) == 0 {
		return nil, ErrEmptyUserQuestion
	}

	id, err := strconv.ParseInt(userQuestionIDs[0], 10, 64)
	if err != nil {
		return nil, err
	}

	return &UserQuestion{id: id}, nil
}

// UserConfirmAnswer returns true of false if the answer is correct and it gives the user's points wtih new question as a result of choosing the right answer
func UserConfirmAnswer(userID int64, choice string) (bool, error) {
	uq, err := GetUserQuestion(userID)
	if err != nil {
		return false, err
	}

	q, err := uq.GetQuestion()
	if err != nil {
		return false, err
	}

	qt, err := GetQuestion(q)
	if err != nil {
		return false, err
	}

	// if choice is incorrect return false without error
	if choice != qt.Answer {
		return false, nil
	}

	qs, err := uq.GetUnansweredQuestions()
	if err != nil {
		return false, err
	}

	currPts, err := uq.GetPoints()
	if err != nil {
		return false, err
	}

	// the ff. should not return true until all the steps are completed.
	// FIXME: consider more atomic transaction

	
	questionID := qs[rand.Intn(len(qs))].id
	
	pipe := client.Pipeline()
	
	key := fmt.Sprintf("user-question:%d", uq.id)
	pipe.HSet(key, "points", currPts+1)
	pipe.HSet(key, "question_id", questionID)
	pipe.LPush("questions", questionID)
	if _, err := pipe.Exec(); err != nil {
		return false, err
	}

	return true, nil
}
