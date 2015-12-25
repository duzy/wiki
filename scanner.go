//
//  Copyright (C) 2013, Duzy Chan <code@duzy.info>, all rights reserverd.
//
package wiki

import "fmt"

type SyntaxError struct {
	msg    string // description of error
	Offset int64  // error occurred after reading Offset bytes
}

func (e *SyntaxError) Error() string { return e.msg }

type scanner struct {
	// json/scanner.go:
	//     The step is a func to be called to execute the next transition.
	//     Also tried using an integer constant and a single func
	//     with a switch, but using the func directly was 10% faster
	//     on a 64-bit Mac Mini, and it's nicer to read.
	step func(*scanner, int) int // = states[stateTop]
	states []func(*scanner, int) int
	stateTop int // = len(states) - 1

	// Stack of what we're in the middle of other entity.
	parsing []EntityType
	parsingPos []int
	parsingTopState EntityType	// = parsing[parsingTop]
	parsingTop int				// = len(parsing) - 1

	state EntityType	// parsed entity type
	shift [2]int		// the special chars lengths ({{, }})

	indent int // counting indent of newline '\n'

	// extra left offset for newline entities (==, *, #, etc)
	newlineOffset int

	// n-bytes rewind on each failed stepping
	rewind int

	err error

	pos func() int
	push func(state EntityType)
	pop func(state EntityType, pos1, pos2, off1, off2 int)
}

func (s *scanner) reset() {
	s.step = stateBeginWiki
	s.states = s.states[0:0]
	s.stateTop = -1
	s.parsing = s.parsing[0:0]
	s.parsingPos = s.parsingPos[0:0]
	s.parsingTopState, s.parsingTop = parseUnknown, -1
	s.indent, s.newlineOffset = 0, 0
	s.rewind = 0
	s.err = nil
}

// pushParseState push the current state, must call this before changing step
func (s *scanner) pushParseState(state EntityType) {
	//fmt.Printf("pushParseState: %v %v, pos=%d\n", s.parsing, state, s.pos())
	if s.push != nil { s.push(state) }
	s.parsingPos = append(s.parsingPos, s.pos())
	s.parsing = append(s.parsing, state); s.parsingTop++
	s.parsingTopState = state
}
func (s *scanner) pushStepState(step func(*scanner, int) int) {
	s.states = append(s.states, step); s.stateTop++
	s.step = step
}

// popParseState pops a parse state (already obtained) off the stack
// and updates s.step accordingly.
func (s *scanner) popParseState(off1, off2, offset, rewind int) {
	//fmt.Printf("popParseState: %v, %v, %d, pos=%d\n", s.parsing, s.parsingPos, s.parsingTop, s.pos())
	if s.pop != nil {
		pos1 := s.parsingPos[s.parsingTop] + offset
		pos2 := s.pos() - rewind
		s.pop(s.parsing[s.parsingTop], pos1, pos2, off1, off2)
	}
	s.parsingPos = s.parsingPos[0:s.parsingTop]
	s.parsing = s.parsing[0:s.parsingTop]
	if s.parsingTop--; s.parsingTop < 0 {
		s.parsingTopState = parseUnknown
	} else {
		s.parsingTopState = s.parsing[s.parsingTop]
	}
}
func (s *scanner) popStepState() {
	if 0 <= s.stateTop {
		s.states = s.states[0:s.stateTop]; s.stateTop--
		if s.stateTop < 0 {
			s.step = stateUnknown
		} else {
			s.step = s.states[s.stateTop]
		}
	} else {
		s.step = stateUnknown
		// TODO: ...
	}
}

// next splits data after the next Entity.
func (s *scanner) next(data []byte) (entity, rest []byte, err error) {
	i, end := 0, len(data)
	s.pos = func() int { return i }
	s.reset()
	for ; i < end; i++ {
		s.rewind = 0 // need to reset 'rewind' every step
		c := data[i]
		v := s.step(s, int(c))
		i -= s.rewind
		//fmt.Printf("%d:%v: %v\n", i, s.parsing, string(data[i:]))
		if scanEnd <= v {
			switch v {
			case scanError:
				return nil, nil, s.err
			case scanEnd:
				//fmt.Printf("%d:%v: %v\n", i, s.parsing, string(data[0:i]))
				return data[0:i], data[i:], nil
			}
		}
	}

	// do a step to zero (EOF)
	s.step(s, 0)

	if 0 <= s.parsingTop {
		s.end(0, 0, 0, 0, 0) // do a final end to pop as text
		if s.state != parseEntityText && len(data) == 1 && data[0] == '\n' {
			s.state = parseEntityText
		}
	}

	//if s.pop != nil { s.pop(s.parsing[s.parsingTop]) }
	//fmt.Printf("next:%d:%v: %v %v %v\n", i, s.parsing, s.state, s.shift, string(data))
	return data, nil, nil
}

func (s *scanner) checkSpecial(c int) bool {
	//fmt.Printf("checkSpecial: %v %v %v, indent=%v\n", string(c), s.parsing, s.parsingTopState, s.indent)

	switch c {
	case '\n':
		s.step, s.indent = stateNewline, 0
		return true
	case '\'':
		s.step = stateQuoteLeft1
		return true
	case '[':
		s.step = stateSqL1
		return true
	case ']':
		s.step = stateSqR1
		return true
	case '{':
		s.step = stateBrL1
		return true
	case '}':
		s.step = stateBrR1
		return true
	case '<':
		s.step = stateLt
		return true
	}

	if /*s.parsingTop == 0*/ s.parsingTopState == parseUnknown {
		switch c {
		case ' ', '\t':
			s.indent++
			return false
		case '*':
			s.step, s.newlineOffset = stateBeginListBulleted, 0
			return true
		case '#':
			s.step, s.newlineOffset = stateBeginListNumbered, 0
			return true
		case ':':
			s.step, s.newlineOffset = stateBeginIndent, 0
			return true
		case '-':
			s.step, s.newlineOffset = stateNewlineDash1, 0
			return true
		case '=':
			s.step, s.newlineOffset = stateNewlineEqualL1, 0
			return true
		}
	}
	return false
}

func (s *scanner) begin(step func(s *scanner, c int) int, state EntityType, scanCode, c, rewind int) int {
	if state != parseEntityText && s.parsingTopState == parseEntityText {
		//fmt.Printf("begin: %v %v %v end text, indent=%v\n", s.pos(), string(c), s.parsing, s.indent)
		s.rewind = rewind
		return s.end(c, 0, 0, 0, 0)
	}

	//fmt.Printf("begin: %v %v %v, indent=%v\n", s.pos(), string(c), s.parsing, s.indent)

	s.pushParseState(state)
	s.pushStepState(step)
	if s.checkSpecial(c) {
		// ...
	}
	return scanCode
}

func (s *scanner) end(c, off1, off2, offset, rewind int) int {
	//fmt.Printf("end: %v %v %v end, indent=%v\n", s.pos(), string(c), s.parsing, s.indent)
	if s.parsingTop < 0 /*|| s.stateTop < 0 || c == 0*/ {
		//fmt.Printf("end: %v %v %v\n", s.pos(), string(c), s.parsing)
		return scanEnd
	}

	s.state, s.shift = s.parsingTopState, [2]int{off1, off2}
	s.popParseState(off1, off2, offset, rewind)
	s.popStepState()

	if /*s.parsingTop < 0 ||*/ s.stateTop < 0 || c == 0 {
		//fmt.Printf("end: %v %v %v\n", s.pos(), string(c), s.parsing)
		return scanEnd
	}

	return s.states[s.stateTop](s, c)
}

// http://en.wiktionary.org/wiki/Help:Wikitext_quick_reference
const (
	scanContinue = iota
	scanBeginWiki
	scanBeginEntity

	scanBeginText
	scanBeginTextBold			// '''bold'''
	scanBeginTextItalic			// ''italic''
	scanBeginTextBoldItalic		// '''''bold italic''''', '''''bold italic'' bold''', '''''bold italic''' italic''

	scanBeginLink1				// [http://www.example.org Link label], [http://www.example.org]
	scanBeginLink2				// [[Page Title|Link label]], [[Page Title]]

	scanBeginListBulleted		// * List item
	scanBeginListNumbered		// # List item

	scanBeginIndent				// :Indented text, ::Indented text

	scanBeginHeader2			// == header 2 ==
	scanBeginHeader3			// === header 3 ===
	scanBeginHeader4			// ==== header 4 ====
	scanBeginHeader5			// ===== header 5 =====

	scanBeginTemplate			// {{object}}, {{object|prop}}
	scanBeginTemplateName		// name
	scanBeginTemplateProp		// |prop

	scanBeginTag				// <ref name="test" />
	scanBeginTagBeg				// <ref name="test">
	scanBeginTagProp			// name="test"
	scanBeginTagEnd				// </ref>

	scanBeginSignature			// ~~~
	scanBeginSignatureStamp		// ~~~~

	scanBeginHR					// ----

	// Don't put symboles after this!
	scanEnd
	scanError
)
const (
	parseUnknown				= WikiEntityWiki
	parseEntityText				= WikiEntityText				// top level normal text
	parseEntityTextItalic		= WikiEntityTextItalic			// ''
	parseEntityTextBold			= WikiEntityTextBold			// '''
	parseEntityTextBoldItalic	= WikiEntityTextBoldItalic		// '''''
	parseEntityLink1			= WikiEntityLinkExternal		// external link: [http://example.com Label], [http://example.org]
	parseEntityLink2			= WikiEntityLinkInternal		// internal link: [[Page Title|Link Label]], [[Page Title]]
	parseEntityLink2Name		= WikiEntityLinkInternalName	// |Link Label, |Wiktionary
	parseEntityLink2Prop		= WikiEntityLinkInternalProp	// |Link Label, |Wiktionary
	parseEntityListBulleted		= WikiEntityListBulleted		// * List item
	parseEntityListNumbered		= WikiEntityListNumbered		// # List item
	parseEntityIndent			= WikiEntityIndent				// :Indented text
	parseEntityHeader2			= WikiEntityHeading2			// == header 2 ==
	parseEntityHeader3			= WikiEntityHeading3			// === header 3 ===
	parseEntityHeader4			= WikiEntityHeading4			// ==== header 4 ====
	parseEntityHeader5			= WikiEntityHeading5			// ===== header 5 =====
	parseEntityTemplate			= WikiEntityTemplate			// {{object}}
	parseEntityTemplateName		= WikiEntityTemplateName		// name
	parseEntityTemplateProp		= WikiEntityTemplateProp		// |prop
	parseEntityTag				= WikiEntityTag					// <ref name="test" />
	parseEntityTagBeg			= WikiEntityTagBeg				// <ref name="test">
	parseEntityTagProp			= WikiEntityTagProp				// name="test"
	parseEntityTagEnd			= WikiEntityTagEnd				// </ref>
	parseEntitySignature		= WikiEntitySignature			// ~~~
	parseEntitySignatureStamp	= WikiEntitySignatureTimestamp	// ~~~~
	parseEntityHR				= WikiEntityHR					// ----
)

func stateUnknown(s *scanner, c int) int {
	s.step = stateError
	s.err = &SyntaxError{
		fmt.Sprintf("invalid character %s at unknown state", string(c)), 0,
	}
	return scanError
}

func stateBeginWiki(s *scanner, c int) int {
	if s.checkSpecial(c) {
		return scanContinue
	}
	return s.begin(stateInEntityText, parseEntityText, scanBeginText, c, 0)
}

// '
func stateQuoteLeft1(s *scanner, c int) int {
	//fmt.Printf("stateQuoteLeft1: %v %v %v\n", s.pos(), string(c), s.parsing)

	if c == '\'' {
		if s.parsingTopState == parseEntityText {
			//fmt.Printf("stateQuoteLeft1: %v %v end text\n", string(c), s.parsing)
			s.rewind = 1;
			return s.end(c, 0, 0, 0, 0)
		}

		s.step = stateQuoteLeft2
		return scanContinue
	}

	if s.stateTop < 0 {
		//fmt.Printf("stateQuoteLeftl: %v %v %v\n", s.pos(), string(c), s.parsing)
		return s.begin(stateInEntityText, parseEntityText, scanBeginText, c, 0)
	}

	s.step = s.states[s.stateTop]
	return s.step(s, c)
}

// ''
func stateQuoteLeft2(s *scanner, c int) int {
	//fmt.Printf("stateQuoteLeft2: %v %v %v\n", s.pos(), string(c), s.parsing)

	if c == '\'' {
		s.step = stateQuoteLeft3
		return scanContinue
	}
	return s.begin(stateInEntityTextItalic, parseEntityTextItalic, scanBeginTextItalic, c, 2)
}

// '''
func stateQuoteLeft3(s *scanner, c int) int {
	//fmt.Printf("stateQuoteLeft3: %v %v %v\n", s.pos(), string(c), s.parsing)
	if c == '\'' {
		s.step = stateQuoteLeft4
		return scanContinue
	}
	return s.begin(stateInEntityTextBold, parseEntityTextBold, scanBeginTextBold, c, 3)
}

// ''''
func stateQuoteLeft4(s *scanner, c int) int {
	//fmt.Printf("stateQuoteLeft4: %v %v %v\n", s.pos(), string(c), s.parsing)
	if c == '\'' {
		s.step = stateQuoteLeft5
		return scanContinue
	}
	// same as stateQuoteLeft3 in this case
	return s.begin(stateInEntityTextBold, parseEntityTextBold, scanBeginTextBold, c, 3)
}

// '''''
func stateQuoteLeft5(s *scanner, c int) int {
	//fmt.Printf("stateQuoteLeft5: %v %v %v\n", s.pos(), string(c), s.parsing)
	// ''a'''''A'''
	return s.begin(stateInEntityTextBoldItalic, parseEntityTextBoldItalic, scanBeginTextBoldItalic, c, 5)
}

// ' (right)
func stateQuoteRight1(s *scanner, c int) int {
	//fmt.Printf("stateQuoteRight1: %v %v %v\n", s.pos(), string(c), s.parsing)

	if c == '\'' {
		s.step = stateQuoteRight2
		return scanContinue
	}

	s.step = s.states[s.stateTop]
	return s.step(s, c) //return scanContinue
}

// '' (right)
func stateQuoteRight2(s *scanner, c int) int {
	//fmt.Printf("stateQuoteRight2: %v %v %v\n", s.pos(), string(c), s.parsing)

	if c == '\'' {
		if s.parsingTopState == parseEntityTextItalic {
			if 0 < s.parsingTop && s.parsing[s.parsingTop-1] == parseEntityTextBold {
				s.end(c, 2, 2, 0, 0)
				s.step = stateQuoteRight1
			} else /* s.parsingTop <= 0 */ {
				// in cases like:
				//   ''italic '''bold''' italic...''
				//   ''a'''''A'''
				s.step = stateQuoteLeft3
			}
		} else {
			s.step = stateQuoteRight3
		}
		return scanContinue
	}

	switch s.parsingTopState {
	case parseEntityTextItalic:
		//fmt.Printf("stateQuoteRight2: end: %v %v %v\n", s.pos(), string(c), s.parsing)
		return s.end(c, 2, 2, 0, 0)
	case parseEntityTextBoldItalic:
		topPos, off1, off2 := s.parsingPos[s.parsingTop], 2, 2
		if 0 < s.parsingTop && s.parsing[s.parsingTop-1] == parseEntityTextItalic {
			// '''a'''''b''
			// discard BoldItalic (rewinds), go back to Bold
			s.parsingPos = s.parsingPos[0:s.parsingTop]
			s.parsing = s.parsing[0:s.parsingTop]; s.parsingTop--
			s.states = s.states[0:s.stateTop]; s.stateTop--
			s.parsingTopState = s.parsing[s.parsingTop]
			s.rewind = s.pos() - topPos + off1 + 1
			if s.pop != nil {
				s.pop(parseEntityTextBoldItalic, 0, 0, 0, 0)
			}
			return s.end(c, off1-1, off2-1, 0, s.rewind-1)
		} else {
			// '''''a'''b''
			// convert BoldItalic into Bold
			s.parsingPos[s.parsingTop] = topPos - 2
			s.parsing[s.parsingTop] = parseEntityTextBold
			s.states[s.stateTop] = stateInEntityTextBold
			s.parsingTopState = parseEntityTextBold
			//if s.pop != nil { s.pop(s.parsing[s.parsingTop], 2, 2) }

			// push new Italic
			s.parsingPos = append(s.parsingPos, topPos)
			s.parsing = append(s.parsing, parseEntityTextItalic); s.parsingTop++
			s.parsingTopState = parseEntityTextItalic
			s.states = append(s.states, stateInEntityTextItalic); s.stateTop++
			s.step = stateInEntityTextItalic
		}
		return s.end(c, off1, off2, 0, 0)
	}

	//fmt.Printf("stateQuoteRight2: end: %v %v %v\n", s.pos(), string(c), s.parsing)

	// in cases like:
	//   '''bold ''italic'' bold...
	//   '''''Any''' may apply.''
	s.step = stateQuoteLeft1
	return stateQuoteLeft2(s, c)
}

func endStateQR3(s *scanner, c, off1, off2 int) int {
	if s.parsingTopState == parseEntityTextBoldItalic {
		topPos := s.parsingPos[s.parsingTop]

		//fmt.Printf("endStateQR3: %v %v %v\n", s.pos(), string(c), s.parsing)
		if 0 < s.parsingTop && s.parsing[s.parsingTop-1] == parseEntityTextItalic {
			// ''a'''''b'''
			// discard BoldItalic (rewinds), go back to Italic
			s.parsingPos = s.parsingPos[0:s.parsingTop]
			s.parsing = s.parsing[0:s.parsingTop]; s.parsingTop--
			s.states = s.states[0:s.stateTop]; s.stateTop--
			s.parsingTopState = s.parsing[s.parsingTop]

			//fmt.Printf("endStateQR3: %v %v %v\n", topPos, s.pos(), s.parsing)

			d := 1
			if s.parsingTop == 0 { d = 0 }
			s.rewind = s.pos() - topPos + off1 + d

			if s.pop != nil {
				s.pop(parseEntityTextBoldItalic, 0, 0, 0, 0)
			}
			return s.end(c, off1-1, off2-1, 0, s.rewind-d)
		} else {
			// '''''a''b'''
			// convert BoldItalic into Italic
			s.parsingPos[s.parsingTop] = topPos - 3
			s.parsing[s.parsingTop] = parseEntityTextItalic
			s.states[s.stateTop] = stateInEntityTextItalic
			s.parsingTopState = parseEntityTextItalic
			//if s.pop != nil { s.pop(s.parsing[s.parsingTop], off1, off2) }

			// push new Bold
			s.parsingPos = append(s.parsingPos, topPos)
			s.parsing = append(s.parsing, parseEntityTextBold); s.parsingTop++
			s.parsingTopState = parseEntityTextBold
			s.states = append(s.states, stateInEntityTextBold); s.stateTop++
			s.step = stateInEntityTextBold
		}
		//fmt.Printf("endStateQR3: %v %v %v\n", s.pos(), string(c), s.parsing)
	}
	return s.end(c, off1, off2, 0, 0)
}

// ''' (right)
func stateQuoteRight3(s *scanner, c int) int {
	//fmt.Printf("stateQuoteRight3: %v %v %v\n", s.pos(), string(c), s.parsing)
	if c == '\'' && s.parsingTopState == parseEntityTextBoldItalic {
		s.step = stateQuoteRight4
		return scanContinue
	}
	return endStateQR3(s, c, 3, 3)
}

// '''' (right)
func stateQuoteRight4(s *scanner, c int) int {
	//fmt.Printf("stateQuoteRight4: %v %v %v\n", s.pos(), string(c), s.parsing)
	if c == '\'' && s.parsingTopState == parseEntityTextBoldItalic {
		s.step = stateQuoteRight5
		return scanContinue
	}
	// same as stateQuoteRight3 in this case
	return endStateQR3(s, c, 3, 3)
}

// ''''' (right)
func stateQuoteRight5(s *scanner, c int) int {
	//fmt.Printf("stateQuoteRight5: %v %v %v\n", s.pos(), string(c), s.parsing)
	return s.end(c, 5, 5, 0, 0)
}

// {
func stateBrL1(s *scanner, c int) int {
	//fmt.Printf("stateBrL1: %v %v %v\n", s.pos(), string(c), s.parsing)

	if c == '{' {
		if s.parsingTopState == parseEntityText {
			//fmt.Printf("stateBrL1: %v %v text end\n", string(c), s.parsing)
			s.rewind = 1;
			return s.end(c, 0, 0, 0, 0)
		}

		s.step = stateBrL2
		return scanContinue
	}

	if s.stateTop < 0 {
		//fmt.Printf("stateBrL1: %v %v %v\n", s.pos(), string(c), s.parsing)
		return s.begin(stateInEntityText, parseEntityText, scanBeginText, c, 0)
	}

	s.step = s.states[s.stateTop]
	return s.step(s, c) //return scanContinue
}

// {{
func stateBrL2(s *scanner, c int) int {
	//fmt.Printf("stateBrL2: %v %v %v\n", s.pos(), string(c), s.parsing)
	code := s.begin(stateInEntityTemplate, parseEntityTemplate, scanBeginTemplate, c, 2)
	//fmt.Printf("stateBrL2: %v %v %v\n", s.pos(), string(c), s.parsing)
	if c != '}' && code != scanEnd {
		s.pushParseState(parseEntityTemplateName)
	}
	return code
}

// }
func stateBrR1(s *scanner, c int) int {
	//fmt.Printf("stateBrR1: %v %v %v\n", s.pos(), string(c), s.parsing)
	if c == '}' {
		s.step = stateBrR2
		return scanContinue
	}

	if s.stateTop < 0 {
		return s.begin(stateInEntityText, parseEntityText, scanBeginText, c, 0)
	}

	s.step = s.states[s.stateTop]
	return s.step(s, c) //return scanContinue
}

// }}
func stateBrR2(s *scanner, c int) int {
	//fmt.Printf("stateBrR2: %v %v %v %v\n", s.pos(), string(c), s.parsing, s.parsingTopState)
	if s.parsingTopState == parseEntityTemplateName {
		s.popParseState(0, 2, 0, 0)
	} else {
		for s.parsingTopState == parseEntityTemplateProp {
			//fmt.Printf("stateBrR2: %v %v %v\n", s.pos(), string(c), s.parsing)
			s.popParseState(1, 2, 0, 0)
		}
	}
	//fmt.Printf("stateBrR2: %v %v %v\n", s.pos(), string(c), s.parsing)
	code := s.end(c, 2, 2, 0, 0)
	//fmt.Printf("stateBrR2: %v %v %v\n", s.pos(), string(c), s.parsing)
	return code
}

// [
func stateSqL1(s *scanner, c int) int {
	//fmt.Printf("stateSqL1: %v %v %v\n", s.pos(), string(c), s.parsing)
	if c == '[' {
		s.step = stateSqL2
		return scanContinue
	}
	return s.begin(stateInEntityLink1, parseEntityLink1, scanBeginLink1, c, 1)
}

// [[
func stateSqL2(s *scanner, c int) int {
	//fmt.Printf("stateSqL2: %v %v %v\n", s.pos(), string(c), s.parsing)
	code := s.begin(stateInEntityLink2, parseEntityLink2, scanBeginLink2, c, 2)
	if c != ']' && code != scanEnd {
		s.pushParseState(parseEntityLink2Name)
	}
	return code
}

// ]
func stateSqR1(s *scanner, c int) int {
	//fmt.Printf("stateSqR1: %v %v %v\n", s.pos(), string(c), s.parsing)

	if s.parsingTopState == parseEntityLink1 {
		//fmt.Printf("stateSqR1: %v %v link end\n", string(c), s.parsing)
		return s.end(c, 1, 1, 0, 0)
	}

	if c == ']' {
		s.step = stateSqR2
		return scanContinue
	}

	if s.stateTop < 0 {
		return s.begin(stateInEntityText, parseEntityText, scanBeginText, c, 0)
	}

	s.step = s.states[s.stateTop]
	return s.step(s, c) //return scanContinue
}

// ]]
func stateSqR2(s *scanner, c int) int {
	//fmt.Printf("stateSqR2: %v %v %v\n", s.pos(), string(c), s.parsing)
	switch s.parsingTopState {
	case parseEntityLink2Name:
		s.popParseState(0, 2, 0, 0)
	case parseEntityLink2Prop:
		//fmt.Printf("stateSqR2: %v %v %v\n", s.pos(), string(c), s.parsing)
		s.popParseState(1, 2, 0, 0)
	}
	//fmt.Printf("stateSqR2: %v %v %v\n", s.pos(), string(c), s.parsing)
	return s.end(c, 2, 2, 0, 0)
}

// <
func stateLt(s *scanner, c int) int {
	//fmt.Printf("stateLt: %v %v %v\n", s.pos(), string(c), s.parsing)
	switch {
	case c == '/':
		s.step = stateInTagEndSlash
		return scanContinue
	case c == '>':
		if s.stateTop < 0 {
			return s.begin(stateInEntityText, parseEntityText, scanBeginText, c, 0)
		}

		s.step = s.states[s.stateTop]
		return s.step(s, c) //return scanContinue
	}
	return s.begin(stateInTagBeg, parseEntityTagBeg, scanBeginTagBeg, c, 1)
}

// in <tag>
func stateInTagBeg(s *scanner, c int) int {
	//fmt.Printf("stateInTagBeg: %v %v %v\n", s.pos(), string(c), s.parsing)
	switch {
	case c == '/':
		s.step = stateInTagBegSlash
		return scanContinue
	case c == '>':
		s.step = stateInTagBegGt
		return scanContinue
	}
	//return s.states[s.stateTop](s, c)
	return scanContinue
}

// in-tag '/' as in "<tag />"
func stateInTagBegSlash(s *scanner, c int) int {
	//fmt.Printf("stateInTagBegSlash: %v %v %v\n", s.pos(), string(c), s.parsing)

	if c == '>' {
		s.step = stateInTagBegSlashGt
		return scanContinue
	}

	s.step = s.states[s.stateTop]
	return s.step(s, c) //return scanContinue
}

// in-tag '>' as in "<tag />"
func stateInTagBegSlashGt(s *scanner, c int) int {
	//fmt.Printf("stateInTagBegGt: %v %v %v\n", s.pos(), string(c), s.parsing)
	s.parsing[s.parsingTop] = parseEntityTag
	s.parsingTopState = parseEntityTag
	return s.end(c, 1, 2, 0, 0)
}

// in-tag '>' as in "<tag>"
func stateInTagBegGt(s *scanner, c int) int {
	//fmt.Printf("stateInTagBegGt: %v %v %v\n", s.pos(), string(c), s.parsing)
	return s.end(c, 1, 1, 0, 0)
}

// in-tag slash as in </tag>
func stateInTagEndSlash(s *scanner, c int) int {
	return s.begin(stateInTagEnd, parseEntityTagEnd, scanBeginTagEnd, c, 2)
}

// in </tag>
func stateInTagEnd(s *scanner, c int) int {
	if c == '>' {
		s.step = stateInTagEndGt
	}
	return scanContinue
}

func stateInTagEndGt(s *scanner, c int) int {
	//fmt.Printf("stateInTagEndGt: %v %v %v\n", s.pos(), string(c), s.parsing)
	return s.end(c, 2, 1, 0, 0)
}

// \n [spaces]
func stateNewline(s *scanner, c int) int {
	//fmt.Printf("stateNewline: %v %v %v\n", s.pos(), string(c), s.parsing)

	switch {
	case c != '\n' && isSpace(rune(c)):
		s.indent++
		return scanContinue

	case c == '*':
		s.step, s.newlineOffset = stateBeginListBulleted, 1
		return scanContinue

	case c == '#':
		s.step, s.newlineOffset = stateBeginListNumbered, 1
		return scanContinue

	case c == ':':
		s.step, s.newlineOffset = stateBeginIndent, 1
		return scanContinue

	case c == '-':
		s.step, s.newlineOffset = stateNewlineDash1, 1
		return scanContinue

	case c == '=':
		s.step, s.newlineOffset = stateNewlineEqualL1, 1
		return scanContinue

	case c == '}':
		if s.parsingTopState == parseEntityTemplate {
			s.step = stateBrR1
			return scanContinue
		}
	}

	if s.stateTop < 0 {
		//fmt.Printf("stateNewline: %v %v new text\n", string(c), s.parsing)
		code := s.begin(stateInEntityText, parseEntityText, scanBeginText, c, 1)

		if c == '\n' {
			//fmt.Printf("stateNewline: %v %v end text\n", string(c), s.parsing)
			s.indent = 0
			return s.end(c, 0, 0, 0, 0)
		}

		if s.checkSpecial(c) {
			// ...
		}
		return code
	} else if s.checkSpecial(c) {
		return scanContinue
	}

	s.step = s.states[s.stateTop]
	return s.step(s, c) //return scanContinue
}

// *
func stateBeginListBulleted(s *scanner, c int) int {
	//fmt.Printf("stateBeginListBulleted: %v %v %v, indent=%v\n", s.pos(), string(c), s.parsing, s.indent)
	step, state, code := selectSubStep(c, stateInListBulleted), parseEntityListBulleted, scanBeginListBulleted
	return s.begin(step, state, code, c, s.indent + s.newlineOffset + 1)
}

// #
func stateBeginListNumbered(s *scanner, c int) int {
	//fmt.Printf("stateBeginListNumbered: %v %v %v, indent=%v\n", s.pos(), string(c), s.parsing, s.indent)
	step, state, code := selectSubStep(c, stateInListNumbered), parseEntityListNumbered, scanBeginListNumbered
	return s.begin(step, state, code, c, s.indent + s.newlineOffset + 1)
}

// :
func stateBeginIndent(s *scanner, c int) int {
	//fmt.Printf("stateBeginIndent: %v %v %v, indent=%v\n", s.pos(), string(c), s.parsing, s.indent)
	step, state, code := selectSubStep(c, stateInIndent), parseEntityIndent, scanBeginIndent
	return s.begin(step, state, code, c, s.indent + s.newlineOffset + 1)
}

func selectSubStep(c int, step func(s *scanner, c int) int) func(s *scanner, c int) int {
	switch c {
	case '*': return stateBeginListBulleted
	case '#': return stateBeginListNumbered
	case ':': return stateBeginIndent
	}
	return step
}

// list item
func stateInLineTerminal(s *scanner, c int) int {
	//fmt.Printf("stateInLineTerminal: %v %v, indent=%v\n", string(c), s.parsing, s.indent)

	if c == '\n' {
		num := s.parsingTop
		for ; 0 < num; num-- {
			switch s.parsing[num] {
			case parseEntityListBulleted:	continue
			case parseEntityListNumbered:	continue
			case parseEntityIndent:		continue
			}
			break
		}
		//fmt.Printf("stateInLineTerminal: %v %v %v\n", num, s.parsingTop, s.parsing)
		for n := s.parsingTop; num < n; n-- {
			s.popParseState(s.indent + s.newlineOffset, 0, 0, 0)
			s.popStepState()
		}
		//fmt.Printf("stateInLineTerminal: %v %v %v\n", num, s.parsingTop, s.parsing) /**/
		/*
		for n := s.parsingTop; 0 < n; n-- {
			s.popParseState(s.indent + s.newlineOffset, 0, 0, 0)
			s.popStepState()
		}
		fmt.Printf("stateInLineTerminal: %v %v\n", s.parsingTop, s.parsing) /**/
		return s.end(c, s.indent + s.newlineOffset, 0, 0, 0)
	}

	s.checkSpecial(c)
	return scanContinue
}

func stateInListBulleted(s *scanner, c int) int {
	return stateInLineTerminal(s, c)
}

func stateInListNumbered(s *scanner, c int) int {
	return stateInLineTerminal(s, c)
}

func stateInIndent(s *scanner, c int) int {
	return stateInLineTerminal(s, c)
}

// -
func stateNewlineDash1(s *scanner, c int) int {
	if c == '-' {
		s.step = stateNewlineDash2
		return scanContinue
	}

	if s.stateTop < 0 {
		return s.begin(stateInEntityText, parseEntityText, scanBeginText, c, 0)
	}

	s.step = s.states[s.stateTop]
	return s.step(s, c) //return scanContinue
}

// --
func stateNewlineDash2(s *scanner, c int) int {
	if c == '-' {
		s.step = stateNewlineDash3
		return scanContinue
	}

	if s.stateTop < 0 {
		return s.begin(stateInEntityText, parseEntityText, scanBeginText, c, 0)
	}

	s.step = s.states[s.stateTop]
	return s.step(s, c) //return scanContinue
}

// ---
func stateNewlineDash3(s *scanner, c int) int {
	if c == '-' {
		s.step = stateNewlineDash4
		return scanContinue
	}

	if s.stateTop < 0 {
		return s.begin(stateInEntityText, parseEntityText, scanBeginText, c, 0)
	}

	s.step = s.states[s.stateTop]
	return s.step(s, c) //return scanContinue
}

// ----
func stateNewlineDash4(s *scanner, c int) int {
	//fmt.Printf("stateNewlineDash4: %v %v %v\n", s.pos(), string(c), s.parsing)
	s.begin(stateInLineTerminal, parseEntityHR, scanBeginHR, c, 0/*s.indent + s.newlineOffset /*+ 4*/)
	//fmt.Printf("stateNewlineDash4: %v %v %v\n", s.pos(), string(c), s.parsing)
	return s.end(c, s.indent + s.newlineOffset, 0, 0, 0)
}

// =
func stateNewlineEqualL1(s *scanner, c int) int {
	if c == '=' {
		s.step = stateNewlineEqualL2
		return scanContinue
	}

	if s.stateTop < 0 {
		return s.begin(stateInEntityText, parseEntityText, scanBeginText, c, 0)
	}

	s.step = s.states[s.stateTop]
	return s.step(s, c) //return scanContinue
}

// ==
func stateNewlineEqualL2(s *scanner, c int) int {
	if c == '=' {
		s.step = stateNewlineEqualL3
		return scanContinue
	}

	return s.begin(stateInHeader, parseEntityHeader2, scanBeginHeader2, c, s.indent + s.newlineOffset + 2)
}

// ===
func stateNewlineEqualL3(s *scanner, c int) int {
	if c == '=' {
		s.step = stateNewlineEqualL4
		return scanContinue
	}

	return s.begin(stateInHeader, parseEntityHeader3, scanBeginHeader3, c, s.indent + s.newlineOffset + 3)
}

// ====
func stateNewlineEqualL4(s *scanner, c int) int {
	if c == '=' {
		s.step = stateNewlineEqualL5
		return scanContinue
	}

	return s.begin(stateInHeader, parseEntityHeader4, scanBeginHeader4, c, s.indent + s.newlineOffset + 4)
}

// =====
func stateNewlineEqualL5(s *scanner, c int) int {
	if c == '=' {
		s.step = stateNewlineEqualR1
		return scanContinue
	}

	return s.begin(stateInHeader, parseEntityHeader5, scanBeginHeader5, c, s.indent + s.newlineOffset + 5)
}

func stateInHeader(s *scanner, c int) int {
	//fmt.Printf("stateInHeader: %v %v, indent=%v\n", string(c), s.parsing, s.indent)

	if c == '=' {
		s.step = stateNewlineEqualR1
		return scanContinue
	}

	s.checkSpecial(c)
	return scanContinue
}

// = (right)
func stateNewlineEqualR1(s *scanner, c int) int {
	//fmt.Printf("stateNewlineEqualR1: %v %v, indent=%v\n", string(c), s.parsing, s.indent)
	if c == '=' {
		s.step = stateNewlineEqualR2
		return scanContinue
	}
	return scanContinue
}

// == (right)
func stateNewlineEqualR2(s *scanner, c int) int {
	//fmt.Printf("stateNewlineEqualR2: %v %v, indent=%v\n", string(c), s.parsing, s.indent)
	if c == '=' {
		s.step = stateNewlineEqualR3
		return scanContinue
	}
	if s.parsingTopState == parseEntityHeader2 {
		return s.end(c, 2 + s.newlineOffset, 2, 0, 0)
	}
	return scanContinue
}

// === (right)
func stateNewlineEqualR3(s *scanner, c int) int {
	//fmt.Printf("stateNewlineEqualR3: %v %v, indent=%v\n", string(c), s.parsing, s.indent)
	if c == '=' {
		s.step = stateNewlineEqualR4
		return scanContinue
	}
	if s.parsingTopState == parseEntityHeader3 {
		return s.end(c, 3 + s.newlineOffset, 3, 0, 0)
	}
	return scanContinue
}

// ==== (right)
func stateNewlineEqualR4(s *scanner, c int) int {
	//fmt.Printf("stateNewlineEqualR4: %v %v, indent=%v\n", string(c), s.parsing, s.indent)
	if c == '=' {
		s.step = stateNewlineEqualR5
		return scanContinue
	}
	if s.parsingTopState == parseEntityHeader4 {
		return s.end(c, 4 + s.newlineOffset, 4, 0, 0)
	}
	return scanContinue
}

// ===== (right)
func stateNewlineEqualR5(s *scanner, c int) int {
	//fmt.Printf("stateNewlineEqualR5: %v %v, indent=%v\n", string(c), s.parsing, s.indent)
	if s.parsingTopState == parseEntityHeader5 {
		return s.end(c, 5 + s.newlineOffset, 5, 0, 0)
	}
	return scanContinue
}

func stateInEntity(s *scanner, c int) int {
	s.checkSpecial(c)
	return scanContinue
}

func stateInEntityText(s *scanner, c int) int {
	return stateInEntity(s, c)
}

func inEntityTextQ(s *scanner, c int) int {
	switch c {
	case '\'':
		s.step = stateQuoteRight1
		return scanContinue
	case '\n':
		/*
		num, foundLineTerm := s.parsingTop, false
		for ; !foundLineTerm && 0 < num; num-- {
			switch s.parsing[num] {
			case parseEntityListBulleted:	foundLineTerm = true
			case parseEntityListNumbered:	foundLineTerm = true
			case parseEntityIndent:		foundLineTerm = true
			}
		}
		//fmt.Printf("inEntityTextQ: %v %v %v\n", num, s.parsingTop, s.parsing)
		if foundLineTerm {
			for n := s.parsingTop; num < n; n-- {
				s.popParseState(s.indent + s.newlineOffset, 0, 0, 0)
				s.popStepState()
			}
		}
		fmt.Printf("inEntityTextQ: %v %v %v\n", num, s.parsingTop, s.parsing) */
		if 0 < s.parsingTop {
			foundLineTerm := false
			switch s.parsing[s.parsingTop-1] {
			case parseEntityListBulleted, parseEntityListNumbered, parseEntityIndent:
				foundLineTerm = true
			}
			if foundLineTerm {
				s.popParseState(s.indent + s.newlineOffset, 0, 0, 0)
				s.popStepState()
				//fmt.Printf("inEntityTextQ: %v %v\n", s.parsingTop, s.parsing)
				return s.states[s.stateTop](s, c)
			}
		}
	}
	return stateInEntity(s, c)
}

func stateInEntityTextBold(s *scanner, c int) int {
	return inEntityTextQ(s, c)
}

func stateInEntityTextItalic(s *scanner, c int) int {
	return inEntityTextQ(s, c)
}

func stateInEntityTextBoldItalic(s *scanner, c int) int {
	return inEntityTextQ(s, c)
}

func stateInEntityTemplateName(s *scanner, c int) int {
	s.pushParseState(parseEntityTemplateName)
	s.step = stateInEntityTemplate
	return stateInEntityTemplate(s, c)
}

func stateInEntityTemplate(s *scanner, c int) int {
	//fmt.Printf("stateInEntityTemplate: %v %v %v\n", string(c), s.parsing, s.parsingTopState)
	switch c {
	case '\n': return scanContinue
	case '|':
		switch s.parsingTopState {
		case parseEntityTemplateName:
			s.popParseState(0, 0, 0, 0)
		case parseEntityTemplateProp:
			s.popParseState(1, 0, 0, 0)
		}
		s.pushParseState(parseEntityTemplateProp)
		//fmt.Printf("stateInEntityTemplate: %v %v\n", string(c), s.parsing)
		return scanContinue
	case '}':
		s.step = stateBrR1
		return scanContinue
	}
	return stateInEntity(s, c)
}

func stateInEntityLink(s *scanner, c int) int {
	//fmt.Printf("stateInEntityLink: %v %v %v\n", string(c), s.parsing, s.parsingTopState)
	if c == ']' {
		s.step = stateSqR1
		return scanContinue
	}
	return stateInEntity(s, c)
}

func stateInEntityLink1(s *scanner, c int) int {
	return stateInEntityLink(s, c)
}

func stateInEntityLink2(s *scanner, c int) int {
	//fmt.Printf("stateInEntityLink2: %v %v %v\n", s.pos(), string(c), s.parsing)
	if c == '|' {
		//fmt.Printf("stateInEntityLink2: %v %v\n", string(c), s.parsing)
		switch s.parsingTopState {
		case parseEntityLink2Name:
			s.popParseState(0, 0, 0, 0)
		case parseEntityLink2Prop:
			s.popParseState(1, 0, 0, 0)
		}
		s.pushParseState(parseEntityLink2Prop)
		//fmt.Printf("stateInEntityLink2: %v %v\n", string(c), s.parsing)
		return scanContinue
	}
	return stateInEntityLink(s, c)
}

func stateError(s *scanner, c int) int {
	return scanError
}

func isSpace(c rune) bool {
	return c == ' ' || c == '\t' || c == '\r' || c == '\n'
}
