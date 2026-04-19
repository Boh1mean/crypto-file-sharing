package ui

import (
	"context"
	"errors"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"

	clientapp "cryptocore/internal/client/app"
	"cryptocore/internal/client/keystore"
)

func Run() error {
	store, err := keystore.NewDefaultStore()
	if err != nil {
		return err
	}

	registerService := clientapp.NewRegisterUserService(store)
	loginService := clientapp.NewLoginService(store)
	sendService := clientapp.NewSendFileService(store)
	receiveService := clientapp.NewReceiveFileService(store)

	a := app.New()
	window := a.NewWindow("CryptoCore")
	window.Resize(fyne.NewSize(800, 500))

	// content — контейнер с одним дочерним элементом, который мы заменяем при навигации.
	content := container.NewStack()

	var showMain func()
	var showAuth func()

	showMain = func() {
		main := NewMainScreen(window, sendService, receiveService, loginService, showAuth)
		content.Objects = []fyne.CanvasObject{main}
		content.Refresh()
	}

	showAuth = func() {
		profile, loadErr := store.Load()
		hasProfile := loadErr == nil

		auth := NewAuthScreen(window, registerService, loginService, hasProfile, showMain)
		content.Objects = []fyne.CanvasObject{auth}
		content.Refresh()

		// Автовход: если профиль есть и сессия ещё не истекла — сразу на главный экран.
		if hasProfile && profile.HasValidSession() {
			showMain()
		}
	}

	// Если профиль есть, но токен истёк — тихий повторный вход в фоне.
	profile, loadErr := store.Load()
	if loadErr == nil && !profile.HasValidSession() {
		go func() {
			_, loginErr := loginService.Login(context.Background())
			fyne.Do(func() {
				if loginErr == nil {
					showMain()
				} else {
					showAuth()
				}
			})
		}()
		// Пока идёт фоновый вход показываем экран авторизации без автоперехода.
		auth := NewAuthScreen(window, registerService, loginService, true, showMain)
		content.Objects = []fyne.CanvasObject{auth}
	} else {
		// Нет профиля или сессия валидна — showAuth разберётся сам.
		showAuth()
	}

	window.SetContent(content)
	window.ShowAndRun()

	return errors.New("desktop app closed")
}
