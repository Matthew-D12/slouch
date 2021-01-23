// Package parser ...
package parser

import (
	"fmt"

	"lukechampine.com/slouch/ast"
	"lukechampine.com/slouch/token"
)

// + has LOWER precedence than *
//
// HIGHER precedence means binds "more tightly"

const (
	precLowest = iota
	precPipe
	precOr
	precAnd
	precNot
	precAssign
	precEquals
	precCmp
	precSum
	precProd
	precNeg
	precCall
	precSplat
	precDot
)

var precedences = map[token.Kind]int{
	token.Pipe:      precPipe,
	token.PipeSplat: precPipe,
	token.PipeRep:   precPipe,
	// token.OR:     OR,
	// token.AND:    AND,
	// token.NOT:    NOT,
	// token.BIND:   ASSIGN,
	// token.ASSIGN: ASSIGN,
	token.Equals:        precEquals,
	token.NotEquals:     precEquals,
	token.Less:          precCmp,
	token.Greater:       precCmp,
	token.LessEquals:    precCmp,
	token.GreaterEquals: precCmp,
	token.And:           precAnd,
	token.Or:            precOr,

	token.Plus:  precSum,
	token.Neg:   precSum,
	token.Star:  precProd,
	token.Slash: precProd,
	token.Splat: precSplat,
	token.Rep:   precSplat,
	token.Dot:   precDot,
	// token.MINUS:    SUM,
	// token.DIVIDE:   PRODUCT,
	// token.MULTIPLY: PRODUCT,
	// token.MODULO:   PRODUCT,
	// token.LPAREN:   CALL,
}

// Parse parses a sequence of tokens as a program.
func Parse(ts []token.Token) ast.Program {
	return newParser(ts).parseProgram()
}

// A Parser parses slouch programs.
type Parser struct {
	ts []token.Token

	prefixFns map[token.Kind]func() ast.Expr
	infixFns  map[token.Kind]func(ast.Expr) ast.Expr
}

func newParser(ts []token.Token) *Parser {
	p := &Parser{
		ts: ts,
	}
	p.prefixFns = map[token.Kind]func() ast.Expr{
		token.Ident:  p.parseIdentifier,
		token.Int:    p.parseInteger,
		token.String: p.parseString,
		//token.Bang:    p.parsePrefixOp,
		token.Lbrace:        p.parseArray,
		token.Lparen:        p.parseParens,
		token.Plus:          p.parsePrefixInfixOp,
		token.Star:          p.parsePrefixInfixOp,
		token.Slash:         p.parsePrefixInfixOp,
		token.Equals:        p.parsePrefixInfixOp,
		token.NotEquals:     p.parsePrefixInfixOp,
		token.Less:          p.parsePrefixInfixOp,
		token.Greater:       p.parsePrefixInfixOp,
		token.LessEquals:    p.parsePrefixInfixOp,
		token.GreaterEquals: p.parsePrefixInfixOp,
		token.And:           p.parsePrefixInfixOp,
		token.Or:            p.parsePrefixInfixOp,
		token.Dot:           p.parsePrefixInfixOp,
		token.Hole:          p.parseHole,
		token.Splat:         p.parseSplat,
		token.Rep:           p.parseRep,
	}
	p.infixFns = map[token.Kind]func(ast.Expr) ast.Expr{
		token.Pipe:          p.parsePipe,
		token.PipeSplat:     p.parsePipe,
		token.PipeRep:       p.parsePipe,
		token.Plus:          p.parseInfixOp,
		token.Neg:           p.parseInfixOp,
		token.Star:          p.parseInfixOp,
		token.Slash:         p.parseInfixOp,
		token.Equals:        p.parseInfixOp,
		token.NotEquals:     p.parseInfixOp,
		token.Less:          p.parseInfixOp,
		token.Greater:       p.parseInfixOp,
		token.LessEquals:    p.parseInfixOp,
		token.GreaterEquals: p.parseInfixOp,
		token.And:           p.parseInfixOp,
		token.Or:            p.parseInfixOp,
		token.Dot:           p.parseInfixOp,
	}
	return p
}

func (p *Parser) cur() token.Token {
	if len(p.ts) == 0 {
		return token.Token{Kind: token.EOF}
	}
	return p.ts[0]
}

func (p *Parser) peek() token.Token {
	if len(p.ts) < 2 {
		return token.Token{Kind: token.EOF}
	}
	return p.ts[1]
}

func (p *Parser) curIs(kinds ...token.Kind) bool {
	t := p.cur()
	for _, k := range kinds {
		if t.Kind == k {
			return true
		}
	}
	return false
}

func (p *Parser) peekIs(kinds ...token.Kind) bool {
	t := p.peek()
	for _, k := range kinds {
		if t.Kind == k {
			return true
		}
	}
	return false
}

func (p *Parser) advance() {
	if len(p.ts) > 0 {
		p.ts = p.ts[1:]
	}
}

func (p *Parser) expect(ks ...token.Kind) {
	if len(p.ts) == 0 {
		panic(fmt.Sprintf("expected one of %v, got EOF", ks))
	}
	for _, k := range ks {
		if p.cur().Kind == k {
			return
		}
	}
	panic(fmt.Sprintf("expected one of %v, got %v", ks, p.cur()))
}

func (p *Parser) consume(k token.Kind) token.Token {
	p.expect(k)
	t := p.cur()
	p.advance()
	return t
}

func (p *Parser) consumeWhitespace() {
	for p.cur().Kind == token.Newline {
		p.advance()
	}
}

func (p *Parser) parseProgram() ast.Program {
	var prog ast.Program
	for p.cur().Kind != token.EOF {
		prog.Stmts = append(prog.Stmts, p.parseStmt())
	}
	return prog
}

func (p *Parser) parseStmt() (s ast.Stmt) {
	p.consumeWhitespace()
	switch t := p.cur(); t.Kind {
	case token.Assign:
		s = p.parseAssignStmt()
	case token.Snippet:
		s = ast.SnippetStmt{Token: t, Code: t.Lit}
	default:
		// any other token indicates an expression statement
		s = ast.ExprStmt{X: p.parseExpr(precLowest)}
	}
	//fmt.Println(ast.Print(s))
	// all statements should end in a newline
	p.advance()
	p.consume(token.Newline)
	return
}

func (p *Parser) parseAssignStmt() ast.Stmt {
	p.consume(token.Assign)
	name := p.consume(token.Ident)
	return ast.AssignStmt{
		Name: ast.Ident{Token: name, Name: name.Lit},
		X:    p.parseExpr(precLowest),
	}
}

func (p *Parser) parseExpr(prec int) ast.Expr {
	t := p.cur()
	pfn, ok := p.prefixFns[t.Kind]
	if !ok {
		panic(fmt.Sprintf("no prefix for %v", t.String()))
	}
	e := pfn()

	// if there are args, this is a fn call
	//
	// TODO: can't shake the feeling that this is buggy...
	if prec < precCall {
		var args []ast.Expr
		for p.peekIs(token.Hole, token.Ident, token.Int, token.String, token.Lbrace, token.Lparen, token.Splat, token.Rep) {
			p.advance()
			args = append(args, p.parseExpr(precCall))
		}
		if len(args) > 0 {
			e = ast.FnCall{
				Token: t,
				Fn:    e,
				Args:  args,
			}
		}
	}

	for !p.peekIs(token.Newline, token.EOF, token.Illegal) && prec < precedences[p.peek().Kind] {
		ifn, ok := p.infixFns[p.peek().Kind]
		if !ok {
			break
		}
		p.advance()
		e = ifn(e)
	}

	return e
}

func (p *Parser) parseInteger() ast.Expr    { return ast.Integer{Token: p.cur(), Value: p.cur().Lit} }
func (p *Parser) parseString() ast.Expr     { return ast.String{Token: p.cur(), Value: p.cur().Lit} }
func (p *Parser) parseIdentifier() ast.Expr { return ast.Ident{Token: p.cur(), Name: p.cur().Lit} }
func (p *Parser) parseHole() ast.Expr       { return ast.Hole{Token: p.cur()} }

func (p *Parser) parseSplat() ast.Expr {
	t := p.cur()
	p.advance()
	return ast.Splat{Token: t, Fn: p.parseExpr(precSplat)}
}

func (p *Parser) parseRep() ast.Expr {
	t := p.cur()
	p.advance()
	return ast.Rep{Token: t, Fn: p.parseExpr(precSplat)}
}

func (p *Parser) parseArray() ast.Expr {
	t := p.cur()
	p.advance() // advance past [
	var elems []ast.Expr
	var assoc bool
	var danglingKey bool
	for !p.curIs(token.Rbrace, token.EOF) {
		elems = append(elems, p.parseExpr(precLowest))
		p.advance()
		if p.curIs(token.Colon) {
			assoc = true
		} else {
			danglingKey = danglingKey || len(elems)%2 != 0
		}
		if !p.curIs(token.Comma, token.Colon) {
			break
		}
		p.advance()
	}
	p.expect(token.Rbrace)
	if assoc && danglingKey {
		panic("dangling key in assoc")
	}
	return ast.Array{Token: t, Elems: elems, Assoc: assoc}
}

func (p *Parser) parseParens() ast.Expr {
	p.advance() // advance past (
	e := p.parseExpr(precLowest)
	p.advance() // advance to )
	return e
}

/*
func (p *Parser) parsePrefixOp() ast.Expr {
	t := p.cur()
	e := ast.PrefixOp{
		Token:    t,
		Operator: t.Kind,
	}
	p.advance()
	e.Right = p.parseExpr(precPrefix)
	return e
}


*/

func (p *Parser) parsePrefixInfixOp() ast.Expr {
	t := p.cur()
	e := ast.InfixOp{
		Token: t,
		Op:    t.Kind,
		Left:  nil,
	}
	if !p.peekIs(token.Illegal, token.EOF, token.Newline, token.Rparen, token.Pipe) { // TODO: kind of a hack...
		p.advance()
		e.Right = p.parseExpr(precedences[t.Kind]) // TODO: wrong?
	}
	return e
}

func (p *Parser) parseInfixOp(left ast.Expr) ast.Expr {
	t := p.cur()
	e := ast.InfixOp{
		Token: t, // TODO: wrong?
		Op:    t.Kind,
		Left:  left,
	}
	p.advance()
	e.Right = p.parseExpr(precedences[t.Kind])
	return e
}

func (p *Parser) parsePipe(left ast.Expr) ast.Expr {
	t := p.cur()
	e := ast.Pipe{
		Token: t,
		Left:  left,
	}
	p.advance()
	e.Right = p.parseExpr(precPipe)
	return e
}
