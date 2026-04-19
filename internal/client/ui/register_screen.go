package ui

import (
	"context"
	"strconv"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"

	clientapp "cryptocore/internal/client/app"
)

func NewRegisterScreen(window fyne.Window, registerService *clientapp.RegisterUserService) fyne.CanvasObject {
	serverEntry := widget.NewEntry()
	serverEntry.SetPlaceHolder("http://localhost:8080")
	serverEntry.SetText("http://localhost:8080")

	userIDEntry := widget.NewEntry()
	userIDEntry.SetPlaceHolder("1")
	userIDEntry.SetText("1")

	statusLabel := widget.NewLabel("Enter server URL and user ID, then register.")

	var registerButton *widget.Button
	registerButton = widget.NewButton("Generate Keys And Register", func() {
		serverURL := strings.TrimSpace(serverEntry.Text)
		if serverURL == "" {
			dialog.ShowError(errMessage("server URL is required"), window)
			return
		}

		rawUserID := strings.TrimSpace(userIDEntry.Text)
		if rawUserID == "" {
			dialog.ShowError(errMessage("user ID is required"), window)
			return
		}

		userID, err := strconv.Atoi(rawUserID)
		if err != nil {
			dialog.ShowError(errMessage("user ID must be a number"), window)
			return
		}

		registerButton.Disable()
		statusLabel.SetText("Registering user and saving local profile...")

		go func() {
			_, err := registerService.Register(context.Background(), clientapp.RegisterUserInput{
				ServerURL: serverURL,
				UserID:    userID,
			})

			fyne.Do(func() {
				registerButton.Enable()
				if err != nil {
					statusLabel.SetText("Registration failed.")
					dialog.ShowError(err, window)
					return
				}

				statusLabel.SetText("Registration completed. Keys are stored locally.")
				dialog.ShowInformation("Success", "User registered and local keys saved.", window)
			})
		}()
	})

	form := widget.NewForm(
		widget.NewFormItem("Server URL", serverEntry),
		widget.NewFormItem("User ID", userIDEntry),
	)

	return container.NewVBox(
		widget.NewLabel("CryptoCore Desktop"),
		widget.NewLabel("First step: generate local keys and register only public keys on the server."),
		form,
		registerButton,
		statusLabel,
	)
}

func errMessage(message string) error {
	return simpleError(message)
}

type simpleError string

func (e simpleError) Error() string {
	return string(e)
}
