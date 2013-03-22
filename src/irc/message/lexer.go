package message

import "fmt"
import "unicode/utf8"
import "strings"

type item struct {
  typ itemType
  val string
}

type itemType int
const (
  itemError itemType = iota
  itemWord
  itemEOF
)

type stateFn func(*lexer) stateFn

type lexer struct {
  name string
  input string
  state stateFn
  start int
  pos int
  width int
  items chan item
}

const ( eof rune = 0 )

func (l *lexer) errorf(format string, args ... interface{} ) stateFn {
  l.items <- item {
    itemError,
    fmt.Sprintf(format, args...),
  }
  return nil
}

func (l * lexer) emit(t itemType) {
  l.items <- item{t, l.input[l.start:l.pos]}
  l.start = l.pos
}

func (l * lexer) next() (r rune) {
  if l.pos >= len(l.input) {
    l.width = 0
    return eof
  }
  r, l.width = utf8.DecodeRuneInString(l.input[l.pos:])
  l.pos += l.width
  return r
}

func (l *lexer) ignore() {
  l.start = l.pos
}

func imInSpace(l *lexer) stateFn {
  for {
    if l.next() != ' ' {
      l.backup()
      l.ignore()
      return lexText
    }
  }
  panic("Shoud not happen\n")
}

func lexText(l *lexer) stateFn {
  for {
    if l.pos < len(l.input) && l.input[ l.pos ] == ' ' {
      l.emit(itemWord)
      return imInSpace
    }
    if l.next() == eof { break }
  }
  if l.pos > l.start {
    l.emit(itemWord)
  }
  l.emit(itemEOF)
  return nil
}

func (l *lexer) backup() {
  l.pos -= l.width
}

func (l *lexer) peek() rune {
  r := l.next()
  l.backup()
  return r
}

func (l * lexer) run() {
  for state := lexText; state != nil; {
    state = state(l)
  }
  close(l.items)
}

func (l *lexer) accept(valid string) bool {
  if strings.IndexRune(valid, l.next()) >= 0 {
    return true
  }
  l.backup()
  return false
}

func (l *lexer) nextItem() item {
  for {
    select {
    case item := <-l.items:
      return item
    default:
      l.state = l.state(l)
    }
  }
  panic("not reached")
}

func (l *lexer) acceptRun(valid string) {
  for strings.IndexRune(valid, l.next()) >= 0 {
  }
  l.backup()
}

func gen_lex(name, input string) *lexer {
  l := &lexer {
    name: name,
    input: input,
    state: lexText,
    items: make(chan item, 2),
  }
  return l
}

func (i item) String() string {
  switch i.typ {
  case itemEOF:
    return "EOF"
  case itemError:
    return i.val
  }
  if len(i.val) > 10 {
    return fmt.Sprintf("%.10q...", i.val)
  }
  return fmt.Sprintf("%q", i.val)
}
