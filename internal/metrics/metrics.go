package metrics

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"
)

type Display struct {
	Connected bool     `json:"connected"`
	Connector string   `json:"connector,omitempty"`
	Modes     []string `json:"modes,omitempty"`
}

type GPU struct {
	Vendor        string  `json:"vendor,omitempty"`
	Name          string  `json:"name,omitempty"`
	UsagePercent  float64 `json:"usage_percent,omitempty"`
	MemoryUsedMB  float64 `json:"memory_used_mb,omitempty"`
	MemoryTotalMB float64 `json:"memory_total_mb,omitempty"`
	TemperatureC  float64 `json:"temperature_c,omitempty"`
}

type Snapshot struct {
	Timestamp       time.Time `json:"timestamp"`
	Hostname        string    `json:"hostname"`
	OS              string    `json:"os"`
	Kernel          string    `json:"kernel"`
	CPUPercent      float64   `json:"cpu_percent"`
	Load1           float64   `json:"load_1"`
	MemoryUsedBytes uint64    `json:"memory_used_bytes"`
	MemoryTotal     uint64    `json:"memory_total_bytes"`
	MemoryPercent   float64   `json:"memory_percent"`
	DiskUsedBytes   uint64    `json:"disk_used_bytes"`
	DiskTotalBytes  uint64    `json:"disk_total_bytes"`
	DiskPercent     float64   `json:"disk_percent"`
	NetworkRXBps    float64   `json:"network_rx_bps"`
	NetworkTXBps    float64   `json:"network_tx_bps"`
	TemperatureC    float64   `json:"temperature_c,omitempty"`
	UptimeSeconds   float64   `json:"uptime_seconds"`
	IPAddresses     []string  `json:"ip_addresses"`
	Display         Display   `json:"display"`
	GPU             GPU       `json:"gpu"`
}

type cpuCounters struct {
	total uint64
	idle  uint64
}

type networkCounters struct {
	rx uint64
	tx uint64
}

type Collector struct {
	mu          sync.Mutex
	dataPath    string
	previousCPU cpuCounters
	previousNet networkCounters
	previousAt  time.Time
	staticGPU   GPU
	osName      string
	kernel      string
}

func NewCollector(dataPath string) *Collector {
	return &Collector{
		dataPath:  dataPath,
		staticGPU: detectGPU(),
		osName:    readOSName(),
		kernel:    readKernel(),
	}
}

func (c *Collector) Sample(ctx context.Context) Snapshot {
	c.mu.Lock()
	defer c.mu.Unlock()

	now := time.Now()
	hostname, _ := os.Hostname()
	snapshot := Snapshot{
		Timestamp:     now,
		Hostname:      hostname,
		OS:            c.osName,
		Kernel:        c.kernel,
		IPAddresses:   ipAddresses(),
		Display:       readDisplay(),
		GPU:           c.staticGPU,
		TemperatureC:  readTemperature(),
		UptimeSeconds: readUptime(),
	}

	currentCPU := readCPU()
	if c.previousCPU.total > 0 && currentCPU.total > c.previousCPU.total {
		totalDelta := currentCPU.total - c.previousCPU.total
		idleDelta := currentCPU.idle - c.previousCPU.idle
		snapshot.CPUPercent = round(100*(1-float64(idleDelta)/float64(totalDelta)), 1)
	}
	c.previousCPU = currentCPU

	snapshot.Load1 = readLoad1()
	snapshot.MemoryUsedBytes, snapshot.MemoryTotal = readMemory()
	if snapshot.MemoryTotal > 0 {
		snapshot.MemoryPercent = round(100*float64(snapshot.MemoryUsedBytes)/float64(snapshot.MemoryTotal), 1)
	}
	snapshot.DiskUsedBytes, snapshot.DiskTotalBytes = readDisk(c.dataPath)
	if snapshot.DiskTotalBytes > 0 {
		snapshot.DiskPercent = round(100*float64(snapshot.DiskUsedBytes)/float64(snapshot.DiskTotalBytes), 1)
	}

	currentNet := readNetwork()
	if !c.previousAt.IsZero() {
		seconds := now.Sub(c.previousAt).Seconds()
		if seconds > 0 {
			if currentNet.rx >= c.previousNet.rx {
				snapshot.NetworkRXBps = float64(currentNet.rx-c.previousNet.rx) / seconds
			}
			if currentNet.tx >= c.previousNet.tx {
				snapshot.NetworkTXBps = float64(currentNet.tx-c.previousNet.tx) / seconds
			}
		}
	}
	c.previousNet = currentNet
	c.previousAt = now

	if snapshot.GPU.Vendor == "NVIDIA" {
		snapshot.GPU = sampleNVIDIA(ctx, snapshot.GPU)
	}
	return snapshot
}

func WriteSnapshot(path string, snapshot Snapshot) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	data, err := json.Marshal(snapshot)
	if err != nil {
		return err
	}
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, append(data, '\n'), 0o644); err != nil {
		return err
	}
	return os.Rename(tmp, path)
}

func readCPU() cpuCounters {
	file, err := os.Open("/proc/stat")
	if err != nil {
		return cpuCounters{}
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)
	if !scanner.Scan() {
		return cpuCounters{}
	}
	fields := strings.Fields(scanner.Text())
	if len(fields) < 5 || fields[0] != "cpu" {
		return cpuCounters{}
	}
	var values []uint64
	for _, field := range fields[1:] {
		value, _ := strconv.ParseUint(field, 10, 64)
		values = append(values, value)
	}
	var total uint64
	for _, value := range values {
		total += value
	}
	idle := values[3]
	if len(values) > 4 {
		idle += values[4]
	}
	return cpuCounters{total: total, idle: idle}
}

func readMemory() (used, total uint64) {
	values := readKeyValueFile("/proc/meminfo")
	total = values["MemTotal"] * 1024
	available := values["MemAvailable"] * 1024
	if total >= available {
		used = total - available
	}
	return used, total
}

func readDisk(path string) (used, total uint64) {
	var stat syscall.Statfs_t
	if err := syscall.Statfs(path, &stat); err != nil {
		if err := syscall.Statfs("/", &stat); err != nil {
			return 0, 0
		}
	}
	total = stat.Blocks * uint64(stat.Bsize)
	available := stat.Bavail * uint64(stat.Bsize)
	if total >= available {
		used = total - available
	}
	return used, total
}

func readNetwork() networkCounters {
	entries, _ := os.ReadDir("/sys/class/net")
	var counters networkCounters
	for _, entry := range entries {
		name := entry.Name()
		if skipInterface(name) {
			continue
		}
		state, _ := os.ReadFile(filepath.Join("/sys/class/net", name, "operstate"))
		if strings.TrimSpace(string(state)) != "up" {
			continue
		}
		counters.rx += readUint(filepath.Join("/sys/class/net", name, "statistics/rx_bytes"))
		counters.tx += readUint(filepath.Join("/sys/class/net", name, "statistics/tx_bytes"))
	}
	return counters
}

func skipInterface(name string) bool {
	for _, prefix := range []string{
		"lo", "docker", "veth", "br", "virbr", "vmnet", "vnet", "nbd",
		"tailscale", "zt", "tun", "tap", "wg", "cni", "flannel", "lxc",
		"kube", "podman", "dummy", "ifb", "macvlan", "ipvlan", "bond",
	} {
		if strings.HasPrefix(name, prefix) {
			return true
		}
	}
	return false
}

func readDisplay() Display {
	connectors, _ := filepath.Glob("/sys/class/drm/card*-*/status")
	sort.Strings(connectors)
	result := Display{}
	for _, statusPath := range connectors {
		status, _ := os.ReadFile(statusPath)
		if strings.TrimSpace(string(status)) != "connected" {
			continue
		}
		connector := filepath.Base(filepath.Dir(statusPath))
		modesData, _ := os.ReadFile(filepath.Join(filepath.Dir(statusPath), "modes"))
		result.Connected = true
		result.Connector = connector
		for _, mode := range strings.Fields(string(modesData)) {
			if mode != "" {
				result.Modes = append(result.Modes, mode)
			}
		}
		break
	}
	return result
}

func detectGPU() GPU {
	cards, _ := filepath.Glob("/sys/class/drm/card*/device/vendor")
	for _, card := range cards {
		value, _ := os.ReadFile(card)
		switch strings.TrimSpace(string(value)) {
		case "0x8086":
			return GPU{Vendor: "Intel", Name: "Intel Graphics"}
		case "0x1002":
			return GPU{Vendor: "AMD", Name: "AMD Graphics"}
		case "0x10de":
			return GPU{Vendor: "NVIDIA", Name: "NVIDIA Graphics"}
		}
	}
	return GPU{}
}

func sampleNVIDIA(parent context.Context, fallback GPU) GPU {
	ctx, cancel := context.WithTimeout(parent, 800*time.Millisecond)
	defer cancel()
	command := exec.CommandContext(ctx, "nvidia-smi", "--query-gpu=name,utilization.gpu,memory.used,memory.total,temperature.gpu", "--format=csv,noheader,nounits")
	output, err := command.Output()
	if err != nil {
		return fallback
	}
	fields := strings.Split(strings.TrimSpace(strings.Split(string(output), "\n")[0]), ",")
	if len(fields) < 5 {
		return fallback
	}
	return GPU{
		Vendor:        "NVIDIA",
		Name:          strings.TrimSpace(fields[0]),
		UsagePercent:  parseFloat(fields[1]),
		MemoryUsedMB:  parseFloat(fields[2]),
		MemoryTotalMB: parseFloat(fields[3]),
		TemperatureC:  parseFloat(fields[4]),
	}
}

func readTemperature() float64 {
	patterns := []string{"/sys/class/thermal/thermal_zone*/temp", "/sys/class/hwmon/hwmon*/temp*_input"}
	var highest float64
	for _, pattern := range patterns {
		paths, _ := filepath.Glob(pattern)
		for _, path := range paths {
			value := float64(readUint(path))
			if value > 1000 {
				value /= 1000
			}
			if value > highest && value < 150 {
				highest = value
			}
		}
	}
	return round(highest, 1)
}

func readUptime() float64 {
	data, _ := os.ReadFile("/proc/uptime")
	fields := strings.Fields(string(data))
	if len(fields) == 0 {
		return 0
	}
	return parseFloat(fields[0])
}

func readLoad1() float64 {
	data, _ := os.ReadFile("/proc/loadavg")
	fields := strings.Fields(string(data))
	if len(fields) == 0 {
		return 0
	}
	return parseFloat(fields[0])
}

func readOSName() string {
	data, err := os.ReadFile("/etc/os-release")
	if err != nil {
		return runtime.GOOS
	}
	for _, line := range strings.Split(string(data), "\n") {
		if strings.HasPrefix(line, "PRETTY_NAME=") {
			return strings.Trim(strings.TrimPrefix(line, "PRETTY_NAME="), "\"")
		}
	}
	return runtime.GOOS
}

func readKernel() string {
	output, err := exec.Command("uname", "-r").Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(output))
}

func ipAddresses() []string {
	interfaces, _ := net.Interfaces()
	var candidates []interfaceAddress
	seen := make(map[string]struct{})
	for _, networkInterface := range interfaces {
		if networkInterface.Flags&net.FlagUp == 0 || networkInterface.Flags&net.FlagLoopback != 0 {
			continue
		}
		priority, ok := displayInterfacePriority(networkInterface.Name, "/sys/class/net")
		if !ok {
			continue
		}
		addresses, err := networkInterface.Addrs()
		if err != nil {
			continue
		}
		for _, address := range addresses {
			ip, _, err := net.ParseCIDR(address.String())
			if err != nil || !ip.IsGlobalUnicast() || ip.IsLinkLocalUnicast() {
				continue
			}
			value := ip.String()
			if _, exists := seen[value]; exists {
				continue
			}
			seen[value] = struct{}{}
			family := 1
			if ip.To4() != nil {
				family = 0
			}
			candidates = append(candidates, interfaceAddress{
				value: value, family: family, priority: priority, name: networkInterface.Name,
			})
		}
	}
	sort.Slice(candidates, func(i, j int) bool {
		if candidates[i].family != candidates[j].family {
			return candidates[i].family < candidates[j].family
		}
		if candidates[i].priority != candidates[j].priority {
			return candidates[i].priority < candidates[j].priority
		}
		if candidates[i].name != candidates[j].name {
			return candidates[i].name < candidates[j].name
		}
		return candidates[i].value < candidates[j].value
	})
	result := make([]string, 0, len(candidates))
	for _, candidate := range candidates {
		result = append(result, candidate.value)
	}
	return result
}

type interfaceAddress struct {
	value            string
	family, priority int
	name             string
}

func displayInterfacePriority(name, sysfsRoot string) (int, bool) {
	lowerName := strings.ToLower(name)
	if skipInterface(lowerName) {
		return 0, false
	}
	base := filepath.Join(sysfsRoot, name)
	if isThunderboltInterface(lowerName, base) {
		return 1, true
	}
	if pathExists(filepath.Join(base, "wireless")) || strings.HasPrefix(lowerName, "wl") {
		return 2, true
	}
	if pathExists(filepath.Join(base, "device")) {
		return 0, true
	}
	if _, err := os.Stat(sysfsRoot); os.IsNotExist(err) && commonPhysicalInterfaceName(lowerName) {
		return 0, true
	}
	return 0, false
}

func isThunderboltInterface(name, sysfsBase string) bool {
	if strings.Contains(name, "thunderbolt") || strings.HasPrefix(name, "tb") {
		return true
	}
	devicePath, err := filepath.EvalSymlinks(filepath.Join(sysfsBase, "device"))
	return err == nil && strings.Contains(strings.ToLower(devicePath), "thunderbolt")
}

func commonPhysicalInterfaceName(name string) bool {
	for _, prefix := range []string{"eth", "en", "wl", "wlan", "thunderbolt", "tb"} {
		if strings.HasPrefix(name, prefix) {
			return true
		}
	}
	return false
}

func pathExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func readKeyValueFile(path string) map[string]uint64 {
	result := map[string]uint64{}
	file, err := os.Open(path)
	if err != nil {
		return result
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		fields := strings.Fields(scanner.Text())
		if len(fields) < 2 {
			continue
		}
		value, err := strconv.ParseUint(fields[1], 10, 64)
		if err == nil {
			result[strings.TrimSuffix(fields[0], ":")] = value
		}
	}
	return result
}

func readUint(path string) uint64 {
	data, err := os.ReadFile(path)
	if err != nil {
		return 0
	}
	value, _ := strconv.ParseUint(strings.TrimSpace(string(data)), 10, 64)
	return value
}

func parseFloat(value string) float64 {
	parsed, _ := strconv.ParseFloat(strings.TrimSpace(value), 64)
	return parsed
}

func round(value float64, precision int) float64 {
	factor := 1.0
	for i := 0; i < precision; i++ {
		factor *= 10
	}
	return float64(int64(value*factor+0.5)) / factor
}

func (s Snapshot) String() string {
	return fmt.Sprintf("cpu=%.1f%% memory=%.1f%% disk=%.1f%%", s.CPUPercent, s.MemoryPercent, s.DiskPercent)
}
