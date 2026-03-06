package views

import (
	"context"
	"fmt"
	"image"
	"image/color"
	_ "image/png"
	"math"
	"os"
	"sort"
	"strings"
	"time"

	"gioui.org/app"
	"gioui.org/f32"
	"gioui.org/font"
	"gioui.org/font/gofont"
	"gioui.org/io/system"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/text"
	"gioui.org/widget"
	"gioui.org/widget/material"

	"github.com/MiltonJ23/Orus/internal/adapters/ui/theme"
	"github.com/MiltonJ23/Orus/internal/domain"
	"github.com/MiltonJ23/Orus/internal/service"
)

type AppState int

const (
	StateSplash AppState = iota
	StateHome
)

// WindowManager orchestre toutes les vues de l'application
type WindowManager struct {
	window       *app.Window
	theme        *material.Theme
	libSvc       *service.LibraryService
	trackSvc     *service.TrackerService
	sheetSvc     *service.ReadingSheetService
	reminderSvc  *service.ReminderService
	sharingSvc   *service.SharingService
	state        AppState
	appStartTime time.Time
	logo         image.Image

	// Contrôles macOS
	btnClose widget.Clickable
	btnMin   widget.Clickable
	btnMax   widget.Clickable

	// Navigation
	tabs      []string
	tabClicks []widget.Clickable
	activeTab int

	// Recherche
	searchEditor widget.Editor
	searchClick  widget.Clickable

	// Bibliothèque
	books       []*domain.Book
	booksLoaded bool
	gridList    widget.List

	// Fiches de lecture
	sheets       []*domain.ReadingSheet
	sheetsLoaded bool
	sheetForm    sheetFormState

	// Rappels
	reminders       []*domain.Reminder
	remindersLoaded bool
	reminderForm    reminderFormState

	// Notification in-app (bannière rappel)
	activeReminder    *domain.Reminder
	reminderBannerBtn widget.Clickable

	// Partage
	shareStatusMsg string
	shareExportBtn widget.Clickable
	shareLibBtn    widget.Clickable
}

func NewWindowManager(
	lib *service.LibraryService,
	track *service.TrackerService,
	sheet *service.ReadingSheetService,
	reminder *service.ReminderService,
	sharing *service.SharingService,
) *WindowManager {
	th := material.NewTheme()
	th.Shaper = text.NewShaper(text.WithCollection(gofont.Collection()))

	var logoImg image.Image
	f, err := os.Open("internal/adapters/ui/assets/eye.png")
	if err == nil {
		defer f.Close()
		logoImg, _, _ = image.Decode(f)
	}

	menuTabs := []string{
		"Home", "Tous", "À lire", "Terminés",
		"Fiches de lecture", "Rappels", "Partager", "Métriques",
	}
	clicks := make([]widget.Clickable, len(menuTabs))

	wm := &WindowManager{
		window:       new(app.Window),
		theme:        th,
		libSvc:       lib,
		trackSvc:     track,
		sheetSvc:     sheet,
		reminderSvc:  reminder,
		sharingSvc:   sharing,
		state:        StateSplash,
		appStartTime: time.Now(),
		logo:         logoImg,
		tabs:         menuTabs,
		tabClicks:    clicks,
		activeTab:    0,
		searchEditor: widget.Editor{SingleLine: true, Submit: true},
		gridList: widget.List{
			List: layout.List{Axis: layout.Vertical},
		},
	}

	// Enregistrer le hook rappel → bannière in-app
	if reminder != nil {
		reminder.SetCallback(func(r *domain.Reminder) {
			wm.activeReminder = r
			wm.window.Invalidate()
		})
	}

	return wm
}

func (wm *WindowManager) Run() error {
	wm.window.Option(
		app.Title("Orus"),
		app.Size(1440, 900),
		app.Decorated(false),
	)

	var ops op.Ops

	for {
		switch e := wm.window.Event().(type) {
		case app.DestroyEvent:
			return e.Err
		case app.FrameEvent:
			gtx := app.NewContext(&ops, e)
			elapsed := time.Since(wm.appStartTime).Seconds()

			if wm.state == StateSplash && elapsed > 6.5 {
				wm.state = StateHome
			}

			if wm.state == StateSplash {
				wm.layoutSplashScreen(gtx, elapsed)
				wm.window.Invalidate()
			} else {
				wm.layoutHomeScreen(gtx)
			}

			e.Frame(gtx.Ops)
		}
	}
}

// ==========================================================
// OUTILS DE DESSIN
// ==========================================================

func drawSophisticatedObelisk(gtx layout.Context, x, y, width, height float32, c color.NRGBA) {
	goldColor := theme.WithAlpha(theme.ColorSandGold, c.A)

	var pBase clip.Path
	pBase.Begin(gtx.Ops)
	pBase.MoveTo(f32.Pt(x-width/2, y))
	pBase.LineTo(f32.Pt(x+width/2, y))
	pBase.LineTo(f32.Pt(x+width*0.4, y-height*0.04))
	pBase.LineTo(f32.Pt(x-width*0.4, y-height*0.04))
	pBase.Close()
	paint.FillShape(gtx.Ops, c, clip.Outline{Path: pBase.End()}.Op())

	var pBody clip.Path
	pBody.Begin(gtx.Ops)
	pBody.MoveTo(f32.Pt(x-width*0.35, y-height*0.04))
	pBody.LineTo(f32.Pt(x+width*0.35, y-height*0.04))
	pBody.LineTo(f32.Pt(x+width*0.15, y-height*0.85))
	pBody.LineTo(f32.Pt(x-width*0.15, y-height*0.85))
	pBody.Close()
	paint.FillShape(gtx.Ops, c, clip.Outline{Path: pBody.End()}.Op())

	var pTop clip.Path
	pTop.Begin(gtx.Ops)
	pTop.MoveTo(f32.Pt(x-width*0.15, y-height*0.85))
	pTop.LineTo(f32.Pt(x+width*0.15, y-height*0.85))
	pTop.LineTo(f32.Pt(x, y-height))
	pTop.Close()
	paint.FillShape(gtx.Ops, goldColor, clip.Outline{Path: pTop.End()}.Op())

	macroLine := op.Record(gtx.Ops)
	clLine := clip.Rect{
		Min: image.Point{X: int(x) - 1, Y: int(y - height*0.80)},
		Max: image.Point{X: int(x) + 1, Y: int(y - height*0.10)},
	}.Push(gtx.Ops)
	paint.Fill(gtx.Ops, theme.WithAlpha(theme.ColorGlassWhite, c.A))
	clLine.Pop()
	callLine := macroLine.Stop()
	callLine.Add(gtx.Ops)
}

func drawDiamond(gtx layout.Context, x, y, size float32, c color.NRGBA) {
	var p clip.Path
	p.Begin(gtx.Ops)
	p.MoveTo(f32.Pt(x, y-size))
	p.LineTo(f32.Pt(x+size/2, y))
	p.LineTo(f32.Pt(x, y+size))
	p.LineTo(f32.Pt(x-size/2, y))
	p.Close()
	paint.FillShape(gtx.Ops, c, clip.Outline{Path: p.End()}.Op())
}

func drawCircle(gtx layout.Context, x, y, radius float32, c color.NRGBA) {
	macro := op.Record(gtx.Ops)
	cl := clip.UniformRRect(image.Rectangle{
		Min: image.Point{X: int(x - radius), Y: int(y - radius)},
		Max: image.Point{X: int(x + radius), Y: int(y + radius)},
	}, int(radius)).Push(gtx.Ops)
	paint.Fill(gtx.Ops, c)
	cl.Pop()
	call := macro.Stop()
	call.Add(gtx.Ops)
}

func (wm *WindowManager) drawMacControls(gtx layout.Context) {
	radius := float32(6.5)
	spacing := float32(20)
	startX, startY := float32(20), float32(20)

	red := color.NRGBA{R: 255, G: 95, B: 86, A: 255}
	yellow := color.NRGBA{R: 255, G: 189, B: 46, A: 255}
	green := color.NRGBA{R: 39, G: 201, B: 63, A: 255}

	if wm.btnClose.Clicked(gtx) {
		wm.window.Perform(system.ActionClose)
	}
	if wm.btnMin.Clicked(gtx) {
		wm.window.Perform(system.ActionMinimize)
	}
	if wm.btnMax.Clicked(gtx) {
		wm.window.Perform(system.ActionMaximize)
	}

	drawBtn := func(c *widget.Clickable, btnX, btnY float32, col color.NRGBA) {
		area := 20
		defer op.Offset(image.Pt(int(btnX)-area/2, int(btnY)-area/2)).Push(gtx.Ops).Pop()
		c.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
			finalCol := col
			if c.Hovered() {
				finalCol = theme.WithAlpha(col, 200)
			}
			if c.Pressed() {
				finalCol = theme.WithAlpha(col, 150)
			}
			drawCircle(gtx, float32(area/2), float32(area/2), radius, finalCol)
			return layout.Dimensions{Size: image.Point{X: area, Y: area}}
		})
	}

	drawBtn(&wm.btnClose, startX, startY, red)
	drawBtn(&wm.btnMin, startX+spacing, startY, yellow)
	drawBtn(&wm.btnMax, startX+spacing*2, startY, green)
}

// drawPillButton dessine un bouton pill réutilisable
func (wm *WindowManager) drawPillButton(gtx layout.Context, label string, btn *widget.Clickable, col color.NRGBA) layout.Dimensions {
	return btn.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		bgAlpha := uint8(20)
		if btn.Hovered() {
			bgAlpha = 40
		}
		if btn.Pressed() {
			bgAlpha = 60
		}
		cl := clip.UniformRRect(image.Rectangle{
			Max: image.Point{X: gtx.Constraints.Max.X, Y: 34},
		}, 17).Push(gtx.Ops)
		paint.Fill(gtx.Ops, theme.WithAlpha(col, bgAlpha))
		cl.Pop()

		return layout.Inset{Top: 8, Bottom: 8, Left: 16, Right: 16}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
			lbl := material.Label(wm.theme, 13, label)
			lbl.Color = col
			lbl.Font.Weight = font.SemiBold
			return lbl.Layout(gtx)
		})
	})
}

// drawLabeledField dessine un champ de saisie avec label
func (wm *WindowManager) drawLabeledField(gtx layout.Context, label string, editor *widget.Editor, hint string) layout.Dimensions {
	return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			lbl := material.Label(wm.theme, 12, label)
			lbl.Color = theme.WithAlpha(theme.ColorPureBlack, 160)
			return layout.Inset{Bottom: 4}.Layout(gtx, lbl.Layout)
		}),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			cl := clip.UniformRRect(image.Rectangle{
				Max: image.Point{X: gtx.Constraints.Max.X, Y: 36},
			}, 6).Push(gtx.Ops)
			paint.Fill(gtx.Ops, theme.WithAlpha(theme.ColorCyberCyan, 12))
			cl.Pop()
			return layout.Inset{Top: 8, Bottom: 8, Left: 12, Right: 12}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
				ed := material.Editor(wm.theme, editor, hint)
				ed.Color = theme.ColorPureBlack
				ed.HintColor = theme.WithAlpha(theme.ColorCyberCyan, 120)
				ed.TextSize = 13
				return ed.Layout(gtx)
			})
		}),
	)
}

// ==========================================================
// SPLASH SCREEN
// ==========================================================

func (wm *WindowManager) layoutSplashScreen(gtx layout.Context, elapsed float64) layout.Dimensions {
	paint.Fill(gtx.Ops, theme.ColorGlassWhite)

	alphaMultiplier := 1.0
	if elapsed > 4.5 {
		alphaMultiplier = 1.0 - ((elapsed - 4.5) / 2.0)
		if alphaMultiplier < 0 {
			alphaMultiplier = 0
		}
	}

	if alphaMultiplier > 0 {
		w := float32(gtx.Constraints.Max.X)
		h := float32(gtx.Constraints.Max.Y)

		progress := elapsed / 4.5
		if progress > 1.0 {
			progress = 1.0
		}
		easeOut := 1.0 - math.Pow(1.0-progress, 3)

		sunRadius := float32(110)
		sunStartY := h * 0.90
		sunEndY := h / 2.0
		sunY := sunStartY - float32(easeOut)*(sunStartY-sunEndY)
		sunX := w / 2

		drawCircle(gtx, sunX, sunY, sunRadius*2.2, theme.WithAlpha(theme.ColorSandGold, uint8(10*alphaMultiplier)))
		drawCircle(gtx, sunX, sunY, sunRadius*1.4, theme.WithAlpha(theme.ColorSandGold, uint8(25*alphaMultiplier)))
		drawCircle(gtx, sunX, sunY, sunRadius, theme.WithAlpha(theme.ColorSandGold, uint8(60*alphaMultiplier)))

		obeliskY := h + 150 - float32(easeOut*120.0)
		slateAlpha := uint8(75 * alphaMultiplier)
		obeliskColor := theme.WithAlpha(theme.ColorCyberCyan, slateAlpha)
		drawSophisticatedObelisk(gtx, w*0.06, obeliskY, 60, h*0.85, obeliskColor)
		drawSophisticatedObelisk(gtx, w*0.94, obeliskY, 60, h*0.85, obeliskColor)

		for i := 0; i < 3; i++ {
			t := float64(elapsed) * 1.5
			offset := float64(i) * (2.0 * math.Pi / 3.0)
			r := 280.0 + 80.0*math.Cos(t*0.5+offset)
			x := float64(w/2) + math.Cos(t+offset)*r

			yProgress := math.Mod(float64(elapsed)*50.0+float64(i*200), float64(h+300))
			yBase := float64(h+150) - yProgress
			y := yBase + math.Sin(t+offset)*120.0

			size := float32(14 + i*4)
			pColor := theme.WithAlpha(theme.ColorSandGold, uint8(240*alphaMultiplier))
			if i == 1 {
				pColor = theme.WithAlpha(theme.ColorCyberCyan, uint8(240*alphaMultiplier))
			}
			drawDiamond(gtx, float32(x), float32(y), size, pColor)
		}

		layout.Center.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
			return layout.Stack{Alignment: layout.Center}.Layout(gtx,
				layout.Stacked(func(gtx layout.Context) layout.Dimensions {
					if wm.logo != nil {
						imgOp := paint.NewImageOp(wm.logo)
						imgOp.Add(gtx.Ops)
						paint.PaintOp{}.Add(gtx.Ops)
						veilAlpha := uint8(255 * (1.0 - alphaMultiplier))
						if veilAlpha > 0 {
							veilColor := theme.WithAlpha(theme.ColorGlassWhite, veilAlpha)
							paint.FillShape(gtx.Ops, veilColor, clip.Rect{Max: wm.logo.Bounds().Size()}.Op())
						}
						return layout.Dimensions{Size: wm.logo.Bounds().Size()}
					}
					return layout.Dimensions{Size: image.Point{X: 300, Y: 300}}
				}),
			)
		})
	}

	wm.drawMacControls(gtx)
	return layout.Dimensions{Size: gtx.Constraints.Max}
}

// ==========================================================
// HOME SCREEN
// ==========================================================

func (wm *WindowManager) layoutHomeScreen(gtx layout.Context) layout.Dimensions {
	paint.Fill(gtx.Ops, theme.ColorGlassWhite)

	layout.Flex{Axis: layout.Horizontal}.Layout(gtx,
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			gtx.Constraints.Min.X = 240
			gtx.Constraints.Max.X = 240
			return wm.drawSidebar(gtx)
		}),
		layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
			return layout.Inset{Top: 40, Left: 40, Right: 40, Bottom: 24}.Layout(gtx, wm.routeMainContent)
		}),
	)

	// Bannière de rappel (par-dessus tout)
	if wm.activeReminder != nil {
		wm.drawReminderBanner(gtx)
	}

	wm.drawMacControls(gtx)
	return layout.Dimensions{Size: gtx.Constraints.Max}
}

// routeMainContent route vers la vue correspondant à l'onglet actif
func (wm *WindowManager) routeMainContent(gtx layout.Context) layout.Dimensions {
	switch wm.activeTab {
	case 0:
		return wm.drawDashboard(gtx)
	case 1, 2, 3:
		return wm.drawBooksGrid(gtx)
	case 4:
		return wm.drawSheetsView(gtx)
	case 5:
		return wm.drawRemindersView(gtx)
	case 6:
		return wm.drawSharingView(gtx)
	case 7:
		return wm.drawMetrics(gtx)
	default:
		return layout.Center.Layout(gtx, material.H4(wm.theme, "Vue en construction").Layout)
	}
}

// drawReminderBanner affiche une bannière en haut de l'écran quand un rappel sonne
func (wm *WindowManager) drawReminderBanner(gtx layout.Context) {
	if wm.reminderBannerBtn.Clicked(gtx) {
		wm.activeReminder = nil
		return
	}

	bannerH := 50
	defer op.Offset(image.Pt(240, 0)).Push(gtx.Ops).Pop()

	cl := clip.UniformRRect(image.Rectangle{
		Max: image.Point{X: gtx.Constraints.Max.X - 240, Y: bannerH},
	}, 0).Push(gtx.Ops)
	paint.Fill(gtx.Ops, theme.ColorSandGold)
	cl.Pop()

	wm.reminderBannerBtn.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		return layout.Inset{Top: 14, Left: 20}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
			msg := fmt.Sprintf("⏰  %s — Appuyer pour fermer", wm.activeReminder.Label)
			lbl := material.Label(wm.theme, 14, msg)
			lbl.Color = theme.ColorGlassWhite
			lbl.Font.Weight = font.Bold
			return lbl.Layout(gtx)
		})
	})
}

// ==========================================================
// SIDEBAR
// ==========================================================

func (wm *WindowManager) drawSidebar(gtx layout.Context) layout.Dimensions {
	paint.FillShape(gtx.Ops, theme.WithAlpha(theme.ColorCyberCyan, 10), clip.Rect{
		Max: gtx.Constraints.Max,
	}.Op())

	return layout.Inset{Top: 40, Bottom: 20}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				return layout.Inset{Left: 15, Right: 15, Bottom: 20}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
					border := clip.UniformRRect(image.Rectangle{Max: image.Point{X: gtx.Constraints.Max.X, Y: 32}}, 6).Push(gtx.Ops)
					paint.Fill(gtx.Ops, theme.WithAlpha(theme.ColorCyberCyan, 20))
					border.Pop()
					return layout.Inset{Top: 6, Bottom: 6, Left: 10, Right: 10}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
						ed := material.Editor(wm.theme, &wm.searchEditor, "Rechercher...")
						ed.Color = theme.ColorPureBlack
						return ed.Layout(gtx)
					})
				})
			}),
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				return layout.Inset{Left: 10, Right: 10}.Layout(gtx, wm.drawSidebarTabs)
			}),
		)
	})
}

func (wm *WindowManager) drawSidebarTabs(gtx layout.Context) layout.Dimensions {
	var children []layout.FlexChild

	// Icônes par onglet
	icons := []string{"⌂", "☰", "📌", "✓", "📝", "⏰", "↗", "📊"}

	for i := range wm.tabs {
		idx := i
		if wm.tabClicks[idx].Clicked(gtx) {
			wm.activeTab = idx
		}

		// Séparateurs de sections
		if idx == 1 {
			children = append(children, layout.Rigid(layout.Spacer{Height: 25}.Layout))
			children = append(children, layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				lbl := material.Label(wm.theme, 12, "BIBLIOTHÈQUE")
				lbl.Color = theme.WithAlpha(theme.ColorCyberCyan, 180)
				lbl.Font.Weight = font.Bold
				return layout.Inset{Left: 10, Bottom: 10}.Layout(gtx, lbl.Layout)
			}))
		}
		if idx == 4 {
			children = append(children, layout.Rigid(layout.Spacer{Height: 25}.Layout))
			children = append(children, layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				lbl := material.Label(wm.theme, 12, "OUTILS")
				lbl.Color = theme.WithAlpha(theme.ColorCyberCyan, 180)
				lbl.Font.Weight = font.Bold
				return layout.Inset{Left: 10, Bottom: 10}.Layout(gtx, lbl.Layout)
			}))
		}
		// Métriques poussées en bas
		if idx == 7 {
			children = append(children, layout.Flexed(1, layout.Spacer{}.Layout))
		}

		child := layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return wm.tabClicks[idx].Layout(gtx, func(gtx layout.Context) layout.Dimensions {
				bgColor := color.NRGBA{A: 0}
				textColor := theme.ColorPureBlack
				fontWeight := font.Normal

				if wm.activeTab == idx {
					bgColor = theme.WithAlpha(theme.ColorCyberCyan, 25)
					fontWeight = font.Bold
				} else if wm.tabClicks[idx].Hovered() {
					bgColor = theme.WithAlpha(theme.ColorCyberCyan, 10)
				}

				cl := clip.UniformRRect(image.Rectangle{Max: image.Point{X: gtx.Constraints.Max.X, Y: 36}}, 6).Push(gtx.Ops)
				paint.Fill(gtx.Ops, bgColor)
				cl.Pop()

				return layout.Inset{Top: 8, Left: 12, Bottom: 8}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
					return layout.Flex{Axis: layout.Horizontal, Alignment: layout.Middle}.Layout(gtx,
						layout.Rigid(func(gtx layout.Context) layout.Dimensions {
							icon := "•"
							if idx < len(icons) {
								icon = icons[idx]
							}
							lbl := material.Label(wm.theme, 14, icon)
							lbl.Color = theme.WithAlpha(textColor, 180)
							return lbl.Layout(gtx)
						}),
						layout.Rigid(layout.Spacer{Width: 10}.Layout),
						layout.Rigid(func(gtx layout.Context) layout.Dimensions {
							lbl := material.Label(wm.theme, 14, wm.tabs[idx])
							lbl.Color = textColor
							lbl.Font.Weight = fontWeight
							return lbl.Layout(gtx)
						}),
					)
				})
			})
		})
		children = append(children, child, layout.Rigid(layout.Spacer{Height: 4}.Layout))
	}

	return layout.Flex{Axis: layout.Vertical}.Layout(gtx, children...)
}

// ==========================================================
// SHARING VIEW
// ==========================================================

func (wm *WindowManager) drawSharingView(gtx layout.Context) layout.Dimensions {
	return layout.Flex{Axis: layout.Vertical}.Layout(gtx,

		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			lbl := material.H5(wm.theme, "↗ Partager")
			lbl.Font.Weight = font.Bold
			return layout.Inset{Bottom: 8}.Layout(gtx, lbl.Layout)
		}),

		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			lbl := material.Label(wm.theme, 14, "Exportez vos livres et fiches de lecture vers un fichier Markdown ou JSON.")
			lbl.Color = theme.WithAlpha(theme.ColorPureBlack, 160)
			return layout.Inset{Bottom: 32}.Layout(gtx, lbl.Layout)
		}),

		// Carte : Exporter la bibliothèque
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return wm.drawShareCard(gtx,
				"📚 Exporter toute la bibliothèque",
				"Génère un fichier Markdown avec tous tes livres et leurs fiches de lecture.",
				&wm.shareLibBtn,
				func() {
					if wm.sharingSvc == nil {
						wm.shareStatusMsg = "Service non disponible."
						return
					}
					path, err := wm.sharingSvc.ExportLibrary(context.Background(), service.ShareFormatMarkdown, ".")
					if err != nil {
						wm.shareStatusMsg = "Erreur : " + err.Error()
					} else {
						wm.shareStatusMsg = "✓ Bibliothèque exportée → " + path
					}
				},
			)
		}),

		layout.Rigid(layout.Spacer{Height: 16}.Layout),

		// Carte : Exporter les fiches
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return wm.drawShareCard(gtx,
				"📝 Exporter toutes les fiches (JSON)",
				"Génère un fichier JSON de toutes tes fiches de lecture, idéal pour un backup.",
				&wm.shareExportBtn,
				func() {
					if wm.sharingSvc == nil {
						wm.shareStatusMsg = "Service non disponible."
						return
					}
					path, err := wm.sharingSvc.ExportLibrary(context.Background(), service.ShareFormatJSON, ".")
					if err != nil {
						wm.shareStatusMsg = "Erreur : " + err.Error()
					} else {
						wm.shareStatusMsg = "✓ Fiches exportées → " + path
					}
				},
			)
		}),

		layout.Rigid(layout.Spacer{Height: 24}.Layout),

		// Message de statut
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			if wm.shareStatusMsg == "" {
				return layout.Dimensions{}
			}
			lbl := material.Label(wm.theme, 14, wm.shareStatusMsg)
			lbl.Color = theme.ColorCyberCyan
			lbl.Font.Weight = font.SemiBold
			return lbl.Layout(gtx)
		}),
	)
}

func (wm *WindowManager) drawShareCard(gtx layout.Context, title, desc string, btn *widget.Clickable, onPress func()) layout.Dimensions {
	cl := clip.UniformRRect(image.Rectangle{Max: image.Point{X: gtx.Constraints.Max.X, Y: 120}}, 10).Push(gtx.Ops)
	paint.Fill(gtx.Ops, theme.WithAlpha(theme.ColorCyberCyan, 8))
	cl.Pop()

	return layout.Inset{Top: 20, Left: 24, Right: 24, Bottom: 20}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		return layout.Flex{Axis: layout.Horizontal, Alignment: layout.Middle}.Layout(gtx,
			layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
				return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
					layout.Rigid(func(gtx layout.Context) layout.Dimensions {
						lbl := material.Label(wm.theme, 15, title)
						lbl.Font.Weight = font.Bold
						return layout.Inset{Bottom: 6}.Layout(gtx, lbl.Layout)
					}),
					layout.Rigid(func(gtx layout.Context) layout.Dimensions {
						lbl := material.Label(wm.theme, 13, desc)
						lbl.Color = theme.WithAlpha(theme.ColorPureBlack, 150)
						return lbl.Layout(gtx)
					}),
				)
			}),
			layout.Rigid(layout.Spacer{Width: 16}.Layout),
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				if btn.Clicked(gtx) {
					onPress()
				}
				return wm.drawPillButton(gtx, "Exporter", btn, theme.ColorSandGold)
			}),
		)
	})
}

// ==========================================================
// DASHBOARD
// ==========================================================

func (wm *WindowManager) drawDashboard(gtx layout.Context) layout.Dimensions {
	return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return layout.Flex{Axis: layout.Horizontal, Alignment: layout.Middle}.Layout(gtx,
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					w, h := 180, 270
					shadowRect := clip.UniformRRect(image.Rectangle{Min: image.Point{X: 6, Y: 12}, Max: image.Point{X: w + 6, Y: h + 15}}, 8).Push(gtx.Ops)
					paint.Fill(gtx.Ops, theme.WithAlpha(theme.ColorPureBlack, 20))
					shadowRect.Pop()
					coverRect := clip.UniformRRect(image.Rectangle{Max: image.Point{X: w, Y: h}}, 8).Push(gtx.Ops)
					paint.Fill(gtx.Ops, theme.ColorCyberCyan)
					coverRect.Pop()
					return layout.Dimensions{Size: image.Point{X: w + 30, Y: h + 20}}
				}),
				layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
					return layout.Inset{Left: 40}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
						return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
							layout.Rigid(func(gtx layout.Context) layout.Dimensions {
								lbl := material.Label(wm.theme, 13, "REPRENDRE LA LECTURE")
								lbl.Color = theme.WithAlpha(theme.ColorCyberCyan, 150)
								lbl.Font.Weight = font.Bold
								return lbl.Layout(gtx)
							}),
							layout.Rigid(layout.Spacer{Height: 15}.Layout),
							layout.Rigid(func(gtx layout.Context) layout.Dimensions {
								lbl := material.H4(wm.theme, "Les tueurs de la République")
								lbl.Font.Weight = font.Bold
								return lbl.Layout(gtx)
							}),
							layout.Rigid(layout.Spacer{Height: 40}.Layout),
							layout.Rigid(func(gtx layout.Context) layout.Dimensions {
								completion := float32(0.31)
								return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
									layout.Rigid(func(gtx layout.Context) layout.Dimensions {
										lbl := material.Label(wm.theme, 15, "31% terminé (Page 120 sur 380)")
										lbl.Color = theme.ColorCyberCyan
										return lbl.Layout(gtx)
									}),
									layout.Rigid(layout.Spacer{Height: 12}.Layout),
									layout.Rigid(func(gtx layout.Context) layout.Dimensions {
										bar := material.ProgressBar(wm.theme, completion)
										bar.Color = theme.ColorSandGold
										bar.TrackColor = theme.WithAlpha(theme.ColorCyberCyan, 20)
										return bar.Layout(gtx)
									}),
								)
							}),
						)
					})
				}),
			)
		}),
		layout.Flexed(1, layout.Spacer{}.Layout),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return layout.Flex{Axis: layout.Horizontal}.Layout(gtx,
				layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
					return wm.drawAnalyticCard(gtx, "Temps moyen", "45 min / jour", "")
				}),
				layout.Rigid(layout.Spacer{Width: 30}.Layout),
				layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
					return wm.drawAnalyticCard(gtx, "Livres terminés", "12 cette année", "")
				}),
				layout.Rigid(layout.Spacer{Width: 30}.Layout),
				layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
					return wm.drawAnalyticCard(gtx, "Pages lues", "3 450 pages", "")
				}),
			)
		}),
	)
}

// ==========================================================
// METRICS
// ==========================================================

func (wm *WindowManager) computeMetrics() (bestDay, bestHour string, recentSessions []domain.ReadingSession) {
	now := time.Now()
	sessions := []domain.ReadingSession{
		{BookID: "Livre A", CurrentPage: 120, LastReadingTime: now.Add(-2 * time.Hour)},
		{BookID: "Livre B", CurrentPage: 45, LastReadingTime: now.Add(-26 * time.Hour)},
		{BookID: "Livre C", CurrentPage: 300, LastReadingTime: now.Add(-72 * time.Hour)},
		{BookID: "Livre A", CurrentPage: 100, LastReadingTime: now.Add(-96 * time.Hour)},
	}
	if len(sessions) == 0 {
		return "N/A", "N/A", nil
	}
	daysCount := make(map[time.Weekday]int)
	hoursCount := make(map[int]int)
	for _, s := range sessions {
		daysCount[s.LastReadingTime.Weekday()]++
		hoursCount[s.LastReadingTime.Hour()]++
	}
	var maxDay time.Weekday
	maxDCount := 0
	for d, count := range daysCount {
		if count > maxDCount {
			maxDCount = count
			maxDay = d
		}
	}
	var maxHour int
	maxHCount := 0
	for h, count := range hoursCount {
		if count > maxHCount {
			maxHCount = count
			maxHour = h
		}
	}
	frDays := map[time.Weekday]string{
		time.Sunday: "Dimanche", time.Monday: "Lundi", time.Tuesday: "Mardi",
		time.Wednesday: "Mercredi", time.Thursday: "Jeudi", time.Friday: "Vendredi", time.Saturday: "Samedi",
	}
	bestDay = frDays[maxDay]
	bestHour = fmt.Sprintf("%dh00 - %dh00", maxHour, maxHour+2)
	sort.Slice(sessions, func(i, j int) bool {
		return sessions[i].LastReadingTime.After(sessions[j].LastReadingTime)
	})
	return bestDay, bestHour, sessions
}

func (wm *WindowManager) drawMetrics(gtx layout.Context) layout.Dimensions {
	bestDay, bestHour, history := wm.computeMetrics()
	return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return layout.Flex{Axis: layout.Horizontal}.Layout(gtx,
				layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
					return wm.drawAnalyticCard(gtx, "Jour le plus actif", bestDay, "Vos sessions sont plus longues ce jour-là.")
				}),
				layout.Rigid(layout.Spacer{Width: 20}.Layout),
				layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
					return wm.drawAnalyticCard(gtx, "Heure de prédilection", bestHour, "Vous lisez principalement dans cette tranche.")
				}),
			)
		}),
		layout.Rigid(layout.Spacer{Height: 40}.Layout),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			lbl := material.H6(wm.theme, "Historique récent des sessions")
			lbl.Font.Weight = font.Bold
			return layout.Inset{Bottom: 15}.Layout(gtx, lbl.Layout)
		}),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			var rows []layout.FlexChild
			for _, s := range history {
				session := s
				row := layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					cl := clip.UniformRRect(image.Rectangle{Max: image.Point{X: gtx.Constraints.Max.X, Y: 50}}, 6).Push(gtx.Ops)
					paint.Fill(gtx.Ops, theme.WithAlpha(theme.ColorGlassWhite, 150))
					cl.Pop()
					return layout.Inset{Top: 10, Bottom: 10, Left: 15, Right: 15}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
						return layout.Flex{Axis: layout.Horizontal, Alignment: layout.Middle}.Layout(gtx,
							layout.Flexed(0.4, func(gtx layout.Context) layout.Dimensions {
								return material.Label(wm.theme, 14, session.BookID).Layout(gtx)
							}),
							layout.Flexed(0.3, func(gtx layout.Context) layout.Dimensions {
								lbl := material.Label(wm.theme, 13, session.LastReadingTime.Format("02 Jan 15:04"))
								lbl.Color = theme.WithAlpha(theme.ColorPureBlack, 150)
								return lbl.Layout(gtx)
							}),
							layout.Flexed(0.3, func(gtx layout.Context) layout.Dimensions {
								lbl := material.Label(wm.theme, 14, fmt.Sprintf("Page %d atteinte", session.CurrentPage))
								lbl.Color = theme.ColorCyberCyan
								lbl.Font.Weight = font.Bold
								return lbl.Layout(gtx)
							}),
						)
					})
				})
				rows = append(rows, row, layout.Rigid(layout.Spacer{Height: 8}.Layout))
			}
			return layout.Flex{Axis: layout.Vertical}.Layout(gtx, rows...)
		}),
	)
}

func (wm *WindowManager) drawAnalyticCard(gtx layout.Context, title, mainValue, subtitle string) layout.Dimensions {
	cl := clip.UniformRRect(image.Rectangle{Max: image.Point{X: gtx.Constraints.Max.X, Y: 110}}, 8).Push(gtx.Ops)
	paint.Fill(gtx.Ops, theme.WithAlpha(theme.ColorSandGold, 20))
	cl.Pop()
	return layout.Inset{Top: 20, Left: 20, Right: 20}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				lbl := material.Label(wm.theme, 13, title)
				lbl.Color = theme.WithAlpha(theme.ColorPureBlack, 180)
				return lbl.Layout(gtx)
			}),
			layout.Rigid(layout.Spacer{Height: 8}.Layout),
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				lbl := material.Label(wm.theme, 18, mainValue)
				lbl.Font.Weight = font.Bold
				lbl.Color = theme.ColorPureBlack
				return lbl.Layout(gtx)
			}),
			layout.Rigid(layout.Spacer{Height: 5}.Layout),
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				if subtitle == "" {
					return layout.Dimensions{}
				}
				lbl := material.Label(wm.theme, 12, subtitle)
				lbl.Color = theme.ColorCyberCyan
				return lbl.Layout(gtx)
			}),
		)
	})
}

// ==========================================================
// SEARCH BAR (utilitaire)
// ==========================================================

func (wm *WindowManager) drawSearchBar(gtx layout.Context) layout.Dimensions {
	height := 32
	radius := 6
	bg := theme.WithAlpha(theme.ColorCyberCyan, 15)
	if wm.searchClick.Hovered() {
		bg = theme.WithAlpha(theme.ColorCyberCyan, 25)
	}
	return wm.searchClick.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		cl := clip.UniformRRect(image.Rectangle{Max: image.Point{X: gtx.Constraints.Max.X, Y: height}}, radius).Push(gtx.Ops)
		paint.Fill(gtx.Ops, bg)
		cl.Pop()
		return layout.Inset{Top: 6, Left: 12, Right: 12}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
			ed := material.Editor(wm.theme, &wm.searchEditor, "Rechercher...")
			ed.Color = theme.ColorPureBlack
			ed.HintColor = theme.WithAlpha(theme.ColorCyberCyan, 120)
			ed.TextSize = 13
			return ed.Layout(gtx)
		})
	})
}

// ==========================================================
// HELPERS PARTAGÉS
// ==========================================================

func splitTrim(s, sep string) []string {
	var result []string
	for _, part := range strings.Split(s, sep) {
		if t := strings.TrimSpace(part); t != "" {
			result = append(result, t)
		}
	}
	return result
}

func joinStr(items []string, sep string) string {
	return strings.Join(items, sep)
}
