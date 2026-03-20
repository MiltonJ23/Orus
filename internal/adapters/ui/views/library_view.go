package views

import (
	"context"
	"fmt"
	"image"
	"image/color"
	"log"
	"math"
	"time"

	"gioui.org/f32"
	"gioui.org/font"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/unit"
	"gioui.org/widget"
	"gioui.org/widget/material"

	"github.com/MiltonJ23/Orus/internal/adapters/ui/theme"
	"github.com/MiltonJ23/Orus/internal/domain"
)

// ── Cover palette: deep, saturated, not-too-bright ──────────────────────────
var coverPalette = []color.NRGBA{
	{R: 12, G: 105, B: 145, A: 255}, // teal foncé
	{R: 130, G: 60, B: 10, A: 255},  // brun ambré
	{R: 55, G: 90, B: 35, A: 255},   // vert forêt
	{R: 100, G: 20, B: 110, A: 255}, // violet profond
	{R: 140, G: 20, B: 40, A: 255},  // rouge bordeaux
	{R: 30, G: 50, B: 120, A: 255},  // bleu marine
}

// Lighter accent derived from base color for gradient top
func coverAccent(c color.NRGBA) color.NRGBA {
	lighten := func(v uint8, d uint8) uint8 {
		if int(v)+int(d) > 255 {
			return 255
		}
		return v + d
	}
	return color.NRGBA{R: lighten(c.R, 55), G: lighten(c.G, 55), B: lighten(c.B, 55), A: c.A}
}

const (
	coverW = 260
	coverH = 370
)

// ── Load ─────────────────────────────────────────────────────────────────────

func (wm *WindowManager) loadBooks() {
	if wm.libSvc == nil {
		log.Println("[Library] LibraryService est nil.")
		wm.booksLoaded = true
		return
	}
	books, err := wm.libSvc.GetLibrary(context.Background())
	if err != nil {
		log.Printf("[Library] Erreur : %v", err)
	} else {
		wm.books = books
	}
	wm.booksLoaded = true
}

// ── Main grid view ────────────────────────────────────────────────────────────

func (wm *WindowManager) drawBooksGrid(gtx layout.Context) layout.Dimensions {
	if !wm.booksLoaded {
		wm.loadBooks()
	}

	tabLabel := "Tous les livres"
	switch wm.activeTab {
	case 2:
		tabLabel = "À lire"
	case 3:
		tabLabel = "Terminés"
	}

	filtered := wm.filterBooksByTab(wm.books)

	// Grow stable button slices
	for len(wm.bookOpenBtns) < len(wm.books) {
		wm.bookOpenBtns = append(wm.bookOpenBtns, widget.Clickable{})
	}
	for len(wm.overlayReadBtns) < len(wm.books) {
		wm.overlayReadBtns = append(wm.overlayReadBtns, widget.Clickable{})
	}
	for len(wm.bookCoverClickBtns) < len(wm.books) {
		wm.bookCoverClickBtns = append(wm.bookCoverClickBtns, widget.Clickable{})
	}
	for len(wm.bookArchiveBtns) < len(wm.books) {
		wm.bookArchiveBtns = append(wm.bookArchiveBtns, widget.Clickable{})
	}
	for len(wm.bookDeleteBtns) < len(wm.books) {
		wm.bookDeleteBtns = append(wm.bookDeleteBtns, widget.Clickable{})
	}

	// Drive animation clock
	if wm.activeBookCardIdx >= 0 {
		elapsed := time.Since(wm.bookCardAnimStart).Seconds()
		prog := float32(elapsed / 0.38)
		if prog > 1 {
			prog = 1
		}
		wm.bookCardAnimProg = easeOutExpo(prog)
		if prog < 1 {
			wm.window.Invalidate()
		}
	}

	return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
		// Header row
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return layout.Inset{Bottom: unit.Dp(24)}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
				return layout.Flex{Axis: layout.Horizontal, Alignment: layout.Middle}.Layout(gtx,
					layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
						lbl := material.H5(wm.theme, tabLabel)
						lbl.Font.Weight = font.Bold
						return lbl.Layout(gtx)
					}),
					layout.Rigid(func(gtx layout.Context) layout.Dimensions {
						if wm.importBtn.Clicked(gtx) {
							wm.importStatusMsg = "Ouverture du sélecteur..."
							go wm.importBooksFromPicker()
						}
						return wm.drawPillButton(gtx, "+ Importer", &wm.importBtn, theme.ColorSandGold)
					}),
				)
			})
		}),
		// Status line
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			if wm.importStatusMsg == "" {
				return layout.Dimensions{}
			}
			lbl := material.Label(wm.theme, 13, wm.importStatusMsg)
			lbl.Color = theme.ColorCyberCyan
			return layout.Inset{Bottom: unit.Dp(16)}.Layout(gtx, lbl.Layout)
		}),
		// Book grid — Flexed so scroll gets full remaining height
		layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
			if len(filtered) == 0 {
				msg := "Aucun livre. Importez votre premier livre !"
				if wm.searchQuery != "" {
					msg = fmt.Sprintf("Aucun résultat pour « %s ».", wm.searchQuery)
				} else if wm.activeTab == 2 {
					msg = "Aucun livre marqué à lire."
				} else if wm.activeTab == 3 {
					msg = "Aucun livre terminé pour l'instant."
				}
				return layout.Center.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
					lbl := material.Label(wm.theme, 16, msg)
					lbl.Color = theme.WithAlpha(theme.ColorPureBlack, 110)
					return lbl.Layout(gtx)
				})
			}
			return wm.renderBookCards(gtx, filtered)
		}),
	)
}

// idxMap → stable book.ID → original wm.books index
func (wm *WindowManager) idxMap() map[string]int {
	m := make(map[string]int, len(wm.books))
	for i, b := range wm.books {
		m[b.ID] = i
	}
	return m
}

// ── Grid renderer ─────────────────────────────────────────────────────────────

func (wm *WindowManager) renderBookCards(gtx layout.Context, books []*domain.Book) layout.Dimensions {
	imap := wm.idxMap()

	// Fixed card size — always the same, window just fits more columns
	const cardW = coverW + 240 // 500px fixed
	const gap = 40
	availW := gtx.Constraints.Max.X
	cols := (availW + gap) / (cardW + gap)
	if cols < 2 {
		cols = 2
	}

	var rows [][]*domain.Book
	for i := 0; i < len(books); i += cols {
		end := i + cols
		if end > len(books) {
			end = len(books)
		}
		rows = append(rows, books[i:end])
	}

	// Capture into local to avoid closure over loop var
	dims := material.List(wm.theme, &wm.gridList).Layout(gtx, len(rows),
		func(gtx layout.Context, rowIdx int) layout.Dimensions {
			row := rows[rowIdx]
			var cells []layout.FlexChild
			for ci := range row {
				bk := row[ci]
				origIdx := imap[bk.ID]
				coverCol := coverPalette[origIdx%len(coverPalette)]
				o := origIdx   // capture for closure
				c2 := coverCol // capture for closure
				// capture
				cells = append(cells, layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					return layout.Inset{Right: unit.Dp(40), Bottom: unit.Dp(40)}.Layout(gtx,
						func(gtx layout.Context) layout.Dimensions {
							return wm.drawBookCard(gtx, bk, o, c2)
						})
				}))
			}
			return layout.Flex{Axis: layout.Horizontal, Alignment: layout.Start}.Layout(gtx, cells...)
		})

	// Action overlay drawn last (on top)
	if wm.activeBookCardIdx >= 0 {
		wm.drawBookActionOverlay(gtx, imap)
	}

	return dims
}

// ── Single book card ──────────────────────────────────────────────────────────

func (wm *WindowManager) drawBookCard(gtx layout.Context, bk *domain.Book, origIdx int, coverCol color.NRGBA) layout.Dimensions {
	// Handle cover click → open menu
	if origIdx < len(wm.bookCoverClickBtns) && wm.bookCoverClickBtns[origIdx].Clicked(gtx) {
		if wm.activeBookCardIdx == origIdx {
			wm.closeBookCardMenu()
		} else {
			wm.openBookCardMenu(origIdx)
		}
	}

	const cardW = coverW + 220
	const cardH = coverH + 24

	// Background panel: cover color at ~10% opacity
	panelCol := color.NRGBA{R: coverCol.R, G: coverCol.G, B: coverCol.B, A: 24}
	panelCl := clip.UniformRRect(image.Rectangle{Max: image.Point{X: cardW, Y: cardH}}, 14).Push(gtx.Ops)
	paint.Fill(gtx.Ops, panelCol)
	panelCl.Pop()

	// Left accent stripe in cover color
	accentCl := clip.UniformRRect(image.Rectangle{Max: image.Point{X: 4, Y: cardH}}, 2).Push(gtx.Ops)
	paint.Fill(gtx.Ops, theme.WithAlpha(coverCol, 175))
	accentCl.Pop()

	// Inner content — use real gtx so button clicks work
	gtxCard := gtx
	gtxCard.Constraints = layout.Exact(image.Point{X: cardW, Y: cardH})
	layout.Inset{Top: unit.Dp(16), Left: unit.Dp(18), Right: unit.Dp(18), Bottom: unit.Dp(16)}.Layout(
		gtxCard,
		func(gtx layout.Context) layout.Dimensions {
			return layout.Flex{Axis: layout.Horizontal, Alignment: layout.Start}.Layout(gtx,
				// Cover image (clickable)
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					if origIdx >= len(wm.bookCoverClickBtns) {
						return layout.Dimensions{}
					}
					return wm.bookCoverClickBtns[origIdx].Layout(gtx, func(gtx layout.Context) layout.Dimensions {
						return wm.drawRefinedCover(gtx, bk, origIdx, coverCol, 1.0)
					})
				}),
				layout.Rigid(layout.Spacer{Width: unit.Dp(12)}.Layout),
				// Right column
				layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
					return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
						layout.Rigid(func(gtx layout.Context) layout.Dimensions {
							lbl := material.Label(wm.theme, 13, bk.Title)
							lbl.Font.Weight = font.Bold
							lbl.Color = theme.ColorPureBlack
							return layout.Inset{Bottom: unit.Dp(5)}.Layout(gtx, lbl.Layout)
						}),
						layout.Rigid(func(gtx layout.Context) layout.Dimensions {
							a := bk.Author
							if a == "" || a == "Unknown" || a == "Unknown Author" {
								a = "Auteur inconnu"
							}
							lbl := material.Label(wm.theme, 11, a)
							lbl.Color = theme.WithAlpha(theme.ColorPureBlack, 120)
							return lbl.Layout(gtx)
						}),
						// Push Lire button to bottom-right
						layout.Flexed(1, layout.Spacer{}.Layout),
						layout.Rigid(func(gtx layout.Context) layout.Dimensions {
							if origIdx >= len(wm.bookOpenBtns) {
								return layout.Dimensions{}
							}
							if wm.bookOpenBtns[origIdx].Clicked(gtx) {
								wm.openBookInReader(bk)
							}
							return layout.Inset{Bottom: unit.Dp(6)}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
								return wm.drawGlowCTA(gtx, "▶  Lire", &wm.bookOpenBtns[origIdx], coverCol)
							})
						}),
					)
				}),
			)
		},
	)
	return layout.Dimensions{Size: image.Point{X: cardW, Y: cardH}}
}

// ── Refined cover (gradient + spine + texture) ────────────────────────────────

func (wm *WindowManager) drawRefinedCover(gtx layout.Context, bk *domain.Book, origIdx int, base color.NRGBA, scale float32) layout.Dimensions {
	w := int(float32(coverW) * scale)
	h := int(float32(coverH) * scale)
	accent := coverAccent(base)

	// Drop shadow
	sr := clip.UniformRRect(image.Rectangle{
		Min: image.Point{X: 6, Y: 9},
		Max: image.Point{X: w + 6, Y: h + 9},
	}, 10).Push(gtx.Ops)
	paint.Fill(gtx.Ops, color.NRGBA{A: 28})
	sr.Pop()

	// Base cover fill
	cr := clip.UniformRRect(image.Rectangle{Max: image.Point{X: w, Y: h}}, 8).Push(gtx.Ops)
	paint.Fill(gtx.Ops, base)
	cr.Pop()

	// Gradient top sheen: 4 horizontal bands fading from accent
	for i := 0; i < 5; i++ {
		bandH := h / 3
		y0 := i * (bandH / 5)
		y1 := y0 + bandH/3
		alpha := uint8(40 - i*7)
		if y1 > h {
			y1 = h
		}
		cl := clip.UniformRRect(image.Rectangle{
			Max: image.Point{X: w, Y: y1},
		}, 8).Push(gtx.Ops)
		_ = y0
		paint.Fill(gtx.Ops, theme.WithAlpha(accent, alpha))
		cl.Pop()
	}

	// Left spine strip
	spineW := int(float32(w) * 0.06)
	spineCol := color.NRGBA{R: base.R / 2, G: base.G / 2, B: base.B / 2, A: 200}
	scl := clip.UniformRRect(image.Rectangle{
		Max: image.Point{X: spineW, Y: h},
	}, 0).Push(gtx.Ops)
	paint.Fill(gtx.Ops, spineCol)
	scl.Pop()

	// Subtle diagonal texture lines
	for i := 0; i < 6; i++ {
		x0 := w/5 + i*(w/6)
		var lp clip.Path
		lp.Begin(gtx.Ops)
		lp.MoveTo(f32.Pt(float32(x0), float32(h)))
		lp.LineTo(f32.Pt(float32(x0+h), 0))
		lp.LineTo(f32.Pt(float32(x0+h+3), 0))
		lp.LineTo(f32.Pt(float32(x0+3), float32(h)))
		lp.Close()
		paint.FillShape(gtx.Ops,
			color.NRGBA{R: 255, G: 255, B: 255, A: 6},
			clip.Outline{Path: lp.End()}.Op())
	}

	// Title on cover
	ts := op.Offset(image.Pt(spineW+10, 18)).Push(gtx.Ops)
	g2 := gtx
	g2.Constraints.Max.X = w - spineW - 20
	tl := material.Label(wm.theme, unit.Sp(11), bk.Title)
	tl.Color = color.NRGBA{R: 255, G: 255, B: 255, A: 220}
	tl.Font.Weight = font.Bold
	tl.Layout(g2)
	ts.Pop()

	// "Appuyer pour options" hint on hover
	if origIdx < len(wm.bookCoverClickBtns) && wm.bookCoverClickBtns[origIdx].Hovered() {
		cl := clip.UniformRRect(image.Rectangle{Max: image.Point{X: w, Y: h}}, 8).Push(gtx.Ops)
		paint.Fill(gtx.Ops, color.NRGBA{A: 30})
		cl.Pop()
		// Small dot-indicator at bottom center
		drawCircle(gtx, float32(w/2), float32(h-14), 4, color.NRGBA{R: 255, G: 255, B: 255, A: 160})
		drawCircle(gtx, float32(w/2)-12, float32(h-14), 4, color.NRGBA{R: 255, G: 255, B: 255, A: 100})
		drawCircle(gtx, float32(w/2)+12, float32(h-14), 4, color.NRGBA{R: 255, G: 255, B: 255, A: 100})
	}

	return layout.Dimensions{Size: image.Point{X: w + 10, Y: h + 10}}
}

// ── Card menu ─────────────────────────────────────────────────────────────────

func (wm *WindowManager) openBookCardMenu(idx int) {
	wm.activeBookCardIdx = idx
	wm.bookCardAnimStart = time.Now()
	wm.bookCardAnimProg = 0
	wm.window.Invalidate()
}

func (wm *WindowManager) closeBookCardMenu() {
	wm.activeBookCardIdx = -1
	wm.bookCardAnimProg = 0
}

// ── Immortals-style overlay ───────────────────────────────────────────────────
// The book SLIDES to the right side of the screen.
// Everything behind blurs (simulated with layered transparency).
// Action panel rises from the left.

func (wm *WindowManager) drawBookActionOverlay(gtx layout.Context, imap map[string]int) {
	prog := wm.bookCardAnimProg // 0 → 1, eased

	// Backdrop dismiss
	if wm.bookMenuCloseBtn.Clicked(gtx) {
		wm.closeBookCardMenu()
		return
	}

	W := gtx.Constraints.Max.X
	H := gtx.Constraints.Max.Y

	// ── 1. Blur simulation: 3 translucent layers with micro-offsets ──────────
	baseAlpha := uint8(float32(175) * prog)
	wm.bookMenuCloseBtn.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		for _, off := range []image.Point{{}, {2, 1}, {-1, 2}} {
			cl := clip.Rect{
				Min: off,
				Max: image.Point{X: W + off.X, Y: H + off.Y},
			}.Push(gtx.Ops)
			paint.Fill(gtx.Ops, color.NRGBA{R: 12, G: 14, B: 22, A: baseAlpha / 3})
			cl.Pop()
		}
		return layout.Dimensions{Size: gtx.Constraints.Max}
	})

	if wm.activeBookCardIdx < 0 || wm.activeBookCardIdx >= len(wm.books) {
		return
	}

	bk := wm.books[wm.activeBookCardIdx]
	origIdx := imap[bk.ID]
	coverCol := coverPalette[origIdx%len(coverPalette)]

	// ── 2. Animated book cover — slides RIGHT with scale ─────────────────────
	const bigCoverW = int(float32(coverW) * 1.4)
	const bigCoverH = int(float32(coverH) * 1.4)

	// Target: right side, vertically centered
	targetX := W - bigCoverW - 80
	targetY := (H - bigCoverH) / 2

	// Source: screen center
	srcX := (W - bigCoverW) / 2
	srcY := (H - bigCoverH) / 2

	// Lerp + overshoot
	overshoot := float32(1.0)
	if prog < 0.7 {
		overshoot = 1.0 + 0.06*float32(math.Sin(float64(prog)*math.Pi/0.7))
	}
	coverX := srcX + int(float32(targetX-srcX)*prog)
	coverY := srcY + int(float32(targetY-srcY)*prog*0.3)
	coverScale := (1.0 + 0.4*prog) * overshoot

	cStack := op.Offset(image.Pt(coverX, coverY)).Push(gtx.Ops)
	// Glow halo around the book
	glowAlpha := uint8(float32(80) * prog)
	haloSize := int(float32(bigCoverW)*coverScale) + 40
	haloOff := op.Offset(image.Pt(-20, -20)).Push(gtx.Ops)
	for gi := 3; gi >= 0; gi-- {
		extra := gi * 12
		gcl := clip.UniformRRect(image.Rectangle{
			Min: image.Point{X: -extra, Y: -extra},
			Max: image.Point{X: haloSize + extra, Y: bigCoverH + extra*2},
		}, 20+extra).Push(gtx.Ops)
		paint.Fill(gtx.Ops, theme.WithAlpha(coverCol, glowAlpha/uint8(gi+2)))
		gcl.Pop()
	}
	_ = haloOff
	haloOff.Pop()

	// Scaled cover widget (reuse drawRefinedCover with scale)
	wm.drawRefinedCover(layout.Context{
		Ops: gtx.Ops,
		Constraints: layout.Constraints{
			Max: image.Point{X: int(float32(bigCoverW) * coverScale * 2), Y: int(float32(bigCoverH) * coverScale * 2)},
		},
		Metric: gtx.Metric,
	}, bk, origIdx, coverCol, float32(bigCoverW)/float32(coverW)*coverScale)
	cStack.Pop()

	// ── 3. Action panel — slides in from left ────────────────────────────────
	panelW := W/2 - 60
	panelH := 360

	panelTargetX := 60
	panelSrcX := -panelW - 40
	panelX := panelSrcX + int(float32(panelTargetX-panelSrcX)*prog)
	panelY := (H - panelH) / 2

	pStack := op.Offset(image.Pt(panelX, panelY)).Push(gtx.Ops)

	// Panel glass card
	panelAlpha := uint8(float32(245) * prog)
	pcl := clip.UniformRRect(image.Rectangle{Max: image.Point{X: panelW, Y: panelH}}, 18).Push(gtx.Ops)
	paint.Fill(gtx.Ops, color.NRGBA{R: 252, G: 251, B: 248, A: panelAlpha})
	pcl.Pop()
	// Subtle accent border top
	accentCl := clip.UniformRRect(image.Rectangle{
		Max: image.Point{X: panelW, Y: 4},
	}, 0).Push(gtx.Ops)
	paint.Fill(gtx.Ops, theme.WithAlpha(coverCol, panelAlpha))
	accentCl.Pop()

	// Panel content — use real gtx so all buttons work
	actionAlpha := uint8(float64(255) * math.Max(0, float64(prog-0.3)/0.7))
	gtxPanel := gtx
	gtxPanel.Constraints = layout.Exact(image.Point{X: panelW, Y: panelH})
	layout.Inset{Top: unit.Dp(32), Left: unit.Dp(36), Right: unit.Dp(36), Bottom: unit.Dp(28)}.Layout(
		gtxPanel,
		func(gtx layout.Context) layout.Dimensions {
			return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
				// Title
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					lbl := material.H5(wm.theme, bk.Title)
					lbl.Font.Weight = font.Bold
					lbl.Color = theme.WithAlpha(theme.ColorPureBlack, actionAlpha)
					return layout.Inset{Bottom: unit.Dp(4)}.Layout(gtx, lbl.Layout)
				}),
				// Author
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					a := bk.Author
					if a == "" || a == "Unknown" || a == "Unknown Author" {
						a = "Auteur inconnu"
					}
					lbl := material.Label(wm.theme, 13, a)
					lbl.Color = theme.WithAlpha(theme.ColorPureBlack, actionAlpha/2)
					return layout.Inset{Bottom: unit.Dp(32)}.Layout(gtx, lbl.Layout)
				}),
				// ▶ Lire — uses dedicated overlayReadBtns (separate from card button)
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					for len(wm.overlayReadBtns) <= origIdx {
						wm.overlayReadBtns = append(wm.overlayReadBtns, widget.Clickable{})
					}
					if wm.overlayReadBtns[origIdx].Clicked(gtx) {
						wm.closeBookCardMenu()
						wm.openBookInReader(bk)
					}
					return layout.Inset{Bottom: unit.Dp(12)}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
						return wm.overlayActionBtnIcon(gtx, "Lire ce livre", "read", &wm.overlayReadBtns[origIdx], coverCol, 255)
					})
				}),
				// Archive
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					for len(wm.bookArchiveBtns) <= origIdx {
						wm.bookArchiveBtns = append(wm.bookArchiveBtns, widget.Clickable{})
					}
					if wm.bookArchiveBtns[origIdx].Clicked(gtx) {
						go wm.libSvc.DeleteBook(context.Background(), bk.ID)
						wm.booksLoaded = false
						wm.bookStatusLoaded = false
						wm.closeBookCardMenu()
						wm.window.Invalidate()
					}
					archiveCol := color.NRGBA{R: 200, G: 120, B: 10, A: 255}
					return layout.Inset{Bottom: unit.Dp(12)}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
						return wm.overlayActionBtnIcon(gtx, "Archiver / Retirer", "archive", &wm.bookArchiveBtns[origIdx], archiveCol, 255)
					})
				}),
				// Delete
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					for len(wm.bookDeleteBtns) <= origIdx {
						wm.bookDeleteBtns = append(wm.bookDeleteBtns, widget.Clickable{})
					}
					if wm.bookDeleteBtns[origIdx].Clicked(gtx) {
						go wm.libSvc.DeleteBook(context.Background(), bk.ID)
						wm.booksLoaded = false
						wm.bookStatusLoaded = false
						wm.dashboardLoaded = false
						wm.closeBookCardMenu()
						wm.window.Invalidate()
					}
					deleteCol := color.NRGBA{R: 180, G: 40, B: 40, A: 255}
					return layout.Inset{Bottom: unit.Dp(22)}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
						return wm.overlayActionBtnIcon(gtx, "Supprimer définitivement", "delete", &wm.bookDeleteBtns[origIdx], deleteCol, 255)
					})
				}),
				// Cancel
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					if wm.bookMenuCloseBtn.Clicked(gtx) {
						wm.closeBookCardMenu()
					}
					lbl := material.Label(wm.theme, 12, "Annuler  ·  Échap")
					lbl.Color = theme.ColorCyberCyan
					return lbl.Layout(gtx)
				}),
			)
		},
	)
	pStack.Pop()
}

// ── Vector icons for overlay buttons ─────────────────────────────────────────

func drawIconPlay(gtx layout.Context, ox, oy int, col color.NRGBA) {
	var p clip.Path
	p.Begin(gtx.Ops)
	p.MoveTo(f32.Pt(float32(ox+2), float32(oy)))
	p.LineTo(f32.Pt(float32(ox+14), float32(oy+8)))
	p.LineTo(f32.Pt(float32(ox+2), float32(oy+16)))
	p.Close()
	paint.FillShape(gtx.Ops, col, clip.Outline{Path: p.End()}.Op())
}

func drawIconArchive(gtx layout.Context, ox, oy int, col color.NRGBA) {
	// lid rect
	cl := clip.Rect{Min: image.Pt(ox, oy), Max: image.Pt(ox+16, oy+4)}.Push(gtx.Ops)
	paint.Fill(gtx.Ops, col)
	cl.Pop()
	// body
	var p clip.Path
	p.Begin(gtx.Ops)
	p.MoveTo(f32.Pt(float32(ox+1), float32(oy+4)))
	p.LineTo(f32.Pt(float32(ox+15), float32(oy+4)))
	p.LineTo(f32.Pt(float32(ox+14), float32(oy+16)))
	p.LineTo(f32.Pt(float32(ox+2), float32(oy+16)))
	p.Close()
	paint.FillShape(gtx.Ops, col, clip.Outline{Path: p.End()}.Op())
	// inner slot
	slot := clip.Rect{Min: image.Pt(ox+4, oy+7), Max: image.Pt(ox+12, oy+9)}.Push(gtx.Ops)
	paint.Fill(gtx.Ops, color.NRGBA{R: 255, G: 255, B: 255, A: 120})
	slot.Pop()
}

func drawIconTrash(gtx layout.Context, ox, oy int, col color.NRGBA) {
	// handle
	handle := clip.UniformRRect(image.Rectangle{
		Min: image.Pt(ox+5, oy), Max: image.Pt(ox+11, oy+3)}, 1).Push(gtx.Ops)
	paint.Fill(gtx.Ops, col)
	handle.Pop()
	// lid bar
	lid := clip.Rect{Min: image.Pt(ox+1, oy+3), Max: image.Pt(ox+15, oy+6)}.Push(gtx.Ops)
	paint.Fill(gtx.Ops, col)
	lid.Pop()
	// body
	var p clip.Path
	p.Begin(gtx.Ops)
	p.MoveTo(f32.Pt(float32(ox+3), float32(oy+6)))
	p.LineTo(f32.Pt(float32(ox+4), float32(oy+16)))
	p.LineTo(f32.Pt(float32(ox+12), float32(oy+16)))
	p.LineTo(f32.Pt(float32(ox+13), float32(oy+6)))
	p.Close()
	paint.FillShape(gtx.Ops, col, clip.Outline{Path: p.End()}.Op())
}

// overlayActionBtn is kept for any existing call site.
func (wm *WindowManager) overlayActionBtn(gtx layout.Context, label string, btn *widget.Clickable, col color.NRGBA, alpha uint8) layout.Dimensions {
	return wm.overlayActionBtnIcon(gtx, label, "", btn, col, alpha)
}

// overlayActionBtnIcon draws a pill action button with a vector icon on the left.
func (wm *WindowManager) overlayActionBtnIcon(gtx layout.Context, label, iconType string, btn *widget.Clickable, col color.NRGBA, alpha uint8) layout.Dimensions {
	return btn.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		const bh = 52
		bw := gtx.Constraints.Max.X
		if bw <= 0 {
			bw = 280
		}
		// Shadow
		sh := clip.UniformRRect(image.Rectangle{
			Min: image.Pt(2, 3), Max: image.Pt(bw+2, bh+3),
		}, 26).Push(gtx.Ops)
		paint.Fill(gtx.Ops, color.NRGBA{A: 20})
		sh.Pop()
		// Solid pill body — readable on any background
		bgAlpha := uint8(200)
		if btn.Hovered() {
			bgAlpha = 230
		}
		if btn.Pressed() {
			bgAlpha = 170
		}
		cl := clip.UniformRRect(image.Rectangle{Max: image.Point{X: bw, Y: bh}}, 26).Push(gtx.Ops)
		paint.Fill(gtx.Ops, theme.WithAlpha(col, bgAlpha))
		cl.Pop()
		// Top shine
		shine := clip.UniformRRect(image.Rectangle{Max: image.Pt(bw, bh/2)}, 26).Push(gtx.Ops)
		paint.Fill(gtx.Ops, color.NRGBA{R: 255, G: 255, B: 255, A: 22})
		shine.Pop()

		iconCol := color.NRGBA{R: 255, G: 255, B: 255, A: 240}
		const iconOx = 18
		const iconOy = (bh - 16) / 2
		switch iconType {
		case "read":
			drawIconPlay(gtx, iconOx, iconOy, iconCol)
		case "archive":
			drawIconArchive(gtx, iconOx, iconOy, iconCol)
		case "delete":
			drawIconTrash(gtx, iconOx, iconOy, iconCol)
		}

		labelLeft := unit.Dp(20)
		if iconType != "" {
			labelLeft = unit.Dp(46)
		}
		gtx2 := gtx
		gtx2.Constraints = layout.Exact(image.Point{X: bw, Y: bh})
		return layout.Inset{Left: labelLeft, Top: unit.Dp(float32(bh)/2 - 9)}.Layout(gtx2,
			func(gtx layout.Context) layout.Dimensions {
				lbl := material.Label(wm.theme, 14, label)
				lbl.Color = color.NRGBA{R: 255, G: 255, B: 255, A: 255}
				lbl.Font.Weight = font.SemiBold
				return lbl.Layout(gtx)
			})
	})
}

// ── Easing ────────────────────────────────────────────────────────────────────

func easeOutExpo(t float32) float32 {
	if t >= 1 {
		return 1
	}
	return float32(1 - math.Pow(2, float64(-10*t)))
}

func easeOutCubic(t float32) float32 {
	t = 1 - t
	return float32(1 - math.Pow(float64(t), 3))
}
