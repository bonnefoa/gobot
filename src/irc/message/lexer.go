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
                       t, l.start, l.pos, l.input[l.start:l.pos])
        l.tokens <- token{t, l.input[l.start:l.pos]}
        l.start = l.pos
}

func (l * lexer) next() (r rune) {
        if l.pos >= len(l.input) {
                l.width = 0
                return EOL
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
        if l.input[ l.pos ] == ':' {
                l.next()
                l.emit(COLUMN)
        }
        for {
                if l.pos >= len(l.input) {
                        l.emit(INVALID)
                        return nil
                }
                if l.input[ l.pos ] == ' ' {
                        switch l.input[ l.start:l.pos ] {
                        case "QUIT":
                                l.emit(QUIT)
                        case "NICK":
                                l.emit(NICK)
                        case "PING":
                                l.emit(PING)
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
                                return maybeInMsg
                        }
                        return imInSpace
                }
                l.next()
        }
        panic("First word should return in the for loop")
}

func maybeInMsg(l *lexer) stateFn {
        l.next()
        l.ignore()
        for {
                if l.pos >= len(l.input) {
                        l.emit(INVALID)
                        return nil
                }
                if l.input[ l.pos ] == ' ' {
                        if l.input[ l.start:l.pos ] == "PRIVMSG" {
                                l.emit(PRIVMSG)
                                return imInSpace
                        }
                        l.emit(INVALID)
                        return nil
                }
                l.next()
        }
        panic("Shoud not happen")
}

func imInSpaceAfterColumn(l *lexer) stateFn {
        for {
                if l.next() != ' ' {
                        l.backup()
                        l.ignore()
                        return lexTextAfterColumn
                }
        }
        panic("Shoud not happen")
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

func lexTextAfterColumn(l *lexer) stateFn {
        for {
                if l.pos >= len(l.input) { break }
                if l.pos + 2 <= len(l.input) && l.input[ l.pos:l.pos + 2] == "\r\n" {
                        l.emit(WORD)
                        l.next()
                        l.next()
                        l.emit(EOF)
                        if l.pos >= len(l.input) {
                          return nil
                        }
                        return firstWord
                }
                if l.input[ l.pos ] == ' ' {
                        l.emit(WORD)
                        return imInSpaceAfterColumn
                }
                if l.next() == eof { break }
        }
        if l.pos > l.start { l.emit(WORD) }
        l.emit(EOF)
        return nil
}

func lexText(l *lexer) stateFn {
        log.Printf("Lex text, start %d, pos %d\n",
                       l.start, l.pos)
        if l.input[ l.pos ] == ':' {
                l.next()
                l.emit(COLUMN)
                return lexTextAfterColumn
        }
        for {
                if l.pos >= len(l.input) { break }
                if l.pos + 2 <= len(l.input) && l.input[ l.pos:l.pos + 2] == "\r\n" {
                        l.emit(WORD)
                        l.next()
                        l.next()
                        l.emit(EOF)
                        if l.pos >= len(l.input) {
                          return nil
                        }
                        return firstWord
                }
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
        case EOL:
                return "EOL"
        }
        if len(i.val) > 10 {
                return fmt.Sprintf("%.10q...", i.val)
        }
        return fmt.Sprintf("%q", i.val)
}
