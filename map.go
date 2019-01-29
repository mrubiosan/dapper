package dapper

import (
	"io"
	"reflect"
	"sort"
	"strings"

	"github.com/dogmatiq/iago"
	"github.com/dogmatiq/iago/indent"
)

// visitMap formats values with a kind of reflect.Map.
//
// TODO(jmalloc): sort numerically-keyed maps numerically
func (vis *visitor) visitMap(w io.Writer, v Value) {
	if vis.enter(w, v) {
		return
	}
	defer vis.leave(v)

	if v.IsAmbiguousType() {
		iago.MustWriteString(w, v.TypeName())
	}

	if v.Value.Len() == 0 {
		iago.MustWriteString(w, "{}")
		return
	}

	iago.MustWriteString(w, "{\n")
	vis.visitMapElements(indent.NewIndenter(w, vis.indent), v)
	iago.MustWriteByte(w, '}')
}

func (vis *visitor) visitMapElements(w io.Writer, v Value) {
	staticType := v.DynamicType.Elem()
	isInterface := staticType.Kind() == reflect.Interface
	keys, alignment := vis.formatMapKeys(v)

	for _, mk := range keys {
		mv := v.Value.MapIndex(mk.Value)

		// unwrap interface values so that elem has it's actual type/kind, and not
		// that of reflect.Interface.
		if isInterface && !mv.IsNil() {
			mv = mv.Elem()
		}

		iago.MustWriteString(w, mk.String)
		iago.MustWriteString(w, ": ")
		iago.MustWriteString(w, strings.Repeat(" ", alignment-mk.Width))
		vis.visit(
			w,
			Value{
				Value:                  mv,
				DynamicType:            mv.Type(),
				StaticType:             staticType,
				IsAmbiguousDynamicType: isInterface,
				IsAmbiguousStaticType:  false,
				IsUnexported:           v.IsUnexported,
			},
		)
		iago.MustWriteString(w, "\n")
	}
}

type mapKey struct {
	Value  reflect.Value
	String string
	Width  int
}

// formatMapKeys formats the keys in maps, and returns a slice of the keys
// sorted by their string representation.
//
// padding is the number of padding characters to add to the shortest key.
func (vis *visitor) formatMapKeys(v Value) (keys []mapKey, alignment int) {
	var w strings.Builder
	staticType := v.DynamicType.Key()
	isInterface := staticType.Kind() == reflect.Interface
	keys = make([]mapKey, v.Value.Len())
	alignToLastLine := false

	for i, mk := range v.Value.MapKeys() {

		// unwrap interface values so that elem has it's actual type/kind, and not
		// that of reflect.Interface.
		if isInterface && !mk.IsNil() {
			mk = mk.Elem()
		}

		vis.visit(
			&w,
			Value{
				Value:                  mk,
				DynamicType:            mk.Type(),
				StaticType:             staticType,
				IsAmbiguousDynamicType: isInterface,
				IsAmbiguousStaticType:  false,
				IsUnexported:           v.IsUnexported,
			},
		)

		s := w.String()
		w.Reset()

		max, last := widths(s)
		if max > alignment {
			alignment = max
			alignToLastLine = max == last
		}

		keys[i] = mapKey{mk, s, last}
	}

	sort.Slice(
		keys,
		func(i, j int) bool {
			return keys[i].String < keys[j].String
		},
	)

	// compensate for the ":" added to the last line"
	if !alignToLastLine {
		alignment--
	}

	return
}

// widths returns the number of characters in the longest, and last line of s.
func widths(s string) (max int, last int) {
	for {
		i := strings.IndexByte(s, '\n')

		if i == -1 {
			last = len(s)
			if len(s) > max {
				max = len(s)
			}
			return
		}

		if i > max {
			max = i
		}

		s = s[i+1:]
	}
}
