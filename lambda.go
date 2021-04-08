package main

import (
  "fmt"
  "bufio"
  "os"
  "strings"
  "unicode"
  "strconv"
)

type Term interface {
  Print()
  Evaluate() Term
  Substitute(string, Term) Term
  FreeVars() []string
}

type Abstraction struct {
  Parameter string
  Body Term
}

func (a Abstraction) Print() {
  fmt.Print("lambda ")
  fmt.Print(a.Parameter)
  fmt.Print(": (")
  a.Body.Print()
  fmt.Print(")")
}

func (a Abstraction) Evaluate() Term {
  return Abstraction{a.Parameter, a.Body.Evaluate()}
}

func (a Abstraction) Substitute(v string, term Term) Term {
  if (a.Parameter == v) {
    return a
  }
  for _, freeVar := range term.FreeVars() {
    if (a.Parameter == freeVar) {
      // renaming required
      renamed := Abstraction{a.Parameter + "'", a.Body.Substitute(a.Parameter, Variable{a.Parameter + "'"})}
      return renamed.Substitute(v, term)
    }
  }
  return Abstraction{a.Parameter, a.Body.Substitute(v, term)}
}

func (a Abstraction) FreeVars() []string {
  bodyFreeVars := a.Body.FreeVars()
  var freeVars []string
  for _, v := range bodyFreeVars {
    if (v != a.Parameter) {
      freeVars = append(freeVars, v)
    }
  }
  return freeVars
}

type Variable struct {
  Var string
}

func (v Variable) Print()  {
  fmt.Print(v.Var)
}

func (v Variable) Evaluate() Term {
  value, ok := symbols[v.Var]
  if ok {
    return value
  }
  return v
}

func (v Variable) Substitute(x string, term Term) Term {
  if (x == v.Var) {
    return term
  }
  return v
}

func (v Variable) FreeVars() []string {
  return []string{v.Var}
}

type Application struct {
  Function Term
  Argument Term
}

func (a Application) Print() {
  fmt.Print("(")
  a.Function.Print()
  fmt.Print(") (")
  a.Argument.Print()
  fmt.Print(")")
}

func (a Application) Evaluate() Term {
  // a.Print()
  // fmt.Println()
  function := a.Function.Evaluate()
  abs, ok := function.(Abstraction)
  if ok {
    return abs.Body.Substitute(abs.Parameter, a.Argument).Evaluate()
  }
  num, ok := function.(Number)
  if ok {
    var iteratedBody Term
    iteratedBody = Variable{"'z"} // hopefully never gets captured
    i := 0
    for i < num.Number {
      iteratedBody = Application{a.Argument, iteratedBody}
      i++
    }
    return Abstraction{"'z", iteratedBody}.Evaluate()
  }
  argument := a.Argument.Evaluate()
  _, ok = function.(Succ)
  if ok {
    // fmt.Print("original argument ")
    // a.Argument.Print()
    // fmt.Println()
    // fmt.Print("argument ")
    // argument.Print()
    // fmt.Println()
    num, ok = argument.(Number)
    if ok {
      return Number{num.Number + 1}
    }
    abs, ok = argument.(Abstraction)
    if ok {
      argumentS := Application{argument, Variable{"'s"}}
      return Abstraction{"'s", Abstraction{"'z", Application{argumentS, Application{Variable{"'s"}, Variable{"'z"}}}}}
    }
    // continue by default
  }
  return Application{function, argument}
}

func (a Application) Substitute(v string, term Term) Term {
  return Application{a.Function.Substitute(v, term), a.Argument.Substitute(v, term)}
}

func (a Application) FreeVars() []string {
  freeVars := a.Function.FreeVars()
  for _, v := range a.Argument.FreeVars() {
    freeVars = append(freeVars, v)
  }
  return freeVars
}


type Let struct {
  Name string
  Value Term
}

func (l Let) Print() {
  fmt.Printf("let %s = ", l.Name)
  l.Value.Print()
}

var symbols map[string]Term

func (l Let) Evaluate() Term {
  symbols[l.Name] = l.Value
  return l.Value
}

func (l Let) Substitute(v string, term Term) Term {
  fmt.Println("Cannot substitute into a let")
  os.Exit(1)
  return nil
}

func (l Let) FreeVars() []string {
  fmt.Println("Cannot get free variables from let")
  os.Exit(1)
  return nil
}

type Number struct {
  Number int
}

func (n Number) Print() {
  fmt.Printf("%d", n.Number)
}

func (n Number) Evaluate() Term {
  return n
}

func (n Number) Substitute(v string, term Term) Term {
  return n
}

func (n Number) FreeVars() []string {
  return []string{}
}

type Succ struct {

}

func (s Succ) Print() {
  fmt.Print("succ")
}

func (s Succ) Evaluate() Term {
  return s
}

func (s Succ) Substitute(v string, term Term) Term {
  return s
}

func (s Succ) FreeVars() []string {
  return []string{}
}

func Parse(text string) Term {
  term, index := ParseImpl(text, 0)
  ExpectEOF(text, index)
  return term
}

func ExpectEOF(text string, index int) {
  if len([]rune(text)) != index {
    fmt.Printf("Expected EOF at %d\n", index)
    os.Exit(1)
  }
}

func ParseImpl(text string, index int) (Term, int) {
  index = SkipWhitespace(text, index)
  if text[index] == '(' {
    return ParseApplication(text, index)
  } else if strings.Index(string(([]rune(text))[index:]), "lambda") == 0 {
    return ParseAbstraction(text, index)
  } else if strings.Index(string(([]rune(text))[index:]), "let") == 0 {
    return ParseLet(text, index)
  } else if unicode.IsDigit([]rune(text)[index]) {
    return ParseNumber(text, index)
  } else {
    return ParseVariable(text, index)
  }
}

func ParseNumber(text string, index int) (Number, int) {
  num := ""
  length := len([]rune(text))
  start := index
  for index < length && unicode.IsDigit([]rune(text)[index]) {
    num += string([]rune(text)[index])
    index++
  }
  if start == index {
    fmt.Printf("Expected number at %d\n", index)
    os.Exit(1)
  }
  i, _ := strconv.Atoi(num)
  return Number{i}, index
}

func ParseLet(text string, index int) (Let, int) {
  index = SkipWhitespace(text, index)
  index = Expect(text, index, "let ")
  index = SkipWhitespace(text, index)
  name, index := ExpectID(text, index)
  index = SkipWhitespace(text, index)
  index = Expect(text, index, "=")
  index = SkipWhitespace(text, index)
  value, index := ParseImpl(text, index)
  return Let{name, value}, index
}

func ExpectID(text string, index int) (string, int) {
  id := ""
  length := len([]rune(text))
  if index >= length || !unicode.IsLetter([]rune(text)[index]) {
    fmt.Printf("Expected ID at %d\n", index)
    os.Exit(1)
  }
  for index < length && (unicode.IsDigit([]rune(text)[index]) || unicode.IsLetter([]rune(text)[index])) {
    id += string([]rune(text)[index])
    index++
  }
  return id, index
}

func ParseApplication(text string, index int) (Application, int) {
  index = SkipWhitespace(text, index)
  index++ // (
  var function Term
  var argument Term
  function, index = ParseImpl(text, index)
  index = SkipWhitespace(text, index)
  index++ // )
  index = SkipWhitespace(text, index)
  index++ // (
  argument, index = ParseImpl(text, index)
  index = SkipWhitespace(text, index)
  index++ // )
  return Application{function, argument}, SkipWhitespace(text, index)
}

func ParseAbstraction(text string, index int) (Abstraction, int) {
  index = SkipWhitespace(text, index)
  index = Expect(text, index, "lambda ")
  index = SkipWhitespace(text, index)
  parameter, index := ([]rune(text))[index], index + 1
  index = Expect(text, index, ":")
  index = SkipWhitespace(text, index)
  index = Expect(text, index, "(")
  index = SkipWhitespace(text, index)
  body, index := ParseImpl(text, index)
  index = SkipWhitespace(text, index)
  index = Expect(text, index, ")")
  return Abstraction{string(parameter), body}, SkipWhitespace(text, index)
}

func Expect(text string, index int, s string) int {
  if strings.Index(string(([]rune(text))[index:]), s) == 0 {
    return index + len(s)
  } else {
    fmt.Printf("invalid expression, expected %s at %d, found %c instead\n", s, index, []rune(text)[index])
    os.Exit(1)
    return -1
  }
}

func ParseVariable(text string, index int) (Variable, int) {
  index = SkipWhitespace(text, index)
  variable, index := ExpectID(text, index)
  index = SkipWhitespace(text, index)
  return Variable{string(variable)}, index
}

func SkipWhitespace(text string, index int) int {
  for _, v := range ([]rune(text))[index:] {
    if unicode.IsSpace(v) {
      index++
    } else {
      break
    }
  }
  return index
}

func main() {
  symbols = make(map[string]Term)
  symbols["succ"] = Succ{}
  pair := Application{Application{Variable{"b"}, Variable{"f"}}, Variable{"s"}}
  symbols["pair"] = Abstraction{"f", Abstraction{"s", Abstraction{"b", pair}}}
  symbols["fst"] = Abstraction{"p", Application{Variable{"p"}, Abstraction{"t", Abstraction{"f", Variable{"t"}}}}}
  symbols["snd"] = Abstraction{"p", Application{Variable{"p"}, Abstraction{"t", Abstraction{"f", Variable{"f"}}}}}
  reader := bufio.NewReader(os.Stdin)
  for {
    fmt.Print("Enter expression: ")
    text, _ := reader.ReadString('\n')
    term := Parse(text)
    term.Print()
    fmt.Println("")
    term.Evaluate().Print()
    fmt.Println("")
  }
}
