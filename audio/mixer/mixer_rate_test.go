package mixer

import "testing"

func TestSetStepsPerSecondIsContinuous(t *testing.T) {
	m := New(1000)
	if got := m.sampleOf(500); got != 500*SampleRate/1000 {
		t.Fatalf("接點前 %d", got)
	}
	m.SetStepsPerSecond(4000, 500)
	if got, want := m.sampleOf(500), uint64(500*SampleRate/1000); got != want {
		t.Fatalf("接點 %d want %d", got, want)
	}
	if got, want := m.sampleOf(4500), uint64(500*SampleRate/1000+4000*SampleRate/4000); got != want {
		t.Fatalf("接點後 %d want %d", got, want)
	}
	m.SetStepsPerSecond(4000, 9000) // 速率相同：不改接點
	if m.baseStep != 500 {
		t.Fatal("速率不變仍改了接點")
	}
}
