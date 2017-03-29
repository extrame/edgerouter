package math

import "strconv"

// Parse creates a new parser with the recommended
// parameters.
func Parse(tokens []LexToken) Item {
	p := &parser{
		tokens: tokens,
		pos:    -1,
	}
	p.initState = initialParserState
	return p.run()
}

// run starts the statemachine
func (p *parser) run() Item {
	var item Item
	for state := p.initState; state != nil; {
		item, state = state(p)
	}
	return item
}

// parserState represents the state of the scanner
// as a function that returns the next state.
type parserState func(*parser) (Item, parserState)

// nest returns what the next token AND
// advances p.pos.
func (p *parser) next() *LexToken {
	if p.pos >= len(p.tokens)-1 {
		return nil
	}
	p.pos += 1
	return &p.tokens[p.pos]
}

// the parser type
type parser struct {
	tokens []LexToken
	pos    int
	serial int

	initState parserState
}

// the starting state for parsing
func initialParserState(p *parser) (Item, parserState) {
	var root, current Item
	for t := p.next(); t[0] != T_EOF; t = p.next() {
		switch t[0] {
		case T_LCLOSURE_MARK:
			sub, _ := initialParserState(p)
			if current != nil {
				current.addItem(sub)
			} else {
				current = sub
			}
			if root == nil {
				root = sub
			}
		case T_RCLOSURE_MARK:
			root.BeClosure()
			return root, initialParserState
		case T_NUMBER_MARK:
			val, _ := strconv.Atoi(t[1])
			numItem := &NumberItem{val}
			if current != nil {
				current.addItem(numItem)
			} else {
				current = numItem
			}
			if root == nil {
				root = current
			}
		case T_VAR_MARK:
			val, _ := strconv.Atoi(t[1][1:])
			varItem := &VarItem{val}
			if current != nil {
				current.addItem(varItem)
			} else {
				current = varItem
			}
			if root == nil {
				root = current
			}
		case T_OPER_MARK:
			root, current = root.addOperator(Operator(t[1]))
		}
	}
	return root, nil
}
