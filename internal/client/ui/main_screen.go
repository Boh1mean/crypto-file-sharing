package ui

import (
	"context"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"

	clientapp "cryptocore/internal/client/app"
)

// NewMainScreen возвращает главный экран с табами Send/Receive и кнопкой выхода.
// onLogout вызывается после выхода — переключает UI обратно на AuthScreen.
func NewMainScreen(
	window fyne.Window,
	sendService *clientapp.SendFileService,
	receiveService *clientapp.ReceiveFileService,
	loginService *clientapp.LoginService,
	onLogout func(),
) fyne.CanvasObject {
	tabs := container.NewAppTabs(
		container.NewTabItem("Send", NewSendScreen(window, sendService)),
		container.NewTabItem("Receive", NewReceiveScreen(window, receiveService)),
	)

	logoutButton := widget.NewButton("Sign Out", func() {
		dialog.ShowConfirm(
			"Sign Out",
			"Are you sure you want to sign out?\nYou will need to authenticate again next time.",
			func(confirmed bool) {
				if !confirmed {
					return
				}
				if err := loginService.Logout(context.Background()); err != nil {
					dialog.ShowError(err, window)
					return
				}
				onLogout()
			},
			window,
		)
	})

	top := container.NewBorder(nil, nil, nil, logoutButton,
		widget.NewLabelWithStyle("CryptoCore", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
	)

	return container.NewBorder(top, nil, nil, nil, tabs)
}
