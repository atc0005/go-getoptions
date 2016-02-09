// This file is part of go-getoptions.
//
// Copyright (C) 2015  David Gamba Rios
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.
package getoptions

import (
	"io/ioutil"
	"log"
	"os"
	"reflect"
	"testing"
)

func TestIsOption(t *testing.T) {
	Debug.SetOutput(os.Stderr)
	Debug.SetOutput(ioutil.Discard)

	cases := []struct {
		in       string
		mode     string
		options  []string
		argument string
	}{
		{"opt", "bundling", []string{}, ""},
		{"--opt", "bundling", []string{"opt"}, ""},
		{"--opt=arg", "bundling", []string{"opt"}, "arg"},
		{"-opt", "bundling", []string{"o", "p", "t"}, ""},
		{"-opt=arg", "bundling", []string{"o", "p", "t"}, "arg"},
		{"-", "bundling", []string{"-"}, ""},
		{"--", "bundling", []string{"--"}, ""},

		{"opt", "singleDash", []string{}, ""},
		{"--opt", "singleDash", []string{"opt"}, ""},
		{"--opt=arg", "singleDash", []string{"opt"}, "arg"},
		{"-opt", "singleDash", []string{"o"}, "pt"},
		{"-", "singleDash", []string{"-"}, ""},
		{"--", "singleDash", []string{"--"}, ""},

		{"opt", "normal", []string{}, ""},
		{"--opt", "normal", []string{"opt"}, ""},
		{"--opt=arg", "normal", []string{"opt"}, "arg"},
		{"-opt", "normal", []string{"opt"}, ""},
		{"-", "normal", []string{"-"}, ""},
		{"--", "normal", []string{"--"}, ""},
	}
	for _, c := range cases {
		options, argument := isOption(c.in, c.mode)
		if !reflect.DeepEqual(options, c.options) || argument != c.argument {
			t.Errorf("isOption(%q, %q) == (%q, %q), want (%q, %q)",
				c.in, c.mode, options, argument, c.options, c.argument)
		}
	}
}

/*
// Verifies that a panic is reached when the same option is defined twice.
func TestDuplicateDefinition(t *testing.T) {
	opt := GetOptions()
	opt.Bool("flag", false, "f")
	opt.Bool("flag", false, "f")
}
*/

func TestWarningOrErrorOnUnknown(t *testing.T) {
	opt := GetOptions()
	_, err := opt.Parse([]string{"--flags"})
	if err == nil {
		t.Errorf("Unknown option 'flags' didn't raise error")
	}
	if err != nil && err.Error() != "Unknown option 'flags'" {
		t.Errorf("Error string didn't match expected value")
	}
}

func TestOptionals(t *testing.T) {
	// Missing argument without default
	opt := GetOptions()
	opt.String("string", "")
	_, err := opt.Parse([]string{"--string"})
	if err == nil {
		t.Errorf("Missing argument for option 'string' didn't raise error")
	}
	if err != nil && err.Error() != "Missing argument for option 'string'!" {
		t.Errorf("Error string didn't match expected value")
	}

	// Missing argument with default
	opt = GetOptions()
	opt.StringOptional("string", "default")
	_, err = opt.Parse([]string{"--string"})
	if err != nil {
		t.Errorf("Unexpected error: %s", err)
	}
	if opt.Option["string"] != "default" {
		t.Errorf("Default value not set for 'string'")
	}

	// Argument given
	opt = GetOptions()
	opt.StringOptional("string", "default")
	_, err = opt.Parse([]string{"--string=arg"})
	if err != nil {
		t.Errorf("Unexpected error: %s", err)
	}
	if opt.Option["string"] != "arg" {
		t.Errorf("string Optional didn't take argument")
	}
	opt = GetOptions()
	opt.StringOptional("string", "default")
	_, err = opt.Parse([]string{"--string", "arg"})
	if err != nil {
		t.Errorf("Unexpected error: %s", err)
	}
	if opt.Option["string"] != "arg" {
		t.Errorf("string Optional didn't take argument")
	}

	// VarOptional
	var result string
	opt = GetOptions()
	opt.StringVarOptional(&result, "string", "default")
	_, err = opt.Parse([]string{"--string=arg"})
	if err != nil {
		t.Errorf("Unexpected error: %s", err)
	}
	if result != "arg" {
		t.Errorf("StringVarOptional didn't take argument")
	}

	result = ""
	opt = GetOptions()
	opt.StringVarOptional(&result, "string", "default")
	_, err = opt.Parse([]string{"--string=arg"})
	if err != nil {
		t.Errorf("Unexpected error: %s", err)
	}
	if result != "arg" {
		t.Errorf("StringVarOptional didn't take argument")
	}

	result = ""
	opt = GetOptions()
	opt.StringVarOptional(&result, "string", "default")
	_, err = opt.Parse([]string{"--string"})
	if err != nil {
		t.Errorf("Unexpected error: %s", err)
	}
	if result != "default" {
		t.Errorf("Default value not set for 'string'")
	}
}

func TestGetOptBool(t *testing.T) {
	setup := func() *GetOpt {
		opt := GetOptions()
		opt.Bool("flag", false)
		opt.NBool("nflag", false)
		return opt
	}

	cases := []struct {
		opt    *GetOpt
		option string
		input  []string
		value  bool
	}{
		{setup(),
			"flag",
			[]string{"--flag"},
			true,
		},
		{setup(),
			"nflag",
			[]string{"--nflag"},
			true,
		},
		{setup(),
			"nflag",
			[]string{"--no-nflag"},
			false,
		},
	}
	for _, c := range cases {
		_, err := c.opt.Parse(c.input)
		if err != nil {
			t.Errorf("Unexpected error: %s", err)
		}
		if c.opt.Option[c.option] != c.value {
			t.Errorf("Wrong value: %v != %v", c.opt.Option[c.option], c.value)
		}
	}
}

func TestCalled(t *testing.T) {
	opt := GetOptions()
	opt.Bool("hello", false)
	opt.Bool("happy", false)
	opt.Bool("world", false)
	opt.String("string", "")
	opt.String("string2", "")
	opt.Int("int", 0)
	opt.Int("int2", 0)
	_, err := opt.Parse([]string{"--hello", "--world", "--string2", "str", "--int2", "123"})
	if err != nil {
		t.Errorf("Unexpected error: %s", err)
	}
	if !opt.Called["hello"] {
		t.Errorf("hello didn't have expected value %v", false)
	}
	if opt.Called["happy"] {
		t.Errorf("happy didn't have expected value %v", true)
	}
	if !opt.Called["world"] {
		t.Errorf("world didn't have expected value %v", false)
	}
	if opt.Called["string"] {
		t.Errorf("string didn't have expected value %v", true)
	}
	if !opt.Called["string2"] {
		t.Errorf("string2 didn't have expected value %v", false)
	}
	if opt.Called["int"] {
		t.Errorf("int didn't have expected value %v", true)
	}
	if !opt.Called["int2"] {
		t.Errorf("int2 didn't have expected value %v", false)
	}
}

func TestEndOfParsing(t *testing.T) {
	opt := GetOptions()
	opt.Bool("hello", false)
	opt.Bool("world", false)
	remaining, err := opt.Parse([]string{"hola", "--hello", "--", "mundo", "--world"})
	if err != nil {
		t.Errorf("Unexpected error: %s", err)
	}

	if !reflect.DeepEqual(remaining, []string{"hola", "mundo", "--world"}) {
		t.Errorf("remaining didn't have expected value: %v != %v", remaining, []string{"hola", "mundo", "--world"})
	}
}

func TestGetOptAliases(t *testing.T) {
	setup := func() *GetOpt {
		opt := GetOptions()
		opt.Bool("flag", false, "f", "h")
		return opt
	}

	cases := []struct {
		opt    *GetOpt
		option string
		input  []string
		value  bool
	}{
		{setup(),
			"flag",
			[]string{"--flag"},
			true,
		},
		{setup(),
			"flag",
			[]string{"-f"},
			true,
		},
		{setup(),
			"flag",
			[]string{"-h"},
			true,
		},
		// TODO: Add flag to allow for this.
		{setup(),
			"flag",
			[]string{"--fl"},
			true,
		},
	}
	for _, c := range cases {
		_, err := c.opt.Parse(c.input)
		if err != nil {
			t.Errorf("Unexpected error: %s", err)
		}
		if c.opt.Option[c.option] != c.value {
			t.Errorf("Wrong value: %v != %v", c.opt.Option[c.option], c.value)
		}
	}

	opt := GetOptions()
	opt.Bool("flag", false)
	opt.Bool("fleg", false)
	_, err := opt.Parse([]string{"--fl"})
	if err == nil {
		t.Errorf("Ambiguous argument 'fl' didn't raise unknown option error")
	}
	if err != nil && err.Error() != "Unknown option 'fl'" {
		t.Errorf("Error string didn't match expected value")
	}
}

func TestGetOptString(t *testing.T) {
	setup := func() *GetOpt {
		opt := GetOptions()
		opt.String("string", "")
		return opt
	}

	cases := []struct {
		opt    *GetOpt
		option string
		input  []string
		value  string
	}{
		{setup(),
			"string",
			[]string{"--string=hello"},
			"hello",
		},
		{setup(),
			"string",
			[]string{"--string=hello", "world"},
			"hello",
		},
		{setup(),
			"string",
			[]string{"--string", "hello"},
			"hello",
		},
		{setup(),
			"string",
			[]string{"--string", "hello", "world"},
			"hello",
		},
		// TODO: Set a flag to decide wheter or not to allow this
		{setup(),
			"string",
			[]string{"--string", "--hello", "world"},
			"--hello",
		},
		// TODO: Set up a flag to decide wheter or not to err on this
		{setup(),
			"string",
			[]string{"--string", "hello", "--string", "world"},
			"world",
		},
	}
	for _, c := range cases {
		_, err := c.opt.Parse(c.input)
		if err != nil {
			t.Errorf("Unexpected error: %s", err)
		}
		if c.opt.Option[c.option] != c.value {
			t.Errorf("Wrong value: %v != %v", c.opt.Option[c.option], c.value)
		}
	}
}

func TestGetOptInt(t *testing.T) {
	setup := func() *GetOpt {
		opt := GetOptions()
		opt.Int("int", 0)
		return opt
	}

	cases := []struct {
		opt    *GetOpt
		option string
		input  []string
		value  int
	}{
		{setup(),
			"int",
			[]string{"--int=123"},
			123,
		},
		{setup(),
			"int",
			[]string{"--int=123", "world"},
			123,
		},
		{setup(),
			"int",
			[]string{"--int", "123"},
			123,
		},
		{setup(),
			"int",
			[]string{"--int", "123", "world"},
			123,
		},
	}
	for _, c := range cases {
		_, err := c.opt.Parse(c.input)
		if err != nil {
			t.Errorf("Unexpected error: %s", err)
		}
		if c.opt.Option[c.option] != c.value {
			t.Errorf("Wrong value: %v != %v", c.opt.Option[c.option], c.value)
		}
	}

	// Cast errors
	opt := GetOptions()
	opt.Int("int", 0)
	_, err := opt.Parse([]string{"--int=hello"})
	if err == nil {
		t.Errorf("Int cast didn't raise errors")
	}
	if err != nil && err.Error() != "Can't convert string to int: 'hello'" {
		t.Errorf("Error string didn't match expected value '%s'", err)
	}
}

func TestGetOptStringRepeat(t *testing.T) {
	setup := func() *GetOpt {
		opt := GetOptions()
		opt.StringSlice("string")
		return opt
	}

	cases := []struct {
		opt    *GetOpt
		option string
		input  []string
		value  []string
	}{
		{setup(),
			"string",
			[]string{"--string=hello"},
			[]string{"hello"},
		},
		{setup(),
			"string",
			[]string{"--string=hello", "world"},
			[]string{"hello"},
		},
		{setup(),
			"string",
			[]string{"--string", "hello"},
			[]string{"hello"},
		},
		{setup(),
			"string",
			[]string{"--string", "hello", "world"},
			[]string{"hello"},
		},
		// TODO: Set a flag to decide wheter or not to allow this
		{setup(),
			"string",
			[]string{"--string", "--hello", "world"},
			[]string{"--hello"},
		},
		{setup(),
			"string",
			[]string{"--string", "hello", "--string", "happy", "--string", "world"},
			[]string{"hello", "happy", "world"},
		},
	}
	for _, c := range cases {
		_, err := c.opt.Parse(c.input)
		if err != nil {
			t.Errorf("Unexpected error: %s", err)
		}
		if !reflect.DeepEqual(c.opt.Option[c.option], c.value) {
			t.Errorf("Wrong value: %v != %v", c.opt.Option[c.option], c.value)
		}
	}
}

// TODO: Allow passig : as the map divider
func TestGetOptStringMap(t *testing.T) {
	setup := func() *GetOpt {
		opt := GetOptions()
		opt.StringMap("string")
		return opt
	}

	// TODO: Check error when there is no equal sign.

	cases := []struct {
		opt    *GetOpt
		option string
		input  []string
		value  map[string]string
	}{
		{setup(),
			"string",
			[]string{"--string=hello=world"},
			map[string]string{"hello": "world"},
		},
		{setup(),
			"string",
			[]string{"--string=hello=happy", "world"},
			map[string]string{"hello": "happy"},
		},
		{setup(),
			"string",
			[]string{"--string", "hello=world"},
			map[string]string{"hello": "world"},
		},
		{setup(),
			"string",
			[]string{"--string", "hello=happy", "world"},
			map[string]string{"hello": "happy"},
		},
		// TODO: Set a flag to decide wheter or not to allow this
		{setup(),
			"string",
			[]string{"--string", "--hello=happy", "world"},
			map[string]string{"--hello": "happy"},
		},
		{setup(),
			"string",
			[]string{"--string", "hello=world", "--string", "key=value", "--string", "key2=value2"},
			map[string]string{"hello": "world", "key": "value", "key2": "value2"},
		},
	}
	for _, c := range cases {
		_, err := c.opt.Parse(c.input)
		if err != nil {
			t.Errorf("Unexpected error: %s", err)
		}
		if !reflect.DeepEqual(c.opt.Option[c.option], c.value) {
			t.Errorf("Wrong value: %v != %v", c.opt.Option[c.option], c.value)
		}
	}
}

func TestVars(t *testing.T) {
	opt := GetOptions()

	var flag, flag2, flag5, flag6 bool
	opt.BoolVar(&flag, "flag", false)
	opt.BoolVar(&flag2, "flag2", true)
	flag3 := opt.Bool("flag3", false)
	flag4 := opt.Bool("flag4", true)
	opt.BoolVar(&flag5, "flag5", false)
	opt.BoolVar(&flag6, "flag6", true)

	var nflag, nflag2 bool
	opt.NBoolVar(&nflag, "nflag", false)
	opt.NBoolVar(&nflag2, "n2", false)

	var str, str2 string
	opt.StringVar(&str, "stringVar", "")
	opt.StringVar(&str2, "stringVar2", "")

	var integer int
	opt.IntVar(&integer, "intVar", 0)

	_, err := opt.Parse([]string{
		"-flag",
		"-flag2",
		"-flag3",
		"-flag4",
		"-nf",
		"--no-n2",
		"--stringVar", "hello",
		"--stringVar2=world",
		"--intVar", "123",
	})
	if err != nil {
		t.Errorf("Unexpected error: %s", err)
	}

	if flag != true {
		t.Errorf("flag didn't have expected value: %v != %v", flag, true)
	}
	if flag2 != false {
		t.Errorf("flag2 didn't have expected value: %v != %v", flag2, false)
	}
	if *flag3 != true {
		t.Errorf("flag3 didn't have expected value: %v != %v", *flag3, true)
	}
	if *flag4 != false {
		t.Errorf("flag4 didn't have expected value: %v != %v", *flag4, false)
	}
	if flag5 != false {
		t.Errorf("flag5 didn't have expected value: %v != %v", flag5, false)
	}
	if flag6 != true {
		t.Errorf("flag6 didn't have expected value: %v != %v", flag6, true)
	}

	if nflag != true {
		t.Errorf("nflag didn't have expected value: %v != %v", nflag, true)
	}
	if nflag2 != false {
		t.Errorf("nflag2 didn't have expected value: %v != %v", nflag2, false)
	}
	if str != "hello" {
		t.Errorf("str didn't have expected value: %v != %v", str, "hello")
	}
	if str2 != "world" {
		t.Errorf("str2 didn't have expected value: %v != %v", str, "world")
	}
	if integer != 123 {
		t.Errorf("integer didn't have expected value: %v != %v", integer, 123)
	}
}

func TestDefaultValues(t *testing.T) {
	var flag, nflag bool
	var str string
	var integer, integer2 int

	opt := GetOptions()
	opt.Bool("flag", false)
	opt.BoolVar(&flag, "varflag", false)
	opt.NBool("nflag", false)
	opt.NBoolVar(&nflag, "varnflag", false)
	opt.String("string", "")
	opt.String("string2", "default")
	str3 := opt.String("string3", "default")
	opt.StringVar(&str, "stringVar", "")
	opt.StringVar(&str, "stringVar2", "default")
	opt.Int("int", 0)
	int2 := opt.Int("int2", 5)
	opt.IntVar(&integer, "intVar", 0)
	opt.IntVar(&integer2, "intVar2", 5)
	opt.StringSlice("string-repeat")
	opt.StringMap("string-map")

	_, err := opt.Parse([]string{})

	if err != nil {
		log.Println(err)
	}

	expected := map[string]interface{}{
		"flag":          false,
		"varflag":       false,
		"nflag":         false,
		"varnflag":      false,
		"string":        "",
		"string2":       "default",
		"stringVar":     "",
		"stringVar2":    "default",
		"int":           0,
		"intVar":        0,
		"string-repeat": []string{},
		"string-map":    map[string]string{},
	}

	for k := range expected {
		if !reflect.DeepEqual(opt.Option[k], expected[k]) {
			t.Errorf("Wrong value: %s\n%v !=\n%v", k, opt.Option[k], expected[k])
		}
	}

	if flag != false {
		t.Errorf("flag didn't have expected value: %v != %v", flag, true)
	}
	if nflag != false {
		t.Errorf("nflag didn't have expected value: %v != %v", nflag, true)
	}
	if str != "" {
		t.Errorf("str didn't have expected value: %v != %v", str, "hello")
	}
	if *str3 != "default" {
		t.Errorf("str didn't have expected value: %v != %v", str3, "default")
	}
	if integer != 0 {
		t.Errorf("integer didn't have expected value: %v != %v", integer, 123)
	}
	if integer2 != 5 {
		t.Errorf("integer2 didn't have expected value: %v != %v", integer2, 5)
	}
	if *int2 != 5 {
		t.Errorf("int2 didn't have expected value: %v != %v", int2, 5)
	}

	// Tested above, but it gives me a feel for how it would be used

	if opt.Option["flag"].(bool) {
		t.Errorf("flag didn't have expected value: %v != %v", opt.Option["flag"], false)
	}
	if opt.Option["non-used-flag"] != nil && opt.Option["non-used-flag"].(bool) {
		t.Errorf("non-used-flag didn't have expected value: %v != %v", opt.Option["non-used-flag"], nil)
	}
	if opt.Option["flag"] != nil && opt.Option["nflag"].(bool) {
		t.Errorf("nflag didn't have expected value: %v != %v", opt.Option["nflag"], nil)
	}
	if opt.Option["string"] != "" {
		t.Errorf("str didn't have expected value: %v != %v", opt.Option["string"], "")
	}
	if opt.Option["int"] != 0 {
		t.Errorf("int didn't have expected value: %v != %v", opt.Option["int"], 0)
	}
}

func TestAll(t *testing.T) {
	var flag, nflag, nflag2 bool
	var str string
	var integer int
	opt := GetOptions()
	opt.Bool("flag", false)
	opt.BoolVar(&flag, "varflag", false)
	opt.Bool("non-used-flag", false)
	opt.NBool("nflag", false)
	opt.NBool("nftrue", false)
	opt.NBool("nfnil", false)
	opt.NBoolVar(&nflag, "varnflag", false)
	opt.NBoolVar(&nflag2, "varnflag2", false)
	opt.String("string", "")
	opt.StringVar(&str, "stringVar", "")
	opt.Int("int", 0)
	opt.IntVar(&integer, "intVar", 0)
	opt.StringSlice("string-repeat")
	opt.StringMap("string-map")

	// log.Println(opt)

	remaining, err := opt.Parse([]string{
		"hello",
		"--flag",
		"--varflag",
		"--no-nflag",
		"--nft",
		"happy",
		"--varnflag",
		"--no-varnflag2",
		"--string", "hello",
		"--stringVar", "hello",
		"--int", "123",
		"--intVar", "123",
		"--string-repeat", "hello", "--string-repeat", "world",
		"--string-map", "hello=world", "--string-map", "server=name",
		"world",
	})

	if err != nil {
		log.Println(err)
	}

	if !reflect.DeepEqual(remaining, []string{"hello", "happy", "world"}) {
		t.Errorf("remaining didn't have expected value: %v != %v", remaining, []string{"hello", "happy", "world"})
	}

	expected := map[string]interface{}{
		"flag":          true,
		"nflag":         false,
		"nftrue":        true,
		"string":        "hello",
		"int":           123,
		"string-repeat": []string{"hello", "world"},
		"string-map":    map[string]string{"hello": "world", "server": "name"},
	}

	for k := range expected {
		if !reflect.DeepEqual(opt.Option[k], expected[k]) {
			t.Errorf("Wrong value: %v != %v", opt.Option, expected)
		}
	}

	if flag != true {
		t.Errorf("flag didn't have expected value: %v != %v", flag, true)
	}
	if nflag != true {
		t.Errorf("nflag didn't have expected value: %v != %v", nflag, true)
	}
	if nflag2 != false {
		t.Errorf("nflag2 didn't have expected value: %v != %v", nflag2, false)
	}
	if str != "hello" {
		t.Errorf("str didn't have expected value: %v != %v", str, "hello")
	}
	if integer != 123 {
		t.Errorf("int didn't have expected value: %v != %v", integer, 123)
	}

	// Tested above, but it gives me a feel for how it would be used

	if opt.Option["flag"] != nil && !opt.Option["flag"].(bool) {
		t.Errorf("flag didn't have expected value: %v != %v", opt.Option["flag"], true)
	}
	if opt.Option["non-used-flag"] != nil && opt.Option["non-used-flag"].(bool) {
		t.Errorf("non-used-flag didn't have expected value: %v != %v", opt.Option["non-used-flag"], false)
	}
	if opt.Option["flag"] != nil && opt.Option["nflag"].(bool) {
		t.Errorf("nflag didn't have expected value: %v != %v", opt.Option["nflag"], true)
	}
	if opt.Option["string"] != "hello" {
		t.Errorf("str didn't have expected value: %v != %v", opt.Option["string"], "hello")
	}
	if opt.Option["int"] != 123 {
		t.Errorf("int didn't have expected value: %v != %v", opt.Option["int"], 123)
	}
}
