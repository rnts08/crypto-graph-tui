package main

import "testing"

func TestDefaultSymbols(t *testing.T) {
	if got := defaultSymbols(); len(got) == 0 {
		t.Error("expected default symbols, got empty slice")
	}
}

func TestParseSymbols_Defaults(t *testing.T) {
	if got := parseSymbols(""); len(got) != 4 {
		t.Errorf("expected default 4 symbols, got %v", got)
	}
	if got := parseSymbols("   "); len(got) != 4 {
		t.Errorf("expected default on whitespace, got %v", got)
	}
}

func TestParseSymbols_Flag(t *testing.T) {
	if got := parseSymbols("--help"); len(got) != 4 {
		t.Errorf("expected defaults when passing flag, got %v", got)
	}
	if got := parseSymbols("-x"); len(got) != 4 {
		t.Errorf("expected defaults when passing flag, got %v", got)
	}
}

func TestParseSymbols_Custom(t *testing.T) {
	if got := parseSymbols("btc,eth"); len(got) != 2 {
		t.Errorf("expected 2 symbols, got %v", got)
	}
	if got := parseSymbols("BTC-USD, xrp"); got[1] != "XRP-USD" {
		t.Errorf("expected conversion to XRP-USD, got %v", got)
	}
}

func TestSanitizeSymbols(t *testing.T) {
	cases := []struct{
		input []string
		want []string
	}{
		{[]string{"btc","--foo","ETH"}, []string{"BTC-USD","ETH-USD"}},
		{[]string{""}, defaultSymbols()},
		{[]string{"-bar"}, defaultSymbols()},
	}
	for _, c := range cases {
		got := sanitizeSymbols(c.input)
		if len(got) != len(c.want) {
			t.Errorf("sanitize %v = %v, want %v", c.input, got, c.want)
			continue
		}
		for i := range got {
			if got[i] != c.want[i] {
				t.Errorf("sanitize %v = %v, want %v", c.input, got, c.want)
			}
		}
	}
}
