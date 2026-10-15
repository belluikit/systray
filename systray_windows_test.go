//go:build windows

package systray

import (
	"io/ioutil"
	"runtime"
	"sync/atomic"
	"testing"
	"time"
	"unsafe"

	"golang.org/x/sys/windows"
)

const iconFilePath = "example/icon/iconwin.ico"

func TestBaseWindowsTray(t *testing.T) {
	systrayReady = func() {}
	exitMu.Lock()
	systrayExit = func() {}
	systrayExitCalled = false
	exitMu.Unlock()

	runtime.LockOSThread()

	if err := wt.initInstance(); err != nil {
		t.Fatalf("initInstance failed: %s", err)
	}

	if err := wt.createMenu(); err != nil {
		t.Fatalf("createMenu failed: %s", err)
	}
	// The same final step as registerSystray, which this test does by hand.
	wt.initialized.Store(true)

	defer func() {
		pDestroyWindow.Call(uintptr(wt.window))
		wt.wcex.unregister()
	}()

	if err := wt.setIcon(iconFilePath); err != nil {
		t.Errorf("SetIcon failed: %s", err)
	}

	if err := wt.setTooltip("Cyrillic tooltip тест:)"); err != nil {
		t.Errorf("SetIcon failed: %s", err)
	}

	var id atomic.Uint32
	err := wt.addOrUpdateMenuItem(id.Add(1), 0, "Simple enabled", false, false)
	if err != nil {
		t.Errorf("mergeMenuItem failed: %s", err)
	}
	err = wt.addOrUpdateMenuItem(id.Add(1), 0, "Simple disabled", true, false)
	if err != nil {
		t.Errorf("mergeMenuItem failed: %s", err)
	}
	err = wt.addSeparatorMenuItem(id.Add(1), 0)
	if err != nil {
		t.Errorf("addSeparatorMenuItem failed: %s", err)
	}
	err = wt.addOrUpdateMenuItem(id.Add(1), 0, "Simple checked enabled", false, true)
	if err != nil {
		t.Errorf("mergeMenuItem failed: %s", err)
	}
	err = wt.addOrUpdateMenuItem(id.Add(1), 0, "Simple checked disabled", true, true)
	if err != nil {
		t.Errorf("mergeMenuItem failed: %s", err)
	}

	err = wt.hideMenuItem(1, 0)
	if err != nil {
		t.Errorf("hideMenuItem failed: %s", err)
	}

	// Hiding an item that does not exist must not disturb the menu. Whether
	// it reports an error depends on whether RemoveMenu sets a last-error
	// code, which differs between Windows versions (Windows Server 2025 does
	// not), and the public Hide only logs it, so the result is not asserted.
	_ = wt.hideMenuItem(100, 0)

	err = wt.addOrUpdateMenuItem(2, 0, "Simple disabled update", true, false)
	if err != nil {
		t.Errorf("mergeMenuItem failed: %s", err)
	}

	time.AfterFunc(1*time.Second, quit)

	m := struct {
		WindowHandle windows.Handle
		Message      uint32
		Wparam       uintptr
		Lparam       uintptr
		Time         uint32
		Pt           point
	}{}
	for {
		ret, _, err := pGetMessage.Call(uintptr(unsafe.Pointer(&m)), 0, 0, 0)
		res := int32(ret)
		if res == -1 {
			t.Errorf("win32 GetMessage failed: %v", err)
			return
		} else if res == 0 {
			break
		}
		pTranslateMessage.Call(uintptr(unsafe.Pointer(&m)))
		pDispatchMessage.Call(uintptr(unsafe.Pointer(&m)))
	}
}

func TestWindowsRun(t *testing.T) {
	onReady := func() {
		b, err := ioutil.ReadFile(iconFilePath)
		if err != nil {
			t.Fatalf("Can't load icon file: %v", err)
		}
		SetIcon(b)
		SetTitle("Test title с кириллицей")

		bSomeBtn := AddMenuItem("Йа кнопко", "")
		bSomeBtn.Check()
		AddSeparator()
		bQuit := AddMenuItem("Quit", "Quit the whole app")
		go func() {
			<-bQuit.ClickedCh
			t.Log("Quit reqested")
			Quit()
		}()
		time.AfterFunc(1*time.Second, Quit)
	}

	onExit := func() {
		t.Log("Exit success")
	}

	Run(onReady, onExit)
}
