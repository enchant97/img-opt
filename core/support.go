package core

import "strings"

type NonStandardSupport struct {
	WEBP bool
	AVIF bool
}

func NonStandardFromAcceptHeader(headerContent string) NonStandardSupport {
	support := NonStandardSupport{}

	types := strings.Split(headerContent, ",")
	for _, fullType := range types {
		fullType := strings.SplitN(fullType, ";", 2)[0]
		types := strings.SplitN(fullType, "/", 2)
		if len(types) < 2 {
			continue
		}
		subType := types[1]

		if subType == "webp" {
			support.WEBP = true
		} else if subType == "avif" {
			support.AVIF = true
		}
	}
	return support
}
