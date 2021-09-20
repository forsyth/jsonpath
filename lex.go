package JSONPath

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

var (
	ErrUnclosedString = errors.New("unclosed string literal")
)

func lex(r *rd) (Op, interface{}, error) {
	for isSpace(r.look()) {
		r.get()
	}
	switch c := r.get(); c {
	case eof:
		return Oeof, nil, nil
	case '(':
		return Olpar, nil, nil
	case ')':
		return Orpar, nil, nil
	case '[':
		return Obra, nil, nil
	case ']':
		return Oket, nil, nil
	case '@':
		return Ocurrent, nil, nil
	case '*':
		return Ostar, nil, nil // wild, or mul
	case '.':
		return isNext(r, '.', Onest, Odot)
	case '$':
		return Oroot, nil, nil
	case ':':
		return Ocolon, nil, nil
	case '~':
		return Omatch, nil, nil
	case ',':
		return Ocomma, nil, nil
	case '/':
		return Oslash, nil, nil // /re/ or div
	case '%':
		return Omod, nil, nil
	case '+':
		return Oadd, nil, nil
	case '-':
		return Osub, nil, nil // or should - be allowed in Oid as in 2007?
	case '&':
		return isNext(r, '&', Oand, Oerror)
	case '|':
		return isNext(r, '|', Oor, Oerror)
	case '<':
		return isNext(r, '=', Ole, Olt)
	case '>':
		return isNext(r, '=', Oge, Ogt)
	case '=':
		return isNext(r, '=', Oeq, Oerror)
	case '!':
		return isNext(r, '=', One, Oerror)
	case '?':
		return isNext(r, '(', Ofilter, Oerror)
	case '"', '\'':
		s, err := lexString(r, c)
		return Ostring, s, err
	default:
		var sb strings.Builder
		if isDigit(c) {
			r.unget()
			for isDigit(r.look()) {
				sb.WriteByte(byte(r.get()))
			}
			v, err := strconv.ParseInt(sb.String(), 10, 64)
			if err != nil {
				return Oerror, nil, err
			}
			return Oint, v, nil
		}
		if isLetter(c) {
			r.unget()
			for isAlphanumeric(r.look()) {
				sb.WriteByte(byte(r.get()))
			}
			return Oid, sb.String(), nil
		}
		return Oerror, nil, nil
	}
}

// consume a string until the closing quote cq., interpreting escape sequences
func lexString(r *rd, cq int) (string, error) {
	var s strings.Builder
Building:
	for {
		c := r.get()
		if c == cq {
			break
		}
		if c == '\\' {
			switch c = r.get(); c {
			case eof:
				return "", ErrUnclosedString
			case '\\', '/':
				// c
			case 't':
				c = '\t'
			case 'f':
				c = 0xC
			case 'n':
				c = '\n'
			case 'r':
				c = '\r'
			case 'u':
				// unicode (TO DO) 4 hex digits
				var digits strings.Builder
				for i := 0; i < 4; i++ {
					c = r.get()
					if !isHexDigit(c) {
						if c == eof {
							return "", ErrUnclosedString
						}
						r.unget()
						// illegal unicode, what to do?
						s.WriteByte('\\')
						s.WriteString(digits.String())
						continue Building
					} else {
						digits.WriteByte(byte(c))
					}
				}
				v, _ := strconv.ParseInt(digits.String(), 16, 32)
				s.WriteRune(rune(v))
				continue Building
			default:
				if c != cq {
					// illegal escape. what to do?
					s.WriteByte('\\')
				}
			}
		}
		s.WriteByte(byte(c))
	}
	return s.String(), nil
}

// if the next character is c, consume it and return t, otherwise return f
func isNext(r *rd, c int, t Op, f Op) (Op, interface{}, error) {
	if r.look() == c {
		r.get()
		return t, nil, nil
	}
	if f == Oerror {
		return f, nil, fmt.Errorf("unexpected char %#v after %#v", rune(r.look()), rune(c))
	}
	return f, nil, nil
}

func isDigit(c int) bool {
	return c >= '0' && c <= '9'
}

func isHexDigit(c int) bool {
	return isDigit(c) || c >= 'a' && c <= 'f' || c >= 'A' && c <= 'F'
}

func isLetter(c int) bool {
	return c >= 'a' && c <= 'z' || c >= 'A' && c <= 'Z' || c == '_'
}

func isAlphanumeric(c int) bool {
	return isLetter(c) || isDigit(c)
}

func isSpace(c int) bool {
	return c == ' ' || c == '\t' || c == '\n' || c == '\r' || c == '\f'
}
