package helper

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/pkg/errors"
	"gopx.io/gopx-common/arr"
	"gopx.io/gopx-common/str"
)

// DecodeQueryValue decodes the input query values used in the
// user and package search.
func DecodeQueryValue(qv string) string {
	return strings.Replace(qv, "+", " ", -1)
}

// ParseRelationalValues parses the relational values used in search query.
// Examples:
// Normal: packages:<=1000
// 				 packages:1000
// Range: joined:2016-04-30..2016-07-04
// 				joined:2016-04-30..*
//				joined:*..2016-07-04
func ParseRelationalValues(val string) (isRange bool, ranges []string, operatorValue []string, err error) {
	val = strings.TrimSpace(val)

	reNormal, err := regexp.Compile("^((?:>=)|(?:>)|(?:<=)|(?:<))(.+)$")
	if err != nil {
		err = errors.Wrap(err, "Failed to compile the regex")
		return
	}
	reRange, err := regexp.Compile("^([^\\.]+)\\.\\.([^\\.]+)$")
	if err != nil {
		err = errors.Wrap(err, "Failed to compile the regex")
		return
	}

	if matches := reNormal.FindStringSubmatch(val); matches != nil && len(matches) > 0 {
		operatorValue = []string{matches[1], matches[2]}
		isRange = false
		return
	} else if matches = reRange.FindStringSubmatch(val); matches != nil && len(matches) > 0 {
		ranges = []string{matches[1], matches[2]}
		isRange = true
		return
	} else {
		operatorValue = []string{"=", val}
		isRange = false
		return
	}
}

// PrepareMultiWordsQueryClause creates SQL where-clauses based on
// the multi words query values.
func PrepareMultiWordsQueryClause(clauses *[]string, placeholderValues *[]interface{}, column string, value string) {
	for _, qv := range str.SplitSpace(value) {
		*clauses = append(*clauses, fmt.Sprintf("%s like ?", column))
		*placeholderValues = append(*placeholderValues, fmt.Sprintf("%%%s%%", qv))
	}
}

// PrepareRelationalQueryClause creates SQL where-clauses based on
// the relational query values.
func PrepareRelationalQueryClause(placeholderValues *[]interface{}, column, value string) (clause string, err error) {
	if isRange, ranges, operatorValue, err := ParseRelationalValues(value); err != nil {
		err = errors.Wrap(err, "Value parsing failed")
		return "", err
	} else if isRange {
		rVal1, rVal2 := ranges[0], ranges[1]
		if rVal1 == "*" && rVal2 != "*" {
			clause = fmt.Sprintf("(%s <= ?)", column)
			*placeholderValues = append(*placeholderValues, rVal2)
		} else if rVal1 != "*" && rVal2 == "*" {
			clause = fmt.Sprintf("(%s >= ?)", column)
			*placeholderValues = append(*placeholderValues, rVal1)
		} else if rVal1 != "*" && rVal2 != "*" {
			clause = fmt.Sprintf("(%s >= ? and %s <= ?)", column, column)
			*placeholderValues = append(*placeholderValues, rVal1, rVal2)
		}
	} else {
		opr, val := operatorValue[0], operatorValue[1]
		clause = fmt.Sprintf("(%s %s ?)", column, opr)
		*placeholderValues = append(*placeholderValues, val)
	}

	return
}

// PrepareInQualifierClause prepares SQL where-clauses based on
// the search term and 'In' qualifier.
func PrepareInQualifierClause(placeholderValues *[]interface{}, inColumns []string, searchTerm string, inValue string) (clause string) {
	inClauses := []string{}
	ins := strings.Split(inValue, ",")

	for _, in := range ins {
		if str.IsEmpty(in) || arr.FindStr(inColumns, in) == -1 {
			continue
		}

		PrepareMultiWordsQueryClause(&inClauses, placeholderValues, in, searchTerm)
	}

	if len(inClauses) > 0 {
		clause = "(" + strings.Join(inClauses, " or ") + ")"
	} else {
		clause = ""
	}

	return
}

// SanitizeSortByCols sanitizes sortBy values according to
// the specified column list.
func SanitizeSortByCols(sortBy string, cols []string) []string {
	sCols := []string{}
	for _, sb := range strings.Split(sortBy, ",") {
		sb = strings.ToLower(strings.TrimSpace(sb))

		if str.IsEmpty(sb) || arr.FindStr(cols, sb) == -1 {
			continue
		}

		sCols = append(sCols, sb)
	}

	return sCols
}
