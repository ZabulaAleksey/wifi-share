//go:build windows

package main

import (
	"bytes"
	"context"
	"encoding/binary"
	"image"
	"image/color"
	"image/png"
	"log"
	"os/exec"
	"strings"
	"time"

	"github.com/getlantern/systray"
	"github.com/local/wifi-share/internal/app"
)

func runApplication(server *app.App, config localConfig, urls []string) {
	go func() {
		if err := server.ListenAndServe(); err != nil {
			log.Printf("server stopped: %v", err)
			systray.Quit()
		}
	}()
	systray.Run(func() {
		systray.SetIcon(trayIcon())
		systray.SetTooltip("WiFi Share")
		openItem := systray.AddMenuItem("Открыть", "")
		urlItem := systray.AddMenuItem("Показать URL", "")
		folderItem := systray.AddMenuItem("Выбрать папку", "")
		settingsItem := systray.AddMenuItem("Настройки", "")
		systray.AddSeparator()
		exitItem := systray.AddMenuItem("Завершить", "")
		go func() {
			for {
				select {
				case <-openItem.ClickedCh:
					openURL(urls[0])
				case <-urlItem.ClickedCh:
					showMessage("WiFi Share", strings.Join(urls, "\r\n"))
				case <-folderItem.ClickedCh:
					if folder := chooseFolder(); folder != "" {
						if err := server.SetShareDir(folder); err != nil {
							showMessage("Ошибка", err.Error())
							continue
						}
						config.Root = folder
						if err := saveLocalConfig("config.local.json", config); err != nil {
							showMessage("Ошибка", err.Error())
						}
					}
				case <-settingsItem.ClickedCh:
					_ = exec.Command("notepad.exe", "config.local.json").Start()
				case <-exitItem.ClickedCh:
					ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
					_ = server.Shutdown(ctx)
					cancel()
					systray.Quit()
					return
				}
			}
		}()
	}, func() {})
}

func openURL(url string) {
	_ = exec.Command("rundll32.exe", "url.dll,FileProtocolHandler", url).Start()
}

func chooseFolder() string {
	script := `Add-Type -AssemblyName System.Windows.Forms; $d=New-Object System.Windows.Forms.FolderBrowserDialog; $d.Description='Выберите общую папку WiFi Share'; if($d.ShowDialog() -eq 'OK'){[Console]::Write($d.SelectedPath)}`
	output, err := exec.Command("powershell.exe", "-NoProfile", "-STA", "-Command", script).Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(output))
}

func showMessage(title, message string) {
	title = strings.ReplaceAll(title, "'", "''")
	message = strings.ReplaceAll(message, "'", "''")
	script := "Add-Type -AssemblyName System.Windows.Forms; [System.Windows.Forms.MessageBox]::Show('" + message + "','" + title + "')"
	_ = exec.Command("powershell.exe", "-NoProfile", "-STA", "-Command", script).Start()
}

func trayIcon() []byte {
	img := image.NewRGBA(image.Rect(0, 0, 32, 32))
	for y := 0; y < 32; y++ {
		for x := 0; x < 32; x++ {
			dx, dy := x-16, y-16
			if dx*dx+dy*dy <= 13*13 {
				img.Set(x, y, color.RGBA{128, 229, 188, 255})
			}
		}
	}
	var data bytes.Buffer
	_ = png.Encode(&data, img)
	var ico bytes.Buffer
	_ = binary.Write(&ico, binary.LittleEndian, []uint16{0, 1, 1})
	ico.Write([]byte{32, 32, 0, 0})
	_ = binary.Write(&ico, binary.LittleEndian, uint16(1))
	_ = binary.Write(&ico, binary.LittleEndian, uint16(32))
	_ = binary.Write(&ico, binary.LittleEndian, uint32(data.Len()))
	_ = binary.Write(&ico, binary.LittleEndian, uint32(22))
	ico.Write(data.Bytes())
	return ico.Bytes()
}
