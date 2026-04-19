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

// NewAuthScreen возвращает экран авторизации с возможностью переключаться
// между режимами "Sign In" и "Register" прямо внутри экрана.
func NewAuthScreen(
	window fyne.Window,
	registerService *clientapp.RegisterUserService,
	loginService *clientapp.LoginService,
	hasExistingProfile bool,
	onSuccess func(),
) fyne.CanvasObject {
	// stack — единственный дочерний объект, который мы заменяем при переключении режима.
	stack := container.NewStack()

	var showLogin func()
	var showRegister func()

	showLogin = func() {
		stack.Objects = []fyne.CanvasObject{
			buildLoginView(window, loginService, onSuccess, showRegister),
		}
		stack.Refresh()
	}

	showRegister = func() {
		stack.Objects = []fyne.CanvasObject{
			buildRegisterView(window, registerService, onSuccess, showLogin),
		}
		stack.Refresh()
	}

	// Начальный режим зависит от наличия профиля.
	if hasExistingProfile {
		showLogin()
	} else {
		showRegister()
	}

	return stack
}

// buildLoginView строит вид входа с кнопкой-ссылкой на регистрацию.
func buildLoginView(
	window fyne.Window,
	loginService *clientapp.LoginService,
	onSuccess func(),
	onSwitchToRegister func(),
) fyne.CanvasObject {
	statusLabel := widget.NewLabel("")

	var loginButton *widget.Button
	loginButton = widget.NewButton("Sign In", func() {
		loginButton.Disable()
		statusLabel.SetText("Authenticating...")

		go func() {
			_, err := loginService.Login(context.Background())

			fyne.Do(func() {
				loginButton.Enable()
				if err != nil {
					statusLabel.SetText("Authentication failed.")
					dialog.ShowError(err, window)
					return
				}
				onSuccess()
			})
		}()
	})
	loginButton.Importance = widget.HighImportance

	switchButton := widget.NewButton("Create new account", onSwitchToRegister)
	switchButton.Importance = widget.LowImportance

	return container.NewCenter(
		container.NewVBox(
			widget.NewLabelWithStyle("CryptoCore", fyne.TextAlignCenter, fyne.TextStyle{Bold: true}),
			widget.NewLabel("Welcome back. Your keys are stored locally."),
			widget.NewSeparator(),
			loginButton,
			statusLabel,
			widget.NewSeparator(),
			switchButton,
		),
	)
}

// buildRegisterView строит вид регистрации с кнопкой-ссылкой на вход.
func buildRegisterView(
	window fyne.Window,
	registerService *clientapp.RegisterUserService,
	onSuccess func(),
	onSwitchToLogin func(),
) fyne.CanvasObject {
	usernameEntry := widget.NewEntry()
	usernameEntry.SetPlaceHolder("например: alice")

	statusLabel := widget.NewLabel("")

	var registerButton *widget.Button
	registerButton = widget.NewButton("Create Account", func() {
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
		statusLabel.SetText("Generating keys and registering...")

		go func() {
			out, err := registerService.Register(context.Background(), clientapp.RegisterUserInput{
				Username: username,
			})

			fyne.Do(func() {
				registerButton.Enable()
				if err != nil {
					statusLabel.SetText("Registration failed.")
					dialog.ShowError(err, window)
					return
				}
				statusLabel.SetText("Registered! Your ID: " + itoa(out.UserID))
				onSuccess()
			})
		}()
	})
	registerButton.Importance = widget.HighImportance

	switchButton := widget.NewButton("Already have an account? Sign In", onSwitchToLogin)
	switchButton.Importance = widget.LowImportance

	form := widget.NewForm(
		widget.NewFormItem("Никнейм", usernameEntry),
	)

	return container.NewCenter(
		container.NewVBox(
			widget.NewLabelWithStyle("CryptoCore", fyne.TextAlignCenter, fyne.TextStyle{Bold: true}),
			widget.NewLabel("Create your account. Keys are generated locally and never leave your device."),
			widget.NewSeparator(),
			form,
			registerButton,
			statusLabel,
			widget.NewSeparator(),
			switchButton,
		),
	)
}
