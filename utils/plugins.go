// Package utils contains various helpers.
package utils

import (
	"errors"
	"fmt"
	"os"
	"plugin"
	"reflect"
	"strings"

	"net/http"

	"io"

	"path/filepath"

	"github.com/go-home-io/server/plugins/common"
	"github.com/go-home-io/server/providers"
	"github.com/go-home-io/server/systems"
	"gopkg.in/yaml.v2"
)

const (
	// PluginEntryPointMethodName is the name of main plugin method.
	PluginEntryPointMethodName = "Load"
	// PluginInterfaceInitMethodName is the name of first initialization method.
	PluginInterfaceInitMethodName = "Init"
	// PluginCDNUrlFormat is format for bintray CDN.
	PluginCDNUrlFormat = "https://dl.bintray.com/go-home-io/%s/%s"
)

// Arch describes build architecture.
var Arch string

// Version describes build version.
var Version string

// ConstructPluginLoader contains params required for creating a new plugin loader instance.
type ConstructPluginLoader struct {
	PluginsFolder string
	Validator     providers.IValidatorProvider
}

// Plugins loader.
type pluginLoader struct {
	pluginsFolder string
	validator     providers.IValidatorProvider

	loadedPlugins map[string]func() (interface{}, interface{}, error)
}

// NewPluginLoader creates a new plugins loader.
func NewPluginLoader(ctor *ConstructPluginLoader) providers.IPluginLoaderProvider {
	loc := ctor.PluginsFolder
	if "" == loc {
		loc = fmt.Sprintf("%s/plugins", GetCurrentWorkingDir())
	}

	loader := pluginLoader{
		pluginsFolder: loc,
		validator:     ctor.Validator,
		loadedPlugins: make(map[string]func() (interface{}, interface{}, error)),
	}

	return &loader
}

// LoadPlugin loads requested plugin.
// Returns main interface implementation which should be casted to package interface.
func (l *pluginLoader) LoadPlugin(request *providers.PluginLoadRequest) (interface{}, error) {
	pKey := getPluginKey(request.SystemType, request.PluginProvider)
	if method, ok := l.loadedPlugins[pKey]; ok {
		return l.loadPlugin(request, method)
	}

	fileName := l.getActualFileName(pKey)
	if _, err := os.Stat(fileName); err != nil {
		err = l.downloadFile(pKey, fileName)
		if err != nil {
			return nil, err
		}
	}

	p, err := plugin.Open(fileName)
	if err != nil {
		// We want to delete failed plugin
		os.Remove(fileName)
		return nil, errors.New("didn't find plugin file")
	}

	LoadSymbol, err := p.Lookup(PluginEntryPointMethodName)
	if err != nil {
		return nil, errors.New("didn't find entry point")
	}
	LoadMethod := LoadSymbol.(func() (interface{}, interface{}, error))
	if LoadMethod == nil {
		return nil, errors.New("wrong entry point signature")
	}

	l.loadedPlugins[pKey] = LoadMethod

	return l.loadPlugin(request, LoadMethod)
}

// Internal plugin cache key.
func getPluginKey(subSystemType systems.SystemType, pluginName string) string {
	switch subSystemType {
	case systems.SysDevice:
		return pluginName
	default:
		return fmt.Sprintf("%s/%s", subSystemType.String(), pluginName)
	}
}

// Performs actual plugin load
func (l *pluginLoader) loadPlugin(request *providers.PluginLoadRequest,
	loadMethod func() (interface{}, interface{}, error)) (interface{}, error) {
	pluginObject, settingsObject, err := loadMethod()
	if err != nil {
		return nil, err
	}

	if !reflect.TypeOf(pluginObject).AssignableTo(request.ExpectedType) {
		return nil, errors.New("plugin doesn't implement requested interface")
	}

	if nil == request.RawConfig || nil == settingsObject {
		err = l.initPlugin(request, pluginObject)
		if err != nil {
			return nil, err
		}

		return pluginObject, nil
	}

	err = yaml.Unmarshal(request.RawConfig, settingsObject)
	if err != nil {
		return nil, err
	}

	settingsInterface, ok := settingsObject.(common.ISettings)

	if !ok {
		return nil, errors.New("wrong settings signature")
	}
	if !l.validator.Validate(settingsObject) {
		return nil, errors.New("invalid config")
	}

	err = settingsInterface.Validate()
	if err != nil {
		return nil, err
	}

	err = l.initPlugin(request, pluginObject)
	if err != nil {
		return nil, err
	}

	return pluginObject, nil
}

// Calling Init method of a plugin.
func (l *pluginLoader) initPlugin(request *providers.PluginLoadRequest, pluginObject interface{}) error {
	method := reflect.ValueOf(pluginObject).MethodByName(PluginInterfaceInitMethodName)
	if !method.IsValid() {
		return errors.New("init method not found")
	}

	var results []reflect.Value

	if nil == request.InitData {
		results = method.Call(nil)
	} else {
		val := reflect.ValueOf(request.InitData)

		if reflect.ValueOf(request.InitData).Kind() != method.Type().In(0).Kind() {
			val = val.Elem()
		}

		rv := reflect.ValueOf(method.Type().In(0))
		for rv.Kind() == reflect.Ptr || rv.Kind() == reflect.Interface {
			rv = rv.Elem()
		}

		results = method.Call([]reflect.Value{val})
	}

	if len(results) > 0 && results[0].Interface() != nil {
		return results[0].Interface().(error)
	}

	return nil
}

// Gets actual plugin name.
func (l *pluginLoader) getActualFileName(pluginKey string) string {
	actualVersion := ""
	if "" != Version {
		actualVersion = fmt.Sprintf("-%s", Version)
	}

	return fmt.Sprintf("%s/%s%s.so", l.pluginsFolder, pluginKey, actualVersion)
}

// Downloads plugin from bintray CDN.
func (l *pluginLoader) downloadFile(pluginKey string, actualName string) error {
	name := strings.Replace(pluginKey, "/", "_", -1)
	name = fmt.Sprintf("%s-%s.so", name, Version)
	println("Downloading " + name)

	os.MkdirAll(filepath.Dir(actualName), os.ModePerm)
	out, err := os.Create(actualName)
	if err != nil {
		println("Failed to load " + name + ": " + err.Error())
		return err
	}

	defer out.Close()
	downloadURL := fmt.Sprintf(PluginCDNUrlFormat, Arch, name)
	res, err := http.Get(downloadURL)
	if err != nil {
		println("Failed to get " + downloadURL + ": " + err.Error())
		return err
	}

	defer res.Body.Close()
	_, err = io.Copy(out, res.Body)
	if err != nil {
		println("Failed to save " + name + ": " + err.Error())
	}

	return err
}
