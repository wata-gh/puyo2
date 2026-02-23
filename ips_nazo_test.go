package puyo2

import (
	"strings"
	"testing"
)

func TestParseIPSNazoURLExamples(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		field     string
		haipuyo   string
		condition string
		condCode  [3]int
	}{
		{
			name:      "q1",
			input:     "https://ips.karou.jp/simu/pn.html?800F08J08A0EB_8161__270",
			field:     "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaabaaaaafbaabaffaabaddaafadf",
			haipuyo:   "pryr",
			condition: "色ぷよ全て消す",
			condCode:  [3]int{2, 7, 0},
		},
		{
			name:      "q2",
			input:     "http://ips.karou.jp/simu/pn.html?jjgqqqqqqqqq_q1q1q1__u06",
			field:     "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaececeacecececececececece",
			haipuyo:   "gbgbgb",
			condition: "6連鎖する",
			condCode:  [3]int{30, 0, 6},
		},
		{
			name:      "q3",
			input:     "http://ips.karou.jp/simu/pn.html?80080080oM0oM098_4141__u03",
			field:     "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaabaaaaabaaaaabaaacagaaacagaaabbba",
			haipuyo:   "brbr",
			condition: "3連鎖する",
			condCode:  [3]int{30, 0, 3},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseIPSNazoURL(tt.input)
			if err != nil {
				t.Fatalf("ParseIPSNazoURL() error = %v", err)
			}
			if got.InitialField != tt.field {
				t.Fatalf("InitialField = %s, want %s", got.InitialField, tt.field)
			}
			if got.Haipuyo != tt.haipuyo {
				t.Fatalf("Haipuyo = %s, want %s", got.Haipuyo, tt.haipuyo)
			}
			if got.Condition.Text != tt.condition {
				t.Fatalf("Condition.Text = %s, want %s", got.Condition.Text, tt.condition)
			}
			if got.ConditionCode != tt.condCode {
				t.Fatalf("ConditionCode = %v, want %v", got.ConditionCode, tt.condCode)
			}
		})
	}
}

func TestParseIPSNazoURLRawQuery(t *testing.T) {
	urlInput := "http://ips.karou.jp/simu/pn.html?80080080oM0oM098_4141__u03"
	rawInput := "80080080oM0oM098_4141__u03"

	gotURL, err := ParseIPSNazoURL(urlInput)
	if err != nil {
		t.Fatalf("ParseIPSNazoURL(url) error = %v", err)
	}
	gotRaw, err := ParseIPSNazoURL(rawInput)
	if err != nil {
		t.Fatalf("ParseIPSNazoURL(raw) error = %v", err)
	}

	if gotURL.InitialField != gotRaw.InitialField {
		t.Fatalf("InitialField mismatch: %s != %s", gotURL.InitialField, gotRaw.InitialField)
	}
	if gotURL.Haipuyo != gotRaw.Haipuyo {
		t.Fatalf("Haipuyo mismatch: %s != %s", gotURL.Haipuyo, gotRaw.Haipuyo)
	}
	if gotURL.ConditionCode != gotRaw.ConditionCode {
		t.Fatalf("ConditionCode mismatch: %v != %v", gotURL.ConditionCode, gotRaw.ConditionCode)
	}
}

func TestParseIPSNazoURLDataMode(t *testing.T) {
	input := "~" + strings.Repeat("0", 77) + "1"
	got, err := ParseIPSNazoURL(input)
	if err != nil {
		t.Fatalf("ParseIPSNazoURL() error = %v", err)
	}
	wantField := strings.Repeat("a", 77) + "b"
	if got.InitialField != wantField {
		t.Fatalf("InitialField = %s, want %s", got.InitialField, wantField)
	}
	if got.Haipuyo != "" {
		t.Fatalf("Haipuyo = %s, want empty", got.Haipuyo)
	}
}

func TestParseIPSNazoURLInvalidCharacter(t *testing.T) {
	_, err := ParseIPSNazoURL("!")
	if err == nil {
		t.Fatal("ParseIPSNazoURL() must return error for invalid character")
	}
}

func TestParseIPSNazoURLUnsupportedCellValues(t *testing.T) {
	tests := []string{".", "?"}
	for _, in := range tests {
		_, err := ParseIPSNazoURL(in)
		if err == nil {
			t.Fatalf("ParseIPSNazoURL(%q) must return error for unsupported field cell", in)
		}
	}
}
