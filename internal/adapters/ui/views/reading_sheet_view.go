package views

import (
	"context"
	"image"
	"log"
	"strconv"

	"gioui.org/font"
	"gioui.org/layout"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/widget"
	"gioui.org/widget/material"

	"github.com/MiltonJ23/Orus/internal/adapters/ui/theme"
	"github.com/MiltonJ23/Orus/internal/domain"
)

// État du formulaire de création d'une fiche
type sheetFormState struct {
	bookIDEditor   widget.Editor
	summaryEditor  widget.Editor
	quoteEditor    widget.Editor
	tagEditor      widget.Editor
	ratingBtns     [5]widget.Clickable
	selectedRating int
	addQuoteBtn    widget.Clickable
	saveBtn        widget.Clickable
	cancelBtn      widget.Clickable
	showForm       bool
	newSheetBtn    widget.Clickable
	pendingQuotes  []string
	statusMsg      string
}

// drawSheetsView est l'onglet "Fiches de lecture"
func (wm *WindowManager) drawSheetsView(gtx layout.Context) layout.Dimensions {
	// Chargement paresseux
	if !wm.sheetsLoaded && wm.sheetSvc != nil {
		sheets, err := wm.sheetSvc.ListSheets(context.Background())
		if err != nil {
			log.Printf("[Sheets] Erreur chargement : %v", err)
		} else {
			wm.sheets = sheets
		}
		wm.sheetsLoaded = true
	}

	return layout.Flex{Axis: layout.Vertical}.Layout(gtx,

		// ── En-tête ──────────────────────────────────────
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return layout.Flex{Axis: layout.Horizontal, Alignment: layout.Middle}.Layout(gtx,
				layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
					lbl := material.H5(wm.theme, "📝 Fiches de lecture")
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
						label = "✕ Annuler"
					}
					return wm.drawPillButton(gtx, label, &wm.sheetForm.newSheetBtn, theme.ColorCyberCyan)
				}),
			)
		}),

		layout.Rigid(layout.Spacer{Height: 24}.Layout),

		// ── Formulaire (si ouvert) ────────────────────────
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			if !wm.sheetForm.showForm {
				return layout.Dimensions{}
			}
			return wm.drawSheetForm(gtx)
		}),

		// ── Liste des fiches ──────────────────────────────
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			if len(wm.sheets) == 0 {
				lbl := material.Label(wm.theme, 15, "Aucune fiche de lecture. Créez-en une après avoir fini un livre !")
				lbl.Color = theme.WithAlpha(theme.ColorPureBlack, 130)
				return layout.Center.Layout(gtx, lbl.Layout)
			}
			return wm.drawSheetList(gtx)
		}),
	)
}

func (wm *WindowManager) drawSheetForm(gtx layout.Context) layout.Dimensions {
	// Fond de carte
	cl := clip.UniformRRect(image.Rectangle{Max: image.Point{X: gtx.Constraints.Max.X, Y: 500}}, 10).Push(gtx.Ops)
	paint.Fill(gtx.Ops, theme.WithAlpha(theme.ColorSandGold, 15))
	cl.Pop()

	return layout.Inset{Top: 16, Bottom: 24, Left: 24, Right: 24}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		return layout.Flex{Axis: layout.Vertical}.Layout(gtx,

			// Titre formulaire
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				lbl := material.Label(wm.theme, 14, "Nouvelle fiche de lecture")
				lbl.Font.Weight = font.Bold
				return layout.Inset{Bottom: 16}.Layout(gtx, lbl.Layout)
			}),

			// ID du livre
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				return wm.drawLabeledField(gtx, "ID du livre", &wm.sheetForm.bookIDEditor, "Collez l'ID du livre...")
			}),
			layout.Rigid(layout.Spacer{Height: 12}.Layout),

			// Résumé
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				wm.sheetForm.summaryEditor.SingleLine = false
				return wm.drawLabeledField(gtx, "Mon résumé personnel", &wm.sheetForm.summaryEditor, "Ce que j'ai retenu du livre...")
			}),
			layout.Rigid(layout.Spacer{Height: 12}.Layout),

			// Note (étoiles)
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				return layout.Flex{Axis: layout.Horizontal, Alignment: layout.Middle}.Layout(gtx,
					layout.Rigid(func(gtx layout.Context) layout.Dimensions {
						lbl := material.Label(wm.theme, 13, "Ma note : ")
						lbl.Color = theme.WithAlpha(theme.ColorPureBlack, 170)
						return lbl.Layout(gtx)
					}),
					layout.Rigid(layout.Spacer{Width: 8}.Layout),
					layout.Rigid(func(gtx layout.Context) layout.Dimensions {
						return wm.drawStarRating(gtx, &wm.sheetForm)
					}),
				)
			}),
			layout.Rigid(layout.Spacer{Height: 12}.Layout),

			// Citation
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				return layout.Flex{Axis: layout.Horizontal, Alignment: layout.Middle}.Layout(gtx,
					layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
						return wm.drawLabeledField(gtx, "Ajouter une citation", &wm.sheetForm.quoteEditor, "« La citation... »")
					}),
					layout.Rigid(layout.Spacer{Width: 8}.Layout),
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

			// Affichage des citations en attente
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				if len(wm.sheetForm.pendingQuotes) == 0 {
					return layout.Dimensions{}
				}
				var rows []layout.FlexChild
				for _, q := range wm.sheetForm.pendingQuotes {
					quote := q
					rows = append(rows, layout.Rigid(func(gtx layout.Context) layout.Dimensions {
						lbl := material.Label(wm.theme, 13, "› "+quote)
						lbl.Color = theme.ColorCyberCyan
						return layout.Inset{Top: 4}.Layout(gtx, lbl.Layout)
					}))
				}
				return layout.Inset{Top: 8}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
					return layout.Flex{Axis: layout.Vertical}.Layout(gtx, rows...)
				})
			}),
			layout.Rigid(layout.Spacer{Height: 12}.Layout),

			// Tags
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				return wm.drawLabeledField(gtx, "Tags (séparés par des virgules)", &wm.sheetForm.tagEditor, "roman, histoire, incontournable...")
			}),
			layout.Rigid(layout.Spacer{Height: 20}.Layout),

			// Message de statut
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				if wm.sheetForm.statusMsg == "" {
					return layout.Dimensions{}
				}
				lbl := material.Label(wm.theme, 13, wm.sheetForm.statusMsg)
				lbl.Color = theme.ColorSandGold
				return layout.Inset{Bottom: 8}.Layout(gtx, lbl.Layout)
			}),

			// Bouton Enregistrer
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				if wm.sheetForm.saveBtn.Clicked(gtx) {
					wm.submitSheetForm()
				}
				return wm.drawPillButton(gtx, "✓ Enregistrer la fiche", &wm.sheetForm.saveBtn, theme.ColorSandGold)
			}),
		)
	})
}

func (wm *WindowManager) drawStarRating(gtx layout.Context, form *sheetFormState) layout.Dimensions {
	var children []layout.FlexChild
	for i := 0; i < 5; i++ {
		idx := i
		children = append(children, layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			if form.ratingBtns[idx].Clicked(gtx) {
				form.selectedRating = idx + 1
			}
			star := "☆"
			col := theme.WithAlpha(theme.ColorPureBlack, 80)
			if idx < form.selectedRating {
				star = "★"
				col = theme.ColorSandGold
			}
			lbl := material.Label(wm.theme, 22, star)
			lbl.Color = col
			return form.ratingBtns[idx].Layout(gtx, lbl.Layout)
		}))
		children = append(children, layout.Rigid(layout.Spacer{Width: 4}.Layout))
	}
	return layout.Flex{Axis: layout.Horizontal, Alignment: layout.Middle}.Layout(gtx, children...)
}

func (wm *WindowManager) submitSheetForm() {
	if wm.sheetSvc == nil {
		wm.sheetForm.statusMsg = "Service non disponible."
		return
	}

	bookID := wm.sheetForm.bookIDEditor.Text()
	if bookID == "" {
		wm.sheetForm.statusMsg = "L'ID du livre est requis."
		return
	}

	// Conversion des tags
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

	// Réinitialisation
	wm.sheetForm.bookIDEditor.SetText("")
	wm.sheetForm.summaryEditor.SetText("")
	wm.sheetForm.quoteEditor.SetText("")
	wm.sheetForm.tagEditor.SetText("")
	wm.sheetForm.pendingQuotes = nil
	wm.sheetForm.selectedRating = 0
	wm.sheetForm.showForm = false
	wm.sheetsLoaded = false // forcer le rechargement
	wm.sheetForm.statusMsg = ""
}

func (wm *WindowManager) drawSheetList(gtx layout.Context) layout.Dimensions {
	var rows []layout.FlexChild
	for _, sheet := range wm.sheets {
		s := sheet
		rows = append(rows, layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return wm.drawSheetCard(gtx, s)
		}))
		rows = append(rows, layout.Rigid(layout.Spacer{Height: 16}.Layout))
	}
	return layout.Flex{Axis: layout.Vertical}.Layout(gtx, rows...)
}

func (wm *WindowManager) drawSheetCard(gtx layout.Context, sheet *domain.ReadingSheet) layout.Dimensions {
	cl := clip.UniformRRect(image.Rectangle{Max: image.Point{X: gtx.Constraints.Max.X, Y: 150}}, 8).Push(gtx.Ops)
	paint.Fill(gtx.Ops, theme.WithAlpha(theme.ColorCyberCyan, 12))
	cl.Pop()

	return layout.Inset{Top: 16, Left: 20, Right: 20, Bottom: 16}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
			// Titre + étoiles
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				return layout.Flex{Axis: layout.Horizontal, Alignment: layout.Middle}.Layout(gtx,
					layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
						lbl := material.Label(wm.theme, 15, sheet.BookTitle)
						lbl.Font.Weight = font.Bold
						lbl.Color = theme.ColorPureBlack
						return lbl.Layout(gtx)
					}),
					layout.Rigid(func(gtx layout.Context) layout.Dimensions {
						lbl := material.Label(wm.theme, 16, sheet.StarString())
						lbl.Color = theme.ColorSandGold
						return lbl.Layout(gtx)
					}),
				)
			}),
			layout.Rigid(layout.Spacer{Height: 8}.Layout),
			// Résumé tronqué
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				summary := sheet.Summary
				if len(summary) > 180 {
					summary = summary[:180] + "..."
				}
				if summary == "" {
					summary = "Aucun résumé."
				}
				lbl := material.Label(wm.theme, 13, summary)
				lbl.Color = theme.WithAlpha(theme.ColorPureBlack, 160)
				return lbl.Layout(gtx)
			}),
			layout.Rigid(layout.Spacer{Height: 10}.Layout),
			// Méta : nb citations + tags
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				meta := strconv.Itoa(len(sheet.Quotes)) + " citation(s)"
				if len(sheet.Tags) > 0 {
					meta += "  •  " + joinStr(sheet.Tags, " • ")
				}
				lbl := material.Label(wm.theme, 12, meta)
				lbl.Color = theme.ColorCyberCyan
				return lbl.Layout(gtx)
			}),
		)
	})
}
