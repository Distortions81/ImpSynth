package impsynth

import "testing"

func benchmarkSynth(sampleRate int) *Synth {
	opl := New(sampleRate)
	opl.WriteReg(0x01, 0x20)
	opl.WriteReg(0x20, 0x01)
	opl.WriteReg(0x23, 0x01)
	opl.WriteReg(0x40, 0x18)
	opl.WriteReg(0x43, 0x00)
	opl.WriteReg(0x60, 0xF4)
	opl.WriteReg(0x63, 0xF6)
	opl.WriteReg(0x80, 0x55)
	opl.WriteReg(0x83, 0x14)
	opl.WriteReg(0xC0, 0x30)
	opl.WriteReg(0xA0, 0x98)
	opl.WriteReg(0xB0, 0x31)
	return opl
}

func TestGenerateStereoS16ProducesPCM(t *testing.T) {
	opl := New(49716)
	opl.WriteReg(0x01, 0x20)
	opl.WriteReg(0x20, 0x01)
	opl.WriteReg(0x23, 0x01)
	opl.WriteReg(0x60, 0xF3)
	opl.WriteReg(0x63, 0xF3)
	opl.WriteReg(0x80, 0x24)
	opl.WriteReg(0x83, 0x24)
	opl.WriteReg(0xA0, 0x98)
	opl.WriteReg(0xB0, 0x31)
	opl.WriteReg(0x43, 0x00)
	opl.WriteReg(0xC0, 0x30)
	pcm := opl.GenerateStereoS16(256)
	if len(pcm) != 512 {
		t.Fatalf("samples=%d want=512", len(pcm))
	}
	nonZero := false
	for _, s := range pcm {
		if s != 0 {
			nonZero = true
			break
		}
	}
	if !nonZero {
		t.Fatal("expected audible PCM")
	}
}

func TestGenerateStereoS16ReusesBuffer(t *testing.T) {
	opl := New(49716)
	opl.WriteReg(0x20, 0x01)
	opl.WriteReg(0x23, 0x01)
	opl.WriteReg(0xA0, 0x98)
	opl.WriteReg(0xB0, 0x31)
	opl.WriteReg(0x43, 0x00)
	_ = opl.GenerateStereoS16(256)
	allocs := testing.AllocsPerRun(100, func() {
		_ = opl.GenerateStereoS16(256)
	})
	if allocs != 0 {
		t.Fatalf("GenerateStereoS16 allocs=%v want 0", allocs)
	}
}

func BenchmarkGenerateStereoS16_2048Frames(b *testing.B) {
	opl := benchmarkSynth(49716)
	b.ReportAllocs()
	b.SetBytes(2048 * 2 * 2)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = opl.GenerateStereoS16(2048)
	}
}

func BenchmarkGenerateStereoS16_2048Frames_44100Hz(b *testing.B) {
	opl := benchmarkSynth(44100)
	b.ReportAllocs()
	b.SetBytes(2048 * 2 * 2)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = opl.GenerateStereoS16(2048)
	}
}
