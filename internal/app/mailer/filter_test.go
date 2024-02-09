package main

import (
	"fmt"
	"testing"

	"github.com/emersion/go-imap/v2"
	"github.com/stretchr/testify/assert"
)

func TestParseFilterExpression(t *testing.T) {
	tests := []struct {
		filterExpr     []rune
		expectedOutput *imap.SearchCriteria
	}{
		{
			filterExpr: []rune("SEEN"),
			expectedOutput: &imap.SearchCriteria{
				Flag: []imap.Flag{imap.FlagSeen},
			},
		},
		{
			filterExpr: []rune("!SEEN"),
			expectedOutput: &imap.SearchCriteria{
				Not: []imap.SearchCriteria{{Flag: []imap.Flag{imap.FlagSeen}}},
			},
		},
		{
			filterExpr: []rune("SEEN && JUNK && IMPORTANT"),
			expectedOutput: &imap.SearchCriteria{
				Flag: []imap.Flag{imap.FlagSeen, imap.FlagJunk, imap.FlagImportant},
			},
		},
		{
			filterExpr: []rune("SEEN || JUNK || IMPORTANT"),
			expectedOutput: &imap.SearchCriteria{
				Or: [][2]imap.SearchCriteria{{
					{
						Or: [][2]imap.SearchCriteria{{
							{Flag: []imap.Flag{imap.FlagSeen}},
							{Flag: []imap.Flag{imap.FlagJunk}},
						}},
					},
					{Flag: []imap.Flag{imap.FlagImportant}},
				}},
			},
		},
		{
			filterExpr: []rune("(((SEEN)))"),
			expectedOutput: &imap.SearchCriteria{
				Flag: []imap.Flag{imap.FlagSeen},
			},
		},
		{
			filterExpr: []rune("FROM == 'test@test.com'"),
			expectedOutput: &imap.SearchCriteria{
				Header: []imap.SearchCriteriaHeaderField{{
					Key:   "FROM",
					Value: "test@test.com",
				}},
			},
		},
		{
			filterExpr: []rune("FROM == 'test@test.com' && SEEN"),
			expectedOutput: &imap.SearchCriteria{
				Header: []imap.SearchCriteriaHeaderField{{
					Key:   "FROM",
					Value: "test@test.com",
				}},
				Flag: []imap.Flag{imap.FlagSeen},
			},
		},
		{
			filterExpr: []rune("((IMPORTANT || FORWARDED) && !JUNK)"),
			expectedOutput: &imap.SearchCriteria{
				Not: []imap.SearchCriteria{{Flag: []imap.Flag{imap.FlagJunk}}},
				Or: [][2]imap.SearchCriteria{{
					{Flag: []imap.Flag{imap.FlagImportant}},
					{Flag: []imap.Flag{imap.FlagForwarded}},
				}},
			},
		},
		{
			filterExpr: []rune("!JUNK || FROM == 'very.important@contact.com'"),
			expectedOutput: &imap.SearchCriteria{
				Or: [][2]imap.SearchCriteria{{
					{Not: []imap.SearchCriteria{{Flag: []imap.Flag{imap.FlagJunk}}}},
					{Header: []imap.SearchCriteriaHeaderField{{
						Key:   "FROM",
						Value: "very.important@contact.com",
					}}},
				}},
			},
		},
		{
			filterExpr: []rune("!(!JUNK || FROM == 'very.important@contact.com')"),
			expectedOutput: &imap.SearchCriteria{
				Not: []imap.SearchCriteria{{
					Or: [][2]imap.SearchCriteria{{
						{Not: []imap.SearchCriteria{{Flag: []imap.Flag{imap.FlagJunk}}}},
						{Header: []imap.SearchCriteriaHeaderField{{
							Key:   "FROM",
							Value: "very.important@contact.com",
						}}},
					}},
				}},
			},
		},
		{
			filterExpr: []rune("UNSEEN && UNDELETED"),
			expectedOutput: &imap.SearchCriteria{
				NotFlag: []imap.Flag{imap.FlagSeen, imap.FlagDeleted},
			},
		},
		{
			filterExpr: []rune("    SEEN    && ( IMPORTANT )       "),
			expectedOutput: &imap.SearchCriteria{
				Flag: []imap.Flag{imap.FlagSeen, imap.FlagImportant},
			},
		},
		{
			filterExpr: []rune(" (   (UNSEEN ||  !(DELETED && FROM == 'test@test.com' )) )"),
			expectedOutput: &imap.SearchCriteria{
				Or: [][2]imap.SearchCriteria{{
					{NotFlag: []imap.Flag{imap.FlagSeen}},
					{
						Not: []imap.SearchCriteria{{
							Flag: []imap.Flag{imap.FlagDeleted},
							Header: []imap.SearchCriteriaHeaderField{{
								Key:   "FROM",
								Value: "test@test.com",
							}},
						}},
					},
				}},
			},
		},
		{
			filterExpr: []rune("BODY == 'sample body' && TEXT == 'sample text'"),
			expectedOutput: &imap.SearchCriteria{
				Body: []string{"sample body"},
				Text: []string{"sample text"},
			},
		},
	}

	for i, tt := range tests {
		t.Run(fmt.Sprintf("Case_%d", i), func(t *testing.T) {
			actual, _, err := parseFilterExpression(tt.filterExpr, 0)
			assert.NoError(t, err)
			assert.Equal(t, tt.expectedOutput, actual, "failed to parse %q", string(tt.filterExpr))
		})
	}
}
