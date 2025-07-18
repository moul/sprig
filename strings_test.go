package sprig

import (
	"encoding/base32"
	"encoding/base64"
	"errors"
	"fmt"
	"testing"
	"unicode/utf8"

	"github.com/stretchr/testify/assert"
)

type stringerImpl struct {
	Value string
}

// String satisfies the Stringer interface.
func (s stringerImpl) String() string {
	return s.Value
}

func TestSubstr(t *testing.T) {
	tpl := `{{"fooo" | substr 0 3 }}`
	if err := runt(tpl, "foo"); err != nil {
		t.Error(err)
	}
}

func TestSubstr_shorterString(t *testing.T) {
	tpl := `{{"foo" | substr 0 10 }}`
	if err := runt(tpl, "foo"); err != nil {
		t.Error(err)
	}
}

func TestTrunc(t *testing.T) {
	tpl := `{{ "foooooo" | trunc 3 }}`
	if err := runt(tpl, "foo"); err != nil {
		t.Error(err)
	}
	tpl = `{{ "baaaaaar" | trunc -3 }}`
	if err := runt(tpl, "aar"); err != nil {
		t.Error(err)
	}
	tpl = `{{ "baaaaaar" | trunc -999 }}`
	if err := runt(tpl, "baaaaaar"); err != nil {
		t.Error(err)
	}
	tpl = `{{ "baaaaaz" | trunc 0 }}`
	if err := runt(tpl, ""); err != nil {
		t.Error(err)
	}
}

func TestQuote(t *testing.T) {
	tpl := `{{quote "a" "b" "c"}}`
	if err := runt(tpl, `"a" "b" "c"`); err != nil {
		t.Error(err)
	}
	tpl = `{{quote "\"a\"" "b" "c"}}`
	if err := runt(tpl, `"\"a\"" "b" "c"`); err != nil {
		t.Error(err)
	}
	tpl = `{{quote 1 2 3 }}`
	if err := runt(tpl, `"1" "2" "3"`); err != nil {
		t.Error(err)
	}
	tpl = `{{ .value | quote }}`
	values := map[string]interface{}{"value": nil}
	if err := runtv(tpl, ``, values); err != nil {
		t.Error(err)
	}
}
func TestSquote(t *testing.T) {
	tpl := `{{squote "a" "b" "c"}}`
	if err := runt(tpl, `'a' 'b' 'c'`); err != nil {
		t.Error(err)
	}
	tpl = `{{squote 1 2 3 }}`
	if err := runt(tpl, `'1' '2' '3'`); err != nil {
		t.Error(err)
	}
	tpl = `{{ .value | squote }}`
	values := map[string]interface{}{"value": nil}
	if err := runtv(tpl, ``, values); err != nil {
		t.Error(err)
	}
}

func TestContains(t *testing.T) {
	// Mainly, we're just verifying the paramater order swap.
	tests := []string{
		`{{if contains "cat" "fair catch"}}1{{end}}`,
		`{{if hasPrefix "cat" "catch"}}1{{end}}`,
		`{{if hasSuffix "cat" "ducat"}}1{{end}}`,
	}
	for _, tt := range tests {
		if err := runt(tt, "1"); err != nil {
			t.Error(err)
		}
	}
}

func TestTrim(t *testing.T) {
	tests := []string{
		`{{trim "   5.00   "}}`,
		`{{trimAll "$" "$5.00$"}}`,
		`{{trimPrefix "$" "$5.00"}}`,
		`{{trimSuffix "$" "5.00$"}}`,
	}
	for _, tt := range tests {
		if err := runt(tt, "5.00"); err != nil {
			t.Error(err)
		}
	}
}

func TestSplit(t *testing.T) {
	tpl := `{{$v := "foo$bar$baz" | split "$"}}{{$v._0}}`
	if err := runt(tpl, "foo"); err != nil {
		t.Error(err)
	}
}

func TestSplitn(t *testing.T) {
	tpl := `{{$v := "foo$bar$baz" | splitn "$" 2}}{{$v._0}}`
	if err := runt(tpl, "foo"); err != nil {
		t.Error(err)
	}
}

func TestToString(t *testing.T) {
	tests := []interface{}{
		1,
		"string",
		[]byte("bytes"),
		errors.New("error"),
		stringerImpl{
			Value: "stringer",
		},
	}
	for _, test := range tests {
		tpl := `{{ toString .Value | kindOf }}`
		assert.NoError(t, runtv(tpl, "string", map[string]interface{}{"Value": test}))
	}
}

func TestToStrings(t *testing.T) {
	tpl := `{{ $s := list 1 2 3 | toStrings }}{{ index $s 1 | kindOf }}`
	assert.NoError(t, runt(tpl, "string"))
	tpl = `{{ list 1 .value 2 | toStrings }}`
	values := map[string]interface{}{"value": nil}
	if err := runtv(tpl, `[1 2]`, values); err != nil {
		t.Error(err)
	}
}

func TestJoin(t *testing.T) {
	assert.NoError(t, runt(`{{ tuple "a" "b" "c" | join "-" }}`, "a-b-c"))
	assert.NoError(t, runt(`{{ tuple 1 2 3 | join "-" }}`, "1-2-3"))
	assert.NoError(t, runtv(`{{ join "-" .V }}`, "a-b-c", map[string]interface{}{"V": []string{"a", "b", "c"}}))
	assert.NoError(t, runtv(`{{ join "-" .V }}`, "abc", map[string]interface{}{"V": "abc"}))
	assert.NoError(t, runtv(`{{ join "-" .V }}`, "1-2-3", map[string]interface{}{"V": []int{1, 2, 3}}))
	assert.NoError(t, runtv(`{{ join "-" .value }}`, "1-2", map[string]interface{}{"value": []interface{}{"1", nil, "2"}}))
}

func TestSortAlpha(t *testing.T) {
	// Named `append` in the function map
	tests := map[string]string{
		`{{ list "c" "a" "b" | sortAlpha | join "" }}`: "abc",
		`{{ list 2 1 4 3 | sortAlpha | join "" }}`:     "1234",
	}
	for tpl, expect := range tests {
		assert.NoError(t, runt(tpl, expect))
	}
}
func TestBase64EncodeDecode(t *testing.T) {
	magicWord := "coffee"
	expect := base64.StdEncoding.EncodeToString([]byte(magicWord))

	if expect == magicWord {
		t.Fatal("Encoder doesn't work.")
	}

	tpl := `{{b64enc "coffee"}}`
	if err := runt(tpl, expect); err != nil {
		t.Error(err)
	}
	tpl = fmt.Sprintf("{{b64dec %q}}", expect)
	if err := runt(tpl, magicWord); err != nil {
		t.Error(err)
	}
}
func TestBase32EncodeDecode(t *testing.T) {
	magicWord := "coffee"
	expect := base32.StdEncoding.EncodeToString([]byte(magicWord))

	if expect == magicWord {
		t.Fatal("Encoder doesn't work.")
	}

	tpl := `{{b32enc "coffee"}}`
	if err := runt(tpl, expect); err != nil {
		t.Error(err)
	}
	tpl = fmt.Sprintf("{{b32dec %q}}", expect)
	if err := runt(tpl, magicWord); err != nil {
		t.Error(err)
	}
}

func TestGoutils(t *testing.T) {
	tests := map[string]string{
		`{{abbrev 5 "hello world"}}`:           "he...",
		`{{abbrevboth 5 10 "1234 5678 9123"}}`: "...5678...",
		`{{nospace "h e l l o "}}`:             "hello",
		`{{untitle "First Try"}}`:              "first try", //https://youtu.be/44-RsrF_V_w
		`{{initials "First Try"}}`:             "FT",
		`{{wrap 5 "Hello World"}}`:             "Hello\nWorld",
		`{{wrapWith 5 "\t" "Hello World"}}`:    "Hello\tWorld",
	}
	for k, v := range tests {
		t.Log(k)
		if err := runt(k, v); err != nil {
			t.Errorf("Error on tpl %q: %s", k, err)
		}
	}
}

func TestRandomString(t *testing.T) {
	// Random strings are now using Masterminds/goutils's cryptographically secure random string functions
	// by default. Consequently, these tests now have no predictable character sequence. No checks for exact
	// string output are necessary.

	// {{randAlphaNum 5}} should yield five random characters
	if x, _ := runRaw(`{{randAlphaNum 5}}`, nil); utf8.RuneCountInString(x) != 5 {
		t.Errorf("String should be 5 characters; string was %v characters", utf8.RuneCountInString(x))
	}

	// {{randAlpha 5}} should yield five random characters
	if x, _ := runRaw(`{{randAlpha 5}}`, nil); utf8.RuneCountInString(x) != 5 {
		t.Errorf("String should be 5 characters; string was %v characters", utf8.RuneCountInString(x))
	}

	// {{randAscii 5}} should yield five random characters
	if x, _ := runRaw(`{{randAscii 5}}`, nil); utf8.RuneCountInString(x) != 5 {
		t.Errorf("String should be 5 characters; string was %v characters", utf8.RuneCountInString(x))
	}

	// {{randNumeric 5}} should yield five random characters
	if x, _ := runRaw(`{{randNumeric 5}}`, nil); utf8.RuneCountInString(x) != 5 {
		t.Errorf("String should be 5 characters; string was %v characters", utf8.RuneCountInString(x))
	}
}

func TestCat(t *testing.T) {
	tpl := `{{$b := "b"}}{{"c" | cat "a" $b}}`
	if err := runt(tpl, "a b c"); err != nil {
		t.Error(err)
	}
	tpl = `{{ .value | cat "a" "b"}}`
	values := map[string]interface{}{"value": nil}
	if err := runtv(tpl, "a b", values); err != nil {
		t.Error(err)
	}
}

func TestIndent(t *testing.T) {
	tpl := `{{indent 4 "a\nb\nc"}}`
	if err := runt(tpl, "    a\n    b\n    c"); err != nil {
		t.Error(err)
	}
}

func TestNindent(t *testing.T) {
	tpl := `{{nindent 4 "a\nb\nc"}}`
	if err := runt(tpl, "\n    a\n    b\n    c"); err != nil {
		t.Error(err)
	}
}

func TestReplace(t *testing.T) {
	tpl := `{{"I Am Henry VIII" | replace " " "-"}}`
	if err := runt(tpl, "I-Am-Henry-VIII"); err != nil {
		t.Error(err)
	}
}

func TestPlural(t *testing.T) {
	tpl := `{{$num := len "two"}}{{$num}} {{$num | plural "1 char" "chars"}}`
	if err := runt(tpl, "3 chars"); err != nil {
		t.Error(err)
	}
	tpl = `{{len "t" | plural "cheese" "%d chars"}}`
	if err := runt(tpl, "cheese"); err != nil {
		t.Error(err)
	}
}
