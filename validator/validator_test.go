package validator

import "testing"

func Test_Username(t *testing.T) {
	given := map[string]error{
		"a": ErrInvalidUsernameFormat,
		"aa": nil,
		"aaaaaaaaaaaaaaaaaaaa": nil,
		"aaaaaaaaaaaaaaaaaaaaa": ErrInvalidUsernameFormat,
		"azAZ": nil,
		"0123456789": nil,
		"---": nil,
		"___": nil,
		"`~!@#$%^&*()=+": ErrInvalidUsernameFormat,
		"{}|\"\\'/.,?><": ErrInvalidUsernameFormat,
	}

	for k, v := range given {
		valid := Username(k)
		if valid != v {
			t.Fatalf("error did not occured: given=%v expected=%v result=%v", k, valid, v)
		}
	}
}
