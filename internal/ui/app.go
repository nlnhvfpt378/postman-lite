package ui

import (
	"context"
	"fmt"
	"net/url"
	"sort"
	"strings"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
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

type tabDescriptor struct {
	state state.RequestTabState
	item  *tabButton
}

type tabButton struct {
	widget.BaseWidget
	label       *widget.Label
	closeBtn    *widget.Button
	title       string
	active      bool
	onTap       func()
	onMiddleTap func()
	onSecondary func(fyne.Position)
	onClose     func()
}

func newTabButton(title string) *tabButton {
	t := &tabButton{
		label:    widget.NewLabel(title),
		closeBtn: widget.NewButtonWithIcon("", theme.CancelIcon(), nil),
		title:    title,
	}
	t.closeBtn.Importance = widget.LowImportance
	t.closeBtn.OnTapped = func() {
		if t.onClose != nil {
			t.onClose()
		}
	}
	t.label.Truncation = fyne.TextTruncateEllipsis
	t.ExtendBaseWidget(t)
	return t
}

func (t *tabButton) SetTitle(title string) {
	t.title = title
	t.label.SetText(title)
	t.Refresh()
}

func (t *tabButton) SetActive(active bool) {
	t.active = active
	t.Refresh()
}

func (t *tabButton) CreateRenderer() fyne.WidgetRenderer {
	bg := canvasRect(theme.ColorNameInputBackground)
	line := canvasRect(theme.ColorNameSeparator)
	content := container.NewBorder(nil, nil, nil, t.closeBtn, t.label)
	objects := []fyne.CanvasObject{bg, line, content}
	return &tabButtonRenderer{tab: t, bg: bg, line: line, content: content, objects: objects}
}

func (t *tabButton) Tapped(*fyne.PointEvent) {
	if t.onTap != nil {
		t.onTap()
	}
}

func (t *tabButton) TappedSecondary(ev *fyne.PointEvent) {
	if t.onSecondary != nil {
		t.onSecondary(ev.AbsolutePosition)
	}
}

func (t *tabButton) MouseDown(ev *desktop.MouseEvent) {
	if ev.Button == desktop.MouseButtonTertiary && t.onMiddleTap != nil {
		t.onMiddleTap()
	}
}

func (t *tabButton) MouseUp(*desktop.MouseEvent) {}

func (t *tabButton) MinSize() fyne.Size {
	labelSize := t.label.MinSize()
	closeSize := t.closeBtn.MinSize()
	return fyne.NewSize(max(labelSize.Width+closeSize.Width+24, 128), max(labelSize.Height, closeSize.Height)+14)
}

type tabButtonRenderer struct {
	tab     *tabButton
	bg      *canvas.Rectangle
	line    *canvas.Rectangle
	content *fyne.Container
	objects []fyne.CanvasObject
}

func (r *tabButtonRenderer) Layout(size fyne.Size) {
	r.bg.Resize(size)
	r.line.Resize(fyne.NewSize(size.Width, 1))
	r.line.Move(fyne.NewPos(0, size.Height-1))
	r.content.Move(fyne.NewPos(10, 7))
	r.content.Resize(fyne.NewSize(size.Width-20, size.Height-14))
}

func (r *tabButtonRenderer) MinSize() fyne.Size { return r.tab.MinSize() }
func (r *tabButtonRenderer) Refresh() {
	variant := fyne.CurrentApp().Settings().ThemeVariant()
	th := fyne.CurrentApp().Settings().Theme()
	bgColor := th.Color(theme.ColorNameInputBackground, variant)
	if r.tab.active {
		bgColor = th.Color(theme.ColorNameOverlayBackground, variant)
	}
	r.bg.FillColor = bgColor
	r.bg.Refresh()
	r.line.FillColor = th.Color(theme.ColorNameSeparator, variant)
	r.line.Refresh()
	r.tab.label.TextStyle = fyne.TextStyle{Bold: r.tab.active}
	r.tab.label.Refresh()
	r.content.Refresh()
}
func (r *tabButtonRenderer) Objects() []fyne.CanvasObject { return r.objects }
func (r *tabButtonRenderer) Destroy()                     {}

func canvasRect(name fyne.ThemeColorName) *canvas.Rectangle {
	return canvas.NewRectangle(fyne.CurrentApp().Settings().Theme().Color(name, fyne.CurrentApp().Settings().ThemeVariant()))
}

func (d *DesktopApp) Run() error {
	stored, err := state.Load()
	if err != nil {
		stored = state.Default()
	}

	fy := app.NewWithID("postman-lite")
	fy.SetIcon(AppIcon())
	win := fy.NewWindow("Postman Lite")
	win.SetIcon(AppIcon())
	win.Resize(fyne.NewSize(1220, 780))

	methods := []string{"GET", "POST", "PUT", "DELETE", "PATCH", "HEAD", "OPTIONS"}
	methodSelect := widget.NewSelect(methods, nil)
	urlEntry := widget.NewEntry()
	urlEntry.SetPlaceHolder("https://httpbin.org/anything")

	headersEntry := widget.NewMultiLineEntry()
	headersEntry.SetPlaceHolder("Content-Type: application/json")
	headersEntry.Wrapping = fyne.TextWrapWord

	bodyEntry := widget.NewMultiLineEntry()
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

	var (
		tabs            []tabDescriptor
		selectedIndex   int
		ignoreFieldSave bool
		tabBar          *fyne.Container
		tabScroll       *container.Scroll
	)

	saveState := func() {
		if len(tabs) == 0 {
			return
		}
		st := state.FormState{
			Tabs:         make([]state.RequestTabState, len(tabs)),
			SelectedTab:  selectedIndex,
			NextTabIndex: stored.NextTabIndex,
		}
		for i := range tabs {
			st.Tabs[i] = tabs[i].state
		}
		if selectedIndex >= 0 && selectedIndex < len(st.Tabs) {
			current := st.Tabs[selectedIndex]
			st.Method = current.Method
			st.URL = current.URL
			st.Headers = current.Headers
			st.Body = current.Body
		}
		_ = state.Save(st)
	}

	updateCurrentFromFields := func() {
		if ignoreFieldSave || selectedIndex < 0 || selectedIndex >= len(tabs) {
			return
		}
		current := &tabs[selectedIndex].state
		current.Method = methodSelect.Selected
		current.URL = urlEntry.Text
		current.Headers = headersEntry.Text
		current.Body = bodyEntry.Text
		trimmedURL := strings.TrimSpace(current.URL)
		if trimmedURL != "" {
			current.Title = buildTabTitle(current.Method, trimmedURL)
		} else if strings.TrimSpace(current.Title) == "" {
			current.Title = fmt.Sprintf("请求 %d", selectedIndex+1)
		}
		if tabs[selectedIndex].item != nil {
			tabs[selectedIndex].item.SetTitle(current.Title)
		}
		saveState()
	}

	selectTab := func(index int) {}
	refreshTabBar := func() {}
	closeTabAt := func(index int) {}
	closeTabsOtherThan := func(index int) {}
	closeTabsToRight := func(index int) {}
	importOpenAPI := func(raw string) {}
	createTab := func(tab state.RequestTabState, autoSelect bool) {}

	methodSelect.OnChanged = func(string) { updateCurrentFromFields() }
	urlEntry.OnChanged = func(string) { updateCurrentFromFields() }
	headersEntry.OnChanged = func(string) { updateCurrentFromFields() }
	bodyEntry.OnChanged = func(string) { updateCurrentFromFields() }

	var importDialogEntry *widget.Entry

	sendBtn := widget.NewButton("发送请求", nil)
	formatBtn := widget.NewButton("JSON 美化", func() {
		pretty, err := PrettyJSON(bodyEntry.Text)
		if err != nil {
			dialog.ShowError(err, win)
			return
		}
		bodyEntry.SetText(pretty)
		updateCurrentFromFields()
	})
	importBtn := widget.NewButton("导入 OpenAPI JSON", func() {
		importDialog := dialog.NewCustomConfirm("导入 OpenAPI JSON", "导入", "取消", container.NewBorder(
			widget.NewLabel("粘贴 OpenAPI 3 JSON 后导入，会按接口生成请求标签页。"), nil, nil, nil,
			func() *widget.Entry {
				entry := widget.NewMultiLineEntry()
				entry.Wrapping = fyne.TextWrapWord
				entry.SetPlaceHolder("{\n  \"openapi\": \"3.0.0\", ...\n}")
				entry.Resize(fyne.NewSize(720, 420))
				importDialogEntry = entry
				return entry
			}(),
		), func(ok bool) {
			if !ok {
				return
			}
			importOpenAPI(importDialogEntry.Text)
		}, win)
		importDialog.Resize(fyne.NewSize(760, 520))
		importDialog.Show()
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

	requestHint := widget.NewLabel("提示：支持导入 OpenAPI 3 JSON，请使用“导入 OpenAPI JSON”按钮粘贴导入。部分桌面环境下右键/中键行为受 Fyne 平台支持限制，已保留按钮与菜单兜底。")
	requestHint.Wrapping = fyne.TextWrapWord

	queueUI := func(fn func()) {
		if queued, ok := interface{}(win).(interface{ QueueEvent(func()) }); ok {
			queued.QueueEvent(fn)
			return
		}
		fn()
	}

	sendBtn.OnTapped = func() {
		updateCurrentFromFields()

		request := model.Request{
			Method:  methodSelect.Selected,
			URL:     strings.TrimSpace(urlEntry.Text),
			Headers: ParseHeaders(headersEntry.Text),
			Body:    bodyEntry.Text,
		}

		sendBtn.Disable()
		statusValue.SetText("请求中...")
		timeValue.SetText("-")
		sizeValue.SetText("-")
		respHeaders.SetText("")
		respBody.SetText("")

		go func(req model.Request) {
			ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
			defer cancel()

			resp := d.core.Client.Send(ctx, req)
			queueUI(func() {
				sendBtn.Enable()
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
			})
		}(request)
	}

	buildContextMenu := func(index int, pos fyne.Position) {
		menu := fyne.NewMenu("",
			fyne.NewMenuItem("关闭当前", func() { closeTabAt(index) }),
			fyne.NewMenuItem("关闭右侧", func() { closeTabsToRight(index) }),
			fyne.NewMenuItem("关闭其他", func() { closeTabsOtherThan(index) }),
		)
		if len(tabs) <= 1 {
			for _, item := range menu.Items {
				item.Disabled = true
			}
		}
		widget.ShowPopUpMenuAtPosition(menu, win.Canvas(), pos)
	}

	var moreTabsBtn *widget.Button
	showTabListMenu := func() {}

	refreshTabBar = func() {
		objects := make([]fyne.CanvasObject, 0, len(tabs))
		for i := range tabs {
			i := i
			item := tabs[i].item
			item.SetActive(i == selectedIndex)
			item.onTap = func() { selectTab(i) }
			item.onMiddleTap = func() { closeTabAt(i) }
			item.onSecondary = func(pos fyne.Position) { buildContextMenu(i, pos) }
			item.onClose = func() { closeTabAt(i) }
			objects = append(objects, item)
		}
		tabBar.Objects = objects
		tabBar.Refresh()
		if tabScroll != nil {
			tabScroll.Refresh()
		}
	}

	showTabListMenu = func() {
		if moreTabsBtn == nil || len(tabs) == 0 {
			return
		}
		items := make([]*fyne.MenuItem, 0, len(tabs))
		for i := range tabs {
			i := i
			prefix := ""
			if i == selectedIndex {
				prefix = "✓ "
			}
			title := strings.TrimSpace(tabs[i].state.Title)
			if title == "" {
				title = fmt.Sprintf("请求 %d", i+1)
			}
			items = append(items, fyne.NewMenuItem(prefix+title, func() { selectTab(i) }))
		}
		widget.ShowPopUpMenuAtRelativePosition(fyne.NewMenu("", items...), win.Canvas(), fyne.NewPos(0, 0), moreTabsBtn)
	}

	selectTab = func(index int) {
		if index < 0 || index >= len(tabs) {
			return
		}
		selectedIndex = index
		current := tabs[index].state
		ignoreFieldSave = true
		methodSelect.SetSelected(current.Method)
		if methodSelect.Selected == "" {
			methodSelect.SetSelected("GET")
		}
		urlEntry.SetText(current.URL)
		headersEntry.SetText(current.Headers)
		bodyEntry.SetText(current.Body)
		ignoreFieldSave = false
		refreshTabBar()
		saveState()
	}

	createTab = func(tab state.RequestTabState, autoSelect bool) {
		if strings.TrimSpace(tab.ID) == "" {
			tab.ID = fmt.Sprintf("tab-%d", stored.NextTabIndex)
		}
		if strings.TrimSpace(tab.Title) == "" {
			tab.Title = fmt.Sprintf("请求 %d", stored.NextTabIndex)
		}
		if strings.TrimSpace(tab.Method) == "" {
			tab.Method = "GET"
		}
		item := newTabButton(tab.Title)
		tabs = append(tabs, tabDescriptor{state: tab, item: item})
		stored.NextTabIndex++
		refreshTabBar()
		if autoSelect {
			selectTab(len(tabs) - 1)
		} else {
			saveState()
		}
	}

	closeTabAt = func(index int) {
		if len(tabs) <= 1 || index < 0 || index >= len(tabs) {
			return
		}
		tabs = append(tabs[:index], tabs[index+1:]...)
		if selectedIndex >= len(tabs) {
			selectedIndex = len(tabs) - 1
		} else if index < selectedIndex {
			selectedIndex--
		} else if index == selectedIndex && selectedIndex > 0 {
			selectedIndex--
		}
		refreshTabBar()
		selectTab(selectedIndex)
	}

	closeTabsOtherThan = func(index int) {
		if len(tabs) <= 1 || index < 0 || index >= len(tabs) {
			return
		}
		tabs = []tabDescriptor{tabs[index]}
		selectedIndex = 0
		refreshTabBar()
		selectTab(0)
	}

	closeTabsToRight = func(index int) {
		if len(tabs) <= 1 || index < 0 || index >= len(tabs)-1 {
			return
		}
		tabs = tabs[:index+1]
		if selectedIndex > index {
			selectedIndex = index
		}
		refreshTabBar()
		selectTab(selectedIndex)
	}

	importOpenAPI = func(raw string) {
		raw = strings.TrimSpace(raw)
		if raw == "" {
			dialog.ShowInformation("导入 OpenAPI", "请输入或粘贴 OpenAPI 3 JSON。", win)
			return
		}
		items, err := ParseOpenAPIJSON(raw, &stored.NextTabIndex)
		if err != nil {
			dialog.ShowError(err, win)
			return
		}
		for i, item := range items {
			createTab(item, i == len(items)-1)
		}
		dialog.ShowInformation("导入完成", fmt.Sprintf("已导入 %d 个接口标签页。", len(items)), win)
	}

	requestPane := container.NewBorder(
		container.NewVBox(
			widget.NewLabelWithStyle("请求", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
			requestHint,
			container.NewBorder(nil, nil, methodSelect, nil, urlEntry),
			container.NewHBox(sendBtn, formatBtn, importBtn, copyBtn, stateBtn),
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
	split.SetOffset(0.48)
	tabBar = container.NewHBox()
	tabScroll = container.NewHScroll(tabBar)
	newTabBtn := widget.NewButtonWithIcon("新建", theme.ContentAddIcon(), func() {
		createTab(state.RequestTabState{
			ID:      fmt.Sprintf("tab-%d", stored.NextTabIndex),
			Title:   fmt.Sprintf("请求 %d", stored.NextTabIndex),
			Method:  "GET",
			URL:     "",
			Headers: "",
			Body:    "",
		}, true)
	})
	newTabBtn.Importance = widget.LowImportance
	moreTabsBtn = widget.NewButtonWithIcon("更多", theme.MenuExpandIcon(), func() {
		showTabListMenu()
	})
	moreTabsBtn.Importance = widget.LowImportance
	tabHeader := container.NewBorder(nil, nil, nil, container.NewHBox(moreTabsBtn, newTabBtn), tabScroll)
	content := container.NewBorder(tabHeader, nil, nil, nil, split)
	win.SetContent(container.New(layout.NewMaxLayout(), content))

	if len(stored.Tabs) == 0 {
		stored = state.Default()
	}
	for _, tab := range stored.Tabs {
		createTab(tab, false)
	}
	if len(tabs) == 0 {
		createTab(state.Default().Tabs[0], false)
	}
	if stored.SelectedTab < 0 || stored.SelectedTab >= len(tabs) {
		stored.SelectedTab = 0
	}
	selectTab(stored.SelectedTab)

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

func buildTabTitle(method, rawURL string) string {
	method = strings.ToUpper(strings.TrimSpace(method))
	rawURL = strings.TrimSpace(rawURL)
	if rawURL == "" {
		return "新请求"
	}
	if parsed, err := url.Parse(rawURL); err == nil {
		path := parsed.Path
		if path == "" {
			path = "/"
		}
		if parsed.RawQuery != "" {
			path += "?" + parsed.RawQuery
		}
		return fmt.Sprintf("%s %s", method, path)
	}
	return fmt.Sprintf("%s %s", method, rawURL)
}

func max(a, b float32) float32 {
	if a > b {
		return a
	}
	return b
}
