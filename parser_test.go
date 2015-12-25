//
//  Copyright (C) 2013, Duzy Chan <code@duzy.info>, all rights reserverd.
//
package wiki

import (
	"compress/gzip"
	"testing"
	"os"
	"io/ioutil"
	"fmt"
)

type entityTest struct{
	wiki *Entity
	src string
	entities []*entityTestResult
}
type entityTestResult struct {
	t EntityType
	s string
	a []*entityTestResult
}

func dump(t *testing.T, entity *Entity) {
	t.Logf("%d: %v", entity.Type, string(entity.Text))
	for _, ent := range entity.Entities {
		dump(t, ent)
	}
}

func checkEntityResults(t *testing.T, i int, tag string, raw []byte, text string, entity *Entity, results []*entityTestResult, skipEmpty bool) {
	var check func(i, ci int, ty EntityType, raw []byte, text string, ents []*Entity, reses []*entityTestResult)
	/* makeraw := func(ent *Entity, d []byte, a, b int) string {
		off, off2 := 0, -1
		switch ent.Type {
		case WikiEntityTextBold:		off = 3
		case WikiEntityTextItalic:		off = 2
		case WikiEntityTextBoldItalic:		off = 5
		case WikiEntityHeading2:		off = 2
		case WikiEntityHeading3:		off = 3
		case WikiEntityHeading4:		off = 4
		case WikiEntityHeading5:		off = 5
		case WikiEntityLinkExternal:		off = 1
		case WikiEntityLinkInternal:			off = 2
		case WikiEntityLinkInternalProp:		off, off2 = 1, 0
		case WikiEntityTemplate:		off = 2
		case WikiEntityTemplateProp:		off, off2 = 1, 0
		case WikiEntityTag:			off, off2 = 1, 2
		case WikiEntityTagBeg:			off = 1
		case WikiEntityTagProp:			off, off2 = 0, 0
		case WikiEntityTagEnd:			off, off2 = 2, 1
		case WikiEntityListBulleted:		off, off2 = 1, 0
		case WikiEntityListNumbered:		off, off2 = 1, 0
		case WikiEntitySignature:		off, off2 = 0, 0
		case WikiEntitySignatureTimestamp:	off, off2 = 1, 0
		case WikiEntityIndent:			off, off2 = 1, 0
		case WikiEntityHR:			off, off2 = 0, 0
		}
		if off2 < 0 { off2 = off }
		return string(d[a-off:b+off2])
	} */
	check = func(i, ci int, ty EntityType, raw []byte, text string, ents []*Entity, reses []*entityTestResult) {
		n1 := len(ents)
		n2 := len(reses)
		if n1 != n2 && !(skipEmpty && n2 == 0) {
			t.Errorf("%s: [%d] len: %v != %v (%s)", tag, i, n1, n2, string(raw))
			for n := 0; n < n1 && n < n2; n++ {
				if string(ents[n].Text) != reses[n].s {
					t.Errorf("%s: [%d, %d, %d] %v != %v (raw: %v) (%v, %v)", tag, i, ci, n, string(ents[n].Text), reses[n].s, string(ents[n].Raw), reses[n].t, string(ents[n].Type))
					break
				}
			}
			for n := 0; n < n1 && n < n2; n++ {
				//t.Errorf("%s: [%d, %d, %d] parsed: %v", tag, i, ci, n, ents[n].Text)
				//t.Errorf("%s: [%d, %d, %d] expect: %v", tag, i, ci, n, reses[n].s)
			}
			t.Errorf("%s:-%v", tag, ents)
			return
		}
		for n := 0; n < n1 && n < n2; n++ {
			ent, res := ents[n], reses[n]
			if string(ent.Text) != res.s {
				t.Errorf("%s: [%d, child=%d] \"%v\" != \"%v\"", tag, i, n, string(ent.Text), res.s)
			}
			if ent.Type != res.t {
				t.Errorf("%s: [%d, child=%d] \"%v\" != \"%v\"", tag, i, n, ent, res.t)
			}
			//t.Logf("%s: [%d, %d, %d] %v, %v", tag, i, ci, n, string(ent.Raw), ent.Text)
			check(i, n, ent.Type, ent.Raw, ent.Text, ent.Entities, res.a)
		}
	}
	check(i, -1, entity.Type, raw, text, entity.Entities, results)
}

func TestParse(t *testing.T) {
	tests := []entityTest{
		/***** 0 *****/
		{nil, `Normal text.`,
			[]*entityTestResult{
				{WikiEntityText, `Normal text.`, []*entityTestResult{}},
			}},
		/***** 1 *****/
		{nil, `'''Bold text'''`,
			[]*entityTestResult{
				{WikiEntityTextBold, `Bold text`, []*entityTestResult{}},
			}},
		/***** 2 *****/
		{nil, `''Italic text''`,
			[]*entityTestResult{
				{WikiEntityTextItalic, `Italic text`, []*entityTestResult{}},
			}},
		/***** 3 *****/
		{nil, `'''''Bold Italic text'''''`,
			[]*entityTestResult{
				{WikiEntityTextBoldItalic, `Bold Italic text`, []*entityTestResult{}},
			}},
		/***** 4 *****/
		{nil, `Normal '''Bold''' Normal ''Italic'' Normal '''''Bold Italic''''' Normal`,
			[]*entityTestResult{
				{WikiEntityText,	`Normal `, []*entityTestResult{}},
				{WikiEntityTextBold,	`Bold`, []*entityTestResult{}},
				{WikiEntityText,	` Normal `, []*entityTestResult{}},
				{WikiEntityTextItalic,	`Italic`, []*entityTestResult{}},
				{WikiEntityText,	` Normal `, []*entityTestResult{}},
				{WikiEntityTextBoldItalic, `Bold Italic`, []*entityTestResult{}},
				{WikiEntityText,	` Normal`, []*entityTestResult{}},
			}},
		/***** 5 *****/
		{nil, `normal '''bold ''bold italic'' bold [[''title'''bold'''title'']] bold ''bold italic'' bold''' normal`,
			[]*entityTestResult{
				{WikiEntityText,	`normal `, []*entityTestResult{}},
				{WikiEntityTextBold,	`bold ''bold italic'' bold [[''title'''bold'''title'']] bold ''bold italic'' bold`, []*entityTestResult{
					{WikiEntityTextItalic, `bold italic`, []*entityTestResult{}},
					{WikiEntityLinkInternal, `''title'''bold'''title''`, []*entityTestResult{
						{WikiEntityLinkInternalName, `''title'''bold'''title''`, []*entityTestResult{
							{WikiEntityTextItalic, `title'''bold'''title`, []*entityTestResult{
								{WikiEntityTextBold,	`bold`, []*entityTestResult{}},
							}},
						}},
					}},
					{WikiEntityTextItalic, `bold italic`, []*entityTestResult{}},
				}},
				{WikiEntityText,	` normal`, []*entityTestResult{}},
			}},
		/***** 6 *****/
		{nil, `'''bold''bold-italic[[link''italic''link]]bold-italic''bold'''`,
			[]*entityTestResult{
				{WikiEntityTextBold, `bold''bold-italic[[link''italic''link]]bold-italic''bold`, []*entityTestResult{
					{WikiEntityTextItalic, `bold-italic[[link''italic''link]]bold-italic`, []*entityTestResult{
						{WikiEntityLinkInternal, `link''italic''link`, []*entityTestResult{
							{WikiEntityLinkInternalName, `link''italic''link`, []*entityTestResult{
								{WikiEntityTextItalic, `italic`, []*entityTestResult{}},
							}},
						}},
					}},
				}},
			}},

		/***** 7 *****/
		{nil, `
{{markup}}

== English ==
{{markup}}

=== Noun ===
text text text
`,
			[]*entityTestResult{
				{WikiEntityText, "\n", []*entityTestResult{}},
				{WikiEntityTemplate, `markup`, []*entityTestResult{
					{WikiEntityTemplateName, `markup`, []*entityTestResult{}},
				}},
				{WikiEntityText, "\n", []*entityTestResult{}},
				{WikiEntityHeading2, " English ", []*entityTestResult{
					{WikiEntityText, "\n", []*entityTestResult{}},
					{WikiEntityTemplate, `markup`, []*entityTestResult{
						{WikiEntityTemplateName, `markup`, []*entityTestResult{}},
					}},
					{WikiEntityText, "\n", []*entityTestResult{}},
					{WikiEntityHeading3, " Noun ", []*entityTestResult{
						{WikiEntityText, "\ntext text text\n", []*entityTestResult{}},
					}},
				}},
			}},
		/***** 8 *****/
		{nil, `
{{markup}}

== English ==
{{markup}}

=== Noun ===
text text text

==== Synomous ====
text text text

===== header5 =====
text text text

===== header5 =====
text text text

==== Synomous ====
text text text

== Chinese ==
text text text

=== Noun ===
text text text
`,
			[]*entityTestResult{
				{WikiEntityText, "\n", []*entityTestResult{}},
				{WikiEntityTemplate, `markup`, []*entityTestResult{
					{WikiEntityTemplateName, `markup`, []*entityTestResult{}},
				}},
				{WikiEntityText, "\n", []*entityTestResult{}},
				{WikiEntityHeading2, " English ", []*entityTestResult{
					{WikiEntityText, "\n", []*entityTestResult{}},
					{WikiEntityTemplate, `markup`, []*entityTestResult{
						{WikiEntityTemplateName, `markup`, []*entityTestResult{}},
					}},
					{WikiEntityText, "\n", []*entityTestResult{}},
					{WikiEntityHeading3, " Noun ", []*entityTestResult{
						{WikiEntityText, "\ntext text text\n", []*entityTestResult{}},
						{WikiEntityHeading4, " Synomous ", []*entityTestResult{
							{WikiEntityText, "\ntext text text\n", []*entityTestResult{}},
							{WikiEntityHeading5, " header5 ", []*entityTestResult{
								{WikiEntityText, "\ntext text text\n", []*entityTestResult{}},
							}},
							{WikiEntityHeading5, " header5 ", []*entityTestResult{
								{WikiEntityText, "\ntext text text\n", []*entityTestResult{}},
							}},
						}},
						{WikiEntityHeading4, " Synomous ", []*entityTestResult{
							{WikiEntityText, "\ntext text text\n", []*entityTestResult{}},
						}},
					}},
				}},
				{WikiEntityHeading2, " Chinese ", []*entityTestResult{
					{WikiEntityText, "\ntext text text\n", []*entityTestResult{}},
					{WikiEntityHeading3, " Noun ", []*entityTestResult{
						{WikiEntityText, "\ntext text text\n", []*entityTestResult{}},
					}},
				}},
			}},

		// BUG fixes
		/***** 9 *****/
		{nil, `
{{markup
{{inner1}}
markup
{{inner2}}}}
.`,
			[]*entityTestResult{
				{WikiEntityText,	"\n", []*entityTestResult{}},
				{WikiEntityTemplate,	`markup
{{inner1}}
markup
{{inner2}}`,
					[]*entityTestResult{
						{WikiEntityTemplateName, `markup
{{inner1}}
markup
{{inner2}}`,
							[]*entityTestResult{
								{WikiEntityTemplate, `inner1`, []*entityTestResult{
									{WikiEntityTemplateName, `inner1`, []*entityTestResult{}},
								}},
								{WikiEntityTemplate, `inner2`, []*entityTestResult{
									{WikiEntityTemplateName, `inner2`, []*entityTestResult{}},
								}},
							}},
					}},
				{WikiEntityText,	"\n.", []*entityTestResult{}},
			}},
		/***** 10 *****/
		{nil, `
# text 1 {{markup1|prop}} text
# text 2 {{markup2|prop}} text
* text 3 {{markup3|prop}} text
* text 4 {{markup4|prop}} text
: text 5 {{markup5|prop}} text
: text 6 {{markup6|prop}} text
`,
			[]*entityTestResult{
				{WikiEntityListNumbered, " text 1 {{markup1|prop}} text", []*entityTestResult{
					{WikiEntityTemplate, `markup1|prop`, []*entityTestResult{
						{WikiEntityTemplateName, `markup1`, []*entityTestResult{}},
						{WikiEntityTemplateProp, `prop`, []*entityTestResult{}},
					}},
				}},
				{WikiEntityListNumbered, " text 2 {{markup2|prop}} text", []*entityTestResult{
					{WikiEntityTemplate, `markup2|prop`, []*entityTestResult{
						{WikiEntityTemplateName, `markup2`, []*entityTestResult{}},
						{WikiEntityTemplateProp, `prop`, []*entityTestResult{}},
					}},
				}},
				{WikiEntityListBulleted, " text 3 {{markup3|prop}} text", []*entityTestResult{
					{WikiEntityTemplate, `markup3|prop`, []*entityTestResult{
						{WikiEntityTemplateName, `markup3`, []*entityTestResult{}},
						{WikiEntityTemplateProp, `prop`, []*entityTestResult{}},
					}},
				}},
				{WikiEntityListBulleted, " text 4 {{markup4|prop}} text", []*entityTestResult{
					{WikiEntityTemplate, `markup4|prop`, []*entityTestResult{
						{WikiEntityTemplateName, `markup4`, []*entityTestResult{}},
						{WikiEntityTemplateProp, `prop`, []*entityTestResult{}},
					}},
				}},
				{WikiEntityIndent, " text 5 {{markup5|prop}} text", []*entityTestResult{
					{WikiEntityTemplate, `markup5|prop`, []*entityTestResult{
						{WikiEntityTemplateName, `markup5`, []*entityTestResult{}},
						{WikiEntityTemplateProp, `prop`, []*entityTestResult{}},
					}},
				}},
				{WikiEntityIndent, " text 6 {{markup6|prop}} text", []*entityTestResult{
					{WikiEntityTemplate, `markup6|prop`, []*entityTestResult{
						{WikiEntityTemplateName, `markup6`, []*entityTestResult{}},
						{WikiEntityTemplateProp, `prop`, []*entityTestResult{}},
					}},
				}},
				{WikiEntityText,	"\n", []*entityTestResult{}},
			}},
		/***** 11 *****/
		{nil, `
  # text 1 {{markup1|prop}} text
  # text 2 {{markup2|prop}} text
  * text 3 {{markup3|prop}} text
  * text 4 {{markup4|prop}} text
  : text 5 {{markup5|prop}} text
  : text 6 {{markup6|prop}} text
`,
			[]*entityTestResult{
				{WikiEntityListNumbered, " text 1 {{markup1|prop}} text", []*entityTestResult{
					{WikiEntityTemplate, `markup1|prop`, []*entityTestResult{
						{WikiEntityTemplateName, `markup1`, []*entityTestResult{}},
						{WikiEntityTemplateProp, `prop`, []*entityTestResult{}},
					}},
				}},
				{WikiEntityListNumbered, " text 2 {{markup2|prop}} text", []*entityTestResult{
					{WikiEntityTemplate, `markup2|prop`, []*entityTestResult{
						{WikiEntityTemplateName, `markup2`, []*entityTestResult{}},
						{WikiEntityTemplateProp, `prop`, []*entityTestResult{}},
					}},
				}},
				{WikiEntityListBulleted, " text 3 {{markup3|prop}} text", []*entityTestResult{
					{WikiEntityTemplate, `markup3|prop`, []*entityTestResult{
						{WikiEntityTemplateName, `markup3`, []*entityTestResult{}},
						{WikiEntityTemplateProp, `prop`, []*entityTestResult{}},
					}},
				}},
				{WikiEntityListBulleted, " text 4 {{markup4|prop}} text", []*entityTestResult{
					{WikiEntityTemplate, `markup4|prop`, []*entityTestResult{
						{WikiEntityTemplateName, `markup4`, []*entityTestResult{}},
						{WikiEntityTemplateProp, `prop`, []*entityTestResult{}},
					}},
				}},
				{WikiEntityIndent, " text 5 {{markup5|prop}} text", []*entityTestResult{
					{WikiEntityTemplate, `markup5|prop`, []*entityTestResult{
						{WikiEntityTemplateName, `markup5`, []*entityTestResult{}},
						{WikiEntityTemplateProp, `prop`, []*entityTestResult{}},
					}},
				}},
				{WikiEntityIndent, " text 6 {{markup6|prop}} text", []*entityTestResult{
					{WikiEntityTemplate, `markup6|prop`, []*entityTestResult{
						{WikiEntityTemplateName, `markup6`, []*entityTestResult{}},
						{WikiEntityTemplateProp, `prop`, []*entityTestResult{}},
					}},
				}},
				{WikiEntityText,	"\n", []*entityTestResult{}},
			}},
		/***** 12 *****/ // BUGS
		{nil, `
===Pronoun===
'''any'''

# Any thing(s) or person(s).
#: Any may apply.
#: ''Any may apply.''
#: '''Any may apply.'''
#: '''''Any''' may apply.''
#: '''''Any'' may apply.'''

====Translations====
{{trans-top|Any things or persons}}
`,
			[]*entityTestResult{
				{WikiEntityHeading3, "Pronoun", []*entityTestResult{
					{WikiEntityText, "\n", []*entityTestResult{}},
					{WikiEntityTextBold, "any", []*entityTestResult{}},
					{WikiEntityText, "\n", []*entityTestResult{}},
					{WikiEntityListNumbered, " Any thing(s) or person(s).", []*entityTestResult{}},
					{WikiEntityListNumbered, ": Any may apply.", []*entityTestResult{
						{WikiEntityIndent, " Any may apply.", []*entityTestResult{}},
					}},
					{WikiEntityListNumbered, ": ''Any may apply.''", []*entityTestResult{
						{WikiEntityIndent, " ''Any may apply.''", []*entityTestResult{
							{WikiEntityTextItalic, "Any may apply.", []*entityTestResult{}},
						}},
					}},
					{WikiEntityListNumbered, ": '''Any may apply.'''", []*entityTestResult{
						{WikiEntityIndent, " '''Any may apply.'''", []*entityTestResult{
							{WikiEntityTextBold, "Any may apply.", []*entityTestResult{}},
						}},
					}},
					{WikiEntityListNumbered, ": '''''Any''' may apply.''", []*entityTestResult{
						{WikiEntityIndent, " '''''Any''' may apply.''", []*entityTestResult{
							{WikiEntityTextItalic, "'''Any''' may apply.", []*entityTestResult{
								{WikiEntityTextBold, "Any", []*entityTestResult{}},
							}},
						}},
					}},
					{WikiEntityListNumbered, ": '''''Any'' may apply.'''", []*entityTestResult{
						{WikiEntityIndent, " '''''Any'' may apply.'''", []*entityTestResult{
							{WikiEntityTextBold, "''Any'' may apply.", []*entityTestResult{
								{WikiEntityTextItalic, "Any", []*entityTestResult{}},
							}},
						}},
					}},
					{WikiEntityText, "\n", []*entityTestResult{}},
					{WikiEntityHeading4, "Translations", []*entityTestResult{
						{WikiEntityText, "\n", []*entityTestResult{}},
						{WikiEntityTemplate, "trans-top|Any things or persons", []*entityTestResult{
							{WikiEntityTemplateName, `trans-top`, []*entityTestResult{}},
							{WikiEntityTemplateProp, "Any things or persons", []*entityTestResult{}},
						}},
						{WikiEntityText, "\n", []*entityTestResult{}},
					}},
				}},
			}},

		/***** 13 *****/
		{nil, `{{also|A|Appendix:Variations of "a"|𐌳}}
==Translingual==
{{Basic Latin character info|previous='|next=b|image=[[Image:Letter a.svg|50px]]|hex=61|name=LATIN SMALL LETTER A}}
{{wikisource1911Enc|A}}

===Etymology 1===
[[Image:UncialA-01.svg|50px|Approximate form of Greek upper case Α (a, “alpha”) that was the source for both common variants of ''a'']]
Modification of capital letter {{term|A|lang=mul}}, from {{etyl|la|mul}} {{term|A|lang=la}}, from {{etyl|grc|mul}} letter {{term|Α|tr=A|lang=grc}}.

====Pronunciation====
* {{sense|letter, most languages}} {{IPA|/ɑː/|/a/|lang=mul}}
* {{audio|Open_front_unrounded_vowel.ogg|IPA}}

====Letter====
{{mul-letter|upper=A|lower=a|script=Latn}}

# {{non-gloss definition|The first letter of the [[Appendix:Latin script|basic modern Latin alphabet]].}}

====Symbol====

====See also====

====External links====

===Etymology 2===

====Symbol====
{{head|mul|symbol}}

# {{non-gloss definition|[[atto-]], the prefix for 10<sup>-18</sup> in the [[International System of Units]].}}

===Etymology 3===

====Symbol====

===Etymology 4===

====Symbol====
{{head|mul|symbol}}

# {{context|physics|lang=mul}} [[acceleration]]

{{Letter|page=A
|NATO=Alpha
|Morse=·–
|Character=A1
|Braille=⠁
}}
<gallery caption="Letter styles" perrow=3>
Image:Latin A.png|Capital and lowercase versions of '''A''', in normal and italic type
File:Fraktur letter A.png|Uppercase and lowercase '''A''' in [[Fraktur]]
File:UncialA-01.svg|Approximate form of Greek upper case Α (a, “alpha”) that was the source for both common variants of ''a'''''A''' in [[uncial]] script
</gallery>

----

==English==

===Etymology 1===
`,
			[]*entityTestResult{
				{WikiEntityTemplate, `also|A|Appendix:Variations of "a"|𐌳`, []*entityTestResult{
					{WikiEntityTemplateName, `also`, []*entityTestResult{}},
					{WikiEntityTemplateProp, `A`, []*entityTestResult{}},
					{WikiEntityTemplateProp, `Appendix:Variations of "a"`, []*entityTestResult{}},
					{WikiEntityTemplateProp, `𐌳`, []*entityTestResult{}},
				}},
				{WikiEntityHeading2, "Translingual", []*entityTestResult{
					{WikiEntityText, "\n", []*entityTestResult{}},
					{WikiEntityTemplate, "Basic Latin character info|previous='|next=b|image=[[Image:Letter a.svg|50px]]|hex=61|name=LATIN SMALL LETTER A", []*entityTestResult{
						{WikiEntityTemplateName, `Basic Latin character info`, []*entityTestResult{}},
						{WikiEntityTemplateProp, "previous='", []*entityTestResult{}},
						{WikiEntityTemplateProp, "next=b", []*entityTestResult{}},
						{WikiEntityTemplateProp, "image=[[Image:Letter a.svg|50px]]", []*entityTestResult{
							{WikiEntityLinkInternal, "Image:Letter a.svg|50px", []*entityTestResult{
								{WikiEntityLinkInternalName, "Image:Letter a.svg", []*entityTestResult{}},
								{WikiEntityLinkInternalProp, "50px", []*entityTestResult{}},
							}},
						}},
						{WikiEntityTemplateProp, "hex=61", []*entityTestResult{}},
						{WikiEntityTemplateProp, "name=LATIN SMALL LETTER A", []*entityTestResult{}},
					}},
					{WikiEntityText, "\n", []*entityTestResult{}},
					{WikiEntityTemplate, `wikisource1911Enc|A`, []*entityTestResult{
						{WikiEntityTemplateName, `wikisource1911Enc`, []*entityTestResult{}},
						{WikiEntityTemplateProp, `A`, []*entityTestResult{}},
					}},
					{WikiEntityText, "\n", []*entityTestResult{}},
					{WikiEntityHeading3, "Etymology 1", []*entityTestResult{
						{WikiEntityText, "\n", []*entityTestResult{}},
						{WikiEntityLinkInternal, "Image:UncialA-01.svg|50px|Approximate form of Greek upper case Α (a, “alpha”) that was the source for both common variants of ''a''", []*entityTestResult{
							{WikiEntityLinkInternalName, "Image:UncialA-01.svg", []*entityTestResult{}},
							{WikiEntityLinkInternalProp, "50px", []*entityTestResult{}},
							{WikiEntityLinkInternalProp, "Approximate form of Greek upper case Α (a, “alpha”) that was the source for both common variants of ''a''", []*entityTestResult{
								{WikiEntityTextItalic, "a", []*entityTestResult{}},
							}},
						}},
						{WikiEntityText, "\nModification of capital letter ", []*entityTestResult{}},
						{WikiEntityTemplate, `term|A|lang=mul`, []*entityTestResult{
							{WikiEntityTemplateName, `term`, []*entityTestResult{}},
							{WikiEntityTemplateProp, `A`, []*entityTestResult{}},
							{WikiEntityTemplateProp, `lang=mul`, []*entityTestResult{}},
						}},
						{WikiEntityText, ", from ", []*entityTestResult{}},
						{WikiEntityTemplate, `etyl|la|mul`, []*entityTestResult{
							{WikiEntityTemplateName, `etyl`, []*entityTestResult{}},
							{WikiEntityTemplateProp, `la`, []*entityTestResult{}},
							{WikiEntityTemplateProp, `mul`, []*entityTestResult{}},
						}},
						{WikiEntityText, " ", []*entityTestResult{}},
						{WikiEntityTemplate, `term|A|lang=la`, []*entityTestResult{
							{WikiEntityTemplateName, `term`, []*entityTestResult{}},
							{WikiEntityTemplateProp, `A`, []*entityTestResult{}},
							{WikiEntityTemplateProp, `lang=la`, []*entityTestResult{}},
						}},
						{WikiEntityText, ", from ", []*entityTestResult{}},
						{WikiEntityTemplate, `etyl|grc|mul`, []*entityTestResult{
							{WikiEntityTemplateName, `etyl`, []*entityTestResult{}},
							{WikiEntityTemplateProp, `grc`, []*entityTestResult{}},
							{WikiEntityTemplateProp, `mul`, []*entityTestResult{}},
						}},
						{WikiEntityText, " letter ", []*entityTestResult{}},
						{WikiEntityTemplate, `term|Α|tr=A|lang=grc`, []*entityTestResult{
							{WikiEntityTemplateName, `term`, []*entityTestResult{}},
							{WikiEntityTemplateProp, `Α`, []*entityTestResult{}},
							{WikiEntityTemplateProp, `tr=A`, []*entityTestResult{}},
							{WikiEntityTemplateProp, `lang=grc`, []*entityTestResult{}},
						}},
						{WikiEntityText, ".\n", []*entityTestResult{}},
						{WikiEntityHeading4, "Pronunciation", []*entityTestResult{
							{WikiEntityListBulleted, " {{sense|letter, most languages}} {{IPA|/ɑː/|/a/|lang=mul}}", []*entityTestResult{
								{WikiEntityTemplate, `sense|letter, most languages`, []*entityTestResult{
									{WikiEntityTemplateName, `sense`, []*entityTestResult{}},
									{WikiEntityTemplateProp, `letter, most languages`, []*entityTestResult{}},
								}},
								{WikiEntityTemplate, `IPA|/ɑː/|/a/|lang=mul`, []*entityTestResult{
									{WikiEntityTemplateName, `IPA`, []*entityTestResult{}},
									{WikiEntityTemplateProp, `/ɑː/`, []*entityTestResult{}},
									{WikiEntityTemplateProp, `/a/`, []*entityTestResult{}},
									{WikiEntityTemplateProp, `lang=mul`, []*entityTestResult{}},
								}},
							}},
							{WikiEntityListBulleted, " {{audio|Open_front_unrounded_vowel.ogg|IPA}}", []*entityTestResult{
								{WikiEntityTemplate, `audio|Open_front_unrounded_vowel.ogg|IPA`, []*entityTestResult{
									{WikiEntityTemplateName, `audio`, []*entityTestResult{}},
									{WikiEntityTemplateProp, `Open_front_unrounded_vowel.ogg`, []*entityTestResult{}},
									{WikiEntityTemplateProp, `IPA`, []*entityTestResult{}},
								}},
							}},
							{WikiEntityText, "\n", []*entityTestResult{}},
						}},
						{WikiEntityHeading4, "Letter", []*entityTestResult{
							{WikiEntityText, "\n", []*entityTestResult{}},
							{WikiEntityTemplate, `mul-letter|upper=A|lower=a|script=Latn`, []*entityTestResult{
								{WikiEntityTemplateName, `mul-letter`, []*entityTestResult{}},
								{WikiEntityTemplateProp, `upper=A`, []*entityTestResult{}},
								{WikiEntityTemplateProp, `lower=a`, []*entityTestResult{}},
								{WikiEntityTemplateProp, `script=Latn`, []*entityTestResult{}},
							}},
							{WikiEntityText, "\n", []*entityTestResult{}},
							{WikiEntityListNumbered, " {{non-gloss definition|The first letter of the [[Appendix:Latin script|basic modern Latin alphabet]].}}", []*entityTestResult{
								{WikiEntityTemplate, "non-gloss definition|The first letter of the [[Appendix:Latin script|basic modern Latin alphabet]].", []*entityTestResult{
									{WikiEntityTemplateName, `non-gloss definition`, []*entityTestResult{}},
									{WikiEntityTemplateProp, "The first letter of the [[Appendix:Latin script|basic modern Latin alphabet]].", []*entityTestResult{
										{WikiEntityLinkInternal, "Appendix:Latin script|basic modern Latin alphabet", []*entityTestResult{
											{WikiEntityLinkInternalName, "Appendix:Latin script", []*entityTestResult{}},
											{WikiEntityLinkInternalProp, "basic modern Latin alphabet", []*entityTestResult{}},
										}},
									}},
								}},
							}},
							{WikiEntityText, "\n", []*entityTestResult{}},
						}},
						{WikiEntityHeading4, "Symbol", []*entityTestResult{
							{WikiEntityText, "\n", []*entityTestResult{}},
						}},
						{WikiEntityHeading4, "See also", []*entityTestResult{
							{WikiEntityText, "\n", []*entityTestResult{}},
						}},
						{WikiEntityHeading4, "External links", []*entityTestResult{
							{WikiEntityText, "\n", []*entityTestResult{}},
						}},
					}},
					{WikiEntityHeading3, "Etymology 2", []*entityTestResult{
						{WikiEntityText, "\n", []*entityTestResult{}},
						{WikiEntityHeading4, "Symbol", []*entityTestResult{
							{WikiEntityText, "\n", []*entityTestResult{}},
							{WikiEntityTemplate, "head|mul|symbol", []*entityTestResult{
								{WikiEntityTemplateName, `head`, []*entityTestResult{}},
								{WikiEntityTemplateProp, "mul", []*entityTestResult{}},
								{WikiEntityTemplateProp, "symbol", []*entityTestResult{}},
							}},
							{WikiEntityText, "\n", []*entityTestResult{}},
							{WikiEntityListNumbered, " {{non-gloss definition|[[atto-]], the prefix for 10<sup>-18</sup> in the [[International System of Units]].}}", []*entityTestResult{
								{WikiEntityTemplate, "non-gloss definition|[[atto-]], the prefix for 10<sup>-18</sup> in the [[International System of Units]].", []*entityTestResult{
									{WikiEntityTemplateName, `non-gloss definition`, []*entityTestResult{}},
									{WikiEntityTemplateProp, "[[atto-]], the prefix for 10<sup>-18</sup> in the [[International System of Units]].", []*entityTestResult{
										{WikiEntityLinkInternal, "atto-", []*entityTestResult{
											{WikiEntityLinkInternalName, "atto-", []*entityTestResult{}},
										}},
										/*
										{WikiEntityTagBeg, "sup", []*entityTestResult{
											{WikiEntityText, "-18", []*entityTestResult{}},
											{WikiEntityTagEnd, "sup", []*entityTestResult{}},
										}}, */
										{WikiEntityTagBeg, "sup", []*entityTestResult{}},
										{WikiEntityTagEnd, "sup", []*entityTestResult{}},
										{WikiEntityLinkInternal, "International System of Units", []*entityTestResult{
											{WikiEntityLinkInternalName, "International System of Units", []*entityTestResult{}},
										}},
									}},
								}},
							}},
							{WikiEntityText, "\n", []*entityTestResult{}},
						}},
					}},
					{WikiEntityHeading3, "Etymology 3", []*entityTestResult{
						{WikiEntityText, "\n", []*entityTestResult{}},
						{WikiEntityHeading4, "Symbol", []*entityTestResult{
							{WikiEntityText, "\n", []*entityTestResult{}},
						}},
					}},
					{WikiEntityHeading3, "Etymology 4", []*entityTestResult{
						{WikiEntityText, "\n", []*entityTestResult{}},
						{WikiEntityHeading4, "Symbol", []*entityTestResult{
							{WikiEntityText, "\n", []*entityTestResult{}},
							{WikiEntityTemplate, "head|mul|symbol", []*entityTestResult{
								{WikiEntityTemplateName, `head`, []*entityTestResult{}},
								{WikiEntityTemplateProp, "mul", []*entityTestResult{}},
								{WikiEntityTemplateProp, "symbol", []*entityTestResult{}},
							}},
							{WikiEntityText, "\n", []*entityTestResult{}},
							{WikiEntityListNumbered, " {{context|physics|lang=mul}} [[acceleration]]", []*entityTestResult{
								{WikiEntityTemplate, "context|physics|lang=mul", []*entityTestResult{
									{WikiEntityTemplateName, `context`, []*entityTestResult{}},
									{WikiEntityTemplateProp, "physics", []*entityTestResult{}},
									{WikiEntityTemplateProp, "lang=mul", []*entityTestResult{}},
								}},
								{WikiEntityLinkInternal, "acceleration", []*entityTestResult{
									{WikiEntityLinkInternalName, "acceleration", []*entityTestResult{}},
								}},
							}},
							{WikiEntityText, "\n", []*entityTestResult{}},
							{WikiEntityText, "\n", []*entityTestResult{}},
							{WikiEntityTemplate, "Letter|page=A\n|NATO=Alpha\n|Morse=·–\n|Character=A1\n|Braille=⠁\n", []*entityTestResult{
								{WikiEntityTemplateName, `Letter`, []*entityTestResult{}},
								{WikiEntityTemplateProp, "page=A\n", []*entityTestResult{}},
								{WikiEntityTemplateProp, "NATO=Alpha\n", []*entityTestResult{}},
								{WikiEntityTemplateProp, "Morse=·–\n", []*entityTestResult{}},
								{WikiEntityTemplateProp, "Character=A1\n", []*entityTestResult{}},
								{WikiEntityTemplateProp, "Braille=⠁\n", []*entityTestResult{}},
							}},
							{WikiEntityText, "\n", []*entityTestResult{}},
							{WikiEntityTagBeg, `gallery caption="Letter styles" perrow=3`, []*entityTestResult{}},
							{WikiEntityText, "\nImage:Latin A.png|Capital and lowercase versions of ", []*entityTestResult{}},
							{WikiEntityTextBold, `A`, []*entityTestResult{}},
							{WikiEntityText, ", in normal and italic type\nFile:Fraktur letter A.png|Uppercase and lowercase ", []*entityTestResult{}},
							{WikiEntityTextBold, `A`, []*entityTestResult{}},
							{WikiEntityText, ` in `, []*entityTestResult{}},
							{WikiEntityLinkInternal, `Fraktur`, []*entityTestResult{
								{WikiEntityLinkInternalName, `Fraktur`, []*entityTestResult{}},
							}},
							{WikiEntityText, "\nFile:UncialA-01.svg|Approximate form of Greek upper case Α (a, “alpha”) that was the source for both common variants of ", []*entityTestResult{}},
							{WikiEntityTextItalic, `a`, []*entityTestResult{}},
							{WikiEntityTextBold, `A`, []*entityTestResult{}},
							{WikiEntityText, ` in `, []*entityTestResult{}},
							{WikiEntityLinkInternal, `uncial`, []*entityTestResult{
								{WikiEntityLinkInternalName, `uncial`, []*entityTestResult{}},
							}},
							{WikiEntityText, " script\n", []*entityTestResult{}},
							{WikiEntityTagEnd, `gallery`, []*entityTestResult{}},
							{WikiEntityText, "\n", []*entityTestResult{}},
							{WikiEntityHR, `----`, []*entityTestResult{}},
							{WikiEntityText, "\n", []*entityTestResult{}},
						}},
					}},
				}},
				{WikiEntityHeading2, "English", []*entityTestResult{
					{WikiEntityText, "\n", []*entityTestResult{}},
					{WikiEntityHeading3, "Etymology 1", []*entityTestResult{
						{WikiEntityText, "\n", []*entityTestResult{}},
					}},
				}},
			},
		},
		/***** 14 *****/
		{nil, `text<ref>ref text ''italic'' ref text</ref>`,
			[]*entityTestResult{
				{WikiEntityText, "text", []*entityTestResult{}},
				{WikiEntityTagBeg, "ref", []*entityTestResult{}},
				{WikiEntityText, "ref text ", []*entityTestResult{}},
				{WikiEntityTextItalic, "italic", []*entityTestResult{}},
				{WikiEntityText, " ref text", []*entityTestResult{}},
				{WikiEntityTagEnd, "ref", []*entityTestResult{}},
			},
		},
		/***** 15 *****/
		{nil, `{{text<ref>ref text ''italic'' ref text</ref>text}}`,
			[]*entityTestResult{
				{WikiEntityTemplate, "text<ref>ref text ''italic'' ref text</ref>text", []*entityTestResult{
					{WikiEntityTemplateName, `text<ref>ref text ''italic'' ref text</ref>text`, []*entityTestResult{
						{WikiEntityTagBeg, "ref", []*entityTestResult{}},
						{WikiEntityTextItalic, "italic", []*entityTestResult{}},
						{WikiEntityTagEnd, "ref", []*entityTestResult{}},
					}},
				}},
			},
		},
		/***** 16 *****/
		{nil, `{{in ''a'''''A''' out}}`,
			[]*entityTestResult{
				{WikiEntityTemplate, "in ''a'''''A''' out", []*entityTestResult{
					{WikiEntityTemplateName, `in ''a'''''A''' out`, []*entityTestResult{
						{WikiEntityTextItalic, "a", []*entityTestResult{}},
						{WikiEntityTextBold, "A", []*entityTestResult{}},
					}},
				}},
			},
		},
		/***** 17 *****/
		{nil, `in ''a'''''A''' out`,
			[]*entityTestResult{
				{WikiEntityText, "in ", []*entityTestResult{}},
				{WikiEntityTextItalic, "a", []*entityTestResult{}},
				{WikiEntityTextBold, "A", []*entityTestResult{}},
				{WikiEntityText, " out", []*entityTestResult{}},
			},
		},
		/***** 18 *****/
		{nil, `{{in '''A'''''a'' out}}`,
			[]*entityTestResult{
				{WikiEntityTemplate, "in '''A'''''a'' out", []*entityTestResult{
					{WikiEntityTemplateName, `in '''A'''''a'' out`, []*entityTestResult{
						{WikiEntityTextBold, "A", []*entityTestResult{}},
						{WikiEntityTextItalic, "a", []*entityTestResult{}},
					}},
				}},
			},
		},
		/***** 19 *****/
		{nil, `in '''A'''''a'' out`,
			[]*entityTestResult{
				{WikiEntityText, "in ", []*entityTestResult{}},
				{WikiEntityTextBold, "A", []*entityTestResult{}},
				{WikiEntityTextItalic, "a", []*entityTestResult{}},
				{WikiEntityText, " out", []*entityTestResult{}},
			},
		},

		/***** 20 *****/
		{nil, `'''''Any''' may apply.''`,
			[]*entityTestResult{
				{WikiEntityTextItalic, "'''Any''' may apply.", []*entityTestResult{
					{WikiEntityTextBold, "Any", []*entityTestResult{}},
				}},
			},
		},
		/***** 21 *****/
		{nil, `'''''Any'' may apply.'''`,
			[]*entityTestResult{
				{WikiEntityTextBold, "''Any'' may apply.", []*entityTestResult{
					{WikiEntityTextItalic, "Any", []*entityTestResult{}},
				}},
			},
		},
		/***** 22 *****/
		{nil, `[['''''Any'' may apply.'''|prop]]`,
			[]*entityTestResult{
				{WikiEntityLinkInternal, "'''''Any'' may apply.'''|prop", []*entityTestResult{
					{WikiEntityLinkInternalName, "'''''Any'' may apply.'''", []*entityTestResult{
						{WikiEntityTextBold, "''Any'' may apply.", []*entityTestResult{
							{WikiEntityTextItalic, "Any", []*entityTestResult{}},
						}},
					}},
					{WikiEntityLinkInternalProp, "prop", []*entityTestResult{}},
				}},
			},
		},
		/***** 23 *****/
		{nil, `[[Image:UncialA-01.svg|50px|Approximate form of Greek upper case Α (a, “alpha”) that was the source for both common variants of ''a'']]`,
			[]*entityTestResult{
				{WikiEntityLinkInternal, "Image:UncialA-01.svg|50px|Approximate form of Greek upper case Α (a, “alpha”) that was the source for both common variants of ''a''", []*entityTestResult{
					{WikiEntityLinkInternalName, "Image:UncialA-01.svg", []*entityTestResult{}},
					{WikiEntityLinkInternalProp, "50px", []*entityTestResult{}},
					{WikiEntityLinkInternalProp, "Approximate form of Greek upper case Α (a, “alpha”) that was the source for both common variants of ''a''", []*entityTestResult{
						{WikiEntityTextItalic, "a", []*entityTestResult{}},
					}},
				}},
			},
		},
		/***** 24 *****/
		{nil, `text
[[Image:UncialA-01.svg|50px|Approximate form of Greek upper case Α (a, “alpha”) that was the source for both common variants of ''a'']]
Modification of capital letter
`,
			[]*entityTestResult{
				{WikiEntityText, "text\n", []*entityTestResult{}},
				{WikiEntityLinkInternal, "Image:UncialA-01.svg|50px|Approximate form of Greek upper case Α (a, “alpha”) that was the source for both common variants of ''a''", []*entityTestResult{
					{WikiEntityLinkInternalName, "Image:UncialA-01.svg", []*entityTestResult{}},
					{WikiEntityLinkInternalProp, "50px", []*entityTestResult{}},
					{WikiEntityLinkInternalProp, "Approximate form of Greek upper case Α (a, “alpha”) that was the source for both common variants of ''a''", []*entityTestResult{
						{WikiEntityTextItalic, "a", []*entityTestResult{}},
					}},
				}},
				{WikiEntityText, "\nModification of capital letter\n", []*entityTestResult{}},
			},
		},
		/***** 25 *****/
		{nil, `
===Etymology 1===
[[Image:UncialA-01.svg|50px|Approximate form of Greek upper case Α (a, “alpha”) that was the source for both common variants of ''a'']]
Modification of capital letter
`,
			[]*entityTestResult{
				{WikiEntityHeading3, "Etymology 1", []*entityTestResult{
					{WikiEntityText, "\n", []*entityTestResult{}},
					{WikiEntityLinkInternal, "Image:UncialA-01.svg|50px|Approximate form of Greek upper case Α (a, “alpha”) that was the source for both common variants of ''a''", []*entityTestResult{
						{WikiEntityLinkInternalName, "Image:UncialA-01.svg", []*entityTestResult{}},
						{WikiEntityLinkInternalProp, "50px", []*entityTestResult{}},
						{WikiEntityLinkInternalProp, "Approximate form of Greek upper case Α (a, “alpha”) that was the source for both common variants of ''a''", []*entityTestResult{
							{WikiEntityTextItalic, "a", []*entityTestResult{}},
						}},
					}},
					{WikiEntityText, "\nModification of capital letter\n", []*entityTestResult{}},
				}},
			},
		},
		/***** 26 *****/
		{nil, `
# text 1 ''text
: text 2 '''text
* text 3 '''''text
`,
			[]*entityTestResult{
				{WikiEntityListNumbered, " text 1 ''text", []*entityTestResult{
					{WikiEntityTextItalic, "text", []*entityTestResult{}},
				}},
				{WikiEntityIndent, " text 2 '''text", []*entityTestResult{
					{WikiEntityTextBold, "text", []*entityTestResult{}},
				}},
				{WikiEntityListBulleted, " text 3 '''''text", []*entityTestResult{
					{WikiEntityTextBoldItalic, "text", []*entityTestResult{}},
				}},
				{WikiEntityText, "\n", []*entityTestResult{}},
			},
		},
	}

	for i, tc := range tests {
		//if i != 16 && i != 17 { continue }
		//if i != 18 && i != 19 { continue }
		if i != 26 { continue }

		//t.Logf("TestParse: %d", i)
		//t.Logf("TestParse: %d: %v", i, tc.src)

		var err error
		tc.wiki, err = ParseString(tc.src)
		if err != nil {
			t.Errorf("TestParse: [%d: error] %v", i, err)
			continue
		}
		if tc.wiki == nil {
			t.Errorf("TestParse: [%d: nil] %v", i, tc.src)
			continue
		}
		if tc.wiki.Type != WikiEntityWiki {
			t.Errorf("TestParse: [%d: Type] %v", i, tc.wiki.Type)
			continue
		}
		raw, text := []byte(tc.src), tc.src
		checkEntityResults(t, i, "TestParse", raw, text, tc.wiki, tc.entities, false)
	}
}

type entityChildTest struct {
	t EntityType
	raw string
	pos int
	children []*entityChildTest
}

func checkEntityChildren(t *testing.T, tag string, entity *Entity, children []*entityChildTest) {
	for ci, child := range entity.Entities {
		s := fmt.Sprintf("%s.%d", tag, ci)
		if l := len(children); l <= ci {
			t.Errorf("checkEntityChildren: %v: out of index (%v, %v, %v <= %v)", s, entity, child, l, ci)
			break
		}
		result := children[ci]
		if child.Type != result.t {
			t.Errorf("checkEntityChildren: %v: Entity.Type: %v != %v", s, child.Type, result.t)
			continue
		}
		if string(child.Raw) != result.raw {
			t.Errorf("checkEntityChildren: %v: Entity.Raw: %v != %v", s, string(child.Raw), result.raw)
			continue
		}
		if child.Pos != result.pos {
			t.Errorf("checkEntityChildren: %v: Entity.Pos: %v != %v (%v)", s, child.Pos, result.pos, child)
			continue
		}
		if l := len(entity.Raw); l <= result.pos {
			t.Errorf("checkEntityChildren: %v: Entity.Pos: %v <= %v (%v)", s, l, child.Pos, child)
			continue
		}
		l := len(child.Raw)
		x := entity.Raw[child.Pos:child.Pos+l]
		if string(x) != result.raw {
			t.Errorf("checkEntityChildren: %v: extraction: %v <= %v (%v)", s, string(x), result.raw, child)
			continue
		}
		checkEntityChildren(t, s, child, result.children)
	}
}

func TestEntityChildren(t *testing.T) {
	tests := []*entityChildTest{
		{
			WikiEntityWiki, `text {{name|prop1|prop2|prop3}} text`, 0,
			[]*entityChildTest{
				{WikiEntityText, "text ", 0, []*entityChildTest{}},
				{WikiEntityTemplate, "{{name|prop1|prop2|prop3}}", 5, []*entityChildTest{
					{WikiEntityTemplateName, "name", 2, []*entityChildTest{}},
					{WikiEntityTemplateProp, "|prop1", 6, []*entityChildTest{}},
					{WikiEntityTemplateProp, "|prop2", 12, []*entityChildTest{}},
					{WikiEntityTemplateProp, "|prop3", 18, []*entityChildTest{}},
				}},
				{WikiEntityText, " text", 31, []*entityChildTest{}},
			},
		},
		{
			WikiEntityWiki, `''text {{name|prop1|prop2|prop3{{name|prop}}}} text''`, 0,
			[]*entityChildTest{
				{WikiEntityTextItalic, "''text {{name|prop1|prop2|prop3{{name|prop}}}} text''", 0, []*entityChildTest{
					{WikiEntityTemplate, "{{name|prop1|prop2|prop3{{name|prop}}}}", 7, []*entityChildTest{
						{WikiEntityTemplateName, "name", 2, []*entityChildTest{}},
						{WikiEntityTemplateProp, "|prop1", 6, []*entityChildTest{}},
						{WikiEntityTemplateProp, "|prop2", 12, []*entityChildTest{}},
						{WikiEntityTemplateProp, "|prop3{{name|prop}}", 18, []*entityChildTest{
							{WikiEntityTemplate, "{{name|prop}}", 6, []*entityChildTest{
								{WikiEntityTemplateName, "name", 2, []*entityChildTest{}},
								{WikiEntityTemplateProp, "|prop", 6, []*entityChildTest{}},
							}},
						}},
					}},
				}},
			},
		},
		/*
		{
			WikiEntityWiki, `
==title==
''text {{name|prop1|prop2|prop3{{name|prop}}}} text''
`, 0,
			[]*entityChildTest{
				{WikiEntityHeading2, "==title==", 1, []*entityChildTest{
					{WikiEntityText, "\n", 10, []*entityChildTest{}},
					{WikiEntityTextBold, "''text {{name|prop1|prop2|prop3{{name|prop}}}} text''", 11, []*entityChildTest{
						{WikiEntityText, "text ", 0, []*entityChildTest{}},
						{WikiEntityTemplate, "{{name|prop1|prop2|prop3{{name|prop}}}}", 2, []*entityChildTest{
							{WikiEntityTemplate, "{{name|prop}}", 2, []*entityChildTest{
								{WikiEntityTemplateName, "name", 2, []*entityChildTest{}},
								{WikiEntityTemplateProp, "|prop", 6, []*entityChildTest{}},
							}},
						}},
						{WikiEntityText, " text", 0, []*entityChildTest{}},
					}},
				}},
			},
		}, */
	}
	for i, tc := range tests {
		wiki, err := ParseString(tc.raw)
		if err != nil {
			t.Errorf("TestEntityChildren: %d: %v", i, wiki)
		}
		wiki.Raw = []byte(tc.raw)
		wiki.Text = tc.raw
		checkEntityChildren(t, fmt.Sprintf("%d", i), wiki, tc.children)
	}
}

func checkTestDataSquare(t *testing.T, i int, data []byte, wiki *Entity) {
	results := []*entityTestResult{
		{WikiEntityHeading2, "English", []*entityTestResult{
			{WikiEntityText, "\n", []*entityTestResult{}},
			{WikiEntityTemplate, "wikipedia|dab=Square", []*entityTestResult{
				{WikiEntityTemplateName, "wikipedia", []*entityTestResult{}},
				{WikiEntityTemplateProp, "dab=Square", []*entityTestResult{}},
			}},
			{WikiEntityText, "\n", []*entityTestResult{}},
			{WikiEntityLinkInternal, "File:Square - black simple.svg|thumb|A square (polygon)", []*entityTestResult{}},
			{WikiEntityText, "\n", []*entityTestResult{}},
			{WikiEntityLinkInternal, "File:Komsomolskaya Square.JPG|thumb|Komsomolskaya Square at night", []*entityTestResult{}},
			{WikiEntityText, "\n", []*entityTestResult{}},
			{WikiEntityHeading3, "Etymology", []*entityTestResult{
			}},
			{WikiEntityHeading3, "Pronunciation", []*entityTestResult{
			}},
			{WikiEntityHeading3, "Noun", []*entityTestResult{
				/*
				{WikiEntityHeading4, "Synonyms", []*entityTestResult{
				}},
				{WikiEntityHeading4, "Derived terms", []*entityTestResult{
				}},
				{WikiEntityHeading4, "Translations", []*entityTestResult{
				}}, */
			}},
			{WikiEntityHeading3, "Adjective", []*entityTestResult{
				/*
				{WikiEntityHeading4, "Synonyms", []*entityTestResult{
				}},
				{WikiEntityHeading4, "Derived terms", []*entityTestResult{
				}},
				{WikiEntityHeading4, "Translations", []*entityTestResult{
				}}, */
			}},
			{WikiEntityHeading3, "Verb", []*entityTestResult{
				/*
				{WikiEntityHeading4, "Synonyms", []*entityTestResult{
				}},
				{WikiEntityHeading4, "Derived terms", []*entityTestResult{
				}},
				{WikiEntityHeading4, "Translations", []*entityTestResult{
				}}, */
			}},
			{WikiEntityHeading3, "See also", []*entityTestResult{
			}},
		}},
		{WikiEntityHeading2, "French", []*entityTestResult{
			/*
			{WikiEntityHeading3, "Etymology", []*entityTestResult{
			}},
			{WikiEntityHeading3, "Pronunciation", []*entityTestResult{
			}},
			{WikiEntityHeading3, "Noun", []*entityTestResult{
			}},
			{WikiEntityHeading3, "References", []*entityTestResult{
			}},
			{WikiEntityHeading3, "Anagrams", []*entityTestResult{
			}}, */
		}},
	}
	checkEntityResults(t, i, "checkTestDataSquare", data, string(data), wiki, results, true)
}

func checkTestDataWiki(t *testing.T, i int, data []byte, wiki *Entity) {
	results := []*entityTestResult{
		{WikiEntityTemplate, "also|Wiki", []*entityTestResult{
			{WikiEntityTemplateName, "also", []*entityTestResult{}},
			{WikiEntityTemplateProp, "Wiki", []*entityTestResult{}},
		}},
		{WikiEntityHeading2, "English", []*entityTestResult{
			{WikiEntityText, "\n", []*entityTestResult{}},
			{WikiEntityTemplate, "was wotd|2006|December|31", []*entityTestResult{
				{WikiEntityTemplateName, "was wotd", []*entityTestResult{}},
				{WikiEntityTemplateProp, "2006", []*entityTestResult{}},
				{WikiEntityTemplateProp, "December", []*entityTestResult{}},
				{WikiEntityTemplateProp, "31", []*entityTestResult{}},
			}},
			{WikiEntityText, "\n", []*entityTestResult{}},
			{WikiEntityTemplate, "wikipedia", []*entityTestResult{}},
			{WikiEntityText, "\n", []*entityTestResult{}},
			{WikiEntityHeading3, "Etymology", []*entityTestResult{
				{WikiEntityText, "\n1995.", []*entityTestResult{}},
				/*
				{WikiEntityTagBeg, "ref name=\"W. Cunningham, Correspondence on the Etymology of Wiki\"", []*entityTestResult{
					{WikiEntityTemplate, "cite web|url=http://c2.com/doc/etymology.html|title=Correspondence on the Etymology of Wiki|last=Cunningham|first=Ward|date=2005|publisher=Ward Cunningham|accessdate=28 February 2010", []*entityTestResult{
						{WikiEntityTemplateProp, "url=http://c2.com/doc/etymology.html", []*entityTestResult{}},
						{WikiEntityTemplateProp, "title=Correspondence on the Etymology of Wiki", []*entityTestResult{}},
						{WikiEntityTemplateProp, "last=Cunningham", []*entityTestResult{}},
						{WikiEntityTemplateProp, "first=Ward", []*entityTestResult{}},
						{WikiEntityTemplateProp, "date=2005", []*entityTestResult{}},
						{WikiEntityTemplateProp, "publisher=Ward Cunningham", []*entityTestResult{}},
						{WikiEntityTemplateProp, "accessdate=28 February 2010", []*entityTestResult{}},
					}},
					{WikiEntityTagEnd, "ref", []*entityTestResult{}},
				}}, */
				{WikiEntityTagBeg, "ref name=\"W. Cunningham, Correspondence on the Etymology of Wiki\"", []*entityTestResult{}},
				{WikiEntityTemplate, "cite web|url=http://c2.com/doc/etymology.html|title=Correspondence on the Etymology of Wiki|last=Cunningham|first=Ward|date=2005|publisher=Ward Cunningham|accessdate=28 February 2010", []*entityTestResult{
					{WikiEntityTemplateName, "cite web", []*entityTestResult{}},
					{WikiEntityTemplateProp, "url=http://c2.com/doc/etymology.html", []*entityTestResult{}},
					{WikiEntityTemplateProp, "title=Correspondence on the Etymology of Wiki", []*entityTestResult{}},
					{WikiEntityTemplateProp, "last=Cunningham", []*entityTestResult{}},
					{WikiEntityTemplateProp, "first=Ward", []*entityTestResult{}},
					{WikiEntityTemplateProp, "date=2005", []*entityTestResult{}},
					{WikiEntityTemplateProp, "publisher=Ward Cunningham", []*entityTestResult{}},
					{WikiEntityTemplateProp, "accessdate=28 February 2010", []*entityTestResult{}},
				}},
				{WikiEntityTagEnd, "ref", []*entityTestResult{}},
				{WikiEntityText, " Abbreviated from ", []*entityTestResult{}},
				{WikiEntityLinkInternal, "WikiWikiWeb", []*entityTestResult{}},
				{WikiEntityText, ", from ", []*entityTestResult{}},
				{WikiEntityTemplate, "etyl|haw|en", []*entityTestResult{
					{WikiEntityTemplateName, "etyl", []*entityTestResult{}},
					{WikiEntityTemplateProp, "haw", []*entityTestResult{}},
					{WikiEntityTemplateProp, "en", []*entityTestResult{}},
				}},
				{WikiEntityText, " ", []*entityTestResult{}},
				{WikiEntityTemplate, "term|wikiwiki||quick|lang=haw", []*entityTestResult{
					{WikiEntityTemplateName, "term", []*entityTestResult{}},
					{WikiEntityTemplateProp, "wikiwiki", []*entityTestResult{}},
					{WikiEntityTemplateProp, "", []*entityTestResult{}},
					{WikiEntityTemplateProp, "quick", []*entityTestResult{}},
					{WikiEntityTemplateProp, "lang=haw", []*entityTestResult{}},
				}},
				{WikiEntityText, " + ", []*entityTestResult{}},
				{WikiEntityTemplate, "etyl|en|-", []*entityTestResult{
					{WikiEntityTemplateName, "etyl", []*entityTestResult{}},
					{WikiEntityTemplateProp, "en", []*entityTestResult{}},
					{WikiEntityTemplateProp, "-", []*entityTestResult{}},
				}},
				{WikiEntityText, " ", []*entityTestResult{}},
				{WikiEntityTemplate, "term|web|lang=en", []*entityTestResult{
					{WikiEntityTemplateName, "term", []*entityTestResult{}},
					{WikiEntityTemplateProp, "web", []*entityTestResult{}},
					{WikiEntityTemplateProp, "lang=en", []*entityTestResult{}},
				}},
				{WikiEntityText, ".\n", []*entityTestResult{}},
			}},
			{WikiEntityHeading3, "Pronunciation", []*entityTestResult{
				{WikiEntityListBulleted, " {{enPR|wĭʹkē|wēʹkē}}, {{IPA|/ˈwɪki/|/ˈwiːki/}}, {{X-SAMPA|/\"wIki/|/\"wi:ki/}}", []*entityTestResult{
					{WikiEntityTemplate, "enPR|wĭʹkē|wēʹkē", []*entityTestResult{
						{WikiEntityTemplateName, "enPR", []*entityTestResult{}},
						{WikiEntityTemplateProp, "wĭʹkē", []*entityTestResult{}},
						{WikiEntityTemplateProp, "wēʹkē", []*entityTestResult{}},
					}},
					{WikiEntityTemplate, "IPA|/ˈwɪki/|/ˈwiːki/", []*entityTestResult{
						{WikiEntityTemplateName, "IPA", []*entityTestResult{}},
						{WikiEntityTemplateProp, "/ˈwɪki/", []*entityTestResult{}},
						{WikiEntityTemplateProp, "/ˈwiːki/", []*entityTestResult{}},
					}},
					{WikiEntityTemplate, "X-SAMPA|/\"wIki/|/\"wi:ki/", []*entityTestResult{
						{WikiEntityTemplateName, "X-SAMPA", []*entityTestResult{}},
						{WikiEntityTemplateProp, "/\"wIki/", []*entityTestResult{}},
						{WikiEntityTemplateProp, "/\"wi:ki/", []*entityTestResult{}},
					}},
				}},
				{WikiEntityListBulleted, " {{audio|en-us-wiki.ogg|Audio (US)}}", []*entityTestResult{
					{WikiEntityTemplate, "audio|en-us-wiki.ogg|Audio (US)", []*entityTestResult{
						{WikiEntityTemplateName, "audio", []*entityTestResult{}},
						{WikiEntityTemplateProp, "en-us-wiki.ogg", []*entityTestResult{}},
						{WikiEntityTemplateProp, "Audio (US)", []*entityTestResult{}},
					}},
				}},
				{WikiEntityListBulleted, " {{rhymes|ɪki|iːki}}", []*entityTestResult{
					{WikiEntityTemplate, "rhymes|ɪki|iːki", []*entityTestResult{
						{WikiEntityTemplateName, "rhymes", []*entityTestResult{}},
						{WikiEntityTemplateProp, "ɪki", []*entityTestResult{}},
						{WikiEntityTemplateProp, "iːki", []*entityTestResult{}},
					}},
				}},
				{WikiEntityListBulleted, " {{homophones|lang=en|wicky}}", []*entityTestResult{
					{WikiEntityTemplate, "homophones|lang=en|wicky", []*entityTestResult{
						{WikiEntityTemplateName, "homophones", []*entityTestResult{}},
						{WikiEntityTemplateProp, "lang=en", []*entityTestResult{}},
						{WikiEntityTemplateProp, "wicky", []*entityTestResult{}},
					}},
				}},
				{WikiEntityText, "\n", []*entityTestResult{}},
			}},
			{WikiEntityHeading3, "Noun", []*entityTestResult{
				{WikiEntityText, "\n", []*entityTestResult{}},
				{WikiEntityTemplate, "en-noun", []*entityTestResult{}},
				{WikiEntityText, "\n", []*entityTestResult{}},
				{WikiEntityListNumbered, " A [[collaborative]] [[website]] which can be directly [[edit]]ed merely using a web browser, often by anyone with access to it.", []*entityTestResult{
					{WikiEntityLinkInternal, "collaborative", []*entityTestResult{}},
					{WikiEntityLinkInternal, "website", []*entityTestResult{}},
					{WikiEntityLinkInternal, "edit", []*entityTestResult{}},
				}},
				{WikiEntityText, "\n", []*entityTestResult{}},
				{WikiEntityHeading4, "Translations", []*entityTestResult{
					{WikiEntityText, "\n", []*entityTestResult{}},
					{WikiEntityTemplate, "trans-top|collaborative website", []*entityTestResult{
						{WikiEntityTemplateName, "trans-top", []*entityTestResult{}},
						{WikiEntityTemplateProp, "collaborative website", []*entityTestResult{}},
					}},
					{WikiEntityListBulleted, " Afrikaans: {{t-|af|wiki}}", []*entityTestResult{
						{WikiEntityTemplate, "t-|af|wiki", []*entityTestResult{
							{WikiEntityTemplateName, "t-", []*entityTestResult{}},
							{WikiEntityTemplateProp, "af", []*entityTestResult{}},
							{WikiEntityTemplateProp, "wiki", []*entityTestResult{}},
						}},
					}},
					{WikiEntityListBulleted, " Arabic: {{t+|ar|ويكي|tr=wiki|sc=Arab}}", []*entityTestResult{
						{WikiEntityTemplate, "t+|ar|ويكي|tr=wiki|sc=Arab", []*entityTestResult{
							{WikiEntityTemplateName, "t+", []*entityTestResult{}},
							{WikiEntityTemplateProp, "ar", []*entityTestResult{}},
							{WikiEntityTemplateProp, "ويكي", []*entityTestResult{}},
							{WikiEntityTemplateProp, "tr=wiki", []*entityTestResult{}},
							{WikiEntityTemplateProp, "sc=Arab", []*entityTestResult{}},
						}},
					}},
					{WikiEntityListBulleted, " Armenian: {{t-|hy|վիքի|tr=vik'i}}", []*entityTestResult{
						{WikiEntityTemplate, "t-|hy|վիքի|tr=vik'i", []*entityTestResult{
							{WikiEntityTemplateName, "t-", []*entityTestResult{}},
							{WikiEntityTemplateProp, "hy", []*entityTestResult{}},
							{WikiEntityTemplateProp, "վիքի", []*entityTestResult{}},
							{WikiEntityTemplateProp, "tr=vik'i", []*entityTestResult{}},
						}},
					}},
					{WikiEntityListBulleted, " Catalan: {{t+|ca|wiki}}", []*entityTestResult{
						{WikiEntityTemplate, "t+|ca|wiki", []*entityTestResult{
							{WikiEntityTemplateName, "t+", []*entityTestResult{}},
							{WikiEntityTemplateProp, "ca", []*entityTestResult{}},
							{WikiEntityTemplateProp, "wiki", []*entityTestResult{}},
						}},
					}},
					{WikiEntityListBulleted, " Chinese:", []*entityTestResult{}},
					{WikiEntityListBulleted, ": Mandarin: {{t|cmn|維基}}, {{t|cmn|维基|tr=wéijī}}", []*entityTestResult{
						{WikiEntityIndent, " Mandarin: {{t|cmn|維基}}, {{t|cmn|维基|tr=wéijī}}", []*entityTestResult{
							{WikiEntityTemplate, "t|cmn|維基", []*entityTestResult{
								{WikiEntityTemplateName, "t", []*entityTestResult{}},
								{WikiEntityTemplateProp, "cmn", []*entityTestResult{}},
								{WikiEntityTemplateProp, "維基", []*entityTestResult{}},
							}},
							{WikiEntityTemplate, "t|cmn|维基|tr=wéijī", []*entityTestResult{
								{WikiEntityTemplateName, "t", []*entityTestResult{}},
								{WikiEntityTemplateProp, "cmn", []*entityTestResult{}},
								{WikiEntityTemplateProp, "维基", []*entityTestResult{}},
								{WikiEntityTemplateProp, "tr=wéijī", []*entityTestResult{}},
							}},
						}},
					}},
					{WikiEntityListBulleted, " Danish: {{t+|da|wiki}}", []*entityTestResult{
						{WikiEntityTemplate, "t+|da|wiki", []*entityTestResult{
							{WikiEntityTemplateName, "t+", []*entityTestResult{}},
							{WikiEntityTemplateProp, "da", []*entityTestResult{}},
							{WikiEntityTemplateProp, "wiki", []*entityTestResult{}},
						}},
					}},
					{WikiEntityListBulleted, " Dutch: {{t+|nl|wiki}}", []*entityTestResult{
						{WikiEntityTemplate, "t+|nl|wiki", []*entityTestResult{
							{WikiEntityTemplateName, "t+", []*entityTestResult{}},
							{WikiEntityTemplateProp, "nl", []*entityTestResult{}},
							{WikiEntityTemplateProp, "wiki", []*entityTestResult{}},
						}},
					}},
					{WikiEntityListBulleted, " Esperanto: {{t+|eo|vikio}}", []*entityTestResult{}},
					{WikiEntityListBulleted, " Estonian: {{t-|et|viki}}", []*entityTestResult{}},
					{WikiEntityListBulleted, " Finnish: {{t+|fi|wiki}}", []*entityTestResult{}},
					{WikiEntityListBulleted, " French: {{t+|fr|wiki|m}}", []*entityTestResult{}},
					{WikiEntityListBulleted, " Georgian: {{t-|ka|ვიკი|tr=viki|sc=Geor}}", []*entityTestResult{
						{WikiEntityTemplate, "t-|ka|ვიკი|tr=viki|sc=Geor", []*entityTestResult{
							{WikiEntityTemplateName, "t-", []*entityTestResult{}},
							{WikiEntityTemplateProp, "ka", []*entityTestResult{}},
							{WikiEntityTemplateProp, "ვიკი", []*entityTestResult{}},
							{WikiEntityTemplateProp, "tr=viki", []*entityTestResult{}},
							{WikiEntityTemplateProp, "sc=Geor", []*entityTestResult{}},
						}},
					}},
					{WikiEntityListBulleted, " German: {{t+|de|Wiki|n}}", []*entityTestResult{
						{WikiEntityTemplate, "t+|de|Wiki|n", []*entityTestResult{
							{WikiEntityTemplateName, "t+", []*entityTestResult{}},
							{WikiEntityTemplateProp, "de", []*entityTestResult{}},
							{WikiEntityTemplateProp, "Wiki", []*entityTestResult{}},
							{WikiEntityTemplateProp, "n", []*entityTestResult{}},
						}},
					}},
					{WikiEntityListBulleted, " Greek: {{t+|el|βίκι|n|tr=víki}}", []*entityTestResult{
						{WikiEntityTemplate, "t+|el|βίκι|n|tr=víki", []*entityTestResult{
							{WikiEntityTemplateName, "t+", []*entityTestResult{}},
							{WikiEntityTemplateProp, "el", []*entityTestResult{}},
							{WikiEntityTemplateProp, "βίκι", []*entityTestResult{}},
							{WikiEntityTemplateProp, "n", []*entityTestResult{}},
							{WikiEntityTemplateProp, "tr=víki", []*entityTestResult{}},
						}},
					}},
					{WikiEntityListBulleted, " Hebrew: {{t+|he|ויקי|tr=wiki}}", []*entityTestResult{}},
					{WikiEntityListBulleted, " Hungarian: {{t+|hu|viki}}", []*entityTestResult{}},
					{WikiEntityListBulleted, " Interlingua: {{t-|ia|wiki}}", []*entityTestResult{}},
					{WikiEntityListBulleted, " Italian: {{t+|it|wiki}}", []*entityTestResult{}},
					{WikiEntityListBulleted, " Japanese: {{t+|ja|ウィキ|tr=wiki|sc=Jpan}}", []*entityTestResult{
						{WikiEntityTemplate, "t+|ja|ウィキ|tr=wiki|sc=Jpan", []*entityTestResult{
							{WikiEntityTemplateName, "t+", []*entityTestResult{}},
							{WikiEntityTemplateProp, "ja", []*entityTestResult{}},
							{WikiEntityTemplateProp, "ウィキ", []*entityTestResult{}},
							{WikiEntityTemplateProp, "tr=wiki", []*entityTestResult{}},
							{WikiEntityTemplateProp, "sc=Jpan", []*entityTestResult{}},
						}},
					}},
					{WikiEntityText, "\n", []*entityTestResult{}},
					{WikiEntityTemplate, "trans-mid", []*entityTestResult{}},
					{WikiEntityListBulleted, " Khmer: {{t-|km|វិគី|tr=vikī|sc=Khmr}}", []*entityTestResult{
						{WikiEntityTemplate, "t-|km|វិគី|tr=vikī|sc=Khmr", []*entityTestResult{
							{WikiEntityTemplateName, "t-", []*entityTestResult{}},
							{WikiEntityTemplateProp, "km", []*entityTestResult{}},
							{WikiEntityTemplateProp, "វិគី", []*entityTestResult{}},
							{WikiEntityTemplateProp, "tr=vikī", []*entityTestResult{}},
							{WikiEntityTemplateProp, "sc=Khmr", []*entityTestResult{}},
						}},
					}},
					{WikiEntityListBulleted, " Korean: {{t+|ko|위키|tr=wiki|sc=Kore}}", []*entityTestResult{
						{WikiEntityTemplate, "t+|ko|위키|tr=wiki|sc=Kore", []*entityTestResult{
							{WikiEntityTemplateName, "t+", []*entityTestResult{}},
							{WikiEntityTemplateProp, "ko", []*entityTestResult{}},
							{WikiEntityTemplateProp, "위키", []*entityTestResult{}},
							{WikiEntityTemplateProp, "tr=wiki", []*entityTestResult{}},
							{WikiEntityTemplateProp, "sc=Kore", []*entityTestResult{}},
						}},
					}},
					{WikiEntityListBulleted, " Lithuanian: {{t+|lt|viki}}", []*entityTestResult{}},
					{WikiEntityListBulleted, " Luxembourgish: {{t-|lb|Wiki|n}}", []*entityTestResult{}},
					{WikiEntityListBulleted, " Macedonian: {{t-|mk|вики|tr=víki}}", []*entityTestResult{}},
					{WikiEntityListBulleted, " Malay: {{t-|ms|wiki}}", []*entityTestResult{}},
					{WikiEntityListBulleted, " Norwegian: {{t+|no|wiki}}", []*entityTestResult{}},
					{WikiEntityListBulleted, " Persian: {{t-|fa|ویکی|tr=viki|sc=fa-Arab}}", []*entityTestResult{
						{WikiEntityTemplate, "t-|fa|ویکی|tr=viki|sc=fa-Arab", []*entityTestResult{
							{WikiEntityTemplateName, "t-", []*entityTestResult{}},
							{WikiEntityTemplateProp, "fa", []*entityTestResult{}},
							{WikiEntityTemplateProp, "ویکی", []*entityTestResult{}},
							{WikiEntityTemplateProp, "tr=viki", []*entityTestResult{}},
							{WikiEntityTemplateProp, "sc=fa-Arab", []*entityTestResult{}},
						}},
					}},
					{WikiEntityListBulleted, " Polish: {{t+|pl|wiki}}", []*entityTestResult{}},
					{WikiEntityListBulleted, " Punjabi: {{t-|pa|ਵਿਕਿ|tr=wiki|sc=Guru}}", []*entityTestResult{
						{WikiEntityTemplate, "t-|pa|ਵਿਕਿ|tr=wiki|sc=Guru", []*entityTestResult{
							{WikiEntityTemplateName, "t-", []*entityTestResult{}},
							{WikiEntityTemplateProp, "pa", []*entityTestResult{}},
							{WikiEntityTemplateProp, "ਵਿਕਿ", []*entityTestResult{}},
							{WikiEntityTemplateProp, "tr=wiki", []*entityTestResult{}},
							{WikiEntityTemplateProp, "sc=Guru", []*entityTestResult{}},
						}},
					}},
					{WikiEntityListBulleted, " Portuguese: {{t+|pt|wiki}}", []*entityTestResult{}},
					{WikiEntityListBulleted, " Romanian: {{t+|ro|wiki}}", []*entityTestResult{}},
					{WikiEntityListBulleted, " Russian: {{t+|ru|вики|tr=víki}}", []*entityTestResult{
						{WikiEntityTemplate, "t+|ru|вики|tr=víki", []*entityTestResult{
							{WikiEntityTemplateName, "t+", []*entityTestResult{}},
							{WikiEntityTemplateProp, "ru", []*entityTestResult{}},
							{WikiEntityTemplateProp, "вики", []*entityTestResult{}},
							{WikiEntityTemplateProp, "tr=víki", []*entityTestResult{}},
						}},
					}},
					{WikiEntityListBulleted, " Serbo-Croatian:", []*entityTestResult{}},
					{WikiEntityListBulleted, ": Cyrillic: {{t-|sh|вики|sc=Cyrl}}", []*entityTestResult{
						{WikiEntityIndent, " Cyrillic: {{t-|sh|вики|sc=Cyrl}}", []*entityTestResult{
							{WikiEntityTemplate, "t-|sh|вики|sc=Cyrl", []*entityTestResult{
								{WikiEntityTemplateName, "t-", []*entityTestResult{}},
								{WikiEntityTemplateProp, "sh", []*entityTestResult{}},
								{WikiEntityTemplateProp, "вики", []*entityTestResult{}},
								{WikiEntityTemplateProp, "sc=Cyrl", []*entityTestResult{}},
							}},
						}},
					}},
					{WikiEntityListBulleted, ": Roman: {{t-|sh|viki}}", []*entityTestResult{
						{WikiEntityIndent, " Roman: {{t-|sh|viki}}", []*entityTestResult{
							{WikiEntityTemplate, "t-|sh|viki", []*entityTestResult{
								{WikiEntityTemplateName, "t-", []*entityTestResult{}},
								{WikiEntityTemplateProp, "sh", []*entityTestResult{}},
								{WikiEntityTemplateProp, "viki", []*entityTestResult{}},
							}},
						}},
					}},
					{WikiEntityListBulleted, " Spanish: {{t+|es|wiki}}", []*entityTestResult{}},
					{WikiEntityListBulleted, " Swedish: {{t+|sv|wiki}}", []*entityTestResult{}},
					{WikiEntityListBulleted, " Turkish: {{t+|tr|viki}}", []*entityTestResult{}},
					{WikiEntityListBulleted, " Volapük: {{t-|vo|vük}}", []*entityTestResult{}},
					{WikiEntityListBulleted, " Võro: {{tø|vro|viki}}", []*entityTestResult{
						{WikiEntityTemplate, "tø|vro|viki", []*entityTestResult{
							{WikiEntityTemplateName, "tø", []*entityTestResult{}},
							{WikiEntityTemplateProp, "vro", []*entityTestResult{}},
							{WikiEntityTemplateProp, "viki", []*entityTestResult{}},
						}},
					}},
					{WikiEntityListBulleted, " Welsh: {{t+|cy|wici}}", []*entityTestResult{
						{WikiEntityTemplate, "t+|cy|wici", []*entityTestResult{
							{WikiEntityTemplateName, "t+", []*entityTestResult{}},
							{WikiEntityTemplateProp, "cy", []*entityTestResult{}},
							{WikiEntityTemplateProp, "wici", []*entityTestResult{}},
						}},
					}},
					{WikiEntityText, "\n", []*entityTestResult{}},
					{WikiEntityTemplate, "trans-bottom", []*entityTestResult{}},
					{WikiEntityText, "\n", []*entityTestResult{}},
				}},
				{WikiEntityHeading4, "Derived terms", []*entityTestResult{
					{WikiEntityListBulleted, " [[interwiki]]", []*entityTestResult{
						{WikiEntityLinkInternal, "interwiki", []*entityTestResult{}},
					}},
					{WikiEntityText, "\n", []*entityTestResult{}},
				}},
			}},
			{WikiEntityHeading3, "Verb", []*entityTestResult{
				{WikiEntityText, "\n", []*entityTestResult{}},
				{WikiEntityTemplate, "en-verb|wiki", []*entityTestResult{
					{WikiEntityTemplateName, "en-verb", []*entityTestResult{}},
					{WikiEntityTemplateProp, "wiki", []*entityTestResult{}},
				}},
				{WikiEntityText, "\n", []*entityTestResult{}},
				{WikiEntityListNumbered, " {{context|transitive|lang=en}} To [[research]] on [[Wikipedia]] or some similar wiki.", []*entityTestResult{
					{WikiEntityTemplate, "context|transitive|lang=en", []*entityTestResult{
						{WikiEntityTemplateName, "context", []*entityTestResult{}},
						{WikiEntityTemplateProp, "transitive", []*entityTestResult{}},
						{WikiEntityTemplateProp, "lang=en", []*entityTestResult{}},
					}},
					{WikiEntityLinkInternal, "research", []*entityTestResult{}},
					{WikiEntityLinkInternal, "Wikipedia", []*entityTestResult{}},
				}},
				{WikiEntityListNumbered, ": ''To get an understanding of the topics, he quickly went online and '''wikied''' each one.''", []*entityTestResult{
					{WikiEntityIndent, " ''To get an understanding of the topics, he quickly went online and '''wikied''' each one.''", []*entityTestResult{
						{WikiEntityTextItalic, `To get an understanding of the topics, he quickly went online and '''wikied''' each one.`, []*entityTestResult{
							{WikiEntityTextBold, `wikied`, []*entityTestResult{}},
						}},
					}},
				}},
				{WikiEntityListNumbered, "* {{quote-news|title=Son of a Geek: Comics and Growing Up the DC Way|author=GeekDad|work=Wired News|date=December 1|year=2008|passage=I tore through his collection '''wikiing''' any plot points that I missed learning the importance of the players of the DC universe}}", []*entityTestResult{
					{WikiEntityListBulleted, " {{quote-news|title=Son of a Geek: Comics and Growing Up the DC Way|author=GeekDad|work=Wired News|date=December 1|year=2008|passage=I tore through his collection '''wikiing''' any plot points that I missed learning the importance of the players of the DC universe}}", []*entityTestResult{
						{WikiEntityTemplate, "quote-news|title=Son of a Geek: Comics and Growing Up the DC Way|author=GeekDad|work=Wired News|date=December 1|year=2008|passage=I tore through his collection '''wikiing''' any plot points that I missed learning the importance of the players of the DC universe", []*entityTestResult{
							{WikiEntityTemplateName, "quote-news", []*entityTestResult{}},
							{WikiEntityTemplateProp, "title=Son of a Geek: Comics and Growing Up the DC Way", []*entityTestResult{}},
							{WikiEntityTemplateProp, "author=GeekDad", []*entityTestResult{}},
							{WikiEntityTemplateProp, "work=Wired News", []*entityTestResult{}},
							{WikiEntityTemplateProp, "date=December 1", []*entityTestResult{}},
							{WikiEntityTemplateProp, "year=2008", []*entityTestResult{}},
							{WikiEntityTemplateProp, "passage=I tore through his collection '''wikiing''' any plot points that I missed learning the importance of the players of the DC universe", []*entityTestResult{
								{WikiEntityTextBold, `wikiing`, []*entityTestResult{}},
							}},
						}},
					}},
				}},
				{WikiEntityListNumbered, "* {{quote-newsgroup|title=Janus|newsgroup=uk.rec.sheds|date=June 18|year=2009|passage=Her English is no better than my Portuguese, but I '''wikied''' 'influenza' in Portuguese and it came up with 'gripe'|author=Lizz Holmans|url=http://groups.google.com/group/uk.rec.sheds/browse_thread/thread/dfbb1b1c19b06f9b/25af2ce4e2298842?hl=en&ie=UTF-8&q=wikied|wikiing++-india#25af2ce4e2298842}}", []*entityTestResult{
					{WikiEntityListBulleted, " {{quote-newsgroup|title=Janus|newsgroup=uk.rec.sheds|date=June 18|year=2009|passage=Her English is no better than my Portuguese, but I '''wikied''' 'influenza' in Portuguese and it came up with 'gripe'|author=Lizz Holmans|url=http://groups.google.com/group/uk.rec.sheds/browse_thread/thread/dfbb1b1c19b06f9b/25af2ce4e2298842?hl=en&ie=UTF-8&q=wikied|wikiing++-india#25af2ce4e2298842}}", []*entityTestResult{
						{WikiEntityTemplate, "quote-newsgroup|title=Janus|newsgroup=uk.rec.sheds|date=June 18|year=2009|passage=Her English is no better than my Portuguese, but I '''wikied''' 'influenza' in Portuguese and it came up with 'gripe'|author=Lizz Holmans|url=http://groups.google.com/group/uk.rec.sheds/browse_thread/thread/dfbb1b1c19b06f9b/25af2ce4e2298842?hl=en&ie=UTF-8&q=wikied|wikiing++-india#25af2ce4e2298842", []*entityTestResult{
							{WikiEntityTemplateName, "quote-newsgroup", []*entityTestResult{}},
							{WikiEntityTemplateProp, "title=Janus", []*entityTestResult{}},
							{WikiEntityTemplateProp, "newsgroup=uk.rec.sheds", []*entityTestResult{}},
							{WikiEntityTemplateProp, "date=June 18", []*entityTestResult{}},
							{WikiEntityTemplateProp, "year=2009", []*entityTestResult{}},
							{WikiEntityTemplateProp, "passage=Her English is no better than my Portuguese, but I '''wikied''' 'influenza' in Portuguese and it came up with 'gripe'", []*entityTestResult{
								{WikiEntityTextBold, `wikied`, []*entityTestResult{}},
							}},
							{WikiEntityTemplateProp, "author=Lizz Holmans", []*entityTestResult{}},
							{WikiEntityTemplateProp, "url=http://groups.google.com/group/uk.rec.sheds/browse_thread/thread/dfbb1b1c19b06f9b/25af2ce4e2298842?hl=en&ie=UTF-8&q=wikied", []*entityTestResult{}},
							{WikiEntityTemplateProp, "wikiing++-india#25af2ce4e2298842", []*entityTestResult{}},
						}},						
					}},
				}},
				{WikiEntityListNumbered, "* {{quote-book|title=Journey|page=65|author=Noemi Gonzalez|year=2010|passage=I did research on the internet and found out so. I “'''wikied'''” it.}}", []*entityTestResult{
					{WikiEntityListBulleted, " {{quote-book|title=Journey|page=65|author=Noemi Gonzalez|year=2010|passage=I did research on the internet and found out so. I “'''wikied'''” it.}}", []*entityTestResult{
						{WikiEntityTemplate, "quote-book|title=Journey|page=65|author=Noemi Gonzalez|year=2010|passage=I did research on the internet and found out so. I “'''wikied'''” it.", []*entityTestResult{
							{WikiEntityTemplateName, "quote-book", []*entityTestResult{}},
							{WikiEntityTemplateProp, "title=Journey", []*entityTestResult{}},
							{WikiEntityTemplateProp, "page=65", []*entityTestResult{}},
							{WikiEntityTemplateProp, "author=Noemi Gonzalez", []*entityTestResult{}},
							{WikiEntityTemplateProp, "year=2010", []*entityTestResult{}},
							{WikiEntityTemplateProp, "passage=I did research on the internet and found out so. I “'''wikied'''” it.", []*entityTestResult{
								{WikiEntityTextBold, "wikied", []*entityTestResult{}},
							}},
						}},
					}},
				}},
				{WikiEntityListNumbered, " {{context|intransitive|lang=en}} To conduct research on a wiki.", []*entityTestResult{}},
				{WikiEntityListNumbered, " {{context|intransitive|lang=en}} To [[contribute]] to a wiki.", []*entityTestResult{}},
				{WikiEntityListNumbered, "* {{quote-book|title=Deptford.TV Diaries|page=73|author=Deptford Tv|year=2006|passage=Blogging, '''wiki-ing''', coding are all activities that generate authorial product.}}", []*entityTestResult{
					{WikiEntityListBulleted, " {{quote-book|title=Deptford.TV Diaries|page=73|author=Deptford Tv|year=2006|passage=Blogging, '''wiki-ing''', coding are all activities that generate authorial product.}}", []*entityTestResult{
						{WikiEntityTemplate, "quote-book|title=Deptford.TV Diaries|page=73|author=Deptford Tv|year=2006|passage=Blogging, '''wiki-ing''', coding are all activities that generate authorial product.", []*entityTestResult{
							{WikiEntityTemplateName, "quote-book", []*entityTestResult{}},
							{WikiEntityTemplateProp, "title=Deptford.TV Diaries", []*entityTestResult{}},
							{WikiEntityTemplateProp, "page=73", []*entityTestResult{}},
							{WikiEntityTemplateProp, "author=Deptford Tv", []*entityTestResult{}},
							{WikiEntityTemplateProp, "year=2006", []*entityTestResult{}},
							{WikiEntityTemplateProp, "passage=Blogging, '''wiki-ing''', coding are all activities that generate authorial product.", []*entityTestResult{
								{WikiEntityTextBold, "wiki-ing", []*entityTestResult{}},
							}},
						}},
					}},
				}},
				{WikiEntityListNumbered, "* {{quote-book|title=Wikis for dummies|page=17|author=Dan Woods|co-author=Peter Thoeny|year=2007|passage=The best way to start '''wiki-ing''' is to find an existing wiki (that is, a hosted wiki) and start adding to it.}}", []*entityTestResult{
					{WikiEntityListBulleted, " {{quote-book|title=Wikis for dummies|page=17|author=Dan Woods|co-author=Peter Thoeny|year=2007|passage=The best way to start '''wiki-ing''' is to find an existing wiki (that is, a hosted wiki) and start adding to it.}}", []*entityTestResult{
						{WikiEntityTemplate, "quote-book|title=Wikis for dummies|page=17|author=Dan Woods|co-author=Peter Thoeny|year=2007|passage=The best way to start '''wiki-ing''' is to find an existing wiki (that is, a hosted wiki) and start adding to it.", []*entityTestResult{
							{WikiEntityTemplateName, "quote-book", []*entityTestResult{}},
							{WikiEntityTemplateProp, "title=Wikis for dummies", []*entityTestResult{}},
							{WikiEntityTemplateProp, "page=17", []*entityTestResult{}},
							{WikiEntityTemplateProp, "author=Dan Woods", []*entityTestResult{}},
							{WikiEntityTemplateProp, "co-author=Peter Thoeny", []*entityTestResult{}},
							{WikiEntityTemplateProp, "year=2007", []*entityTestResult{}},
							{WikiEntityTemplateProp, "passage=The best way to start '''wiki-ing''' is to find an existing wiki (that is, a hosted wiki) and start adding to it.", []*entityTestResult{
								{WikiEntityTextBold, "wiki-ing", []*entityTestResult{}},
							}},
						}},
					}},
				}},
				{WikiEntityListNumbered, "* {{quote-book|title=Wiki writing: collaborative learning in the college classroom|page=46|author=Robert E. Cummings|coauthors=Matt Barton|year=2008|passage=For example, blog and wiki software can be used to support all sorts of activities that are not commonly associated with the activities of “blogging” or “'''wikiing'''.” This includes activities like sharing syllabi, publishing announcements}}", []*entityTestResult{}},
				{WikiEntityListNumbered, " {{context|transitive|lang=en}} To participate in the wiki-based production of.", []*entityTestResult{
					{WikiEntityTemplate, "context|transitive|lang=en", []*entityTestResult{
						{WikiEntityTemplateName, "context", []*entityTestResult{}},
						{WikiEntityTemplateProp, "transitive", []*entityTestResult{}},
						{WikiEntityTemplateProp, "lang=en", []*entityTestResult{}},
					}},
				}},
				{WikiEntityListNumbered, "* {{quote-journal|journal=Time|title=Cooking Consensus: Will Wiki Work in the Kitchen?|date=October 19|year=2009|passage=The history of '''wikied''' novels isn't pretty (Penguin Books never published the gobbledygook that was \"A Million Penguins\"), and no one has dared '''wiki''' a jazz song.|url=http://www.time.com/time/magazine/article/0,9171,1929212,00.html?iid=tsmodule}}", []*entityTestResult{
					{WikiEntityListBulleted, " {{quote-journal|journal=Time|title=Cooking Consensus: Will Wiki Work in the Kitchen?|date=October 19|year=2009|passage=The history of '''wikied''' novels isn't pretty (Penguin Books never published the gobbledygook that was \"A Million Penguins\"), and no one has dared '''wiki''' a jazz song.|url=http://www.time.com/time/magazine/article/0,9171,1929212,00.html?iid=tsmodule}}", []*entityTestResult{
						{WikiEntityTemplate, "quote-journal|journal=Time|title=Cooking Consensus: Will Wiki Work in the Kitchen?|date=October 19|year=2009|passage=The history of '''wikied''' novels isn't pretty (Penguin Books never published the gobbledygook that was \"A Million Penguins\"), and no one has dared '''wiki''' a jazz song.|url=http://www.time.com/time/magazine/article/0,9171,1929212,00.html?iid=tsmodule", []*entityTestResult{
							{WikiEntityTemplateName, "quote-journal", []*entityTestResult{}},
							{WikiEntityTemplateProp, "journal=Time", []*entityTestResult{}},
							{WikiEntityTemplateProp, "title=Cooking Consensus: Will Wiki Work in the Kitchen?", []*entityTestResult{}},
							{WikiEntityTemplateProp, "date=October 19", []*entityTestResult{}},
							{WikiEntityTemplateProp, "year=2009", []*entityTestResult{}},
							{WikiEntityTemplateProp, "passage=The history of '''wikied''' novels isn't pretty (Penguin Books never published the gobbledygook that was \"A Million Penguins\"), and no one has dared '''wiki''' a jazz song.", []*entityTestResult{}},
							{WikiEntityTemplateProp, "url=http://www.time.com/time/magazine/article/0,9171,1929212,00.html?iid=tsmodule", []*entityTestResult{}},
						}},
					}},
				}},
				{WikiEntityText, "\n", []*entityTestResult{}},
				{WikiEntityHeading4, "Translations", []*entityTestResult{
					{WikiEntityText, "\n", []*entityTestResult{}},
					{WikiEntityTemplate, "trans-top|research on a wiki", []*entityTestResult{
						{WikiEntityTemplateName, "trans-top", []*entityTestResult{}},
						{WikiEntityTemplateProp, "research on a wiki", []*entityTestResult{}},
					}},
					{WikiEntityListBulleted, " Dutch: {{t-|nl|wikiën}}", []*entityTestResult{
						{WikiEntityTemplate, "t-|nl|wikiën", []*entityTestResult{
							{WikiEntityTemplateName, "t-", []*entityTestResult{}},
							{WikiEntityTemplateProp, "nl", []*entityTestResult{}},
							{WikiEntityTemplateProp, "wikiën", []*entityTestResult{}},
						}},
					}},
					{WikiEntityText, "\n", []*entityTestResult{}},
					{WikiEntityTemplate, "trans-mid", []*entityTestResult{}},
					{WikiEntityListBulleted, " Limburgish: {{t-|li|wikieë}}", []*entityTestResult{
						{WikiEntityTemplate, "t-|li|wikieë", []*entityTestResult{
							{WikiEntityTemplateName, "t-", []*entityTestResult{}},
							{WikiEntityTemplateProp, "li", []*entityTestResult{}},
							{WikiEntityTemplateProp, "wikieë", []*entityTestResult{}},
						}},
					}},
					{WikiEntityText, "\n", []*entityTestResult{}},
					{WikiEntityTemplate, "trans-bottom", []*entityTestResult{}},
					{WikiEntityText, "\n", []*entityTestResult{}},
					{WikiEntityText, "\n", []*entityTestResult{}},
					{WikiEntityTemplate, "trans-top|contribute to a wiki", []*entityTestResult{}},
					{WikiEntityListBulleted, " Dutch: {{t-|nl|wikiën}}", []*entityTestResult{
						{WikiEntityTemplate, "t-|nl|wikiën", []*entityTestResult{
							{WikiEntityTemplateName, "t-", []*entityTestResult{}},
							{WikiEntityTemplateProp, "nl", []*entityTestResult{}},
							{WikiEntityTemplateProp, "wikiën", []*entityTestResult{}},
						}},
					}},
					{WikiEntityListBulleted, " Esperanto: {{t-|eo|vikiumi}}", []*entityTestResult{
						{WikiEntityTemplate, "t-|eo|vikiumi", []*entityTestResult{
							{WikiEntityTemplateName, "t-", []*entityTestResult{}},
							{WikiEntityTemplateProp, "eo", []*entityTestResult{}},
							{WikiEntityTemplateProp, "vikiumi", []*entityTestResult{}},
						}},
					}},
					{WikiEntityText, "\n", []*entityTestResult{}},
					{WikiEntityTemplate, "trans-mid", []*entityTestResult{}},
					{WikiEntityListBulleted, " Limburgish: {{t-|li|bewikieë}}", []*entityTestResult{
						{WikiEntityTemplate, "t-|li|bewikieë", []*entityTestResult{
							{WikiEntityTemplateName, "t-", []*entityTestResult{}},
							{WikiEntityTemplateProp, "li", []*entityTestResult{}},
							{WikiEntityTemplateProp, "bewikieë", []*entityTestResult{}},
						}},
					}},
					{WikiEntityText, "\n", []*entityTestResult{}},
					{WikiEntityTemplate, "trans-bottom", []*entityTestResult{}},
					{WikiEntityText, "\n", []*entityTestResult{}},
				}},
				{WikiEntityHeading4, "Derived terms", []*entityTestResult{
					{WikiEntityListBulleted, " [[wikify]]", []*entityTestResult{
						{WikiEntityLinkInternal, "wikify", []*entityTestResult{}},
					}},
					{WikiEntityListBulleted, " [[wikiholic]]", []*entityTestResult{
						{WikiEntityLinkInternal, "wikiholic", []*entityTestResult{}},
					}},
					{WikiEntityListBulleted, " [[wikilink]]", []*entityTestResult{
						{WikiEntityLinkInternal, "wikilink", []*entityTestResult{}},
					}},
					{WikiEntityListBulleted, " The names of many wiki-based Web projects, e.g. [[Wikipedia]], [[Wikisource]], [[w:Wiktionary|Wiktionary]] ([[w:Wiktionarian|Wiktionarian]]), {{w|WikiLeaks}}, {{w|Wikibooks}}, {{w|Wikimedia Foundation}}.", []*entityTestResult{
						{WikiEntityLinkInternal, "Wikipedia", []*entityTestResult{}},
						{WikiEntityLinkInternal, "Wikisource", []*entityTestResult{}},
						{WikiEntityLinkInternal, "w:Wiktionary|Wiktionary", []*entityTestResult{
							{WikiEntityLinkInternalName, "w:Wiktionary", []*entityTestResult{}},
							{WikiEntityLinkInternalProp, "Wiktionary", []*entityTestResult{}},
						}},
						{WikiEntityLinkInternal, "w:Wiktionarian|Wiktionarian", []*entityTestResult{
							{WikiEntityLinkInternalName, "w:Wiktionarian", []*entityTestResult{}},
							{WikiEntityLinkInternalProp, "Wiktionarian", []*entityTestResult{}},
						}},
						{WikiEntityTemplate, "w|WikiLeaks", []*entityTestResult{
							{WikiEntityTemplateName, "w", []*entityTestResult{}},
							{WikiEntityTemplateProp, "WikiLeaks", []*entityTestResult{}},
						}},
						{WikiEntityTemplate, "w|Wikibooks", []*entityTestResult{
							{WikiEntityTemplateName, "w", []*entityTestResult{}},
							{WikiEntityTemplateProp, "Wikibooks", []*entityTestResult{}},
						}},
						{WikiEntityTemplate, "w|Wikimedia Foundation", []*entityTestResult{
							{WikiEntityTemplateName, "w", []*entityTestResult{}},
							{WikiEntityTemplateProp, "Wikimedia Foundation", []*entityTestResult{}},
						}},
					}},
					{WikiEntityText, "\n", []*entityTestResult{}},
				}},
			}},
			{WikiEntityHeading3, "References", []*entityTestResult{
				{WikiEntityListBulleted, " {{R:American Heritage 2000|wiki}}", []*entityTestResult{}},
				{WikiEntityListBulleted, " {{R:Webster’s New Millennium|wiki}}", []*entityTestResult{}},
				{WikiEntityListBulleted, " Notes:", []*entityTestResult{}},
				{WikiEntityText, "\n", []*entityTestResult{}},
				{WikiEntityTag, "references", []*entityTestResult{}},
				{WikiEntityText, "\n", []*entityTestResult{}},
			}},
			{WikiEntityHeading3, "Anagrams", []*entityTestResult{
				{WikiEntityListBulleted, " [[kiwi#English|kiwi]], [[Kiwi#English|Kiwi]]", []*entityTestResult{
					{WikiEntityLinkInternal, "kiwi#English|kiwi", []*entityTestResult{
						{WikiEntityLinkInternalName, "kiwi#English", []*entityTestResult{}},
						{WikiEntityLinkInternalProp, "kiwi", []*entityTestResult{}},
					}},
					{WikiEntityLinkInternal, "Kiwi#English|Kiwi", []*entityTestResult{
						{WikiEntityLinkInternalName, "Kiwi#English", []*entityTestResult{}},
						{WikiEntityLinkInternalProp, "Kiwi", []*entityTestResult{}},
					}},
				}},
				{WikiEntityText, "\n", []*entityTestResult{}},
				{WikiEntityText, "\n", []*entityTestResult{}},
				{WikiEntityLinkInternal, "Category:en:Websites", []*entityTestResult{}},
				{WikiEntityText, "\n", []*entityTestResult{}},
				{WikiEntityHR, "----", []*entityTestResult{}},
				{WikiEntityText, "\n", []*entityTestResult{}},
			}},
		}},
		{WikiEntityHeading2, "Dutch", []*entityTestResult{
			{WikiEntityText, "\n", []*entityTestResult{}},
			{WikiEntityHeading3, "Etymology", []*entityTestResult{
			}},
			{WikiEntityHeading3, "Pronunciation", []*entityTestResult{
			}},
			{WikiEntityHeading3, "Noun", []*entityTestResult{
				{WikiEntityText, "\n", []*entityTestResult{}},
				{WikiEntityTemplate, "nl-noun|m|wiki's|wikietje", []*entityTestResult{
					{WikiEntityTemplateName, "nl-noun", []*entityTestResult{}},
					{WikiEntityTemplateProp, "m", []*entityTestResult{}},
					{WikiEntityTemplateProp, "wiki's", []*entityTestResult{}},
					{WikiEntityTemplateProp, "wikietje", []*entityTestResult{}},
				}},
				{WikiEntityText, "\n", []*entityTestResult{}},
				{WikiEntityListNumbered, " {{l|en|wiki}}", []*entityTestResult{
					{WikiEntityTemplate, "l|en|wiki", []*entityTestResult{
						{WikiEntityTemplateName, "l", []*entityTestResult{}},
						{WikiEntityTemplateProp, "en", []*entityTestResult{}},
						{WikiEntityTemplateProp, "wiki", []*entityTestResult{}},
					}},
				}},
				{WikiEntityText, "\n", []*entityTestResult{}},
				{WikiEntityHeading4, "Derived terms", []*entityTestResult{
					{WikiEntityListBulleted, " [[wikiën]]", []*entityTestResult{
						{WikiEntityLinkInternal, "wikiën", []*entityTestResult{}},
					}},
					{WikiEntityText, "\n", []*entityTestResult{}},
				}},
			}},
			{WikiEntityHeading3, "Anagrams", []*entityTestResult{
				{WikiEntityListBulleted, " [[kiwi]]", []*entityTestResult{
					{WikiEntityLinkInternal, "kiwi", []*entityTestResult{}},
				}},
				{WikiEntityText, "\n", []*entityTestResult{}},
				{WikiEntityHR, "----", []*entityTestResult{}},
				{WikiEntityText, "\n", []*entityTestResult{}},
			}},
		}},
		{WikiEntityHeading2, "French", []*entityTestResult{
			/*
			{WikiEntityHeading3, "Etymology", []*entityTestResult{
			}},
			{WikiEntityHeading3, "Pronunciation", []*entityTestResult{
			}},
			{WikiEntityHeading3, "Noun", []*entityTestResult{
			}},
			{WikiEntityHeading3, "References", []*entityTestResult{
			}},
			{WikiEntityHeading3, "Anagrams", []*entityTestResult{
			}}, */
		}},
		{WikiEntityHeading2, "Hawaiian", []*entityTestResult{
			/*
			{WikiEntityHeading3, "Etymology", []*entityTestResult{
			}},
			{WikiEntityHeading3, "Pronunciation", []*entityTestResult{
			}},
			{WikiEntityHeading3, "Noun", []*entityTestResult{
			}},
			{WikiEntityHeading3, "References", []*entityTestResult{
			}},
			{WikiEntityHeading3, "Anagrams", []*entityTestResult{
			}}, */
		}},
		{WikiEntityHeading2, "Limburgish", []*entityTestResult{
			/*
			{WikiEntityHeading3, "Etymology", []*entityTestResult{
			}},
			{WikiEntityHeading3, "Pronunciation", []*entityTestResult{
			}},
			{WikiEntityHeading3, "Noun", []*entityTestResult{
			}},
			{WikiEntityHeading3, "References", []*entityTestResult{
			}},
			{WikiEntityHeading3, "Anagrams", []*entityTestResult{
			}}, */
		}},
		{WikiEntityHeading2, "Lower Sorbian", []*entityTestResult{
			/*
			{WikiEntityHeading3, "Etymology", []*entityTestResult{
			}},
			{WikiEntityHeading3, "Pronunciation", []*entityTestResult{
			}},
			{WikiEntityHeading3, "Noun", []*entityTestResult{
			}},
			{WikiEntityHeading3, "References", []*entityTestResult{
			}},
			{WikiEntityHeading3, "Anagrams", []*entityTestResult{
			}}, */
		}},
		{WikiEntityHeading2, "Norwegian", []*entityTestResult{
			/*
			{WikiEntityHeading3, "Etymology", []*entityTestResult{
			}},
			{WikiEntityHeading3, "Pronunciation", []*entityTestResult{
			}},
			{WikiEntityHeading3, "Noun", []*entityTestResult{
			}},
			{WikiEntityHeading3, "References", []*entityTestResult{
			}},
			{WikiEntityHeading3, "Anagrams", []*entityTestResult{
			}}, */
		}},
		{WikiEntityHeading2, "Spanish", []*entityTestResult{
			/*
			{WikiEntityHeading3, "Etymology", []*entityTestResult{
			}},
			{WikiEntityHeading3, "Pronunciation", []*entityTestResult{
			}},
			{WikiEntityHeading3, "Noun", []*entityTestResult{
			}},
			{WikiEntityHeading3, "References", []*entityTestResult{
			}},
			{WikiEntityHeading3, "Anagrams", []*entityTestResult{
			}}, */
		}},
		{WikiEntityHeading2, "Swahili", []*entityTestResult{
			/*
			{WikiEntityHeading3, "Etymology", []*entityTestResult{
			}},
			{WikiEntityHeading3, "Pronunciation", []*entityTestResult{
			}},
			{WikiEntityHeading3, "Noun", []*entityTestResult{
			}},
			{WikiEntityHeading3, "References", []*entityTestResult{
			}},
			{WikiEntityHeading3, "Anagrams", []*entityTestResult{
			}}, */
		}},
		{WikiEntityHeading2, "Swedish", []*entityTestResult{
			/*
			{WikiEntityHeading3, "Etymology", []*entityTestResult{
			}},
			{WikiEntityHeading3, "Pronunciation", []*entityTestResult{
			}},
			{WikiEntityHeading3, "Noun", []*entityTestResult{
			}},
			{WikiEntityHeading3, "References", []*entityTestResult{
			}},
			{WikiEntityHeading3, "Anagrams", []*entityTestResult{
			}}, */
		}},
		{WikiEntityHeading2, "Tocharian A", []*entityTestResult{
			{WikiEntityText, "\n", []*entityTestResult{}},
			{WikiEntityHeading3, "Etymology", []*entityTestResult{
				{WikiEntityText, "\nFrom a hypothetical ", []*entityTestResult{}},
				{WikiEntityTemplate, "etyl|ine-toc-pro|xto", []*entityTestResult{}},
				{WikiEntityText, " ", []*entityTestResult{}},
				{WikiEntityTemplate, "term/t|ine-pro|*w'īkän", []*entityTestResult{}},
				{WikiEntityText, ", from ", []*entityTestResult{}},
				{WikiEntityTemplate, "etyl|ine-pro|xto", []*entityTestResult{}},
				{WikiEntityText, " ", []*entityTestResult{}},
				{WikiEntityTemplate, "term/t|ine-pro|*h₁wih₁ḱm̥t", []*entityTestResult{}},
				{WikiEntityText, " or ", []*entityTestResult{}},
				{WikiEntityTemplate, "term/t|ine-pro|*h₁wih₁ḱm̥ti", []*entityTestResult{}},
				{WikiEntityText, ", ", []*entityTestResult{}},
				{WikiEntityTemplate, "term/t|ine-pro|*dwi(h₁)dḱm̥ti", []*entityTestResult{}},
				{WikiEntityText, " (cognate with Latin ", []*entityTestResult{}},
				{WikiEntityTemplate, "term|vīgintī|lang=la", []*entityTestResult{}},
				{WikiEntityText, ", Ancient Greek ", []*entityTestResult{}},
				{WikiEntityTemplate, "term|εἴκοσι||tr=eikosi|lang=grc", []*entityTestResult{}},
				{WikiEntityText, ", Doric ", []*entityTestResult{}},
				{WikiEntityTemplate, "term|ϝείκατι||tr=weikati|lang=grc", []*entityTestResult{}},
				{WikiEntityText, ", Sanskrit ", []*entityTestResult{}},
				{WikiEntityTemplate, "term|विंशति||tr=viṃśati|lang=sa", []*entityTestResult{}},
				{WikiEntityText, ", Avestan ", []*entityTestResult{}},
				{WikiEntityTextItalic, "vīsaiti", []*entityTestResult{}},
				{WikiEntityText, ", Ossetian ", []*entityTestResult{}},
				{WikiEntityTextItalic, "insäi", []*entityTestResult{}},
				{WikiEntityText, ", Armenian ", []*entityTestResult{}},
				{WikiEntityTemplate, "term|քսան||tr=k'san|lang=hy", []*entityTestResult{}},
				{WikiEntityText, ", Albanian ", []*entityTestResult{}},
				{WikiEntityTemplate, "term|njëzet|(një)zet|lang=sq", []*entityTestResult{}},
				{WikiEntityText, ", Sanskrit ", []*entityTestResult{}},
				{WikiEntityTemplate, "term|विंशति||tr=viṃśati|lang=sa", []*entityTestResult{}},
				{WikiEntityText, ", Welsh ", []*entityTestResult{}},
				{WikiEntityTemplate, "term|ugain|lang=cy", []*entityTestResult{}},
				{WikiEntityText, "). Compare Tocharian B ", []*entityTestResult{}},
				{WikiEntityTemplate, "term|ikäṃ|lang=txb", []*entityTestResult{}},
				{WikiEntityText, ".\n", []*entityTestResult{}},
			}},
			{WikiEntityHeading3, "Numeral", []*entityTestResult{
				{WikiEntityText, "\n", []*entityTestResult{}},
				{WikiEntityTemplate, "head|xto|numeral", []*entityTestResult{}},
				{WikiEntityText, "\n", []*entityTestResult{}},
				{WikiEntityListNumbered, " {{context|cardinal|lang=xto}} [[twenty]]", []*entityTestResult{}},
				{WikiEntityText, "\n", []*entityTestResult{}},
				{WikiEntityText, "\n", []*entityTestResult{}},
				{WikiEntityLinkInternal, "ca:wiki", []*entityTestResult{}}, {WikiEntityText, "\n", []*entityTestResult{}},
				{WikiEntityLinkInternal, "cs:wiki", []*entityTestResult{}}, {WikiEntityText, "\n", []*entityTestResult{}},
				{WikiEntityLinkInternal, "cy:wiki", []*entityTestResult{}}, {WikiEntityText, "\n", []*entityTestResult{}},
				{WikiEntityLinkInternal, "da:wiki", []*entityTestResult{}}, {WikiEntityText, "\n", []*entityTestResult{}},
				{WikiEntityLinkInternal, "de:wiki", []*entityTestResult{}}, {WikiEntityText, "\n", []*entityTestResult{}},
				{WikiEntityLinkInternal, "et:wiki", []*entityTestResult{}}, {WikiEntityText, "\n", []*entityTestResult{}},
				{WikiEntityLinkInternal, "el:wiki", []*entityTestResult{}}, {WikiEntityText, "\n", []*entityTestResult{}},
				{WikiEntityLinkInternal, "es:wiki", []*entityTestResult{}}, {WikiEntityText, "\n", []*entityTestResult{}},
				{WikiEntityLinkInternal, "fr:wiki", []*entityTestResult{}}, {WikiEntityText, "\n", []*entityTestResult{}},
				{WikiEntityLinkInternal, "gl:wiki", []*entityTestResult{}}, {WikiEntityText, "\n", []*entityTestResult{}},
				{WikiEntityLinkInternal, "ko:wiki", []*entityTestResult{}}, {WikiEntityText, "\n", []*entityTestResult{}},
				{WikiEntityLinkInternal, "hi:wiki", []*entityTestResult{}}, {WikiEntityText, "\n", []*entityTestResult{}},
				{WikiEntityLinkInternal, "id:wiki", []*entityTestResult{}}, {WikiEntityText, "\n", []*entityTestResult{}},
				{WikiEntityLinkInternal, "is:wiki", []*entityTestResult{}}, {WikiEntityText, "\n", []*entityTestResult{}},
				{WikiEntityLinkInternal, "it:wiki", []*entityTestResult{}}, {WikiEntityText, "\n", []*entityTestResult{}},
				{WikiEntityLinkInternal, "he:wiki", []*entityTestResult{}}, {WikiEntityText, "\n", []*entityTestResult{}},
				{WikiEntityLinkInternal, "jv:wiki", []*entityTestResult{}}, {WikiEntityText, "\n", []*entityTestResult{}},
				{WikiEntityLinkInternal, "kn:wiki", []*entityTestResult{}}, {WikiEntityText, "\n", []*entityTestResult{}},
				{WikiEntityLinkInternal, "sw:wiki", []*entityTestResult{}}, {WikiEntityText, "\n", []*entityTestResult{}},
				{WikiEntityLinkInternal, "lv:wiki", []*entityTestResult{}}, {WikiEntityText, "\n", []*entityTestResult{}},
				{WikiEntityLinkInternal, "lt:wiki", []*entityTestResult{}}, {WikiEntityText, "\n", []*entityTestResult{}},
				{WikiEntityLinkInternal, "li:wiki", []*entityTestResult{}}, {WikiEntityText, "\n", []*entityTestResult{}},
				{WikiEntityLinkInternal, "hu:wiki", []*entityTestResult{}}, {WikiEntityText, "\n", []*entityTestResult{}},
				{WikiEntityLinkInternal, "mg:wiki", []*entityTestResult{}}, {WikiEntityText, "\n", []*entityTestResult{}},
				{WikiEntityLinkInternal, "fj:wiki", []*entityTestResult{}}, {WikiEntityText, "\n", []*entityTestResult{}},
				{WikiEntityLinkInternal, "nl:wiki", []*entityTestResult{}}, {WikiEntityText, "\n", []*entityTestResult{}},
				{WikiEntityLinkInternal, "ja:wiki", []*entityTestResult{}}, {WikiEntityText, "\n", []*entityTestResult{}},
				{WikiEntityLinkInternal, "no:wiki", []*entityTestResult{}}, {WikiEntityText, "\n", []*entityTestResult{}},
				{WikiEntityLinkInternal, "nn:wiki", []*entityTestResult{}}, {WikiEntityText, "\n", []*entityTestResult{}},
				{WikiEntityLinkInternal, "km:wiki", []*entityTestResult{}}, {WikiEntityText, "\n", []*entityTestResult{}},
				{WikiEntityLinkInternal, "pl:wiki", []*entityTestResult{}}, {WikiEntityText, "\n", []*entityTestResult{}},
				{WikiEntityLinkInternal, "pt:wiki", []*entityTestResult{}}, {WikiEntityText, "\n", []*entityTestResult{}},
				{WikiEntityLinkInternal, "ro:wiki", []*entityTestResult{}}, {WikiEntityText, "\n", []*entityTestResult{}},
				{WikiEntityLinkInternal, "ru:wiki", []*entityTestResult{}}, {WikiEntityText, "\n", []*entityTestResult{}},
				{WikiEntityLinkInternal, "scn:wiki", []*entityTestResult{}}, {WikiEntityText, "\n", []*entityTestResult{}},
				{WikiEntityLinkInternal, "simple:wiki", []*entityTestResult{}}, {WikiEntityText, "\n", []*entityTestResult{}},
				{WikiEntityLinkInternal, "sk:wiki", []*entityTestResult{}}, {WikiEntityText, "\n", []*entityTestResult{}},
				{WikiEntityLinkInternal, "fi:wiki", []*entityTestResult{}}, {WikiEntityText, "\n", []*entityTestResult{}},
				{WikiEntityLinkInternal, "sv:wiki", []*entityTestResult{}}, {WikiEntityText, "\n", []*entityTestResult{}},
				{WikiEntityLinkInternal, "tl:wiki", []*entityTestResult{}}, {WikiEntityText, "\n", []*entityTestResult{}},
				{WikiEntityLinkInternal, "te:wiki", []*entityTestResult{}}, {WikiEntityText, "\n", []*entityTestResult{}},
				{WikiEntityLinkInternal, "tr:wiki", []*entityTestResult{}}, {WikiEntityText, "\n", []*entityTestResult{}},
				{WikiEntityLinkInternal, "vi:wiki", []*entityTestResult{}}, {WikiEntityText, "\n", []*entityTestResult{}},
				{WikiEntityLinkInternal, "zh:wiki", []*entityTestResult{}}, {WikiEntityText, "\n", []*entityTestResult{}},
			}},
		}},
	}
	checkEntityResults(t, i, "checkTestDataWiki", data, string(data), wiki, results, true)
}

func checkTestDataTest(t *testing.T, i int, data []byte, wiki *Entity) {
	results := []*entityTestResult{
		{WikiEntityTemplate, "also|Test|țest", []*entityTestResult{}},
		{WikiEntityHeading2, "English", []*entityTestResult{
		}},
		{WikiEntityHeading2, "Breton", []*entityTestResult{
		}},
		{WikiEntityHeading2, "Czech", []*entityTestResult{
			{WikiEntityText, "\n", []*entityTestResult{}},
			{WikiEntityHeading3, "Noun", []*entityTestResult{
				{WikiEntityText, "\n", []*entityTestResult{}},
				{WikiEntityTemplate, "cs-noun|g=m", []*entityTestResult{
					{WikiEntityTemplateName, "cs-noun", []*entityTestResult{}},
					{WikiEntityTemplateProp, "g=m", []*entityTestResult{}},
				}},
				{WikiEntityText, "\n", []*entityTestResult{}},
				{WikiEntityListNumbered, " {{l|en|test}}", []*entityTestResult{}},
				{WikiEntityText, "\n", []*entityTestResult{}},
				{WikiEntityHeading4, "Derived terms", []*entityTestResult{
					{WikiEntityListBulleted, " {{l|cs|testovat}}", []*entityTestResult{}},
					{WikiEntityListBulleted, " {{l|cs|testovací}}", []*entityTestResult{}},
					{WikiEntityListBulleted, " {{l|cs|testový}}", []*entityTestResult{}},
					{WikiEntityText, "\n", []*entityTestResult{}},
					{WikiEntityHR, "----", []*entityTestResult{}},
					{WikiEntityText, "\n", []*entityTestResult{}},
				}},
			}},
		}},
		{WikiEntityHeading2, "French", []*entityTestResult{
			{WikiEntityText, "\n", []*entityTestResult{}},
			{WikiEntityHeading3, "Etymology", []*entityTestResult{
			}},
			{WikiEntityHeading3, "Pronunciation", []*entityTestResult{
			}},
			{WikiEntityHeading3, "Noun", []*entityTestResult{
			}},
		}},
		{WikiEntityHeading2, "Hungarian", []*entityTestResult{
			{WikiEntityText, "\n", []*entityTestResult{}},
			{WikiEntityHeading3, "Etymology", []*entityTestResult{
			}},
			{WikiEntityHeading3, "Pronunciation", []*entityTestResult{
			}},
			{WikiEntityHeading3, "Noun", []*entityTestResult{
			}},
		}},
		{WikiEntityHeading2, "Italian", []*entityTestResult{
			{WikiEntityText, "\n", []*entityTestResult{}},
			{WikiEntityHeading3, "Etymology", []*entityTestResult{
			}},
			{WikiEntityHeading3, "Noun", []*entityTestResult{
			}},
		}},
		{WikiEntityHeading2, "Ladin", []*entityTestResult{
			{WikiEntityText, "\n", []*entityTestResult{}},
			{WikiEntityHeading3, "Noun", []*entityTestResult{
			}},
		}},
		{WikiEntityHeading2, "Old French", []*entityTestResult{
			{WikiEntityText, "\n", []*entityTestResult{}},
			{WikiEntityHeading3, "Etymology", []*entityTestResult{
			}},
			{WikiEntityHeading3, "Noun", []*entityTestResult{
			}},
			{WikiEntityHeading3, "References", []*entityTestResult{
			}},
		}},
		{WikiEntityHeading2, "Spanish", []*entityTestResult{
			{WikiEntityText, "\n", []*entityTestResult{}},
			{WikiEntityHeading3, "Noun", []*entityTestResult{
			}},
		}},
		{WikiEntityHeading2, "Swedish", []*entityTestResult{
			{WikiEntityText, "\n", []*entityTestResult{}},
			{WikiEntityHeading3, "Noun", []*entityTestResult{
			}},
		}},
	}
	checkEntityResults(t, i, "checkTestDataTest", data, string(data), wiki, results, true)
}

func checkTestDataBook(t *testing.T, i int, data []byte, wiki *Entity) {
	results := []*entityTestResult{
		{WikiEntityHeading2, "English", []*entityTestResult{
		}},
		{WikiEntityHeading2, "Limburgish", []*entityTestResult{
		}},
		{WikiEntityHeading2, "Middle English", []*entityTestResult{
		}},
	}
	checkEntityResults(t, i, "checkTestDataBook", data, string(data), wiki, results, true)
}

func checkTestDataAny(t *testing.T, i int, data []byte, wiki *Entity) {
	results := []*entityTestResult{
		{WikiEntityTemplate, "also|-any|-ány|ăn ý", []*entityTestResult{}},
		{WikiEntityHeading2, "English", []*entityTestResult{
			{WikiEntityText, "\n", []*entityTestResult{}},
			{WikiEntityTemplate, "wikipedia", []*entityTestResult{}},
			{WikiEntityText, "\n", []*entityTestResult{}},
			{WikiEntityHeading3, "Alternative forms", []*entityTestResult{
			}},
			{WikiEntityHeading3, "Etymology", []*entityTestResult{
			}},
			{WikiEntityHeading3, "Pronunciation", []*entityTestResult{
			}},
			{WikiEntityHeading3, "Adverb", []*entityTestResult{
			}},
			{WikiEntityHeading3, "Determiner", []*entityTestResult{
			}},
			{WikiEntityHeading3, "Pronoun", []*entityTestResult{
				{WikiEntityText, "\n", []*entityTestResult{}},
				{WikiEntityTextBold, "any", []*entityTestResult{}},
				{WikiEntityText, "\n", []*entityTestResult{}},
				{WikiEntityListNumbered, " Any thing(s) or person(s).", []*entityTestResult{}},
				{WikiEntityListNumbered, ": '''''Any''' may apply.''", []*entityTestResult{
					{WikiEntityIndent, " '''''Any''' may apply.''", []*entityTestResult{
						{WikiEntityTextItalic, "'''Any''' may apply.", []*entityTestResult{
							{WikiEntityTextBold, "Any", []*entityTestResult{}},
						}},
					}},
				}},
				{WikiEntityText, "\n", []*entityTestResult{}},
				{WikiEntityHeading4, "Translations", []*entityTestResult{
				}},
			}},
			{WikiEntityHeading3, "Statistics", []*entityTestResult{
			}},
			{WikiEntityHeading3, "Anagrams", []*entityTestResult{
			}},
		}},
		{WikiEntityHeading2, "Catalan", []*entityTestResult{
			{WikiEntityText, "\n", []*entityTestResult{}},
			{WikiEntityHeading3, "Etymology", []*entityTestResult{
			}},
			{WikiEntityHeading3, "Pronunciation", []*entityTestResult{
			}},
			{WikiEntityHeading3, "Noun", []*entityTestResult{
			}},
		}},
	}
	checkEntityResults(t, i, "checkTestDataAny", data, string(data), wiki, results, true)
}

func checkTestDataA(t *testing.T, i int, data []byte, wiki *Entity) {
	results := []*entityTestResult{
		{WikiEntityTemplate, `also|A|Appendix:Variations of "a"|𐌳`, []*entityTestResult{}},
		{WikiEntityHeading2, "Translingual", []*entityTestResult{
			{WikiEntityText, "\n", []*entityTestResult{}},
			{WikiEntityTemplate, "Basic Latin character info|previous=`|next=b|image=[[Image:Letter a.svg|50px]]|hex=61|name=LATIN SMALL LETTER A", []*entityTestResult{}},
			{WikiEntityText, "\n", []*entityTestResult{}},
			{WikiEntityTemplate, `wikisource1911Enc|A`, []*entityTestResult{}},
			{WikiEntityText, "\n", []*entityTestResult{}},
			{WikiEntityHeading3, "Etymology 1", []*entityTestResult{
				{WikiEntityText, "\n", []*entityTestResult{}},
				{WikiEntityLinkInternal, "Image:UncialA-01.svg|50px|Approximate form of Greek upper case Α (a, “alpha”) that was the source for both common variants of ''a''", []*entityTestResult{}},
				{WikiEntityText, "\nModification of capital letter ", []*entityTestResult{}},
				{WikiEntityTemplate, "term|A|lang=mul", []*entityTestResult{}},
				{WikiEntityText, ", from ", []*entityTestResult{}},
				{WikiEntityTemplate, "etyl|la|mul", []*entityTestResult{}},
				{WikiEntityText, " ", []*entityTestResult{}},
				{WikiEntityTemplate, "term|A|lang=la", []*entityTestResult{}},
				{WikiEntityText, ", from ", []*entityTestResult{}},
				{WikiEntityTemplate, "etyl|grc|mul", []*entityTestResult{}},
				{WikiEntityText, " letter ", []*entityTestResult{}},
				{WikiEntityTemplate, "term|Α|tr=A|lang=grc", []*entityTestResult{}},
				{WikiEntityText, ".\n", []*entityTestResult{}},
				{WikiEntityHeading4, "Pronunciation", []*entityTestResult{
				}},
				{WikiEntityHeading4, "Letter", []*entityTestResult{
				}},
				{WikiEntityHeading4, "Symbol", []*entityTestResult{
				}},
				{WikiEntityHeading4, "See also", []*entityTestResult{
				}},
				{WikiEntityHeading4, "External links", []*entityTestResult{
				}},
			}},
			{WikiEntityHeading3, "Etymology 2", []*entityTestResult{
				/*
				{WikiEntityHeading4, "Symbol", []*entityTestResult{
				}}, */
			}},
			{WikiEntityHeading3, "Etymology 3", []*entityTestResult{
				/*
				{WikiEntityHeading4, "Symbol", []*entityTestResult{
				}}, */
			}},
			{WikiEntityHeading3, "Etymology 4", []*entityTestResult{
				/*
				{WikiEntityHeading4, "Symbol", []*entityTestResult{
				}}, */
			}},
		}},
		{WikiEntityHeading2, "English", []*entityTestResult{
			{WikiEntityText, "\n", []*entityTestResult{}},
			{WikiEntityHeading3, "Etymology 1", []*entityTestResult{
				{WikiEntityText, "\n", []*entityTestResult{}},
				{WikiEntityLinkInternal, "Image:Runic letter ansuz.png|left|30px|Runic letter {{term|lang=mul||ᚫ|tr=a|ansuz}}, source for Anglo-Saxon Futhorc letters replaced by ''a''", []*entityTestResult{}},
				{WikiEntityText, "\nFrom ", []*entityTestResult{}},
				{WikiEntityTemplate, "etyl|enm|en", []*entityTestResult{}},
				{WikiEntityText, " and ", []*entityTestResult{}},
				{WikiEntityTemplate, "etyl|ang|en", []*entityTestResult{}},
				{WikiEntityText, " lower case letter ", []*entityTestResult{}},
				{WikiEntityTemplate, "term|a|lang=enm", []*entityTestResult{}},
				{WikiEntityText, " and split of ", []*entityTestResult{}},
				{WikiEntityTemplate, "etyl|enm|en", []*entityTestResult{}},
				{WikiEntityText, " and ", []*entityTestResult{}},
				{WikiEntityTemplate, "etyl|ang|en", []*entityTestResult{}},
				{WikiEntityText, " lower case letter ", []*entityTestResult{}},
				{WikiEntityTemplate, "term|æ|lang=enm", []*entityTestResult{}},
				{WikiEntityText, ".", []*entityTestResult{}},
				{WikiEntityIndent, "* [[Image:Rune-Ac.png|10px|Anglo-Saxon Futhorc letter {{term|lang=mul||ᚪ|tr=a|āc}}]] {{etyl|ang|en}} lower case letter {{term|a|lang=enm}} from 7th century replacement by Latin lower case letter {{term|a|lang=la}} of the Anglo-Saxon Futhorc letter {{term|lang=mul|sc=Runr|ᚪ||tr=a|āc}}, derived from Runic letter {{term|lang=mul|sc=Runr|ᚫ||tr=a|Ansuz}}.", []*entityTestResult{}},
				{WikiEntityIndent, "* [[Image:Rune-Æsc.png|10px|Anglo-Saxon Futhorc letter {{term|lang=mul||ᚫ|tr=æ|æsc}}]] {{etyl|ang|en}} lower case letter {{term|æ|lang=enm}} from 7th century replacement by Latin lower case ligature {{term|æ|lang=la}} of the Anglo-Saxon Futhorc letter {{term|lang=mul|sc=Runr|ᚫ||tr=æ|æsc}}, also derived from Runic letter {{term|lang=mul|sc=Runr|ᚫ||tr=a|Ansuz}}.", []*entityTestResult{}},
				{WikiEntityText, "\n", []*entityTestResult{}},
				{WikiEntityHeading4, "Alternative forms", []*entityTestResult{
				}},
				{WikiEntityHeading4, "Pronunciation", []*entityTestResult{
				}},
				{WikiEntityHeading4, "Letter", []*entityTestResult{
					/*
					{WikiEntityHeading5, "Usage notes", []*entityTestResult{
					}},
					{WikiEntityHeading5, "Derived terms", []*entityTestResult{
					}},
					{WikiEntityHeading5, "See also", []*entityTestResult{
					}}, /**/
				}},
				{WikiEntityHeading4, "Cardinal number", []*entityTestResult{
				}},
				{WikiEntityHeading4, "Noun", []*entityTestResult{
					{WikiEntityText, "\n", []*entityTestResult{}},
					{WikiEntityTemplate, "en-noun|a's|pl2=as|pl3=aes", []*entityTestResult{
						//{WikiEntityTemplateName, "en-noun", []*entityTestResult{}},
					}},
					/*
					{WikiEntityTagBeg, "ref name = WI3", []*entityTestResult{
						{WikiEntityText, "Gove, Philip Babcock, (1976)", []*entityTestResult{}},
						{WikiEntityTagEnd, "ref", []*entityTestResult{}},
					}}, */
					{WikiEntityTagBeg, "ref name = WI3", []*entityTestResult{}},
					{WikiEntityText, "Gove, Philip Babcock, (1976)", []*entityTestResult{}},
					{WikiEntityTagEnd, "ref", []*entityTestResult{}},
					{WikiEntityText, "\n", []*entityTestResult{}},
					{WikiEntityListNumbered, " {{non-gloss definition|The name of the [[Appendix:Latin script|Latin script]] letter '''[[A]]'''/'''[[a]]'''.}}", []*entityTestResult{}},
					{WikiEntityListNumbered, " {{rfd-sense|fragment=a 2}} A spoken sound represented by the letter ''a'' or ''A'', as in map, mall, or male.", []*entityTestResult{}},
					{WikiEntityListNumbered, " {{rfd-sense|fragment=a 2}} A written representation of the letter ''A'' or ''a''.", []*entityTestResult{}},
					{WikiEntityListNumbered, " {{rfd-sense|fragment=a 2}} A printer's type or stamp used to reproduce the letter ''a''.", []*entityTestResult{}},
					{WikiEntityListNumbered, " {{rfd-sense|fragment=a 2}} An item having the shape of the letter ''a'' or ''A''.", []*entityTestResult{}},
					{WikiEntityText, "\n", []*entityTestResult{}},
					{WikiEntityHeading5, "See also", []*entityTestResult{
					}},
					{WikiEntityHeading5, "Translations", []*entityTestResult{
					}},
				}},
			}},
			{WikiEntityHeading3, "Etymology 2", []*entityTestResult{
				/*
				{WikiEntityHeading4, "Pronunciation", []*entityTestResult{
				}},
				{WikiEntityHeading4, "Article", []*entityTestResult{
					{WikiEntityHeading5, "Usage notes", []*entityTestResult{
					}},
					{WikiEntityHeading5, "Translations", []*entityTestResult{
					}},
				}}, /**/
			}},
			{WikiEntityHeading3, "Etymology 3", []*entityTestResult{
				/*
				{WikiEntityHeading4, "Pronunciation", []*entityTestResult{
				}},
				{WikiEntityHeading4, "Preposition", []*entityTestResult{
					{WikiEntityHeading4, "Usage notes", []*entityTestResult{
					}},
				}}, /**/
			}},
			{WikiEntityHeading3, "Etymology 4", []*entityTestResult{
				/*
				{WikiEntityHeading4, "Alternative forms", []*entityTestResult{
				}},
				{WikiEntityHeading4, "Pronunciation", []*entityTestResult{
				}},
				{WikiEntityHeading4, "Verb", []*entityTestResult{
				}},
				{WikiEntityHeading4, "Derived terms", []*entityTestResult{
				}},
				{WikiEntityHeading4, "Usage notes", []*entityTestResult{
				}}, /**/
			}},
			{WikiEntityHeading3, "Etymology 5", []*entityTestResult{
				/*
				{WikiEntityHeading4, "Alternative forms", []*entityTestResult{
				}},
				{WikiEntityHeading4, "Pronunciation", []*entityTestResult{
				}},
				{WikiEntityHeading4, "Pronoun", []*entityTestResult{
				}}, /**/
			}},
			{WikiEntityHeading3, "Etymology 6", []*entityTestResult{
			}},
			{WikiEntityHeading3, "Etymology 7", []*entityTestResult{
			}},
			{WikiEntityHeading3, "Etymology 8", []*entityTestResult{
			}},
			{WikiEntityHeading3, "Etymology 9", []*entityTestResult{
			}},
			{WikiEntityHeading3, "See also", []*entityTestResult{
			}},
			{WikiEntityHeading3, "Statistics", []*entityTestResult{
			}},
			{WikiEntityHeading3, "Footnotes", []*entityTestResult{
			}},
			{WikiEntityHeading3, "References", []*entityTestResult{
			}},
			{WikiEntityHeading3, "External links", []*entityTestResult{
			}},
		}},
		{WikiEntityHeading2, "Abau", []*entityTestResult{
		}},
		{WikiEntityHeading2, "Afar", []*entityTestResult{
		}},
		{WikiEntityHeading2, "Albanian", []*entityTestResult{
		}},
		{WikiEntityHeading2, "Ama", []*entityTestResult{
		}},
		{WikiEntityHeading2, "Aragonese", []*entityTestResult{
		}},
		{WikiEntityHeading2, "Asturian", []*entityTestResult{
		}},
		{WikiEntityHeading2, "Azeri", []*entityTestResult{
		}},
		{WikiEntityHeading2, "Bavarian", []*entityTestResult{
		}},
		{WikiEntityHeading2, "Catalan", []*entityTestResult{
		}},
		{WikiEntityHeading2, "Czech", []*entityTestResult{
		}},
		{WikiEntityHeading2, "Dalmatian", []*entityTestResult{
		}},
		{WikiEntityHeading2, "Danish", []*entityTestResult{
		}},
		{WikiEntityHeading2, "Dutch", []*entityTestResult{
		}},
		{WikiEntityHeading2, "Esperanto", []*entityTestResult{
		}},
		{WikiEntityHeading2, "Fala", []*entityTestResult{
		}},
		{WikiEntityHeading2, "Faroese", []*entityTestResult{
		}},
		{WikiEntityHeading2, "Finnish", []*entityTestResult{
		}},
		{WikiEntityHeading2, "French", []*entityTestResult{
		}},
		{WikiEntityHeading2, "Galician", []*entityTestResult{
		}},
		{WikiEntityHeading2, "German", []*entityTestResult{
		}},
		{WikiEntityHeading2, "Gilbertese", []*entityTestResult{
		}},
		{WikiEntityHeading2, "Haitian Creole", []*entityTestResult{
		}},
		{WikiEntityHeading2, "Hawaiian", []*entityTestResult{
		}},
		{WikiEntityHeading2, "Hungarian", []*entityTestResult{
		}},
		{WikiEntityHeading2, "Ido", []*entityTestResult{
		}},
		{WikiEntityHeading2, "Indo-Portuguese", []*entityTestResult{
		}},
		{WikiEntityHeading2, "Interlingua", []*entityTestResult{
		}},
		{WikiEntityHeading2, "Irish", []*entityTestResult{
		}},
		{WikiEntityHeading2, "Istriot", []*entityTestResult{
		}},
		{WikiEntityHeading2, "Italian", []*entityTestResult{
		}},
		{WikiEntityHeading2, "Japanese", []*entityTestResult{
		}},
		{WikiEntityHeading2, "Krisa", []*entityTestResult{
		}},
		{WikiEntityHeading2, "Ladin", []*entityTestResult{
		}},
		{WikiEntityHeading2, "Latgalian", []*entityTestResult{
		}},
		{WikiEntityHeading2, "Latin", []*entityTestResult{
		}},
		{WikiEntityHeading2, "Latvian", []*entityTestResult{
		}},
		{WikiEntityHeading2, "Lower Sorbian", []*entityTestResult{
		}},
		{WikiEntityHeading2, "Malay", []*entityTestResult{
		}},
		{WikiEntityHeading2, "Mandarin", []*entityTestResult{
		}},
		{WikiEntityHeading2, "Maori", []*entityTestResult{
		}},
		{WikiEntityHeading2, "Middle French", []*entityTestResult{
		}},
		{WikiEntityHeading2, "Min Nan", []*entityTestResult{
		}},
		{WikiEntityHeading2, "Mopan Maya", []*entityTestResult{
		}},
		{WikiEntityHeading2, "Navajo", []*entityTestResult{
		}},
		{WikiEntityHeading2, "Neapolitan", []*entityTestResult{
		}},
		{WikiEntityHeading2, "Norwegian Bokmål", []*entityTestResult{
		}},
		{WikiEntityHeading2, "Norwegian Nynorsk", []*entityTestResult{
		}},
		{WikiEntityHeading2, "Novial", []*entityTestResult{
		}},
		{WikiEntityHeading2, "Old English", []*entityTestResult{
		}},
		{WikiEntityHeading2, "Old French", []*entityTestResult{
		}},
		{WikiEntityHeading2, "Old Irish", []*entityTestResult{
		}},
		{WikiEntityHeading2, "Polish", []*entityTestResult{
		}},
		{WikiEntityHeading2, "Portuguese", []*entityTestResult{
		}},
		{WikiEntityHeading2, "Rapa Nui", []*entityTestResult{
		}},
		{WikiEntityHeading2, "Romanian", []*entityTestResult{
		}},
		{WikiEntityHeading2, "Scots", []*entityTestResult{
		}},
		{WikiEntityHeading2, "Scottish Gaelic", []*entityTestResult{
		}},
		{WikiEntityHeading2, "Serbo-Croatian", []*entityTestResult{
		}},
		{WikiEntityHeading2, "Slovak", []*entityTestResult{
		}},
		{WikiEntityHeading2, "Slovene", []*entityTestResult{
		}},
		{WikiEntityHeading2, "Spanish", []*entityTestResult{
		}},
		{WikiEntityHeading2, "Sranan Tongo", []*entityTestResult{
		}},
		{WikiEntityHeading2, "Swedish", []*entityTestResult{
		}},
		{WikiEntityHeading2, "Tagalog", []*entityTestResult{
		}},
		{WikiEntityHeading2, "Tarantino", []*entityTestResult{
		}},
		{WikiEntityHeading2, "Tok Pisin", []*entityTestResult{
		}},
		{WikiEntityHeading2, "Turkish", []*entityTestResult{
		}},
		{WikiEntityHeading2, "Walloon", []*entityTestResult{
		}},
		{WikiEntityHeading2, "Welsh", []*entityTestResult{
		}},
		{WikiEntityHeading2, "Yoruba", []*entityTestResult{
		}},
		{WikiEntityHeading2, "Zhuang", []*entityTestResult{
		}},
	}
	checkEntityResults(t, i, "checkTestDataA", data, string(data), wiki, results, true)
}

func checkTestDataRain(t *testing.T, i int, data []byte, wiki *Entity) {
	results := []*entityTestResult{
		{WikiEntityTemplate, `also|Rain|räin`, []*entityTestResult{}},
		{WikiEntityHeading2, "English", []*entityTestResult{
			// TODO: ...
		}},
		{WikiEntityHeading2, "Japanese", []*entityTestResult{
			{WikiEntityText, "\n", []*entityTestResult{}},
			{WikiEntityHeading3, "Romanization", []*entityTestResult{
				// TODO: ...
			}},
		}},
	}
	checkEntityResults(t, i, "checkTestDataRain", data, string(data), wiki, results, true)
}

func checkTestDataBetter(t *testing.T, i int, data []byte, wiki *Entity) {
	results := []*entityTestResult{
		{WikiEntityHeading2, "English", []*entityTestResult{
			// TODO: ...
		}},
		{WikiEntityHeading2, "Scots", []*entityTestResult{
			// TODO: ...
		}},
		{WikiEntityHeading2, "West Frisian", []*entityTestResult{
			{WikiEntityText, "\n", []*entityTestResult{}},
			{WikiEntityHeading3, "Etymology", []*entityTestResult{
				// TODO: ...
			}},
			{WikiEntityHeading3, "Adjective", []*entityTestResult{
				// TODO: ...
			}},
		}},
	}
	checkEntityResults(t, i, "checkTestDataBetter", data, string(data), wiki, results, true)
}

func TestParseData(t *testing.T) {
	files := []struct{
		file string
		check func(t *testing.T, i int, data []byte, wiki *Entity)
	}{
		{"testdata/square.wiki.gz", checkTestDataSquare},
		{"testdata/square2.wiki.gz", checkTestDataSquare},
		{"testdata/wiki.wiki.gz", checkTestDataWiki},
		{"testdata/test.wiki.gz", checkTestDataTest},
		{"testdata/book.wiki.gz", checkTestDataBook},
		{"testdata/any.wiki.gz", checkTestDataAny},
		{"testdata/a.wiki.gz", checkTestDataA},
		{"testdata/rain.wiki.gz", checkTestDataRain},
		{"testdata/better.wiki.gz", checkTestDataBetter},
	}
	for i, s := range files {
		//if i != 1 { continue }

		file, err := os.Open(s.file)
		if err != nil {
			t.Error("os.Open(%s): %v", s.file, err)
			continue
		}
		gz, err := gzip.NewReader(file)
		if err != nil {
			t.Error("gzip.NewReader(%s): %v", s.file, err)
			continue
		}
		b, err := ioutil.ReadAll(gz)
		if err != nil {
			t.Error("ioutil.ReadAll(%s): %v", s.file, err)
			continue
		}
		wiki, err :=  Parse(b)
		if err != nil {
			t.Error("Parse(%s): %v", s.file, err)
			continue
		}

		s.check(t, i, b, wiki)

		/*
		t.Logf("%v", s.file)
		for _, ent := range wiki.Entities {
			t.Logf("%v: %v", ent.Type, ent.Text)
		} /**/
	}
}
