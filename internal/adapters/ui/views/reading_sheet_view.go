package views

import (
	"context"
	"fmt"
	"image"
	"log"
	"os/exec"
	"runtime"
	"strconv"
	"strings"

	"gioui.org/font"
	"gioui.org/layout"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/unit"
	"gioui.org/widget"
	"gioui.org/widget/material"

	"github.com/MiltonJ23/Orus/internal/adapters/ui/theme"
	"github.com/MiltonJ23/Orus/internal/domain"
)

// sheetFormState — state for the "new sheet" creation form
type sheetFormState struct {
	bookIDEditor   widget.Editor
	summaryEditor  widget.Editor
	quoteEditor    widget.Editor
	tagEditor      widget.Editor
	ratingBtns     [5]widget.Clickable
	selectedRating int
	addQuoteBtn    widget.Clickable
	saveBtn        widget.Clickable
	showForm       bool
	newSheetBtn    widget.Clickable
	pendingQuotes  []string
	statusMsg      string
}

// ======================================================
// MAIN SHEETS VIEW
// ======================================================

func (wm *WindowManager) drawSheetsView(gtx layout.Context) layout.Dimensions {
	// Load sheets
	if !wm.sheetsLoaded && wm.sheetSvc != nil {
		sheets, err := wm.sheetSvc.ListSheets(context.Background())
		if err != nil {
			log.Printf("[Sheets] Erreur : %v", err)
		} else {
			wm.sheets = sheets
			// Build filter buttons (one per unique book title)
			wm.rebuildSheetFilterBtns()
		}
		wm.sheetsLoaded = true
	}
	// Ensure books are loaded for the selector
	if !wm.booksLoaded {
		wm.loadBooks()
	}

	return layout.Flex{Axis: layout.Vertical}.Layout(gtx,

		// Header row
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return layout.Flex{Axis: layout.Horizontal, Alignment: layout.Middle}.Layout(gtx,
				layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
					lbl := material.H5(wm.theme, "Fiches de lecture")
					lbl.Font.Weight = font.Bold
					lbl.Color = theme.ColorPureBlack
					return lbl.Layout(gtx)
				}),
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					if wm.sheetForm.newSheetBtn.Clicked(gtx) {
						wm.sheetForm.showForm = !wm.sheetForm.showForm
						wm.sheetForm.statusMsg = ""
					}
					label := "+ Nouvelle fiche"
					if wm.sheetForm.showForm {
						label = "Annuler"
					}
					return wm.drawPillButton(gtx, label, &wm.sheetForm.newSheetBtn, theme.ColorCyberCyan)
				}),
			)
		}),

		layout.Rigid(layout.Spacer{Height: unit.Dp(20)}.Layout),

		// Filter bar — one chip per book that has a sheet
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return wm.drawSheetFilterBar(gtx)
		}),

		layout.Rigid(layout.Spacer{Height: unit.Dp(20)}.Layout),

		// Creation form
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			if !wm.sheetForm.showForm {
				return layout.Dimensions{}
			}
			return wm.drawSheetForm(gtx)
		}),

		// Sheet list (filtered)
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			filtered := wm.filteredSheets()
			if len(filtered) == 0 {
				lbl := material.Label(wm.theme, 15, "Aucune fiche. Creez-en une apres avoir fini un livre !")
				lbl.Color = theme.WithAlpha(theme.ColorPureBlack, 130)
				return layout.Center.Layout(gtx, lbl.Layout)
			}
			return wm.drawSheetList(gtx, filtered)
		}),
	)
}

// ======================================================
// FILTER BAR
// ======================================================

func (wm *WindowManager) rebuildSheetFilterBtns() {
	seen := map[string]bool{"Tous": true}
	titles := []string{"Tous"}
	for _, s := range wm.sheets {
		if !seen[s.BookTitle] {
			seen[s.BookTitle] = true
			titles = append(titles, s.BookTitle)
		}
	}
	// Grow button slice if needed
	for len(wm.sheetFilterBtns) < len(titles) {
		wm.sheetFilterBtns = append(wm.sheetFilterBtns, widget.Clickable{})
	}
}

func (wm *WindowManager) filterTitles() []string {
	seen := map[string]bool{"Tous": true}
	titles := []string{"Tous"}
	for _, s := range wm.sheets {
		if !seen[s.BookTitle] {
			seen[s.BookTitle] = true
			titles = append(titles, s.BookTitle)
		}
	}
	return titles
}

func (wm *WindowManager) filteredSheets() []*domain.ReadingSheet {
	titles := wm.filterTitles()
	if wm.sheetFilterIdx <= 0 || wm.sheetFilterIdx >= len(titles) {
		return wm.sheets
	}
	target := titles[wm.sheetFilterIdx]
	var result []*domain.ReadingSheet
	for _, s := range wm.sheets {
		if s.BookTitle == target {
			result = append(result, s)
		}
	}
	return result
}

func (wm *WindowManager) drawSheetFilterBar(gtx layout.Context) layout.Dimensions {
	titles := wm.filterTitles()
	if len(titles) <= 1 {
		return layout.Dimensions{}
	}

	var chips []layout.FlexChild
	for i, title := range titles {
		idx := i
		t := title
		if idx < len(wm.sheetFilterBtns) && wm.sheetFilterBtns[idx].Clicked(gtx) {
			wm.sheetFilterIdx = idx
		}
		isActive := wm.sheetFilterIdx == idx

		chips = append(chips, layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			bgA := uint8(12)
			textCol := theme.WithAlpha(theme.ColorPureBlack, 180)
			if isActive {
				bgA = 55
				textCol = theme.ColorCyberCyan
			}

			if idx < len(wm.sheetFilterBtns) {
				return wm.sheetFilterBtns[idx].Layout(gtx, func(gtx layout.Context) layout.Dimensions {
					cl := clip.UniformRRect(image.Rectangle{
						Max: image.Point{X: gtx.Constraints.Max.X, Y: 32},
					}, 16).Push(gtx.Ops)
					paint.Fill(gtx.Ops, theme.WithAlpha(theme.ColorCyberCyan, bgA))
					cl.Pop()
					return layout.Inset{
						Top: unit.Dp(7), Bottom: unit.Dp(7),
						Left: unit.Dp(16), Right: unit.Dp(16),
					}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
						lbl := material.Label(wm.theme, 13, t)
						lbl.Color = textCol
						if isActive {
							lbl.Font.Weight = font.Bold
						}
						return lbl.Layout(gtx)
					})
				})
			}
			return layout.Dimensions{}
		}))
		chips = append(chips, layout.Rigid(layout.Spacer{Width: unit.Dp(8)}.Layout))
	}
	return layout.Flex{Axis: layout.Horizontal, Alignment: layout.Middle}.Layout(gtx, chips...)
}

// ======================================================
// CREATION FORM
// ======================================================

func (wm *WindowManager) drawSheetForm(gtx layout.Context) layout.Dimensions {
	cl := clip.UniformRRect(image.Rectangle{Max: image.Point{X: gtx.Constraints.Max.X, Y: 540}}, 10).Push(gtx.Ops)
	paint.Fill(gtx.Ops, theme.WithAlpha(theme.ColorSandGold, 12))
	cl.Pop()

	return layout.Inset{
		Top:    unit.Dp(16),
		Bottom: unit.Dp(24),
		Left:   unit.Dp(24),
		Right:  unit.Dp(24),
	}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		return layout.Flex{Axis: layout.Vertical}.Layout(gtx,

			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				lbl := material.Label(wm.theme, 14, "Nouvelle fiche de lecture")
				lbl.Font.Weight = font.Bold
				return layout.Inset{Bottom: unit.Dp(16)}.Layout(gtx, lbl.Layout)
			}),

			// Book selector (visual buttons, no ID needed)
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				return wm.drawBookSelector(gtx)
			}),
			layout.Rigid(layout.Spacer{Height: unit.Dp(12)}.Layout),

			// Summary
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				wm.sheetForm.summaryEditor.SingleLine = false
				return wm.drawLabeledField(gtx, "Mon resume personnel", &wm.sheetForm.summaryEditor, "Ce que j'ai retenu du livre...")
			}),
			layout.Rigid(layout.Spacer{Height: unit.Dp(12)}.Layout),

			// Star rating
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				return layout.Flex{Axis: layout.Horizontal, Alignment: layout.Middle}.Layout(gtx,
					layout.Rigid(func(gtx layout.Context) layout.Dimensions {
						lbl := material.Label(wm.theme, 13, "Ma note : ")
						lbl.Color = theme.WithAlpha(theme.ColorPureBlack, 170)
						return lbl.Layout(gtx)
					}),
					layout.Rigid(layout.Spacer{Width: unit.Dp(8)}.Layout),
					layout.Rigid(func(gtx layout.Context) layout.Dimensions {
						return wm.drawStarRating(gtx, &wm.sheetForm)
					}),
				)
			}),
			layout.Rigid(layout.Spacer{Height: unit.Dp(12)}.Layout),

			// Quote input
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				return layout.Flex{Axis: layout.Horizontal, Alignment: layout.Middle}.Layout(gtx,
					layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
						return wm.drawLabeledField(gtx, "Ajouter une citation", &wm.sheetForm.quoteEditor, "La citation...")
					}),
					layout.Rigid(layout.Spacer{Width: unit.Dp(8)}.Layout),
					layout.Rigid(func(gtx layout.Context) layout.Dimensions {
						if wm.sheetForm.addQuoteBtn.Clicked(gtx) {
							q := wm.sheetForm.quoteEditor.Text()
							if q != "" {
								wm.sheetForm.pendingQuotes = append(wm.sheetForm.pendingQuotes, q)
								wm.sheetForm.quoteEditor.SetText("")
							}
						}
						return wm.drawPillButton(gtx, "+ Ajouter", &wm.sheetForm.addQuoteBtn, theme.ColorCyberCyan)
					}),
				)
			}),

			// Pending quotes preview
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				if len(wm.sheetForm.pendingQuotes) == 0 {
					return layout.Dimensions{}
				}
				var rows []layout.FlexChild
				for _, q := range wm.sheetForm.pendingQuotes {
					quote := q
					rows = append(rows, layout.Rigid(func(gtx layout.Context) layout.Dimensions {
						lbl := material.Label(wm.theme, 12, "  \""+quote+"\"")
						lbl.Color = theme.ColorCyberCyan
						return layout.Inset{Top: unit.Dp(4)}.Layout(gtx, lbl.Layout)
					}))
				}
				return layout.Inset{Top: unit.Dp(8)}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
					return layout.Flex{Axis: layout.Vertical}.Layout(gtx, rows...)
				})
			}),
			layout.Rigid(layout.Spacer{Height: unit.Dp(12)}.Layout),

			// Tags
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				return wm.drawLabeledField(gtx, "Tags (separes par des virgules)", &wm.sheetForm.tagEditor, "roman, histoire, incontournable...")
			}),
			layout.Rigid(layout.Spacer{Height: unit.Dp(20)}.Layout),

			// Status message
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				if wm.sheetForm.statusMsg == "" {
					return layout.Dimensions{}
				}
				lbl := material.Label(wm.theme, 13, wm.sheetForm.statusMsg)
				lbl.Color = theme.ColorSandGold
				return layout.Inset{Bottom: unit.Dp(8)}.Layout(gtx, lbl.Layout)
			}),

			// Save button
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				if wm.sheetForm.saveBtn.Clicked(gtx) {
					wm.submitSheetForm()
				}
				return wm.drawGlowCTA(gtx, "Enregistrer la fiche", &wm.sheetForm.saveBtn, theme.ColorSandGold)
			}),
		)
	})
}

// drawBookSelector shows library books as clickable chips — no ID typing needed.
func (wm *WindowManager) drawBookSelector(gtx layout.Context) layout.Dimensions {
	return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			lbl := material.Label(wm.theme, 12, "Livre concerne")
			lbl.Color = theme.WithAlpha(theme.ColorPureBlack, 160)
			return layout.Inset{Bottom: unit.Dp(6)}.Layout(gtx, lbl.Layout)
		}),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			if len(wm.books) == 0 {
				lbl := material.Label(wm.theme, 13, "Aucun livre dans la bibliotheque.")
				lbl.Color = theme.WithAlpha(theme.ColorPureBlack, 100)
				return lbl.Layout(gtx)
			}

			var chips []layout.FlexChild
			for i, b := range wm.books {
				idx := i
				book := b
				isSelected := wm.selectedBookIdx == idx

				chips = append(chips, layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					if idx < len(wm.bookSelectBtns) && wm.bookSelectBtns[idx].Clicked(gtx) {
						wm.selectedBookIdx = idx
						wm.sheetForm.bookIDEditor.SetText(book.ID)
					}

					bgA := uint8(12)
					textCol := theme.WithAlpha(theme.ColorPureBlack, 180)
					if isSelected {
						bgA = 60
						textCol = theme.ColorCyberCyan
					}

					if idx >= len(wm.bookSelectBtns) {
						return layout.Dimensions{}
					}

					return wm.bookSelectBtns[idx].Layout(gtx, func(gtx layout.Context) layout.Dimensions {
						cl := clip.UniformRRect(image.Rectangle{
							Max: image.Point{X: gtx.Constraints.Max.X, Y: 34},
						}, 17).Push(gtx.Ops)
						paint.Fill(gtx.Ops, theme.WithAlpha(theme.ColorCyberCyan, bgA))
						cl.Pop()
						return layout.Inset{
							Top: unit.Dp(8), Bottom: unit.Dp(8),
							Left: unit.Dp(14), Right: unit.Dp(14),
						}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
							lbl := material.Label(wm.theme, 13, book.Title)
							lbl.Color = textCol
							if isSelected {
								lbl.Font.Weight = font.Bold
							}
							return lbl.Layout(gtx)
						})
					})
				}))
				chips = append(chips, layout.Rigid(layout.Spacer{Width: unit.Dp(8)}.Layout))
			}
			return layout.Flex{Axis: layout.Horizontal}.Layout(gtx, chips...)
		}),
	)
}

func (wm *WindowManager) drawStarRating(gtx layout.Context, form *sheetFormState) layout.Dimensions {
	var children []layout.FlexChild
	for i := 0; i < 5; i++ {
		idx := i
		children = append(children, layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			if form.ratingBtns[idx].Clicked(gtx) {
				form.selectedRating = idx + 1
			}
			star := "o"
			if idx < form.selectedRating {
				star = "*"
			}
			col := theme.WithAlpha(theme.ColorPureBlack, 80)
			if idx < form.selectedRating {
				col = theme.ColorSandGold
			}
			lbl := material.Label(wm.theme, 24, star)
			lbl.Color = col
			return form.ratingBtns[idx].Layout(gtx, lbl.Layout)
		}))
		children = append(children, layout.Rigid(layout.Spacer{Width: unit.Dp(4)}.Layout))
	}
	return layout.Flex{Axis: layout.Horizontal, Alignment: layout.Middle}.Layout(gtx, children...)
}

func (wm *WindowManager) submitSheetForm() {
	if wm.sheetSvc == nil {
		wm.sheetForm.statusMsg = "Service non disponible."
		return
	}
	bookID := wm.sheetForm.bookIDEditor.Text()
	if bookID == "" && wm.selectedBookIdx >= 0 && wm.selectedBookIdx < len(wm.books) {
		bookID = wm.books[wm.selectedBookIdx].ID
	}
	if bookID == "" {
		wm.sheetForm.statusMsg = "Selectionnez un livre d'abord."
		return
	}

	tags := splitTrim(wm.sheetForm.tagEditor.Text(), ",")
	_, err := wm.sheetSvc.CreateSheet(
		context.Background(),
		bookID,
		wm.sheetForm.summaryEditor.Text(),
		wm.sheetForm.selectedRating,
		wm.sheetForm.pendingQuotes,
		tags,
	)
	if err != nil {
		wm.sheetForm.statusMsg = "Erreur : " + err.Error()
		return
	}
	// Reset form
	wm.sheetForm.bookIDEditor.SetText("")
	wm.sheetForm.summaryEditor.SetText("")
	wm.sheetForm.quoteEditor.SetText("")
	wm.sheetForm.tagEditor.SetText("")
	wm.sheetForm.pendingQuotes = nil
	wm.sheetForm.selectedRating = 0
	wm.sheetForm.showForm = false
	wm.sheetForm.statusMsg = ""
	wm.sheetsLoaded = false
	wm.selectedBookIdx = -1
}

// ======================================================
// SHEET LIST
// ======================================================

func (wm *WindowManager) drawSheetList(gtx layout.Context, sheets []*domain.ReadingSheet) layout.Dimensions {
	// Ensure one clickable per sheet in the current list
	for len(wm.sheetDetailBtns) < len(sheets) {
		wm.sheetDetailBtns = append(wm.sheetDetailBtns, widget.Clickable{})
	}

	var rows []layout.FlexChild
	for i, sheet := range sheets {
		idx := i
		s := sheet
		rows = append(rows, layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			if idx < len(wm.sheetDetailBtns) && wm.sheetDetailBtns[idx].Clicked(gtx) {
				wm.activeSheetDetail = s
			}
			return wm.drawSheetCard(gtx, s)
		}))
		rows = append(rows, layout.Rigid(layout.Spacer{Height: unit.Dp(14)}.Layout))
	}
	return layout.Flex{Axis: layout.Vertical}.Layout(gtx, rows...)
}

func (wm *WindowManager) drawSheetCard(gtx layout.Context, sheet *domain.ReadingSheet) layout.Dimensions {
	cl := clip.UniformRRect(image.Rectangle{Max: image.Point{X: gtx.Constraints.Max.X, Y: 160}}, 8).Push(gtx.Ops)
	paint.Fill(gtx.Ops, theme.WithAlpha(theme.ColorCyberCyan, 10))
	cl.Pop()

	return layout.Inset{
		Top: unit.Dp(16), Left: unit.Dp(20),
		Right: unit.Dp(20), Bottom: unit.Dp(16),
	}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
			// Header row: title + stars
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				return layout.Flex{Axis: layout.Horizontal, Alignment: layout.Middle}.Layout(gtx,
					layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
						lbl := material.Label(wm.theme, 15, sheet.BookTitle)
						lbl.Font.Weight = font.Bold
						lbl.Color = theme.ColorPureBlack
						return lbl.Layout(gtx)
					}),
					layout.Rigid(func(gtx layout.Context) layout.Dimensions {
						lbl := material.Label(wm.theme, 15, starStr(sheet.Rating))
						lbl.Color = theme.ColorSandGold
						return lbl.Layout(gtx)
					}),
				)
			}),
			layout.Rigid(layout.Spacer{Height: unit.Dp(8)}.Layout),
			// Summary preview
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				summary := sheet.Summary
				if len(summary) > 200 {
					summary = summary[:200] + "..."
				}
				if summary == "" {
					summary = "Aucun resume."
				}
				lbl := material.Label(wm.theme, 13, summary)
				lbl.Color = theme.WithAlpha(theme.ColorPureBlack, 160)
				return lbl.Layout(gtx)
			}),
			layout.Rigid(layout.Spacer{Height: unit.Dp(10)}.Layout),
			// Footer: quotes count + tags + "Ouvrir"
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				return layout.Flex{Axis: layout.Horizontal, Alignment: layout.Middle}.Layout(gtx,
					layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
						meta := fmt.Sprintf("%d citation(s)", len(sheet.Quotes))
						if len(sheet.Tags) > 0 {
							meta += "   " + strings.Join(sheet.Tags, " · ")
						}
						lbl := material.Label(wm.theme, 12, meta)
						lbl.Color = theme.ColorCyberCyan
						return lbl.Layout(gtx)
					}),
					layout.Rigid(func(gtx layout.Context) layout.Dimensions {
						lbl := material.Label(wm.theme, 12, "Ouvrir ->")
						lbl.Color = theme.WithAlpha(theme.ColorCyberCyan, 160)
						lbl.Font.Weight = font.SemiBold
						return lbl.Layout(gtx)
					}),
				)
			}),
		)
	})
}

// ======================================================
// SHEET DETAIL VIEW — full screen overlay
// ======================================================

func (wm *WindowManager) drawSheetDetailView(gtx layout.Context) layout.Dimensions {
	sheet := wm.activeSheetDetail
	if sheet == nil {
		return layout.Dimensions{}
	}

	// Background
	paint.Fill(gtx.Ops, theme.ColorGlassWhite)

	return layout.Inset{
		Top: unit.Dp(50), Bottom: unit.Dp(40),
		Left: unit.Dp(80), Right: unit.Dp(80),
	}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		return layout.Flex{Axis: layout.Vertical}.Layout(gtx,

			// Top bar: close + share
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				return layout.Flex{Axis: layout.Horizontal, Alignment: layout.Middle}.Layout(gtx,
					layout.Rigid(func(gtx layout.Context) layout.Dimensions {
						if wm.closeSheetDetailBtn.Clicked(gtx) {
							wm.activeSheetDetail = nil
							wm.shareSheetStatus = ""
						}
						return wm.closeSheetDetailBtn.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
							lbl := material.Label(wm.theme, 18, "<- Retour")
							lbl.Color = theme.ColorCyberCyan
							lbl.Font.Weight = font.SemiBold
							return lbl.Layout(gtx)
						})
					}),
					layout.Flexed(1, layout.Spacer{}.Layout),
					// Share button
					layout.Rigid(func(gtx layout.Context) layout.Dimensions {
						if wm.shareSheetBtn.Clicked(gtx) {
							wm.shareSheetToClipboard(sheet)
						}
						return wm.drawPillButton(gtx, "Copier / Partager", &wm.shareSheetBtn, theme.ColorSandGold)
					}),
				)
			}),

			// Share status
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				if wm.shareSheetStatus == "" {
					return layout.Dimensions{}
				}
				lbl := material.Label(wm.theme, 13, wm.shareSheetStatus)
				lbl.Color = theme.ColorSandGold
				return layout.Inset{Top: unit.Dp(8)}.Layout(gtx, lbl.Layout)
			}),

			layout.Rigid(layout.Spacer{Height: unit.Dp(32)}.Layout),

			// Book title heading
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				lbl := material.H4(wm.theme, sheet.BookTitle)
				lbl.Font.Weight = font.Bold
				lbl.Color = theme.ColorPureBlack
				return layout.Inset{Bottom: unit.Dp(4)}.Layout(gtx, lbl.Layout)
			}),

			// Star rating
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				if sheet.Rating == 0 {
					return layout.Dimensions{}
				}
				lbl := material.Label(wm.theme, 20, starStr(sheet.Rating))
				lbl.Color = theme.ColorSandGold
				return layout.Inset{Bottom: unit.Dp(8)}.Layout(gtx, lbl.Layout)
			}),

			// Tags
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				if len(sheet.Tags) == 0 {
					return layout.Dimensions{}
				}
				lbl := material.Label(wm.theme, 13, strings.Join(sheet.Tags, " · "))
				lbl.Color = theme.ColorCyberCyan
				return layout.Inset{Bottom: unit.Dp(24)}.Layout(gtx, lbl.Layout)
			}),

			// Summary
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				if sheet.Summary == "" {
					return layout.Dimensions{}
				}
				return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
					layout.Rigid(func(gtx layout.Context) layout.Dimensions {
						lbl := material.Label(wm.theme, 12, "RESUME PERSONNEL")
						lbl.Color = theme.WithAlpha(theme.ColorCyberCyan, 160)
						lbl.Font.Weight = font.Bold
						return layout.Inset{Bottom: unit.Dp(8)}.Layout(gtx, lbl.Layout)
					}),
					layout.Rigid(func(gtx layout.Context) layout.Dimensions {
						// Summary in a framed box
						cl := clip.UniformRRect(image.Rectangle{Max: image.Point{X: gtx.Constraints.Max.X, Y: 999}}, 8).Push(gtx.Ops)
						paint.Fill(gtx.Ops, theme.WithAlpha(theme.ColorCyberCyan, 7))
						cl.Pop()
						return layout.Inset{
							Top: unit.Dp(16), Bottom: unit.Dp(16),
							Left: unit.Dp(20), Right: unit.Dp(20),
						}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
							lbl := material.Label(wm.theme, 15, sheet.Summary)
							lbl.Color = theme.ColorPureBlack
							return lbl.Layout(gtx)
						})
					}),
					layout.Rigid(layout.Spacer{Height: unit.Dp(24)}.Layout),
				)
			}),

			// Quotes
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				if len(sheet.Quotes) == 0 {
					return layout.Dimensions{}
				}
				var rows []layout.FlexChild
				rows = append(rows, layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					lbl := material.Label(wm.theme, 12, "CITATIONS")
					lbl.Color = theme.WithAlpha(theme.ColorCyberCyan, 160)
					lbl.Font.Weight = font.Bold
					return layout.Inset{Bottom: unit.Dp(12)}.Layout(gtx, lbl.Layout)
				}))
				for i, q := range sheet.Quotes {
					quote := q
					num := strconv.Itoa(i + 1)
					rows = append(rows, layout.Rigid(func(gtx layout.Context) layout.Dimensions {
						return layout.Inset{Bottom: unit.Dp(12)}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
							return layout.Flex{Axis: layout.Horizontal}.Layout(gtx,
								layout.Rigid(func(gtx layout.Context) layout.Dimensions {
									lbl := material.Label(wm.theme, 13, num+".")
									lbl.Color = theme.ColorSandGold
									lbl.Font.Weight = font.Bold
									return layout.Inset{Right: unit.Dp(10)}.Layout(gtx, lbl.Layout)
								}),
								layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
									lbl := material.Label(wm.theme, 14, "\""+quote+"\"")
									lbl.Color = theme.WithAlpha(theme.ColorPureBlack, 180)
									return lbl.Layout(gtx)
								}),
							)
						})
					}))
				}
				return layout.Flex{Axis: layout.Vertical}.Layout(gtx, rows...)
			}),
		)
	})
}

// ======================================================
// CLIPBOARD SHARE
// ======================================================

// shareSheetToClipboard formats the sheet as text and copies to system clipboard.
func (wm *WindowManager) shareSheetToClipboard(sheet *domain.ReadingSheet) {
	if sheet == nil {
		return
	}
	var sb strings.Builder
	sb.WriteString("=== FICHE DE LECTURE ===\n\n")
	sb.WriteString("LIVRE : " + sheet.BookTitle + "\n")
	if sheet.Rating > 0 {
		sb.WriteString("NOTE  : " + sheet.StarString() + "\n")
	}
	if len(sheet.Tags) > 0 {
		sb.WriteString("TAGS  : " + strings.Join(sheet.Tags, ", ") + "\n")
	}
	sb.WriteString("\n")
	if sheet.Summary != "" {
		sb.WriteString("RESUME :\n" + sheet.Summary + "\n\n")
	}
	if len(sheet.Quotes) > 0 {
		sb.WriteString("CITATIONS FAVORITES :\n")
		for i, q := range sheet.Quotes {
			sb.WriteString(fmt.Sprintf("  %d. \"%s\"\n", i+1, q))
		}
	}
	sb.WriteString("\n— Lu avec Orus")

	content := sb.String()
	err := copyToClipboard(content)
	if err != nil {
		log.Printf("[Share] Clipboard error: %v", err)
		wm.shareSheetStatus = "Copie echouee. Verifiez votre configuration."
	} else {
		wm.shareSheetStatus = "Copie dans le presse-papier ! Collez sur vos reseaux sociaux."
	}
	wm.window.Invalidate()
}

// copyToClipboard writes text to the system clipboard via OS-native command.
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
		// On Windows, use clip
		cmd = exec.Command("clip")
	default:
		return fmt.Errorf("clipboard not supported on %s", runtime.GOOS)
	}
	cmd.Stdin = strings.NewReader(text)
	return cmd.Run()
}

// ======================================================
// HELPERS
// ======================================================

// starStr returns a string of * and o for a rating 0–5 using ASCII chars (font-safe)
func starStr(rating int) string {
	var sb strings.Builder
	for i := 1; i <= 5; i++ {
		if i <= rating {
			sb.WriteString("* ")
		} else {
			sb.WriteString("o ")
		}
	}
	return strings.TrimSpace(sb.String())
}
