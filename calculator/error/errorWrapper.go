package error

import (
	"errors"
	"fmt"
)

func WrapErrorExample() {
	// 1. Root Cause
	err1 := errors.New("connection refused")

	// 2. Middle layer wraps err1 inside err2
	err2 := fmt.Errorf("db failed: %w", err1)

	// 3. Top layer wraps err2 inside err3
	err3 := fmt.Errorf("http 500: %w", err2)

	fmt.Println("the final error err3:", err3)
}


func GetUserNameExample(id int) (string, error) {
	if id <= 0 {
		return "", fmt.Errorf("invalid user ID: %d", id)
	}
	userName , err := fetchUserFromDb(id);
	return userName, fmt.Errorf("failed to get user name for ID %d: %w", id, err)
}

func fetchUserFromDb(id int) (string, error) {
	// some logic here to fetch

	// assuming logic fails and we get an error from db or query

	return "", errors.New("something went wrong while querying from DB")
}