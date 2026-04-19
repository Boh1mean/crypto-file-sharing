package ui

import (
	"context"
	"strconv"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/storage"
	"fyne.io/fyne/v2/widget"

	clientapp "cryptocore/internal/client/app"
)

func NewSendScreen(window fyne.Window, sendService *clientapp.SendFileService) fyne.CanvasObject {
	fileIDEntry := widget.NewEntry()
	fileIDEntry.SetPlaceHolder("1001")
	fileIDEntry.SetText("1001")

	recipientIDEntry := widget.NewEntry()
	recipientIDEntry.SetPlaceHolder("2")
	recipientIDEntry.SetText("2")

	filePathEntry := widget.NewEntry()
	filePathEntry.SetPlaceHolder("/path/to/file.txt")

	chooseFileButton := widget.NewButton("Choose File", func() {
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

	statusLabel := widget.NewLabel("Choose a file and send it to a registered recipient.")

	var sendButton *widget.Button
	sendButton = widget.NewButton("Encrypt And Send", func() {
		rawFileID := strings.TrimSpace(fileIDEntry.Text)
		if rawFileID == "" {
			dialog.ShowError(errMessage("file ID is required"), window)
			return
		}

		fileID, err := strconv.Atoi(rawFileID)
		if err != nil {
			dialog.ShowError(errMessage("file ID must be a number"), window)
			return
		}

		rawRecipientID := strings.TrimSpace(recipientIDEntry.Text)
		if rawRecipientID == "" {
			dialog.ShowError(errMessage("recipient ID is required"), window)
			return
		}

		recipientID, err := strconv.Atoi(rawRecipientID)
		if err != nil {
			dialog.ShowError(errMessage("recipient ID must be a number"), window)
			return
		}

		filePath := strings.TrimSpace(filePathEntry.Text)
		if filePath == "" {
			dialog.ShowError(errMessage("file path is required"), window)
			return
		}

		sendButton.Disable()
		statusLabel.SetText("Encrypting file and uploading container...")

		go func() {
			_, err := sendService.Send(context.Background(), clientapp.SendFileInput{
				FileID:      fileID,
				RecipientID: recipientID,
				FilePath:    filePath,
			})

			fyne.Do(func() {
				sendButton.Enable()
				if err != nil {
					statusLabel.SetText("Send failed.")
					dialog.ShowError(err, window)
					return
				}

				statusLabel.SetText("File encrypted and uploaded successfully.")
				dialog.ShowInformation("Success", "Encrypted container uploaded to the server.", window)
			})
		}()
	})

	form := widget.NewForm(
		widget.NewFormItem("File ID", fileIDEntry),
		widget.NewFormItem("Recipient ID", recipientIDEntry),
		widget.NewFormItem("File Path", filePathEntry),
	)

	return container.NewVBox(
		widget.NewLabel("Send File"),
		widget.NewLabel("This uses your locally stored private keys and sends only an encrypted container."),
		form,
		chooseFileButton,
		sendButton,
		statusLabel,
	)
}
