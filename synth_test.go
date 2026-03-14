package impsynth

import "testing"

var benchmarkVoiceChannels = []int{0, 1, 2}

func writeBenchmarkVoice(opl *Synth, ch int) {
	base := uint16(0)
	localCh := ch
	if ch >= 9 {
		base = 0x100
		localCh = ch - 9
	}

	modSlots := [9]uint16{0, 1, 2, 8, 9, 10, 16, 17, 18}
	carSlots := [9]uint16{3, 4, 5, 11, 12, 13, 19, 20, 21}
	mod := modSlots[localCh]
	car := carSlots[localCh]

	opl.WriteReg(base+0x20+mod, 0x01)
	opl.WriteReg(base+0x20+car, 0x01)
	opl.WriteReg(base+0x40+mod, 0x18)
	opl.WriteReg(base+0x40+car, 0x00)
	opl.WriteReg(base+0x60+mod, 0xF4)
	opl.WriteReg(base+0x60+car, 0xF6)
	opl.WriteReg(base+0x80+mod, 0x55)
	opl.WriteReg(base+0x80+car, 0x14)
	opl.WriteReg(base+0xC0+uint16(localCh), 0x30)
	opl.WriteReg(base+0xA0+uint16(localCh), 0x98)
	opl.WriteReg(base+0xB0+uint16(localCh), 0x31)
}

func benchmarkSynth(sampleRate int) *Synth {
	opl := New(sampleRate)
	opl.WriteReg(0x01, 0x20)
	for _, ch := range benchmarkVoiceChannels {
		writeBenchmarkVoice(opl, ch)
	}
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
