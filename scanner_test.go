//
//  Copyright (C) 2013, Duzy Chan <code@duzy.info>, all rights reserverd.
//
package wiki

import (
	"testing"
)

func TestScanEntity(t *testing.T) {
	type result struct{
		parseType EntityType
		s string
	}
	tests := []struct{
		src string
		res []result
	}{
		/***** 0 *****/
		{`normal text`,
			[]result{
				{parseEntityText, `normal text`},
			},
		},
		/***** 1 *****/
		{`normal 'quote' normal`,
			[]result{
				{parseEntityText, `normal 'quote' normal`},
			},
		},
		/***** 2 *****/
		{`''italic''`,
			[]result{
				{parseEntityTextItalic, `''italic''`},
			},
		},
		/***** 3 *****/
		{`'''bold'''`,
			[]result{
				{parseEntityTextBold, `'''bold'''`},
			},
		},
		/***** 4 *****/
		{`'''''bold italic'''''`,
			[]result{
				{parseEntityTextBoldItalic, `'''''bold italic'''''`},
			},
		},
		/***** 5 *****/
		{`normal ''italic'' normal`,
			[]result{
				{parseEntityText, `normal `},
				{parseEntityTextItalic, `''italic''`},
				{parseEntityText, ` normal`},
			},
		},
		/***** 6 *****/
		{`normal '''bold''' normal`,
			[]result{
				{parseEntityText, `normal `},
				{parseEntityTextBold, `'''bold'''`},
				{parseEntityText, ` normal`},
			},
		},
		/***** 7 *****/
		{`normal ''italic'' normal '''bold''' normal '''''bold italic''''' normal`,
			[]result{
				{parseEntityText, `normal `},
				{parseEntityTextItalic, `''italic''`},
				{parseEntityText, ` normal `},
				{parseEntityTextBold, `'''bold'''`},
				{parseEntityText, ` normal `},
				{parseEntityTextBoldItalic, `'''''bold italic'''''`},
				{parseEntityText, ` normal`},
			},
		},
		/***** 8 *****/
		{`normal '''bold ''italic'' bold''' normal`,
			[]result{
				{parseEntityText, `normal `},
				{parseEntityTextBold, `'''bold ''italic'' bold'''`},
				{parseEntityText, ` normal`},
			},
		},
		/***** 9 *****/
		{`normal ''italic '''bold''' italic '''bold''' italic'' normal`,
			[]result{
				{parseEntityText, `normal `},
				{parseEntityTextItalic, `''italic '''bold''' italic '''bold''' italic''`},
				{parseEntityText, ` normal`},
			},
		},
		/***** 10 *****/
		{`normal ''italic 'abc' italic'' normal`,
			[]result{
				{parseEntityText, `normal `},
				{parseEntityTextItalic, `''italic 'abc' italic''`},
				{parseEntityText, ` normal`},
			},
		},
		/***** 11 *****/
		{`normal '''bold 'a''b''c' bold''' normal`,
			[]result{
				{parseEntityText, `normal `},
				{parseEntityTextBold, `'''bold 'a''b''c' bold'''`},
				{parseEntityText, ` normal`},
			},
		},


		/***** 12 *****/
		{`text [http://www.example.com label] text`,
			[]result{
				{parseEntityText, `text `},
				{parseEntityLink1, `[http://www.example.com label]`},
				{parseEntityText, ` text`},
			},
		},
		/***** 13 *****/
		{`text [http://www.example.com] text`,
			[]result{
				{parseEntityText, `text `},
				{parseEntityLink1, `[http://www.example.com]`},
				{parseEntityText, ` text`},
			},
		},
		/***** 14 *****/
		{`text [[Page Title|Link Label]] text`,
			[]result{
				{parseEntityText, `text `},
				{parseEntityLink2, `[[Page Title|Link Label]]`},
				{parseEntityText, ` text`},
			},
		},
		/***** 15 *****/
		{`text [[Page Title]] text`,
			[]result{
				{parseEntityText, `text `},
				{parseEntityLink2, `[[Page Title]]`},
				{parseEntityText, ` text`},
			},
		},

		/***** 16 *****/
		{`text [[''Page '''BOLD''' Title'']] text`,
			[]result{
				{parseEntityText, `text `},
				{parseEntityLink2, `[[''Page '''BOLD''' Title'']]`},
				{parseEntityText, ` text`},
			},
		},
		/***** 17 *****/
		{`text [[Page '''BOLD ''italic'' BOLD''' Title]] text`,
			[]result{
				{parseEntityText, `text `},
				{parseEntityLink2, `[[Page '''BOLD ''italic'' BOLD''' Title]]`},
				{parseEntityText, ` text`},
			},
		},

		/***** 18 *****/
		{`
text text
  * item 1
  * item 2
  * item 3
text text
`,
			[]result{
				{parseEntityText, `
text text`},
				{parseEntityListBulleted, `
  * item 1`},
				{parseEntityListBulleted, `
  * item 2`},
				{parseEntityListBulleted, `
  * item 3`},
				{parseEntityText, `
text text
`},
			},
		},
		/***** 19 *****/
		{`
text text
  # item 1
  # item 2
  # item 3
text text
`,
			[]result{
				{parseEntityText, `
text text`},
				{parseEntityListNumbered, `
  # item 1`},
				{parseEntityListNumbered, `
  # item 2`},
				{parseEntityListNumbered, `
  # item 3`},
				{parseEntityText, `
text text
`},
			},
		},

		/***** 20 *****/
		{`
  # item 1
  # item 2
  # item 3
`,
			[]result{
				{parseEntityListNumbered, `
  # item 1`},
				{parseEntityListNumbered, `
  # item 2`},
				{parseEntityListNumbered, `
  # item 3`},
				{parseEntityText, "\n"},
			},
		},

		/***** 21 *****/
		{`# item 1
# item 2
# item 3`,
			[]result{
				{parseEntityListNumbered, `# item 1`},
				{parseEntityListNumbered, `
# item 2`},
				{parseEntityListNumbered, `
# item 3`},
			},
		},

		/***** 22 *****/
		{`
# ''item 1, italic''
# '''item 2, bold'''
# [http://www.example.com]
# [[example]]
`,
			[]result{
				{parseEntityListNumbered, `
# ''item 1, italic''`},
				{parseEntityListNumbered, `
# '''item 2, bold'''`},
				{parseEntityListNumbered, `
# [http://www.example.com]`},
				{parseEntityListNumbered, `
# [[example]]`},
				{parseEntityListNumbered, "\n"},
			},
		},

		/***** 23 *****/
		{`
text text
  :indent 1
  ::indent 2
text text
:indent 3
text text`,
			[]result{
				{parseEntityText, `
text text`},
				{parseEntityIndent, `
  :indent 1`},
				{parseEntityIndent, `
  ::indent 2`},
				{parseEntityText, `
text text`},
				{parseEntityIndent, `
:indent 3`},
				{parseEntityText, `
text text`},
			},
		},
		/***** 24 *****/
		{`
text text
  :'''indent 1, bold'''
  ::''indent 2, italic''
text text`,
			[]result{
				{parseEntityText, `
text text`},
				{parseEntityIndent, `
  :'''indent 1, bold'''`},
				{parseEntityIndent, `
  ::''indent 2, italic''`},
				{parseEntityText, `
text text`},
			},
		},


		/***** 25 *****/
		{`
== header 2 ==
text text`,
			[]result{
				{parseEntityHeader2, `
== header 2 ==`},
				{parseEntityText, `
text text`},
			},
		},
		/***** 26 *****/
		{`
=== header 3 ===
text text`,
			[]result{
				{parseEntityHeader3, `
=== header 3 ===`},
				{parseEntityText, `
text text`},
			},
		},
		/***** 27 *****/
		{`
==== header 4 ====
text text`,
			[]result{
				{parseEntityHeader4, `
==== header 4 ====`},
				{parseEntityText, `
text text`},
			},
		},
		/***** 28 *****/
		{`
===== header 5 =====
text text`,
			[]result{
				{parseEntityHeader5, `
===== header 5 =====`},
				{parseEntityText, `
text text`},
			},
		},
		/***** 29 *****/
		{`== header 2 ==`,
			[]result{
				{parseEntityHeader2, `== header 2 ==`},
			},
		},

		/***** 30 *****/
		{`{{markup}}`,
			[]result{
				{parseEntityTemplate, `{{markup}}`},
			},
		},
		/***** 31 *****/
		{`text {{markup|prop1|prop2}} text`,
			[]result{
				{parseEntityText, `text `},
				{parseEntityTemplate, `{{markup|prop1|prop2}}`},
				{parseEntityText, ` text`},
			},
		},
		/***** 32 *****/
		{`text {{markup|[[prop1]]|[prop2]}} text`,
			[]result{
				{parseEntityText, `text `},
				{parseEntityTemplate, `{{markup|[[prop1]]|[prop2]}}`},
				{parseEntityText, ` text`},
			},
		},
		/***** 33 *****/
		{`text {{markup
|[[prop1]]
|[prop2]
}} text`,
			[]result{
				{parseEntityText, `text `},
				{parseEntityTemplate, `{{markup
|[[prop1]]
|[prop2]
}}`},
				{parseEntityText, ` text`},
			},
		},


		/***** 34 *****/
		{`text <ref name="test" /> text`,
			[]result{
				{parseEntityText, `text `},
				{parseEntityTag, `<ref name="test" />`},
				{parseEntityText, ` text`},
			},
		},
		/***** 35 *****/
		{`text <ref name="test">test test</ref>`,
			[]result{
				{parseEntityText, `text `},
				{parseEntityTagBeg, `<ref name="test">`},
				{parseEntityText, `test test`},
				{parseEntityTagEnd, `</ref>`},
				{parseEntityText, ` text`},
			},
		},
		/***** 36 *****/
		{`text <any name="value" /> text <any name="test">test test</any> text`,
			[]result{
				{parseEntityText, `text `},
				{parseEntityTag, `<any name="value" />`},
				{parseEntityText, ` text `},
				{parseEntityTagBeg, `<any name="test">`},
				{parseEntityText, `test test`},
				{parseEntityTagEnd, `</any>`},
				{parseEntityText, ` text`},
			},
		},

		// BUG fixes:
		/***** 37 *****/
		{`'''bold''bold-italic[[link''italic''link]]bold-italic''bold'''`,
			[]result{
				{parseEntityTextBold, `'''bold''bold-italic[[link''italic''link]]bold-italic''bold'''`},
			},
		},
		// BUG fixes:
		/***** 38 *****/
		{`
{{markup}}

== English ==
{{markup}}

=== Noun ===
text text text
`,
			[]result{
				{parseEntityText, "\n"},
				{parseEntityTemplate, `{{markup}}`},
				{parseEntityText, "\n"},
				{parseEntityHeader2, `
== English ==`},
				{parseEntityText, "\n"},
				{parseEntityTemplate, `{{markup}}`},
				{parseEntityText, "\n"},
				{parseEntityHeader3, `
=== Noun ===`},
				{parseEntityText, `
text text text
`},
			},
		},

		/***** 39 *****/
		{`
  # item 1
  ## item 2
  ### item 3
  #### item 4
  ##### item 5
`,
			[]result{
				{parseEntityListNumbered, `
  # item 1`},
				{parseEntityListNumbered, `
  ## item 2`},
				{parseEntityListNumbered, `
  ### item 3`},
				{parseEntityListNumbered, `
  #### item 4`},
				{parseEntityListNumbered, `
  ##### item 5`},
				{parseEntityText, "\n"},
			},
		},
		/***** 40 *****/
		{`
  * item 1
  ** item 2
  *** item 3
  **** item 4
  ***** item 5
`,
			[]result{
				{parseEntityListBulleted, `
  * item 1`},
				{parseEntityListBulleted, `
  ** item 2`},
				{parseEntityListBulleted, `
  *** item 3`},
				{parseEntityListBulleted, `
  **** item 4`},
				{parseEntityListBulleted, `
  ***** item 5`},
				{parseEntityText, "\n"},
			},
		},
		/***** 41 *****/
		{`
  : item 1
  :: item 2
  ::: item 3
  :::: item 4
  ::::: item 5
`,
			[]result{
				{parseEntityIndent, `
  : item 1`},
				{parseEntityIndent, `
  :: item 2`},
				{parseEntityIndent, `
  ::: item 3`},
				{parseEntityIndent, `
  :::: item 4`},
				{parseEntityIndent, `
  ::::: item 5`},
				{parseEntityText, "\n"},
			},
		},
		/***** 42 *****/
		{`
  : item 1
  *: item 2
  #:: item 3
  :*#*: item 4
  :#*#*: item 5
`,
			[]result{
				{parseEntityIndent, `
  : item 1`},
				{parseEntityListBulleted, `
  *: item 2`},
				{parseEntityListNumbered, `
  #:: item 3`},
				{parseEntityIndent, `
  :*#*: item 4`},
				{parseEntityIndent, `
  :#*#*: item 5`},
				{parseEntityText, "\n"},
			},
		},
		/***** 43 *****/ //BUG fixing
		{`
===Pronoun===
'''any'''

# Any thing(s) or person(s).
#: '''''Any''' may apply.''

====Translations====
{{trans-top|Any things or persons}}
`,
			[]result{
				{parseEntityHeader3, `
===Pronoun===`},
				{parseEntityText, "\n"},
				{parseEntityTextBold, "'''any'''"},
				{parseEntityText, "\n"},
				{parseEntityListNumbered, `
# Any thing(s) or person(s).`},
				{parseEntityListNumbered, `
#: '''''Any''' may apply.''`},
				{parseEntityText, "\n"},
				{parseEntityHeader4, `
====Translations====`},
				{parseEntityText, "\n"},
				{parseEntityTemplate, "{{trans-top|Any things or persons}}"},
				{parseEntityText, "\n"},
			},
		},
		/***** 44 *****/ //BUG fixing
		{`{{in '''A'''''a'' out}}`,
			[]result{
				{parseEntityTemplate, "{{in '''A'''''a'' out}}"},
			},
		},
		/***** 45 *****/ //BUG fixing
		{`{{in ''a'''''A''' out}}`,
			[]result{
				{parseEntityTemplate, "{{in ''a'''''A''' out}}"},
			},
		},
		/***** 46 *****/ //BUG fixing
		{`in '''A'''''a'' out`,
			[]result{
				{parseEntityText, "in "},
				{parseEntityTextBold, "'''A'''"},
				{parseEntityTextItalic, "''a''"},
				{parseEntityText, " out"},
			},
		},
		/***** 47 *****/ //BUG fixing
		{`in ''a'''''A''' out`,
			[]result{
				{parseEntityText, "in "},
				{parseEntityTextItalic, "''a''"},
				{parseEntityTextBold, "'''A'''"},
				{parseEntityText, " out"},
			},
		},
		/***** 48 *****/ //BUG fixing
		{`in '''''a''A''' out`,
			[]result{
				{parseEntityText, "in "},
				{parseEntityTextBold, "'''''a''A'''"},
				{parseEntityText, " out"},
			},
		},
		/***** 49 *****/ //BUG fixing
		{`in '''''A'''a'' out`,
			[]result{
				{parseEntityText, "in "},
				{parseEntityTextItalic, "'''''A'''a''"},
				{parseEntityText, " out"},
			},
		},

		/***** 50 *****/ //BUG fixing
		{`
===Etymology 2===
{{abbreviation-old|mul}} of {{term|atto-|lang=mul}}, from {{etyl|da|mul}} and {{etyl|no|mul}} {{term|atten||eighteen|lang=no}}.

====Symbol====
{{head|mul|symbol}}

# {{non-gloss definition|[[atto-]], the prefix for 10<sup>-18</sup> in the [[International System of Units]].}}

===Etymology 3===
From {{etyl|la|mul}} {{term|annus|lang=la}}

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
			[]result{
				{parseEntityHeader3, `
===Etymology 2===`},
				{parseEntityText, "\n"},
				{parseEntityTemplate, "{{abbreviation-old|mul}}"},
				{parseEntityText, " of "},
				{parseEntityTemplate, "{{term|atto-|lang=mul}}"},
				{parseEntityText, ", from "},
				{parseEntityTemplate, "{{etyl|da|mul}}"},
				{parseEntityText, " and "},
				{parseEntityTemplate, "{{etyl|no|mul}}"},
				{parseEntityText, " "},
				{parseEntityTemplate, "{{term|atten||eighteen|lang=no}}"},
				{parseEntityText, ".\n"},
				{parseEntityHeader4, `
====Symbol====`},
				{parseEntityText, "\n"},
				{parseEntityTemplate, "{{head|mul|symbol}}"},
				{parseEntityText, "\n"},
				{parseEntityListNumbered, `
# {{non-gloss definition|[[atto-]], the prefix for 10<sup>-18</sup> in the [[International System of Units]].}}`},
				{parseEntityText, "\n"},
				{parseEntityHeader3, `
===Etymology 3===`},
				{parseEntityText, "\nFrom "},
				{parseEntityTemplate, "{{etyl|la|mul}}"},
				{parseEntityText, " "},
				{parseEntityTemplate, "{{term|annus|lang=la}}"},
				{parseEntityText, "\n"},
				{parseEntityHeader3, "\n===Etymology 4==="},
				{parseEntityText, "\n"},
				{parseEntityHeader4, "\n====Symbol===="},
				{parseEntityText, "\n"},
				{parseEntityTemplate, "{{head|mul|symbol}}"},
				{parseEntityText, "\n"},
				{parseEntityListNumbered, "\n# {{context|physics|lang=mul}} [[acceleration]]"},
				{parseEntityText, "\n"},
				{parseEntityText, "\n"},
				{parseEntityTemplate, "{{Letter|page=A\n|NATO=Alpha\n|Morse=·–\n|Character=A1\n|Braille=⠁\n}}"},
				{parseEntityText, "\n"},
				{parseEntityTagBeg, `<gallery caption="Letter styles" perrow=3>`},
				{parseEntityText, "\nImage:Latin A.png|Capital and lowercase versions of "},
				{parseEntityTextBold, "'''A'''"},
				{parseEntityText, ", in normal and italic type\nFile:Fraktur letter A.png|Uppercase and lowercase "},
				{parseEntityTextBold, "'''A'''"},
				{parseEntityText, " in "},
				{parseEntityLink2, "[[Fraktur]]"},
				{parseEntityText, "\nFile:UncialA-01.svg|Approximate form of Greek upper case Α (a, “alpha”) that was the source for both common variants of "},
				{parseEntityTextItalic, "''a''"},
				{parseEntityTextBold, "'''A'''"},
				{parseEntityText, " in "},
				{parseEntityLink2, "[[uncial]]"},
				{parseEntityText, " script\n"},
				{parseEntityTagEnd, `</gallery>`},
				{parseEntityText, "\n"},
				{parseEntityHR, "\n----"},
				{parseEntityText, "\n"},
				{parseEntityHeader2, "\n==English=="},
				{parseEntityText, "\n"},
				{parseEntityHeader3, "\n===Etymology 1==="},
				{parseEntityText, "\n"},
			},
		},
		/***** 51 *****/ //BUG fixing
		{`{{context|cricket|lang=en}} In line with the [[batsman]]'s [[popping crease]].`,
			[]result{
				{parseEntityTemplate, "{{context|cricket|lang=en}}"},
				{parseEntityText, " In line with the "},
				{WikiEntityLinkInternal, "[[batsman]]"},
				{parseEntityText, "'s "},
				{WikiEntityLinkInternal, "[[popping crease]]"},
				{parseEntityText, "."},
			},
		},
		/***** 52 *****/ //BUG fixing
		{`{{context|cricket|lang=en}} In line with the [[batsman]]{s [[popping crease]].`,
			[]result{
				{parseEntityTemplate, "{{context|cricket|lang=en}}"},
				{parseEntityText, " In line with the "},
				{WikiEntityLinkInternal, "[[batsman]]"},
				{parseEntityText, "{s "},
				{WikiEntityLinkInternal, "[[popping crease]]"},
				{parseEntityText, "."},
			},
		},
		/***** 53 *****/ //BUG fixing
		{`{{context|cricket|lang=en}} In line with the [[batsman]]<s />[[popping crease]].`,
			[]result{
				{parseEntityTemplate, "{{context|cricket|lang=en}}"},
				{parseEntityText, " In line with the "},
				{WikiEntityLinkInternal, "[[batsman]]"},
				{parseEntityTag, "<s />"},
				{WikiEntityLinkInternal, "[[popping crease]]"},
				{parseEntityText, "."},
			},
		},
		/***** 54 *****/ //BUG fixing
		{`{{context|cricket|lang=en}} In line with the [[batsman]][[s]] [[popping crease]].`,
			[]result{
				{parseEntityTemplate, "{{context|cricket|lang=en}}"},
				{parseEntityText, " In line with the "},
				{WikiEntityLinkInternal, "[[batsman]]"},
				{WikiEntityLinkInternal, "[[s]]"},
				{parseEntityText, " "},
				{WikiEntityLinkInternal, "[[popping crease]]"},
				{parseEntityText, "."},
			},
		},
		/***** 55 *****/ //BUG fixing
		{`-`,
			[]result{
				{parseEntityText, "-"},
			},
		},
		/***** 56 *****/ //BUG fixing
		{`--`,
			[]result{
				{parseEntityText, "--"},
			},
		},
		/***** 57 *****/ //BUG fixing
		{`---`,
			[]result{
				{parseEntityText, "---"},
			},
		},
		/***** 58 *****/ //BUG fixing
		{`----`,
			[]result{
				{parseEntityHR, "----"},
			},
		},
		/***** 59 *****/ //BUG fixing
		{`from pre-Germanic {{m|ine-pro|*Hréǵ-no}}-, from`,
			[]result{
				{parseEntityText, "from pre-Germanic "},
				{parseEntityTemplate, "{{m|ine-pro|*Hréǵ-no}}"},
				{parseEntityText, "-, from"},
			},
		},
		/***** 60 *****/ //BUG fixing
		{`=`,
			[]result{
				{parseEntityText, "="},
			},
		},
	}
	for i, tc := range tests {
		//if i != 44 { continue }

		s := ""
		data := []byte(tc.src)
		scan := new(scanner)
		n := 0
		for ; ; n++ {
			entity, rest, err := scan.next(data)
			s := string(entity)

			if err != nil {
				t.Errorf("TestScanEntity: [%d: next] error: %v", i, err)
				break
			}

			//t.Logf("TestScanEntity: [%d: next] entity: %v, %v (rest=%v)", i, scan.parsing, string(entity), string(rest))
			//t.Logf("TestScanEntity: [%d: next] entity: %v, %v", i, scan.parsing, string(entity))

			if len(tc.res) == n && (s == "\n" || s == ""){
				break
			}
			if len(tc.res) <= n {
				t.Errorf("TestScanEntity: [%d: len] %v <= %v (%v)", i, len(tc.res), n, string(entity))
				break
			}

			res := tc.res[n]
			if res.s != string(entity) {
				t.Errorf("TestScanEntity: [%d, %d: entity] %v != %v, parsing=%v", i, n, s, res.s, scan.parsing)
				break
			}

			if len(scan.parsing) != 0 {
				t.Errorf("TestScanEntity: [%d: pos] %d, parsing %v is not empty, %v", i, n, scan.parsing, s)
				break
			}

			if res.parseType != scan.state {
				if s != "\n" {
					t.Errorf("TestScanEntity: [%d: type] %v != %v (%v, %v) (parsing=%v)", i, scan.state, res.parseType, s, res.s, scan.parsing)
				}
				break
			}

			if data = rest; data == nil || len(data) <= 0 {
				break
			}
		}
		if n == len(tc.res) && s == "" {
			//continue
		}
		if n+1 != len(tc.res) && s != "" {
			t.Errorf("TestScanEntity: [%d: len] %v != %v (%v, %s)", i, n, len(tc.res), scan.parsing, s)
			t.Logf("TestScanEntity: [%d] %v", i, tc.src)
		}
	}
}
