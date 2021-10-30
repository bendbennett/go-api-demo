package search

const searchTermMinLen = 3

// alphaWithHyphen removes all characters from string except
// alpha and hyphen.
type alphaWithHyphen func(string) (string, error)
