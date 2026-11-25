package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestTheSheetPairsChineseWithItsEnglish(t *testing.T) {
	dir := t.TempDir()
	content := `- text: 一位大娘站在家门口。
  english: An auntie stands at her door.
  choices: [请给一点儿斋饭。, 喂！]
  english_choices: [Please spare a little food., Hey!]
  thanks: 大娘笑了。
  thanks_english: The auntie smiles.
  origin: Seven sisters of Silk Cave (盘丝洞).
`
	if err := os.WriteFile(filepath.Join(dir, "alms.yaml"), []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
	var b strings.Builder
	n, err := write(&b, dir)
	if err != nil || n != 4 {
		t.Fatalf("%d lines %v:\n%s", n, err, b.String())
	}
	for _, want := range []string{
		"| [0].text | 一位大娘站在家门口。 | An auntie stands at her door. | |",
		"| [0].choices[0] | 请给一点儿斋饭。 | Please spare a little food. | |",
		"| [0].thanks | 大娘笑了。 | The auntie smiles. | |",
	} {
		if !strings.Contains(b.String(), want) {
			t.Errorf("sheet lacks %q:\n%s", want, b.String())
		}
	}
	if strings.Contains(b.String(), "Silk Cave") {
		t.Error("an English sentence quoting Chinese is not a line to proofread")
	}
}
