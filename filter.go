package main

import (
	"errors"
	"fmt"
	"strings"
	"unicode"

	"github.com/emersion/go-imap/v2"
)

func parseFilter(filterExpr string) (*imap.SearchCriteria, error) {
	criteria, _, err := parseFilterExpression([]rune(filterExpr), 0)
	return criteria, err
}

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
		HeaderToken == String
		HeaderToken != String
		MsgToken == String
		MsgToken != String
		!Primary
		( Expression )
*/

func parseFilterExpression(filterExpr []rune, i int) (*imap.SearchCriteria, int, error) {
	var criteria *imap.SearchCriteria
	var err error

	criteria, i, err = parseFilterTerm(filterExpr, i)
	if err != nil {
		return nil, i, err
	}

	var (
		t      *imap.SearchCriteria
		opFunc filterBoolFunc
	)

	for i < len(filterExpr) {
		c := filterExpr[i]

		switch {
		case unicode.IsSpace(c):
			i++

		case c == '|':
			opFunc, i, err = parseFilterBoolOp(filterExpr, i+1, '|')
			if err != nil {
				return nil, i, err
			}

			t, i, err = parseFilterTerm(filterExpr, i)
			if err != nil {
				return nil, i, err
			}

			criteria = opFunc(criteria, t)

		default:
			return criteria, i + 1, nil
		}
	}

	return criteria, i, nil
}

func parseFilterTerm(filterExpr []rune, i int) (*imap.SearchCriteria, int, error) {
	var criteria *imap.SearchCriteria
	var err error

	criteria, i, err = parseFilterPrimary(filterExpr, i)
	if err != nil {
		return nil, i, err
	}

	var (
		t      *imap.SearchCriteria
		opFunc filterBoolFunc
	)

	for i < len(filterExpr) {
		c := filterExpr[i]

		switch {
		case unicode.IsSpace(c):
			i++

		case c == '&':
			opFunc, i, err = parseFilterBoolOp(filterExpr, i+1, '&')
			if err != nil {
				return nil, i, err
			}

			t, i, err = parseFilterExpression(filterExpr, i)
			if err != nil {
				return nil, i, err
			}

			criteria = opFunc(criteria, t)
			i++

		default:
			return criteria, i, nil
		}
	}

	return criteria, i, nil
}

func parseFilterPrimary(filterExpr []rune, i int) (*imap.SearchCriteria, int, error) {
	var criteria *imap.SearchCriteria
	var t *imap.SearchCriteria
	var err error

parseLoop:
	for i < len(filterExpr) {
		c := filterExpr[i]

		switch {
		case unicode.IsSpace(c):
			i++

		case c == '!':
			criteria = &imap.SearchCriteria{}
			t, i, err = parseFilterPrimary(filterExpr, i+1)
			if err != nil {
				return nil, i, err
			}

			criteria.Not = append(criteria.Not, *t)
			return criteria, i, nil

		case c == '(':
			return parseFilterExpression(filterExpr, i+1)

		default:
			break parseLoop
		}
	}

	var (
		opFunc filterCmpFunc
		t1, t2 string
	)

	t1, i = parseFilterToken(filterExpr, i)
	if _, ok := flagTokens[strings.ToUpper(t1)]; ok {
		criteria = &imap.SearchCriteria{}
		return assignFlag(criteria, strings.ToUpper(t1)), i, nil
	}

	for i < len(filterExpr) {
		c := filterExpr[i]

		switch {
		case unicode.IsSpace(c):
			i++

		case c == '=' || c == '!':
			opFunc, i, err = parseFilterCmpOp(filterExpr, i+1, c)
			if err != nil {
				return nil, i, err
			}
			i++

			t2, i, err = parseFilterQuotedToken(filterExpr, i)
			if err != nil {
				return nil, i, err
			}
			criteria = &imap.SearchCriteria{}

			return opFunc(criteria, t1, t2), i, nil
		}
	}

	return criteria, i, nil
}

func parseFilterToken(filterExpr []rune, i int) (string, int) {
	var sb strings.Builder

	for i < len(filterExpr) {
		c := filterExpr[i]

		switch {
		case unicode.IsLetter(c) || c == '-':
			sb.WriteRune(c)
			i++

		default:
			return sb.String(), i
		}
	}

	return sb.String(), i
}

func parseFilterQuotedToken(filterExpr []rune, i int) (string, int, error) {
	var sb strings.Builder
	var startQuote rune

	for i < len(filterExpr) {
		c := filterExpr[i]

		switch {
		case startQuote == 0 && (c == '\'' || c == '"'):
			startQuote = c
			i++

		case startQuote == 0 && unicode.IsSpace(c):
			i++

		case startQuote == 0:
			return "", i, fmt.Errorf("expected starting quote but got '%c'", c)

		case startQuote != 0 && c != startQuote:
			sb.WriteRune(c)
			i++

		case startQuote != 0 && c == startQuote:
			return sb.String(), i + 1, nil
		}
	}

	if startQuote != 0 {
		return "", i, errors.New("missing closing quote")
	}

	return sb.String(), i, nil
}

func parseFilterBoolOp(filterExpr []rune, i int, opChar rune) (filterBoolFunc, int, error) {
	for i < len(filterExpr) {
		c := filterExpr[i]

		switch {
		case c == opChar && opChar == '|':
			return addOrCriteria, i + 1, nil

		case c == opChar && opChar == '&':
			return addAndCriteria, i + 1, nil

		default:
			return nil, i, fmt.Errorf("unexpected '%c' token while parsing '%c' bool function", c, opChar)
		}
	}

	return nil, i, errors.New("bool operation parsing stopped unexpectedly")
}

func parseFilterCmpOp(filterExpr []rune, i int, opChar rune) (filterCmpFunc, int, error) {
	for i < len(filterExpr) {
		c := filterExpr[i]

		switch {
		case opChar == '=' && c == '=':
			return addEqCmpCriteriaOp, i, nil

		case opChar == '!' && c == '=':
			return addNotEqCmpCriteriaOp, i, nil

		default:
			return nil, i, fmt.Errorf("unexpected token '%c'", c)
		}
	}

	return nil, i, errors.New("compare operation parsing stopped unexpectedly")
}

type filterCmpFunc func(*imap.SearchCriteria, string, string) *imap.SearchCriteria

func addEqCmpCriteriaOp(c *imap.SearchCriteria, k, v string) *imap.SearchCriteria {
	if _, ok := msgTokens[k]; ok {
		if k == "BODY" {
			c.Body = append(c.Body, v)
			return c
		}

		if k == "TEXT" {
			c.Text = append(c.Text, v)
			return c
		}
	}

	c.Header = append(c.Header, imap.SearchCriteriaHeaderField{
		Key:   k,
		Value: v,
	})
	return c
}

func addNotEqCmpCriteriaOp(c *imap.SearchCriteria, k, v string) *imap.SearchCriteria {
	c.Not = append(c.Not, *addEqCmpCriteriaOp(&imap.SearchCriteria{}, k, v))
	return c
}

var flagTokens = map[string]imap.Flag{
	"JUNK":       imap.FlagJunk,
	"SEEN":       imap.FlagSeen,
	"UNSEEN":     imap.FlagSeen,
	"DRAFT":      imap.FlagDraft,
	"UNDRAFT":    imap.FlagDraft,
	"DELETED":    imap.FlagDeleted,
	"UNDELETED":  imap.FlagDeleted,
	"FLAGGED":    imap.FlagFlagged,
	"UNFLAGGED":  imap.FlagFlagged,
	"PHISHING":   imap.FlagPhishing,
	"WILDCARD":   imap.FlagWildcard,
	"FORWARDED":  imap.FlagForwarded,
	"IMPORTANT":  imap.FlagImportant,
	"ANSWERED":   imap.FlagAnswered,
	"UNANSWERED": imap.FlagAnswered,
}

var msgTokens = map[string]struct{}{
	"TEXT": {},
	"BODY": {},
}

type filterBoolFunc func(c1, c2 *imap.SearchCriteria) *imap.SearchCriteria

func addAndCriteria(c1, c2 *imap.SearchCriteria) *imap.SearchCriteria {
	if c1 == nil {
		return c2
	}
	if c2 == nil {
		return c1
	}

	c1.And(c2)
	return c1
}

func addOrCriteria(c1, c2 *imap.SearchCriteria) *imap.SearchCriteria {
	return &imap.SearchCriteria{
		Or: [][2]imap.SearchCriteria{{*c1, *c2}},
	}
}

func addNotCriteria(c1 *imap.SearchCriteria) *imap.SearchCriteria {
	return &imap.SearchCriteria{
		Not: []imap.SearchCriteria{*c1},
	}
}

func intersectCriteria(c1, c2 *imap.SearchCriteria) *imap.SearchCriteria {
	return c1
}

func assignFlag(c *imap.SearchCriteria, flagToken string) *imap.SearchCriteria {
	flag, ok := flagTokens[flagToken]
	if !ok {
		return c
	}

	if _, ok := strings.CutPrefix(flagToken, "UN"); ok {
		c.NotFlag = append(c.NotFlag, flag)
		return c
	}

	c.Flag = append(c.Flag, flag)
	return c
}

func assignHeader(c *imap.SearchCriteria, k, v string) *imap.SearchCriteria {
	c.Header = append(c.Header, imap.SearchCriteriaHeaderField{
		Key:   k,
		Value: v,
	})
	return c
}
