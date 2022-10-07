// Copyright (c) 2021 rookie-ninja
//
// Use of this source code is governed by an Apache-style
// license that can be found in the LICENSE file.
//
// Copied from https://github.com/helm/helm/tree/master/pkg/strvals

package rku

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"
)

// errNotList indicates that a non-list was treated as a list.
var errNotList = errors.New("not a list")

// parser is a simple parser that takes a strvals line and parses it into a
// map representation.
//
// where sc is the source of the original data being parsed
// where data is the final parsed data from the parses with correct types
// where st is a boolean to figure out if we're forcing it to parse values as string
type parser struct {
	sc         *bytes.Buffer
	data       map[interface{}]interface{}
	runesToVal runesToVal
}

type runesToVal func([]rune) (interface{}, error)

func newParser(sc *bytes.Buffer, data map[interface{}]interface{}, stringBool bool) *parser {
	rs2v := func(rs []rune) (interface{}, error) {
		return typedVal(rs, stringBool), nil
	}
	return &parser{sc: sc, data: data, runesToVal: rs2v}
}

func (t *parser) parse() error {
	for {
		err := t.key(t.data)
		if err == nil {
			continue
		}
		if err == io.EOF {
			return nil
		}
		return err
	}
}

func runeSet(r []rune) map[rune]bool {
	s := make(map[rune]bool, len(r))
	for _, rr := range r {
		s[rr] = true
	}
	return s
}

func (t *parser) key(data map[interface{}]interface{}) error {
	stop := runeSet([]rune{'=', '[', ',', '.'})
	for {
		switch k, last, err := runesUntil(t.sc, stop); {
		case err != nil:
			if len(k) == 0 {
				return err
			}
			return fmt.Errorf("key %q has no value", string(k))
			//set(data, string(k), "")
			//return err
		case last == '[':
			// We are in a list index context, so we need to set an index.
			i, err := t.keyIndex()
			if err != nil {
				return fmt.Errorf("error parsing index: %s", err)
			}
			kk := string(k)
			// Find or create target list
			list := []interface{}{}
			if _, ok := data[kk]; ok {
				if v, ok := data[kk].([]interface{}); ok {
					list = v
				} else {
					return fmt.Errorf("invalid format")
				}
			}

			// Now we need to get the value after the ].
			list, err = t.listItem(list, i)
			set(data, kk, list)
			return err
		case last == '=':
			//End of key. Consume =, Get value.
			// FIXME: Get value list first
			vl, e := t.valList()
			switch e {
			case nil:
				set(data, string(k), vl)
				return nil
			case io.EOF:
				set(data, string(k), "")
				return e
			case errNotList:
				rs, e := t.val()
				if e != nil && e != io.EOF {
					return e
				}
				v, e := t.runesToVal(rs)
				set(data, string(k), v)
				return e
			default:
				return e
			}

		case last == ',':
			// No value given. Set the value to empty string. Return error.
			set(data, string(k), "")
			return fmt.Errorf("key %q has no value (cannot end with ,)", string(k))
		case last == '.':
			// First, create or find the target map.
			inner := map[interface{}]interface{}{}
			if _, ok := data[string(k)]; ok {
				if v, ok := data[string(k)].(map[interface{}]interface{}); ok {
					inner = v
				} else {
					return fmt.Errorf("invalid format")
				}
			}

			// Recurse
			e := t.key(inner)
			if len(inner) == 0 {
				return fmt.Errorf("key map %q has no value", string(k))
			}
			set(data, string(k), inner)
			return e
		}
	}
}

func set(data map[interface{}]interface{}, key string, val interface{}) {
	// If key is empty, don't set it.
	if len(key) == 0 {
		return
	}
	data[key] = val
}

func setIndex(list []interface{}, index int, val interface{}) []interface{} {
	if len(list) <= index {
		newlist := make([]interface{}, index+1)
		copy(newlist, list)
		list = newlist
	}
	list[index] = val
	return list
}

func (t *parser) keyIndex() (int, error) {
	// First, get the key.
	stop := runeSet([]rune{']'})
	v, _, err := runesUntil(t.sc, stop)
	if err != nil {
		return 0, err
	}
	// v should be the index
	return strconv.Atoi(string(v))

}

func (t *parser) listItem(list []interface{}, i int) ([]interface{}, error) {
	stop := runeSet([]rune{'[', '.', '='})
	switch k, last, err := runesUntil(t.sc, stop); {
	case len(k) > 0:
		return list, fmt.Errorf("unexpected data at end of array index: %q", k)
	case err != nil:
		return list, err
	case last == '=':
		vl, e := t.valList()
		switch e {
		case nil:
			return setIndex(list, i, vl), nil
		case io.EOF:
			return setIndex(list, i, ""), err
		case errNotList:
			rs, e := t.val()
			if e != nil && e != io.EOF {
				return list, e
			}
			v, e := t.runesToVal(rs)
			return setIndex(list, i, v), e
		default:
			return list, e
		}
	case last == '[':
		// now we have a nested list. Read the index and handle.
		nextI, err := t.keyIndex()
		if err != nil {
			return list, fmt.Errorf("error parsing index: %s", err)
		}
		var crtList []interface{}
		if len(list) > i {
			// If nested list already exists, take the value of list to next cycle.
			existed := list[i]
			if existed != nil {
				if v, ok := list[i].([]interface{}); ok {
					crtList = v
				} else {
					return list, fmt.Errorf("invalid source")
				}
			}
		}
		// Now we need to get the value after the ].
		list2, err := t.listItem(crtList, nextI)
		return setIndex(list, i, list2), err
	case last == '.':
		// We have a nested object. Send to t.key
		inner := map[interface{}]interface{}{}
		if len(list) > i {
			var ok bool
			inner, ok = list[i].(map[interface{}]interface{})
			if !ok {
				// We have indices out of order. Initialize empty value.
				list[i] = map[interface{}]interface{}{}
				inner = list[i].(map[interface{}]interface{})
			}
		}

		// Recurse
		e := t.key(inner)
		return setIndex(list, i, inner), e
	default:
		return nil, fmt.Errorf("parse error: unexpected token %v", last)
	}
}

func (t *parser) val() ([]rune, error) {
	stop := runeSet([]rune{','})
	v, _, err := runesUntil(t.sc, stop)
	return v, err
}

func (t *parser) valList() ([]interface{}, error) {
	r, _, e := t.sc.ReadRune()
	if e != nil {
		return []interface{}{}, e
	}

	if r != '{' {
		t.sc.UnreadRune()
		return []interface{}{}, errNotList
	}

	list := []interface{}{}
	stop := runeSet([]rune{',', '}'})
	for {
		switch rs, last, err := runesUntil(t.sc, stop); {
		case err != nil:
			if err == io.EOF {
				err = errors.New("list must terminate with '}'")
			}
			return list, err
		case last == '}':
			// If this is followed by ',', consume it.
			if r, _, e := t.sc.ReadRune(); e == nil && r != ',' {
				t.sc.UnreadRune()
			}
			v, e := t.runesToVal(rs)
			list = append(list, v)
			return list, e
		case last == ',':
			v, e := t.runesToVal(rs)
			if e != nil {
				return list, e
			}
			list = append(list, v)
		}
	}
}

func runesUntil(in io.RuneReader, stop map[rune]bool) ([]rune, rune, error) {
	v := []rune{}
	for {
		switch r, _, e := in.ReadRune(); {
		case e != nil:
			return v, r, e
		case inMap(r, stop):
			return v, r, nil
		case r == '\\':
			next, _, e := in.ReadRune()
			if e != nil {
				return v, next, e
			}
			v = append(v, next)
		default:
			v = append(v, r)
		}
	}
}

func inMap(k rune, m map[rune]bool) bool {
	_, ok := m[k]
	return ok
}

func typedVal(v []rune, st bool) interface{} {
	val := string(v)

	if st {
		return val
	}

	if strings.EqualFold(val, "true") {
		return true
	}

	if strings.EqualFold(val, "false") {
		return false
	}

	if strings.EqualFold(val, "null") {
		return nil
	}

	if strings.EqualFold(val, "0") {
		return 0
	}

	// If this value does not start with zero, try parsing it to an int
	if len(val) != 0 && val[0] != '0' {
		if iv, err := strconv.Atoi(val); err == nil {
			return iv
		}
	}

	return val
}
