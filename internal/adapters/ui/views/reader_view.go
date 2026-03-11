package views

import (
	"context"
	"fmt"
	"image"
	"log"

	"gioui.org/font"
	"gioui.org/layout"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/text"
	"gioui.org/unit"
	"gioui.org/widget"
	"gioui.org/widget/material"

	"github.com/MiltonJ23/Orus/internal/adapters/ui/theme"
	"github.com/MiltonJ23/Orus/internal/domain"
)

func (wm *WindowManager) drawReaderView(gtx layout.Context) layout.Dimensions {
	// Chargement asynchrone du texte
	if len(wm.readerContent) == 0 && wm.readerBook != nil && !wm.readerLoading {
		wm.readerLoading = true
		go func() {
			if wm.contentReader == nil {
				wm.readerContent = []string{"Lecteur non disponible."}
				wm.readerLoading = false
				wm.window.Invalidate()
				return
			}
			chunks, err := wm.contentReader.ReadBookText(context.Background(), wm.readerBook.FilePath)
			if err != nil {
				log.Printf("[Reader] Erreur : %v", err)
				wm.readerContent = []string{fmt.Sprintf("Impossible de lire ce fichier :\n%v\n\nFormats supportés : PDF, EPUB.", err)}
			} else {
				wm.readerContent = chunks
			}
			wm.readerPage = 0
			wm.readerLoading = false
			wm.window.Invalidate()
		}()
	}

	// Fond du lecteur (blanc ou sombre selon dimming)
	if wm.readerDimAlpha > 128 {
		paint.Fill(gtx.Ops, theme.ColorVoidDark)
	} else {
		paint.Fill(gtx.Ops, theme.ColorGlassWhite)
	}

	return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
		layout.Rigid(wm.drawReaderTopBar),
		layout.Flexed(1, wm.drawReaderContent),
		layout.Rigid(wm.drawReaderBottomBar),
	)
}

func (wm *WindowManager) drawReaderTopBar(gtx layout.Context) layout.Dimensions {
	barH := 56
	cl := clip.Rect{Max: image.Point{X: gtx.Constraints.Max.X, Y: barH}}.Push(gtx.Ops)
	if wm.readerDimAlpha > 128 {
		paint.Fill(gtx.Ops, theme.WithAlpha(theme.ColorVoidDark, 240))
	} else {
		paint.Fill(gtx.Ops, theme.WithAlpha(theme.ColorGlassWhite, 240))
	}
	cl.Pop()

	textCol := theme.ColorPureBlack
	if wm.readerDimAlpha > 128 {
		textCol = theme.ColorGlassWhite
	}

	return layout.Inset{Top: 14, Bottom: 10, Left: 24, Right: 24}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		return layout.Flex{Axis: layout.Horizontal, Alignment: layout.Middle}.Layout(gtx,
			// Fermer ←
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				if wm.closeReaderBtn.Clicked(gtx) {
					wm.closeReader()
				}
				return wm.closeReaderBtn.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
					lbl := material.Label(wm.theme, 20, "←")
					lbl.Color = theme.ColorCyberCyan
					return lbl.Layout(gtx)
				})
			}),
			layout.Rigid(layout.Spacer{Width: 16}.Layout),
			// Titre
			layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
				title := "Lecture"
				if wm.readerBook != nil {
					title = wm.readerBook.Title
				}
				lbl := material.Label(wm.theme, 14, title)
				lbl.Font.Weight = font.Bold
				lbl.Color = textCol
				lbl.Alignment = text.Middle
				return lbl.Layout(gtx)
			}),
			// Contrôles
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				return layout.Flex{Axis: layout.Horizontal, Alignment: layout.Middle}.Layout(gtx,
					layout.Rigid(func(gtx layout.Context) layout.Dimensions {
						if wm.dimPlusBtn.Clicked(gtx) && wm.readerDimAlpha < 200 {
							wm.readerDimAlpha += 20
						}
						return wm.readerIconBtn(gtx, "🌙", &wm.dimPlusBtn)
					}),
					layout.Rigid(func(gtx layout.Context) layout.Dimensions {
						if wm.dimMinusBtn.Clicked(gtx) && wm.readerDimAlpha >= 20 {
							wm.readerDimAlpha -= 20
						}
						return wm.readerIconBtn(gtx, "☀", &wm.dimMinusBtn)
					}),
					layout.Rigid(layout.Spacer{Width: 12}.Layout),
					layout.Rigid(func(gtx layout.Context) layout.Dimensions {
						if wm.fontMinusBtn.Clicked(gtx) && wm.readerFontSize > 11 {
							wm.readerFontSize -= 1.5
						}
						return wm.readerIconBtn(gtx, "A−", &wm.fontMinusBtn)
					}),
					layout.Rigid(func(gtx layout.Context) layout.Dimensions {
						if wm.fontPlusBtn.Clicked(gtx) && wm.readerFontSize < 32 {
							wm.readerFontSize += 1.5
						}
						return wm.readerIconBtn(gtx, "A+", &wm.fontPlusBtn)
					}),
				)
			}),
		)
	})
}

func (wm *WindowManager) drawReaderContent(gtx layout.Context) layout.Dimensions {
	textCol := theme.ColorPureBlack
	if wm.readerDimAlpha > 128 {
		textCol = theme.ColorGlassWhite
	}

	if wm.readerLoading {
		return layout.Center.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
			lbl := material.Label(wm.theme, 16, "Chargement du livre en cours...")
			lbl.Color = theme.WithAlpha(textCol, 160)
			return lbl.Layout(gtx)
		})
	}
	if len(wm.readerContent) == 0 {
		return layout.Center.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
			lbl := material.Label(wm.theme, 15, "Aucun contenu à afficher.")
			lbl.Color = theme.WithAlpha(textCol, 130)
			return lbl.Layout(gtx)
		})
	}

	pageText := wm.readerContent[wm.readerPage]

	return layout.Center.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		maxW := 720
		if gtx.Constraints.Max.X-80 < maxW {
			maxW = gtx.Constraints.Max.X - 80
		}
		gtx.Constraints.Max.X = maxW
		gtx.Constraints.Min.X = maxW

		return layout.Inset{Top: 48, Bottom: 32, Left: 0, Right: 0}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
			lbl := material.Label(wm.theme, unit.Sp(wm.readerFontSize), pageText)
			lbl.Color = textCol
			return lbl.Layout(gtx)
		})
	})
}

func (wm *WindowManager) drawReaderBottomBar(gtx layout.Context) layout.Dimensions {
	if len(wm.readerContent) == 0 {
		return layout.Dimensions{Size: image.Point{Y: 60}}
	}
	total := len(wm.readerContent)
	progress := float32(wm.readerPage+1) / float32(total)

	cl := clip.Rect{Max: image.Point{X: gtx.Constraints.Max.X, Y: 64}}.Push(gtx.Ops)
	if wm.readerDimAlpha > 128 {
		paint.Fill(gtx.Ops, theme.WithAlpha(theme.ColorVoidDark, 220))
	} else {
		paint.Fill(gtx.Ops, theme.WithAlpha(theme.ColorGlassWhite, 220))
	}
	cl.Pop()

	textCol := theme.ColorPureBlack
	if wm.readerDimAlpha > 128 {
		textCol = theme.ColorGlassWhite
	}

	return layout.Inset{Top: 10, Bottom: 10, Left: 32, Right: 32}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		return layout.Flex{Axis: layout.Horizontal, Alignment: layout.Middle}.Layout(gtx,
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				if wm.readerPrevBtn.Clicked(gtx) && wm.readerPage > 0 {
					wm.readerPage--
				}
				col := theme.ColorCyberCyan
				if wm.readerPage == 0 {
					col = theme.WithAlpha(textCol, 60)
				}
				return wm.readerPrevBtn.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
					lbl := material.Label(wm.theme, 14, "◀  Précédent")
					lbl.Color = col
					lbl.Font.Weight = font.SemiBold
					return lbl.Layout(gtx)
				})
			}),
			layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
				return layout.Flex{Axis: layout.Vertical, Alignment: layout.Middle}.Layout(gtx,
					layout.Rigid(func(gtx layout.Context) layout.Dimensions {
						bar := material.ProgressBar(wm.theme, progress)
						bar.Color = theme.ColorSandGold
						bar.TrackColor = theme.WithAlpha(theme.ColorCyberCyan, 20)
						return layout.Inset{Left: 20, Right: 20, Bottom: 4}.Layout(gtx, bar.Layout)
					}),
					layout.Rigid(func(gtx layout.Context) layout.Dimensions {
						lbl := material.Label(wm.theme, 11, fmt.Sprintf("%d / %d", wm.readerPage+1, total))
						lbl.Color = theme.WithAlpha(textCol, 130)
						lbl.Alignment = text.Middle
						return lbl.Layout(gtx)
					}),
				)
			}),
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				if wm.readerNextBtn.Clicked(gtx) && wm.readerPage < total-1 {
					wm.readerPage++
				}
				col := theme.ColorCyberCyan
				if wm.readerPage >= total-1 {
					col = theme.WithAlpha(textCol, 60)
				}
				return wm.readerNextBtn.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
					lbl := material.Label(wm.theme, 14, "Suivant  ▶")
					lbl.Color = col
					lbl.Font.Weight = font.SemiBold
					return lbl.Layout(gtx)
				})
			}),
		)
	})
}

func (wm *WindowManager) readerIconBtn(gtx layout.Context, icon string, btn *widget.Clickable) layout.Dimensions {
	return btn.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		alpha := uint8(150)
		if btn.Hovered() {
			alpha = 255
		}
		lbl := material.Label(wm.theme, 13, icon)
		lbl.Color = theme.WithAlpha(theme.ColorCyberCyan, alpha)
		lbl.Font.Weight = font.Bold
		return layout.Inset{Left: 8, Right: 8}.Layout(gtx, lbl.Layout)
	})
}

func (wm *WindowManager) closeReader() {
	wm.readerActive = false
	wm.readerBook = nil
	wm.readerContent = nil
	wm.readerPage = 0
	wm.readerLoading = false
}

func (wm *WindowManager) openBookInReader(book *domain.Book) {
	wm.readerActive = true
	wm.readerBook = book
	wm.readerContent = nil
	wm.readerPage = 0
	wm.readerLoading = false
	if wm.readerFontSize == 0 {
		wm.readerFontSize = 16
	}
}
