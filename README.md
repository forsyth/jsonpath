# jsonpath
Experimental implementation of JSONpath

This is a placeholder for an eventual implementation of lexing and parsing of JSONpath.

IETF grammar:

~~~~
json-path = root-selector *selector
   root-selector = "$"               ; $ selects document root node
   
selector = dot-child              ; see below for alternatives
   dot-child = "." dot-child-name / ; .<dot-child-name>
               "." "*"             ; .*
   dot-child-name = 1*(
                      "-" /         ; -
                      DIGIT /
                      ALPHA /
                      "_" /         ; _
                      %x80-10FFFF    ; any non-ASCII Unicode character
                    )
   DIGIT =  %x30-39                  ; 0-9
   ALPHA = %x41-5A / %x61-7A         ; A-Z / a-z
  
  selector =/ union
   union = "[" ws union-elements ws "]" ; [...]
   ws = *" "                             ; zero or more spaces
   union-elements = union-element /
                    union-element ws "," ws union-elements
                                          ; ,-separated list
  
  double-quoted = dq-unescaped /
             escape (
                 %x22 /         ; "    quotation mark  U+0022
                 "/" /          ; /    solidus         U+002F
                 "\" /          ; \    reverse solidus U+005C
                 "b" /          ; b    backspace       U+0008
                 "f" /          ; f    form feed       U+000C
                 "n" /          ; n    line feed       U+000A
                 "r" /          ; r    carriage return U+000D
                 "t" /          ; t    tab             U+0009
                 "u" 4HEXDIG )  ; uXXXX                U+XXXX


   dq-unescaped = %x20-21 / %x23-5B / %x5D-10FFFF

   single-quoted = sq-unescaped /
             escape (
                 "'" /          ; '    apostrophe      U+0027
                 "/" /          ; /    solidus         U+002F
                 "\" /          ; \    reverse solidus U+005C
                 "b" /          ; b    backspace       U+0008
                 "f" /          ; f    form feed       U+000C
                 "n" /          ; n    line feed       U+000A
                 "r" /          ; r    carriage return U+000D
                 "t" /          ; t    tab             U+0009
                 "u" 4HEXDIG )  ; uXXXX                U+XXXX

   sq-unescaped = %x20-26 / %x28-5B / %x5D-10FFFF

   escape = "\"                 ; \

   HEXDIG =  DIGIT / "A" / "B" / "C" / "D" / "E" / "F"
                                 ; case insensitive hex digit
  
  
   union-element =/ array-index / array-slice
  
  array-index = integer

   integer = ["-"] ("0" / (DIGIT1 *DIGIT))
                               ; optional - followed by 0 or
                               ; sequence of digits with no leading zero
   DIGIT1 = %x31-39            ; non-zero digit

  
   array-slice = [ start ] ws ":" ws [ end ]
                      [ ws ":" ws [ step ] ]
   start = integer
   end = integer
   step = integer

  Bits of text from the IETF doc:
  +=======+==================+=====================================+
    | XPath | JSONPath         | Description                         |
    +=======+==================+=====================================+
    | /     | $                | the root element/item               |
    +-------+------------------+-------------------------------------+
    | .     | @                | the current element/item            |
    +-------+------------------+-------------------------------------+
    | /     | "." or "[]"      | child operator                      |
    +-------+------------------+-------------------------------------+
    | ..    | n/a              | parent operator                     |
    +-------+------------------+-------------------------------------+
    | //    | ..               | nested descendants (JSONPath        |
    |       |                  | borrows this syntax from E4X)       |
    +-------+------------------+-------------------------------------+
    | *     | *                | wildcard: All elements/items        |
    |       |                  | regardless of their names           |
    +-------+------------------+-------------------------------------+
    | @     | n/a              | attribute access: JSON data items   |
    |       |                  | do not have attributes              |
    +-------+------------------+-------------------------------------+
    | []    | []               | subscript operator: XPath uses it   |
    |       |                  | to iterate over element collections |
    |       |                  | and for predicates; native array    |
    |       |                  | indexing as in JavaScript here      |
    +-------+------------------+-------------------------------------+
    | |     | [,]              | Union operator in XPath (results in |
    |       |                  | a combination of node sets);        |
    |       |                  | JSONPath allows alternate names or  |
    |       |                  | array indices as a set              |
    +-------+------------------+-------------------------------------+
    | n/a   | [start:end:step] | array slice operator borrowed from  |
    |       |                  | ES4                                 |
    +-------+------------------+-------------------------------------+
    | []    | ?()              | applies a filter (script)           |
    |       |                  | expression                          |
    +-------+------------------+-------------------------------------+
    | n/a   | ()               | expression engine                   |
    +-------+------------------+-------------------------------------+
    | ()    | n/a              | grouping in Xpath                   |
    +-------+------------------+-------------------------------------+
  
  +=================+===================+
                  | Escape Sequence | Unicode Character |
                  +=================+===================+
                  |     "" %x22     |       U+0022      |
                  +-----------------+-------------------+
                  |      "" "'"     |       U+0027      |
                  +-----------------+-------------------+
                  |      "" "/"     |       U+002F      |
                  +-----------------+-------------------+
                  |      "" ""      |       U+005C      |
                  +-----------------+-------------------+
                  |      "" "b"     |       U+0008      |
                  +-----------------+-------------------+
                  |      "" "f"     |       U+000C      |
                  +-----------------+-------------------+
                  |      "" "n"     |       U+000A      |
                  +-----------------+-------------------+
                  |      "" "r"     |       U+000D      |
                  +-----------------+-------------------+
                  |      "" "t"     |       U+0009      |
                  +-----------------+-------------------+
                  |     "" uXXXX    |       U+XXXX      |
                  +-----------------+-------------------+
~~~~
