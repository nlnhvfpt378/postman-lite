package ui

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
	appcore "postman-lite/internal/app"
	"postman-lite/internal/model"
	"postman-lite/internal/state"
)

type DesktopApp struct {
	core *appcore.App
}

func New(core *appcore.App) *DesktopApp {
	return &DesktopApp{core: core}
}

func (d *DesktopApp) Run() error {
	stored, err := state.Load()
	if err != nil {
		stored = state.Default()
	}

	fy := app.NewWithID("postman-lite")
	win := fy.NewWindow("Postman Lite")
	win.Resize(fyne.NewSize(1180, 760))

	methods := []string{"GET", "POST", "PUT", "DELETE", "PATCH", "HEAD", "OPTIONS"}
	methodSelect := widget.NewSelect(methods, nil)
	methodSelect.SetSelected(stored.Method)
	if methodSelect.Selected == "" {
		methodSelect.SetSelected("GET")
	}

	urlEntry := widget.NewEntry()
	urlEntry.SetText(stored.URL)
	urlEntry.SetPlaceHolder("https://httpbin.org/anything")

	headersEntry := widget.NewMultiLineEntry()
	headersEntry.SetText(stored.Headers)
	headersEntry.SetPlaceHolder("Content-Type: application/json")
	headersEntry.Wrapping = fyne.TextWrapWord

	bodyEntry := widget.NewMultiLineEntry()
	bodyEntry.SetText(stored.Body)
	bodyEntry.SetPlaceHolder("{\n  \"hello\": \"world\"\n}")
	bodyEntry.Wrapping = fyne.TextWrapWord

	statusValue := widget.NewLabel("未发送")
	timeValue := widget.NewLabel("-")
	sizeValue := widget.NewLabel("-")
	respHeaders := widget.NewMultiLineEntry()
	respHeaders.Disable()
	respHeaders.Wrapping = fyne.TextWrapWord
	respBody := widget.NewMultiLineEntry()
	respBody.Disable()
	respBody.Wrapping = fyne.TextWrapWord

	saveState := func() {
		_ = state.Save(state.FormState{
			Method:  methodSelect.Selected,
			URL:     urlEntry.Text,
			Headers: headersEntry.Text,
			Body:    bodyEntry.Text,
		})
	}

	methodSelect.OnChanged = func(string) { saveState() }
	urlEntry.OnChanged = func(string) { saveState() }
	headersEntry.OnChanged = func(string) { saveState() }
	bodyEntry.OnChanged = func(string) { saveState() }

	sendBtn := widget.NewButton("发送请求", nil)
	formatBtn := widget.NewButton("JSON 美化", func() {
		pretty, err := PrettyJSON(bodyEntry.Text)
		if err != nil {
			dialog.ShowError(err, win)
			return
		}
		bodyEntry.SetText(pretty)
		saveState()
	})
	copyBtn := widget.NewButton("复制响应体", func() {
		win.Clipboard().SetContent(respBody.Text)
	})
	stateBtn := widget.NewButton("查看状态文件", func() {
		path, err := state.Path()
		if err != nil {
			dialog.ShowError(err, win)
			return
		}
		dialog.ShowInformation("状态文件", path, win)
	})

	sendBtn.OnTapped = func() {
		saveState()
		sendBtn.Disable()
		statusValue.SetText("请求中...")
		timeValue.SetText("-")
		sizeValue.SetText("-")
		respHeaders.SetText("")
		respBody.SetText("")

		ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
		defer cancel()
		resp := d.core.Client.Send(ctx, model.Request{
			Method:  methodSelect.Selected,
			URL:     strings.TrimSpace(urlEntry.Text),
			Headers: ParseHeaders(headersEntry.Text),
			Body:    bodyEntry.Text,
		})
		defer sendBtn.Enable()
		if resp.Error != "" {
			statusValue.SetText("失败")
			respBody.SetText(resp.Error)
			return
		}
		statusValue.SetText(resp.Status)
		timeValue.SetText(resp.DurationText())
		sizeValue.SetText(fmt.Sprintf("%d bytes", resp.Size))
		respHeaders.SetText(FormatHeaders(resp.Headers))
		respBody.SetText(MaybePrettyBody(resp.Body, resp.Headers))
	}

	requestPane := container.NewBorder(
		container.NewVBox(
			widget.NewLabelWithStyle("请求", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
			container.NewBorder(nil, nil, methodSelect, nil, urlEntry),
			container.NewHBox(sendBtn, formatBtn, copyBtn, stateBtn),
		),
		nil,
		nil,
		nil,
		container.NewVSplit(
			container.NewBorder(widget.NewLabel("Headers"), nil, nil, nil, headersEntry),
			container.NewBorder(widget.NewLabel("Body"), nil, nil, nil, bodyEntry),
		),
	)

	meta := container.NewGridWithColumns(3,
		container.NewVBox(widget.NewLabel("状态"), statusValue),
		container.NewVBox(widget.NewLabel("耗时"), timeValue),
		container.NewVBox(widget.NewLabel("大小"), sizeValue),
	)

	responsePane := container.NewBorder(
		container.NewVBox(
			widget.NewLabelWithStyle("响应", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
			meta,
		),
		nil,
		nil,
		nil,
		container.NewVSplit(
			container.NewBorder(widget.NewLabel("响应头"), nil, nil, nil, respHeaders),
			container.NewBorder(widget.NewLabel("响应体"), nil, nil, nil, respBody),
		),
	)

	split := container.NewHSplit(requestPane, responsePane)
	split.SetOffset(0.46)
	content := container.NewBorder(nil, nil, nil, nil, split)
	win.SetContent(container.New(layout.NewMaxLayout(), content))
	win.ShowAndRun()
	return nil
}

func ParseHeaders(input string) []model.HeaderKV {
	lines := strings.Split(input, "\n")
	out := make([]model.HeaderKV, 0, len(lines))
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		parts := strings.SplitN(line, ":", 2)
		if len(parts) != 2 {
			continue
		}
		out = append(out, model.HeaderKV{Key: strings.TrimSpace(parts[0]), Value: strings.TrimSpace(parts[1])})
	}
	return out
}

func FormatHeaders(headers map[string][]string) string {
	if len(headers) == 0 {
		return ""
	}
	keys := make([]string, 0, len(headers))
	for k := range headers {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	lines := make([]string, 0, len(keys))
	for _, k := range keys {
		lines = append(lines, fmt.Sprintf("%s: %s", k, strings.Join(headers[k], ", ")))
	}
	return strings.Join(lines, "\n")
}

func MaybePrettyBody(body string, headers map[string][]string) string {
	trimmed := strings.TrimSpace(body)
	if trimmed == "" {
		return ""
	}
	contentType := strings.ToLower(strings.Join(headers["Content-Type"], ","))
	if strings.Contains(contentType, "json") || strings.HasPrefix(trimmed, "{") || strings.HasPrefix(trimmed, "[") {
		pretty, err := PrettyJSON(body)
		if err == nil {
			return pretty
		}
	}
	return body
}
