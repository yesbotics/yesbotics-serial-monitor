package gui

import (
	"github.com/rivo/tview"
	"yesbotics/ysm/internal/config"
)

type Favorites struct {
	list         *tview.List
	config       *config.Config
	grid         *tview.Grid
	buttonNew    *tview.Button
	buttonEdit   *tview.Button
	buttonDelete *tview.Button
	buttonSend   *tview.Button
}

// This also determines the maximum number of favorites
const favShortcuts = "123456789abcdefghijklmnopqrstuvwxyz"

func NewFavorites(
	configP *config.Config,
	onNew func(),
	onEdit func(int, config.MessageFavorite),
	onSend func(string),
) *Favorites {

	list := tview.NewList()
	list.SetTitle(" Favorites ")

	buttonSend := tview.NewButton("Send")

	buttonDuplicate := tview.NewButton("Copy")

	buttonNew := tview.NewButton("New")
	buttonNew.SetSelectedFunc(onNew)

	buttonEdit := tview.NewButton("Edit")
	buttonEdit.SetSelectedFunc(func() {
		if list.GetItemCount() < 1 {
			return
		}
		index := list.GetCurrentItem()
		main, secondary := list.GetItemText(index)
		onEdit(index, config.MessageFavorite{
			Name:    main,
			Message: secondary,
		})
	})

	buttonDelete := tview.NewButton("Delete")

	buttonGrid := tview.NewGrid()
	buttonGrid.AddItem(buttonSend, 0, 0, 1, 4, 0, 0, false)
	buttonGrid.AddItem(buttonNew, 1, 0, 1, 1, 0, 0, false)
	buttonGrid.AddItem(buttonDuplicate, 1, 1, 1, 1, 0, 0, false)
	buttonGrid.AddItem(buttonEdit, 1, 2, 1, 1, 0, 0, false)
	buttonGrid.AddItem(buttonDelete, 1, 3, 1, 1, 0, 0, false)

	grid := tview.NewGrid()
	grid.SetTitle(" Favorites ")
	grid.SetBorder(true)
	grid.SetRows(0, 2)
	grid.AddItem(list, 0, 0, 1, 1, 0, 0, false)
	grid.AddItem(buttonGrid, 1, 0, 1, 1, 0, 0, false)

	fav := Favorites{
		config:       configP,
		list:         list,
		grid:         grid,
		buttonNew:    buttonNew,
		buttonEdit:   buttonEdit,
		buttonDelete: buttonDelete,
		buttonSend:   buttonSend,
	}

	fav.RenderList(0)

	buttonDelete.SetSelectedFunc(func() {
		if list.GetItemCount() < 1 {
			return
		}
		message := fav.getSelectedItem()
		configP.RemoveMessageFavorite(message.Name, message.Message)
		fav.RenderList(0)
	})

	buttonSend.SetSelectedFunc(func() {
		if list.GetItemCount() < 1 {
			return
		}
		message := fav.getSelectedItem()
		onSend(message.Message)
	})

	buttonDuplicate.SetSelectedFunc(func() {
		if list.GetItemCount() < 1 {
			return
		}
		message := fav.getSelectedItem()
		index := list.GetCurrentItem()
		fav.config.AddMessageFavorite(&message, true)
		fav.RenderList(index)
	})

	fav.updateButtonAvailability()

	return &fav
}

func (f *Favorites) RenderList(selectIndex int) {
	f.list.Clear()

	runes := []rune(favShortcuts)
	for i, favorite := range f.config.GetMessageFavorites() {
		shortcut := runes[i%len(runes)]
		f.list.AddItem(favorite.Name, favorite.Message, shortcut, nil)
	}

	f.list.SetCurrentItem(selectIndex)
	f.updateButtonAvailability()
}

func (f *Favorites) GetPrimitive() *tview.Grid {
	return f.grid
}

func (f *Favorites) updateButtonAvailability() {
	runes := []rune(favShortcuts)
	newItemsAllowed := f.list.GetItemCount() < len(runes)
	editDeleteSendAllowed := f.list.GetItemCount() > 0 && f.list.GetCurrentItem() >= 0
	f.buttonSend.SetDisabled(!editDeleteSendAllowed)
	f.buttonNew.SetDisabled(!newItemsAllowed)
	f.buttonEdit.SetDisabled(!editDeleteSendAllowed)
	f.buttonDelete.SetDisabled(!editDeleteSendAllowed)
}

func (f *Favorites) getSelectedItem() config.MessageFavorite {
	index := f.list.GetCurrentItem()
	main, secondary := f.list.GetItemText(index)
	return config.MessageFavorite{
		Name:    main,
		Message: secondary,
	}
}
