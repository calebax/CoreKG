package dbplugins

type DatabaseType string

const (
	DatabaseTypeMySQL DatabaseType = "mysql"
)

type Engine struct {
	Plugins []*Plugin
}

func (e *Engine) RegistryPlugin(plugin *Plugin) {
	if e.Plugins == nil {
		e.Plugins = []*Plugin{}
	}
	e.Plugins = append(e.Plugins, plugin)
}

func (e *Engine) ChoosePlugin(dbType DatabaseType) *Plugin {
	for _, plugin := range e.Plugins {
		if plugin.Type == dbType {
			return plugin
		}
	}
	return nil
}
