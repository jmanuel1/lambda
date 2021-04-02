package main

import (
  "fmt"
  "bufio"
  "os"
  "strings"
  "unicode"
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
  abs, ok := a.Function.(Abstraction)
  if ok {
    return abs.Body.Substitute(abs.Parameter, a.Argument).Evaluate()
  }
  return Application{a.Function.Evaluate(), a.Argument.Evaluate()}
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

func Parse(text string) Term {
  term, _ := ParseImpl(text, 0)
  return term
}

func ParseImpl(text string, index int) (Term, int) {
  index = SkipWhitespace(text, index)
  if text[index] == '(' {
    return ParseApplication(text, index)
  } else if strings.Index(string(([]rune(text))[index:]), "lambda") == 0 {
    return ParseAbstraction(text, index)
  } else {
    return ParseVariable(text, index)
  }
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
    fmt.Printf("invalid expression, expected %s at %d\n", s, index)
    os.Exit(1)
    return -1
  }
}

func ParseVariable(text string, index int) (Variable, int) {
  index = SkipWhitespace(text, index)
  variable, index := ([]rune(text))[index], index + 1
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
  reader := bufio.NewReader(os.Stdin)
  fmt.Print("Enter expression: ")
  text, _ := reader.ReadString('\n')
  term := Parse(text)
  term.Print()
  fmt.Println("")
  term.Evaluate().Print()
}
