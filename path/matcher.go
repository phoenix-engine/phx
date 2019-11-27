package path

// Matcher matches on some string path.
type Matcher interface{ Match(string) bool }
