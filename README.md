# jsonpath
Experimental Go implementation of JSONpath

This is a placeholder for an eventual Go implementation of lexing and parsing of JSONpath.

Paths have the following grammar:

path ::= "$" step*
step ::= "." member | ".." member | "[" subscript "]" | ".." "[" subscript "]"
member ::= "*" | identifier | expr | integer
subscript ::= subscript-expression | union-element ("," union-element)
subscript-expression ::= "*" | expr | filter
union-element ::=  array-index | string-literal | array-slice   // could include identifier?
array-index ::= integer
array-slice ::= start? ":" end? (":" stride?)?
start ::= integer | expr
end ::= integer | expr
stride ::= integer | expr
expr ::= "(" script-expression ")"
filter ::= "?(" script-expression ")"
step ::= ...  "[" subscript "]" ... | ".." "[" subscript "]"
subscript ::= subscript-expression | union-element ("," union-element)
subscript-expression ::= "*" | expr | filter
union-element ::=  array-index | string-literal | array-slice   // could include identifier?
array-index ::= integer
array-slice ::= start? ":" end? (":" step?)?
member ::= "*" | identifier | expr | integer
expr ::= "(" script-expression ")"

Script expressions (filters and calculations) share the same syntax:

script-expression ::= e   // both filters and values share the same syntax
e ::= primary | e binary-op e | unary-op e
binary-op ::= "+" | "-" | "*" | "/" | "%" | "<" | ">" | ">=" | "<=" | "==" | "!=" | "~" | "in" | "nin"  | "&&" | "||"
unary-op ::= "-" | "!"
unary ::= ("-" | "!")+ primary
primary ::= primary1 ("(" e-list ")" | "[" e-list "]" | "." identifier)*
e-list ::= e ("," e)*
primary1 ::= identifier | integer | string | "/" re "/" | "@" | "$" | "(" e ")"
