package filter

import "testing"

func TestRulesMatches(t *testing.T) {
	tests := []struct {
		name    string
		rules   Rules
		title   string
		want    bool
		wantErr bool
	}{
		{name: "no rules", rules: Rules{}, title: "anything", want: true},
		{name: "include match", rules: Rules{Include: []string{"1080p"}}, title: "[Sub] Show - 01 [1080p]", want: true},
		{name: "include miss", rules: Rules{Include: []string{"1080p"}}, title: "[Sub] Show - 01 [720p]", want: false},
		{name: "exclude match", rules: Rules{Exclude: []string{"HEVC"}}, title: "[Sub] Show - 01 [1080p][HEVC]", want: false},
		{name: "include and exclude", rules: Rules{Include: []string{"1080p"}, Exclude: []string{"720p"}}, title: "[Sub] Show - 01 [1080p]", want: true},
		{name: "case insensitive substring", rules: Rules{Include: []string{"1080P"}}, title: "[Sub] Show - 01 [1080p]", want: true},
		{name: "regex include", rules: Rules{Include: []string{`1080[pP]|2160[pP]`}, Regex: true}, title: "[Sub] Show - 01 [2160P]", want: true},
		{name: "invalid regex", rules: Rules{Include: []string{"["}, Regex: true}, title: "x", wantErr: true},
		{
			name: "exclude unless keeps jianfan",
			rules: Rules{
				Include: []string{"1080p", "简"},
				ExcludeUnless: []UnlessExclude{{Pattern: "繁", Unless: "简"}},
			},
			title: "[LoliHouse] Show - 01 [1080p][简繁内封字幕]",
			want:  true,
		},
		{
			name: "exclude unless drops fan only",
			rules: Rules{
				Include: []string{"1080p", "简"},
				ExcludeUnless: []UnlessExclude{{Pattern: "繁", Unless: "简"}},
			},
			title: "[Sub] Show - 08 [1080p][繁日双语]",
			want:  false,
		},
		{
			name: "exclude fan blocks jianfan with plain exclude",
			rules: Rules{
				Include: []string{"1080p", "简"},
				Exclude: []string{"繁"},
			},
			title: "[LoliHouse] Show - 01 [1080p][简繁内封字幕]",
			want:  false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := tc.rules.Matches(tc.title)
			if tc.wantErr {
				if err == nil {
					t.Fatal("expected error")
				}
				return
			}
			if err != nil {
				t.Fatalf("Matches: %v", err)
			}
			if got != tc.want {
				t.Errorf("Matches(%q) = %v, want %v", tc.title, got, tc.want)
			}
		})
	}
}