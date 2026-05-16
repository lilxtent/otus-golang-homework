package hw10programoptimization

import (
	"bufio"
	"io"
	"strings"
)

//easyjson:json
type User struct {
	Email string
}

type DomainStat map[string]int

func GetDomainStat(r io.Reader, domain string) (DomainStat, error) {
	domainStat := make(DomainStat)
	scanner := bufio.NewScanner(r)
	expectedEmailSuffix := "." + domain

	for scanner.Scan() {
		user := User{}
		if err := user.UnmarshalJSON(scanner.Bytes()); err != nil {
			return nil, err
		}

		if !strings.HasSuffix(user.Email, expectedEmailSuffix) {
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
