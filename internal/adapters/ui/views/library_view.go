package views

import (
	"context"
	"image"
	"image/color"
	"log"

	"gioui.org/font"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/widget"
	"gioui.org/widget/material"

	"github.com/MiltonJ23/Orus/internal/adapters/ui/theme"
	"github.com/MiltonJ23/Orus/internal/domain"
)

// coverPalette — a set of rich cover colors cycling through the grid
var coverPalette = []color.NRGBA{
	{R: 0, G: 188, B: 212, A: 255},  // cyan
	{R: 255, G: 193, B: 7, A: 255},  // amber
	{R: 76, G: 175, B: 80, A: 255},  // green
	{R: 156, G: 39, B: 176, A: 255}, // purple
	{R: 233, G: 30, B: 99, A: 255},  // pink
	{R: 63, G: 81, B: 181, A: 255},  // indigo
}

// loadBooks fetches books from the SQLite store via LibraryService.
func (wm *WindowManager) loadBooks() {
	if wm.libSvc == nil {
		log.Println("[Library] LibraryService est nil.")
		wm.booksLoaded = true
		return
	}
	books, err := wm.libSvc.GetLibrary(context.Background())
	if err != nil {
		log.Printf("[Library] Erreur chargement : %v", err)
		wm.books = nil
	} else {
		wm.books = books
		// Sync book selector buttons for the sheet form
		wm.bookSelectBtns = make([]widget.Clickable, len(books))
	}
	wm.booksLoaded = true
}

// drawBooksGrid renders the book grid with beautiful glow CTA buttons.
func (wm *WindowManager) drawBooksGrid(gtx layout.Context) layout.Dimensions {
	if !wm.booksLoaded {
		wm.loadBooks()
	}

	if len(wm.books) == 0 {
		return layout.Center.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
			lbl := material.Label(wm.theme, 16, "Aucun livre trouve. Importez votre premier livre !")
			lbl.Color = theme.WithAlpha(theme.ColorPureBlack, 130)
			return lbl.Layout(gtx)
		})
	}

	// Ensure button slices match book count
	for len(wm.bookOpenBtns) < len(wm.books) {
		wm.bookOpenBtns = append(wm.bookOpenBtns, widget.Clickable{})
	}

	var gridElements []layout.FlexChild

	for i, book := range wm.books {
		idx := i
		bk := book
		coverCol := coverPalette[idx%len(coverPalette)]
		btnCol := coverCol
		btnCol.A = 255

		gridElements = append(gridElements, layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			// Handle click — open in reader
			if idx < len(wm.bookOpenBtns) && wm.bookOpenBtns[idx].Clicked(gtx) {
				wm.openBookInReader(bk)
			}

			return layout.Flex{Axis: layout.Vertical}.Layout(gtx,

				// ---- Cover card ----
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					const w, h = 180, 265

					// Drop shadow
					sr := clip.UniformRRect(image.Rectangle{
						Min: image.Point{X: 5, Y: 8},
						Max: image.Point{X: w + 5, Y: h + 10},
					}, 10).Push(gtx.Ops)
					paint.Fill(gtx.Ops, theme.WithAlpha(theme.ColorPureBlack, 20))
					sr.Pop()

					// Cover background
					cr := clip.UniformRRect(image.Rectangle{Max: image.Point{X: w, Y: h}}, 8).Push(gtx.Ops)
					paint.Fill(gtx.Ops, coverCol)
					cr.Pop()

					// Top sheen
					sheen := clip.UniformRRect(image.Rectangle{Max: image.Point{X: w, Y: 60}}, 8).Push(gtx.Ops)
					paint.Fill(gtx.Ops, theme.WithAlpha(theme.ColorGlassWhite, 18))
					sheen.Pop()

					// Bottom gradient strip
					botStrip := clip.UniformRRect(image.Rectangle{
						Min: image.Point{X: 0, Y: h - 70},
						Max: image.Point{X: w, Y: h},
					}, 8).Push(gtx.Ops)
					paint.Fill(gtx.Ops, theme.WithAlpha(theme.ColorPureBlack, 60))
					botStrip.Pop()

					// Title on cover
					tStack := op.Offset(image.Pt(12, 14)).Push(gtx.Ops)
					gtx2 := gtx
					gtx2.Constraints.Max.X = w - 24
					tLbl := material.Label(wm.theme, 12, bk.Title)
					tLbl.Color = theme.ColorGlassWhite
					tLbl.Font.Weight = font.Bold
					tLbl.Layout(gtx2)
					tStack.Pop()

					// Author on cover bottom
					aStack := op.Offset(image.Pt(12, h-20)).Push(gtx.Ops)
					gtx3 := gtx
					gtx3.Constraints.Max.X = w - 24
					au := bk.Author
					if au == "" {
						au = "Auteur inconnu"
					}
					aLbl := material.Label(wm.theme, 10, au)
					aLbl.Color = theme.WithAlpha(theme.ColorGlassWhite, 200)
					aLbl.Layout(gtx3)
					aStack.Pop()

					return layout.Dimensions{Size: image.Point{X: w + 12, Y: h + 12}}
				}),

				layout.Rigid(layout.Spacer{Height: 14}.Layout),

				// ---- Glow CTA "Lire" button ----
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					if idx >= len(wm.bookOpenBtns) {
						return layout.Dimensions{}
					}
					return wm.drawGlowCTA(gtx, "  Lire  ", &wm.bookOpenBtns[idx], btnCol)
				}),
			)
		}))

		gridElements = append(gridElements, layout.Rigid(layout.Spacer{Width: 32}.Layout))
	}

	return layout.Inset{Top: 8}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		return layout.Flex{Axis: layout.Horizontal, Alignment: layout.Start}.Layout(gtx, gridElements...)
	})
}

// drawSingleBook renders a compact single book card.
func (wm *WindowManager) drawSingleBook(gtx layout.Context, book *domain.Book, index int) layout.Dimensions {
	coverCol := coverPalette[index%len(coverPalette)]
	return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			const w, h = 180, 265
			sr := clip.UniformRRect(image.Rectangle{
				Min: image.Point{X: 5, Y: 8},
				Max: image.Point{X: w + 5, Y: h + 10},
			}, 10).Push(gtx.Ops)
			paint.Fill(gtx.Ops, theme.WithAlpha(theme.ColorPureBlack, 22))
			sr.Pop()
			cr := clip.UniformRRect(image.Rectangle{Max: image.Point{X: w, Y: h}}, 8).Push(gtx.Ops)
			paint.Fill(gtx.Ops, coverCol)
			cr.Pop()
			return layout.Dimensions{Size: image.Point{X: w + 12, Y: h + 12}}
		}),
		layout.Rigid(layout.Spacer{Height: 10}.Layout),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			lbl := material.Label(wm.theme, 14, book.Title)
			lbl.Color = theme.ColorPureBlack
			lbl.Font.Weight = font.Bold
			return lbl.Layout(gtx)
		}),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			author := book.Author
			if author == "" {
				author = "Auteur inconnu"
			}
			lbl := material.Label(wm.theme, 12, author)
			lbl.Color = theme.WithAlpha(theme.ColorPureBlack, 150)
			return lbl.Layout(gtx)
		}),
	)
}
