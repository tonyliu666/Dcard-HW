package util

import(
	"fmt"
	"dcardapp/param"
)

func GenerateHash(query param.Query) string {
	// Concatenate struct fields into a string
	hash := fmt.Sprintf("%s:%s:%s:%s:%d:%d",
		query.Age, query.Country, query.Platform, query.Gender, query.Offset, query.Limit)
	return hash
}

