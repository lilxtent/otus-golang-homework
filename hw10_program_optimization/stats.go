package hw10programoptimization

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"strings"
)

type User struct {
	ID       int
	Name     string
	Username string
	Email    string
	Phone    string
	Password string
	Address  string
}

type DomainStat map[string]int

func GetDomainStat(r io.Reader, domain string) (DomainStat, error) {
	u, err := getUsers(r)
	if err != nil {
		return nil, fmt.Errorf("get users error: %w", err)
	}
	return countDomains(u, domain)
}

type users [100_000]User

func getUsers(r io.Reader) (result users, err error) {
	scanner := bufio.NewScanner(r)
	index := 0

	for scanner.Scan() {
		var user User
		if err = json.Unmarshal(scanner.Bytes(), &user); err != nil {
			return
		}

		result[index] = user
		index++
	}

	if err = scanner.Err(); err != nil {
		return
	}

	return
}

func countDomains(u users, domain string) (DomainStat, error) {
	result := make(DomainStat)

	for _, user := range u {
		if !strings.HasSuffix(user.Email, "."+domain) {
			continue
		}

		indexOfAtSymbol := strings.Index(user.Email, "@")
		key := strings.ToLower(user.Email[indexOfAtSymbol+1:])
		result[key]++
	}

	return result, nil
}
