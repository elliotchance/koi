# koi

## Introduction

This is the reference manual for the koi programming language.

## Lexical Elements

### Comments

There is only one form of single line comment, using `#`:

```
# This is a comment.

a = 123 # Comment same line.
```

### Identifiers

Identifiers are used for variable, function and type names. They must start with
a letter, then followed by any combinations of letters, numbers or underscore.

Identifiers that start with a capital letter are public. This means they can be
referenced outside of this package.

### Keywords

The following keywords are reserved and may not be used as identifiers.

```
and      break   continue   else
false    for     func       if
import   is      match      map
not      or      return     true
type
```

### Operators and Punctuation

```
+   +=   &    &=    and   ==   !=   (   )
-   -=   |    |=    or    <    <=   [   ]
*   *=   ~    ~=    not   >    >=   {   }
/   /=   <<   <<=   ++    =         .   ,
%   %=   >>   >>=   --              :   @
```

### Integer Literals

```
46
```

### Floating-point Literals

```
0.0
12.3
```

### Character Literals

Characters use single quotes and may contain any Unicode point.

```
'a'
'ä'
'本'
'\t'
```

Special escape values:

```
\a   U+0007 alert or bell
\b   U+0008 backspace
\f   U+000C form feed
\n   U+000A line feed or newline
\r   U+000D carriage return
\t   U+0009 horizontal tab
\v   U+000B vertical tab
\\   U+005C backslash
\'   U+0027 single quote  (valid escape only within rune literals)
\"   U+0022 double quote  (valid escape only within string literals)
```

### String Literals

String literals use double-quotes and can contain any character escape literal.

```
""
"abc"
"日本語"
"Hello, world!\n"
```

## Variables

Variables are defined with `=` and are assigned a value and type from that value
on creation.

```
a = 123       # int32
b = 1.23      # float64
c = 'a'       # int32
d = "hello"   # string
e = true      # bool

a = [1, 2, 3]         # []int32
b = ["hi", "there"]   # []string
b = ["hi", 123]       # []any
b = []                # []any
c = [].([]int8)       # []int8

a = {a: 1, b: 2}              # map[string]int32
b = {a: 1.2, b: 3.4}          # map[string]float64
c = {}                        # map[string]any
d = {}.(map[string]float64)   # map[string]float64
```

## Types

### Boolean Types

The type is `bool` and can only contain a `true` or `false` value.

### Numeric Types

```
int8    float32
int16   float64
int32
int64
```

### String Types

```
string
```

### Array Types

```
[]string
```

### Map Types

```
map[string]int32
```

### Custom Types

```
type emptyType() {}

# customType is:
#   field1 int32
#   Field2 bool
#   field3 int32
#   Field4 float64
type customType(field1 int32, Field2 bool) {
  field3 = 'a'
  Field4 = 1.2
}
```

### Function Types

```
func()
func(x int32) int32
func(a int32, _ int32, z float32) bool
func(a int32, b int32, z float32) (bool)
func(int32, int32, float64) (float64, []int32)
func(n int32) func(p T)
```

### Enum Types

```
enum Colors {
  red    # 0
  green  # 1
  blue   # 2
}

enum Names {
  bob  = 'Robert'
  jane = 'Jane'
}
```

## Declarations and Scope

### Exported Identifiers

### Method Declarations

```
type Person(name string) {
  func SayHello() {
    io.Println("Hi, {Name}!")
  }
}
```

### Index Expressions

### Type Assertions

### Operators

#### Operator Precedence

#### Arithmetic Operators

#### String Interpolation

#### Comparison Operators

#### Logical Operators

### Conversions

## Statements
