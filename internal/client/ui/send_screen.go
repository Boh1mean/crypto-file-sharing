package ui

import (
	"context"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/storage"
	"fyne.io/fyne/v2/widget"

	clientapp "cryptocore/internal/client/app"
)

func NewSendScreen(window fyne.Window, sendService *clientapp.SendFileService) fyne.CanvasObject {
	recipientEntry := widget.NewEntry()
	recipientEntry.SetPlaceHolder("никнейм получателя")

	filePathEntry := widget.NewEntry()
	filePathEntry.SetPlaceHolder("/путь/к/файлу")

	chooseFileButton := widget.NewButton("Выбрать файл", func() {
		fileDialog := dialog.NewFileOpen(func(reader fyne.URIReadCloser, err error) {
			if err != nil {
				dialog.ShowError(err, window)
				return
			}
			if reader == nil {
				return
			}
			filePathEntry.SetText(reader.URI().Path())
			_ = reader.Close()
		}, window)
		fileDialog.SetFilter(storage.NewExtensionFileFilter([]string{".txt", ".pdf", ".jpg", ".jpeg", ".png", ".bin"}))
		fileDialog.Show()
	})

	statusLabel := widget.NewLabel("Выберите файл и введите никнейм получателя.")

	var sendButton *widget.Button
	sendButton = widget.NewButton("Зашифровать и отправить", func() {
		recipient := strings.TrimSpace(recipientEntry.Text)
		if recipient == "" {
			dialog.ShowError(errMessage("никнейм получателя не может быть пустым"), window)
			return
		}

		filePath := strings.TrimSpace(filePathEntry.Text)
		if filePath == "" {
			dialog.ShowError(errMessage("выберите файл"), window)
			return
		}

		sendButton.Disable()
		statusLabel.SetText("Шифрование и загрузка на сервер...")

		go func() {
			out, err := sendService.Send(context.Background(), clientapp.SendFileInput{
				RecipientUsername: recipient,
				FilePath:          filePath,
			})

			fyne.Do(func() {
				sendButton.Enable()
				if err != nil {
					statusLabel.SetText("Ошибка отправки.")
					dialog.ShowError(err, window)
					return
				}

				statusLabel.SetText("Файл успешно зашифрован и отправлен.")
				dialog.ShowInformation("Успех",
					"Файл отправлен!\nID файла для получателя: "+itoa(out.FileID),
					window)
			})
		}()
	})

	form := widget.NewForm(
		widget.NewFormItem("Получатель", recipientEntry),
		widget.NewFormItem("Файл", filePathEntry),
	)

	return container.NewVBox(
		widget.NewLabel("Отправить файл"),
		widget.NewLabel("Файл будет зашифрован локально перед отправкой."),
		form,
		chooseFileButton,
		sendButton,
		statusLabel,
	)
}
