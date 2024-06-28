package main

import (
	"errors"
	"flag"
	"fmt"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
	"log"
	"os"
	"shadercompat/repository"
	"shadercompat/runtime"
	"strings"
)

var (
	repositoryPath = flag.String("repository", "./repository", "Path to local repository")
)

func main() {
	flag.Parse()

	// Init app
	a := app.New()
	w := a.NewWindow("ShaderCompat")
	w.SetFixedSize(true)
	w.CenterOnScreen()

	context := &runtime.Context{
		LocalPath: *repositoryPath,
		Window:    w,
	}

	waitBar := widget.NewProgressBarInfinite()
	w.SetContent(container.NewCenter(container.NewVBox(
		waitBar,
		widget.NewLabel("Loading"),
	)))
	waitBar.Resize(fyne.NewSize(100, 20))

	go setup(context)

	w.ShowAndRun()
}

func setup(context *runtime.Context) {
	// Get repository version
	updateLocalVersion(context)

	// Get repository version
	version, err := repository.GetVersion()
	if err != nil {
		context.RemoteVersion = runtime.UnknownVersion
		context.ShowError(fmt.Errorf("failed to get repository version: %v", err))
	} else {
		context.RemoteVersion = version
	}

	log.Printf("Local version: %s; Repository version: %s\n", context.LocalVersion, context.RemoteVersion)

	// Load mappings
	if err := context.LoadMappings(); err != nil {
		context.ShowFatal(err)
	}

	// Main view
	versionView := runtime.NewVersionView()
	context.VersionView = versionView

	versionCanvas := versionView.Build(context)
	if err := versionView.Update(context); err != nil {
		context.ShowFatal(err)
	}

	shaderView := runtime.NewShaderView()
	context.ShaderView = shaderView
	shaderCanvas := shaderView.Build(context)

	modsView := runtime.NewModsView()
	context.ModsView = modsView
	modsCanvas := modsView.Build(context)

	patchView := runtime.NewPatchView()
	context.PatchView = patchView
	patchCanvas := patchView.Build(context)

	mainView := container.NewVBox(
		container.NewHBox(
			container.NewVBox(
				runtime.Header("Repository"),
				versionCanvas,
				runtime.Header("Shader"),
				shaderCanvas,
			),
			widget.NewSeparator(),
			container.NewVBox(
				runtime.Header("Mods"),
				modsCanvas,
				layout.NewSpacer(),
				widget.NewSeparator(),
				runtime.Header("Patch"),
				patchCanvas,
			),
		),
	)

	context.Window.SetContent(mainView)

}

func updateLocalVersion(context *runtime.Context) {
	versionFilePath := fmt.Sprintf("%s/%s", context.LocalPath, repository.VersionPath)
	if f, err := os.Stat(versionFilePath); errors.Is(err, os.ErrNotExist) {
		context.LocalVersion = runtime.UnknownVersion
	} else if f.IsDir() {
		context.ShowFatal(fmt.Errorf("version file is a directory"))
	} else {
		data, err := os.ReadFile(versionFilePath)
		if err != nil {
			context.ShowFatal(err)
		}
		context.LocalVersion = strings.TrimSuffix(string(data), "\n")
	}
}

/*
func main() {
	a := app.New()
	w := a.NewWindow("ShaderCompat")
	w.SetFixedSize(true)
	w.Resize(fyne.NewSize(400, 300))

	context := &gui.Context{
		Window:       w,
		SelectedMods: make(map[string]bool),
	}

	view, err := gui.View(context)
	if err != nil {
		gui.ShowError(context, err)
		return
	}
	w.SetContent(view)
	w.ShowAndRun()
}

*/
