package runtime

import "fyne.io/fyne/v2/widget"

func unsetLabel() *widget.Label {
	return widget.NewLabel("...")
}

func unsetButton() *widget.Button {
	return widget.NewButton("...", nil)
}

func Header(text string) *widget.Label {
	label := widget.NewLabel(text)
	label.TextStyle.Bold = true
	return label
}
