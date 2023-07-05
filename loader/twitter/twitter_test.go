package twitter

import (
	"testing"
)

func Test_extractShortenURL(t *testing.T) {
	tests := []struct {
		name  string
		args  []urlShortenEntity
		input string
		want  string
	}{
		{
			name: "from https://developer.twitter.com/en/docs/twitter-api/v1/data-dictionary/object-model/entities",
			args: []urlShortenEntity{
				{
					DisplayURL:  "dev.twitter.com",
					ExpandedURL: "http://dev.twitter.com",
					Indices:     []int{32, 54},
					URL:         "http://t.co/p5dOtmnZyu",
				},
				{
					DisplayURL:  "pic.twitter.com/ZSvIEMOPb8",
					ExpandedURL: "https://ton.twitter.com/1.1/ton/data/dm/411031503817039874/411031503833792512/cOkcq9FS.jpg",
					Indices:     []int{55, 78},
					URL:         "https://t.co/ZSvIEMOPb8",
				},
			},
			input: "test $TWTR @twitterapi #hashtag http://t.co/p5dOtmnZyu https://t.co/ZSvIEMOPb8",
			want:  "test $TWTR @twitterapi #hashtag <http://dev.twitter.com|dev.twitter.com> <https://ton.twitter.com/1.1/ton/data/dm/411031503817039874/411031503833792512/cOkcq9FS.jpg|pic.twitter.com/ZSvIEMOPb8>",
		},
		{
			name: "from https://twitter.com/WhiteHouse/status/1663319306820546563",
			args: []urlShortenEntity{
				{
					DisplayURL:  "VA.gov/PAC",
					ExpandedURL: "http://VA.gov/PACT",
					Indices:     []int{194, 217},
					URL:         "https://t.co/ZSvIEMOPb8",
				},
				{
					DisplayURL:  "pic.twitter.com/buPi4Aanfm",
					ExpandedURL: "https://twitter.com/WhiteHouse/status/1663319306820546563/photo/1",
					Indices:     []int{219, 242},
					URL:         "https://t.co/DGNEd7CNQ",
				},
			},
			input: "With the PACT Act, President Biden is delivering on his promise to strengthen health care and benefits for America’s veterans and their survivors.\n\nEligible veterans can sign up for services at https://t.co/emuAN7Ctzw. https://t.co/buPi4Aanfm",
			want:  "With the PACT Act, President Biden is delivering on his promise to strengthen health care and benefits for America’s veterans and their survivors.\n\nEligible veterans can sign up for services at <http://VA.gov/PACT|VA.gov/PAC>. <https://twitter.com/WhiteHouse/status/1663319306820546563/photo/1|pic.twitter.com/buPi4Aanfm>",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := extractShortenURL(tt.input, tt.args); got != tt.want {
				t.Errorf("extractShortenURL() = %v, want %v", got, tt.want)
			}
		})
	}
}
