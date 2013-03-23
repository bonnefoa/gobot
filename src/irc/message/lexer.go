package message

import "fmt"
import "unicode/utf8"
import "strings"
import "log"

type token struct {
        tok int
        val string
}

type stateFn func(*lexer) stateFn

type lexer struct {
        name string
        input string
        state stateFn
        start int
        pos int
        width int
        tokens chan token
}

const ( eof rune = 0 )

func (l * lexer) emit(t int) {
        log.Printf("Emitting %d, start %d, pos %d, str '%s'\n",
                       t, l.pos, l.start, l.input[l.start:l.pos])
        l.tokens <- token{t, l.input[l.start:l.pos]}
        l.start = l.pos
}

func (l * lexer) next() (r rune) {
        if l.pos >= len(l.input) {
                l.width = 0
                return eof
        }
        if l.input[l.pos:] == "\r\n" {
                l.width = 0
                return eof
        }
        r, l.width = utf8.DecodeRuneInString(l.input[l.pos:])
        l.pos += l.width
        return r
}

func (l *lexer) ignore() { l.start = l.pos }

func (l *lexer) backup() { l.pos -= l.width }

func firstWord(l *lexer) stateFn {
        log.Printf("First word state, start %d, pos %d\n",
                       l.start, l.pos)
        for {
                if l.pos < len(l.input) && l.input[ l.pos ] == ' ' {
                        switch l.input[ l.start:l.pos ] {
                        case "NICK":
                                l.emit(NICK)
                        case "PONG":
                                l.emit(PONG)
                        case "USER":
                                l.emit(USER)
                        case "JOIN":
                                l.emit(JOIN)
                        case "PRIVMSG":
                                l.emit(PRIVMSG)
                        default:
                                l.emit(WORD)
                        }
                        return imInSpace
                }
                l.next()
        }
        return lexText
}

func imInSpace(l *lexer) stateFn {
        for {
                if l.next() != ' ' {
                        l.backup()
                        l.ignore()
                        return lexText
                }
        }
        panic("Shoud not happen")
}

func lexText(l *lexer) stateFn {
        for {
                if l.pos > len(l.input) { break }
                if l.input[ l.pos ] == ' ' {
                        l.emit(WORD)
                        return imInSpace
                }
                if l.next() == eof { break }
        }
        if l.pos > l.start { l.emit(WORD) }
        l.emit(EOF)
        return nil
}

func (l *lexer) peek() rune {
        r := l.next()
        l.backup()
        return r
}

func (l * lexer) run() {
        for state := firstWord; state != nil; {
                state = state(l)
        }
        close(l.tokens)
}

func (l *lexer) accept(valid string) bool {
        if strings.IndexRune(valid, l.next()) >= 0 {
                return true
        }
        l.backup()
        return false
}

func (l *lexer) nextItem() token {
        for {
                select {
                case token := <-l.tokens:
                        return token
                default:
                        l.state = l.state(l)
                }
        }
        panic("not reached")
}

func (l *lexer) acceptRun(valid string) {
        for strings.IndexRune(valid, l.next()) >= 0 { }
        l.backup()
}

func gen_lex(name, input string) *lexer {
        l := &lexer {
                name: name,
                input: input,
                state: lexText,
                tokens: make(chan token, 2),
        }
        return l
}

func (i token) String() string {
        switch i.tok {
        case EOF:
                return "EOF"
        }
        if len(i.val) > 10 {
                return fmt.Sprintf("%.10q...", i.val)
        }
        return fmt.Sprintf("%q", i.val)
}
