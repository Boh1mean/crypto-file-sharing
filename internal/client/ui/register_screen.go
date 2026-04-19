package ui

import (
	"context"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"

	clientapp "cryptocore/internal/client/app"
)

func NewRegisterScreen(window fyne.Window, registerService *clientapp.RegisterUserService) fyne.CanvasObject {
	usernameEntry := widget.NewEntry()
	usernameEntry.SetPlaceHolder("например: alice")

	statusLabel := widget.NewLabel("Введите никнейм и нажмите Register.")

	var registerButton *widget.Button
	registerButton = widget.NewButton("Generate Keys And Register", func() {
		username := strings.TrimSpace(usernameEntry.Text)
		if username == "" {
			dialog.ShowError(errMessage("никнейм не может быть пустым"), window)
			return
		}
		if len([]rune(username)) > 32 {
			dialog.ShowError(errMessage("никнейм не может быть длиннее 32 символов"), window)
			return
		}

		registerButton.Disable()
		statusLabel.SetText("Регистрация пользователя и сохранение локального профиля...")

		go func() {
			out, err := registerService.Register(context.Background(), clientapp.RegisterUserInput{
				Username: username,
			})

			fyne.Do(func() {
				registerButton.Enable()
				if err != nil {
					statusLabel.SetText("Регистрация не удалась.")
					dialog.ShowError(err, window)
					return
				}

				statusLabel.SetText("Регистрация завершена. Ключи сохранены локально.")
				dialog.ShowInformation("Успех",
					"Пользователь зарегистрирован!\nВаш ID: "+itoa(out.UserID),
					window)
			})
		}()
	})

	form := widget.NewForm(
		widget.NewFormItem("Никнейм", usernameEntry),
	)

	return container.NewVBox(
		widget.NewLabel("CryptoCore Desktop"),
		widget.NewLabel("Шаг 1: введите никнейм, сгенерируйте ключи и зарегистрируйтесь на сервере."),
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
