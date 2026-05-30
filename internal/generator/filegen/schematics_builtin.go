package filegen

import "sync"

var (
	builtinOnce     sync.Once
	builtinRegistry *SchematicRegistry
)

// BuiltinRegistry returns the singleton registry containing all built-in schematics.
// It is initialized lazily on first call.
func BuiltinRegistry() *SchematicRegistry {
	builtinOnce.Do(func() {
		builtinRegistry = NewRegistry()
		registerBuiltins(builtinRegistry)
	})
	return builtinRegistry
}

func registerBuiltins(r *SchematicRegistry) {
	r.Register(schematicService())
	r.Register(schematicRepository())
	r.Register(schematicCommand())
	r.Register(schematicObserver())
	r.Register(schematicFactory())
	r.Register(schematicSingleton())
	r.Register(schematicModule())
	r.Register(schematicController())
	r.Register(schematicModel())
	r.Register(schematicViewModel())
	r.Register(schematicWindow())
	r.Register(schematicDialog())
	// Games
	r.Register(schematicSystem())
	r.Register(schematicComponent())
	r.Register(schematicSubsystem())
	r.Register(schematicEvent())
}

// layerPath returns a template fragment that resolves to:
//
//	<base>/{{.Layer}}/{{.NameSnake}}/<file>   when .Layer is non-empty
//	<base>/{{.NameSnake}}/<file>              otherwise
//
// base and file are literal strings (not template expressions).
func layerPath(base, file string) string {
	return `{{if .Layer}}` + base + `/{{.Layer}}/{{.NameSnake}}/` + file +
		`{{else}}` + base + `/{{.NameSnake}}/` + file + `{{end}}`
}

// layerPathWithDefault behaves like layerPath but substitutes defaultLayer
// when .Layer is empty, rather than omitting it.
func layerPathWithDefault(base, defaultLayer, file string) string {
	return `{{if .Layer}}` + base + `/{{.Layer}}/{{.NameSnake}}/` + file +
		`{{else}}` + base + `/` + defaultLayer + `/{{.NameSnake}}/` + file + `{{end}}`
}

// ---------------------------------------------------------------------------
// Domain / application schematics
// ---------------------------------------------------------------------------

func schematicService() Schematic {
	return Schematic{
		Name:        "service",
		Description: "Domain service with interface, implementation and test.",
		Files: []SchematicFileSpec{
			{
				Role:        "interface",
				PathTmpl:    layerPath(`{{.Include}}`, `i_{{.NameSnake}}_service.hpp`),
				ContentTmpl: "interface_hpp",
				ClassPrefix: "I",
				ClassSuffix: "Service",
			},
			{
				Role:        "impl_header",
				PathTmpl:    layerPath(`{{.Include}}`, `{{.NameSnake}}_service.hpp`),
				ContentTmpl: "class_hpp",
				ClassSuffix: "Service",
			},
			{
				Role:        "impl",
				PathTmpl:    layerPath(`{{.Src}}`, `{{.NameSnake}}_service.cpp`),
				ContentTmpl: "class_cpp",
				ClassSuffix: "Service",
			},
			{
				Role:        "test",
				PathTmpl:    `{{.Tests}}/{{.NameSnake}}/{{.NameSnake}}_service_test.cpp`,
				ContentTmpl: "test_catch2",
				ClassSuffix: "Service",
			},
		},
	}
}

func schematicRepository() Schematic {
	return Schematic{
		Name:        "repository",
		Description: "Repository pattern with interface, implementation and test.",
		Files: []SchematicFileSpec{
			{
				Role:        "interface",
				PathTmpl:    layerPath(`{{.Include}}`, `i_{{.NameSnake}}_repository.hpp`),
				ContentTmpl: "interface_hpp",
				ClassPrefix: "I",
				ClassSuffix: "Repository",
			},
			{
				Role:        "impl_header",
				PathTmpl:    layerPath(`{{.Include}}`, `{{.NameSnake}}_repository.hpp`),
				ContentTmpl: "class_hpp",
				ClassSuffix: "Repository",
			},
			{
				Role:        "impl",
				PathTmpl:    layerPath(`{{.Src}}`, `{{.NameSnake}}_repository.cpp`),
				ContentTmpl: "class_cpp",
				ClassSuffix: "Repository",
			},
			{
				Role:        "test",
				PathTmpl:    `{{.Tests}}/{{.NameSnake}}/{{.NameSnake}}_repository_test.cpp`,
				ContentTmpl: "test_catch2",
				ClassSuffix: "Repository",
			},
		},
	}
}

func schematicCommand() Schematic {
	return Schematic{
		Name:        "command",
		Description: "Command pattern with base interface, implementation and test.",
		Files: []SchematicFileSpec{
			{
				Role:        "base",
				PathTmpl:    `{{.Include}}/i_command.hpp`,
				ContentTmpl: "interface_hpp",
				Brief:       "Base interface for all commands.",
				ClassPrefix: "I",
				ClassSuffix: "Command",
			},
			{
				Role:        "impl_header",
				PathTmpl:    layerPath(`{{.Include}}`, `{{.NameSnake}}_command.hpp`),
				ContentTmpl: "class_hpp",
				ClassSuffix: "Command",
			},
			{
				Role:        "impl",
				PathTmpl:    layerPath(`{{.Src}}`, `{{.NameSnake}}_command.cpp`),
				ContentTmpl: "class_cpp",
				ClassSuffix: "Command",
			},
			{
				Role:        "test",
				PathTmpl:    `{{.Tests}}/{{.NameSnake}}/{{.NameSnake}}_command_test.cpp`,
				ContentTmpl: "test_catch2",
				ClassSuffix: "Command",
			},
		},
	}
}

func schematicObserver() Schematic {
	return Schematic{
		Name:        "observer",
		Description: "Observer pattern with interface, subject and concrete observer.",
		Files: []SchematicFileSpec{
			{
				Role:        "interface",
				PathTmpl:    `{{.Include}}/i_observer.hpp`,
				ContentTmpl: "interface_hpp",
				Brief:       "Observer interface.",
				ClassPrefix: "I",
				ClassSuffix: "Observer",
			},
			{
				Role:        "subject",
				PathTmpl:    layerPath(`{{.Include}}`, `{{.NameSnake}}_subject.hpp`),
				ContentTmpl: "class_hpp",
				Brief:       "Observable subject for {{.NamePascal}}.",
				ClassSuffix: "Subject",
			},
			{
				Role:        "impl_header",
				PathTmpl:    layerPath(`{{.Include}}`, `{{.NameSnake}}_observer.hpp`),
				ContentTmpl: "class_hpp",
				ClassSuffix: "Observer",
			},
			{
				Role:        "impl",
				PathTmpl:    layerPath(`{{.Src}}`, `{{.NameSnake}}_observer.cpp`),
				ContentTmpl: "class_cpp",
				ClassSuffix: "Observer",
			},
		},
	}
}

func schematicFactory() Schematic {
	return Schematic{
		Name:        "factory",
		Description: "Factory pattern with product interface, factory header, source and test.",
		Files: []SchematicFileSpec{
			{
				Role:        "interface",
				PathTmpl:    layerPath(`{{.Include}}`, `i_{{.NameSnake}}.hpp`),
				ContentTmpl: "interface_hpp",
				ClassPrefix: "I",
			},
			{
				Role:        "factory_hdr",
				PathTmpl:    layerPath(`{{.Include}}`, `{{.NameSnake}}_factory.hpp`),
				ContentTmpl: "class_hpp",
				ClassSuffix: "Factory",
			},
			{
				Role:        "factory_cpp",
				PathTmpl:    layerPath(`{{.Src}}`, `{{.NameSnake}}_factory.cpp`),
				ContentTmpl: "class_cpp",
				ClassSuffix: "Factory",
			},
			{
				Role:        "test",
				PathTmpl:    `{{.Tests}}/{{.NameSnake}}/{{.NameSnake}}_factory_test.cpp`,
				ContentTmpl: "test_catch2",
				ClassSuffix: "Factory",
			},
		},
	}
}

func schematicSingleton() Schematic {
	return Schematic{
		Name:        "singleton",
		Description: "Thread-safe Meyer's singleton.",
		Files: []SchematicFileSpec{
			{
				Role:        "header",
				PathTmpl:    layerPath(`{{.Include}}`, `{{.NameSnake}}.hpp`),
				ContentTmpl: "class_hpp",
				Brief:       "Thread-safe Meyer's singleton.",
			},
			{
				Role:        "impl",
				PathTmpl:    layerPath(`{{.Src}}`, `{{.NameSnake}}.cpp`),
				ContentTmpl: "class_cpp",
			},
		},
	}
}

// ---------------------------------------------------------------------------
// Structural / module schematics
// ---------------------------------------------------------------------------

func schematicModule() Schematic {
	return Schematic{
		Name:        "module",
		Description: "Self-contained module with types, class and test.",
		Files: []SchematicFileSpec{
			{
				Role:        "types",
				PathTmpl:    `{{.Src}}/{{.NameSnake}}/{{.NameSnake}}_types.hpp`,
				ContentTmpl: "struct_hpp",
				Brief:       "Public types for the {{.NamePascal}} module.",
			},
			{
				Role:        "header",
				PathTmpl:    `{{.Src}}/{{.NameSnake}}/{{.NameSnake}}.hpp`,
				ContentTmpl: "class_hpp",
			},
			{
				Role:        "impl",
				PathTmpl:    `{{.Src}}/{{.NameSnake}}/{{.NameSnake}}.cpp`,
				ContentTmpl: "class_cpp",
			},
			{
				Role:        "test",
				PathTmpl:    `{{.Tests}}/{{.NameSnake}}/{{.NameSnake}}_test.cpp`,
				ContentTmpl: "test_catch2",
			},
		},
	}
}

// ---------------------------------------------------------------------------
// MVC / MVVM schematics
// ---------------------------------------------------------------------------

func schematicController() Schematic {
	return Schematic{
		Name:        "controller",
		Description: "MVC controller with header, source and test.",
		Files: []SchematicFileSpec{
			{
				Role:        "header",
				PathTmpl:    layerPathWithDefault(`{{.Include}}`, "controllers", `{{.NameSnake}}_controller.hpp`),
				ContentTmpl: "class_hpp",
				ClassSuffix: "Controller",
			},
			{
				Role:        "impl",
				PathTmpl:    layerPathWithDefault(`{{.Src}}`, "controllers", `{{.NameSnake}}_controller.cpp`),
				ContentTmpl: "class_cpp",
				ClassSuffix: "Controller",
			},
			{
				Role:        "test",
				PathTmpl:    `{{.Tests}}/{{.NameSnake}}/{{.NameSnake}}_controller_test.cpp`,
				ContentTmpl: "test_catch2",
				ClassSuffix: "Controller",
			},
		},
	}
}

func schematicModel() Schematic {
	return Schematic{
		Name:        "model",
		Description: "MVC data model with header, source and test.",
		Files: []SchematicFileSpec{
			{
				Role:        "header",
				PathTmpl:    layerPathWithDefault(`{{.Include}}`, "models", `{{.NameSnake}}_model.hpp`),
				ContentTmpl: "class_hpp",
				ClassSuffix: "Model",
			},
			{
				Role:        "impl",
				PathTmpl:    layerPathWithDefault(`{{.Src}}`, "models", `{{.NameSnake}}_model.cpp`),
				ContentTmpl: "class_cpp",
				ClassSuffix: "Model",
			},
			{
				Role:        "test",
				PathTmpl:    `{{.Tests}}/{{.NameSnake}}/{{.NameSnake}}_model_test.cpp`,
				ContentTmpl: "test_catch2",
				ClassSuffix: "Model",
			},
		},
	}
}

func schematicViewModel() Schematic {
	return Schematic{
		Name:        "viewmodel",
		Description: "MVVM ViewModel with header, source and test.",
		Files: []SchematicFileSpec{
			{
				Role:        "header",
				PathTmpl:    layerPathWithDefault(`{{.Include}}`, "viewmodels", `{{.NameSnake}}_viewmodel.hpp`),
				ContentTmpl: "class_hpp",
				ClassSuffix: "ViewModel",
			},
			{
				Role:        "impl",
				PathTmpl:    layerPathWithDefault(`{{.Src}}`, "viewmodels", `{{.NameSnake}}_viewmodel.cpp`),
				ContentTmpl: "class_cpp",
				ClassSuffix: "ViewModel",
			},
			{
				Role:        "test",
				PathTmpl:    `{{.Tests}}/{{.NameSnake}}/{{.NameSnake}}_viewmodel_test.cpp`,
				ContentTmpl: "test_catch2",
				ClassSuffix: "ViewModel",
			},
		},
	}
}

// ---------------------------------------------------------------------------
// UI schematics
// ---------------------------------------------------------------------------

func schematicWindow() Schematic {
	return Schematic{
		Name:        "window",
		Description: "Desktop application window with header and source.",
		Files: []SchematicFileSpec{
			{
				Role:        "header",
				PathTmpl:    layerPathWithDefault(`{{.Include}}`, "views", `{{.NameSnake}}_window.hpp`),
				ContentTmpl: "class_hpp",
				Brief:       "{{.NamePascal}} application window.",
				ClassSuffix: "Window",
			},
			{
				Role:        "impl",
				PathTmpl:    layerPathWithDefault(`{{.Src}}`, "views", `{{.NameSnake}}_window.cpp`),
				ContentTmpl: "class_cpp",
				ClassSuffix: "Window",
			},
		},
	}
}

func schematicDialog() Schematic {
	return Schematic{
		Name:        "dialog",
		Description: "Modal dialog with header and source.",
		Files: []SchematicFileSpec{
			{
				Role:        "header",
				PathTmpl:    layerPathWithDefault(`{{.Include}}`, "views", `{{.NameSnake}}_dialog.hpp`),
				ContentTmpl: "class_hpp",
				Brief:       "{{.NamePascal}} modal dialog.",
				ClassSuffix: "Dialog",
			},
			{
				Role:        "impl",
				PathTmpl:    layerPathWithDefault(`{{.Src}}`, "views", `{{.NameSnake}}_dialog.cpp`),
				ContentTmpl: "class_cpp",
				ClassSuffix: "Dialog",
			},
		},
	}
}

// ---------------------------------------------------------------------------
// Game / ECS schematics
// ---------------------------------------------------------------------------

func schematicSystem() Schematic {
	return Schematic{
		Name:        "system",
		Description: "ECS System with base interface, implementation and test.",
		Files: []SchematicFileSpec{
			{
				Role:        "interface",
				PathTmpl:    `{{.Include}}/i_system.hpp`,
				ContentTmpl: "interface_hpp",
				Brief:       "Base interface for all ECS systems.",
				ClassPrefix: "I",
				ClassSuffix: "System",
			},
			{
				Role:        "header",
				PathTmpl:    layerPathWithDefault(`{{.Include}}`, "systems", `{{.NameSnake}}_system.hpp`),
				ContentTmpl: "class_hpp",
				ClassSuffix: "System",
			},
			{
				Role:        "impl",
				PathTmpl:    layerPathWithDefault(`{{.Src}}`, "systems", `{{.NameSnake}}_system.cpp`),
				ContentTmpl: "class_cpp",
				ClassSuffix: "System",
			},
			{
				Role:        "test",
				PathTmpl:    `{{.Tests}}/{{.NameSnake}}/{{.NameSnake}}_system_test.cpp`,
				ContentTmpl: "test_catch2",
				ClassSuffix: "System",
			},
		},
	}
}

func schematicComponent() Schematic {
	return Schematic{
		Name:        "component",
		Description: "ECS Component — pure data struct.",
		Files: []SchematicFileSpec{
			{
				Role:        "header",
				PathTmpl:    layerPathWithDefault(`{{.Include}}`, "components", `{{.NameSnake}}_component.hpp`),
				ContentTmpl: "struct_hpp",
				Brief:       "{{.NamePascal}} ECS component data.",
			},
		},
	}
}

func schematicSubsystem() Schematic {
	return Schematic{
		Name:        "subsystem",
		Description: "Subsystem with Init/Update/Shutdown lifecycle.",
		Files: []SchematicFileSpec{
			{
				Role:        "interface",
				PathTmpl:    `{{.Include}}/i_subsystem.hpp`,
				ContentTmpl: "interface_hpp",
				Brief:       "Base interface: Init, Update, Shutdown.",
				ClassPrefix: "I",
				ClassSuffix: "Subsystem",
			},
			{
				Role:        "header",
				PathTmpl:    layerPathWithDefault(`{{.Include}}`, "systems", `{{.NameSnake}}_subsystem.hpp`),
				ContentTmpl: "class_hpp",
				ClassSuffix: "Subsystem",
			},
			{
				Role:        "impl",
				PathTmpl:    layerPathWithDefault(`{{.Src}}`, "systems", `{{.NameSnake}}_subsystem.cpp`),
				ContentTmpl: "class_cpp",
				ClassSuffix: "Subsystem",
			},
		},
	}
}

func schematicEvent() Schematic {
	return Schematic{
		Name:        "event",
		Description: "Event data struct.",
		Files: []SchematicFileSpec{
			{
				Role:        "header",
				PathTmpl:    layerPathWithDefault(`{{.Include}}`, "events", `{{.NameSnake}}_event.hpp`),
				ContentTmpl: "struct_hpp",
				Brief:       "{{.NamePascal}} event data.",
			},
		},
	}
}
