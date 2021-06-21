// SPDX-License-Identifier: Apache-2.0

package modules

import (
	"errors"

	"github.com/gookit/color"
	"github.com/i582/cfmt/cmd/cfmt"
	log "github.com/sirupsen/logrus"

	"spdx-sbom-generator/internal/models"
	"spdx-sbom-generator/internal/modules/cargo"
	"spdx-sbom-generator/internal/modules/composer"
	"spdx-sbom-generator/internal/modules/gem"
	"spdx-sbom-generator/internal/modules/gomod"
	"spdx-sbom-generator/internal/modules/javamaven"
	"spdx-sbom-generator/internal/modules/npm"
	"spdx-sbom-generator/internal/modules/nuget"
	"spdx-sbom-generator/internal/modules/pip"
	"spdx-sbom-generator/internal/modules/yarn"
)

var (
	errNoPluginAvailable   = errors.New("no plugin system available for current path")
	errNoModulesInstalled  = errors.New("there are no components in the BOM. The project may not contain dependencies, please install modules")
	errFailedToReadModules = errors.New("failed to read modules")
)

var registeredPlugins []models.IPlugin

func init() {
	registeredPlugins = append(registeredPlugins,
		cargo.New(),
		composer.New(),
		gomod.New(),
		gem.New(),
		npm.New(),
		javamaven.New(),
		nuget.New(),
		yarn.New(),
		pip.New(),
	)
}

// Manager ...
type Manager struct {
	Config  Config
	Plugin  models.IPlugin
	modules []models.Module
}

// Config ...
type Config struct {
	Path string
}

// New ...
func New(cfg Config) ([]*Manager, error) {
	var usePlugin models.IPlugin
	var managerSlice []*Manager
	for _, plugin := range registeredPlugins {
		if plugin.IsValid(cfg.Path) {
			if err := plugin.SetRootModule(cfg.Path); err != nil {
				return nil, err
			}

			usePlugin = plugin
			if usePlugin == nil {
				return nil, errNoPluginAvailable
			}

			managerSlice = append(managerSlice, &Manager{
				Config: cfg,
				Plugin: usePlugin,
			})
		}
	}

	return managerSlice, nil
}

// Run ...
func (m *Manager) Run() error {
	modulePath := m.Config.Path
	version, err := m.Plugin.GetVersion()
	if err != nil {
		return err
	}

	cfmt.Print(cfmt.Sprintf("{{Versions detected: }}::cyan|bold %s", color.Yellow.Sprintf(version)))
	if err := m.Plugin.HasModulesInstalled(modulePath); err != nil {
		return err
	}

	modules, err := m.Plugin.ListModulesWithDeps(modulePath)
	if err != nil {
		log.Error(err)
		return errFailedToReadModules
	}

	m.modules = modules

	return nil
}

// GetSource ...
func (m *Manager) GetSource() []models.Module {
	return m.modules
}
