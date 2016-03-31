# wiki

This is a Wikipedia text parser transforming wiki text into Abstract Syntax Tree
(AST). It was written in year 2013, I think it's better to open source it.

The [wiki markup](https://en.wikipedia.org/wiki/Help:Wiki_markup) format is 
documented on Wikipedia. This parser is not complying the 
[markup spec](https://www.mediawiki.org/wiki/Markup_spec), the entities (type of
the AST nodes) is abstracted from the [wiki markup](https://en.wikipedia.org/wiki/Help:Wiki_markup)
instead. Be aware that *bugs still exists*, so please be careful if you're using
it. You're welcome to submit your **bug fixes**, to do so, please use the 
[pull requests](https://help.github.com/articles/using-pull-requests/).

The following entities (AST nodes) are abstracted currently.

* [WikiEntityWiki]() - The root node.
* [WikiEntityText]() - Normal text without formatting.
* [WikiEntityTextBold]() - Bold text: `'''bold'''`
* [WikiEntityTextItalic]() - Italic text: `''italic''`
* [WikiEntityTextBoldItalic]() - Bold italic text: `'''''bold italic'''''`
* [WikiEntityHeading2]() - Level 2 Head Line: `== Heading text ==`
* [WikiEntityHeading3]() - Level 3 Head Line: `=== Heading text ===`
* [WikiEntityHeading4]() - Level 4 Head Line: `==== Heading text ====`
* [WikiEntityHeading5]() - Level 5 Head Line: `===== Heading text =====`
* [WikiEntityLinkExternal]() - External Linkage: `[http://example.com Link label]`
* [WikiEntityLinkInternal]() - Internal Linkage: `[[Title|Link label]]`
* [WikiEntityLinkInternalName]()
* [WikiEntityLinkInternalProp]()
* [WikiEntityTemplate]()
* [WikiEntityTemplateName]()
* [WikiEntityTemplateProp]()
* [WikiEntityTag]()
* [WikiEntityTagBeg]()
* [WikiEntityTagProp]()
* [WikiEntityTagEnd]()
* [WikiEntityListBulleted]()
* [WikiEntityListNumbered]()
* [WikiEntitySignature]()
* [WikiEntitySignatureTimestamp]()
* [WikiEntityIndent]()
* [WikiEntityHR]()
