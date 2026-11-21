package drill

import (
	"fmt"
	"math/rand"
	"strings"

	"github.com/therealsiege/tuilaoshi/internal/journey"
	"github.com/therealsiege/tuilaoshi/internal/mnemonic"
	"github.com/therealsiege/tuilaoshi/internal/model"
	"github.com/therealsiege/tuilaoshi/internal/ui"
	"github.com/therealsiege/tuilaoshi/internal/ui/imgview"
)

func (m Model) View() string {
	var b strings.Builder

	title := ui.TitleStyle.Render("Drill 练习")
	b.WriteString(title)
	b.WriteString("\n\n")
	b.WriteString(ui.AudioError(m.ttsErr))
	if m.saveErr != nil {
		b.WriteString(ui.ErrorStyle.Render(fmt.Sprintf("  Could not save progress: %v", m.saveErr)))
		b.WriteString("\n\n")
	}

	switch m.state {
	case stateModePick:
		b.WriteString(ui.SubtitleStyle.Render("  Select Drill Mode"))
		b.WriteString("\n\n")
		for i, opt := range modeOptions {
			label := fmt.Sprintf("  %s", opt.label)
			desc := ui.MutedStyle.Render("    " + opt.desc)
			if i == m.modeCursor {
				b.WriteString(ui.SelectedStyle.Render("  ▸ "+label) + "\n")
				b.WriteString(desc + "\n")
			} else {
				b.WriteString(ui.UnselectedStyle.Render("    "+label) + "\n")
			}
		}
		b.WriteString("\n")
		b.WriteString(ui.HelpStyle.Render("  ↑↓ select · enter confirm · esc back"))

	case stateLoading:
		b.WriteString(ui.MutedStyle.Render("Loading cards..."))

	case statePrompt:
		if m.current < len(m.cards) {
			if m.encounter != nil {
				b.WriteString(m.renderEncounter())
				b.WriteString("\n")
			}
			card := m.cards[m.current]
			b.WriteString(m.renderCard(card))
			b.WriteString("\n\n")
			b.WriteString("  " + m.input.View())
			b.WriteString("\n\n")
			b.WriteString(ui.HelpStyle.Render("  enter submit · ctrl+space toggle input · ctrl+r replay · esc back"))
		}

	case stateEval:
		if m.encounter != nil {
			b.WriteString(m.renderEncounter())
			b.WriteString("\n")
		}
		b.WriteString(ui.MutedStyle.Render("  Evaluating with Claude..."))

	case stateDifficulty:
		if m.encounter != nil {
			b.WriteString(m.renderEncounter())
			b.WriteString("\n")
		}
		if m.current < len(m.cards) {
			card := m.cards[m.current]
			b.WriteString(m.renderCard(card))
			b.WriteString("\n\n")
		}
		if m.result.Correct {
			b.WriteString(ui.SuccessStyle.Render("  ✓ Correct!"))
		} else {
			b.WriteString(ui.ErrorStyle.Render("  ✗ Incorrect"))
		}
		// Show encounter victory message
		if m.encounter != nil && m.encounterHits >= 3 {
			b.WriteString("  ")
			b.WriteString(ui.SuccessStyle.Render(
				m.encounter.Demon.English + " defeated!"))
		}
		b.WriteString("\n\n  ")
		b.WriteString(m.feedback)
		if m.current < len(m.cards) {
			if line := formatDecomposition(m.cards[m.current].Chinese); line != "" {
				b.WriteString("\n\n")
				b.WriteString(ui.MutedStyle.Render(line))
			}
		}
		// Companion dialogue
		if m.result.Correct {
			if d := journey.RandomDialogue("streak", m.dialogueCount); d != nil {
				b.WriteString("\n")
				b.WriteString(ui.MutedStyle.Render("  " + ui.ChineseStyle.Render(d.Speaker) + ": " + d.Line))
			}
		} else {
			if d := journey.RandomDialogue("mistake", m.dialogueCount); d != nil {
				b.WriteString("\n")
				b.WriteString(ui.MutedStyle.Render("  " + ui.ChineseStyle.Render(d.Speaker) + ": " + d.Line))
			}
		}
		b.WriteString("\n\n")
		b.WriteString("  Rate difficulty:\n")
		fmt.Fprintf(&b, "  %s  %s  %s  %s\n",
			ui.DiffAgainStyle.Render("[1] Again"),
			ui.DiffHardStyle.Render("[2] Hard"),
			ui.DiffGoodStyle.Render("[3] Good"),
			ui.DiffEasyStyle.Render("[4] Easy"),
		)
		b.WriteString("\n")
		b.WriteString(ui.HelpStyle.Render("  1-4 rate · enter default (Good) · esc back"))

	case stateFeedback:
		// Transitional state, same as difficulty
		b.WriteString(ui.MutedStyle.Render("  Processing..."))

	case stateDone:
		if m.err != nil {
			b.WriteString(ui.ErrorStyle.Render(fmt.Sprintf("Error: %v", m.err)))
		} else if m.total == 0 {
			b.WriteString(ui.MutedStyle.Render("  No cards available for drilling."))
			b.WriteString("\n\n")
			b.WriteString(ui.MutedStyle.Render("  Add character, vocabulary, or grammar pages to the wiki to start drilling."))
		} else {
			b.WriteString(ui.SubtitleStyle.Render("  Session Complete 完成"))
			b.WriteString("\n\n")
			fmt.Fprintf(&b, "  Cards:     %d\n", m.total)
			fmt.Fprintf(&b, "  Correct:   %d\n", m.correct)
			if m.maxCombo > 0 {
				fmt.Fprintf(&b, "  Best combo: %d\n", m.maxCombo)
			}
			pct := float64(m.correct) / float64(m.total) * 100
			scoreText := fmt.Sprintf("  Score:   %.0f%%", pct)
			if pct >= 80 {
				b.WriteString(ui.SuccessStyle.Render(scoreText))
			} else if pct >= 50 {
				b.WriteString(ui.WarningStyle.Render(scoreText))
			} else {
				b.WriteString(ui.ErrorStyle.Render(scoreText))
			}
			b.WriteString("\n\n")
			b.WriteString("  " + ui.ProgressBar(m.correct, m.total, 30))
			if d := journey.RandomDialogue("review_complete", m.dialogueCount); d != nil {
				b.WriteString("\n\n")
				b.WriteString(ui.MutedStyle.Render("  " + ui.ChineseStyle.Render(d.Speaker) + ": " + d.Line))
			}
		}
		b.WriteString("\n\n")
		b.WriteString(ui.HelpStyle.Render("  esc back to dashboard"))

	}

	// Progress bar with combo counter
	if m.state != stateDone && m.state != stateModePick && len(m.cards) > 0 {
		if m.comboMsg != "" {
			b.WriteString("\n")
			b.WriteString(ui.SuccessStyle.Render(m.comboMsg))
		}
		b.WriteString("\n\n")
		progress := fmt.Sprintf("  Card %d/%d", m.current+1, len(m.cards))
		if m.total > 0 {
			progress += fmt.Sprintf("  •  %d/%d correct", m.correct, m.total)
		}
		progress += "  •  " + comboIndicator(m.combo)
		if m.maxCombo > 0 {
			progress += "  •  " + ui.MutedStyle.Render(fmt.Sprintf("Best: %d", m.maxCombo))
		}
		b.WriteString(ui.StatusBarStyle.Render(progress))
	}

	return b.String()
}

// renderEncounter renders the demon encounter narrative frame.
func (m Model) renderEncounter() string {
	enc := m.encounter
	var b strings.Builder

	// Demon art: prefer terminal image, fall back to ASCII art
	imgPath := fmt.Sprintf("assets/images/demons/%s.png", enc.Demon.Name)
	artContent := imgview.RenderWithFallback(imgPath, 40, 15, enc.Demon.Art)
	for _, line := range strings.Split(artContent, "\n") {
		b.WriteString(ui.ErrorStyle.Render("      " + line))
		b.WriteString("\n")
	}
	b.WriteString("\n")

	// Demon name and title
	b.WriteString("      ")
	b.WriteString(ui.ChineseStyle.Render(enc.Demon.Name))
	b.WriteString("  ")
	b.WriteString(ui.MutedStyle.Render(enc.Demon.English))
	b.WriteString("\n")

	// Narrative intro
	b.WriteString("  ")
	b.WriteString(ui.PinyinStyle.Render(enc.IntroEng))
	b.WriteString("\n")

	// Challenge text and progress
	b.WriteString("  ")
	b.WriteString(ui.WarningStyle.Render(enc.Challenge))
	b.WriteString("\n")
	hits := m.encounterHits
	if hits > 3 {
		hits = 3
	}
	swords := strings.Repeat("X", hits) + strings.Repeat("-", 3-hits)
	fmt.Fprintf(&b, "  %s [%s] %d/3",
		ui.ErrorStyle.Render("⚔"),
		swords,
		hits,
	)

	return b.String()
}

func (m Model) renderCard(card model.Card) string {
	mode := m.effectiveMode()

	var content string
	switch mode {
	case model.ModeRecognition:
		chinese := ui.ChineseStyle.Render(card.Chinese)
		content = fmt.Sprintf("  What does this mean?\n\n      %s", chinese)
		if card.Pinyin != "" {
			content += "\n      " + ui.PinyinStyle.Render(card.Pinyin)
		}
	case model.ModeProduction:
		content = fmt.Sprintf("  How do you say this in Chinese?\n\n      %s", card.English)
		if card.Pinyin != "" {
			content += "\n      " + ui.PinyinStyle.Render("("+card.Pinyin+")")
		}
	case model.ModeListening:
		content = "  What did you hear? Type it in Chinese.\n\n"
		content += "      " + ui.ChineseStyle.Render("🔊 [Listen]")
	case model.ModeSentence:
		if len(card.Examples) > 0 {
			ex := card.Examples[rand.Intn(len(card.Examples))]
			ex = strings.TrimPrefix(ex, "- ")
			content = fmt.Sprintf("  Fill in the blank:\n\n      %s\n\n  Hint: %s",
				blankOut(ex, card.Chinese),
				ui.PinyinStyle.Render(card.English))
		} else {
			chinese := ui.ChineseStyle.Render(card.Chinese)
			content = fmt.Sprintf("  What does this mean?\n\n      %s", chinese)
		}
	}

	typeLabel := ui.MutedStyle.Render(fmt.Sprintf("[%s]", card.Type))
	modeLabel := ui.MutedStyle.Render(fmt.Sprintf(" · %s", mode))
	header := typeLabel + modeLabel
	separator := ui.MutedStyle.Render("  " + strings.Repeat("─", 24))
	return ui.CardStyle.Render(header + "\n" + separator + "\n" + content)
}

// blankOut replaces the target word in a sentence with _____.
func blankOut(sentence, word string) string {
	if word == "" {
		return sentence
	}
	if idx := strings.Index(sentence, word); idx >= 0 {
		return sentence[:idx] + ui.ChineseStyle.Render("_____") + sentence[idx+len(word):]
	}
	return sentence + " " + ui.ChineseStyle.Render("_____")
}

// formatDecomposition returns a single mnemonic line for a character, or ""
// if no decomposition is known. Example output:
//
//	部 女 (woman) + 子 (child) → A woman holding her child — that is the definition of GOOD.
func formatDecomposition(char string) string {
	d, ok := mnemonic.Lookup(char)
	if !ok {
		return ""
	}
	var parts []string
	for _, c := range d.Components {
		parts = append(parts, fmt.Sprintf("%s (%s)", c.Part, c.Meaning))
	}
	return fmt.Sprintf("  部 %s → %s", strings.Join(parts, " + "), d.Story)
}

// comboIndicator returns the styled combo display string.
func comboIndicator(combo int) string {
	if combo == 0 {
		return ui.MutedStyle.Render("Combo: 0")
	}
	switch {
	case combo >= 20:
		return ui.SuccessStyle.Bold(true).Render(fmt.Sprintf("****x%d**** UNSTOPPABLE", combo))
	case combo >= 10:
		return ui.SuccessStyle.Render(fmt.Sprintf("***x%d***", combo))
	case combo >= 5:
		return ui.WarningStyle.Render(fmt.Sprintf("**x%d**", combo))
	case combo >= 3:
		return fmt.Sprintf("*x%d*", combo)
	default:
		return fmt.Sprintf("Combo: %d", combo)
	}
}

// comboMilestone returns a celebration message for milestone combos.
// Returns empty string for non-milestone values.
func comboMilestone(combo int) string {
	switch {
	case combo >= 30 && combo%5 == 0:
		return fmt.Sprintf("  连续%d个！齐天大圣！", combo)
	case combo == 25:
		return "  连续25个！势不可挡！"
	case combo == 20:
		return "  连续20个！势不可挡！"
	case combo == 15:
		return "  连续15个！太厉害了！"
	case combo == 10:
		return "  连续10个！太厉害了！"
	case combo == 5:
		return "  连续5个！不错！"
	default:
		return ""
	}
}
