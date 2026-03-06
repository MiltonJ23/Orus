package views

import (
	"context"
	"image"
	"log"

	"gioui.org/font"
	"gioui.org/layout"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/widget/material"

	"github.com/MiltonJ23/Orus/internal/adapters/ui/theme"
	"github.com/MiltonJ23/Orus/internal/domain"
)

// loadBooks appelle le service backend pour récupérer les livres de la BDD SQLite
func (wm *WindowManager) loadBooks() {
	// Protection anti-crash si le service n'est pas instancié dans le main
	if wm.libSvc == nil {
		log.Println("[Library] Avertissement : LibraryService est nil. Backend non connecté.")
		wm.books = nil
		wm.booksLoaded = true
		return
	}

	books, err := wm.libSvc.GetLibrary(context.Background())
	if err != nil {
		log.Printf("[Library] Erreur lors du chargement des livres : %v", err)
		wm.books = nil
	} else {
		wm.books = books
	}
	wm.booksLoaded = true
}

// drawBooksGrid dessine la grille dynamique avec gestion du Scroll (défilement)
func (wm *WindowManager) drawBooksGrid(gtx layout.Context) layout.Dimensions {
	// 1. On charge les livres de la BDD UNE SEULE FOIS
	if !wm.booksLoaded && wm.libSvc != nil {
		books, err := wm.libSvc.GetLibrary(context.Background())
		if err != nil {
			log.Printf("[Library] Erreur DB : %v", err)
		} else {
			wm.books = books
		}
		wm.booksLoaded = true
	}

	// S'il n'y a pas de livres
	if len(wm.books) == 0 {
		lbl := material.Label(wm.theme, 16, "Aucun livre trouvé dans la bibliothèque.")
		return layout.Center.Layout(gtx, lbl.Layout)
	}

	// 2. On dessine ta grille exactement comme tu l'avais codée
	var gridElements []layout.FlexChild

	for i, book := range wm.books {
		idx := i
		bk := book // Copie locale pour éviter les bugs de pointeurs

		gridElements = append(gridElements, layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
				// A. Couverture
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					w, h := 180, 270
					shadowRect := clip.UniformRRect(image.Rectangle{Min: image.Point{X: 4, Y: 8}, Max: image.Point{X: w + 4, Y: h + 10}}, 4).Push(gtx.Ops)
					paint.Fill(gtx.Ops, theme.WithAlpha(theme.ColorPureBlack, 25))
					shadowRect.Pop()

					coverColor := theme.WithAlpha(theme.ColorCyberCyan, 220)
					if idx%2 != 0 {
						coverColor = theme.ColorSandGold
					}

					coverRect := clip.UniformRRect(image.Rectangle{Max: image.Point{X: w, Y: h}}, 4).Push(gtx.Ops)
					paint.Fill(gtx.Ops, coverColor)
					coverRect.Pop()

					return layout.Dimensions{Size: image.Point{X: w + 10, Y: h + 12}}
				}),
				layout.Rigid(layout.Spacer{Height: 12}.Layout),

				// B. Titre du livre venant de la base de données
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					lbl := material.Label(wm.theme, 14, bk.Title) // <- VRAI TITRE ICI
					lbl.Color = theme.ColorPureBlack
					lbl.Font.Weight = font.Bold
					return lbl.Layout(gtx)
				}),
			)
		}))

		// Espace entre les livres
		gridElements = append(gridElements, layout.Rigid(layout.Spacer{Width: 30}.Layout))
	}

	// On affiche les livres horizontalement
	return layout.Flex{Axis: layout.Horizontal}.Layout(gtx, gridElements...)
}

// drawSingleBook dessine la carte premium d'un seul livre
func (wm *WindowManager) drawSingleBook(gtx layout.Context, book *domain.Book, index int) layout.Dimensions {
	return layout.Flex{Axis: layout.Vertical}.Layout(gtx,

		// 1. La Couverture
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			w, h := 180, 270

			// Ombre portée subtile
			shadowRect := clip.UniformRRect(image.Rectangle{
				Min: image.Point{X: 4, Y: 8}, Max: image.Point{X: w + 4, Y: h + 10},
			}, 4).Push(gtx.Ops)
			paint.Fill(gtx.Ops, theme.WithAlpha(theme.ColorPureBlack, 25))
			shadowRect.Pop()

			coverColor := theme.WithAlpha(theme.ColorCyberCyan, 220)
			if index%2 != 0 {
				coverColor = theme.ColorSandGold
			}

			coverRect := clip.UniformRRect(image.Rectangle{Max: image.Point{X: w, Y: h}}, 4).Push(gtx.Ops)
			paint.Fill(gtx.Ops, coverColor)
			coverRect.Pop()

			return layout.Dimensions{Size: image.Point{X: w + 10, Y: h + 12}}
		}),
		layout.Rigid(layout.Spacer{Height: 12}.Layout),

		// 2. Le Titre du livre
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			lbl := material.Label(wm.theme, 14, book.Title)
			lbl.Color = theme.ColorPureBlack
			lbl.Font.Weight = font.Bold
			return lbl.Layout(gtx)
		}),

		// 3. L'Auteur
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			authorText := book.Author
			if authorText == "" {
				authorText = "Auteur inconnu"
			}
			lbl := material.Label(wm.theme, 12, authorText)
			lbl.Color = theme.WithAlpha(theme.ColorPureBlack, 150)
			return lbl.Layout(gtx)
		}),
	)
}
