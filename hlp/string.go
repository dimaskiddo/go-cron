package hlp

import (
	"strings"
)

func IsStringsContains(s string, str []string) bool {
	if len(str) > 0 {
		for _, v := range str {
			if v == s {
				return true
			}
		}
	}

	return false
}

func SplitWithEscapeN(s string, sep string, n int, trim bool) []string {
	ret := []string{""}

	if trim {
		s = strings.TrimSpace(s)
	}

	if len(s) > 0 {
		var j int

		split := strings.SplitN(s, sep, n)
		for i := 0; i < len(split); i++ {
			if trim {
				ret[j] = strings.TrimSpace(split[i])
			} else {
				ret[j] = split[i]
			}

			for (strings.Count(ret[j], "'") == 1 || strings.Count(ret[j], "\"") == 1) && i != len(split)-1 {
				if trim {
					ret[j] = ret[j] + sep + strings.TrimSpace(split[i+1])
				} else {
					ret[j] = ret[j] + sep + split[i+1]
				}
				i++
			}

			if i != len(split)-1 {
				ret = append(ret, "")
				j++
			}
		}
	}

	return ret
}
