package runtime

import (
	"fmt"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
	"log"
	"path"
	"shadercompat/repository"
	"sort"
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
	defer func() {
		v.initialized = true
	}()

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
			}

			ctx.LocalVersion = ctx.RemoteVersion

			if err := v.Update(ctx); err != nil {
				ctx.ShowError(fmt.Errorf("failed to update version view: %v", err))
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
		container.NewHBox(widget.NewLabel("Local:\nRemote:"), layout.NewSpacer(), v.versionLabel),
		v.actionButton,
	)
}

func (v *versionView) Update(ctx *Context) error {
	if !v.initialized {
		return fmt.Errorf("versionView not initialized")
	}

	v.versionLabel.SetText(ctx.LocalVersion + "\n" + ctx.RemoteVersion)

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
	v.initialized = true

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

	scroll := container.NewVScroll(v.list)
	scroll.SetMinSize(fyne.NewSize(200, 200))

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
		ctx.SelectedFilePath = reader.URI().String()
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
