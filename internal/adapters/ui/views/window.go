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
	"github.com/MiltonJ23/Orus/internal/port"
	"github.com/MiltonJ23/Orus/internal/service"
)

type AppState int

const (
	StateSplash AppState = iota
	StateHome
)

type WindowManager struct {
	window        *app.Window
	theme         *material.Theme
	libSvc        *service.LibraryService
	trackSvc      *service.TrackerService
	sheetSvc      *service.ReadingSheetService
	reminderSvc   *service.ReminderService
	sharingSvc    *service.SharingService
	contentReader port.ContentReader
	state         AppState
	appStartTime  time.Time
	logo          image.Image

	// Mac window controls
	btnClose widget.Clickable
	btnMin   widget.Clickable
	btnMax   widget.Clickable

	// Sidebar navigation
	tabs      []string
	tabClicks []widget.Clickable
	activeTab int

	// Search
	searchEditor widget.Editor

	// Library
	books        []*domain.Book
	booksLoaded  bool
	gridList     widget.List
	bookOpenBtns []widget.Clickable

	// Dashboard — real data
	continueReadBtn widget.Clickable
	dashboardLoaded bool
	currentBook     *domain.Book
	currentSession  *domain.ReadingSession

	// Reading sheets
	sheets       []*domain.ReadingSheet
	sheetsLoaded bool
	sheetForm    sheetFormState

	// Sheet detail view
	activeSheetDetail   *domain.ReadingSheet
	sheetDetailBtns     []widget.Clickable
	closeSheetDetailBtn widget.Clickable
	shareSheetBtn       widget.Clickable
	shareSheetStatus    string

	// Sheet filter
	sheetFilterIdx  int
	sheetFilterBtns []widget.Clickable // one per unique book in sheets

	// Book selector for sheet form
	bookSelectBtns  []widget.Clickable
	selectedBookIdx int

	// Reminders
	reminders       []*domain.Reminder
	remindersLoaded bool
	reminderForm    reminderFormState

	// Reminder in-app banner
	activeReminder    *domain.Reminder
	reminderBannerBtn widget.Clickable

	// Sharing
	shareStatusMsg string
	shareLibBtn    widget.Clickable
	shareExportBtn widget.Clickable

	// Reader
	readerActive   bool
	readerBook     *domain.Book
	readerContent  []string
	readerPage     int
	readerFontSize float32
	readerDimAlpha uint8
	readerLoading  bool
	closeReaderBtn widget.Clickable
	fontPlusBtn    widget.Clickable
	fontMinusBtn   widget.Clickable
	dimPlusBtn     widget.Clickable
	dimMinusBtn    widget.Clickable
	readerPrevBtn  widget.Clickable
	readerNextBtn  widget.Clickable

	// Metrics
	metricsLoaded  bool
	recentSessions []*domain.ReadingSession
}

func NewWindowManager(
	lib *service.LibraryService,
	track *service.TrackerService,
	sheet *service.ReadingSheetService,
	reminder *service.ReminderService,
	sharing *service.SharingService,
	contentReader port.ContentReader,
) *WindowManager {
	th := material.NewTheme()
	th.Shaper = text.NewShaper(text.WithCollection(gofont.Collection()))

	var logoImg image.Image
	if f, err := os.Open("internal/adapters/ui/assets/eye.png"); err == nil {
		defer f.Close()
		logoImg, _, _ = image.Decode(f)
	}

	menuTabs := []string{
		"Home", "Tous", "A lire", "Termines",
		"Fiches", "Rappels", "Partager", "Metriques",
	}
	clicks := make([]widget.Clickable, len(menuTabs))

	wm := &WindowManager{
		window:          new(app.Window),
		theme:           th,
		libSvc:          lib,
		trackSvc:        track,
		sheetSvc:        sheet,
		reminderSvc:     reminder,
		sharingSvc:      sharing,
		contentReader:   contentReader,
		state:           StateSplash,
		appStartTime:    time.Now(),
		logo:            logoImg,
		tabs:            menuTabs,
		tabClicks:       clicks,
		activeTab:       0,
		searchEditor:    widget.Editor{SingleLine: true, Submit: true},
		readerFontSize:  16,
		gridList:        widget.List{List: layout.List{Axis: layout.Vertical}},
		selectedBookIdx: -1,
	}

	if reminder != nil {
		reminder.SetCallback(func(r *domain.Reminder) {
			wm.activeReminder = r
			wm.window.Invalidate()
		})
	}
	return wm
}

func (wm *WindowManager) Run() error {
	wm.window.Option(app.Title("Orus"), app.Size(1440, 900), app.Decorated(false))
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
// PRIMITIVE SHAPES
// ==========================================================

func drawArcThick(gtx layout.Context, cx, cy, r, thickness, startDeg, endDeg float32, c color.NRGBA) {
	if endDeg <= startDeg {
		return
	}
	startRad := float64(startDeg) * math.Pi / 180
	endRad := float64(endDeg) * math.Pi / 180
	inner := r - thickness/2
	outer := r + thickness/2
	steps := 80
	var p clip.Path
	p.Begin(gtx.Ops)
	for i := 0; i <= steps; i++ {
		t := float64(i) / float64(steps)
		angle := startRad + (endRad-startRad)*t
		x := cx + outer*float32(math.Cos(angle))
		y := cy + outer*float32(math.Sin(angle))
		if i == 0 {
			p.MoveTo(f32.Pt(x, y))
		} else {
			p.LineTo(f32.Pt(x, y))
		}
	}
	for i := steps; i >= 0; i-- {
		t := float64(i) / float64(steps)
		angle := startRad + (endRad-startRad)*t
		x := cx + inner*float32(math.Cos(angle))
		y := cy + inner*float32(math.Sin(angle))
		p.LineTo(f32.Pt(x, y))
	}
	p.Close()
	paint.FillShape(gtx.Ops, c, clip.Outline{Path: p.End()}.Op())
}

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
	m := op.Record(gtx.Ops)
	cl := clip.Rect{Min: image.Point{X: int(x) - 1, Y: int(y - height*0.80)}, Max: image.Point{X: int(x) + 1, Y: int(y - height*0.10)}}.Push(gtx.Ops)
	paint.Fill(gtx.Ops, theme.WithAlpha(theme.ColorGlassWhite, c.A))
	cl.Pop()
	m.Stop().Add(gtx.Ops)
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
	m := op.Record(gtx.Ops)
	cl := clip.UniformRRect(image.Rectangle{
		Min: image.Point{X: int(x - radius), Y: int(y - radius)},
		Max: image.Point{X: int(x + radius), Y: int(y + radius)},
	}, int(radius)).Push(gtx.Ops)
	paint.Fill(gtx.Ops, c)
	cl.Pop()
	m.Stop().Add(gtx.Ops)
}

// ==========================================================
// MAC WINDOW CONTROLS  — fixed: no defer in loop
// ==========================================================

func (wm *WindowManager) drawMacControls(gtx layout.Context) {
	if wm.btnClose.Clicked(gtx) {
		wm.window.Perform(system.ActionClose)
	}
	if wm.btnMin.Clicked(gtx) {
		wm.window.Perform(system.ActionMinimize)
	}
	if wm.btnMax.Clicked(gtx) {
		wm.window.Perform(system.ActionMaximize)
	}

	radius := float32(6.5)
	spacing := float32(20)
	sx, sy := float32(20), float32(20)
	colors := []color.NRGBA{
		{R: 255, G: 95, B: 86, A: 255},  // red — close
		{R: 255, G: 189, B: 46, A: 255}, // yellow — minimize
		{R: 39, G: 201, B: 63, A: 255},  // green — maximize
	}
	btns := []*widget.Clickable{&wm.btnClose, &wm.btnMin, &wm.btnMax}

	for i, btn := range btns {
		b := btn
		col := colors[i]
		area := 20
		bx := int(sx+float32(i)*spacing) - area/2
		by := int(sy) - area/2

		// FIX: use immediate push/pop, NOT defer (defer in a loop breaks all three buttons)
		stack := op.Offset(image.Pt(bx, by)).Push(gtx.Ops)
		b.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
			c := col
			if b.Hovered() {
				c = theme.WithAlpha(col, 200)
			}
			if b.Pressed() {
				c = theme.WithAlpha(col, 140)
			}
			drawCircle(gtx, float32(area/2), float32(area/2), radius, c)
			// Inner symbol on hover
			if b.Hovered() {
				sym := theme.WithAlpha(color.NRGBA{R: 80, G: 30, B: 0, A: 255}, 180)
				switch i {
				case 0: // × close
					sz := float32(3)
					half := float32(area / 2)
					drawCircle(gtx, half-sz, half-sz, 1.2, sym)
					drawCircle(gtx, half+sz, half-sz, 1.2, sym)
					drawCircle(gtx, half, half, 1.2, sym)
					drawCircle(gtx, half-sz, half+sz, 1.2, sym)
					drawCircle(gtx, half+sz, half+sz, 1.2, sym)
				case 1: // – minimize
					cl := clip.Rect{Min: image.Point{X: area/2 - 3, Y: area / 2}, Max: image.Point{X: area/2 + 3, Y: area/2 + 1}}.Push(gtx.Ops)
					paint.Fill(gtx.Ops, sym)
					cl.Pop()
				}
			}
			return layout.Dimensions{Size: image.Point{X: area, Y: area}}
		})
		stack.Pop()
	}
}

// ==========================================================
// SHARED COMPONENTS
// ==========================================================

// drawPillButton — lightweight ghost pill button
func (wm *WindowManager) drawPillButton(gtx layout.Context, label string, btn *widget.Clickable, col color.NRGBA) layout.Dimensions {
	return btn.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		bgA := uint8(18)
		if btn.Hovered() {
			bgA = 40
		}
		if btn.Pressed() {
			bgA = 65
		}
		cl := clip.UniformRRect(image.Rectangle{Max: image.Point{X: gtx.Constraints.Max.X, Y: 34}}, 17).Push(gtx.Ops)
		paint.Fill(gtx.Ops, theme.WithAlpha(col, bgA))
		cl.Pop()
		return layout.Inset{Top: 8, Bottom: 8, Left: 16, Right: 16}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
			lbl := material.Label(wm.theme, 13, label)
			lbl.Color = col
			lbl.Font.Weight = font.SemiBold
			return lbl.Layout(gtx)
		})
	})
}

// drawGlowCTA — solid filled CTA button with bloom glow on hover
func (wm *WindowManager) drawGlowCTA(gtx layout.Context, label string, btn *widget.Clickable, col color.NRGBA) layout.Dimensions {
	const btnW, btnH = 150, 40
	const glowPad = 14

	// Total widget size includes glow padding on all sides
	totalW := btnW + glowPad*2
	totalH := btnH + glowPad*2

	return btn.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		if btn.Hovered() {
			// Outer glow layers (largest → most transparent)
			for _, g := range []struct{ pad, alpha int }{
				{0, 5}, {3, 10}, {6, 18}, {9, 10},
			} {
				cl := clip.UniformRRect(image.Rectangle{
					Min: image.Point{X: g.pad, Y: g.pad},
					Max: image.Point{X: totalW - g.pad, Y: totalH - g.pad},
				}, 22+g.pad).Push(gtx.Ops)
				paint.Fill(gtx.Ops, theme.WithAlpha(col, uint8(g.alpha)))
				cl.Pop()
			}
		}

		// The button itself sits inside the glow padding area
		btnStack := op.Offset(image.Pt(glowPad, glowPad)).Push(gtx.Ops)

		bgAlpha := uint8(215)
		if btn.Hovered() {
			bgAlpha = 245
		}
		if btn.Pressed() {
			bgAlpha = 175
		}

		// Subtle inner shine at top
		if btn.Hovered() {
			shineR := clip.UniformRRect(image.Rectangle{Max: image.Point{X: btnW, Y: btnH / 2}}, 20).Push(gtx.Ops)
			paint.Fill(gtx.Ops, theme.WithAlpha(theme.ColorGlassWhite, 15))
			shineR.Pop()
		}

		btnR := clip.UniformRRect(image.Rectangle{Max: image.Point{X: btnW, Y: btnH}}, 20).Push(gtx.Ops)
		paint.Fill(gtx.Ops, theme.WithAlpha(col, bgAlpha))
		btnR.Pop()

		layout.Inset{Top: 11, Left: 22}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
			lbl := material.Label(wm.theme, 13, label)
			lbl.Color = theme.ColorGlassWhite
			lbl.Font.Weight = font.Bold
			return lbl.Layout(gtx)
		})

		btnStack.Pop()

		return layout.Dimensions{Size: image.Point{X: totalW, Y: totalH}}
	})
}

// drawLabeledField — text field with label above
func (wm *WindowManager) drawLabeledField(gtx layout.Context, label string, editor *widget.Editor, hint string) layout.Dimensions {
	return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			if label == "" {
				return layout.Dimensions{}
			}
			lbl := material.Label(wm.theme, 12, label)
			lbl.Color = theme.WithAlpha(theme.ColorPureBlack, 155)
			return layout.Inset{Bottom: 4}.Layout(gtx, lbl.Layout)
		}),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			cl := clip.UniformRRect(image.Rectangle{Max: image.Point{X: gtx.Constraints.Max.X, Y: 36}}, 6).Push(gtx.Ops)
			paint.Fill(gtx.Ops, theme.WithAlpha(theme.ColorCyberCyan, 12))
			cl.Pop()
			return layout.Inset{Top: 8, Bottom: 8, Left: 12, Right: 12}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
				ed := material.Editor(wm.theme, editor, hint)
				ed.Color = theme.ColorPureBlack
				ed.HintColor = theme.WithAlpha(theme.ColorCyberCyan, 110)
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
	alphaM := 1.0
	if elapsed > 4.5 {
		alphaM = 1.0 - (elapsed-4.5)/2.0
		if alphaM < 0 {
			alphaM = 0
		}
	}
	if alphaM > 0 {
		w := float32(gtx.Constraints.Max.X)
		h := float32(gtx.Constraints.Max.Y)
		prog := math.Min(elapsed/4.5, 1.0)
		ease := 1.0 - math.Pow(1.0-prog, 3)
		sunR := float32(110)
		sunY := h*0.90 - float32(ease)*(h*0.90-h/2)
		sunX := w / 2
		drawCircle(gtx, sunX, sunY, sunR*2.2, theme.WithAlpha(theme.ColorSandGold, uint8(10*alphaM)))
		drawCircle(gtx, sunX, sunY, sunR*1.4, theme.WithAlpha(theme.ColorSandGold, uint8(25*alphaM)))
		drawCircle(gtx, sunX, sunY, sunR, theme.WithAlpha(theme.ColorSandGold, uint8(60*alphaM)))
		obY := h + 150 - float32(ease*120)
		obColor := theme.WithAlpha(theme.ColorCyberCyan, uint8(75*alphaM))
		drawSophisticatedObelisk(gtx, w*0.06, obY, 60, h*0.85, obColor)
		drawSophisticatedObelisk(gtx, w*0.94, obY, 60, h*0.85, obColor)
		for i := 0; i < 3; i++ {
			t := elapsed * 1.5
			off := float64(i) * (2 * math.Pi / 3)
			r := 280.0 + 80.0*math.Cos(t*0.5+off)
			x := float64(w/2) + math.Cos(t+off)*r
			yProg := math.Mod(elapsed*50+float64(i*200), float64(h+300))
			y := float64(h+150) - yProg + math.Sin(t+off)*120
			size := float32(14 + i*4)
			pc := theme.WithAlpha(theme.ColorSandGold, uint8(240*alphaM))
			if i == 1 {
				pc = theme.WithAlpha(theme.ColorCyberCyan, uint8(240*alphaM))
			}
			drawDiamond(gtx, float32(x), float32(y), size, pc)
		}
		layout.Center.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
			return layout.Stack{Alignment: layout.Center}.Layout(gtx,
				layout.Stacked(func(gtx layout.Context) layout.Dimensions {
					if wm.logo != nil {
						paint.NewImageOp(wm.logo).Add(gtx.Ops)
						paint.PaintOp{}.Add(gtx.Ops)
						veil := uint8(255 * (1.0 - alphaM))
						if veil > 0 {
							paint.FillShape(gtx.Ops, theme.WithAlpha(theme.ColorGlassWhite, veil), clip.Rect{Max: wm.logo.Bounds().Size()}.Op())
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
	// Sheet detail overlay takes precedence
	if wm.activeSheetDetail != nil {
		wm.drawSheetDetailView(gtx)
		wm.drawMacControls(gtx)
		return layout.Dimensions{Size: gtx.Constraints.Max}
	}
	// Reader takes full screen
	if wm.readerActive {
		wm.drawReaderView(gtx)
		return layout.Dimensions{Size: gtx.Constraints.Max}
	}

	paint.Fill(gtx.Ops, theme.ColorGlassWhite)

	layout.Flex{Axis: layout.Horizontal}.Layout(gtx,
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			gtx.Constraints.Min.X = 240
			gtx.Constraints.Max.X = 240
			return wm.drawSidebar(gtx)
		}),
		layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
			return layout.Inset{Top: 40, Left: 44, Right: 44, Bottom: 24}.Layout(gtx, wm.routeMainContent)
		}),
	)

	if wm.activeReminder != nil {
		wm.drawReminderBanner(gtx)
	}
	wm.drawMacControls(gtx)
	return layout.Dimensions{Size: gtx.Constraints.Max}
}

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
		return layout.Center.Layout(gtx, material.H4(wm.theme, "En construction").Layout)
	}
}

// drawReminderBanner — gold banner at top with dismiss
func (wm *WindowManager) drawReminderBanner(gtx layout.Context) {
	if wm.reminderBannerBtn.Clicked(gtx) {
		rem := wm.activeReminder
		wm.activeReminder = nil
		if wm.reminderSvc != nil && rem != nil {
			go func() {
				_ = wm.reminderSvc.DismissReminder(context.Background(), rem.ID)
				wm.remindersLoaded = false
				wm.window.Invalidate()
			}()
		}
		return
	}
	bannerH := 48
	stack := op.Offset(image.Pt(240, 0)).Push(gtx.Ops)
	w := gtx.Constraints.Max.X - 240
	cl := clip.UniformRRect(image.Rectangle{Max: image.Point{X: w, Y: bannerH}}, 0).Push(gtx.Ops)
	paint.Fill(gtx.Ops, theme.ColorSandGold)
	cl.Pop()
	wm.reminderBannerBtn.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		return layout.Inset{Top: 13, Left: 24}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
			msg := fmt.Sprintf(">>  %s    |    Cliquer pour fermer", wm.activeReminder.Label)
			lbl := material.Label(wm.theme, 14, msg)
			lbl.Color = theme.ColorGlassWhite
			lbl.Font.Weight = font.Bold
			return lbl.Layout(gtx)
		})
	})
	stack.Pop()
}

// ==========================================================
// SIDEBAR — icons drawn as shapes (no emoji font dependency)
// ==========================================================

func (wm *WindowManager) drawSidebar(gtx layout.Context) layout.Dimensions {
	paint.FillShape(gtx.Ops, theme.WithAlpha(theme.ColorCyberCyan, 8), clip.Rect{Max: gtx.Constraints.Max}.Op())
	return layout.Inset{Top: 40, Bottom: 20}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				return layout.Inset{Left: 15, Right: 15, Bottom: 20}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
					cl := clip.UniformRRect(image.Rectangle{Max: image.Point{X: gtx.Constraints.Max.X, Y: 32}}, 6).Push(gtx.Ops)
					paint.Fill(gtx.Ops, theme.WithAlpha(theme.ColorCyberCyan, 18))
					cl.Pop()
					return layout.Inset{Top: 6, Bottom: 6, Left: 10}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
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

	for i := range wm.tabs {
		idx := i
		if wm.tabClicks[idx].Clicked(gtx) {
			wm.activeTab = idx
		}

		// Section headers
		if idx == 1 {
			children = append(children,
				layout.Rigid(layout.Spacer{Height: 22}.Layout),
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					lbl := material.Label(wm.theme, 10, "BIBLIOTHEQUE")
					lbl.Color = theme.WithAlpha(theme.ColorCyberCyan, 170)
					lbl.Font.Weight = font.Bold
					return layout.Inset{Left: 10, Bottom: 8}.Layout(gtx, lbl.Layout)
				}))
		}
		if idx == 4 {
			children = append(children,
				layout.Rigid(layout.Spacer{Height: 22}.Layout),
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					lbl := material.Label(wm.theme, 10, "OUTILS")
					lbl.Color = theme.WithAlpha(theme.ColorCyberCyan, 170)
					lbl.Font.Weight = font.Bold
					return layout.Inset{Left: 10, Bottom: 8}.Layout(gtx, lbl.Layout)
				}))
		}
		if idx == 7 {
			children = append(children, layout.Flexed(1, layout.Spacer{}.Layout))
		}

		child := layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return wm.tabClicks[idx].Layout(gtx, func(gtx layout.Context) layout.Dimensions {
				isActive := wm.activeTab == idx
				bg := color.NRGBA{}
				if isActive {
					bg = theme.WithAlpha(theme.ColorCyberCyan, 22)
				} else if wm.tabClicks[idx].Hovered() {
					bg = theme.WithAlpha(theme.ColorCyberCyan, 9)
				}
				cl := clip.UniformRRect(image.Rectangle{Max: image.Point{X: gtx.Constraints.Max.X, Y: 36}}, 6).Push(gtx.Ops)
				paint.Fill(gtx.Ops, bg)
				cl.Pop()

				return layout.Inset{Top: 9, Left: 12, Bottom: 9}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
					return layout.Flex{Axis: layout.Horizontal, Alignment: layout.Middle}.Layout(gtx,
						// Icon: drawn shape instead of emoji/unicode
						layout.Rigid(func(gtx layout.Context) layout.Dimensions {
							const iconSz = 16
							m := op.Record(gtx.Ops)
							wm.drawTabIcon(gtx, idx, isActive, iconSz)
							m.Stop().Add(gtx.Ops)
							return layout.Dimensions{Size: image.Point{X: iconSz, Y: iconSz}}
						}),
						layout.Rigid(layout.Spacer{Width: 10}.Layout),
						layout.Rigid(func(gtx layout.Context) layout.Dimensions {
							fw := font.Normal
							tc := theme.WithAlpha(theme.ColorPureBlack, 200)
							if isActive {
								fw = font.Bold
								tc = theme.ColorPureBlack
							}
							lbl := material.Label(wm.theme, 14, wm.tabs[idx])
							lbl.Color = tc
							lbl.Font.Weight = fw
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

// drawTabIcon draws a small geometric icon for each sidebar tab using Gio paths.
// No font dependency — always renders correctly.
func (wm *WindowManager) drawTabIcon(gtx layout.Context, idx int, active bool, sz int) {
	col := theme.WithAlpha(theme.ColorCyberCyan, 130)
	if active {
		col = theme.ColorCyberCyan
	}
	half := float32(sz / 2)
	switch idx {
	case 0: // Home — circle with dot center
		drawCircle(gtx, half, half, half-2, theme.WithAlpha(col, 60))
		drawCircle(gtx, half, half, 3, col)
	case 1: // Tous — three horizontal bars
		for row := 0; row < 3; row++ {
			y0 := row*5 + 2
			cl := clip.Rect{Min: image.Point{X: 1, Y: y0}, Max: image.Point{X: sz - 1, Y: y0 + 2}}.Push(gtx.Ops)
			paint.Fill(gtx.Ops, col)
			cl.Pop()
		}
	case 2: // A lire — bookmark shape
		var p clip.Path
		p.Begin(gtx.Ops)
		p.MoveTo(f32.Pt(3, 1))
		p.LineTo(f32.Pt(float32(sz-3), 1))
		p.LineTo(f32.Pt(float32(sz-3), float32(sz-1)))
		p.LineTo(f32.Pt(half, float32(sz-5)))
		p.LineTo(f32.Pt(3, float32(sz-1)))
		p.Close()
		paint.FillShape(gtx.Ops, theme.WithAlpha(col, 80), clip.Outline{Path: p.End()}.Op())
		drawCircle(gtx, half, 5, 2, col)
	case 3: // Termines — check mark
		var p clip.Path
		p.Begin(gtx.Ops)
		p.MoveTo(f32.Pt(2, 8))
		p.LineTo(f32.Pt(6, 13))
		p.LineTo(f32.Pt(14, 3))
		p.LineTo(f32.Pt(13, 2))
		p.LineTo(f32.Pt(6, 11))
		p.LineTo(f32.Pt(3, 7))
		p.Close()
		paint.FillShape(gtx.Ops, col, clip.Outline{Path: p.End()}.Op())
	case 4: // Fiches — document lines
		cl := clip.UniformRRect(image.Rectangle{Min: image.Point{X: 2, Y: 1}, Max: image.Point{X: sz - 2, Y: sz - 1}}, 2).Push(gtx.Ops)
		paint.Fill(gtx.Ops, theme.WithAlpha(col, 50))
		cl.Pop()
		for row := 0; row < 3; row++ {
			y0 := row*4 + 4
			cl2 := clip.Rect{Min: image.Point{X: 4, Y: y0}, Max: image.Point{X: sz - 4, Y: y0 + 1}}.Push(gtx.Ops)
			paint.Fill(gtx.Ops, col)
			cl2.Pop()
		}
	case 5: // Rappels — clock circle
		drawCircle(gtx, half, half, half-1, theme.WithAlpha(col, 50))
		drawCircle(gtx, half, half, 2, col)
		// Clock hands
		cl1 := clip.Rect{Min: image.Point{X: int(half) - 1, Y: int(half) - 4}, Max: image.Point{X: int(half) + 1, Y: int(half)}}.Push(gtx.Ops)
		paint.Fill(gtx.Ops, col)
		cl1.Pop()
		cl2 := clip.Rect{Min: image.Point{X: int(half), Y: int(half) - 1}, Max: image.Point{X: int(half) + 4, Y: int(half) + 1}}.Push(gtx.Ops)
		paint.Fill(gtx.Ops, col)
		cl2.Pop()
	case 6: // Partager — arrow up-right
		var p clip.Path
		p.Begin(gtx.Ops)
		p.MoveTo(f32.Pt(3, float32(sz-2)))
		p.LineTo(f32.Pt(float32(sz-2), 3))
		p.LineTo(f32.Pt(float32(sz-2), 8))
		p.LineTo(f32.Pt(float32(sz-2), 3))
		p.LineTo(f32.Pt(float32(sz-8), 3))
		p.Close()
		paint.FillShape(gtx.Ops, col, clip.Outline{Path: p.End()}.Op())
		drawCircle(gtx, 3, float32(sz-2), 2, col)
	case 7: // Metriques — bar chart
		bars := []struct{ x, h int }{{2, 8}, {6, 12}, {10, 6}, {14, 10}}
		for _, b := range bars {
			cl := clip.Rect{
				Min: image.Point{X: b.x, Y: sz - b.h},
				Max: image.Point{X: b.x + 3, Y: sz},
			}.Push(gtx.Ops)
			paint.Fill(gtx.Ops, col)
			cl.Pop()
		}
	default:
		drawCircle(gtx, half, half, 3, col)
	}
}

// ==========================================================
// DASHBOARD — loads real data from services
// ==========================================================

func (wm *WindowManager) loadDashboard() {
	if wm.trackSvc == nil || wm.libSvc == nil {
		wm.dashboardLoaded = true
		return
	}
	book, session, err := wm.trackSvc.GetMostRecentBook(context.Background())
	if err == nil {
		wm.currentBook = book
		wm.currentSession = session
	}
	wm.dashboardLoaded = true
}

func (wm *WindowManager) drawDashboard(gtx layout.Context) layout.Dimensions {
	if !wm.dashboardLoaded {
		wm.loadDashboard()
	}

	// Compute progress from real session
	var completion float32
	var centerLabel = "0%"
	var subLabel = "Aucune session"
	var bookTitle = "Aucun livre en cours"
	var authorLabel = "Importez un livre pour commencer"

	if wm.currentBook != nil && wm.currentSession != nil {
		completion = float32(wm.currentSession.CalculateCompletion()) / 100.0
		centerLabel = fmt.Sprintf("%.0f%%", wm.currentSession.CalculateCompletion())
		subLabel = fmt.Sprintf("Page %d / %d", wm.currentSession.CurrentPage, wm.currentSession.TotalPages)
		bookTitle = wm.currentBook.Title
		authorLabel = wm.currentBook.Author
		if authorLabel == "" {
			authorLabel = "Auteur inconnu"
		}
	}

	// CTA button handler
	if wm.continueReadBtn.Clicked(gtx) && wm.currentBook != nil {
		wm.openBookInReader(wm.currentBook)
	}

	return layout.Flex{Axis: layout.Vertical}.Layout(gtx,

		// Top section: cover | speedometer | info + CTA
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return layout.Flex{Axis: layout.Horizontal, Alignment: layout.Middle}.Layout(gtx,

				// Book cover
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					w, h := 180, 270
					// Shadow
					sr := clip.UniformRRect(image.Rectangle{
						Min: image.Point{X: 6, Y: 10},
						Max: image.Point{X: w + 6, Y: h + 12},
					}, 10).Push(gtx.Ops)
					paint.Fill(gtx.Ops, theme.WithAlpha(theme.ColorPureBlack, 18))
					sr.Pop()
					// Cover
					cr := clip.UniformRRect(image.Rectangle{Max: image.Point{X: w, Y: h}}, 10).Push(gtx.Ops)
					paint.Fill(gtx.Ops, theme.ColorCyberCyan)
					cr.Pop()
					// Top sheen
					sheen := clip.UniformRRect(image.Rectangle{Max: image.Point{X: w, Y: 55}}, 10).Push(gtx.Ops)
					paint.Fill(gtx.Ops, theme.WithAlpha(theme.ColorGlassWhite, 14))
					sheen.Pop()
					// Book title on cover
					stack := op.Offset(image.Pt(12, 14)).Push(gtx.Ops)
					gtx2 := gtx
					gtx2.Constraints.Max.X = w - 24
					lbl := material.Label(wm.theme, 13, bookTitle)
					lbl.Color = theme.ColorGlassWhite
					lbl.Font.Weight = font.Bold
					lbl.Layout(gtx2)
					stack.Pop()
					return layout.Dimensions{Size: image.Point{X: w + 20, Y: h + 16}}
				}),

				layout.Rigid(layout.Spacer{Width: 48}.Layout),

				// Speedometer (circular progress gauge)
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					return wm.drawSpeedometer(gtx, completion, centerLabel, subLabel)
				}),

				layout.Rigid(layout.Spacer{Width: 48}.Layout),

				// Info panel + CTA
				layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
					return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
						layout.Rigid(func(gtx layout.Context) layout.Dimensions {
							lbl := material.Label(wm.theme, 12, "EN COURS DE LECTURE")
							lbl.Color = theme.WithAlpha(theme.ColorCyberCyan, 160)
							lbl.Font.Weight = font.Bold
							return layout.Inset{Bottom: 10}.Layout(gtx, lbl.Layout)
						}),
						layout.Rigid(func(gtx layout.Context) layout.Dimensions {
							lbl := material.H4(wm.theme, bookTitle)
							lbl.Font.Weight = font.Bold
							lbl.Color = theme.ColorPureBlack
							return layout.Inset{Bottom: 6}.Layout(gtx, lbl.Layout)
						}),
						layout.Rigid(func(gtx layout.Context) layout.Dimensions {
							lbl := material.Label(wm.theme, 14, authorLabel)
							lbl.Color = theme.WithAlpha(theme.ColorPureBlack, 150)
							return layout.Inset{Bottom: 36}.Layout(gtx, lbl.Layout)
						}),
						// "Continue" CTA — glow button, uses real currentBook
						layout.Rigid(func(gtx layout.Context) layout.Dimensions {
							label := "Continuer la lecture"
							if wm.currentBook == nil {
								label = "Importer un livre"
							}
							return wm.drawGlowCTA(gtx, label, &wm.continueReadBtn, theme.ColorCyberCyan)
						}),
					)
				}),
			)
		}),

		layout.Flexed(1, layout.Spacer{}.Layout),

		// Bottom stat cards
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			// Compute real stats from sessions if available
			totalPages := 0
			booksFinished := 0
			sessionCount := 0
			if !wm.metricsLoaded {
				wm.loadMetrics()
			}
			for _, s := range wm.recentSessions {
				totalPages += s.CurrentPage
				sessionCount++
				if s.IsBookComplete() {
					booksFinished++
				}
			}
			avgMin := 0
			if sessionCount > 0 {
				avgMin = (totalPages * 2) / sessionCount // rough estimate: 2 min/page
			}
			avgStr := fmt.Sprintf("%d min / jour", avgMin)
			if avgMin == 0 {
				avgStr = "—"
			}
			pagesStr := fmt.Sprintf("%d pages", totalPages)
			if totalPages == 0 {
				pagesStr = "—"
			}
			finStr := fmt.Sprintf("%d livres", booksFinished)
			if booksFinished == 0 {
				finStr = "—"
			}

			return layout.Flex{Axis: layout.Horizontal}.Layout(gtx,
				layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
					return wm.drawAnalyticCard(gtx, "Temps moyen estimé", avgStr, "")
				}),
				layout.Rigid(layout.Spacer{Width: 24}.Layout),
				layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
					return wm.drawAnalyticCard(gtx, "Livres terminés", finStr, "")
				}),
				layout.Rigid(layout.Spacer{Width: 24}.Layout),
				layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
					return wm.drawAnalyticCard(gtx, "Pages lues au total", pagesStr, "")
				}),
			)
		}),
	)
}

// drawSpeedometer — circular arc gauge with gold glow
func (wm *WindowManager) drawSpeedometer(gtx layout.Context, completion float32, centerLabel, subLabel string) layout.Dimensions {
	const size = 260
	cx, cy := float32(size/2), float32(size/2)
	r := float32(105)
	thickness := float32(16)
	startDeg := float32(135)
	sweepDeg := float32(270)

	gtx.Constraints = layout.Exact(image.Point{X: size, Y: size})

	endTrack := startDeg + sweepDeg
	endProg := startDeg + sweepDeg*completion

	// Track (background arc)
	drawArcThick(gtx, cx, cy, r, thickness, startDeg, endTrack, theme.WithAlpha(theme.ColorCyberCyan, 16))

	if completion > 0.005 {
		// Concentric glow halos
		drawArcThick(gtx, cx, cy, r, thickness+34, startDeg, endProg, theme.WithAlpha(theme.ColorSandGold, 4))
		drawArcThick(gtx, cx, cy, r, thickness+20, startDeg, endProg, theme.WithAlpha(theme.ColorSandGold, 10))
		drawArcThick(gtx, cx, cy, r, thickness+8, startDeg, endProg, theme.WithAlpha(theme.ColorSandGold, 24))
		// Main arc
		drawArcThick(gtx, cx, cy, r, thickness, startDeg, endProg, theme.ColorSandGold)
		// Glowing dot at tip
		endRad := float64(endProg) * math.Pi / 180
		dotX := cx + r*float32(math.Cos(endRad))
		dotY := cy + r*float32(math.Sin(endRad))
		drawCircle(gtx, dotX, dotY, 18, theme.WithAlpha(theme.ColorSandGold, 22))
		drawCircle(gtx, dotX, dotY, 11, theme.WithAlpha(theme.ColorSandGold, 50))
		drawCircle(gtx, dotX, dotY, 7, theme.ColorSandGold)
	}

	// Centre labels
	layout.Center.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		return layout.Flex{Axis: layout.Vertical, Alignment: layout.Middle}.Layout(gtx,
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				lbl := material.Label(wm.theme, 38, centerLabel)
				lbl.Font.Weight = font.Bold
				lbl.Color = theme.ColorPureBlack
				lbl.Alignment = text.Middle
				return lbl.Layout(gtx)
			}),
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				lbl := material.Label(wm.theme, 12, subLabel)
				lbl.Color = theme.WithAlpha(theme.ColorPureBlack, 140)
				lbl.Alignment = text.Middle
				return layout.Inset{Top: 4}.Layout(gtx, lbl.Layout)
			}),
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				lbl := material.Label(wm.theme, 10, "COMPLETE")
				lbl.Color = theme.WithAlpha(theme.ColorCyberCyan, 160)
				lbl.Font.Weight = font.Bold
				lbl.Alignment = text.Middle
				return layout.Inset{Top: 2}.Layout(gtx, lbl.Layout)
			}),
		)
	})

	return layout.Dimensions{Size: image.Point{X: size, Y: size}}
}

// ==========================================================
// SHARING VIEW
// ==========================================================

func (wm *WindowManager) drawSharingView(gtx layout.Context) layout.Dimensions {
	return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			lbl := material.H5(wm.theme, "Partager")
			lbl.Font.Weight = font.Bold
			return layout.Inset{Bottom: 8}.Layout(gtx, lbl.Layout)
		}),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			lbl := material.Label(wm.theme, 14, "Exportez votre bibliotheque et vos fiches dans le dossier de votre choix.")
			lbl.Color = theme.WithAlpha(theme.ColorPureBlack, 155)
			return layout.Inset{Bottom: 32}.Layout(gtx, lbl.Layout)
		}),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return wm.drawShareCard(gtx,
				"Bibliotheque complete — Markdown",
				"Un fichier .md lisible avec tous vos livres et fiches.",
				&wm.shareLibBtn,
				func() {
					if wm.sharingSvc == nil {
						wm.shareStatusMsg = "Service non disponible."
						return
					}
					dir := service.PickExportDirectory()
					path, err := wm.sharingSvc.ExportLibrary(context.Background(), service.ShareFormatMarkdown, dir)
					if err != nil {
						wm.shareStatusMsg = "Erreur : " + err.Error()
					} else {
						wm.shareStatusMsg = "Exporte -> " + path
					}
				},
			)
		}),
		layout.Rigid(layout.Spacer{Height: 16}.Layout),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return wm.drawShareCard(gtx,
				"Bibliotheque complete — JSON",
				"Format structure ideal pour un backup ou une integration.",
				&wm.shareExportBtn,
				func() {
					if wm.sharingSvc == nil {
						wm.shareStatusMsg = "Service non disponible."
						return
					}
					dir := service.PickExportDirectory()
					path, err := wm.sharingSvc.ExportLibrary(context.Background(), service.ShareFormatJSON, dir)
					if err != nil {
						wm.shareStatusMsg = "Erreur : " + err.Error()
					} else {
						wm.shareStatusMsg = "Exporte -> " + path
					}
				},
			)
		}),
		layout.Rigid(layout.Spacer{Height: 28}.Layout),
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
	cl := clip.UniformRRect(image.Rectangle{Max: image.Point{X: gtx.Constraints.Max.X, Y: 110}}, 10).Push(gtx.Ops)
	paint.Fill(gtx.Ops, theme.WithAlpha(theme.ColorCyberCyan, 7))
	cl.Pop()
	return layout.Inset{Top: 20, Left: 24, Right: 24, Bottom: 20}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		return layout.Flex{Axis: layout.Horizontal, Alignment: layout.Middle}.Layout(gtx,
			layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
				return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
					layout.Rigid(func(gtx layout.Context) layout.Dimensions {
						lbl := material.Label(wm.theme, 15, title)
						lbl.Font.Weight = font.Bold
						return layout.Inset{Bottom: 5}.Layout(gtx, lbl.Layout)
					}),
					layout.Rigid(func(gtx layout.Context) layout.Dimensions {
						lbl := material.Label(wm.theme, 13, desc)
						lbl.Color = theme.WithAlpha(theme.ColorPureBlack, 150)
						return lbl.Layout(gtx)
					}),
				)
			}),
			layout.Rigid(layout.Spacer{Width: 20}.Layout),
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				if btn.Clicked(gtx) {
					go onPress()
					wm.shareStatusMsg = "Selection du dossier..."
				}
				return wm.drawPillButton(gtx, "Exporter", btn, theme.ColorSandGold)
			}),
		)
	})
}

// ==========================================================
// METRICS — real session data
// ==========================================================

func (wm *WindowManager) loadMetrics() {
	if wm.trackSvc == nil {
		wm.metricsLoaded = true
		return
	}
	sessions, err := wm.trackSvc.GetRecentSessions(context.Background())
	if err == nil {
		wm.recentSessions = sessions
	}
	wm.metricsLoaded = true
}

func (wm *WindowManager) computeMetrics() (bestDay, bestHour string, sessions []*domain.ReadingSession) {
	if !wm.metricsLoaded {
		wm.loadMetrics()
	}
	sessions = wm.recentSessions

	if len(sessions) == 0 {
		return "—", "—", nil
	}

	daysCount := make(map[time.Weekday]int)
	hoursCount := make(map[int]int)
	for _, s := range sessions {
		daysCount[s.LastReadingTime.Weekday()]++
		hoursCount[s.LastReadingTime.Hour()]++
	}
	var maxDay time.Weekday
	maxDC := 0
	for d, c := range daysCount {
		if c > maxDC {
			maxDC = c
			maxDay = d
		}
	}
	var maxHour, maxHC int
	for h, c := range hoursCount {
		if c > maxHC {
			maxHC = c
			maxHour = h
		}
	}
	frDays := map[time.Weekday]string{
		time.Sunday: "Dimanche", time.Monday: "Lundi", time.Tuesday: "Mardi",
		time.Wednesday: "Mercredi", time.Thursday: "Jeudi",
		time.Friday: "Vendredi", time.Saturday: "Samedi",
	}
	bestDay = frDays[maxDay]
	bestHour = fmt.Sprintf("%dh00 — %dh00", maxHour, maxHour+2)
	sort.Slice(sessions, func(i, j int) bool {
		return sessions[i].LastReadingTime.After(sessions[j].LastReadingTime)
	})
	return
}

func (wm *WindowManager) drawMetrics(gtx layout.Context) layout.Dimensions {
	if !wm.metricsLoaded {
		wm.loadMetrics()
	}
	bestDay, bestHour, history := wm.computeMetrics()

	// Compute totals
	totalPages := 0
	totalSessions := len(history)
	completedBooks := 0
	uniqueBooks := map[string]bool{}
	for _, s := range history {
		totalPages += s.CurrentPage
		uniqueBooks[s.BookID] = true
		if s.IsBookComplete() {
			completedBooks++
		}
	}

	return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
		// Title
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			lbl := material.H5(wm.theme, "Statistiques de lecture")
			lbl.Font.Weight = font.Bold
			return layout.Inset{Bottom: 28}.Layout(gtx, lbl.Layout)
		}),

		// Top stat cards row
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return layout.Flex{Axis: layout.Horizontal}.Layout(gtx,
				layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
					return wm.drawAnalyticCard(gtx, "Sessions totales", fmt.Sprintf("%d", totalSessions), "Toutes les ouvertures de livres")
				}),
				layout.Rigid(layout.Spacer{Width: 20}.Layout),
				layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
					return wm.drawAnalyticCard(gtx, "Pages lues", fmt.Sprintf("%d", totalPages), "Cumul de toutes les sessions")
				}),
				layout.Rigid(layout.Spacer{Width: 20}.Layout),
				layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
					return wm.drawAnalyticCard(gtx, "Livres termines", fmt.Sprintf("%d", completedBooks), "Sessions arrivees en fin")
				}),
				layout.Rigid(layout.Spacer{Width: 20}.Layout),
				layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
					return wm.drawAnalyticCard(gtx, "Livres distincts", fmt.Sprintf("%d", len(uniqueBooks)), "")
				}),
			)
		}),
		layout.Rigid(layout.Spacer{Height: 24}.Layout),

		// Patterns row
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return layout.Flex{Axis: layout.Horizontal}.Layout(gtx,
				layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
					return wm.drawAnalyticCard(gtx, "Jour le plus actif", bestDay, "Vos sessions sont plus longues ce jour-la.")
				}),
				layout.Rigid(layout.Spacer{Width: 20}.Layout),
				layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
					return wm.drawAnalyticCard(gtx, "Heure de predilection", bestHour, "Vous lisez principalement dans cette tranche.")
				}),
			)
		}),
		layout.Rigid(layout.Spacer{Height: 32}.Layout),

		// Session history list
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			lbl := material.H6(wm.theme, "Historique des sessions")
			lbl.Font.Weight = font.Bold
			return layout.Inset{Bottom: 14}.Layout(gtx, lbl.Layout)
		}),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			if len(history) == 0 {
				lbl := material.Label(wm.theme, 14, "Aucune session enregistree. Commencez a lire !")
				lbl.Color = theme.WithAlpha(theme.ColorPureBlack, 130)
				return lbl.Layout(gtx)
			}
			var rows []layout.FlexChild
			limit := len(history)
			if limit > 10 {
				limit = 10
			}
			for _, s := range history[:limit] {
				session := s
				row := layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					cl := clip.UniformRRect(image.Rectangle{Max: image.Point{X: gtx.Constraints.Max.X, Y: 50}}, 6).Push(gtx.Ops)
					paint.Fill(gtx.Ops, theme.WithAlpha(theme.ColorCyberCyan, 7))
					cl.Pop()
					return layout.Inset{Top: 10, Bottom: 10, Left: 16, Right: 16}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
						return layout.Flex{Axis: layout.Horizontal, Alignment: layout.Middle}.Layout(gtx,
							layout.Flexed(0.4, func(gtx layout.Context) layout.Dimensions {
								// Try to look up book title
								title := session.BookID[:min(8, len(session.BookID))] + "..."
								for _, b := range wm.books {
									if b.ID == session.BookID {
										title = b.Title
										break
									}
								}
								lbl := material.Label(wm.theme, 14, title)
								lbl.Font.Weight = font.SemiBold
								return lbl.Layout(gtx)
							}),
							layout.Flexed(0.3, func(gtx layout.Context) layout.Dimensions {
								lbl := material.Label(wm.theme, 13, session.LastReadingTime.Format("02 Jan 15:04"))
								lbl.Color = theme.WithAlpha(theme.ColorPureBlack, 140)
								return lbl.Layout(gtx)
							}),
							layout.Flexed(0.15, func(gtx layout.Context) layout.Dimensions {
								lbl := material.Label(wm.theme, 14, fmt.Sprintf("p.%d", session.CurrentPage))
								lbl.Color = theme.ColorCyberCyan
								lbl.Font.Weight = font.Bold
								return lbl.Layout(gtx)
							}),
							layout.Flexed(0.15, func(gtx layout.Context) layout.Dimensions {
								pct := session.CalculateCompletion()
								lbl := material.Label(wm.theme, 12, fmt.Sprintf("%.0f%%", pct))
								lbl.Color = theme.ColorSandGold
								lbl.Font.Weight = font.Bold
								return lbl.Layout(gtx)
							}),
						)
					})
				})
				rows = append(rows, row, layout.Rigid(layout.Spacer{Height: 6}.Layout))
			}
			return layout.Flex{Axis: layout.Vertical}.Layout(gtx, rows...)
		}),
	)
}

func (wm *WindowManager) drawAnalyticCard(gtx layout.Context, title, mainValue, subtitle string) layout.Dimensions {
	cl := clip.UniformRRect(image.Rectangle{Max: image.Point{X: gtx.Constraints.Max.X, Y: 110}}, 8).Push(gtx.Ops)
	paint.Fill(gtx.Ops, theme.WithAlpha(theme.ColorSandGold, 16))
	cl.Pop()
	return layout.Inset{Top: 20, Left: 20, Right: 20}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				lbl := material.Label(wm.theme, 13, title)
				lbl.Color = theme.WithAlpha(theme.ColorPureBlack, 175)
				return lbl.Layout(gtx)
			}),
			layout.Rigid(layout.Spacer{Height: 8}.Layout),
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				lbl := material.Label(wm.theme, 18, mainValue)
				lbl.Font.Weight = font.Bold
				lbl.Color = theme.ColorPureBlack
				return lbl.Layout(gtx)
			}),
			layout.Rigid(layout.Spacer{Height: 4}.Layout),
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
// HELPERS
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

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
