package gui

import (
	"fmt"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"log"
	"os"
	"slices"
	"strconv"
	"time"
	"yesbotics/ysm/internal/config"
	"yesbotics/ysm/internal/hex"
	"yesbotics/ysm/internal/serialcon"
)

const DEFAULT_VERSION = "0.1.0"
const SHORTCUTS = "Ctrl-q: Quit, Ctrl-r: Reload devices, Ctrl-l: Clear, Ctrl-d: Disconnect, Ctrl-h: Hex"
const PAGENAME_MODAL = "modal"
const MIN_SIZE = 80

type Gui struct {
	appConfig config.AppConfig
	//serialChannel    chan string
	//showHex          bool
	serialConnection *serialcon.Serialcon

	app             *tview.Application
	pages           *tview.Pages
	grid            *tview.Grid
	messagesTable   *tview.Table
	input           *tview.InputField
	deviceList      *tview.DropDown
	deviceListEmpty *tview.TextView
	favorites       *Favorites
	baudrateList    *tview.DropDown
}

type MessageType int

type Element interface {
	GetPrimitive() tview.Primitive
}

const (
	MessageIn MessageType = iota
	MessageOut
	MessageStatus
	MessageError
)

func New(appConfig config.AppConfig) (*Gui, error) {
	gui := &Gui{
		appConfig: appConfig,
	}

	gui.messagesTable = gui.getMessagesTable()

	inputGrid := gui.getInputGrid()
	bottomRow := gui.getBottomRow()
	sidebar := gui.getSidebar()

	gui.grid = tview.NewGrid().
		SetRows(0, 3, 1).
		SetColumns(0, 30).
		SetBorders(false)

	gui.grid.
		AddItem(gui.messagesTable, 0, 0, 1, 2, 0, 0, false).
		AddItem(gui.messagesTable, 0, 0, 1, 1, 0, MIN_SIZE, false).
		AddItem(sidebar, 0, 1, 1, 1, 0, MIN_SIZE, false).
		AddItem(inputGrid, 1, 0, 1, 2, 0, 0, true).
		AddItem(bottomRow, 2, 0, 1, 2, 0, 0, false)

	gui.pages = tview.NewPages().
		AddPage("background", gui.grid, true, true)
	//AddPage("modal", modal(box, 40, 10), true, true)

	app := tview.NewApplication()
	app.SetRoot(gui.pages, true)
	app.EnableMouse(true)
	app.SetFocus(gui.input)
	app.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyCtrlR:
			gui.updateDeviceListAsync(true)
		case tcell.KeyCtrlN:
			gui.showModalFavNew()
		case tcell.KeyCtrlL:
			gui.clearMessages()
		case tcell.KeyCtrlD:
			_ = gui.disconnect()
		case tcell.KeyCtrlH:
			gui.appConfig.Config.ShowHex = !gui.appConfig.Config.ShowHex
		case tcell.KeyCtrlQ:
			gui.quit()
		default:
		}
		return event
	})

	gui.app = app

	if len(appConfig.Config.SerialConfig.SerialPort) > 0 {
		serialPorts, serialError := serialcon.GetSerialPorts()
		if serialError == nil {
			if slices.Contains(serialPorts, appConfig.Config.SerialConfig.SerialPort) {
				_ = gui.connect(appConfig.Config.SerialConfig)
			} else {
				gui.appConfig.Config.SerialConfig.SerialPort = ""
			}
		}
	}

	err := app.Run()
	if err != nil {
		panic(err)
	}

	return gui, nil
}

func (g *Gui) quit() {
	g.app.Stop()
	_ = g.disconnect()
	os.Exit(0)
}

func (g *Gui) connect(serialConfig config.SerialConfig) error {
	if g.serialConnection == nil {
		g.serialConnection = serialcon.New()
	} else if g.serialConnection.IsConnected() {
		portName, _ := g.serialConnection.GetCurrentPortName()
		baudrate := g.serialConnection.GetCurrentBaudrate()

		if serialConfig.SerialPort == portName && serialConfig.SerialMode.BaudRate == baudrate {
			log.Println("Port already connected")
			return nil
		}

		_ = g.disconnect()
	}

	err := g.serialConnection.Open(serialConfig, g.serialDataCallback)
	if err != nil {
		g.addTableDataAsync(
			fmt.Sprintf("Could not connect to %s", serialConfig.SerialPort),
			MessageStatus,
		)
		log.Printf("Could not connect to serial device \"%s\".\n", serialConfig.SerialPort)
		return err
	}

	g.messagesTable.SetTitle(" Messages ")
	g.addTableDataAsync(
		fmt.Sprintf("Connected to port %s at %d baud", serialConfig.SerialPort, serialConfig.SerialMode.BaudRate),
		MessageStatus,
	)
	return nil
}

func (g *Gui) serialDataCallback(message string) {
	g.addTableDataAsync(message, MessageIn)
}

func (g *Gui) addTableDataAsync(data string, messageType MessageType) {
	go g.addTableData(data, messageType)
}

func (g *Gui) addTableData(data string, messageType MessageType) {
	g.app.QueueUpdateDraw(func() {

		cellColor := tcell.ColorWhite
		selectable := true

		switch messageType {
		case MessageIn:
			cellColor = tcell.ColorWhite
			if g.appConfig.Config.ShowHex {
				data = hex.GetHexString(data)
			}
		case MessageOut:
			cellColor = tcell.ColorYellow
			if g.appConfig.Config.ShowHex {
				data = hex.GetHexString(data)
			}
		case MessageStatus:
			cellColor = tcell.ColorGrey
			selectable = false
		case MessageError:
			cellColor = tcell.ColorRed
			selectable = false
		}

		row := g.messagesTable.GetRowCount()

		g.messagesTable.SetCell(row, 0, tview.
			NewTableCell(time.Now().Format("15:04:05.000")).
			SetTextColor(tcell.ColorGray).
			SetSelectable(false),
		)

		g.messagesTable.SetCell(row, 1, tview.
			NewTableCell(data).
			SetTextColor(cellColor).
			SetSelectable(selectable),
		)

		//if g.appConfig.Config.ShowHex {
		//	g.messagesTable.SetCell(row, 2, tview.
		//		NewTableCell(hex.GetHexString(data)))
		//}
	})
}

func (g *Gui) sendMessage(text string) {
	if len(text) <= 0 {
		return
	}

	text = hex.ReplaceHexValuesToLatin1(text)

	bytes := []byte(text)

	go func() {
		if g.serialConnection == nil {
			log.Println("No connection when tried to send message.")
			return
		}
		if !g.serialConnection.IsConnected() {
			log.Println("Not connected to a serial device.")
			return
		}
		write, err := g.serialConnection.Write(bytes)
		if err != nil {
			log.Println("Could not write to serial port.")
			g.addTableDataAsync("Could not write to serial port.", MessageError)
			return
		}
		g.addTableDataAsync(text, MessageOut)
		log.Printf("Wrote %d bytes.", write)
	}()
}

func (g *Gui) getInputGrid() *tview.Grid {
	input := tview.NewInputField()
	input.SetDoneFunc(func(key tcell.Key) {
		g.sendMessage(input.GetText())
	})

	sendButton := tview.NewButton("Send")
	sendButton.SetSelectedFunc(func() {
		g.sendMessage(input.GetText())
	})

	inputGrid := tview.NewGrid()
	inputGrid.SetGap(0, 1)
	inputGrid.AddItem(input, 0, 0, 1, 1, 0, 0, false)
	inputGrid.AddItem(sendButton, 0, 1, 1, 1, 0, 0, false)

	inputGrid.SetRows(1)
	inputGrid.SetColumns(0, 10)
	inputGrid.SetTitle(" Send message ")
	inputGrid.SetTitleAlign(tview.AlignLeft)
	inputGrid.SetBorder(true)

	g.input = input

	return inputGrid
}

func (g *Gui) getBottomRow() *tview.Grid {
	bottomTextLeft := tview.NewTextView().SetText(SHORTCUTS)
	bottomTextRight := tview.NewTextView().
		SetText(fmt.Sprintf("YSM %s", DEFAULT_VERSION)).
		SetTextAlign(tview.AlignRight)

	bottomRow := tview.NewGrid().
		SetColumns(0, 10).
		AddItem(bottomTextLeft, 0, 0, 1, 1, 0, 0, false).
		AddItem(bottomTextRight, 0, 1, 1, 1, 0, 0, false)
	return bottomRow
}

func (g *Gui) getSidebar() *tview.Grid {

	g.deviceList = g.getDeviceList()
	g.baudrateList = g.getBaudrateList()
	g.favorites = NewFavorites(&g.appConfig.Config,
		func() {
			g.showModalFavNew()
		},
		func(i int, favorite config.MessageFavorite) {
			g.showModalFavEdit(g.appConfig.Config.GetMessageFavorites()[i])
		},
		func(message string) {
			g.sendMessage(message)
		},
	)

	grid := tview.NewGrid()
	grid.SetRows(0, 3, 3)
	grid.AddItem(g.favorites.GetPrimitive(), 0, 0, 1, 1, 0, 0, false)
	grid.AddItem(g.baudrateList, 1, 0, 1, 1, 0, 0, false)
	grid.AddItem(g.deviceList, 2, 0, 1, 1, 0, 0, false)

	g.updateDeviceListAsync(false)

	return grid
}

func (g *Gui) updateDeviceListAsync(showUpdateMessage bool) {
	go g.updateDeviceList(showUpdateMessage)
}

func (g *Gui) updateDeviceList(showUpdateMessage bool) {
	g.app.QueueUpdateDraw(func() {
		optionCount := g.deviceList.GetOptionCount()
		for i := 0; i < optionCount; i++ {
			g.deviceList.RemoveOption(0)
		}

		serialPorts, err := serialcon.GetSerialPorts()
		if err != nil {
			return
		}

		selectedOption := -1

		for index, serialPort := range serialPorts {
			label := serialPort
			g.deviceList.AddOption(label, nil)

			if g.serialConnection != nil {
				currentPort, _ := g.serialConnection.GetCurrentPortName()
				if serialPort == currentPort {
					selectedOption = index
				}
			}
		}

		if selectedOption > -1 {
			g.deviceList.SetCurrentOption(selectedOption)
		}

		if showUpdateMessage {
			message := fmt.Sprintf("Refreshed device list. %d devices found.", g.deviceList.GetOptionCount())
			g.addTableDataAsync(message, MessageStatus)
		}
	})
}

func (g *Gui) getBaudrateList() *tview.DropDown {
	list := tview.NewDropDown()
	list.SetBorder(true)
	list.SetTitle(" Baudrate ")
	list.SetFieldWidth(50)
	list.SetTextOptions("", "", "", "", "- no baudrate selected -")

	baudrates := serialcon.GetBaudrates()
	currentIndex := -1

	for i, baudrate := range baudrates {
		if baudrate == g.appConfig.Config.SerialConfig.SerialMode.BaudRate {
			currentIndex = i
		}
		list.AddOption(strconv.Itoa(baudrate), nil)
	}

	list.SetCurrentOption(currentIndex)

	list.SetSelectedFunc(func(text string, index int) {
		if index < 0 {
			return
		}

		baudrate, err := strconv.Atoi(text)
		if err != nil {
			return
		}

		if baudrate == g.appConfig.Config.SerialConfig.SerialMode.BaudRate {
			return
		}

		g.appConfig.Config.SerialConfig.SerialMode.BaudRate = baudrate
		err = g.connect(g.appConfig.Config.SerialConfig)
		if err != nil {
			return
		}
	})

	return list
}

func (g *Gui) getDeviceList() *tview.DropDown {
	deviceList := tview.NewDropDown()
	deviceList.SetBorder(true)
	deviceList.SetTitle(" Device ")
	deviceList.SetFieldWidth(50)
	deviceList.SetTextOptions("", "", "", " (connected)", "- not port selected -")
	deviceList.SetSelectedFunc(func(text string, index int) {
		log.Println("selected: ", index)
		//log.Println("port1: ", g.appConfig.Config.SerialConfig.SerialPort)
		//log.Println("port2: ", text)
		if index < 0 {
			return
		}

		if g.appConfig.Config.SerialConfig.SerialPort == text {
			return
		}

		g.appConfig.Config.SerialConfig.SerialPort = text
		err := g.connect(g.appConfig.Config.SerialConfig)
		if err != nil {
			go deviceList.SetCurrentOption(-1)
			return
		}
	})
	return deviceList
}

func (g *Gui) getMessagesTable() *tview.Table {
	table := tview.NewTable()
	table.SetBorders(false)
	table.SetSelectable(true, true)
	table.SetBorder(true)
	table.SetTitle(" Messages ")
	table.SetTitleAlign(tview.AlignLeft)
	//table.SetBorderPadding(0, 0, 0, 0)
	return table
}

func (g *Gui) showModalFavNew() {
	message := g.input.GetText()
	modal := NewModalFavNew(
		message,
		func(fav config.MessageFavorite) {
			g.appConfig.Config.AddMessageFavorite(&fav, true)
			g.hideModal()
			g.favorites.RenderList(g.appConfig.Config.GetMessageFavoriteIndex(&fav))
		},
		g.hideModal,
	)

	g.pages.AddPage(PAGENAME_MODAL, modal.GetPrimitive(), true, true)
}

func (g *Gui) showModalFavEdit(favorite *config.MessageFavorite) {
	modal := NewModalFavEdit(
		favorite,
		func() {
			g.hideModal()
			g.favorites.RenderList(g.appConfig.Config.GetMessageFavoriteIndex(favorite))
		},
		g.hideModal,
	)

	g.pages.AddPage(PAGENAME_MODAL, modal.GetPrimitive(), true, true)
}

func (g *Gui) hideModal() {
	if g.pages.HasPage(PAGENAME_MODAL) {
		g.pages.RemovePage(PAGENAME_MODAL)
	}
}

func (g *Gui) clearMessages() {
	log.Println("clearMessages()")
	g.messagesTable.Clear()
}

func (g *Gui) disconnect() error {
	g.deviceList.SetCurrentOption(-1)
	g.appConfig.Config.SerialConfig.SerialPort = ""

	if g.serialConnection.IsConnected() {
		portName, _ := g.serialConnection.GetCurrentPortName()
		err := g.serialConnection.Close()
		if err != nil {
			return err
		}

		g.addTableDataAsync(
			fmt.Sprintf("Disconnected from %s", portName),
			MessageStatus,
		)
	}

	return nil
}
