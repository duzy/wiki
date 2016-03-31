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

	* [WikiEntityWiki]()
	* [WikiEntityText]()
	* [WikiEntityTextBold]()
	* [WikiEntityTextItalic]()
	* [WikiEntityTextBoldItalic]()
	* [WikiEntityHeading2]()
	* [WikiEntityHeading3]()
	* [WikiEntityHeading4]()
	* [WikiEntityHeading5]()
	* [WikiEntityLinkExternal]()
	* [WikiEntityLinkInternal]()
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
