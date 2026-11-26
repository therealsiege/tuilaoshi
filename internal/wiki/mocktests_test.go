package wiki

import "testing"

func TestMockSittingsAreSavedInOrder(t *testing.T) {
	s := NewStore(t.TempDir())
	if got, err := s.MockSittings(); err != nil || len(got) != 0 {
		t.Fatalf("fresh store: %v %v", got, err)
	}
	first := MockSitting{Date: "2026-10-26", Level: 4, Total: 150, Max: 300,
		Sections: []MockSection{{Name: "Listening", Score: 40}, {Name: "Reading", Score: 60}, {Name: "Writing", Score: 50}}}
	second := MockSitting{Date: "2026-11-26", Level: 5, Total: 140, Max: 200, Passed: true,
		Sections: []MockSection{{Name: "Listening", Skipped: true}, {Name: "Reading", Score: 70}, {Name: "Writing", Score: 70}}}
	for _, m := range []MockSitting{first, second} {
		if err := s.SaveMockSitting(m); err != nil {
			t.Fatal(err)
		}
	}
	got, err := s.MockSittings()
	if err != nil || len(got) != 2 {
		t.Fatalf("got %v %v", got, err)
	}
	if got[0].Date != first.Date || got[1].Level != 5 || !got[1].Passed || !got[1].Sections[0].Skipped || got[1].Sections[1].Score != 70 {
		t.Errorf("sittings came back as %+v", got)
	}
}
