package main

import (
	"fmt"
	"strings"
	"unicode"

	"github.com/emersion/go-imap/v2"
)

func buildSearchCriteria(filters []string) (*imap.SearchCriteria, error) {
	var searchCriteria *imap.SearchCriteria

	for _, filterExpr := range filters {
		if strings.TrimSpace(filterExpr) == "" {
			continue
		}

		newCriteria, err := parseFilter(filterExpr)
		if err != nil {
			return nil, fmt.Errorf("failed to parse filter expression %q: %w", filterExpr, err)
		}

		if searchCriteria == nil {
			searchCriteria = newCriteria
			continue
		}

		searchCriteria.And(newCriteria)
	}

	return searchCriteria, nil
}

func parseFilter(filterExpr string) (*imap.SearchCriteria, error) {
	criteria, _, err := parseFilterExpression([]rune(filterExpr), 0)
	return criteria, err
}

func parseFilterExpression(filterExpr []rune, i int) (*imap.SearchCriteria, int, error) {
	var (
		criteria *imap.SearchCriteria
		t        *imap.SearchCriteria
		opFunc   filterBoolFunc
		err      error
	)

	/*
		Parser syntax

		Expression:
			Expression || Term
			Term

		Term:
			Term && Expression
			Primary

		Primary:
			FlagToken
			HeaderToken == string
			HeaderToken ~= string
			HeaderToken != string
			!Expression

		Examples:
			JUNK || IMPORTANT && SEEN
			JUNK || IMPORTANT && !SEEN
			JUNK || FLAG && BODY ~= 'sas'
			JUNK && !(IMPORTANT && SEEN)
			!JUNK && !IMPORTANT || SEEN
	*/

parseLoop:
	for i < len(filterExpr) {
		c := filterExpr[i]

		switch {
		case unicode.IsLetter(c):
			criteria, i, err = parseFilterTerm(filterExpr, i)

		case c == '|' || c == '&' || c == '!':
			opFunc, i, err = parseFilterOp(filterExpr, i)

		case c == '(' && opFunc != nil:
			t, i, err = parseFilterExpression(filterExpr, i+1)
			criteria = opFunc(criteria, t)
			opFunc = nil

		case c == '(' && opFunc == nil:
			criteria, i, err = parseFilterExpression(filterExpr, i+1)

		case c == ')':
			i++
			break parseLoop

		case unicode.IsSpace(c):
			i++

		default:
			return nil, i, fmt.Errorf("unrecognized character '%c' at index %d", c, i)
		}

		if err != nil {
			return nil, i, err
		}
	}

	return criteria, 0, nil
}

func parseFilterTerm(filterExpr []rune, i int) (*imap.SearchCriteria, int, error) {
	return nil, 0, nil
}

func parseFilterOp(filterExpr []rune, i int) (filterBoolFunc, int, error) {
	return nil, 0, nil
}

func parseFilterPrimary(filterExpr []rune, i int) (*imap.SearchCriteria, int, error) {
	return nil, 0, nil
}

var flagTokens = []string{"JUNK", "SEEN", "DRAFT", "DELETED", "FLAGGED", "PHISHING", "WILDCARD", "FORWARDED", "IMPORTANT", "ANSWERED"}

type filterBoolFunc func(c1, c2 *imap.SearchCriteria) *imap.SearchCriteria

func addAndCriteria(c1, c2 *imap.SearchCriteria) *imap.SearchCriteria {
	return c1
}

func addOrCriteria(c1, c2 *imap.SearchCriteria) *imap.SearchCriteria {
	return c1
}

func addNotCriteria(c1, c2 *imap.SearchCriteria) *imap.SearchCriteria {
	return c1
}
