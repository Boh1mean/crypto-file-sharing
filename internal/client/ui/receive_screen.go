package ui

import (
	"context"
	"fmt"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	clientapp "cryptocore/internal/client/app"
)

const pollInterval = 30 * time.Second

func NewReceiveScreen(window fyne.Window, receiveService *clientapp.ReceiveFileService) fyne.CanvasObject {
	// --- состояние экрана ---
	var items []clientapp.InboxItem

	// --- выбор папки для сохранения ---
	outputDirEntry := widget.NewEntry()
	outputDirEntry.SetPlaceHolder("Choose output folder...")

	chooseDirButton := widget.NewButton("Browse...", func() {
		dialog.NewFolderOpen(func(uri fyne.ListableURI, err error) {
			if err != nil {
				dialog.ShowError(err, window)
				return
			}
			if uri != nil {
				outputDirEntry.SetText(uri.Path())
			}
		}, window).Show()
	})

	// --- список входящих ---
	statusLabel := widget.NewLabel("Loading inbox...")

	list := widget.NewList(
		func() int { return len(items) },
		func() fyne.CanvasObject {
			// Шаблон строки: иконка + имя файла, отправитель, размер, дата, кнопка скачать
			nameLabel := widget.NewLabel("filename.pdf")
			nameLabel.TextStyle = fyne.TextStyle{Bold: true}
			nameLabel.Truncation = fyne.TextTruncateEllipsis

			senderLabel := widget.NewLabel("from: username")
			sizeLabel := widget.NewLabel("0 KB")
			dateLabel := widget.NewLabel("Jan 02 15:04")

			downloadBtn := widget.NewButtonWithIcon("", theme.DownloadIcon(), func() {})
			downloadBtn.Importance = widget.LowImportance

			left := container.NewVBox(nameLabel, senderLabel)
			right := container.NewVBox(
				container.NewHBox(sizeLabel, dateLabel),
				container.NewHBox(widget.NewLabel(""), downloadBtn),
			)
			return container.NewBorder(nil, nil, nil, right, left)
		},
		func(id widget.ListItemID, obj fyne.CanvasObject) {
			if id >= len(items) {
				return
			}
			item := items[id]

			row := obj.(*fyne.Container)
			left := row.Objects[0].(*fyne.Container)
			right := row.Objects[1].(*fyne.Container)

			nameLabel := left.Objects[0].(*widget.Label)
			senderLabel := left.Objects[1].(*widget.Label)

			topRow := right.Objects[0].(*fyne.Container)
			botRow := right.Objects[1].(*fyne.Container)

			sizeLabel := topRow.Objects[0].(*widget.Label)
			dateLabel := topRow.Objects[1].(*widget.Label)
			downloadBtn := botRow.Objects[1].(*widget.Button)

			nameLabel.SetText(item.FileName)
			senderLabel.SetText("from: " + item.SenderUsername)
			sizeLabel.SetText(formatFileSize(item.Size))
			dateLabel.SetText(item.CreatedAt.Local().Format("Jan 02 15:04"))

			// Захватываем локальную копию для замыкания кнопки.
			captured := item
			downloadBtn.OnTapped = func() {
				outputDir := outputDirEntry.Text
				if outputDir == "" {
					dialog.ShowError(errMessage("please choose an output folder first"), window)
					return
				}

				downloadBtn.Disable()
				go func() {
					out, err := receiveService.Receive(context.Background(), clientapp.ReceiveFileInput{
						FileID:    captured.ID,
						OutputDir: outputDir,
					})
					fyne.Do(func() {
						downloadBtn.Enable()
						if err != nil {
							dialog.ShowError(err, window)
							return
						}
						dialog.ShowInformation("Saved",
							fmt.Sprintf("File saved to:\n%s", out.OutputFilePath), window)
					})
				}()
			}
		},
	)

	// --- обновление списка ---
	refresh := func() {
		statusLabel.SetText("Refreshing...")
		go func() {
			loaded, err := receiveService.ListInbox(context.Background())
			fyne.Do(func() {
				if err != nil {
					statusLabel.SetText("Failed to load inbox.")
					return
				}
				items = loaded
				list.Refresh()
				if len(items) == 0 {
					statusLabel.SetText("Inbox is empty.")
				} else {
					statusLabel.SetText(fmt.Sprintf("%d file(s) waiting for you.", len(items)))
				}
			})
		}()
	}

	refreshButton := widget.NewButtonWithIcon("Refresh", theme.ViewRefreshIcon(), refresh)

	// --- polling каждые 30 секунд ---
	go func() {
		ticker := time.NewTicker(pollInterval)
		defer ticker.Stop()
		for range ticker.C {
			fyne.Do(refresh)
		}
	}()

	// Первоначальная загрузка
	refresh()

	// --- сборка экрана ---
	folderRow := container.NewBorder(nil, nil, nil, chooseDirButton, outputDirEntry)

	topBar := container.NewBorder(nil, nil, nil, refreshButton,
		widget.NewLabelWithStyle("Inbox", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
	)

	bottom := container.NewVBox(
		widget.NewSeparator(),
		widget.NewLabel("Save to folder:"),
		folderRow,
		statusLabel,
	)

	return container.NewBorder(topBar, bottom, nil, nil, list)
}

func formatFileSize(bytes int64) string {
	switch {
	case bytes < 1024:
		return fmt.Sprintf("%d B", bytes)
	case bytes < 1024*1024:
		return fmt.Sprintf("%.1f KB", float64(bytes)/1024)
	default:
		return fmt.Sprintf("%.1f MB", float64(bytes)/(1024*1024))
	}
}
