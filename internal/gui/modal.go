package gui

import (
	"github.com/rivo/tview"
	"yesbotics/ysm/internal/config"
)

type Modal struct {
	grid *tview.Grid
}

func NewModalFavNew(
	defaultText string,
	onAdd func(fav config.MessageFavorite),
	onCancel func(),
) *Modal {
	message := config.MessageFavorite{
		Name:    "",
		Message: defaultText,
	}

	form := tview.NewForm()
	checkValid := func() {
		form.GetButton(0).SetDisabled(!message.IsValid())
	}

	form.
		AddInputField("Title", "", 0, nil, func(text string) {
			message.Name = text
			checkValid()
		}).
		AddInputField("Message", defaultText, 0, nil, func(text string) {
			message.Message = text
			checkValid()
		}).
		AddButton("Create", func() {
			onAdd(message)
		}).
		AddButton("Cancel", onCancel)

	form.SetTitle(" New favorite ")
	form.SetBorder(true)
	form.SetCancelFunc(onCancel)
	checkValid()

	return newModal(form, 50, 9)
}

func NewModalFavEdit(
	message *config.MessageFavorite,
	onSave func(),
	onCancel func(),
) *Modal {
	messageTemp := config.MessageFavorite{
		Name:    message.Name,
		Message: message.Message,
	}

	form := tview.NewForm()
	checkValid := func() {
		form.GetButton(0).SetDisabled(!messageTemp.IsValid())
	}

	form.
		AddInputField("Title", messageTemp.Name, 0, nil, func(text string) {
			messageTemp.Name = text
			checkValid()
		}).
		AddInputField("Message", messageTemp.Message, 0, nil, func(text string) {
			messageTemp.Message = text
			checkValid()
		}).
		AddButton("Save", func() {
			message.Name = messageTemp.Name
			message.Message = messageTemp.Message
			onSave()
		}).
		AddButton("Cancel", onCancel)

	form.SetTitle(" Edit favorite ")
	form.SetBorder(true)
	form.SetCancelFunc(onCancel)
	checkValid()

	return newModal(form, 50, 9)
}

func newModal(p tview.Primitive, width int, height int) *Modal {
	grid := tview.NewGrid().
		SetColumns(0, width, 0).
		SetRows(0, height, 0).
		AddItem(p, 1, 1, 1, 1, 0, 0, true)
	return &Modal{
		grid: grid,
	}
}

func (m *Modal) GetPrimitive() tview.Primitive {
	return m.grid
}
