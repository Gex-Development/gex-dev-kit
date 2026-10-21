package theme

import (
	"testing"

	"charm.land/bubbles/v2/key"
)

func TestHuhLightDiffersFromDark(t *testing.T) {
	dark := New(Options{Dark: true}).Huh().Theme(true)
	light := New(Options{Dark: false}).Huh().Theme(false)
	if got := tokensHex(dark.Focused.Title.GetForeground()); got != "#7fc3ef" {
		t.Fatalf("título escuro do huh = %s, esperado #7fc3ef", got)
	}
	if got := tokensHex(light.Focused.Title.GetForeground()); got != "#1a5fc0" {
		t.Fatalf("título claro do huh = %s, esperado #1a5fc0", got)
	}
	if dark.Focused.Title.Render("x") == light.Focused.Title.Render("x") {
		t.Fatalf("título do huh igual nos temas claro e escuro")
	}
}

func TestHuhUsesThemeMode(t *testing.T) {
	styles := New(Options{Dark: true}).Huh().Theme(false)
	if got := tokensHex(styles.Focused.Title.GetForeground()); got != "#7fc3ef" {
		t.Fatalf("huh ignorou o modo do tema: título = %s, esperado #7fc3ef", got)
	}
}

func TestHuhKeyMapInPortuguese(t *testing.T) {
	km := New(Options{}).HuhKeyMap()
	cases := []struct {
		name    string
		binding key.Binding
		desc    string
	}{
		{"Input.Next", km.Input.Next, "próximo"},
		{"Input.Prev", km.Input.Prev, "voltar"},
		{"Input.Submit", km.Input.Submit, "enviar"},
		{"Select.Next", km.Select.Next, "selecionar"},
		{"Select.Filter", km.Select.Filter, "filtrar"},
		{"MultiSelect.Toggle", km.MultiSelect.Toggle, "alternar"},
		{"Confirm.Accept", km.Confirm.Accept, "sim"},
	}
	for _, c := range cases {
		if got := c.binding.Help().Desc; got != c.desc {
			t.Fatalf("%s: descrição = %q, esperado %q", c.name, got, c.desc)
		}
	}
	if got := km.Select.Next.Help().Key; got != "enter" {
		t.Fatalf("tecla de Select.Next = %q, esperado enter", got)
	}
}
