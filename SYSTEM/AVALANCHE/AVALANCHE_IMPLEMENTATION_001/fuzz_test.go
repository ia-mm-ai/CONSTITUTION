package main

import "testing"

func FuzzParseGenesisNeverPanics(f *testing.F) {
	f.Add([]byte(`{"format":"PRESENCE_RUNTIME_GENESIS"}`))
	f.Add([]byte(`null`))
	f.Fuzz(func(_ *testing.T, input []byte) {
		_, _ = parseGenesis(input)
	})
}

func FuzzParseTransitionNeverPanics(f *testing.F) {
	f.Add([]byte(`{"unsigned":{},"signature":""}`))
	f.Add([]byte(`null`))
	f.Fuzz(func(_ *testing.T, input []byte) {
		_, _, _ = parseTransition(input)
	})
}

func FuzzParseBlockNeverPanics(f *testing.F) {
	f.Add([]byte(`{"schema":"LOCALITY_BLOCK_003"}`))
	f.Add([]byte(`[]`))
	f.Fuzz(func(_ *testing.T, input []byte) {
		_, _, _, _ = parseWireBlock(input)
	})
}
