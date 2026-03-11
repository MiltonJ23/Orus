package views

import (
	"context"
	"fmt"
	"image"
	"image/color"
	"log"
	"math"
	"strings"
	"time"

	"gioui.org/f32"
	"gioui.org/font"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/text"
	"gioui.org/unit"
	"gioui.org/widget"
	"gioui.org/widget/material"

	"github.com/MiltonJ23/Orus/internal/adapters/ui/theme"
	"github.com/MiltonJ23/Orus/internal/domain"
)

// ── Background presets ────────────────────────────────────────────────────────
var readerBgPresets = []color.NRGBA{
	{R: 250, G: 249, B: 246, A: 255}, // 0 – parchemin clair
	{R: 18, G: 18, B: 28, A: 255},    // 1 – nuit profonde
	{},                               // 2 – XMB animé
	{R: 30, G: 30, B: 46, A: 255},    // 3 – bleu nuit
	{R: 10, G: 40, B: 30, A: 255},    // 4 – vert forêt
	{R: 45, G: 20, B: 50, A: 255},    // 5 – violet sombre
	{R: 60, G: 30, B: 10, A: 255},    // 6 – brun tabac
	{R: 240, G: 235, B: 215, A: 255}, // 7 – crème journal
	{R: 200, G: 220, B: 240, A: 255}, // 8 – ciel pâle
}

var readerBgLabels = []string{
	"Clair", "Nuit", "XMB", "Bleu", "Forêt", "Violet", "Brun", "Crème", "Ciel",
}

// ── Main entry point ──────────────────────────────────────────────────────────
func (wm *WindowManager) drawReaderView(gtx layout.Context) layout.Dimensions {
	// Async text loading
	if len(wm.readerContent) == 0 && wm.readerBook != nil && !wm.readerLoading {
		wm.readerLoading = true
		book := wm.readerBook
		go func() {
			if wm.contentReader == nil {
				wm.readerContent = []string{"Lecteur non disponible."}
				wm.readerLoading = false
				wm.window.Invalidate()
				return
			}
			chunks, err := wm.contentReader.ReadBookText(context.Background(), book.FilePath)
			if err != nil {
				log.Printf("[Reader] Erreur lecture : %v", err)
				wm.readerContent = []string{fmt.Sprintf(
					"Impossible de lire ce fichier.\n\nErreur : %v\n\nFormats supportés : PDF, EPUB.", err)}
			} else {
				wm.readerContent = chunks
			}
			if wm.readerSession != nil && wm.readerSession.CurrentPage > 1 {
				p := wm.readerSession.CurrentPage - 1
				if p < len(wm.readerContent) {
					wm.readerPage = p
				}
			}
			wm.readerLoading = false
			wm.window.Invalidate()
		}()
	}

	// Background
	wm.drawReaderBackground(gtx)
	if wm.readerBgMode == 2 {
		wm.window.Invalidate()
	}

	// Dim overlay (applied on top of background, under content)
	if wm.readerDimAlpha > 0 {
		cl := clip.Rect{Max: gtx.Constraints.Max}.Push(gtx.Ops)
		paint.Fill(gtx.Ops, color.NRGBA{A: wm.readerDimAlpha})
		cl.Pop()
	}

	return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
		layout.Rigid(wm.drawReaderTopBar),
		layout.Flexed(1, wm.drawReaderContent),
		layout.Rigid(wm.drawReaderBottomBar),
		layout.Rigid(wm.drawReaderBgPanel),
	)
}

// ── Background ────────────────────────────────────────────────────────────────
func (wm *WindowManager) drawReaderBackground(gtx layout.Context) {
	if wm.readerBgMode == 2 {
		wm.drawXMBBackground(gtx)
		return
	}
	idx := wm.readerBgMode
	if idx < 0 || idx >= len(readerBgPresets) {
		idx = 0
	}
	paint.Fill(gtx.Ops, readerBgPresets[idx])
}

func (wm *WindowManager) drawXMBBackground(gtx layout.Context) {
	if wm.readerBgAnimStart.IsZero() {
		wm.readerBgAnimStart = time.Now()
	}
	t := time.Since(wm.readerBgAnimStart).Seconds()
	w := float32(gtx.Constraints.Max.X)
	h := float32(gtx.Constraints.Max.Y)

	// Gradient bands: navy → deep purple
	paint.Fill(gtx.Ops, color.NRGBA{R: 8, G: 12, B: 35, A: 255})
	bands := 8
	for i := 0; i < bands; i++ {
		fi := float64(i) / float64(bands)
		r := uint8(8 + fi*50)
		g := uint8(12 + fi*5)
		b := uint8(35 + fi*55)
		bandH := h / float32(bands)
		y := float32(i) * bandH
		cl := clip.Rect{
			Min: image.Point{X: 0, Y: int(y)},
			Max: image.Point{X: int(w), Y: int(y + bandH + 1)},
		}.Push(gtx.Ops)
		paint.Fill(gtx.Ops, color.NRGBA{R: r, G: g, B: b, A: 255})
		cl.Pop()
	}

	// Flowing wave lines
	for wi := 0; wi < 14; wi++ {
		wf := float64(wi)
		speed := 0.18 + wf*0.04
		amplitude := h * (0.025 + float32(wi%3)*0.015)
		yCenter := h * float32(0.1+wf/14.0*0.85)
		alpha := uint8(12 + wi*5)
		if wi > 8 {
			alpha = uint8(wi * 3)
		}
		lineW := float32(0.8 + float64(wi%4)*0.4)
		var p clip.Path
		p.Begin(gtx.Ops)
		steps := 120
		for s := 0; s <= steps; s++ {
			x := w * float32(s) / float32(steps)
			phase := t*speed + wf*0.7 + float64(s)*0.06
			y := yCenter + amplitude*float32(math.Sin(phase))
			if s == 0 {
				p.MoveTo(f32.Pt(x, y))
			} else {
				p.LineTo(f32.Pt(x, y))
			}
		}
		for s := steps; s >= 0; s-- {
			x := w * float32(s) / float32(steps)
			phase := t*speed + wf*0.7 + float64(s)*0.06
			y := yCenter + amplitude*float32(math.Sin(phase)) + lineW
			p.LineTo(f32.Pt(x, y))
		}
		p.Close()
		waveCol := color.NRGBA{R: 80, G: 140, B: 255, A: alpha}
		if wi%3 == 1 {
			waveCol = color.NRGBA{R: 140, G: 80, B: 255, A: alpha}
		} else if wi%3 == 2 {
			waveCol = color.NRGBA{R: 60, G: 190, B: 220, A: alpha}
		}
		paint.FillShape(gtx.Ops, waveCol, clip.Outline{Path: p.End()}.Op())
	}

	// Floating particles
	for pi := 0; pi < 28; pi++ {
		pf := float64(pi)
		progress := math.Mod(t*(0.06+pf*0.003)+pf/28.0, 1.0)
		px := w * float32(progress)
		baseY := h * float32(math.Mod(pf*37.0+11.0, 84.0)/84.0)
		wobble := float32(math.Sin(t*0.4+pf*1.1)) * 18
		py := baseY + wobble
		size := float32(2 + (pi%3)*2)
		alpha := uint8(30 + (pi%4)*18)
		if progress < 0.08 {
			alpha = uint8(float64(alpha) * progress / 0.08)
		} else if progress > 0.92 {
			alpha = uint8(float64(alpha) * (1 - progress) / 0.08)
		}
		pCol := color.NRGBA{R: 160, G: 200, B: 255, A: alpha}
		if pi%4 == 0 {
			var dp clip.Path
			dp.Begin(gtx.Ops)
			dp.MoveTo(f32.Pt(px, py-size))
			dp.LineTo(f32.Pt(px+size/2, py))
			dp.LineTo(f32.Pt(px, py+size))
			dp.LineTo(f32.Pt(px-size/2, py))
			dp.Close()
			paint.FillShape(gtx.Ops, pCol, clip.Outline{Path: dp.End()}.Op())
		} else {
			drawCircle(gtx, px, py, size/2, pCol)
		}
	}

	// Horizontal glow streak
	streakY := h * (0.38 + float32(math.Sin(t*0.12))*0.06)
	for _, g := range []struct{ ow, a int }{{80, 3}, {50, 6}, {20, 12}, {6, 20}} {
		sr := clip.Rect{
			Min: image.Point{X: 0, Y: int(streakY) - g.ow},
			Max: image.Point{X: int(w), Y: int(streakY) + g.ow},
		}.Push(gtx.Ops)
		paint.Fill(gtx.Ops, color.NRGBA{R: 100, G: 160, B: 255, A: uint8(g.a)})
		sr.Pop()
	}
}

// ── Top bar ───────────────────────────────────────────────────────────────────
func (wm *WindowManager) drawReaderTopBar(gtx layout.Context) layout.Dimensions {
	const barH = 56
	textCol := wm.readerTextColor()
	bgCol := wm.readerBarBgColor()

	cl := clip.Rect{Max: image.Point{X: gtx.Constraints.Max.X, Y: barH}}.Push(gtx.Ops)
	paint.Fill(gtx.Ops, bgCol)
	cl.Pop()

	// Left inset = 110dp to clear macOS traffic lights (which end at ~65px)
	return layout.Inset{
		Top: unit.Dp(10), Bottom: unit.Dp(10),
		Left: unit.Dp(110), Right: unit.Dp(20),
	}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		return layout.Flex{Axis: layout.Horizontal, Alignment: layout.Middle}.Layout(gtx,
			// ← Back pill
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				if wm.closeReaderBtn.Clicked(gtx) {
					wm.closeReader()
				}
				return wm.readerPillBtn(gtx, "← Retour", &wm.closeReaderBtn, theme.ColorCyberCyan)
			}),
			layout.Rigid(layout.Spacer{Width: unit.Dp(16)}.Layout),
			// Title
			layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
				title := "Lecture"
				if wm.readerBook != nil {
					title = wm.readerBook.Title
				}
				lbl := material.Label(wm.theme, 13, title)
				lbl.Font.Weight = font.Bold
				lbl.Color = textCol
				lbl.Alignment = text.Middle
				return lbl.Layout(gtx)
			}),
			// Controls row
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				return layout.Flex{Axis: layout.Horizontal, Alignment: layout.Middle}.Layout(gtx,
					// dim group
					layout.Rigid(func(gtx layout.Context) layout.Dimensions {
						if wm.dimPlusBtn.Clicked(gtx) && wm.readerDimAlpha < 200 {
							wm.readerDimAlpha += 20
						}
						return wm.readerIconPill(gtx, "🌙+", &wm.dimPlusBtn)
					}),
					layout.Rigid(func(gtx layout.Context) layout.Dimensions {
						if wm.dimMinusBtn.Clicked(gtx) && wm.readerDimAlpha >= 20 {
							wm.readerDimAlpha -= 20
						}
						return wm.readerIconPill(gtx, "🌙−", &wm.dimMinusBtn)
					}),
					layout.Rigid(layout.Spacer{Width: unit.Dp(6)}.Layout),
					// font group
					layout.Rigid(func(gtx layout.Context) layout.Dimensions {
						if wm.fontMinusBtn.Clicked(gtx) && wm.readerFontSize > 11 {
							wm.readerFontSize -= 1.5
						}
						return wm.readerIconPill(gtx, "A−", &wm.fontMinusBtn)
					}),
					layout.Rigid(func(gtx layout.Context) layout.Dimensions {
						if wm.fontPlusBtn.Clicked(gtx) && wm.readerFontSize < 32 {
							wm.readerFontSize += 1.5
						}
						return wm.readerIconPill(gtx, "A+", &wm.fontPlusBtn)
					}),
					layout.Rigid(layout.Spacer{Width: unit.Dp(6)}.Layout),
					// BG palette
					layout.Rigid(func(gtx layout.Context) layout.Dimensions {
						if wm.readerBgPanelBtn.Clicked(gtx) {
							wm.readerBgPanelOpen = !wm.readerBgPanelOpen
						}
						active := wm.readerBgPanelOpen
						return wm.readerIconPillActive(gtx, "🎨", &wm.readerBgPanelBtn, active)
					}),
				)
			}),
		)
	})
}

// ── Content ───────────────────────────────────────────────────────────────────
func (wm *WindowManager) drawReaderContent(gtx layout.Context) layout.Dimensions {
	textCol := wm.readerTextColor()
	if wm.readerLoading {
		return layout.Center.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
			lbl := material.Label(wm.theme, 16, "Chargement…")
			lbl.Color = theme.WithAlpha(textCol, 150)
			return lbl.Layout(gtx)
		})
	}
	if len(wm.readerContent) == 0 {
		return layout.Center.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
			lbl := material.Label(wm.theme, 15, "Aucun contenu à afficher.")
			lbl.Color = theme.WithAlpha(textCol, 120)
			return lbl.Layout(gtx)
		})
	}

	pageText := wm.readerContent[wm.readerPage]
	// Clean raw text: normalize whitespace, collapse runs of blank lines
	pageText = cleanReaderText(pageText)

	// Detect if this page starts with a chapter marker
	isChapter, chapterLabel, bodyText := parseChapterHeader(pageText)

	return layout.Center.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		maxW := 680
		if gtx.Constraints.Max.X-80 < maxW {
			maxW = gtx.Constraints.Max.X - 80
		}
		gtx.Constraints.Max.X = maxW
		gtx.Constraints.Min.X = maxW

		return layout.Inset{Top: unit.Dp(44), Bottom: unit.Dp(32)}.Layout(gtx,
			func(gtx layout.Context) layout.Dimensions {
				if !isChapter {
					// Plain body text
					lbl := material.Label(wm.theme, unit.Sp(wm.readerFontSize), pageText)
					lbl.Color = textCol
					lbl.LineHeight = unit.Sp(wm.readerFontSize * 1.65)
					return lbl.Layout(gtx)
				}
				// Chapter header + body
				return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
					// Chapter label — small uppercase, accent colour
					layout.Rigid(func(gtx layout.Context) layout.Dimensions {
						return layout.Inset{Bottom: unit.Dp(6)}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
							lbl := material.Label(wm.theme, unit.Sp(11), chapterLabel)
							lbl.Color = theme.WithAlpha(textCol, 120)
							lbl.Font.Weight = font.Bold
							return lbl.Layout(gtx)
						})
					}),
					// Hairline separator
					layout.Rigid(func(gtx layout.Context) layout.Dimensions {
						return layout.Inset{Bottom: unit.Dp(28)}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
							line := clip.Rect{Max: image.Pt(maxW, 1)}.Push(gtx.Ops)
							paint.Fill(gtx.Ops, theme.WithAlpha(textCol, 30))
							line.Pop()
							return layout.Dimensions{Size: image.Pt(maxW, 1)}
						})
					}),
					// Body
					layout.Rigid(func(gtx layout.Context) layout.Dimensions {
						lbl := material.Label(wm.theme, unit.Sp(wm.readerFontSize), bodyText)
						lbl.Color = textCol
						lbl.LineHeight = unit.Sp(wm.readerFontSize * 1.65)
						return lbl.Layout(gtx)
					}),
				)
			})
	})
}

// cleanReaderText normalises raw extracted text for comfortable reading.
func cleanReaderText(s string) string {
	// Replace literal \n sequences (if any escaped) with newlines
	s = strings.ReplaceAll(s, `\n`, "\n")
	// Collapse 3+ consecutive newlines into 2
	for strings.Contains(s, "\n\n\n") {
		s = strings.ReplaceAll(s, "\n\n\n", "\n\n")
	}
	return strings.TrimSpace(s)
}

// parseChapterHeader detects lines like "— Chapitre N —" or "Chapter N" at the
// start of a page. Returns (isChapter, label, bodyText).
func parseChapterHeader(s string) (bool, string, string) {
	lines := strings.SplitN(s, "\n", 4)
	if len(lines) == 0 {
		return false, "", s
	}
	first := strings.TrimSpace(lines[0])
	// Match "— Chapitre N —" or "─── Chapitre N ───" style markers
	isMarker := (strings.HasPrefix(first, "—") || strings.HasPrefix(first, "─") ||
		strings.HasPrefix(first, "===") || strings.HasPrefix(first, "***")) &&
		(strings.Contains(strings.ToLower(first), "chapitre") ||
			strings.Contains(strings.ToLower(first), "chapter") ||
			strings.Contains(strings.ToLower(first), "partie") ||
			strings.Contains(strings.ToLower(first), "part"))
	if !isMarker {
		return false, "", s
	}
	// Strip decoration chars from label
	label := strings.Trim(first, "—─=* \t")
	body := ""
	if len(lines) > 1 {
		body = strings.TrimSpace(strings.Join(lines[1:], "\n"))
	}
	return true, strings.ToUpper(label), body
}

// ── Bottom bar — navigation pills ─────────────────────────────────────────────
func (wm *WindowManager) drawReaderBottomBar(gtx layout.Context) layout.Dimensions {
	if len(wm.readerContent) == 0 {
		return layout.Dimensions{Size: image.Point{Y: 60}}
	}
	total := len(wm.readerContent)
	progress := float32(wm.readerPage+1) / float32(total)
	textCol := wm.readerTextColor()

	const barH = 68
	cl := clip.Rect{Max: image.Point{X: gtx.Constraints.Max.X, Y: barH}}.Push(gtx.Ops)
	paint.Fill(gtx.Ops, wm.readerBarBgColor())
	cl.Pop()

	return layout.Inset{
		Top: unit.Dp(12), Bottom: unit.Dp(10),
		Left: unit.Dp(32), Right: unit.Dp(32),
	}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		return layout.Flex{Axis: layout.Horizontal, Alignment: layout.Middle}.Layout(gtx,

			// ← Précédent pill
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				disabled := wm.readerPage == 0
				if !disabled && wm.readerPrevBtn.Clicked(gtx) {
					wm.readerPage--
					wm.saveReaderProgress()
				}
				return wm.navPill(gtx, "← Précédent", &wm.readerPrevBtn, disabled, textCol)
			}),

			// Center: progress bar + label
			layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
				return layout.Flex{Axis: layout.Vertical, Alignment: layout.Middle}.Layout(gtx,
					layout.Rigid(func(gtx layout.Context) layout.Dimensions {
						bar := material.ProgressBar(wm.theme, progress)
						bar.Color = theme.ColorSandGold
						bar.TrackColor = theme.WithAlpha(theme.ColorCyberCyan, 22)
						return layout.Inset{
							Left: unit.Dp(24), Right: unit.Dp(24), Bottom: unit.Dp(4),
						}.Layout(gtx, bar.Layout)
					}),
					layout.Rigid(func(gtx layout.Context) layout.Dimensions {
						lbl := material.Label(wm.theme, 11,
							fmt.Sprintf("%d / %d  ·  ← → pour naviguer  ·  ESC pour quitter",
								wm.readerPage+1, total))
						lbl.Color = theme.WithAlpha(textCol, 90)
						lbl.Alignment = text.Middle
						return lbl.Layout(gtx)
					}),
				)
			}),

			// Suivant → pill
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				disabled := wm.readerPage >= total-1
				if !disabled && wm.readerNextBtn.Clicked(gtx) {
					wm.readerPage++
					wm.saveReaderProgress()
				}
				return wm.navPill(gtx, "Suivant →", &wm.readerNextBtn, disabled, textCol)
			}),
		)
	})
}

// navPill draws a proper pill-shaped navigation button.
func (wm *WindowManager) navPill(gtx layout.Context, label string, btn *widget.Clickable, disabled bool, textCol color.NRGBA) layout.Dimensions {
	return btn.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		col := theme.ColorSandGold
		bgAlpha := uint8(22)
		if disabled {
			col = theme.WithAlpha(textCol, 45)
			bgAlpha = 8
		} else if btn.Hovered() {
			bgAlpha = 40
		} else if btn.Pressed() {
			bgAlpha = 60
		}
		const pillH = 36
		const pillW = 130
		cl := clip.UniformRRect(image.Rectangle{
			Max: image.Point{X: pillW, Y: pillH},
		}, pillH/2).Push(gtx.Ops)
		paint.Fill(gtx.Ops, theme.WithAlpha(col, bgAlpha))
		cl.Pop()
		gtxPill := gtx
		gtxPill.Constraints = layout.Exact(image.Point{X: pillW, Y: pillH})
		return layout.Center.Layout(
			gtxPill,
			func(gtx layout.Context) layout.Dimensions {
				lbl := material.Label(wm.theme, 13, label)
				if disabled {
					lbl.Color = theme.WithAlpha(textCol, 45)
				} else {
					lbl.Color = theme.ColorSandGold
				}
				lbl.Font.Weight = font.SemiBold
				return lbl.Layout(gtx)
			},
		)
	})
}

// ── Background palette panel ───────────────────────────────────────────────────
func (wm *WindowManager) drawReaderBgPanel(gtx layout.Context) layout.Dimensions {
	if !wm.readerBgPanelOpen {
		return layout.Dimensions{}
	}
	const panelH = 80

	cl := clip.Rect{Max: image.Point{X: gtx.Constraints.Max.X, Y: panelH}}.Push(gtx.Ops)
	paint.Fill(gtx.Ops, color.NRGBA{R: 18, G: 18, B: 28, A: 245})
	cl.Pop()

	return layout.Inset{Top: unit.Dp(12), Left: unit.Dp(24), Right: unit.Dp(24)}.Layout(gtx,
		func(gtx layout.Context) layout.Dimensions {
			var chips []layout.FlexChild
			for i := range readerBgPresets {
				idx := i // capture
				// ── Clicked() MUST be checked BEFORE Layout() ──────────────────
				clicked := wm.readerBgBtns[idx].Clicked(gtx)
				chips = append(chips, layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					if clicked {
						wm.readerBgMode = idx
						if idx == 2 {
							wm.readerBgAnimStart = time.Now()
						}
					}
					return layout.Inset{Right: unit.Dp(10)}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
						return wm.readerBgBtns[idx].Layout(gtx, func(gtx layout.Context) layout.Dimensions {
							active := wm.readerBgMode == idx
							const bw, bh = 54, 38
							swatchCol := readerBgPresets[idx]
							if idx == 2 {
								swatchCol = color.NRGBA{R: 40, G: 60, B: 160, A: 255}
							}
							outer := image.Rectangle{Max: image.Point{X: bw, Y: bh}}
							// Active ring behind swatch
							if active {
								ring := clip.UniformRRect(image.Rectangle{
									Min: image.Point{X: -3, Y: -3},
									Max: image.Point{X: bw + 3, Y: bh + 3},
								}, 9).Push(gtx.Ops)
								paint.Fill(gtx.Ops, color.NRGBA{R: 255, G: 200, B: 50, A: 220})
								ring.Pop()
							}
							// Swatch fill
							cl2 := clip.UniformRRect(outer, 6).Push(gtx.Ops)
							paint.Fill(gtx.Ops, swatchCol)
							cl2.Pop()
							// XMB mini waves decoration
							if idx == 2 {
								for wi := 0; wi < 3; wi++ {
									var p clip.Path
									p.Begin(gtx.Ops)
									for s := 0; s <= 20; s++ {
										x := float32(s) / 20 * bw
										y := float32(bh)/2 + float32(wi*6-6) +
											float32(math.Sin(float64(s)*0.7+float64(wi)))*4
										if s == 0 {
											p.MoveTo(f32.Pt(x, y))
										} else {
											p.LineTo(f32.Pt(x, y+1))
											p.LineTo(f32.Pt(x, y))
										}
									}
									p.Close()
									paint.FillShape(gtx.Ops,
										color.NRGBA{R: 120, G: 180, B: 255, A: 80},
										clip.Outline{Path: p.End()}.Op())
								}
							}
							// Hover highlight
							if wm.readerBgBtns[idx].Hovered() && !active {
								h := clip.UniformRRect(outer, 6).Push(gtx.Ops)
								paint.Fill(gtx.Ops, color.NRGBA{R: 255, G: 255, B: 255, A: 30})
								h.Pop()
							}
							// Label below swatch
							lbl := material.Label(wm.theme, 9, readerBgLabels[idx])
							lbl.Color = color.NRGBA{R: 255, G: 255, B: 255, A: 200}
							lbl.Alignment = text.Middle
							ts := op.Offset(image.Pt(0, bh+3)).Push(gtx.Ops)
							lbl.Layout(gtx)
							ts.Pop()
							return layout.Dimensions{Size: image.Point{X: bw, Y: bh + 14}}
						})
					})
				}))
			}
			return layout.Flex{Axis: layout.Horizontal, Alignment: layout.Start}.Layout(gtx, chips...)
		})
}

// ── Helpers ───────────────────────────────────────────────────────────────────

func (wm *WindowManager) readerTextColor() color.NRGBA {
	switch wm.readerBgMode {
	case 0, 7, 8:
		return theme.ColorPureBlack
	default:
		return theme.ColorGlassWhite
	}
}

func (wm *WindowManager) readerBarBgColor() color.NRGBA {
	switch wm.readerBgMode {
	case 0, 7, 8:
		return theme.WithAlpha(theme.ColorGlassWhite, 235)
	case 2:
		return color.NRGBA{R: 8, G: 12, B: 35, A: 220}
	default:
		return theme.WithAlpha(readerBgPresets[wm.readerBgMode], 230)
	}
}

// readerIconPill draws a small pill icon button.
func (wm *WindowManager) readerIconPill(gtx layout.Context, icon string, btn *widget.Clickable) layout.Dimensions {
	return wm.readerIconPillActive(gtx, icon, btn, false)
}

func (wm *WindowManager) readerIconPillActive(gtx layout.Context, icon string, btn *widget.Clickable, active bool) layout.Dimensions {
	return btn.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		isDark := wm.readerBgMode != 0 && wm.readerBgMode != 7 && wm.readerBgMode != 8
		const pH, pW = 34, 56

		var bgCol color.NRGBA
		var iconCol color.NRGBA
		var borderCol color.NRGBA

		if isDark {
			bgCol = color.NRGBA{R: 255, G: 255, B: 255, A: 28}
			borderCol = color.NRGBA{R: 255, G: 255, B: 255, A: 70}
			iconCol = color.NRGBA{R: 255, G: 255, B: 255, A: 255}
			if active {
				bgCol = color.NRGBA{R: 245, G: 166, B: 35, A: 55}
				borderCol = theme.WithAlpha(theme.ColorSandGold, 200)
				iconCol = theme.ColorSandGold
			}
			if btn.Hovered() {
				bgCol.A += 35
			}
			if btn.Pressed() {
				bgCol.A += 55
			}
		} else {
			bgCol = color.NRGBA{R: 30, G: 30, B: 45, A: 18}
			borderCol = color.NRGBA{R: 30, G: 30, B: 45, A: 60}
			iconCol = color.NRGBA{R: 30, G: 30, B: 45, A: 230}
			if active {
				bgCol = theme.WithAlpha(theme.ColorSandGold, 35)
				borderCol = theme.WithAlpha(theme.ColorSandGold, 180)
				iconCol = color.NRGBA{R: 160, G: 100, B: 10, A: 255}
			}
			if btn.Hovered() {
				bgCol.A += 20
			}
			if btn.Pressed() {
				bgCol.A += 35
			}
		}

		// Visible border ring
		border := clip.UniformRRect(image.Rectangle{Max: image.Point{X: pW, Y: pH}}, pH/2).Push(gtx.Ops)
		paint.Fill(gtx.Ops, borderCol)
		border.Pop()
		// Inner fill (1px inset)
		inner := clip.UniformRRect(image.Rectangle{
			Min: image.Point{X: 1, Y: 1},
			Max: image.Point{X: pW - 1, Y: pH - 1},
		}, pH/2).Push(gtx.Ops)
		paint.Fill(gtx.Ops, bgCol)
		inner.Pop()

		gtxBg := gtx
		gtxBg.Constraints = layout.Exact(image.Point{X: pW, Y: pH})
		return layout.Center.Layout(gtxBg, func(gtx layout.Context) layout.Dimensions {
			lbl := material.Label(wm.theme, 13, icon)
			lbl.Color = iconCol
			lbl.Font.Weight = font.Bold
			lbl.Alignment = text.Middle
			return lbl.Layout(gtx)
		})
	})
}

// readerPillBtn draws a full pill button (for Back button).
func (wm *WindowManager) readerPillBtn(gtx layout.Context, label string, btn *widget.Clickable, col color.NRGBA) layout.Dimensions {
	return btn.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		// Use contrasting color on dark backgrounds
		isDark := wm.readerBgMode != 0 && wm.readerBgMode != 7 && wm.readerBgMode != 8
		var textCol color.NRGBA
		var bgCol color.NRGBA
		if isDark {
			textCol = color.NRGBA{R: 255, G: 255, B: 255, A: 230}
			bgCol = color.NRGBA{R: 255, G: 255, B: 255, A: 28}
		} else {
			textCol = col
			bgCol = theme.WithAlpha(col, 22)
		}
		if btn.Hovered() {
			bgCol.A += 30
			textCol.A = 255
		}
		if btn.Pressed() {
			bgCol.A += 20
		}
		const pH, pW = 32, 100
		cl := clip.UniformRRect(image.Rectangle{Max: image.Point{X: pW, Y: pH}}, pH/2).Push(gtx.Ops)
		paint.Fill(gtx.Ops, bgCol)
		cl.Pop()
		gtxPillB := gtx
		gtxPillB.Constraints = layout.Exact(image.Point{X: pW, Y: pH})
		return layout.Center.Layout(
			gtxPillB,
			func(gtx layout.Context) layout.Dimensions {
				lbl := material.Label(wm.theme, 13, label)
				lbl.Color = textCol
				lbl.Font.Weight = font.SemiBold
				lbl.Alignment = text.Middle
				return lbl.Layout(gtx)
			},
		)
	})
}

func (wm *WindowManager) closeReader() {
	wm.readerActive = false
	wm.readerBook = nil
	wm.readerSession = nil
	wm.readerContent = nil
	wm.readerPage = 0
	wm.readerLoading = false
	wm.readerBgPanelOpen = false
	wm.dashboardLoaded = false
	wm.metricsLoaded = false
	wm.bookStatusLoaded = false
}

func (wm *WindowManager) openBookInReader(book *domain.Book) {
	wm.readerActive = true
	wm.readerOpenedAt = time.Now()
	wm.readerBook = book
	wm.readerContent = nil
	wm.readerPage = 0
	wm.readerLoading = false
	if wm.readerFontSize == 0 {
		wm.readerFontSize = 16
	}
	if wm.trackSvc != nil {
		go func() {
			session, err := wm.trackSvc.OpenBook(context.Background(), book.ID)
			if err != nil {
				log.Printf("[Reader] OpenBook: %v", err)
				return
			}
			wm.readerSession = session
			wm.window.Invalidate()
		}()
	}
}
