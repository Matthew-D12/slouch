// Package token ...
package token

import "strconv"

type Kind uint8

const (
	Illegal Kind = iota
	EOF
	Newline
	Ident
	Int
	String
	Assign
	Equals
	NotEquals
	Less
	Greater
	LessEquals
	GreaterEquals
	DivisibleBy
	And
	Or
	Comma
	Dot
	Colon
	Plus
	Neg
	Star
	Slash
	Mod
	Splat
	Rep
	Snippet
	Lbrace
	Rbrace
	Lbracket
	Rbracket
	Lparen
	Rparen
	Pipe
	PipeSplat
	PipeRep
	Buck
	Hole
)

func (k Kind) String() string {
	return [...]string{
		Illegal:       "Illegal",
		EOF:           "EOF",
		Newline:       "Newline",
		Ident:         "Ident",
		Int:           "Int",
		String:        "String",
		Assign:        "Assign",
		Equals:        "Equals",
		NotEquals:     "NotEquals",
		Less:          "Less",
		Greater:       "Greater",
		LessEquals:    "LessEquals",
		GreaterEquals: "GreaterEquals",
		DivisibleBy:   "DivisibleBy",
		And:           "And",
		Or:            "Or",
		Comma:         "Comma",
		Dot:           "Dot",
		Colon:         "Colon",
		Plus:          "Plus",
		Neg:           "Neg",
		Star:          "Star",
		Slash:         "Slash",
		Mod:           "Mod",
		Splat:         "Splat",
		Rep:           "Rep",
		Snippet:       "Snippet",
		Lbrace:        "Lbrace",
		Rbrace:        "Rbrace",
		Lbracket:      "Lbracket",
		Rbracket:      "Rbracket",
		Lparen:        "Lparen",
		Rparen:        "Rparen",
		Pipe:          "Pipe",
		PipeSplat:     "PipeSplat",
		PipeRep:       "PipeRep",
		Buck:          "Buck",
		Hole:          "Hole",
	}[k]
}

type Token struct {
	Kind Kind
	Lit  string
}

func (t Token) String() string { return t.Kind.String() + " " + strconv.Quote(t.Lit) }
