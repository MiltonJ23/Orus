package views

import (
	"context"
	"fmt"
	"image"
	"log"
	"strconv"
	"strings"

	"gioui.org/font"
	"gioui.org/layout"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/widget"
	"gioui.org/widget/material"

	"github.com/MiltonJ23/Orus/internal/adapters/ui/theme"
	"github.com/MiltonJ23/Orus/internal/domain"
)

type reminderFormState struct {
	labelEditor  widget.Editor
	hourEditor   widget.Editor
	minuteEditor widget.Editor
	freqBtns     [4]widget.Clickable
	selectedFreq int // 0=daily,1=weekly,2=weekdays,3=once
	saveBtn      widget.Clickable
	showForm     bool
	newBtn       widget.Clickable
	statusMsg    string
}

var freqOptions = []struct {
	Label string
	Value domain.ReminderFrequency
}{
	{"Tous les jours", domain.FrequencyDaily},
	{"Chaque semaine", domain.FrequencyWeekly},
	{"Lun–Ven", domain.FrequencyWeekdays},
	{"Une seule fois", domain.FrequencyOnce},
}

func (wm *WindowManager) drawRemindersView(gtx layout.Context) layout.Dimensions {
	if !wm.remindersLoaded && wm.reminderSvc != nil {
		reminders, err := wm.reminderSvc.ListReminders(context.Background())
		if err != nil {
			log.Printf("[Reminders] Erreur chargement : %v", err)
		} else {
			wm.reminders = reminders
		}
		wm.remindersLoaded = true
	}

	return layout.Flex{Axis: layout.Vertical}.Layout(gtx,

		// En-tête
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return layout.Flex{Axis: layout.Horizontal, Alignment: layout.Middle}.Layout(gtx,
				layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
					lbl := material.H5(wm.theme, "⏰ Rappels de lecture")
					lbl.Font.Weight = font.Bold
					lbl.Color = theme.ColorPureBlack
					return lbl.Layout(gtx)
				}),
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					if wm.reminderForm.newBtn.Clicked(gtx) {
						wm.reminderForm.showForm = !wm.reminderForm.showForm
					}
					label := "+ Nouveau rappel"
					if wm.reminderForm.showForm {
						label = "✕ Annuler"
					}
					return wm.drawPillButton(gtx, label, &wm.reminderForm.newBtn, theme.ColorCyberCyan)
				}),
			)
		}),

		layout.Rigid(layout.Spacer{Height: 24}.Layout),

		// Formulaire
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			if !wm.reminderForm.showForm {
				return layout.Dimensions{}
			}
			return wm.drawReminderForm(gtx)
		}),

		// Liste
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			if len(wm.reminders) == 0 {
				lbl := material.Label(wm.theme, 15, "Aucun rappel configuré. Créez-en un pour maintenir votre habitude de lecture !")
				lbl.Color = theme.WithAlpha(theme.ColorPureBlack, 130)
				return layout.Center.Layout(gtx, lbl.Layout)
			}
			return wm.drawReminderList(gtx)
		}),
	)
}

func (wm *WindowManager) drawReminderForm(gtx layout.Context) layout.Dimensions {
	cl := clip.UniformRRect(image.Rectangle{Max: image.Point{X: gtx.Constraints.Max.X, Y: 380}}, 10).Push(gtx.Ops)
	paint.Fill(gtx.Ops, theme.WithAlpha(theme.ColorCyberCyan, 10))
	cl.Pop()

	return layout.Inset{Top: 16, Bottom: 24, Left: 24, Right: 24}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		return layout.Flex{Axis: layout.Vertical}.Layout(gtx,

			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				lbl := material.Label(wm.theme, 14, "Nouveau rappel de lecture")
				lbl.Font.Weight = font.Bold
				return layout.Inset{Bottom: 16}.Layout(gtx, lbl.Layout)
			}),

			// Message
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				return wm.drawLabeledField(gtx, "Message du rappel", &wm.reminderForm.labelEditor, "📖 C'est l'heure de lire !")
			}),
			layout.Rigid(layout.Spacer{Height: 12}.Layout),

			// Heure HH:MM
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				return layout.Flex{Axis: layout.Horizontal, Alignment: layout.Middle}.Layout(gtx,
					layout.Rigid(func(gtx layout.Context) layout.Dimensions {
						gtx.Constraints.Max.X = 80
						return wm.drawLabeledField(gtx, "Heure", &wm.reminderForm.hourEditor, "21")
					}),
					layout.Rigid(func(gtx layout.Context) layout.Dimensions {
						lbl := material.Label(wm.theme, 20, " h ")
						lbl.Color = theme.ColorCyberCyan
						return lbl.Layout(gtx)
					}),
					layout.Rigid(func(gtx layout.Context) layout.Dimensions {
						gtx.Constraints.Max.X = 80
						return wm.drawLabeledField(gtx, "Minute", &wm.reminderForm.minuteEditor, "00")
					}),
				)
			}),
			layout.Rigid(layout.Spacer{Height: 12}.Layout),

			// Fréquence
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				lbl := material.Label(wm.theme, 13, "Fréquence :")
				lbl.Color = theme.WithAlpha(theme.ColorPureBlack, 170)
				return layout.Inset{Bottom: 8}.Layout(gtx, lbl.Layout)
			}),
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				var chips []layout.FlexChild
				for i, opt := range freqOptions {
					idx := i
					chips = append(chips, layout.Rigid(func(gtx layout.Context) layout.Dimensions {
						if wm.reminderForm.freqBtns[idx].Clicked(gtx) {
							wm.reminderForm.selectedFreq = idx
						}
						active := wm.reminderForm.selectedFreq == idx
						bgAlpha := uint8(15)
						if active {
							bgAlpha = 60
						}
						cl := clip.UniformRRect(image.Rectangle{
							Max: image.Point{X: gtx.Constraints.Max.X, Y: 32},
						}, 16).Push(gtx.Ops)
						paint.Fill(gtx.Ops, theme.WithAlpha(theme.ColorCyberCyan, bgAlpha))
						cl.Pop()
						return layout.Inset{Left: 8}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
							return wm.reminderForm.freqBtns[idx].Layout(gtx, func(gtx layout.Context) layout.Dimensions {
								lbl := material.Label(wm.theme, 13, freqOptions[idx].Label)
								if active {
									lbl.Font.Weight = font.Bold
									lbl.Color = theme.ColorCyberCyan
								}
								_ = opt
								return layout.Inset{Top: 7, Bottom: 7, Left: 12, Right: 12}.Layout(gtx, lbl.Layout)
							})
						})
					}))
					chips = append(chips, layout.Rigid(layout.Spacer{Width: 8}.Layout))
				}
				return layout.Flex{Axis: layout.Horizontal}.Layout(gtx, chips...)
			}),
			layout.Rigid(layout.Spacer{Height: 20}.Layout),

			// Statut
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				if wm.reminderForm.statusMsg == "" {
					return layout.Dimensions{}
				}
				lbl := material.Label(wm.theme, 13, wm.reminderForm.statusMsg)
				lbl.Color = theme.ColorSandGold
				return layout.Inset{Bottom: 8}.Layout(gtx, lbl.Layout)
			}),

			// Sauvegarder
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				if wm.reminderForm.saveBtn.Clicked(gtx) {
					wm.submitReminderForm()
				}
				return wm.drawPillButton(gtx, "✓ Créer le rappel", &wm.reminderForm.saveBtn, theme.ColorSandGold)
			}),
		)
	})
}

func (wm *WindowManager) submitReminderForm() {
	if wm.reminderSvc == nil {
		wm.reminderForm.statusMsg = "Service non disponible."
		return
	}

	hour, err := strconv.Atoi(strings.TrimSpace(wm.reminderForm.hourEditor.Text()))
	if err != nil || hour < 0 || hour > 23 {
		wm.reminderForm.statusMsg = "Heure invalide (0–23)."
		return
	}
	minute, err := strconv.Atoi(strings.TrimSpace(wm.reminderForm.minuteEditor.Text()))
	if err != nil || minute < 0 || minute > 59 {
		wm.reminderForm.statusMsg = "Minute invalide (0–59)."
		return
	}

	label := wm.reminderForm.labelEditor.Text()
	freq := freqOptions[wm.reminderForm.selectedFreq].Value

	_, err = wm.reminderSvc.AddReminder(context.Background(), "", "", label, hour, minute, freq)
	if err != nil {
		wm.reminderForm.statusMsg = "Erreur : " + err.Error()
		return
	}

	wm.reminderForm.labelEditor.SetText("")
	wm.reminderForm.hourEditor.SetText("")
	wm.reminderForm.minuteEditor.SetText("")
	wm.reminderForm.selectedFreq = 0
	wm.reminderForm.showForm = false
	wm.remindersLoaded = false
	wm.reminderForm.statusMsg = ""
}

func (wm *WindowManager) drawReminderList(gtx layout.Context) layout.Dimensions {
	var rows []layout.FlexChild
	for _, r := range wm.reminders {
		rem := r
		rows = append(rows, layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return wm.drawReminderCard(gtx, rem)
		}))
		rows = append(rows, layout.Rigid(layout.Spacer{Height: 12}.Layout))
	}
	return layout.Flex{Axis: layout.Vertical}.Layout(gtx, rows...)
}

func (wm *WindowManager) drawReminderCard(gtx layout.Context, r *domain.Reminder) layout.Dimensions {
	bgAlpha := uint8(12)
	if !r.Enabled {
		bgAlpha = 5
	}

	cl := clip.UniformRRect(image.Rectangle{Max: image.Point{X: gtx.Constraints.Max.X, Y: 90}}, 8).Push(gtx.Ops)
	paint.Fill(gtx.Ops, theme.WithAlpha(theme.ColorCyberCyan, bgAlpha))
	cl.Pop()

	return layout.Inset{Top: 14, Left: 20, Right: 20, Bottom: 14}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		return layout.Flex{Axis: layout.Horizontal, Alignment: layout.Middle}.Layout(gtx,

			// Heure
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				timeStr := fmt.Sprintf("%02d:%02d", r.Hour, r.Minute)
				lbl := material.Label(wm.theme, 28, timeStr)
				lbl.Font.Weight = font.Bold
				if !r.Enabled {
					lbl.Color = theme.WithAlpha(theme.ColorPureBlack, 80)
				} else {
					lbl.Color = theme.ColorCyberCyan
				}
				return layout.Inset{Right: 20}.Layout(gtx, lbl.Layout)
			}),

			// Label + fréquence
			layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
				return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
					layout.Rigid(func(gtx layout.Context) layout.Dimensions {
						lbl := material.Label(wm.theme, 14, r.Label)
						lbl.Font.Weight = font.Bold
						if !r.Enabled {
							lbl.Color = theme.WithAlpha(theme.ColorPureBlack, 100)
						}
						return lbl.Layout(gtx)
					}),
					layout.Rigid(layout.Spacer{Height: 4}.Layout),
					layout.Rigid(func(gtx layout.Context) layout.Dimensions {
						lbl := material.Label(wm.theme, 12, r.FrequencyLabel())
						lbl.Color = theme.WithAlpha(theme.ColorPureBlack, 140)
						return lbl.Layout(gtx)
					}),
				)
			}),

			// Badge ON/OFF
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				badge := "● Actif"
				col := theme.ColorSandGold
				if !r.Enabled {
					badge = "○ Inactif"
					col = theme.WithAlpha(theme.ColorPureBlack, 100)
				}
				lbl := material.Label(wm.theme, 12, badge)
				lbl.Color = col
				lbl.Font.Weight = font.Bold
				return lbl.Layout(gtx)
			}),
		)
	})
}
