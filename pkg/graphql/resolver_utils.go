package graphql

import (
	"encoding/base64"
	"fmt"
	log "github.com/golang/glog"
	"strconv"
	"strings"
)

const (
	cursorInternalDelimiter = "|||"

	// Represents a cursor that uses row offset as the position
	cursorTypeOffset = "offset"

	defaultCursorType  = cursorTypeOffset
	defaultCursorValue = "0"
)

var defaultPaginationCursor = &paginationCursor{
	typeName: defaultCursorType,
	value:    defaultCursorValue,
}

func decodeToPaginationCursor(s string) (*paginationCursor, error) {
	bys, err := base64.StdEncoding.DecodeString(s)
	if err != nil {
		return nil, err
	}
	parts := strings.Split(string(bys), cursorInternalDelimiter)
	if len(parts) != 2 {
		return nil, fmt.Errorf("Invalid cursor decode: %v", parts)
	}
	return &paginationCursor{
		typeName: parts[0],
		value:    parts[1],
	}, nil
}

type paginationCursor struct {
	// The name of the cursor type
	typeName string
	// The target cursor value from which to continue pagination
	value string
}

func (c *paginationCursor) ValueInt() int {
	v, err := strconv.Atoi(c.value)
	if err != nil {
		log.Errorf("Error converting pagination value: err: %v", err)
		return 0
	}
	return v
}

func (c *paginationCursor) ValueFromInt(v int) {
	c.value = fmt.Sprintf("%v", v)
}

func (c *paginationCursor) Encode() string {
	baseStr := fmt.Sprintf(
		"%v%v%v",
		c.typeName,
		cursorInternalDelimiter,
		c.value,
	)
	return base64.StdEncoding.EncodeToString([]byte(baseStr))
}

func paginationOffsetFromCursor(cursor *paginationCursor,
	after *string) (int, *paginationCursor, error) {
	afterCursor, err := decodeToPaginationCursor(*after)
	if err != nil {
		return 0, nil, err
	}

	startOffset := 0

	if afterCursor.typeName == cursorTypeOffset {
		cursorIntValue := afterCursor.ValueInt()
		// Increment the offset and get the next item
		startOffset = cursorIntValue + 1
		afterCursor.ValueFromInt(cursorIntValue + 1)
		cursor = afterCursor
	}

	return startOffset, cursor, nil
}

func criteriaCount(first *int) int {
	// Default count value
	criteriaCount := defaultCriteriaCount
	if first != nil {
		criteriaCount = *first
	}

	// Add 1 to all of these to see if there are additional items
	// If we see items beyond what we truly requested, then that warrants
	// another query by the caller.
	criteriaCount++
	return criteriaCount
}
