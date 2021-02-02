package models

import (
	"fmt"
	"strconv"

	"github.com/go-redis/redis"
)

const leaderboard = "leaderboard"

// UpdateRank replaces the user's points with a new value
func UpdateRank(userID, points int64) error {
	z := redis.Z{Score: float64(points), Member: userID}
	_, err := client.ZAdd(leaderboard, z).Result()
	return err
}

// GetRank returns the current rank of the user
func GetRank(userID int64) int64 {
	return client.ZRank(leaderboard, fmt.Sprint(userID)).Val()
}

// RankT is a simple data struct for a sorted leaderboard
type RankT struct {
	Rank  int
	Name  string
	Score int64
}

func listLeaderboard(z *redis.ZSliceCmd) ([]RankT, error) {
	ut := []RankT{}
	for i, data := range z.Val() {
		s, ok := data.Member.(string)
		if !ok {
			return nil, ErrTypeMismatch
		}
		id, err := strconv.ParseInt(s, 10, 64)
		if err != nil {
			return nil, err
		}

		u, err := GetUserByUserID(id)
		if err != nil {
			return nil, err
		}

		un, err := u.GetUsername()
		if err != nil {
			return nil, err
		}

		ut = append(ut, RankT{Rank: i + 1, Name: un, Score: int64(data.Score)})
	}
	return ut, nil
}

// TopRanks lists the overall top 25
func TopRanks() ([]RankT, error) {
	z := client.ZRevRangeWithScores(leaderboard, 0, 24)
	return listLeaderboard(z)
}

// GetCurrentStandings list the raks of the 12 users above and below your current standing
func GetCurrentStandings(userID int64) ([]RankT, error) {
	zRank := client.ZRank(leaderboard, fmt.Sprint(userID))
	lower := zRank.Val() - 12
	upper := zRank.Val() + 12

	z := client.ZRangeWithScores(leaderboard, lower, upper)
	return listLeaderboard(z)
}
