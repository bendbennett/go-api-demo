package sanitise

import "regexp"

func AlphaWithHyphen(str string) (string, error) {
	reg := regexp.MustCompile("[^a-zA-Z-]+")

	return reg.ReplaceAllString(str, ""), nil
}
