package runtime

import (
	"errors"
	"fmt"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
	"log"
	"os"
	"path"
	"shadercompat/repository"
	"sort"
	"strings"
)

type View interface {
	Build(ctx *Context) fyne.CanvasObject
	Update(ctx *Context) error
}

const (
	versionActionUpdate  = "Update"
	versionActionCheck   = "Check"
	versionActionInstall = "Install"
)

type versionView struct {
	initialized                  bool
	versionLabel, remoteVerLabel *widget.Label
	actionButton                 *widget.Button
	actionLocked                 bool
}

func (v *versionView) Build(ctx *Context) fyne.CanvasObject {
	defer func() { v.initialized = true }()

	v.versionLabel = unsetLabel()
	v.remoteVerLabel = unsetLabel()
	v.actionButton = unsetButton()

	v.actionButton.OnTapped = func() {
		if v.actionLocked {
			return
		}
		v.actionLocked = true
		v.actionButton.Disable()
		defer func() {
			v.actionLocked = false
			v.actionButton.Enable()
		}()

		switch v.actionButton.Text {
		case versionActionInstall:
			log.Print("Cloning repository... ")

			errChan := make(chan error)
			d := dialog.NewCustomWithoutButtons("Cloning repository", widget.NewLabel("This may take a while..."), ctx.Window)
			d.Show()
			go func() {
				err := repository.Clone(ctx.LocalPath)
				if err != nil {
					errChan <- err
				} else {
					errChan <- nil
				}
			}()
			err := <-errChan
			d.Hide()
			if err != nil {
				ctx.ShowFatal(err)
				return
			}

			ctx.LocalVersion = ctx.RemoteVersion

			if err := ctx.LoadMappings(); err != nil {
				ctx.ShowFatal(err)
				return
			}

			if err := v.Update(ctx); err != nil {
				ctx.ShowError(fmt.Errorf("failed to update version view: %v", err))
			}

			// Need to rebuild mods view to show new mods
			ctx.ModsView.Build(ctx)
			if err := ctx.ModsView.Update(ctx); err != nil {
				ctx.ShowError(fmt.Errorf("failed to mods view: %v", err))
			}
		case versionActionUpdate:
			if err := repository.Update(ctx.LocalPath); err != nil {
				ctx.ShowError(fmt.Errorf("failed to update repository: %v", err))
				return
			} else {
				ctx.LocalVersion = ctx.RemoteVersion
			}
			if err := v.Update(ctx); err != nil {
				ctx.ShowError(fmt.Errorf("failed to update version view: %v", err))
			}
		case versionActionCheck:
			version, err := repository.GetVersion()
			if err != nil {
				ctx.ShowError(fmt.Errorf("failed to get repository version: %v", err))
				return
			}

			if ctx.LocalVersion == version {
				dialog.ShowCustom("Up to date", "Ok", widget.NewLabel("The repository is up to date"), ctx.Window)
			} else {
				dialog.ShowCustom("Out of date", "Ok", widget.NewLabel("The repository is out of date"), ctx.Window)
			}

			ctx.RemoteVersion = version
			if err := v.Update(ctx); err != nil {
				ctx.ShowError(fmt.Errorf("failed to update version view: %v", err))
			}
		default:
			ctx.ShowError(fmt.Errorf("unknown action: %s", v.actionButton.Text))
		}
	}

	return container.NewHBox(
		v.actionButton,
		container.NewHBox(widget.NewLabel("Local:\nRemote:"), layout.NewSpacer(), v.versionLabel),
	)
}

func (v *versionView) Update(ctx *Context) error {
	if !v.initialized {
		return fmt.Errorf("versionView not initialized")
	}

	remoteVer := ctx.RemoteVersion
	if remoteVer == UnknownVersion {
		remoteVer = "unknown"
	}

	v.versionLabel.SetText(ctx.LocalVersion + "\n" + remoteVer)

	if ctx.LocalVersion == UnknownVersion {
		v.actionButton.SetText(versionActionInstall)
	} else if ctx.RemoteVersion == ctx.LocalVersion || ctx.RemoteVersion == UnknownVersion {
		v.actionButton.SetText(versionActionCheck)
	} else {
		v.actionButton.SetText(versionActionUpdate)
	}

	return nil
}

func NewVersionView() View {
	return &versionView{}
}

type shaderView struct {
	initialized bool
	fileName    *widget.Label
	list        *widget.List
}

func (v *shaderView) Build(ctx *Context) fyne.CanvasObject {
	defer func() { v.initialized = true }()

	v.list = widget.NewList(
		func() int {
			return len(ctx.Shaders)
		},
		func() fyne.CanvasObject {
			return widget.NewLabel("")
		},
		func(i widget.ListItemID, o fyne.CanvasObject) {
			var shaderNames []string
			for k := range ctx.Shaders {
				shaderNames = append(shaderNames, k)
			}
			sort.Slice(shaderNames, func(i, j int) bool {
				return shaderNames[i] < shaderNames[j]
			})
			o.(*widget.Label).SetText(shaderNames[i])
		},
	)
	v.list.OnSelected = func(i widget.ListItemID) {
		var shaderNames []string
		for k := range ctx.Shaders {
			shaderNames = append(shaderNames, k)
		}
		sort.Slice(shaderNames, func(i, j int) bool {
			return shaderNames[i] < shaderNames[j]
		})
		ctx.SelectedShaderName = shaderNames[i]
	}
	v.list.Select(0)

	scroll := container.NewScroll(v.list)
	scroll.SetMinSize(fyne.NewSize(300, 200))

	filePicker := dialog.NewFileOpen(func(reader fyne.URIReadCloser, err error) {
		if err != nil {
			ctx.ShowError(err)
			return
		}
		if reader == nil {
			dialog.ShowCustom("No file selected", "Ok", container.NewHBox(), ctx.Window)
			return
		}
		defer func() {
			if err := reader.Close(); err != nil {
				ctx.ShowError(err)
			}
		}()
		ctx.SelectedFilePath = strings.TrimPrefix(reader.URI().String(), "file://")
		if err := v.Update(ctx); err != nil {
			ctx.ShowError(fmt.Errorf("failed to update shader view: %v", err))
		}
	}, ctx.Window)

	v.fileName = unsetLabel()
	v.fileName.SetText("no file selected")
	browseButton := widget.NewButton("Browse", func() {
		filePicker.Show()
	})

	return container.NewVBox(
		container.NewHBox(
			browseButton,
			v.fileName,
		),
		widget.NewLabel("Profile"),
		scroll,
	)
}

func (v *shaderView) Update(ctx *Context) error {
	if !v.initialized {
		return fmt.Errorf("shaderView not initialized")
	}

	v.list.Refresh()
	v.fileName.SetText(path.Base(ctx.SelectedFilePath))

	return nil
}

func NewShaderView() View {
	return &shaderView{}
}

type modsView struct {
	initialized bool
	radios      map[string]*widget.Check
}

func (m *modsView) Build(ctx *Context) fyne.CanvasObject {
	defer func() { m.initialized = true }()

	m.radios = make(map[string]*widget.Check)

	sortedMods := make([]string, 0)
	for k := range ctx.ModStates {
		sortedMods = append(sortedMods, k)
	}

	sort.Slice(sortedMods, func(i, j int) bool {
		return sortedMods[i] < sortedMods[j]
	})

	modRadios := make([]fyne.CanvasObject, 0)

	for _, k := range sortedMods {
		state := ctx.ModStates[k]
		name := strings.TrimSuffix(path.Base(k), path.Ext(k))
		radio := widget.NewCheck(" "+name, func(b bool) {
			ctx.ModStates[k] = b
			if err := ctx.PatchView.Update(ctx); err != nil {
				ctx.ShowError(err)
			}
		})
		radio.SetChecked(state)
		m.radios[k] = radio
		modRadios = append(modRadios, radio)
	}

	scroll := container.NewScroll(container.NewVBox(modRadios...))
	scroll.SetMinSize(fyne.NewSize(300, 200))
	return scroll
}

func (m *modsView) Update(ctx *Context) error {
	if !m.initialized {
		return fmt.Errorf("modsView not initialized")
	}

	return nil
}

func NewModsView() View {
	return &modsView{}
}

type patchView struct {
	initialized bool
	infoLabel   *widget.Label
	patchButton *widget.Button
}

func (v *patchView) Build(ctx *Context) fyne.CanvasObject {
	v.infoLabel = unsetLabel()

	saveDialog := dialog.NewFileSave(func(uw fyne.URIWriteCloser, err error) {
		if err != nil {
			ctx.ShowError(err)
			return
		}
		if uw == nil {
			dialog.ShowError(fmt.Errorf("no output file selected"), ctx.Window)
			return
		}
		if err := uw.Close(); err != nil {
			ctx.ShowError(err)
			return
		}

		filePath := strings.TrimPrefix(uw.URI().String(), "file://")
		if err := os.Remove(filePath); err != nil && errors.Is(err, os.ErrNotExist) {
			ctx.ShowError(err)
			return
		}

		if path.Ext(strings.ToLower(filePath)) != ".zip" {
			filePath += ".zip"
		}

		ctx.OutputFilePath = filePath

		errChan := make(chan error)
		d := dialog.NewCustomWithoutButtons("Patching shader", widget.NewLabel("This may take a while..."), ctx.Window)
		d.Show()
		go func() {
			err := ctx.Patch(filePath)
			if err != nil {
				errChan <- err
			} else {
				errChan <- nil
			}
		}()
		e := <-errChan
		d.Hide()
		if e != nil {
			ctx.ShowFatal(e)
			return
		}

		dialog.ShowInformation("Patch successful", fmt.Sprintf("Saved to %s", filePath), ctx.Window)
	}, ctx.Window)

	v.patchButton = widget.NewButton("Patch shader", func() {
		if len(ctx.SelectedFilePath) == 0 {
			dialog.ShowError(fmt.Errorf("no shader source file selected"), ctx.Window)
			return
		}
		saveDialog.Show()
	})

	defer func() {
		v.initialized = true

		if err := v.Update(ctx); err != nil {
			ctx.ShowError(err)
		}
	}()

	return container.NewVBox(
		container.NewHBox(
			v.infoLabel,
			layout.NewSpacer(),
			v.patchButton,
		),
	)
}

func (v *patchView) Update(ctx *Context) error {
	if !v.initialized {
		return fmt.Errorf("patchView not initialized")
	}

	modsSelected := 0
	for _, v := range ctx.ModStates {
		if v {
			modsSelected++
		}
	}

	shaderName := ctx.SelectedShaderName

	text := fmt.Sprintf("%d mods, profile %s", modsSelected, shaderName)
	v.infoLabel.SetText(text)

	return nil
}

func NewPatchView() View {
	return &patchView{}
}
