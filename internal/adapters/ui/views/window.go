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
	"gioui.org/io/key"
	"gioui.org/io/system"
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
	"github.com/MiltonJ23/Orus/internal/port"
	"github.com/MiltonJ23/Orus/internal/service"
)

type AppState int

const (
	StateSplash AppState = iota
	StateHome
)

// WindowManager is the root UI controller.
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

	// macOS traffic lights
	btnClose widget.Clickable
	btnMin   widget.Clickable
	btnMax   widget.Clickable

	// Sidebar
	tabs      []string
	tabClicks []widget.Clickable
	activeTab int

	// Search
	searchEditor widget.Editor
	searchQuery  string // current live filter

	// Library
	books           []*domain.Book
	booksLoaded     bool
	gridList        widget.List
	bookOpenBtns    []widget.Clickable
	overlayReadBtns []widget.Clickable // separate from card buttons
	importBtn       widget.Clickable
	importStatusMsg string

	// Book completion cache — "unread" | "reading" | "done"
	bookStatus       map[string]string
	bookStatusLoaded bool

	// Dashboard
	continueReadBtn widget.Clickable
	dashboardLoaded bool
	currentBook     *domain.Book
	currentSession  *domain.ReadingSession

	// Sheets
	sheets            []*domain.ReadingSheet
	sheetsLoaded      bool
	sheetForm         sheetFormState
	sheetPickerSearch widget.Editor
	sheetPickerList   widget.List

	// Sheet detail
	activeSheetDetail     *domain.ReadingSheet
	sheetDetailBtns       []widget.Clickable
	sheetDetailScrollList widget.List
	closeSheetDetailBtn   widget.Clickable
	shareSheetBtn         widget.Clickable
	shareSheetStatus      string
	// Book completion achievement modal
	achievementBook       *domain.Book
	achievementReadMin    int
	achievementDismiss    widget.Clickable
	achievementShare      widget.Clickable
	achievementTwitter    widget.Clickable
	achievementLinkedIn   widget.Clickable
	confettiStart         time.Time // when confetti burst started
	confettiActive        bool
	dismissedAchievements map[string]bool // bookID → already shown; never re-show

	// Social share buttons in detail view
	shareTwitterBtn widget.Clickable
	shareCopyBtn    widget.Clickable

	// Sheet filter
	sheetFilterIdx  int
	sheetFilterBtns []widget.Clickable

	// Book selector inside sheet form
	bookSelectBtns  []widget.Clickable
	selectedBookIdx int

	// Reminders
	reminders       []*domain.Reminder
	remindersLoaded bool
	reminderForm    reminderFormState

	// In-app reminder banner
	activeReminder    *domain.Reminder
	reminderBannerBtn widget.Clickable

	// Sharing
	shareStatusMsg string
	shareLibBtn    widget.Clickable
	shareExportBtn widget.Clickable

	// Book card action overlay (cover click → slide → archive/delete)
	activeBookCardIdx  int // -1 = none
	bookCoverClickBtns []widget.Clickable
	bookArchiveBtns    []widget.Clickable
	bookDeleteBtns     []widget.Clickable
	bookMenuCloseBtn   widget.Clickable // backdrop click to dismiss
	bookCardAnimStart  time.Time
	bookCardAnimProg   float32 // 0..1

	// uiChan carries UI-state mutations from goroutines to the main render loop.
	// All writes to WindowManager fields MUST go through this channel to avoid
	// data races with Gio's single-threaded frame model.
	uiChan chan func()

	// Reader
	readerOpenedAt   time.Time
	readerScrollList widget.List // vertical scroll within a reader page
	readerActive     bool
	readerBook       *domain.Book
	readerSession    *domain.ReadingSession // live session for progress saving
	readerContent    []string
	readerPage       int
	readerFontSize   float32
	readerDimAlpha   uint8
	readerLoading    bool
	closeReaderBtn   widget.Clickable
	fontPlusBtn      widget.Clickable
	fontMinusBtn     widget.Clickable
	dimPlusBtn       widget.Clickable
	dimMinusBtn      widget.Clickable
	readerPrevBtn    widget.Clickable
	readerNextBtn    widget.Clickable

	// Reader background (color palette + XMB animated mode)
	// Mode: 0=light 1=dark 2=xmb 3..9=preset colors
	readerBgMode      int
	readerBgAnimStart time.Time
	readerBgPanelOpen bool
	readerBgPanelBtn  widget.Clickable
	readerBgBtns      [9]widget.Clickable

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
		window:                new(app.Window),
		theme:                 th,
		libSvc:                lib,
		trackSvc:              track,
		sheetSvc:              sheet,
		reminderSvc:           reminder,
		sharingSvc:            sharing,
		contentReader:         contentReader,
		state:                 StateSplash,
		appStartTime:          time.Now(),
		logo:                  logoImg,
		tabs:                  menuTabs,
		tabClicks:             clicks,
		activeTab:             0,
		searchEditor:          widget.Editor{SingleLine: true, Submit: true},
		readerFontSize:        16,
		gridList:              widget.List{List: layout.List{Axis: layout.Vertical}},
		sheetPickerList:       widget.List{List: layout.List{Axis: layout.Vertical}},
		sheetDetailScrollList: widget.List{List: layout.List{Axis: layout.Vertical}},
		selectedBookIdx:       -1,
		activeBookCardIdx:     -1,
		overlayReadBtns:       []widget.Clickable{},
		readerBgMode:          0,
		uiChan:                make(chan func(), 128),
	}

	if reminder != nil {
		reminder.SetCallback(func(r *domain.Reminder) {
			wm.uiChan <- func() {
				wm.activeReminder = r
				wm.window.Invalidate()
			}
		})
	}
	return wm
}

func (wm *WindowManager) Run() error {
	wm.window.Option(app.Title("Orus"), app.Size(1440, 900), app.MinSize(980, 640), app.Decorated(false))
	var ops op.Ops
	for {
		switch e := wm.window.Event().(type) {
		case app.DestroyEvent:
			return e.Err
		case app.FrameEvent:
			// Drain all pending UI mutations from goroutines — runs on main thread.
			for drained := false; !drained; {
				select {
				case fn := <-wm.uiChan:
					fn()
				default:
					drained = true
				}
			}
			gtx := app.NewContext(&ops, e)
			elapsed := time.Since(wm.appStartTime).Seconds()
			if wm.state == StateSplash && elapsed > 6.5 {
				wm.state = StateHome
			}
			if wm.state == StateSplash {
				wm.layoutSplashScreen(gtx, elapsed)
				wm.window.Invalidate()
			} else {
				// Live search: capture editor changes every frame
				wm.searchQuery = wm.searchEditor.Text()
				wm.layoutHomeScreen(gtx)
			}
			e.Frame(gtx.Ops)
		case key.Event:
			if e.State == key.Press {
				// ESC closes overlay or reader
				if e.Name == key.NameEscape {
					if wm.achievementBook != nil {
						if wm.dismissedAchievements == nil {
							wm.dismissedAchievements = make(map[string]bool)
						}
						wm.dismissedAchievements[wm.achievementBook.ID] = true
						wm.achievementBook = nil
						wm.confettiActive = false
						wm.window.Invalidate()
					} else if wm.activeBookCardIdx >= 0 {
						wm.closeBookCardMenu()
						wm.window.Invalidate()
					} else if wm.activeSheetDetail != nil {
						wm.activeSheetDetail = nil
						wm.window.Invalidate()
					} else if wm.sheetForm.showForm {
						wm.sheetForm.showForm = false
						wm.window.Invalidate()
					} else if wm.readerActive {
						wm.closeReader()
						wm.window.Invalidate()
					}
				}
				if wm.readerActive {
					total := len(wm.readerContent)
					switch e.Name {
					case key.NameLeftArrow:
						if wm.readerPage > 0 {
							wm.readerPage--
							wm.saveReaderProgress()
							wm.window.Invalidate()
						}
					case key.NameRightArrow:
						if total > 0 && wm.readerPage < total-1 {
							wm.readerPage++
							wm.saveReaderProgress()
							wm.window.Invalidate()
						}
					}
				}
			}
		}
	}
}

// ==========================================================
// SHAPE PRIMITIVES
// ==========================================================

func drawArcThick(gtx layout.Context, cx, cy, r, thickness, startDeg, endDeg float32, c color.NRGBA) {
	if endDeg <= startDeg {
		return
	}
	startRad := float64(startDeg) * math.Pi / 180
	endRad := float64(endDeg) * math.Pi / 180
	inner := r - thickness/2
	outer := r + thickness/2
	var p clip.Path
	p.Begin(gtx.Ops)
	for i := 0; i <= 80; i++ {
		t := float64(i) / 80
		angle := startRad + (endRad-startRad)*t
		x := cx + outer*float32(math.Cos(angle))
		y := cy + outer*float32(math.Sin(angle))
		if i == 0 {
			p.MoveTo(f32.Pt(x, y))
		} else {
			p.LineTo(f32.Pt(x, y))
		}
	}
	for i := 80; i >= 0; i-- {
		t := float64(i) / 80
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
		{R: 255, G: 95, B: 86, A: 255},
		{R: 255, G: 189, B: 46, A: 255},
		{R: 39, G: 201, B: 63, A: 255},
	}
	btns := []*widget.Clickable{&wm.btnClose, &wm.btnMin, &wm.btnMax}
	for i, btn := range btns {
		b := btn
		col := colors[i]
		area := 20
		bx := int(sx+float32(i)*spacing) - area/2
		by := int(sy) - area/2
		// NOTE: immediate push/pop, never defer in a loop
		stack := op.Offset(image.Pt(bx, by)).Push(gtx.Ops)
		b.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
			c := col
			if b.Hovered() {
				c = theme.WithAlpha(col, 195)
			}
			if b.Pressed() {
				c = theme.WithAlpha(col, 140)
			}
			drawCircle(gtx, float32(area/2), float32(area/2), radius, c)
			return layout.Dimensions{Size: image.Point{X: area, Y: area}}
		})
		stack.Pop()
	}
}

// ==========================================================
// REUSABLE WIDGETS
// ==========================================================

// drawPillButton — natural-width solid pill button
func (wm *WindowManager) drawPillButton(gtx layout.Context, label string, btn *widget.Clickable, col color.NRGBA) layout.Dimensions {
	return btn.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		bgA := uint8(210)
		if btn.Hovered() {
			bgA = 240
		}
		if btn.Pressed() {
			bgA = 160
		}
		const h = 36
		const hPad = 20

		// Measure text width using op.Record (discard draw, keep dims)
		gtxM := gtx
		gtxM.Constraints = layout.Constraints{Max: image.Point{X: 2000, Y: h}}
		macro := op.Record(gtxM.Ops)
		lblM := material.Label(wm.theme, 13, label)
		lblM.Font.Weight = font.SemiBold
		textDims := lblM.Layout(gtxM)
		macro.Stop() // discard — we only needed the size
		w := textDims.Size.X + hPad*2

		// Shadow
		sh := clip.UniformRRect(image.Rectangle{Min: image.Pt(1, 2), Max: image.Pt(w+1, h+2)}, h/2).Push(gtx.Ops)
		paint.Fill(gtx.Ops, color.NRGBA{A: 25})
		sh.Pop()
		// Body
		cl := clip.UniformRRect(image.Rectangle{Max: image.Pt(w, h)}, h/2).Push(gtx.Ops)
		paint.Fill(gtx.Ops, theme.WithAlpha(col, bgA))
		cl.Pop()
		// Shine
		shine := clip.UniformRRect(image.Rectangle{Max: image.Pt(w, h/2)}, h/2).Push(gtx.Ops)
		paint.Fill(gtx.Ops, color.NRGBA{R: 255, G: 255, B: 255, A: 22})
		shine.Pop()
		// Label
		gtx2 := gtx
		gtx2.Constraints = layout.Exact(image.Pt(w, h))
		layout.Center.Layout(gtx2, func(gtx layout.Context) layout.Dimensions {
			lbl := material.Label(wm.theme, 13, label)
			lbl.Color = color.NRGBA{R: 255, G: 255, B: 255, A: 255}
			lbl.Font.Weight = font.SemiBold
			return lbl.Layout(gtx)
		})
		return layout.Dimensions{Size: image.Pt(w, h)}
	})
}

// drawGlowCTA — premium CTA with layered glow and shine
func (wm *WindowManager) drawGlowCTA(gtx layout.Context, label string, btn *widget.Clickable, col color.NRGBA) layout.Dimensions {
	const btnW, btnH = 180, 44
	const glowPad = 10
	totalW := btnW + glowPad*2
	totalH := btnH + glowPad*2

	return btn.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		// Outer bloom layers on hover
		if btn.Hovered() {
			for gi, g := range []struct{ pad, alpha int }{{12, 6}, {8, 12}, {4, 22}, {1, 10}} {
				_ = gi
				cl := clip.UniformRRect(image.Rectangle{
					Min: image.Pt(g.pad, g.pad),
					Max: image.Pt(totalW-g.pad, totalH-g.pad),
				}, 24+g.pad).Push(gtx.Ops)
				paint.Fill(gtx.Ops, theme.WithAlpha(col, uint8(g.alpha)))
				cl.Pop()
			}
		}

		stack := op.Offset(image.Pt(glowPad, glowPad)).Push(gtx.Ops)

		bgA := uint8(230)
		if btn.Hovered() {
			bgA = 255
		}
		if btn.Pressed() {
			bgA = 175
		}

		// Shadow
		shcl := clip.UniformRRect(image.Rectangle{
			Min: image.Pt(2, 4), Max: image.Pt(btnW+2, btnH+4),
		}, 22).Push(gtx.Ops)
		paint.Fill(gtx.Ops, color.NRGBA{A: 40})
		shcl.Pop()

		// Body
		btnR := clip.UniformRRect(image.Rectangle{Max: image.Pt(btnW, btnH)}, 22).Push(gtx.Ops)
		paint.Fill(gtx.Ops, theme.WithAlpha(col, bgA))
		btnR.Pop()

		// Top half shine
		shineR := clip.UniformRRect(image.Rectangle{Max: image.Pt(btnW, btnH/2)}, 22).Push(gtx.Ops)
		paint.Fill(gtx.Ops, color.NRGBA{R: 255, G: 255, B: 255, A: 22})
		shineR.Pop()

		// Label centered
		gtxBtn := gtx
		gtxBtn.Constraints = layout.Exact(image.Pt(btnW, btnH))
		layout.Center.Layout(gtxBtn, func(gtx layout.Context) layout.Dimensions {
			lbl := material.Label(wm.theme, 14, label)
			lbl.Color = color.NRGBA{R: 255, G: 255, B: 255, A: 255}
			lbl.Font.Weight = font.Bold
			return lbl.Layout(gtx)
		})
		stack.Pop()
		return layout.Dimensions{Size: image.Pt(totalW, totalH)}
	})
}

// drawLabeledField — text field with label
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
							paint.FillShape(gtx.Ops, theme.WithAlpha(theme.ColorGlassWhite, veil),
								clip.Rect{Max: wm.logo.Bounds().Size()}.Op())
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
	if wm.activeSheetDetail != nil {
		wm.drawSheetDetailView(gtx)
		wm.drawMacControls(gtx)
		return layout.Dimensions{Size: gtx.Constraints.Max}
	}
	if wm.readerActive {
		wm.drawReaderView(gtx)
		// Achievement fires from within the reader — render it on top immediately.
		if wm.achievementBook != nil {
			wm.drawAchievementModal(gtx)
		}
		wm.drawMacControls(gtx)
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
	if wm.achievementBook != nil {
		wm.drawAchievementModal(gtx)
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

// drawReminderBanner — top gold banner with dismiss
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
// SIDEBAR
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
			// Refresh status cache on tab switch to library sections
			if idx >= 1 && idx <= 3 {
				wm.bookStatusLoaded = false
			}
		}
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
			isActive := wm.activeTab == idx
			return wm.tabClicks[idx].Layout(gtx, func(gtx layout.Context) layout.Dimensions {
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

// drawTabIcon draws a geometric sidebar icon (no emoji/unicode dependency).
func (wm *WindowManager) drawTabIcon(gtx layout.Context, idx int, active bool, sz int) {
	col := theme.WithAlpha(theme.ColorCyberCyan, 130)
	if active {
		col = theme.ColorCyberCyan
	}
	half := float32(sz / 2)
	switch idx {
	case 0: // Home
		drawCircle(gtx, half, half, half-2, theme.WithAlpha(col, 60))
		drawCircle(gtx, half, half, 3, col)
	case 1: // Tous — three bars
		for row := 0; row < 3; row++ {
			y0 := row*5 + 2
			cl := clip.Rect{Min: image.Point{X: 1, Y: y0}, Max: image.Point{X: sz - 1, Y: y0 + 2}}.Push(gtx.Ops)
			paint.Fill(gtx.Ops, col)
			cl.Pop()
		}
	case 2: // A lire — bookmark
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
	case 3: // Termines — checkmark
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
	case 4: // Fiches — document
		cl := clip.UniformRRect(image.Rectangle{Min: image.Point{X: 2, Y: 1}, Max: image.Point{X: sz - 2, Y: sz - 1}}, 2).Push(gtx.Ops)
		paint.Fill(gtx.Ops, theme.WithAlpha(col, 50))
		cl.Pop()
		for row := 0; row < 3; row++ {
			y0 := row*4 + 4
			cl2 := clip.Rect{Min: image.Point{X: 4, Y: y0}, Max: image.Point{X: sz - 4, Y: y0 + 1}}.Push(gtx.Ops)
			paint.Fill(gtx.Ops, col)
			cl2.Pop()
		}
	case 5: // Rappels — clock
		drawCircle(gtx, half, half, half-1, theme.WithAlpha(col, 50))
		drawCircle(gtx, half, half, 2, col)
		cl1 := clip.Rect{Min: image.Point{X: int(half) - 1, Y: int(half) - 4}, Max: image.Point{X: int(half) + 1, Y: int(half)}}.Push(gtx.Ops)
		paint.Fill(gtx.Ops, col)
		cl1.Pop()
		cl2 := clip.Rect{Min: image.Point{X: int(half), Y: int(half) - 1}, Max: image.Point{X: int(half) + 4, Y: int(half) + 1}}.Push(gtx.Ops)
		paint.Fill(gtx.Ops, col)
		cl2.Pop()
	case 6: // Partager — arrow
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
		for _, b := range []struct{ x, h int }{{2, 8}, {6, 12}, {10, 6}, {14, 10}} {
			cl := clip.Rect{Min: image.Point{X: b.x, Y: sz - b.h}, Max: image.Point{X: b.x + 3, Y: sz}}.Push(gtx.Ops)
			paint.Fill(gtx.Ops, col)
			cl.Pop()
		}
	default:
		drawCircle(gtx, half, half, 3, col)
	}
}

// ==========================================================
// DASHBOARD
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
	if !wm.booksLoaded {
		wm.loadBooks()
	}

	var completion float32
	centerLabel := "0%"
	subLabel := "Aucune session"
	bookTitle := "Aucun livre en cours"
	authorLabel := "Importez un livre pour commencer"

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

	if wm.continueReadBtn.Clicked(gtx) {
		if wm.currentBook != nil {
			wm.openBookInReader(wm.currentBook)
		} else {
			// No book yet — trigger import
			wm.importStatusMsg = "Ouverture du sélecteur..."
			go wm.importBooksFromPicker()
		}
	}

	return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return layout.Flex{Axis: layout.Horizontal, Alignment: layout.Middle}.Layout(gtx,
				// Cover
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					w, h := 180, 270
					sr := clip.UniformRRect(image.Rectangle{Min: image.Point{X: 6, Y: 10}, Max: image.Point{X: w + 6, Y: h + 12}}, 10).Push(gtx.Ops)
					paint.Fill(gtx.Ops, theme.WithAlpha(theme.ColorPureBlack, 18))
					sr.Pop()
					cr := clip.UniformRRect(image.Rectangle{Max: image.Point{X: w, Y: h}}, 10).Push(gtx.Ops)
					paint.Fill(gtx.Ops, theme.ColorCyberCyan)
					cr.Pop()
					sheen := clip.UniformRRect(image.Rectangle{Max: image.Point{X: w, Y: 55}}, 10).Push(gtx.Ops)
					paint.Fill(gtx.Ops, theme.WithAlpha(theme.ColorGlassWhite, 14))
					sheen.Pop()
					ts := op.Offset(image.Pt(12, 14)).Push(gtx.Ops)
					gtx2 := gtx
					gtx2.Constraints.Max.X = w - 24
					tl := material.Label(wm.theme, 13, bookTitle)
					tl.Color = theme.ColorGlassWhite
					tl.Font.Weight = font.Bold
					tl.Layout(gtx2)
					ts.Pop()
					return layout.Dimensions{Size: image.Point{X: w + 20, Y: h + 16}}
				}),
				layout.Rigid(layout.Spacer{Width: 48}.Layout),
				// Speedometer
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					return wm.drawSpeedometer(gtx, completion, centerLabel, subLabel)
				}),
				layout.Rigid(layout.Spacer{Width: 48}.Layout),
				// Info + CTA
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
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			if !wm.metricsLoaded {
				wm.loadMetrics()
			}
			totalPages, booksFinished, sessionCount := 0, 0, 0
			for _, s := range wm.recentSessions {
				totalPages += s.CurrentPage
				sessionCount++
				if s.IsBookComplete() {
					booksFinished++
				}
			}
			avgStr := "—"
			if sessionCount > 0 {
				avgStr = fmt.Sprintf("%d min estimees", (totalPages*2)/sessionCount)
			}
			pagesStr, finStr := "—", "—"
			if totalPages > 0 {
				pagesStr = fmt.Sprintf("%d pages", totalPages)
			}
			if booksFinished > 0 {
				finStr = fmt.Sprintf("%d livre(s)", booksFinished)
			}
			return layout.Flex{Axis: layout.Horizontal}.Layout(gtx,
				layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
					return wm.drawAnalyticCard(gtx, "Temps moyen estime", avgStr, "")
				}),
				layout.Rigid(layout.Spacer{Width: 24}.Layout),
				layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
					return wm.drawAnalyticCard(gtx, "Livres termines", finStr, "")
				}),
				layout.Rigid(layout.Spacer{Width: 24}.Layout),
				layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
					return wm.drawAnalyticCard(gtx, "Pages lues au total", pagesStr, "")
				}),
			)
		}),
	)
}

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
	drawArcThick(gtx, cx, cy, r, thickness, startDeg, endTrack, theme.WithAlpha(theme.ColorCyberCyan, 16))
	if completion > 0.005 {
		drawArcThick(gtx, cx, cy, r, thickness+34, startDeg, endProg, theme.WithAlpha(theme.ColorSandGold, 4))
		drawArcThick(gtx, cx, cy, r, thickness+20, startDeg, endProg, theme.WithAlpha(theme.ColorSandGold, 10))
		drawArcThick(gtx, cx, cy, r, thickness+8, startDeg, endProg, theme.WithAlpha(theme.ColorSandGold, 24))
		drawArcThick(gtx, cx, cy, r, thickness, startDeg, endProg, theme.ColorSandGold)
		endRad := float64(endProg) * math.Pi / 180
		dotX := cx + r*float32(math.Cos(endRad))
		dotY := cy + r*float32(math.Sin(endRad))
		drawCircle(gtx, dotX, dotY, 18, theme.WithAlpha(theme.ColorSandGold, 22))
		drawCircle(gtx, dotX, dotY, 11, theme.WithAlpha(theme.ColorSandGold, 50))
		drawCircle(gtx, dotX, dotY, 7, theme.ColorSandGold)
	}
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
			return wm.drawShareCard(gtx, "Bibliotheque complete — Markdown",
				"Un fichier .md lisible avec tous vos livres et fiches.",
				&wm.shareLibBtn, func() {
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
				})
		}),
		layout.Rigid(layout.Spacer{Height: 16}.Layout),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return wm.drawShareCard(gtx, "Bibliotheque complete — JSON",
				"Format structure ideal pour un backup ou une integration.",
				&wm.shareExportBtn, func() {
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
				})
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
// METRICS
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
	// For "heure de prédilection": weight by pages read in each session,
	// not just session count — reflects actual productivity, not just presence.
	hoursPages := make(map[int]int)
	for _, s := range sessions {
		daysCount[s.LastReadingTime.Weekday()]++
		hoursPages[s.LastReadingTime.Hour()] += s.CurrentPage
	}
	var maxDay time.Weekday
	maxDC := 0
	for d, c := range daysCount {
		if c > maxDC {
			maxDC = c
			maxDay = d
		}
	}
	var maxHour, maxHP int
	for h, p := range hoursPages {
		if p > maxHP {
			maxHP = p
			maxHour = h
		}
	}
	frDays := map[time.Weekday]string{
		time.Sunday: "Dimanche", time.Monday: "Lundi", time.Tuesday: "Mardi",
		time.Wednesday: "Mercredi", time.Thursday: "Jeudi",
		time.Friday: "Vendredi", time.Saturday: "Samedi",
	}
	bestDay = frDays[maxDay]
	// Show a clean 2-hour window label
	endH := maxHour + 2
	if endH > 23 {
		endH = 23
	}
	bestHour = fmt.Sprintf("%02dh — %02dh", maxHour, endH)
	sort.Slice(sessions, func(i, j int) bool {
		return sessions[i].LastReadingTime.After(sessions[j].LastReadingTime)
	})
	return
}

func (wm *WindowManager) drawMetrics(gtx layout.Context) layout.Dimensions {
	if !wm.metricsLoaded {
		wm.loadMetrics()
	}
	if !wm.booksLoaded {
		wm.loadBooks()
	}
	bestDay, bestHour, history := wm.computeMetrics()
	totalPages, completedBooks, sessionCount := 0, 0, 0
	uniqueBooks := map[string]bool{}
	for _, s := range history {
		totalPages += s.CurrentPage
		sessionCount++
		uniqueBooks[s.BookID] = true
		if s.IsBookComplete() {
			completedBooks++
		}
	}

	return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			lbl := material.H5(wm.theme, "Statistiques de lecture")
			lbl.Font.Weight = font.Bold
			return layout.Inset{Bottom: 28}.Layout(gtx, lbl.Layout)
		}),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return layout.Flex{Axis: layout.Horizontal}.Layout(gtx,
				layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
					return wm.drawAnalyticCard(gtx, "Sessions totales", fmt.Sprintf("%d", sessionCount), "Toutes les ouvertures de livres")
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
			limit := len(history)
			if limit > 10 {
				limit = 10
			}
			var rows []layout.FlexChild
			for _, s := range history[:limit] {
				session := s
				rows = append(rows, layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					cl := clip.UniformRRect(image.Rectangle{Max: image.Point{X: gtx.Constraints.Max.X, Y: 50}}, 6).Push(gtx.Ops)
					paint.Fill(gtx.Ops, theme.WithAlpha(theme.ColorCyberCyan, 7))
					cl.Pop()
					return layout.Inset{Top: 10, Bottom: 10, Left: 16, Right: 16}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
						return layout.Flex{Axis: layout.Horizontal, Alignment: layout.Middle}.Layout(gtx,
							layout.Flexed(0.4, func(gtx layout.Context) layout.Dimensions {
								title := bookTitleFor(session.BookID, wm.books)
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
								lbl := material.Label(wm.theme, 12, fmt.Sprintf("%.0f%%", session.CalculateCompletion()))
								lbl.Color = theme.ColorSandGold
								lbl.Font.Weight = font.Bold
								return lbl.Layout(gtx)
							}),
						)
					})
				}))
				rows = append(rows, layout.Rigid(layout.Spacer{Height: 6}.Layout))
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

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// bookTitleFor looks up a book title by ID from the cached books slice.
func bookTitleFor(bookID string, books []*domain.Book) string {
	for _, b := range books {
		if b.ID == bookID {
			return b.Title
		}
	}
	n := minInt(8, len(bookID))
	return bookID[:n] + "..."
}

// matchesSearch returns true if the book matches the current search query.
func (wm *WindowManager) matchesSearch(b *domain.Book) bool {
	q := strings.ToLower(strings.TrimSpace(wm.searchQuery))
	if q == "" {
		return true
	}
	return strings.Contains(strings.ToLower(b.Title), q) ||
		strings.Contains(strings.ToLower(b.Author), q)
}

// saveReaderProgress persists the current page via TrackerService in a goroutine.
// UI state mutations are dispatched through uiChan to the main thread — no data races.
func (wm *WindowManager) saveReaderProgress() {
	if wm.trackSvc == nil || wm.readerSession == nil {
		return
	}
	ses := wm.readerSession
	book := wm.readerBook
	page := wm.readerPage + 1
	openedAt := wm.readerOpenedAt // capture before goroutine
	go func() {
		_ = wm.trackSvc.UpdateProgress(context.Background(), page, ses)
		// All UI mutations sent to main thread via buffered channel
		wm.uiChan <- func() {
			wm.dashboardLoaded = false
			wm.metricsLoaded = false
			wm.bookStatusLoaded = false
			if ses.IsBookComplete() && book != nil && wm.achievementBook == nil {
				if wm.dismissedAchievements == nil {
					wm.dismissedAchievements = make(map[string]bool)
				}
				if !wm.dismissedAchievements[book.ID] {
					wm.achievementBook = book
					wm.achievementReadMin = int(time.Since(openedAt).Minutes())
					wm.confettiStart = time.Now()
					wm.confettiActive = true
				}
			}
		}
		wm.window.Invalidate()
	}()
}

// loadBookStatus fetches completion status for all books (background-safe).
func (wm *WindowManager) loadBookStatus() {
	if wm.trackSvc == nil {
		wm.bookStatusLoaded = true
		return
	}
	status, err := wm.trackSvc.BookCompletionStatus(context.Background())
	if err == nil {
		wm.bookStatus = status
	}
	wm.bookStatusLoaded = true
}

// filterBooksByTab applies the tab filter and search query.
func (wm *WindowManager) filterBooksByTab(books []*domain.Book) []*domain.Book {
	if !wm.bookStatusLoaded {
		wm.loadBookStatus()
	}
	var out []*domain.Book
	for _, b := range books {
		if !wm.matchesSearch(b) {
			continue
		}
		switch wm.activeTab {
		case 2: // A lire — unread only
			if wm.bookStatus[b.ID] == "unread" {
				out = append(out, b)
			}
		case 3: // Termines — done only
			if wm.bookStatus[b.ID] == "done" {
				out = append(out, b)
			}
		default: // Tous — all
			out = append(out, b)
		}
	}
	return out
}

// importBooksFromPicker opens a native file dialog for multi-file import.
func (wm *WindowManager) importBooksFromPicker() {
	if wm.libSvc == nil {
		wm.importStatusMsg = "LibraryService non disponible."
		return
	}
	paths := pickMultipleFiles()
	if len(paths) == 0 {
		wm.importStatusMsg = "Aucun fichier selectionne."
		wm.window.Invalidate()
		return
	}
	books, errs := wm.libSvc.ImportBooks(context.Background(), paths)
	if len(books) > 0 {
		wm.booksLoaded = false
		wm.dashboardLoaded = false
		wm.bookStatusLoaded = false
		wm.importStatusMsg = fmt.Sprintf("%d livre(s) importe(s).", len(books))
	}
	if len(errs) > 0 {
		wm.importStatusMsg += fmt.Sprintf(" %d erreur(s).", len(errs))
	}
	wm.window.Invalidate()
}

// deleteBook removes a book from the library.
func (wm *WindowManager) deleteBook(bookID string) {
	if wm.libSvc == nil {
		return
	}
	go func() {
		_ = wm.libSvc.DeleteBook(context.Background(), bookID)
		wm.booksLoaded = false
		wm.dashboardLoaded = false
		wm.bookStatusLoaded = false
		wm.window.Invalidate()
	}()
}

// Keyboard navigation handled directly in Run() via case key.Event.

// =============================================================================
// ACHIEVEMENT MODAL — livre terminé
// =============================================================================

func (wm *WindowManager) drawAchievementModal(gtx layout.Context) {
	book := wm.achievementBook
	if book == nil {
		return
	}
	W := gtx.Constraints.Max.X
	H := gtx.Constraints.Max.Y

	// ── Backdrop dismiss (click outside card) ────────────────────────────────
	// IMPORTANT: check Clicked() BEFORE Layout() — Gio processes events from
	// the previous frame. Layout() registers the hit area for the NEXT frame.
	backdropClicked := wm.achievementDismiss.Clicked(gtx)
	wm.achievementDismiss.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		cl := clip.Rect{Max: gtx.Constraints.Max}.Push(gtx.Ops)
		paint.Fill(gtx.Ops, color.NRGBA{R: 8, G: 6, B: 22, A: 220})
		cl.Pop()
		return layout.Dimensions{Size: gtx.Constraints.Max}
	})
	if backdropClicked {
		if wm.dismissedAchievements == nil {
			wm.dismissedAchievements = make(map[string]bool)
		}
		wm.dismissedAchievements[book.ID] = true
		wm.achievementBook = nil
		wm.confettiActive = false
		return
	}

	// ── Full-screen confetti burst ────────────────────────────────────────────
	if wm.confettiActive {
		elapsed := time.Since(wm.confettiStart).Seconds()
		if elapsed < 4.0 {
			wm.drawConfetti(gtx, elapsed, W, H)
			wm.window.Invalidate() // keep animating
		} else {
			wm.confettiActive = false
		}
	}

	// Book accent color
	bookCol := coverPalette[0]
	for i, b := range wm.books {
		if b.ID == book.ID {
			bookCol = coverPalette[i%len(coverPalette)]
			break
		}
	}

	// ── Card geometry ─────────────────────────────────────────────────────────
	cardW := W * 6 / 10
	if cardW < 520 {
		cardW = 520
	}
	if cardW > 720 {
		cardW = 720
	}
	const cardH = 480
	offX := (W - cardW) / 2
	offY := (H - cardH) / 2

	// Drop shadow
	shOff := op.Offset(image.Pt(offX+4, offY+10)).Push(gtx.Ops)
	scl := clip.UniformRRect(image.Rectangle{Max: image.Pt(cardW, cardH)}, 24).Push(gtx.Ops)
	paint.Fill(gtx.Ops, color.NRGBA{A: 90})
	scl.Pop()
	shOff.Pop()

	// ── Card (one push, one pop at end) ──────────────────────────────────────
	// Outer glow halo around the card
	for gi := 4; gi >= 1; gi-- {
		e := gi * 8
		halocl := clip.UniformRRect(image.Rectangle{
			Min: image.Pt(offX-e, offY-e),
			Max: image.Pt(offX+cardW+e, offY+cardH+e),
		}, 24+e).Push(gtx.Ops)
		paint.Fill(gtx.Ops, theme.WithAlpha(bookCol, uint8(10/gi)))
		halocl.Pop()
	}

	cardOff := op.Offset(image.Pt(offX, offY)).Push(gtx.Ops)
	cardClip := clip.UniformRRect(image.Rectangle{Max: image.Pt(cardW, cardH)}, 24).Push(gtx.Ops)
	paint.Fill(gtx.Ops, color.NRGBA{R: 253, G: 251, B: 246, A: 255})

	// Top gradient wash — tall, dramatic, 10 bands
	for i := 0; i < 10; i++ {
		bandY := i * 10
		alpha := uint8(180 - i*16)
		band := clip.Rect{Min: image.Pt(0, bandY), Max: image.Pt(cardW, bandY+12)}.Push(gtx.Ops)
		paint.Fill(gtx.Ops, theme.WithAlpha(bookCol, alpha))
		band.Pop()
	}
	// Top solid accent stripe
	stripe := clip.UniformRRect(image.Rectangle{Max: image.Pt(cardW, 6)}, 0).Push(gtx.Ops)
	paint.Fill(gtx.Ops, bookCol)
	stripe.Pop()

	// Trophy icon (centered, below stripe) — 64×64
	drawTrophyIcon(gtx, cardW/2-32, 14)

	// Content area — real gtx so buttons work
	gtxC := gtx
	gtxC.Constraints = layout.Exact(image.Point{X: cardW, Y: cardH})
	layout.Inset{Top: unit.Dp(120), Left: unit.Dp(60), Right: unit.Dp(60), Bottom: unit.Dp(36)}.Layout(gtxC,
		func(gtx layout.Context) layout.Dimensions {
			return layout.Flex{Axis: layout.Vertical, Alignment: layout.Middle}.Layout(gtx,
				// Badge
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					return layout.Inset{Bottom: unit.Dp(14)}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
						// Pill badge
						const bw, bh = 210, 32
						gtxB := gtx
						gtxB.Constraints = layout.Exact(image.Pt(bw, bh))
						pill := clip.UniformRRect(image.Rectangle{Max: image.Pt(bw, bh)}, 16).Push(gtx.Ops)
						paint.Fill(gtx.Ops, theme.WithAlpha(bookCol, 50))
						pill.Pop()
						layout.Center.Layout(gtxB, func(gtx layout.Context) layout.Dimensions {
							lbl := material.Label(wm.theme, 12, "LIVRE TERMINÉ  ✓")
							lbl.Font.Weight = font.Bold
							lbl.Color = bookCol
							return lbl.Layout(gtx)
						})
						return layout.Dimensions{Size: image.Pt(bw, bh)}
					})
				}),
				// Title
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					title := book.Title
					if len([]rune(title)) > 44 {
						title = string([]rune(title)[:44]) + "..."
					}
					return layout.Inset{Bottom: unit.Dp(8)}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
						return layout.Center.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
							lbl := material.Label(wm.theme, 22, title)
							lbl.Font.Weight = font.Bold
							lbl.Color = theme.ColorPureBlack
							return lbl.Layout(gtx)
						})
					})
				}),
				// Reading time
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					m := wm.achievementReadMin
					var timeStr string
					switch {
					case m < 1:
						timeStr = "Session de lecture terminée"
					case m < 60:
						timeStr = fmt.Sprintf("%d minutes de lecture cette session", m)
					default:
						timeStr = fmt.Sprintf("%dh%02d de lecture cette session", m/60, m%60)
					}
					return layout.Inset{Bottom: unit.Dp(32)}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
						return layout.Center.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
							lbl := material.Label(wm.theme, 13, timeStr)
							lbl.Color = theme.WithAlpha(theme.ColorPureBlack, 155)
							return lbl.Layout(gtx)
						})
					})
				}),
				// Share label
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					return layout.Inset{Bottom: unit.Dp(16)}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
						lbl := material.Label(wm.theme, 13, "Partager cet exploit :")
						lbl.Color = theme.WithAlpha(theme.ColorPureBlack, 150)
						lbl.Font.Weight = font.SemiBold
						return lbl.Layout(gtx)
					})
				}),
				// Share buttons — 3 large readable buttons in a row
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					return layout.Inset{Bottom: unit.Dp(24)}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
						return layout.Flex{Axis: layout.Horizontal, Alignment: layout.Middle}.Layout(gtx,
							layout.Rigid(func(gtx layout.Context) layout.Dimensions {
								if wm.achievementTwitter.Clicked(gtx) {
									t := wm.buildAchievementShareText(book)
									openURL("https://twitter.com/intent/tweet?text=" + urlEncode(t))
								}
								return layout.Inset{Right: unit.Dp(12)}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
									return wm.achievementTwitter.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
										return drawShareBadge(gtx, "X / Twitter", color.NRGBA{R: 20, G: 20, B: 20, A: 255}, wm.theme,
											wm.achievementTwitter.Hovered(), wm.achievementTwitter.Pressed())
									})
								})
							}),
							layout.Rigid(func(gtx layout.Context) layout.Dimensions {
								if wm.achievementLinkedIn.Clicked(gtx) {
									t := wm.buildAchievementShareText(book)
									openURL("https://www.linkedin.com/shareArticle?mini=true&summary=" + urlEncode(t))
								}
								return layout.Inset{Right: unit.Dp(12)}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
									return wm.achievementLinkedIn.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
										return drawShareBadge(gtx, "LinkedIn", color.NRGBA{R: 10, G: 102, B: 194, A: 255}, wm.theme,
											wm.achievementLinkedIn.Hovered(), wm.achievementLinkedIn.Pressed())
									})
								})
							}),
							layout.Rigid(func(gtx layout.Context) layout.Dimensions {
								if wm.achievementShare.Clicked(gtx) {
									t := wm.buildAchievementShareText(book)
									_ = copyToClipboard(t)
								}
								return wm.achievementShare.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
									return drawShareBadge(gtx, "Copier le texte", color.NRGBA{R: 68, G: 68, B: 80, A: 255}, wm.theme,
										wm.achievementShare.Hovered(), wm.achievementShare.Pressed())
								})
							}),
						)
					})
				}),
				// Dismiss hint — prominent
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					return layout.Center.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
						lbl := material.Label(wm.theme, 12, "Cliquer en dehors  ·  Échap pour fermer")
						lbl.Color = theme.WithAlpha(theme.ColorPureBlack, 80)
						return lbl.Layout(gtx)
					})
				}),
			)
		})
	cardClip.Pop()
	cardOff.Pop()
}

// buildAchievementShareText builds an engagement-driven share message.
// Rotates between several storytelling angles to feel human, not generated.
func (wm *WindowManager) buildAchievementShareText(book *domain.Book) string {
	m := wm.achievementReadMin
	title := book.Title

	capitalise := func(s string) string {
		if s == "" {
			return s
		}
		r := []rune(s)
		if r[0] >= 'a' && r[0] <= 'z' {
			r[0] -= 32
		}
		return string(r)
	}

	// Time phrase — honest and specific
	var timePhrase string
	switch {
	case m < 1:
		timePhrase = ""
	case m < 30:
		timePhrase = fmt.Sprintf("en %d minutes chrono", m)
	case m < 60:
		timePhrase = fmt.Sprintf("en %d minutes de lecture pure", m)
	case m == 60:
		timePhrase = "en exactement 1 heure"
	default:
		timePhrase = fmt.Sprintf("en %dh%02d", m/60, m%60)
	}

	// Pick angle based on title hash — deterministic but varied
	angle := 0
	for _, c := range title {
		angle += int(c)
	}
	angle = angle % 5

	var sb strings.Builder
	switch angle {
	case 0:
		sb.WriteString(fmt.Sprintf("Vient de tomber le rideau sur \"%s\".", title))
		if timePhrase != "" {
			sb.WriteString(fmt.Sprintf(" %s.", capitalise(timePhrase)))
		}
		sb.WriteString("\n\nCe livre m'a forcé à remettre en question des choses que je croyais acquises.")
		sb.WriteString("\n\nVous l'avez lu ? Dites-moi ce que vous en avez pensé. 👇")
	case 1:
		sb.WriteString(fmt.Sprintf("Terminé : \"%s\"", title))
		if timePhrase != "" {
			sb.WriteString(fmt.Sprintf(" (%s)", timePhrase))
		}
		sb.WriteString(".\n\nIl y a des livres qu'on ferme et qu'on oublie.")
		sb.WriteString("\nIl y en a d'autres qui continuent de tourner dans la tête longtemps après.")
		sb.WriteString("\n\nCelui-ci fait partie de la deuxième catégorie.")
		sb.WriteString("\n\nQuel livre vous a le plus marqué ces 6 derniers mois ?")
	case 2:
		sb.WriteString(fmt.Sprintf("Pages finales de \"%s\" — done.", title))
		if timePhrase != "" {
			sb.WriteString(fmt.Sprintf("\n%s.", capitalise(timePhrase)))
		}
		sb.WriteString("\n\nOn sous-estime à quel point lire régulièrement transforme la façon dont on pense,")
		sb.WriteString("\ndont on parle, dont on prend des décisions.")
		sb.WriteString("\n\nC'est le seul entraînement mental qui ne se voit pas mais qui change tout.")
		sb.WriteString("\n\nVous lisez quoi en ce moment ?")
	case 3:
		if timePhrase != "" {
			sb.WriteString(fmt.Sprintf("%s pour finir \"%s\".", capitalise(timePhrase), title))
		} else {
			sb.WriteString(fmt.Sprintf("Fin de \"%s\".", title))
		}
		sb.WriteString("\n\nChaque livre terminé est une version de vous-même en plus.")
		sb.WriteString("\nNon pas parce que vous avez \"lu\", mais parce que vous avez réfléchi.")
		sb.WriteString("\n\nMerci à ceux qui m'ont recommandé ce titre. Vous savez qui vous êtes. 🙏")
		sb.WriteString("\n\nProchaine lecture ? Je suis preneur de suggestions sérieuses.")
	default:
		sb.WriteString(fmt.Sprintf("Un livre de plus dans la colonne des terminés : \"%s\".", title))
		if timePhrase != "" {
			sb.WriteString(fmt.Sprintf(" %s.", capitalise(timePhrase)))
		}
		sb.WriteString("\n\nOn parle beaucoup de productivité, de croissance, d'optimisation.")
		sb.WriteString("\nMais rien ne compresse autant d'expérience en aussi peu de temps qu'un bon livre.")
		sb.WriteString("\n\nC'est le meilleur ROI qui existe. Point.")
		sb.WriteString("\n\nQu'est-ce que vous lisez pour progresser en ce moment ?")
	}

	sb.WriteString("\n\n#lecture #livres #Orus")
	return sb.String()
}

// drawTrophyIcon draws a gold trophy cup vector at ox,oy (64×64px).
func drawTrophyIcon(gtx layout.Context, ox, oy int) {
	gold := theme.ColorSandGold
	shine := color.NRGBA{R: 255, G: 230, B: 120, A: 255}

	// Outer glow rings
	for gi := 3; gi >= 1; gi-- {
		e := gi * 6
		gcl := clip.UniformRRect(image.Rectangle{
			Min: image.Pt(ox-e, oy-e),
			Max: image.Pt(ox+64+e, oy+64+e),
		}, 32+e).Push(gtx.Ops)
		paint.Fill(gtx.Ops, color.NRGBA{R: 245, G: 166, B: 35, A: uint8(14 / gi)})
		gcl.Pop()
	}

	// Cup body
	var cup clip.Path
	cup.Begin(gtx.Ops)
	cup.MoveTo(f32.Pt(float32(ox+12), float32(oy+5)))
	cup.LineTo(f32.Pt(float32(ox+52), float32(oy+5)))
	cup.LineTo(f32.Pt(float32(ox+47), float32(oy+32)))
	cup.LineTo(f32.Pt(float32(ox+38), float32(oy+42)))
	cup.LineTo(f32.Pt(float32(ox+26), float32(oy+42)))
	cup.LineTo(f32.Pt(float32(ox+17), float32(oy+32)))
	cup.Close()
	paint.FillShape(gtx.Ops, gold, clip.Outline{Path: cup.End()}.Op())

	// Inner shine highlight on cup
	var sh clip.Path
	sh.Begin(gtx.Ops)
	sh.MoveTo(f32.Pt(float32(ox+16), float32(oy+8)))
	sh.LineTo(f32.Pt(float32(ox+30), float32(oy+8)))
	sh.LineTo(f32.Pt(float32(ox+26), float32(oy+22)))
	sh.LineTo(f32.Pt(float32(ox+14), float32(oy+22)))
	sh.Close()
	paint.FillShape(gtx.Ops, theme.WithAlpha(shine, 80), clip.Outline{Path: sh.End()}.Op())

	// Stem
	stem := clip.UniformRRect(image.Rectangle{
		Min: image.Pt(ox+27, oy+42), Max: image.Pt(ox+37, oy+54),
	}, 2).Push(gtx.Ops)
	paint.Fill(gtx.Ops, gold)
	stem.Pop()

	// Base
	base := clip.UniformRRect(image.Rectangle{
		Min: image.Pt(ox+18, oy+54), Max: image.Pt(ox+46, oy+62),
	}, 4).Push(gtx.Ops)
	paint.Fill(gtx.Ops, gold)
	base.Pop()

	// Left handle
	var lh clip.Path
	lh.Begin(gtx.Ops)
	lh.MoveTo(f32.Pt(float32(ox+12), float32(oy+5)))
	lh.LineTo(f32.Pt(float32(ox+4), float32(oy+5)))
	lh.LineTo(f32.Pt(float32(ox+4), float32(oy+24)))
	lh.LineTo(f32.Pt(float32(ox+17), float32(oy+24)))
	lh.Close()
	paint.FillShape(gtx.Ops, gold, clip.Outline{Path: lh.End()}.Op())

	// Right handle
	var rh clip.Path
	rh.Begin(gtx.Ops)
	rh.MoveTo(f32.Pt(float32(ox+52), float32(oy+5)))
	rh.LineTo(f32.Pt(float32(ox+60), float32(oy+5)))
	rh.LineTo(f32.Pt(float32(ox+60), float32(oy+24)))
	rh.LineTo(f32.Pt(float32(ox+47), float32(oy+24)))
	rh.Close()
	paint.FillShape(gtx.Ops, gold, clip.Outline{Path: rh.End()}.Op())
}

// drawConfetti renders a full-screen celebration burst.
// elapsed goes from 0→4 seconds. Particles launch upward from bottom, fade out.
func (wm *WindowManager) drawConfetti(gtx layout.Context, elapsed float64, W, H int) {
	// 7 vivid celebration colors
	confettiColors := []color.NRGBA{
		{R: 255, G: 215, B: 0, A: 255},   // gold
		{R: 255, G: 82, B: 82, A: 255},   // coral
		{R: 72, G: 199, B: 142, A: 255},  // mint
		{R: 99, G: 179, B: 237, A: 255},  // sky
		{R: 214, G: 158, B: 46, A: 255},  // amber
		{R: 198, G: 125, B: 255, A: 255}, // lavender
		{R: 255, G: 255, B: 255, A: 255}, // white
	}

	const numParticles = 120

	for i := 0; i < numParticles; i++ {
		fi := float64(i)

		// Each particle has a unique launch angle and speed derived from its index
		seed := fi*137.508 + 42.0 // golden ratio spread, never repeats
		launchDelay := math.Mod(fi*0.031, 0.9)
		age := elapsed - launchDelay
		if age <= 0 {
			continue
		}

		// Normalised age 0→1 over particle lifetime (1.8s each)
		lifetime := 2.2
		norm := age / lifetime
		if norm > 1 {
			norm = 1
		}

		// Horizontal position — spread across full width
		xFrac := math.Mod(seed*0.00731, 1.0)
		startX := float32(xFrac * float64(W))

		// Horizontal drift
		driftSpeed := (math.Mod(seed*0.0411, 1.0) - 0.5) * 160.0
		px := startX + float32(driftSpeed*age)

		// Vertical — launch from bottom, arc upward then gravity pulls down
		// v0 varies per particle
		v0 := 600.0 + math.Mod(seed*0.317, 1.0)*500.0
		gravity := 420.0
		py := float32(H) - float32(v0*age-0.5*gravity*age*age)

		// Skip particles that left the screen
		if px < -20 || px > float32(W)+20 || py > float32(H)+20 {
			continue
		}

		// Alpha fade — full for first 60%, then fade out
		alpha := float64(1.0)
		if norm > 0.6 {
			alpha = 1.0 - (norm-0.6)/0.4
		}
		if alpha <= 0 {
			continue
		}

		col := confettiColors[i%len(confettiColors)]
		col.A = uint8(float64(col.A) * alpha)

		// Shape: alternate between rectangles (ribbon) and diamonds
		size := float32(6 + math.Mod(fi*0.13, 1.0)*6)
		// Rotation angle
		angle := math.Mod(seed*6.28+elapsed*(math.Mod(seed*0.3, 1.0)*4+1), 6.28)
		cosA := float32(math.Cos(angle))
		sinA := float32(math.Sin(angle))

		if i%3 == 0 {
			// Diamond
			hw := size * 0.55
			hh := size * 0.85
			var dp clip.Path
			dp.Begin(gtx.Ops)
			dp.MoveTo(f32.Pt(px+cosA*0-sinA*(-hh), py+sinA*0+cosA*(-hh)))
			dp.LineTo(f32.Pt(px+cosA*hw-sinA*0, py+sinA*hw+cosA*0))
			dp.LineTo(f32.Pt(px+cosA*0-sinA*hh, py+sinA*0+cosA*hh))
			dp.LineTo(f32.Pt(px+cosA*(-hw)-sinA*0, py+sinA*(-hw)+cosA*0))
			dp.Close()
			paint.FillShape(gtx.Ops, col, clip.Outline{Path: dp.End()}.Op())
		} else {
			// Ribbon rectangle
			hw := size * 0.3
			hh := size * 0.9
			corners := [4]f32.Point{
				{X: px + cosA*(-hw) - sinA*(-hh), Y: py + sinA*(-hw) + cosA*(-hh)},
				{X: px + cosA*hw - sinA*(-hh), Y: py + sinA*hw + cosA*(-hh)},
				{X: px + cosA*hw - sinA*hh, Y: py + sinA*hw + cosA*hh},
				{X: px + cosA*(-hw) - sinA*hh, Y: py + sinA*(-hw) + cosA*hh},
			}
			var rp clip.Path
			rp.Begin(gtx.Ops)
			rp.MoveTo(corners[0])
			rp.LineTo(corners[1])
			rp.LineTo(corners[2])
			rp.LineTo(corners[3])
			rp.Close()
			paint.FillShape(gtx.Ops, col, clip.Outline{Path: rp.End()}.Op())
		}
	}
}

// drawShareBadge draws a proper wide button with label — readable and classy.
// hovered and pressed must be passed from the parent widget.Clickable.
func drawShareBadge(gtx layout.Context, label string, col color.NRGBA, th *material.Theme, hovered, pressed bool) layout.Dimensions {
	const bh = 46
	const bw = 148

	// Dynamic background alpha for tactile feedback
	bgAlpha := uint8(255)
	if pressed {
		bgAlpha = 155
	} else if hovered {
		bgAlpha = 210
	}
	bgCol := theme.WithAlpha(col, bgAlpha)

	// Scale down slightly on press for a physical "click" feel
	scaleOffset := 0
	if pressed {
		scaleOffset = 1
	}

	// Shadow (softened on press)
	shadowAlpha := uint8(35)
	if pressed {
		shadowAlpha = 12
	}
	shcl := clip.UniformRRect(image.Rectangle{
		Min: image.Pt(1+scaleOffset, 2+scaleOffset),
		Max: image.Pt(bw+1-scaleOffset, bh+2-scaleOffset),
	}, bh/2).Push(gtx.Ops)
	paint.Fill(gtx.Ops, color.NRGBA{A: shadowAlpha})
	shcl.Pop()

	// Button body
	bg := clip.UniformRRect(image.Rectangle{
		Min: image.Pt(scaleOffset, scaleOffset),
		Max: image.Pt(bw-scaleOffset, bh-scaleOffset),
	}, bh/2).Push(gtx.Ops)
	paint.Fill(gtx.Ops, bgCol)
	bg.Pop()

	// Top shine (dimmed on hover, absent on press)
	if !pressed {
		shineAlpha := uint8(18)
		if hovered {
			shineAlpha = 38
		}
		shine := clip.UniformRRect(image.Rectangle{
			Min: image.Pt(scaleOffset, scaleOffset),
			Max: image.Pt(bw-scaleOffset, bh/2),
		}, bh/2).Push(gtx.Ops)
		paint.Fill(gtx.Ops, color.NRGBA{R: 255, G: 255, B: 255, A: shineAlpha})
		shine.Pop()
	}

	// Label centered
	gtx2 := gtx
	gtx2.Constraints = layout.Exact(image.Pt(bw, bh))
	layout.Center.Layout(gtx2, func(gtx layout.Context) layout.Dimensions {
		lbl := material.Label(th, 13, label)
		lbl.Color = color.NRGBA{R: 255, G: 255, B: 255, A: 255}
		lbl.Font.Weight = font.SemiBold
		return lbl.Layout(gtx)
	})
	return layout.Dimensions{Size: image.Pt(bw, bh)}
}
