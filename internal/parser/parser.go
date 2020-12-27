package parser

// Parser defines the interface a parser object should provide in order to
// manage the functions of parsing of a blockchain structure
type Parser interface {
	InfinitelyParse() error
	Parse() error
}
