package dapper_test

import (
	"errors"
	"fmt"
	"io"
	"reflect"
	"strings"
	"sync"
	"testing"

	. "github.com/dogmatiq/dapper"
)

type syncmaps struct {
	Map sync.Map
}

// This test verifies that that sync.Map key/value types are always rendered.
func TestPrinter_SyncMap(t *testing.T) {
	var m sync.Map

	test(t, "zero-value sync.Map", &m, "*sync.Map{}")

	m.Store(1, 100)
	m.Store(2, 200)

	test(
		t,
		"sync.Map",
		&m,
		"*sync.Map{",
		"    int(1): int(100)",
		"    int(2): int(200)",
		"}",
	)

	m.Delete(1)
	m.Delete(2)

	test(t, "empty sync.Map", &m, "*sync.Map{}")
}

// This test verifies the formatting of sync.Map key/values in the named structs.
func TestPrinter_SyncMapInNamedStruct(t *testing.T) {
	test(
		t,
		"empty sync.Map",
		&syncmaps{},
		"*dapper_test.syncmaps{",
		"    Map: {}",
		"}",
	)

	sm := &syncmaps{}

	sm.Map.Store(1, 100)
	sm.Map.Store(2, 200)

	test(
		t,
		"non-empty sync.Map",
		sm,
		"*dapper_test.syncmaps{",
		"    Map: {",
		"        int(1): int(100)",
		"        int(2): int(200)",
		"    }",
		"}",
	)
}

// This test verifies that sync.Map keys are sorted by their formatted string
// representation.
func TestPrinter_SyncMapKeySorting(t *testing.T) {
	var m sync.Map

	m.Store("foo", 1)
	m.Store("bar", 2)

	test(
		t,
		"keys are sorted by their string representation",
		&m,
		"*sync.Map{",
		`    "bar": int(2)`,
		`    "foo": int(1)`,
		"}",
	)
}

// This test verifies that values associated with sync.Map keys that have a
// multiline string representation are aligned correctly.
func TestPrinter_MultilineSyncMapKeyAlignment(t *testing.T) {
	var m sync.Map

	m.Store("short", "one")
	m.Store("the longest key in the galaxy", "two")
	m.Store(multiline{Key: "multiline key"}, "three")

	test(
		t,
		"keys are aligned correctly",
		&m,
		"*sync.Map{",
		`    "short":                         "one"`,
		`    "the longest key in the galaxy": "two"`,
		"    dapper_test.multiline{",
		`        Key: "multiline key"`,
		`    }: "three"`,
		"}",
	)

	m.Delete("the longest key in the galaxy")

	test(
		t,
		"keys are aligned correctly when the longest line is part of a multiline key",
		&m,
		"*sync.Map{",
		`    "short":                 "one"`,
		"    dapper_test.multiline{",
		`        Key: "multiline key"`,
		`    }: "three"`,
		"}",
	)
}

// This test verifies that recursive sync.Map is detected, and do not produce
// an infinite loop or stack overflow.
func TestPrinter_SyncMapRecursion(t *testing.T) {
	var m sync.Map
	m.Store("child", &m)

	test(
		t,
		"recursive sync.Map",
		&m,
		"*sync.Map{",
		`    "child": *sync.Map(<recursion>)`,
		"}",
	)
}

// This test verifies that recursive sync.Map is detected, and do not produce
// an infinite loop or stack overflow.
func TestPrinter_SyncMapFormatFunctionErr(t *testing.T) {
	m := &sync.Map{}
	m.Store("foo", 1)

	v := Value{
		Value:                  reflect.ValueOf(m).Elem(),
		DynamicType:            reflect.TypeOf(m).Elem(),
		StaticType:             reflect.TypeOf(m).Elem(),
		IsAmbiguousDynamicType: false,
		IsAmbiguousStaticType:  false,
		IsUnexported:           false,
	}

	terr := errors.New("test key format function error")
	err := SyncFilter(
		&strings.Builder{},
		v,
		func(_ io.Writer, v Value) error {
			if s, ok := v.Value.Interface().(string); ok && s == "foo" {
				return terr
			}
			return nil
		},
	)

	t.Log(fmt.Sprintf("expected:\n\n%v\n", terr))

	if terr != err {
		t.Fatal(fmt.Sprintf("actual:\n\n%v\n", err))
	}

	terr = errors.New("test value format function error")
	err = SyncFilter(
		&strings.Builder{},
		v,
		func(_ io.Writer, v Value) error {
			if i, ok := v.Value.Interface().(int); ok && i == 1 {
				return terr
			}
			return nil
		},
	)
}
