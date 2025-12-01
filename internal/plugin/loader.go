package plugin

import (
	"fmt"
	"os"
	"path/filepath"
	"plugin"
	"reflect"
	"runtime"

	"github.com/ops-cli/internal/core"
	"github.com/spf13/cobra"
)

// Loader manages loading of plugins
type Loader struct {
	pluginDir string
	xdg       *core.XDGPaths
}

// NewLoader creates a new plugin loader
func NewLoader() (*Loader, error) {
	xdg := core.NewXDGPaths("ops-cli")
	pluginDir, err := xdg.PluginDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get plugin directory: %w", err)
	}

	return &Loader{
		pluginDir: pluginDir,
		xdg:       xdg,
	}, nil
}

// LoadPlugin loads a plugin by name
func (l *Loader) LoadPlugin(name string) (Plugin, error) {
	pluginPath := filepath.Join(l.pluginDir, fmt.Sprintf("%s.so", name))

	// Check if plugin file exists
	if _, err := os.Stat(pluginPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("plugin %s not found at %s", name, pluginPath)
	}

	// Load the plugin
	p, err := plugin.Open(pluginPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open plugin %s: %w", name, err)
	}

	// Lookup the Plugin symbol
	symbol, err := p.Lookup(PluginSymbol)
	if err != nil {
		return nil, fmt.Errorf("plugin %s does not export %s symbol: %w", name, PluginSymbol, err)
	}

	// Try direct type assertion first
	if pluginInstance, ok := symbol.(Plugin); ok {
		return pluginInstance, nil
	}

	// If direct assertion fails, use reflection as workaround for Go plugin type limitations
	// This handles cases where the plugin is in package 'main' and interface is in another package
	val := reflect.ValueOf(symbol)

	// Handle pointers - keep a pointer to access pointer receiver methods
	// Dereference until we have a single pointer (not double pointer)
	for val.Kind() == reflect.Ptr && val.Elem().Kind() == reflect.Ptr {
		if val.IsNil() {
			return nil, fmt.Errorf("plugin %s exported nil Plugin", name)
		}
		val = val.Elem()
	}

	// Ensure we have a pointer (methods might be on pointer receiver)
	if val.Kind() != reflect.Ptr {
		// Create a pointer to the value
		ptr := reflect.New(val.Type())
		ptr.Elem().Set(val)
		val = ptr
	}

	// Check if the value has the required methods (using reflection)
	if !hasMethod(val, "Name") || !hasMethod(val, "Version") || !hasMethod(val, "Register") {
		return nil, fmt.Errorf("plugin %s does not implement required Plugin methods (type: %s)", name, val.Type())
	}

	// Create a wrapper that uses reflection to call methods
	wrapper := &reflectionPluginWrapper{
		val:  val,
		name: name,
	}

	return wrapper, nil
}

// reflectionPluginWrapper wraps a plugin using reflection to work around Go plugin type limitations
type reflectionPluginWrapper struct {
	val  reflect.Value
	name string
}

func (w *reflectionPluginWrapper) Name() string {
	method := w.val.MethodByName("Name")
	if !method.IsValid() {
		return w.name
	}
	result := method.Call(nil)
	if len(result) > 0 {
		if str, ok := result[0].Interface().(string); ok {
			return str
		}
	}
	return w.name
}

func (w *reflectionPluginWrapper) Version() string {
	method := w.val.MethodByName("Version")
	if !method.IsValid() {
		return "unknown"
	}
	result := method.Call(nil)
	if len(result) > 0 {
		if str, ok := result[0].Interface().(string); ok {
			return str
		}
	}
	return "unknown"
}

func (w *reflectionPluginWrapper) Register(rootCmd *cobra.Command) error {
	method := w.val.MethodByName("Register")
	if !method.IsValid() {
		return fmt.Errorf("plugin %s does not have Register method", w.name)
	}
	result := method.Call([]reflect.Value{reflect.ValueOf(rootCmd)})
	if len(result) > 0 {
		if err, ok := result[0].Interface().(error); ok && err != nil {
			return err
		}
	}
	return nil
}

func hasMethod(val reflect.Value, methodName string) bool {
	method := val.MethodByName(methodName)
	return method.IsValid()
}

// DiscoverPlugins discovers all installed plugins
func (l *Loader) DiscoverPlugins() ([]Plugin, error) {
	var plugins []Plugin

	// Ensure plugin directory exists
	if err := os.MkdirAll(l.pluginDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create plugin directory: %w", err)
	}

	// Find all .so files in plugin directory
	pattern := filepath.Join(l.pluginDir, "*.so")
	files, err := filepath.Glob(pattern)
	if err != nil {
		return nil, fmt.Errorf("failed to scan plugin directory: %w", err)
	}

	for _, file := range files {
		name := filepath.Base(file)
		name = name[:len(name)-3] // Remove .so extension

		p, err := l.LoadPlugin(name)
		if err != nil {
			// Log error but continue loading other plugins
			fmt.Fprintf(os.Stderr, "Warning: failed to load plugin %s: %v\n", name, err)
			continue
		}

		plugins = append(plugins, p)
	}

	return plugins, nil
}

// RegisterPlugins discovers and registers all plugins with the root command
func (l *Loader) RegisterPlugins(rootCmd *cobra.Command) error {
	plugins, err := l.DiscoverPlugins()
	if err != nil {
		return fmt.Errorf("failed to discover plugins: %w", err)
	}

	for _, p := range plugins {
		if err := p.Register(rootCmd); err != nil {
			fmt.Fprintf(os.Stderr, "Warning: failed to register plugin %s: %v\n", p.Name(), err)
			continue
		}
	}

	return nil
}

// GetPluginDir returns the plugin directory path
func (l *Loader) GetPluginDir() string {
	return l.pluginDir
}

// CheckVersionCompatibility checks if a plugin is compatible with the current Go version
func (l *Loader) CheckVersionCompatibility(pluginPath string) error {
	// This is a placeholder for future version checking
	// For now, we rely on the fact that plugins are built with the same Go version
	// as the core binary (since they're distributed together)

	// Future: could read metadata from plugin or check Go version used to build
	currentGoVersion := runtime.Version()
	_ = currentGoVersion // Placeholder

	return nil
}
