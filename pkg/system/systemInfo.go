package system

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"

	"slices"

	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/mem"
)

type SystemInfo struct {
	OS             string `json:"os"`
	Architecture   string `json:"architecture"`
	GoVersion      string `json:"go_version"`
	ExecutablePath string `json:"executorPath"`
	ExecutableDir  string `json:"executableDir"`
	WorkingDir     string `json:"workingDir"`

	CPU struct {
		Type        string  `json:"type"`
		Cores       int     `json:"cores"`
		Thread      int     `json:"thread"`
		UsedPercent float64 `json:"percent"`
	} `json:"cpu"`

	Memory struct {
		Total       float64 `json:"total"`
		Used        float64 `json:"used"`
		UsedPercent float64 `json:"used_percent"`
	} `json:"memory"`
}

func GetSystemInfo() (*SystemInfo, error) {

	// allocate memory in stack
	systemInfo := new(SystemInfo)

	systemInfo.OS = runtime.GOOS
	systemInfo.Architecture = runtime.GOARCH
	systemInfo.GoVersion = runtime.Version()
	// get executable path
	exePath, err := os.Executable()
	if err != nil {
		fmt.Println("get executable path error:", err)
	}
	systemInfo.ExecutablePath = exePath
	systemInfo.ExecutableDir = filepath.Dir(exePath)
	// get working dir
	workingDir, err := os.Getwd()
	if err != nil {
		fmt.Println("get working dir error:", err)
	}
	systemInfo.WorkingDir = workingDir

	// CPU

	cpuInfo, err := cpu.Info()
	if err != nil {
		systemInfo.CPU.Type = fmt.Sprintf("Get cpu type error: %v", err)
	} else {
		systemInfo.CPU.Type = strings.TrimSpace(cpuInfo[0].ModelName)
	}

	cores, err := cpu.Counts(true)
	if err != nil {
		systemInfo.CPU.Cores = 0
	} else {
		systemInfo.CPU.Cores = cores
	}

	thread, err := cpu.Counts(false)
	if err != nil {
		systemInfo.CPU.Thread = 0
	} else {
		systemInfo.CPU.Thread = thread
	}

	percent, err := cpu.Percent(500*time.Millisecond, false)
	if err != nil {
		systemInfo.CPU.UsedPercent = 0
	} else {
		systemInfo.CPU.UsedPercent, _ = strconv.ParseFloat(fmt.Sprintf("%.2f", percent[0]), 64)
	}

	// Memory

	vmStat, err := mem.VirtualMemory()
	if err != nil {
		systemInfo.Memory.Total = 0
		systemInfo.Memory.Used = 0
		systemInfo.Memory.UsedPercent = 0
	} else {
		// unit: GB

		systemInfo.Memory.Total, _ = strconv.ParseFloat(fmt.Sprintf("%.2f", float64(vmStat.Total)/1024/1024/1024), 64)
		systemInfo.Memory.Used, _ = strconv.ParseFloat(fmt.Sprintf("%.2f", float64(vmStat.Used)/1024/1024/1024), 64)
		systemInfo.Memory.UsedPercent, _ = strconv.ParseFloat(fmt.Sprintf("%.2f", systemInfo.Memory.Used/systemInfo.Memory.Total*100), 64)
	}

	return systemInfo, nil

}

func IsUnixSystem() bool {
	UnixSystems := []string{"linux", "darwin", "freebsd", "netbsd", "openbsd"}
	return slices.Contains(UnixSystems, runtime.GOOS)
}

func IsWindowsSystem() bool {
	return runtime.GOOS == "windows"
}
