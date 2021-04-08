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
  Infer(map[string]Type) Type
}

type Abstraction struct {
  Parameter string
  ParameterType Type
  Body Term
}

func (a Abstraction) Print() {
  fmt.Print("lambda ")
  fmt.Print(a.Parameter)
  fmt.Print(": ")
  a.ParameterType.Print()
  fmt.Print(". (")
  a.Body.Print()
  fmt.Print(")")
}

func (a Abstraction) Evaluate() Term {
  return Abstraction{a.Parameter, a.ParameterType, a.Body.Evaluate()}
}

func (a Abstraction) Substitute(v string, term Term) Term {
  if (a.Parameter == v) {
    return a
  }
  for _, freeVar := range term.FreeVars() {
    if (a.Parameter == freeVar) {
      // renaming required
      renamed := Abstraction{a.Parameter + "'", a.ParameterType, a.Body.Substitute(a.Parameter, Variable{a.Parameter + "'"})}
      return renamed.Substitute(v, term)
    }
  }
  return Abstraction{a.Parameter, a.ParameterType, a.Body.Substitute(v, term)}
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

func (a Abstraction) Infer(context map[string]Type) Type {
  context[a.Parameter] = a.ParameterType // QUESTION: Mutation seen outside?
  return FunctionType{a.ParameterType, a.Body.Infer(context)}
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

func (v Variable) Infer(context map[string]Type) Type {
  return context[v.Var]
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

func (a Application) Infer(context map[string]Type) Type {
  fun, ok := a.Function.Infer(context).(FunctionType)
  if !ok {
    fmt.Println("type error")
    os.Exit(1)
    return nil
  }
  arg := a.Argument.Infer(context)

  if fun.Argument == arg {
    return fun.Return
  } else {
    fmt.Println("type error")
    os.Exit(1)
    return nil
  }
}

type Type interface {
  Print()
}

type BaseType struct {

}

func (o BaseType) Print() {
  fmt.Print("o")
}

type FunctionType struct {
  Argument Type
  Return Type
}

func (f FunctionType) Print() {
  fmt.Print("(")
  f.Argument.Print()
  fmt.Print(" -> ")
  f.Return.Print()
  fmt.Print(")")
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
  typ, index := ParseType(text, index)
  index = Expect(text, index, ".")
  index = Expect(text, index, "(")
  index = SkipWhitespace(text, index)
  body, index := ParseImpl(text, index)
  index = SkipWhitespace(text, index)
  index = Expect(text, index, ")")
  return Abstraction{string(parameter), typ, body}, SkipWhitespace(text, index)
}

func ParseType(text string, index int) (Type, int) {
  index = SkipWhitespace(text, index)
  var typ Type
  if (text[index] == 'o') {
    index++
    index = SkipWhitespace(text, index)
    typ = BaseType{}
  } else if (text[index] == '(') {
    index++
    typ, index = ParseType(text, index)
  } else {
    fmt.Printf("invalid type")
    os.Exit(1)
    return nil, -1
  }
  if (text[index:index+2] == "->") {
    ret, index := ParseType(text, index + 2)
    return FunctionType{typ, ret}, index
  } else {
    return typ, index
  }
}

func Expect(text string, index int, s string) int {
  index = SkipWhitespace(text, index)
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
  fmt.Println("STLC with only one base type o and no base constants")
  fmt.Print("Enter expression: ")
  text, _ := reader.ReadString('\n')
  term := Parse(text)
  term.Print()
  fmt.Println("")
  typ := term.Infer(map[string]Type{})
  typ.Print()
  fmt.Println("")
  term.Evaluate().Print()
}
