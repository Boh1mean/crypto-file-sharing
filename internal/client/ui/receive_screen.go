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

func NewReceiveScreen(window fyne.Window, receiveService *clientapp.ReceiveFileService) fyne.CanvasObject {
	fileIDEntry := widget.NewEntry()
	fileIDEntry.SetPlaceHolder("1001")
	fileIDEntry.SetText("1001")

	outputDirEntry := widget.NewEntry()
	outputDirEntry.SetPlaceHolder("/path/to/output-directory")

	chooseDirButton := widget.NewButton("Choose Output Folder", func() {
		folderDialog := dialog.NewFolderOpen(func(uri fyne.ListableURI, err error) {
			if err != nil {
				dialog.ShowError(err, window)
				return
			}
			if uri == nil {
				return
			}
			outputDirEntry.SetText(uri.Path())
		}, window)
		folderDialog.Show()
	})

	statusLabel := widget.NewLabel("Download an encrypted container from the server and decrypt it locally.")

	var receiveButton *widget.Button
	receiveButton = widget.NewButton("Download And Decrypt", func() {
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

		outputDir := strings.TrimSpace(outputDirEntry.Text)
		if outputDir == "" {
			dialog.ShowError(errMessage("output directory is required"), window)
			return
		}

		receiveButton.Disable()
		statusLabel.SetText("Downloading container and decrypting file...")

		go func() {
			out, err := receiveService.Receive(context.Background(), clientapp.ReceiveFileInput{
				FileID:    fileID,
				OutputDir: outputDir,
			})

			fyne.Do(func() {
				receiveButton.Enable()
				if err != nil {
					statusLabel.SetText("Receive failed.")
					dialog.ShowError(err, window)
					return
				}

				statusLabel.SetText("File decrypted and saved successfully.")
				dialog.ShowInformation("Success", "File saved to:\n"+out.OutputFilePath, window)
			})
		}()
	})

	form := widget.NewForm(
		widget.NewFormItem("File ID", fileIDEntry),
		widget.NewFormItem("Output Folder", outputDirEntry),
	)

	return container.NewVBox(
		widget.NewLabel("Receive File"),
		widget.NewLabel("This downloads the encrypted container, verifies the sender signature and decrypts locally."),
		form,
		chooseDirButton,
		receiveButton,
		statusLabel,
	)
}
