package dapper

import (
	"io"
	"os"
	"reflect"
	"strings"

	"github.com/dogmatiq/iago/count"
	"github.com/dogmatiq/iago/must"
)

// DefaultIndent is the default indent string used to indent nested values.
var DefaultIndent = []byte("    ")

// DefaultRecursionMarker is the default string to display when recursion
// is detected within a Go value.
const DefaultRecursionMarker = "<recursion>"

// Config holds the configuration for a printer.
type Config struct {
	// Filters is the set of filters to apply when formatting values.
	Filters []Filter

	// Indent is the string used to indent nested values.
	// If it is empty, DefaultIndent is used.
	Indent []byte

	// RecursionMarker is a string that is displayed instead of a value's
	// representation when recursion has been detected.
	// If it is empty, DefaultRecursionMarker is used.
	RecursionMarker string
}

// Printer generates human-readable representations of Go values.
//
// The output format is intended to be as minimal as possible, without being
// ambiguous. To that end, type information is only included where it can not be
// reliably inferred from the structure of the value.
type Printer struct {
	// Config is the configuration for the printer.
	Config Config
}

// emptyInterfaceType is the reflect.Type for interface{}.
var emptyInterfaceType = reflect.TypeOf((*interface{})(nil)).Elem()

// Write writes a pretty-printed representation of v to w.
//
// It returns the number of bytes written.
func (p *Printer) Write(w io.Writer, v interface{}) (n int, err error) {
	defer must.Recover(&err)

	vis := visitor{
		config: p.Config,
	}

	if len(vis.config.Indent) == 0 {
		vis.config.Indent = DefaultIndent
	}

	if vis.config.RecursionMarker == "" {
		vis.config.RecursionMarker = DefaultRecursionMarker
	}

	rv := reflect.ValueOf(v)
	var rt reflect.Type

	if rv.Kind() != reflect.Invalid {
		rt = rv.Type()
	}

	cw := count.NewWriter(w)

	vis.mustVisit(
		cw,
		Value{
			Value:                  rv,
			DynamicType:            rt,
			StaticType:             emptyInterfaceType,
			IsAmbiguousDynamicType: true,
			IsAmbiguousStaticType:  true,
			IsUnexported:           false,
		},
	)

	n = cw.Count()
	return
}

// Format returns a pretty-printed representation of v.
func (p *Printer) Format(v interface{}) string {
	var b strings.Builder

	if _, err := p.Write(&b, v); err != nil {
		// CODE COVERAGE: At the time of writing, strings.Builder.Write() never
		// returns an error.
		panic(err)
	}

	return b.String()
}

var defaultPrinter = Printer{
	Config: Config{
		Filters: []Filter{
			ReflectTypeFilter,
			TimeFilter,
			DurationFilter,
			SyncFilter,
		},
	},
}

// Write writes a pretty-printed representation of v to w using the default
// printer settings.
//
// It returns the number of bytes written.
func Write(w io.Writer, v interface{}) (int, error) {
	return defaultPrinter.Write(w, v)
}

// Format returns a pretty-printed representation of v.
func Format(v interface{}) string {
	return defaultPrinter.Format(v)
}

var newLine = []byte{'\n'}

// Print writes a pretty-printed representation of v to os.Stdout.
func Print(values ...interface{}) {
	for _, v := range values {
		defaultPrinter.Write(os.Stdout, v)
		os.Stdout.Write(newLine)
	}
}
