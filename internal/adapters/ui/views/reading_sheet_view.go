package views

import (
	"context"
	"fmt"
	"image"
	"image/color"
	"log"
	"os/exec"
	"runtime"
	"strings"

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

// sheetFormState holds all editors and sub-view navigation state.
type sheetFormState struct {
	summaryEditor widget.Editor
	quoteEditor   widget.Editor
	tagsEditor    widget.Editor
	ratingEditor  widget.Editor
	submitBtn     widget.Clickable // "Enregistrer la fiche"
	openBtn       widget.Clickable // "+ Nouvelle fiche"
	cancelBtn     widget.Clickable // "← Retour" from create view
	showForm      bool             // true = create view, false = list view
}

// =============================================================================
// ENTRY POINT — routes between list / create / detail
// =============================================================================

func (wm *WindowManager) drawSheetsView(gtx layout.Context) layout.Dimensions {
	if !wm.sheetsLoaded {
		wm.loadSheets()
	}
	if !wm.booksLoaded {
		wm.loadBooks()
	}

	// Ensure widget slices
	for len(wm.sheetDetailBtns) < len(wm.sheets) {
		wm.sheetDetailBtns = append(wm.sheetDetailBtns, widget.Clickable{})
	}
	for len(wm.sheetFilterBtns) < len(uniqueBookTitles(wm.sheets))+1 {
		wm.sheetFilterBtns = append(wm.sheetFilterBtns, widget.Clickable{})
	}

	// Route
	if wm.activeSheetDetail != nil {
		return wm.drawSheetDetailPage(gtx)
	}
	if wm.sheetForm.showForm {
		return wm.drawSheetCreatePage(gtx)
	}
	return wm.drawSheetListPage(gtx)
}

// =============================================================================
// SUB-VIEW 1 — LIST
// =============================================================================

func (wm *WindowManager) drawSheetListPage(gtx layout.Context) layout.Dimensions {
	return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
		// Header
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return layout.Inset{Bottom: unit.Dp(24)}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
				return layout.Flex{Axis: layout.Horizontal, Alignment: layout.Middle}.Layout(gtx,
					layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
						lbl := material.H5(wm.theme, "Fiches de lecture")
						lbl.Font.Weight = font.Bold
						return lbl.Layout(gtx)
					}),
					layout.Rigid(func(gtx layout.Context) layout.Dimensions {
						if wm.sheetForm.submitBtn.Clicked(gtx) {
							wm.sheetForm.showForm = true
							wm.selectedBookIdx = -1
							wm.sheetForm.summaryEditor.SetText("")
							wm.sheetForm.quoteEditor.SetText("")
							wm.sheetForm.tagsEditor.SetText("")
							wm.sheetForm.ratingEditor.SetText("")
						}
						return wm.drawPillButton(gtx, "+ Nouvelle fiche", &wm.sheetForm.submitBtn, theme.ColorCyberCyan)
					}),
				)
			})
		}),

		// Filter chips
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return layout.Inset{Bottom: unit.Dp(20)}.Layout(gtx, wm.drawSheetFilterBar)
		}),

		// Cards
		layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
			filtered := wm.filteredSheets()
			if len(filtered) == 0 {
				return layout.Center.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
					return layout.Flex{Axis: layout.Vertical, Alignment: layout.Middle}.Layout(gtx,
						layout.Rigid(func(gtx layout.Context) layout.Dimensions {
							drawSheetEmptyIcon(gtx, gtx.Constraints.Max.X/2-20, 20)
							return layout.Dimensions{Size: image.Point{X: gtx.Constraints.Max.X, Y: 60}}
						}),
						layout.Rigid(func(gtx layout.Context) layout.Dimensions {
							lbl := material.Label(wm.theme, 15, "Aucune fiche de lecture")
							lbl.Color = theme.WithAlpha(theme.ColorPureBlack, 100)
							lbl.Alignment = 2 // center
							return layout.Inset{Top: unit.Dp(12)}.Layout(gtx, lbl.Layout)
						}),
						layout.Rigid(func(gtx layout.Context) layout.Dimensions {
							lbl := material.Label(wm.theme, 13, "Créez votre première fiche en cliquant sur "+"Nouvelle fiche")
							lbl.Color = theme.WithAlpha(theme.ColorPureBlack, 70)
							lbl.Alignment = 2
							return layout.Inset{Top: unit.Dp(6)}.Layout(gtx, lbl.Layout)
						}),
					)
				})
			}
			var rows []layout.FlexChild
			for i, sheet := range filtered {
				s := sheet
				origIdx := sheetIndexInAll(wm.sheets, s)
				rows = append(rows, layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					for len(wm.sheetDetailBtns) <= origIdx {
						wm.sheetDetailBtns = append(wm.sheetDetailBtns, widget.Clickable{})
					}
					if wm.sheetDetailBtns[origIdx].Clicked(gtx) {
						wm.activeSheetDetail = s
					}
					_ = i
					return layout.Inset{Bottom: unit.Dp(14)}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
						return wm.drawSheetCard(gtx, s, origIdx)
					})
				}))
			}
			return layout.Flex{Axis: layout.Vertical}.Layout(gtx, rows...)
		}),
	)
}

// drawSheetCard — beautiful horizontal card with colored left accent.
func (wm *WindowManager) drawSheetCard(gtx layout.Context, s *domain.ReadingSheet, btnIdx int) layout.Dimensions {
	for len(wm.sheetDetailBtns) <= btnIdx {
		wm.sheetDetailBtns = append(wm.sheetDetailBtns, widget.Clickable{})
	}
	// Derive accent color from book index
	bookIdx := sheetBookIdx(wm.books, s.BookTitle)
	accentCol := coverPalette[bookIdx%len(coverPalette)]

	return wm.sheetDetailBtns[btnIdx].Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		const cardH = 100
		w := gtx.Constraints.Max.X

		// Card background
		bgAlpha := uint8(8)
		if wm.sheetDetailBtns[btnIdx].Hovered() {
			bgAlpha = 18
		}
		if wm.sheetDetailBtns[btnIdx].Pressed() {
			bgAlpha = 32
		}
		bg := clip.UniformRRect(image.Rectangle{Max: image.Point{X: w, Y: cardH}}, 10).Push(gtx.Ops)
		paint.Fill(gtx.Ops, theme.WithAlpha(accentCol, bgAlpha))
		bg.Pop()

		// Left accent stripe
		stripe := clip.UniformRRect(image.Rectangle{Max: image.Point{X: 4, Y: cardH}}, 2).Push(gtx.Ops)
		paint.Fill(gtx.Ops, theme.WithAlpha(accentCol, 200))
		stripe.Pop()

		// Content
		layout.Inset{Top: unit.Dp(14), Left: unit.Dp(20), Right: unit.Dp(20), Bottom: unit.Dp(14)}.Layout(gtx,
			func(gtx layout.Context) layout.Dimensions {
				return layout.Flex{Axis: layout.Horizontal, Alignment: layout.Middle}.Layout(gtx,
					// Left: mini color square
					layout.Rigid(func(gtx layout.Context) layout.Dimensions {
						sq := clip.UniformRRect(image.Rectangle{Max: image.Point{X: 36, Y: 52}}, 6).Push(gtx.Ops)
						paint.Fill(gtx.Ops, theme.WithAlpha(accentCol, 180))
						sq.Pop()
						// Book letter
						lbl := material.Label(wm.theme, 18, strings.ToUpper(s.BookTitle[:1]))
						lbl.Color = color.NRGBA{R: 255, G: 255, B: 255, A: 220}
						lbl.Font.Weight = font.Bold
						gtxSq := gtx
						gtxSq.Constraints = layout.Exact(image.Point{X: 36, Y: 52})
						layout.Center.Layout(gtxSq, lbl.Layout)
						return layout.Dimensions{Size: image.Point{X: 36, Y: 52}}
					}),
					layout.Rigid(layout.Spacer{Width: unit.Dp(16)}.Layout),
					// Center: title + summary + tags
					layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
						return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
							layout.Rigid(func(gtx layout.Context) layout.Dimensions {
								lbl := material.Label(wm.theme, 15, s.BookTitle)
								lbl.Font.Weight = font.Bold
								return layout.Inset{Bottom: unit.Dp(4)}.Layout(gtx, lbl.Layout)
							}),
							layout.Rigid(func(gtx layout.Context) layout.Dimensions {
								summary := "Aucun résumé."
								if s.Summary != "" {
									summary = s.Summary
									if len([]rune(summary)) > 70 {
										summary = string([]rune(summary)[:70]) + "…"
									}
								}
								lbl := material.Label(wm.theme, 12, summary)
								lbl.Color = theme.WithAlpha(theme.ColorPureBlack, 140)
								return lbl.Layout(gtx)
							}),
							layout.Rigid(func(gtx layout.Context) layout.Dimensions {
								if len(s.Tags) == 0 {
									return layout.Dimensions{}
								}
								tagStr := strings.Join(s.Tags, "  ·  ")
								lbl := material.Label(wm.theme, 11, tagStr)
								lbl.Color = theme.WithAlpha(accentCol, 200)
								return layout.Inset{Top: unit.Dp(4)}.Layout(gtx, lbl.Layout)
							}),
						)
					}),
					layout.Rigid(layout.Spacer{Width: unit.Dp(16)}.Layout),
					// Right: stars + arrow
					layout.Rigid(func(gtx layout.Context) layout.Dimensions {
						return layout.Flex{Axis: layout.Vertical, Alignment: layout.End}.Layout(gtx,
							layout.Rigid(func(gtx layout.Context) layout.Dimensions {
								lbl := material.Label(wm.theme, 14, sheetStarStr(s.Rating))
								lbl.Color = theme.ColorSandGold
								return layout.Inset{Bottom: unit.Dp(8)}.Layout(gtx, lbl.Layout)
							}),
							layout.Rigid(func(gtx layout.Context) layout.Dimensions {
								const aw, ah = 76, 28
								pill := clip.UniformRRect(image.Rectangle{Max: image.Pt(aw, ah)}, 14).Push(gtx.Ops)
								paint.Fill(gtx.Ops, theme.WithAlpha(accentCol, 24))
								pill.Pop()
								gtxA := gtx
								gtxA.Constraints = layout.Exact(image.Pt(aw, ah))
								layout.Center.Layout(gtxA, func(gtx layout.Context) layout.Dimensions {
									lbl := material.Label(wm.theme, 12, "Ouvrir  →")
									lbl.Color = accentCol
									lbl.Font.Weight = font.SemiBold
									return lbl.Layout(gtx)
								})
								return layout.Dimensions{Size: image.Pt(aw, ah)}
							}),
						)
					}),
				)
			})
		return layout.Dimensions{Size: image.Point{X: w, Y: cardH}}
	})
}

// =============================================================================
// SUB-VIEW 2 — CREATE
// =============================================================================

func (wm *WindowManager) drawSheetCreatePage(gtx layout.Context) layout.Dimensions {
	return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
		// Top bar: back + title
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return layout.Inset{Bottom: unit.Dp(28)}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
				return layout.Flex{Axis: layout.Horizontal, Alignment: layout.Middle}.Layout(gtx,
					layout.Rigid(func(gtx layout.Context) layout.Dimensions {
						if wm.sheetForm.cancelBtn.Clicked(gtx) {
							wm.sheetForm.showForm = false
						}
						return wm.drawPillButton(gtx, "← Retour", &wm.sheetForm.cancelBtn,
							theme.WithAlpha(theme.ColorCyberCyan, 180))
					}),
					layout.Rigid(layout.Spacer{Width: unit.Dp(20)}.Layout),
					layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
						lbl := material.H5(wm.theme, "Nouvelle fiche de lecture")
						lbl.Font.Weight = font.Bold
						return lbl.Layout(gtx)
					}),
				)
			})
		}),

		// Section: Choisir un livre
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return layout.Inset{Bottom: unit.Dp(24)}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
				return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
					layout.Rigid(func(gtx layout.Context) layout.Dimensions {
						lbl := material.Label(wm.theme, 11, "CHOISIR UN LIVRE")
						lbl.Font.Weight = font.Bold
						lbl.Color = theme.WithAlpha(theme.ColorPureBlack, 110)
						return layout.Inset{Bottom: unit.Dp(12)}.Layout(gtx, lbl.Layout)
					}),
					layout.Rigid(func(gtx layout.Context) layout.Dimensions {
						return wm.drawBookPickerGrid(gtx)
					}),
				)
			})
		}),

		// Section: Résumé
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return layout.Inset{Bottom: unit.Dp(18)}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
				return wm.drawSheetFieldArea(gtx, "RÉSUMÉ", &wm.sheetForm.summaryEditor,
					"Écrivez un résumé de ce que vous avez retenu…", 120)
			})
		}),

		// Section: Citation
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return layout.Inset{Bottom: unit.Dp(18)}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
				return wm.drawSheetFieldArea(gtx, "CITATION CLÉ", &wm.sheetForm.quoteEditor,
					"La phrase qui vous a le plus marqué…", 72)
			})
		}),

		// Section: Tags + Note side by side
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return layout.Inset{Bottom: unit.Dp(28)}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
				return layout.Flex{Axis: layout.Horizontal, Alignment: layout.Start}.Layout(gtx,
					layout.Flexed(3, func(gtx layout.Context) layout.Dimensions {
						return layout.Inset{Right: unit.Dp(16)}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
							return wm.drawSheetFieldArea(gtx, "TAGS (séparés par virgule)",
								&wm.sheetForm.tagsEditor, "roman, histoire, philosophie…", 48)
						})
					}),
					layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
						return wm.drawSheetFieldArea(gtx, "NOTE /5",
							&wm.sheetForm.ratingEditor, "0–5", 48)
					}),
				)
			})
		}),

		// Submit button
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			if wm.sheetForm.submitBtn.Clicked(gtx) {
				wm.submitNewSheet()
			}
			return wm.drawGlowCTA(gtx, "✓  Enregistrer la fiche", &wm.sheetForm.submitBtn, theme.ColorCyberCyan)
		}),
	)
}

// drawBookPickerGrid — compact WOW horizontal carousel with search.
// Single row, horizontally scrollable, gold glow on selection.
func (wm *WindowManager) drawBookPickerGrid(gtx layout.Context) layout.Dimensions {
	if len(wm.books) == 0 {
		lbl := material.Label(wm.theme, 13, "Aucun livre importé.")
		lbl.Color = theme.WithAlpha(theme.ColorPureBlack, 120)
		return lbl.Layout(gtx)
	}

	return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
		// ── Search bar ─────────────────────────────────────────────────────────
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return layout.Inset{Bottom: unit.Dp(12)}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
				w := gtx.Constraints.Max.X
				const h = 36
				sbBg := clip.UniformRRect(image.Rectangle{Max: image.Pt(w, h)}, 18).Push(gtx.Ops)
				paint.Fill(gtx.Ops, color.NRGBA{R: 232, G: 232, B: 238, A: 255})
				sbBg.Pop()
				gtxSB := gtx
				gtxSB.Constraints.Min.Y = h
				gtxSB.Constraints.Max.Y = h
				return layout.Inset{Left: unit.Dp(14), Right: unit.Dp(14), Top: unit.Dp(8), Bottom: unit.Dp(8)}.Layout(gtxSB,
					func(gtx layout.Context) layout.Dimensions {
						e := material.Editor(wm.theme, &wm.sheetPickerSearch, "🔍  Rechercher un livre…")
						e.Color = theme.ColorPureBlack
						e.HintColor = theme.WithAlpha(theme.ColorPureBlack, 100)
						return e.Layout(gtx)
					})
			})
		}),
		// ── Carousel row ────────────────────────────────────────────────────────
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			query := strings.ToLower(strings.TrimSpace(wm.sheetPickerSearch.Text()))
			var filtered []int
			for i, b := range wm.books {
				if query == "" || strings.Contains(strings.ToLower(b.Title), query) ||
					strings.Contains(strings.ToLower(b.Author), query) {
					filtered = append(filtered, i)
				}
			}
			if len(filtered) == 0 {
				lbl := material.Label(wm.theme, 12, "Aucun résultat.")
				lbl.Color = theme.WithAlpha(theme.ColorPureBlack, 100)
				return lbl.Layout(gtx)
			}

			// Build cells
			const cw, ch = 96, 132
			var cells []layout.Widget
			for _, bi := range filtered {
				bk := wm.books[bi]
				idx := bi
				for len(wm.bookSelectBtns) <= idx {
					wm.bookSelectBtns = append(wm.bookSelectBtns, widget.Clickable{})
				}
				baseCol := coverPalette[idx%len(coverPalette)]
				cells = append(cells, func(gtx layout.Context) layout.Dimensions {
					if wm.bookSelectBtns[idx].Clicked(gtx) {
						wm.selectedBookIdx = idx
					}
					active := wm.selectedBookIdx == idx
					return layout.Inset{Right: unit.Dp(14)}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
						return wm.bookSelectBtns[idx].Layout(gtx, func(gtx layout.Context) layout.Dimensions {
							// Glow rings when selected
							if active {
								for gi := 3; gi >= 1; gi-- {
									e := gi * 3
									gl := clip.UniformRRect(image.Rectangle{
										Min: image.Pt(-e, -e), Max: image.Pt(cw+e, ch+e),
									}, 10+e).Push(gtx.Ops)
									paint.Fill(gtx.Ops, theme.WithAlpha(theme.ColorSandGold, uint8(44/gi)))
									gl.Pop()
								}
								ring := clip.UniformRRect(image.Rectangle{
									Min: image.Pt(-2, -2), Max: image.Pt(cw+2, ch+2),
								}, 9).Push(gtx.Ops)
								paint.Fill(gtx.Ops, theme.ColorSandGold)
								ring.Pop()
							}
							// Cover
							cvr := clip.UniformRRect(image.Rectangle{Max: image.Pt(cw, ch)}, 8).Push(gtx.Ops)
							paint.Fill(gtx.Ops, baseCol)
							cvr.Pop()
							// Spine
							sp := clip.Rect{Max: image.Pt(5, ch)}.Push(gtx.Ops)
							paint.Fill(gtx.Ops, theme.WithAlpha(baseCol, 180))
							sp.Pop()
							// Shine
							var shine clip.Path
							shine.Begin(gtx.Ops)
							shine.MoveTo(f32.Pt(float32(cw)*0.35, 0))
							shine.LineTo(f32.Pt(float32(cw)*0.65, 0))
							shine.LineTo(f32.Pt(float32(cw)*0.35, float32(ch)))
							shine.LineTo(f32.Pt(0, float32(ch)))
							shine.Close()
							sa := uint8(16)
							if active {
								sa = 32
							}
							paint.FillShape(gtx.Ops, color.NRGBA{R: 255, G: 255, B: 255, A: sa},
								clip.Outline{Path: shine.End()}.Op())
							// Hover dim
							if wm.bookSelectBtns[idx].Hovered() && !active {
								ho := clip.UniformRRect(image.Rectangle{Max: image.Pt(cw, ch)}, 8).Push(gtx.Ops)
								paint.Fill(gtx.Ops, color.NRGBA{R: 255, G: 255, B: 255, A: 25})
								ho.Pop()
							}
							// Title text on cover
							gtxT := gtx
							gtxT.Constraints = layout.Exact(image.Pt(cw, ch))
							layout.Inset{Top: unit.Dp(7), Left: unit.Dp(7), Right: unit.Dp(7)}.Layout(gtxT,
								func(gtx layout.Context) layout.Dimensions {
									lbl := material.Label(wm.theme, 9, bk.Title)
									lbl.Color = color.NRGBA{R: 255, G: 255, B: 255, A: 215}
									lbl.Font.Weight = font.SemiBold
									return lbl.Layout(gtx)
								})
							// Check badge
							if active {
								bx := clip.UniformRRect(image.Rectangle{
									Min: image.Pt(cw-20, ch-20), Max: image.Pt(cw-2, ch-2),
								}, 9).Push(gtx.Ops)
								paint.Fill(gtx.Ops, theme.ColorSandGold)
								bx.Pop()
								ts := op.Offset(image.Pt(cw-15, ch-17)).Push(gtx.Ops)
								chk := material.Label(wm.theme, 9, "✓")
								chk.Color = color.NRGBA{R: 255, G: 255, B: 255, A: 255}
								chk.Font.Weight = font.Bold
								chk.Layout(gtx)
								ts.Pop()
							}
							return layout.Dimensions{Size: image.Pt(cw, ch)}
						})
					})
				})
			}

			// Horizontal scroll wrapper — axis on widget.List, not on ListStyle
			wm.sheetPickerList.List.Axis = layout.Horizontal
			gtxCar := gtx
			gtxCar.Constraints.Max.Y = ch + 4
			gtxCar.Constraints.Min.Y = ch
			dims := material.List(wm.theme, &wm.sheetPickerList).Layout(gtxCar, len(cells), func(gtx layout.Context, i int) layout.Dimensions {
				return cells[i](gtx)
			})

			// Selected title below
			if wm.selectedBookIdx >= 0 && wm.selectedBookIdx < len(wm.books) {
				sb := wm.books[wm.selectedBookIdx]
				scol := coverPalette[wm.selectedBookIdx%len(coverPalette)]
				titleOff := op.Offset(image.Pt(0, ch+10)).Push(gtx.Ops)
				layout.Flex{Axis: layout.Horizontal, Alignment: layout.Middle}.Layout(gtx,
					layout.Rigid(func(gtx layout.Context) layout.Dimensions {
						dot := clip.UniformRRect(image.Rectangle{Max: image.Pt(8, 8)}, 4).Push(gtx.Ops)
						paint.Fill(gtx.Ops, scol)
						dot.Pop()
						return layout.Dimensions{Size: image.Pt(8, 8)}
					}),
					layout.Rigid(layout.Spacer{Width: unit.Dp(8)}.Layout),
					layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
						lbl := material.Label(wm.theme, 12, sb.Title)
						lbl.Color = theme.ColorPureBlack
						lbl.Font.Weight = font.SemiBold
						return lbl.Layout(gtx)
					}),
				)
				titleOff.Pop()
				dims.Size.Y += 24
			}
			return dims
		}),
	)
}

// drawSheetFieldArea draws a labeled text-input area.
func (wm *WindowManager) drawSheetFieldArea(gtx layout.Context, label string, ed *widget.Editor, hint string, minH int) layout.Dimensions {
	return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			lbl := material.Label(wm.theme, 11, label)
			lbl.Font.Weight = font.Bold
			lbl.Color = theme.WithAlpha(theme.ColorPureBlack, 110)
			return layout.Inset{Bottom: unit.Dp(6)}.Layout(gtx, lbl.Layout)
		}),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			w := gtx.Constraints.Max.X
			h := minH
			// Draw background
			bg := clip.UniformRRect(image.Rectangle{Max: image.Point{X: w, Y: h}}, 8).Push(gtx.Ops)
			paint.Fill(gtx.Ops, color.NRGBA{R: 245, G: 245, B: 248, A: 255})
			bg.Pop()

			// Force min height: set Min.Y so layout gives us at least h px
			gtx.Constraints.Min.Y = h
			gtx.Constraints.Max.Y = h
			gtx.Constraints.Max.X = w
			return layout.Inset{Top: unit.Dp(10), Left: unit.Dp(14), Right: unit.Dp(14), Bottom: unit.Dp(10)}.Layout(gtx,
				func(gtx layout.Context) layout.Dimensions {
					e := material.Editor(wm.theme, ed, hint)
					e.Color = theme.ColorPureBlack
					e.HintColor = theme.WithAlpha(theme.ColorPureBlack, 90)
					return e.Layout(gtx)
				})
		}),
	)
}

// =============================================================================
// SUB-VIEW 3 — DETAIL PAGE  ·  editorial redesign
// Philosophy: maximum white space, typography as the hero, colour as accent only.
// =============================================================================

func (wm *WindowManager) drawSheetDetailPage(gtx layout.Context) layout.Dimensions {
	s := wm.activeSheetDetail
	if s == nil {
		return layout.Dimensions{}
	}
	bookIdx := sheetBookIdx(wm.books, s.BookTitle)
	accentCol := coverPalette[bookIdx%len(coverPalette)]

	return material.List(wm.theme, &wm.sheetDetailScrollList).Layout(gtx, 1,
		func(gtx layout.Context, _ int) layout.Dimensions {
			W := gtx.Constraints.Max.X

			return layout.Flex{Axis: layout.Vertical}.Layout(gtx,

				// ── Thin top chrome bar (traffic-light safe) ─────────────────
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					// Hairline separator at bottom
					sep := clip.Rect{Min: image.Pt(0, 55), Max: image.Pt(W, 56)}.Push(gtx.Ops)
					paint.Fill(gtx.Ops, theme.WithAlpha(accentCol, 40))
					sep.Pop()
					return layout.Inset{
						Top: unit.Dp(12), Bottom: unit.Dp(10),
						Left: unit.Dp(110), Right: unit.Dp(32),
					}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
						return layout.Flex{Axis: layout.Horizontal, Alignment: layout.Middle}.Layout(gtx,
							layout.Rigid(func(gtx layout.Context) layout.Dimensions {
								if wm.closeSheetDetailBtn.Clicked(gtx) {
									wm.activeSheetDetail = nil
								}
								return wm.drawPillButton(gtx, "← Retour", &wm.closeSheetDetailBtn, accentCol)
							}),
						)
					})
				}),

				// ── Hero header block ─────────────────────────────────────────
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					// Full-width 3px accent bar at very top of hero
					bar := clip.Rect{Max: image.Pt(W, 3)}.Push(gtx.Ops)
					paint.Fill(gtx.Ops, accentCol)
					bar.Pop()

					return layout.Inset{
						Top: unit.Dp(56), Left: unit.Dp(80),
						Right: unit.Dp(80), Bottom: unit.Dp(52),
					}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
						return layout.Flex{Axis: layout.Vertical, Alignment: layout.Middle}.Layout(gtx,

							// Stars — large, airy
							layout.Rigid(func(gtx layout.Context) layout.Dimensions {
								if s.Rating == 0 {
									return layout.Dimensions{}
								}
								return layout.Inset{Bottom: unit.Dp(20)}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
									return layout.Center.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
										lbl := material.Label(wm.theme, unit.Sp(28), sheetStarStr(s.Rating))
										lbl.Color = theme.ColorSandGold
										return lbl.Layout(gtx)
									})
								})
							}),

							// Book title — the headline
							layout.Rigid(func(gtx layout.Context) layout.Dimensions {
								return layout.Inset{Bottom: unit.Dp(10)}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
									return layout.Center.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
										lbl := material.Label(wm.theme, unit.Sp(30), s.BookTitle)
										lbl.Font.Weight = font.Bold
										lbl.Color = theme.ColorPureBlack
										lbl.Alignment = text.Middle
										return lbl.Layout(gtx)
									})
								})
							}),

							// Rating label — muted, secondary
							layout.Rigid(func(gtx layout.Context) layout.Dimensions {
								if s.Rating == 0 {
									return layout.Dimensions{}
								}
								return layout.Center.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
									lbl := material.Label(wm.theme, unit.Sp(13), fmt.Sprintf("%d / 5", s.Rating))
									lbl.Color = theme.WithAlpha(theme.ColorPureBlack, 90)
									return lbl.Layout(gtx)
								})
							}),
						)
					})
				}),

				// ── Hairline divider ──────────────────────────────────────────
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					return layout.Inset{Left: unit.Dp(80), Right: unit.Dp(80), Bottom: unit.Dp(52)}.Layout(gtx,
						func(gtx layout.Context) layout.Dimensions {
							w := gtx.Constraints.Max.X
							line := clip.Rect{Max: image.Pt(w, 1)}.Push(gtx.Ops)
							paint.Fill(gtx.Ops, theme.WithAlpha(theme.ColorPureBlack, 14))
							line.Pop()
							return layout.Dimensions{Size: image.Pt(w, 1)}
						})
				}),

				// ── Body — centred column, max 680px reading width ────────────
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					bodyW := W - 160
					if bodyW > 680 {
						bodyW = 680
					}
					leftPad := (W - bodyW) / 2

					return layout.Inset{
						Left:   unit.Dp(float32(leftPad)),
						Right:  unit.Dp(float32(leftPad)),
						Bottom: unit.Dp(72),
					}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
						return layout.Flex{Axis: layout.Vertical}.Layout(gtx,

							// Résumé
							layout.Rigid(func(gtx layout.Context) layout.Dimensions {
								if s.Summary == "" {
									return layout.Dimensions{}
								}
								return layout.Inset{Bottom: unit.Dp(52)}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
									return wm.drawEditorialSection(gtx, "Résumé", accentCol, func(gtx layout.Context) layout.Dimensions {
										lbl := material.Label(wm.theme, unit.Sp(16), s.Summary)
										lbl.Color = theme.WithAlpha(theme.ColorPureBlack, 200)
										lbl.LineHeight = unit.Sp(26)
										return lbl.Layout(gtx)
									})
								})
							}),

							// Citations
							layout.Rigid(func(gtx layout.Context) layout.Dimensions {
								if len(s.Quotes) == 0 {
									return layout.Dimensions{}
								}
								return layout.Inset{Bottom: unit.Dp(52)}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
									return wm.drawEditorialSection(gtx, "Citations", accentCol, func(gtx layout.Context) layout.Dimensions {
										var rows []layout.FlexChild
										for _, q := range s.Quotes {
											quote := q
											rows = append(rows, layout.Rigid(func(gtx layout.Context) layout.Dimensions {
												return layout.Inset{Bottom: unit.Dp(24)}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
													w := gtx.Constraints.Max.X
													strip := clip.UniformRRect(image.Rectangle{Max: image.Pt(3, 52)}, 2).Push(gtx.Ops)
													paint.Fill(gtx.Ops, theme.WithAlpha(accentCol, 160))
													strip.Pop()
													return layout.Inset{Left: unit.Dp(18)}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
														gtx.Constraints.Max.X = w - 18
														lbl := material.Label(wm.theme, unit.Sp(16), "« "+quote+" »")
														lbl.Color = theme.WithAlpha(theme.ColorPureBlack, 165)
														lbl.LineHeight = unit.Sp(26)
														lbl.Font.Style = font.Italic
														return lbl.Layout(gtx)
													})
												})
											}))
										}
										return layout.Flex{Axis: layout.Vertical}.Layout(gtx, rows...)
									})
								})
							}),

							// Thèmes / tags
							layout.Rigid(func(gtx layout.Context) layout.Dimensions {
								if len(s.Tags) == 0 {
									return layout.Dimensions{}
								}
								return layout.Inset{Bottom: unit.Dp(52)}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
									return wm.drawEditorialSection(gtx, "Thèmes", accentCol, func(gtx layout.Context) layout.Dimensions {
										var chips []layout.FlexChild
										for _, tag := range s.Tags {
											t := tag
											chips = append(chips, layout.Rigid(func(gtx layout.Context) layout.Dimensions {
												return layout.Inset{Right: unit.Dp(10), Bottom: unit.Dp(8)}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
													const ph = 32
													gtxM := gtx
													gtxM.Constraints = layout.Constraints{Max: image.Pt(400, ph)}
													macro := op.Record(gtxM.Ops)
													lblM := material.Label(wm.theme, unit.Sp(13), "#"+t)
													txtDims := lblM.Layout(gtxM)
													macro.Stop()
													const hPad = 16
													cw := txtDims.Size.X + hPad*2

													bg := clip.UniformRRect(image.Rectangle{Max: image.Pt(cw, ph)}, ph/2).Push(gtx.Ops)
													paint.Fill(gtx.Ops, theme.WithAlpha(accentCol, 18))
													bg.Pop()

													gtxC := gtx
													gtxC.Constraints = layout.Exact(image.Pt(cw, ph))
													layout.Center.Layout(gtxC, func(gtx layout.Context) layout.Dimensions {
														lbl := material.Label(wm.theme, unit.Sp(13), "#"+t)
														lbl.Color = theme.WithAlpha(accentCol, 220)
														lbl.Font.Weight = font.Medium
														return lbl.Layout(gtx)
													})
													return layout.Dimensions{Size: image.Pt(cw, ph)}
												})
											}))
										}
										return layout.Flex{Axis: layout.Horizontal, Alignment: layout.Middle}.Layout(gtx, chips...)
									})
								})
							}),

							// Partager
							layout.Rigid(func(gtx layout.Context) layout.Dimensions {
								return wm.drawSheetSocialShare(gtx, s)
							}),
							layout.Rigid(func(gtx layout.Context) layout.Dimensions {
								if wm.shareSheetStatus == "" {
									return layout.Dimensions{}
								}
								lbl := material.Label(wm.theme, unit.Sp(12), wm.shareSheetStatus)
								lbl.Color = theme.WithAlpha(accentCol, 180)
								return layout.Inset{Top: unit.Dp(10)}.Layout(gtx, lbl.Layout)
							}),
						)
					})
				}),
			)
		})
}

// drawEditorialSection — minimal titled section, no background box.
// Small all-caps label in accent colour, content directly below.
func (wm *WindowManager) drawEditorialSection(gtx layout.Context, title string, accent color.NRGBA, content layout.Widget) layout.Dimensions {
	return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return layout.Inset{Bottom: unit.Dp(16)}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
				lbl := material.Label(wm.theme, unit.Sp(11), strings.ToUpper(title))
				lbl.Font.Weight = font.Bold
				lbl.Color = theme.WithAlpha(accent, 200)
				return lbl.Layout(gtx)
			})
		}),
		layout.Rigid(content),
	)
}

func (wm *WindowManager) drawSheetDetailView(gtx layout.Context) layout.Dimensions {
	return wm.drawSheetDetailPage(gtx)
}

// drawRichSection draws a titled content section with accent left bar and subtle bg.
func (wm *WindowManager) drawRichSection(gtx layout.Context, title string, accent, dark color.NRGBA, content layout.Widget) layout.Dimensions {
	w := gtx.Constraints.Max.X
	return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
		// Section header
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return layout.Inset{Bottom: unit.Dp(10)}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
				return layout.Flex{Axis: layout.Horizontal, Alignment: layout.Middle}.Layout(gtx,
					layout.Rigid(func(gtx layout.Context) layout.Dimensions {
						acl := clip.UniformRRect(image.Rectangle{Max: image.Pt(3, 18)}, 2).Push(gtx.Ops)
						paint.Fill(gtx.Ops, accent)
						acl.Pop()
						return layout.Dimensions{Size: image.Pt(3, 18)}
					}),
					layout.Rigid(layout.Spacer{Width: unit.Dp(10)}.Layout),
					layout.Rigid(func(gtx layout.Context) layout.Dimensions {
						lbl := material.Label(wm.theme, 11, title)
						lbl.Font.Weight = font.Bold
						lbl.Color = dark
						return lbl.Layout(gtx)
					}),
				)
			})
		}),
		// Content box
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			// Background
			bgH := gtx.Constraints.Max.Y
			if bgH == 0 {
				bgH = 200
			}
			bg := clip.UniformRRect(image.Rectangle{Max: image.Pt(w, bgH+32)}, 10).Push(gtx.Ops)
			paint.Fill(gtx.Ops, color.NRGBA{R: 248, G: 248, B: 252, A: 255})
			bg.Pop()
			// Left accent bar
			bar := clip.Rect{Max: image.Pt(4, bgH+32)}.Push(gtx.Ops)
			paint.Fill(gtx.Ops, theme.WithAlpha(accent, 80))
			bar.Pop()
			return layout.Inset{Top: unit.Dp(14), Left: unit.Dp(18), Right: unit.Dp(16), Bottom: unit.Dp(14)}.Layout(gtx, content)
		}),
	)
}

// drawSolidPillBtn draws a button with natural width, solid background and text color.
func (wm *WindowManager) drawSolidPillBtn(gtx layout.Context, label string, btn *widget.Clickable, textCol, bgCol color.NRGBA) layout.Dimensions {
	return btn.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		if btn.Hovered() {
			bgCol.A = bgCol.A + 30
		}
		if btn.Pressed() {
			bgCol.A = bgCol.A + 20
		}
		const h = 36
		const hPad = 18
		// Measure via op.Record — discard draw, keep size
		gtxM := gtx
		gtxM.Constraints = layout.Constraints{Max: image.Point{X: 2000, Y: h}}
		macro := op.Record(gtxM.Ops)
		lblM := material.Label(wm.theme, 13, label)
		lblM.Font.Weight = font.SemiBold
		textDims := lblM.Layout(gtxM)
		macro.Stop()
		w := textDims.Size.X + hPad*2
		cl := clip.UniformRRect(image.Rectangle{Max: image.Point{X: w, Y: h}}, 18).Push(gtx.Ops)
		paint.Fill(gtx.Ops, bgCol)
		cl.Pop()
		gtx2 := gtx
		gtx2.Constraints = layout.Exact(image.Pt(w, h))
		layout.Center.Layout(gtx2, func(gtx layout.Context) layout.Dimensions {
			lbl := material.Label(wm.theme, 13, label)
			lbl.Color = textCol
			lbl.Font.Weight = font.SemiBold
			return lbl.Layout(gtx)
		})
		return layout.Dimensions{Size: image.Pt(w, h)}
	})
}

// darkenColor darkens a color by the given factor (0=same, 1=black).
func darkenColor(col color.NRGBA, factor float32) color.NRGBA {
	f := 1 - factor
	return color.NRGBA{
		R: uint8(float32(col.R) * f),
		G: uint8(float32(col.G) * f),
		B: uint8(float32(col.B) * f),
		A: col.A,
	}
}

// =============================================================================
// FILTER BAR
// =============================================================================

func (wm *WindowManager) drawSheetFilterBar(gtx layout.Context) layout.Dimensions {
	titles := uniqueBookTitles(wm.sheets)
	all := append([]string{"Tous"}, titles...)
	for len(wm.sheetFilterBtns) < len(all) {
		wm.sheetFilterBtns = append(wm.sheetFilterBtns, widget.Clickable{})
	}
	var chips []layout.FlexChild
	for i, label := range all {
		idx := i
		lbl := label
		chips = append(chips, layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			if wm.sheetFilterBtns[idx].Clicked(gtx) {
				wm.sheetFilterIdx = idx
			}
			active := wm.sheetFilterIdx == idx
			accentCol := theme.ColorCyberCyan
			bgA := uint8(10)
			if active {
				bgA = 28
			}
			if wm.sheetFilterBtns[idx].Hovered() {
				bgA += 10
			}
			return layout.Inset{Right: unit.Dp(8)}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
				return wm.sheetFilterBtns[idx].Layout(gtx, func(gtx layout.Context) layout.Dimensions {
					cl := clip.UniformRRect(image.Rectangle{Max: image.Point{X: gtx.Constraints.Max.X, Y: 30}}, 15).Push(gtx.Ops)
					paint.Fill(gtx.Ops, theme.WithAlpha(accentCol, bgA))
					cl.Pop()
					return layout.Inset{Top: unit.Dp(6), Bottom: unit.Dp(6),
						Left: unit.Dp(16), Right: unit.Dp(16)}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
						l := material.Label(wm.theme, 13, lbl)
						if active {
							l.Color = theme.ColorCyberCyan
							l.Font.Weight = font.SemiBold
						} else {
							l.Color = theme.WithAlpha(theme.ColorCyberCyan, 140)
						}
						return l.Layout(gtx)
					})
				})
			})
		}))
	}
	return layout.Flex{Axis: layout.Horizontal, Alignment: layout.Middle}.Layout(gtx, chips...)
}

// =============================================================================
// SOCIAL SHARE
// =============================================================================

func (wm *WindowManager) drawSheetSocialShare(gtx layout.Context, s *domain.ReadingSheet) layout.Dimensions {
	text := buildShareText(s)
	return layout.Flex{Axis: layout.Horizontal, Alignment: layout.Middle}.Layout(gtx,
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			if wm.shareCopyBtn.Clicked(gtx) {
				if err := copyToClipboard(text); err != nil {
					wm.shareSheetStatus = "Erreur : " + err.Error()
				} else {
					wm.shareSheetStatus = "Copié dans le presse-papier !"
				}
			}
			return layout.Inset{Right: unit.Dp(12)}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
				return wm.drawPillButton(gtx, "Copier", &wm.shareCopyBtn,
					color.NRGBA{R: 80, G: 80, B: 80, A: 255})
			})
		}),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			if wm.shareTwitterBtn.Clicked(gtx) {
				openURL("https://twitter.com/intent/tweet?text=" + urlEncode(text))
				wm.shareSheetStatus = "Ouverture de Twitter / X…"
			}
			return layout.Inset{Right: unit.Dp(12)}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
				return wm.drawPillButton(gtx, "Twitter / X", &wm.shareTwitterBtn,
					color.NRGBA{R: 29, G: 155, B: 240, A: 255})
			})
		}),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			if wm.shareSheetBtn.Clicked(gtx) {
				openURL("https://www.linkedin.com/sharing/share-offsite/?url=" +
					urlEncode("https://orus.app") + "&summary=" + urlEncode(text))
				wm.shareSheetStatus = "Ouverture de LinkedIn…"
			}
			return wm.drawPillButton(gtx, "LinkedIn", &wm.shareSheetBtn,
				color.NRGBA{R: 10, G: 102, B: 194, A: 255})
		}),
	)
}

// =============================================================================
// HELPERS
// =============================================================================

func (wm *WindowManager) loadSheets() {
	if wm.sheetSvc == nil {
		wm.sheetsLoaded = true
		return
	}
	sheets, err := wm.sheetSvc.ListSheets(context.Background())
	if err != nil {
		log.Printf("[Sheets] Erreur : %v", err)
	} else {
		wm.sheets = sheets
	}
	wm.sheetsLoaded = true
}

func (wm *WindowManager) filteredSheets() []*domain.ReadingSheet {
	titles := uniqueBookTitles(wm.sheets)
	if wm.sheetFilterIdx == 0 || wm.sheetFilterIdx > len(titles) {
		return wm.sheets
	}
	target := titles[wm.sheetFilterIdx-1]
	var out []*domain.ReadingSheet
	for _, s := range wm.sheets {
		if s.BookTitle == target {
			out = append(out, s)
		}
	}
	return out
}

func (wm *WindowManager) submitNewSheet() {
	if wm.sheetSvc == nil || wm.selectedBookIdx < 0 || wm.selectedBookIdx >= len(wm.books) {
		return
	}
	book := wm.books[wm.selectedBookIdx]
	summary := strings.TrimSpace(wm.sheetForm.summaryEditor.Text())
	quoteRaw := strings.TrimSpace(wm.sheetForm.quoteEditor.Text())
	tags := splitTrim(wm.sheetForm.tagsEditor.Text(), ",")
	var quotes []string
	if quoteRaw != "" {
		quotes = []string{quoteRaw}
	}
	go func() {
		_, err := wm.sheetSvc.CreateSheet(context.Background(), book.ID, summary, 0, quotes, tags)
		if err != nil {
			log.Printf("[Sheets] CreateSheet: %v", err)
			return
		}
		wm.sheetsLoaded = false
		wm.sheetForm.showForm = false
		wm.window.Invalidate()
	}()
}

func uniqueBookTitles(sheets []*domain.ReadingSheet) []string {
	seen := map[string]bool{}
	var out []string
	for _, s := range sheets {
		if !seen[s.BookTitle] {
			seen[s.BookTitle] = true
			out = append(out, s.BookTitle)
		}
	}
	return out
}

func sheetIndexInAll(all []*domain.ReadingSheet, s *domain.ReadingSheet) int {
	for i, sh := range all {
		if sh.ID == s.ID {
			return i
		}
	}
	return 0
}

// sheetBookIdx returns the index of the book matching a title in wm.books.
func sheetBookIdx(books []*domain.Book, title string) int {
	for i, b := range books {
		if b.Title == title {
			return i
		}
	}
	return 0
}

func sheetStarStr(rating int) string {
	s := ""
	for i := 0; i < 5; i++ {
		if i < rating {
			s += "★"
		} else {
			s += "☆"
		}
	}
	return s
}

// drawSheetEmptyIcon draws a simple open-book vector shape.
func drawSheetEmptyIcon(gtx layout.Context, ox, oy int) {
	col := color.NRGBA{R: 180, G: 180, B: 195, A: 180}
	// Left page
	var lp clip.Path
	lp.Begin(gtx.Ops)
	lp.MoveTo(f32.Pt(float32(ox), float32(oy+4)))
	lp.LineTo(f32.Pt(float32(ox+22), float32(oy)))
	lp.LineTo(f32.Pt(float32(ox+22), float32(oy+32)))
	lp.LineTo(f32.Pt(float32(ox), float32(oy+36)))
	lp.Close()
	paint.FillShape(gtx.Ops, col, clip.Outline{Path: lp.End()}.Op())
	// Right page
	var rp clip.Path
	rp.Begin(gtx.Ops)
	rp.MoveTo(f32.Pt(float32(ox+24), float32(oy)))
	rp.LineTo(f32.Pt(float32(ox+46), float32(oy+4)))
	rp.LineTo(f32.Pt(float32(ox+46), float32(oy+36)))
	rp.LineTo(f32.Pt(float32(ox+24), float32(oy+32)))
	rp.Close()
	paint.FillShape(gtx.Ops, col, clip.Outline{Path: rp.End()}.Op())
}

func buildShareText(s *domain.ReadingSheet) string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Ma fiche de lecture : \"%s\"\n", s.BookTitle))
	sb.WriteString(fmt.Sprintf("Note : %s\n", sheetStarStr(s.Rating)))
	if s.Summary != "" {
		sb.WriteString(fmt.Sprintf("\nRésumé : %s\n", s.Summary))
	}
	if len(s.Quotes) > 0 {
		sb.WriteString("\nCitations :\n")
		for i, q := range s.Quotes {
			sb.WriteString(fmt.Sprintf("  %d. %s\n", i+1, q))
		}
	}
	if len(s.Tags) > 0 {
		sb.WriteString("\n#" + strings.Join(s.Tags, " #"))
	}
	sb.WriteString("\n\n#Orus #lecture #livres")
	return sb.String()
}

func copyToClipboard(text string) error {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("pbcopy")
	case "linux":
		if _, err := exec.LookPath("xclip"); err == nil {
			cmd = exec.Command("xclip", "-selection", "clipboard")
		} else {
			cmd = exec.Command("xsel", "--clipboard", "--input")
		}
	case "windows":
		cmd = exec.Command("clip")
	default:
		return fmt.Errorf("plateforme non supportée")
	}
	cmd.Stdin = strings.NewReader(text)
	return cmd.Run()
}

func openURL(url string) {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("open", url)
	case "linux":
		cmd = exec.Command("xdg-open", url)
	case "windows":
		cmd = exec.Command("rundll32", "url.dll,FileProtocolHandler", url)
	default:
		return
	}
	_ = cmd.Start()
}

func urlEncode(s string) string {
	var b strings.Builder
	for _, c := range s {
		switch {
		case c >= 'A' && c <= 'Z', c >= 'a' && c <= 'z', c >= '0' && c <= '9',
			c == '-', c == '_', c == '.', c == '~':
			b.WriteRune(c)
		default:
			for _, by := range []byte(string(c)) {
				fmt.Fprintf(&b, "%%%02X", by)
			}
		}
	}
	return b.String()
}
