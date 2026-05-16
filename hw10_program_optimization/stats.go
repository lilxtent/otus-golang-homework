package hw10programoptimization

import (
	"bufio"
	"io"
	"strings"

	"github.com/mailru/easyjson"
)

//easyjson:json
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
	domainStat := make(DomainStat)
	scanner := bufio.NewScanner(r)

	for scanner.Scan() {
		var user User
		if err := easyjson.Unmarshal(scanner.Bytes(), &user); err != nil {
			return nil, err
		}

		if !strings.HasSuffix(user.Email, "."+domain) {
			continue
		}

		indexOfAtSymbol := strings.Index(user.Email, "@")
		key := strings.ToLower(user.Email[indexOfAtSymbol+1:])
		domainStat[key]++
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return domainStat, nil
}