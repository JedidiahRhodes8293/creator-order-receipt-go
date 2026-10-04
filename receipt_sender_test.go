package main

import "testing"

func TestProcessDownloadURLTrimsAssetLink(t *testing.T) {
	cases := []struct{ name, input, want string }{{"signed asset", "  https://cdn.example/book.pdf ", "https://cdn.example/book.pdf"}, {"empty", "", ""}}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := processDownloadURL(tc.input); got != tc.want {
				t.Fatalf("got %q want %q", got, tc.want)
			}
		})
	}
}
