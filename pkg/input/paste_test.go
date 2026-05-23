package input

import "testing"

func TestBuildPasteMacro(t *testing.T) {
	steps, invalid := BuildPasteMacro("en_US", "A!\n", 35)
	if len(invalid) != 0 {
		t.Fatalf("unexpected invalid runes: %v", invalid)
	}
	if len(steps) != 6 {
		t.Fatalf("expected 6 macro steps, got %d", len(steps))
	}
	if steps[0].Modifier != 0x02 || steps[0].Keys[0] != 4 || steps[0].Delay != 20 {
		t.Fatalf("unexpected first step: %+v", steps[0])
	}
	if steps[1].Modifier != 0 || steps[1].Keys[0] != 0 || steps[1].Delay != 35 {
		t.Fatalf("unexpected reset step: %+v", steps[1])
	}
	if steps[2].Modifier != 0x02 || steps[2].Keys[0] != 30 {
		t.Fatalf("unexpected punctuation step: %+v", steps[2])
	}
	if steps[4].Modifier != 0 || steps[4].Keys[0] != 40 {
		t.Fatalf("unexpected enter step: %+v", steps[4])
	}
}

func TestBuildPasteMacroReportsInvalidRunes(t *testing.T) {
	_, invalid := BuildPasteMacro("en_US", "ok€é", 20)
	if len(invalid) != 2 || invalid[0] != '€' || invalid[1] != 'é' {
		t.Fatalf("unexpected invalid runes: %v", invalid)
	}
}

func TestBuildPasteMacro_AllSupportedLayouts(t *testing.T) {
	tests := []struct {
		layout string
		text   string
	}{
		{"en-US", "Hello, World! 123 @#$"},
		{"en-UK", "Hello, World! 123 @#$"},
		{"pt-PT", "Olá! Ação,ação. 123"},
		{"de-DE", "Hallo! Über äöü ß 123"},
		{"es-ES", "¡Hola! ñ 123 @#"},
		{"fr-FR", "Bonjour! é è à ç 123"},
		{"it-IT", "Ciao! è é à ò ù 123"},
	}
	for _, tt := range tests {
		t.Run(tt.layout, func(t *testing.T) {
			steps, invalid := BuildPasteMacro(tt.layout, tt.text, 25)
			if len(steps) == 0 {
				t.Fatalf("produced zero steps for %q", tt.text)
			}
			if len(invalid) > 0 {
				t.Errorf("invalid runes for layout %s: %v", tt.layout, invalid)
			}
			for i, s := range steps {
				if i%2 == 0 && s.Delay != 20 {
					continue
				}
				if i%2 == 1 && s.Delay != 25 {
					t.Errorf("step %d: release delay = %d, want 25", i, s.Delay)
				}
			}
		})
	}
}

func TestBuildPasteMacro_UnknownLayoutFallsBackToEnUS(t *testing.T) {
	stepsUnknown, invalidUnknown := BuildPasteMacro("xx-XX", "Hello", 25)
	stepsEnUS, invalidEnUS := BuildPasteMacro("en-US", "Hello", 25)

	if len(stepsUnknown) != len(stepsEnUS) {
		t.Fatalf("unknown layout produced %d steps, en-US produced %d — fallback broken",
			len(stepsUnknown), len(stepsEnUS))
	}
	if len(invalidUnknown) != len(invalidEnUS) {
		t.Fatalf("unknown layout invalid runes %d, en-US %d — fallback broken",
			len(invalidUnknown), len(invalidEnUS))
	}
	for i := range stepsUnknown {
		if stepsUnknown[i] != stepsEnUS[i] {
			t.Fatalf("step %d differs between unknown and en-US fallback", i)
		}
	}
}

func TestBuildPasteMacro_NFCNormalization(t *testing.T) {
	// NFD "á" = 'a' + combining acute (U+0301). NFC normalizes to U+00E1.
	nfd := "a\u0301"
	steps, invalid := BuildPasteMacro("pt-PT", nfd, 25)
	if len(invalid) > 0 {
		t.Fatalf("NFD á reported as invalid on pt-PT: %v", invalid)
	}
	if len(steps) == 0 {
		t.Fatal("NFD á produced zero steps")
	}
}
