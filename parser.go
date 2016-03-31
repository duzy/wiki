//
//  Copyright (C) 2013, Duzy Chan <code@duzy.info>, all rights reserverd.
//
package wiki

import (
	"fmt"
	//"strings"
)

// https://www.mediawiki.org/wiki/Markup_spec
// https://www.mediawiki.org/wiki/Markup_spec/BNF
// https://www.mediawiki.org/wiki/Markup_spec/EBNF
// https://www.mediawiki.org/wiki/Preprocessor_ABNF
//
// Formatting
// ----------
//	Italic				''Italic text''
//	Bold				'''Bold text'''
//	Bold & italic		'''''Bold & italic text'''''
//
// Links
// -----
//	Internal link                           [[Page title|Link label]]
//						[[Page title]]
//	External link                           [http://www.example.org Link label]
//						[http://www.example.org]
//						http://www.example.org
// Lists
// -----
//	Bulleted list		* List item
//						* List item
//	Numbered list		# List item
//						# List item
// Files
// -----
//	Embedded file		[[File:Example.png|thumb|Caption text]]
//
// Markups
// -------
//	<any name="value" />
//	<any name="value">blah blah</any>
//
// References
// ----------
//	Reference							Page text.<ref name="test">[http://www.example.org Link text], additional text.</ref>                    Page text.[1]
//	Additional use of same reference	<ref name="test" />			Page text.[1]
//	Display references					<references />
//
// Discusstion
// -----------
//	Signature with timestamp	~~~~			wikieditor-toolbar-help-content-signaturetimestamp-result: Parse error at position 19 in input: Username (talk) 15:54, 10 June 2009 (UTC)
//	Signature			~~~			wikieditor-toolbar-help-content-signature-result: Parse error at position 19 in input: Username (talk)
//	Indent				Normal text
//					:Indented text
//					::Indented text
const (
	WikiEntityWiki EntityType	= iota // The root wiki entity.
	/*		      *///
	WikiEntityText		// Normal text.
	/*		      *///
	WikiEntityTextBold	// '''Bold text'''
	/*		      *///
	WikiEntityTextItalic	// ''Italic text''
	WikiEntityTextBoldItalic// '''''bold italic'''''
	/*		      */// '''''bold italic'' bold'''
	/*		      */// '''''bold italic''' italic''
	/*		      *///
	/*		      */// ''a'''''A'''		italic + bold
	/*		      */// '''a'''''A''		bold + italic
	/*		      *///
	WikiEntityHeading2	// == Heading text ==
	/*		      *///
	WikiEntityHeading3	// === Heading text ===
	/*		      *///
	WikiEntityHeading4	// ==== Heading text ====
	/*		      *///
	WikiEntityHeading5	// ===== Heading text =====
	/*		      *///
	WikiEntityLinkExternal	// [http://... Link label]
	/*		      */// [http://...]
	/*		      */// http://...
	/*		      *///
	//WikiEntityLinkLabel	// Link label 
	/*		      *///
	WikiEntityLinkInternal	// [[Title|Link label]]
	/*		      */// [[File:Example.png|thumb|Caption Text]]
	WikiEntityLinkInternalName
	/*		      */// Title
	/*		      */// File:Example.png
	WikiEntityLinkInternalProp
	/*		      */// |Link label
	/*		      */// |thumb, |Caption Text
	/*		      *///
	WikiEntityTemplate	// {{wikipedia}}
	/*		      */// {{IPA|/ˈwɪki/|/ˈwiːki/}}
	/*		      */// {{audio|en-us-wiki.ogg|Audio (US)}}
	/*		      */// {{homophones|lang=en|wicky}}
	/*		      */// {{etyl|haw|en}}
	/*		      */// {{term|wikiwiki||quick|lang=haw}}
	/*		      *///
	WikiEntityTemplateName	//
	WikiEntityTemplateProp	// |prop
	/*		      *///
	WikiEntityTag		// <tag />
	WikiEntityTagBeg	// <tag>
	WikiEntityTagProp	// name="value"		(x)
	WikiEntityTagEnd	// </tag>
	/*		      *///
	WikiEntityListBulleted	// * item
	/*		      */// 
	WikiEntityListNumbered	// # item
	/*		      */// 
	WikiEntitySignature	// ~~~
	/*		      */// 
	WikiEntitySignatureTimestamp // ~~~~
	/*		      */// 
	WikiEntityIndent	// :Indented text, ::Indented text
	/*		      */// 
	WikiEntityHR		// ----
	/*		      */// 
)

var entityTypeNames = []string{
	WikiEntityWiki:				"WikiEntityWiki",
	WikiEntityText:				"WikiEntityText",
	WikiEntityTextBold:			"WikiEntityTextBold",
	WikiEntityTextItalic:                   "WikiEntityTextItalic",
	WikiEntityTextBoldItalic:               "WikiEntityTextBoldItalic",
	WikiEntityHeading2:			"WikiEntityHeading2",
	WikiEntityHeading3:			"WikiEntityHeading3",
	WikiEntityHeading4:			"WikiEntityHeading4",
	WikiEntityHeading5:			"WikiEntityHeading5",
	WikiEntityLinkExternal:                 "WikiEntityLinkExternal",
	WikiEntityLinkInternal:                 "WikiEntityLinkInternal",
	WikiEntityLinkInternalName:             "WikiEntityLinkInternalName",
	WikiEntityLinkInternalProp:             "WikiEntityLinkInternalProp",
	WikiEntityTemplate:			"WikiEntityTemplate",
	WikiEntityTemplateName:                 "WikiEntityTemplateName",
	WikiEntityTemplateProp:                 "WikiEntityTemplateProp",
	WikiEntityTag:				"WikiEntityTag",
	WikiEntityTagBeg:			"WikiEntityTagBeg",
	WikiEntityTagProp:			"WikiEntityTagProp", /***/
	WikiEntityTagEnd:			"WikiEntityTagEnd",
	WikiEntityListBulleted:                 "WikiEntityListBulleted",
	WikiEntityListNumbered:                 "WikiEntityListNumbered",
	WikiEntitySignature:                    "WikiEntitySignature",
	WikiEntitySignatureTimestamp:           "WikiEntitySignatureTimestamp",
	WikiEntityIndent:			"WikiEntityIndent",
	WikiEntityHR:				"WikiEntityHR",
}

type EntityType int8
type Entity struct {
	Type EntityType
	Pos int
	Raw []byte
	Text string
	Entities []*Entity // all child entities
}

func (e Entity) String() string {
	return fmt.Sprintf("%v{%v}", e.Type, string(e.Raw))
}

func (t EntityType) String() string {
	return entityTypeNames[int(t)]
}

type parser struct {
	scan *scanner
	data []byte

	// stacks
	state []EntityType
	//pos []int
	//off []int
	entities []*Entity // parsed entity stack (parents)

	entity *Entity
}

func (p *parser) push(state EntityType) {
	if 0 < len(p.state) {
		//fmt.Printf("push: [stack=%v, state=%v, entities=%v]\n", p.state, state, p.entities)
		p.entities = append(p.entities, p.entity)
		p.entity = new(Entity)
		p.entity.Pos = p.scan.pos()
		//p.entity.Type = state
	}
	p.state = append(p.state, state)
}

func (p *parser) pop(state EntityType, pos1, pos2, off1, off2 int) {
	//fmt.Printf("pop: [stack=%v, state=%v, off=[%v, %v], entities=%v] %s, %s\n", p.state, state, off1, off2, p.entities, string(p.data[pos1:pos2]), string(p.data[pos2:]))

	top := len(p.state) - 1
	if top < 0 {
		//fmt.Printf("pop: %s [stack=%v, state=%v, pos-stack=%v, off=[%v, %v], empty]\n", p.entity.Text, p.state, state, p.pos, off1, off2)
		fmt.Printf("pop: %s [stack=%v, state=%v, off=[%v, %v], empty]\n", p.entity.Text, p.state, state, off1, off2)
		return
	}

	//pos1, pos2 := pos0 /*p.pos[top]*/, p.scan.pos() - p.scan.rewind
	doPush, doPop := false, true
	switch {
	case p.state[top] == state && pos1 == 0 && pos2 == 0 && off1 == 0 && off2 == 0:
		p.state = p.state[0:top]
		if l := len(p.entities); 0 < l {
			p.entity = p.entities[l-1]
			p.entities = p.entities[0:l-1]
			//fmt.Printf("pop: %v [stack=%v]\n", p.entities, p.state)
		}
		return
	case p.state[top] == WikiEntityTextBoldItalic:
		switch state {
		case WikiEntityTextBold:
			if 0 < top && p.state[top-1] == WikiEntityTextItalic {
				// in case of: ''a'''''A'''
				fmt.Printf("pop-error: %s [stack=%v, state=%v, pos1=%v, pos2=%v, off=[%v, %v]]\n", p.entity, p.state, state, pos1, pos2, off1, off2)
			} else {
				p.state[top] = WikiEntityTextItalic
				//p.pos[top] -= off1 // step back for '''
				doPush = true
			}
		case WikiEntityTextItalic:
			if 0 < top && p.state[top-1] == WikiEntityTextBold {
				// in case of: '''A'''''a''
				fmt.Printf("pop-error: %s [stack=%v, state=%v, pos1=%v, pos2=%v, off=[%v, %v]]\n", p.entity, p.state, state, pos1, pos2, off1, off2)
			} else {
				p.state[top] = WikiEntityTextBold
				//p.pos[top] -= off1 // step back for ''
				doPush = true
			}
		}
	}
	if doPush {
		p.entities = append(p.entities, p.entity)
		p.entity = new(Entity)
	} else if doPop {
		//p.pos, p.state = p.pos[0:top], p.state[0:top]
		p.state = p.state[0:top]
	}

	switch state {
	case WikiEntityLinkInternalProp, WikiEntityTemplateProp:
		pos1++ // skip '|'
	}

	// make p.entity
	p.entity.Type = state
	p.entity.Text = string(p.data[pos1 : pos2-off2])

	if top = len(p.entities)-1; top < 0 {
		//fmt.Printf("pop: %v [stack=%v, state=%v, entities=%v] (no parents)\n", p.entity, p.state, state, p.entities)
		return
	}

	parent := p.entities[top]
	p.entity.Pos = pos1 - off1 - parent.Pos

	// Entity.Pos
	switch state {
	case WikiEntityTemplateName, WikiEntityTemplateProp:
		if 0 < top { p.entity.Pos += 2 } // FIXME: ...
	}

	// Entity.Raw
	switch state {
	default:
		if l := len(p.data); pos1-off1 < 0 || l < pos1-off1 || l < pos2 {
			fmt.Printf("pop: out-of-range: %s [state=%v, stack=%v, pos1=%v, pos2=%v, off=[%v, %v], len=%v]\n", p.entity.Text, state, p.state, pos1, pos2, off1, off2, l)
			//panic(s)
		} else {
			p.entity.Raw = p.data[pos1-off1 : pos2]
		}
	case WikiEntityLinkInternalProp, WikiEntityTemplateProp:
		p.entity.Raw = p.data[pos1-off1 : pos2-off2]
	}

	//fmt.Printf("pop: %v %v\n", p.entity, p.entities)
	//for k, e := range p.entity.Entities { fmt.Printf("\t%v: %v\n", k, e) }

	parent.Entities = append(parent.Entities, p.entity)
	//fmt.Printf("pop: %v [stack=%v, state=%v, parents=%v, parent=%v%v]\n", p.entity, p.state, state, p.entities, parent, parent.Entities)
	p.entity, p.entities = parent, p.entities[0:top]
	//fmt.Printf("pop: %v [stack=%v, state=%v, parents=%v, parent=%v%v]\n", p.entity, p.state, state, p.entities, parent, parent.Entities)
}

func (p *parser) parse(wiki *Entity, data []byte) (err error) {
	p.data = data
	p.scan.push = p.push
	p.scan.pop = p.pop
	pos, parent := 0, wiki
	parents := make([]*Entity, WikiEntityHeading5 - WikiEntityHeading2 + 1)
	//tags := make([]*Entity, 0, 5) // all WikiEntityTagBeg except tags[0]
	for i, _ := range parents {
		// Default parents are the root entity 'wiki'
		parents[i] = wiki
	}
	for {
		p.entity = new(Entity)

		ent, rest, e := p.scan.next(p.data)
		if e != nil {
			err = e
			return
		}

		l := len(ent)
		state, shift := p.scan.state, p.scan.shift
		p.entity.Raw = ent
		p.entity.Pos = pos

		if state != WikiEntityText && 0 < l && ent[0] == '\n' {
			p.entity.Raw = p.entity.Raw[1:]
			p.entity.Pos++
		}

		switch state {
		case WikiEntityIndent, WikiEntityListBulleted, WikiEntityListNumbered:
			p.entity.Raw = p.entity.Raw[p.scan.indent:]
			p.entity.Pos += p.scan.indent
		}

		//fmt.Printf("scan: %v (state=%v, shift=%v)\n", string(ent), state, shift)

		if shift[0] < l && shift[1] < l {
			a, b := shift[0], l - shift[1]
			if a <= b {
				switch state {
				case WikiEntityIndent, WikiEntityListBulleted, WikiEntityListNumbered:
					a++ // skip ':', '*', '#'
				}
				p.entity.Type = state
				p.entity.Text = string(ent[a:b])
				shift[0] = a
			}
		}

		//fmt.Printf("scan: %v (%v, stack=%v, state=%v, shift=%v, children=%v)\n", string(ent), p.entity.Text, p.state, state, shift, len(p.entity.Entities))
		/* for k, e := range p.entity.Entities {
			fmt.Printf("scan: %v: %v: %v\n", k, e.Type, e.Text)
		} */

		// If the header level is less or equaled to the parent,
		// we need to reset the parent.
		isHeading := WikiEntityHeading2 <= p.entity.Type && p.entity.Type <= WikiEntityHeading5
		if isHeading && p.entity.Type <= parent.Type {
			i := p.entity.Type - WikiEntityHeading2
			parent = parents[i]
		}

		// Add p.entity to the current 'parent'
		parent.Entities = append(parent.Entities, p.entity)
		pos += l

		// Select new parent
		switch {
		case isHeading:
			// Change parent for all other entities and sub-levels.
			parent = p.entity
			i := p.entity.Type - WikiEntityHeading2
			for i++; int(i) < len(parents); i++ {
				parents[i] = parent
			}
			/*
		case state == WikiEntityTagBeg:
			tags = append(tags, parent)
			parent = p.entity
		case state == WikiEntityTagEnd:
			if n := len(tags); 0 < n {
				parent = tags[n-1]
				tags = tags[0:n-1]
			} */
		}

		if p.data = rest; len(p.data) <= 0 {
			break
		}
	}
	return
}

func Parse(data []byte) (wiki *Entity, err error) {
	p := new(parser)
	p.scan = new(scanner)

	wiki = new(Entity)
	wiki.Type = WikiEntityWiki
	err = p.parse(wiki, data)
	return
}

func ParseString(s string) (wiki *Entity, err error) {
	return Parse([]byte(s))
}
