# gdt-dotnet

C# and .NET tooling plugin for [gdt](https://github.com/monkeymonk/gdt) (Godot Developer Toolchain).

## Installation

```sh
gdt plugin install monkeymonk/gdt-dotnet
```

Or build manually:

```sh
git clone https://github.com/monkeymonk/gdt-dotnet.git
cd gdt-dotnet
make install
```

## Commands

### doctor

Check that .NET SDK, MSBuild, and project files are properly configured:

```sh
gdt dotnet doctor
```

Verifies:
- .NET SDK is installed
- SDK version is compatible with Godot (net6.0 or net8.0)
- MSBuild is available
- `.csproj` files exist in the project root

### restore

Restore NuGet packages:

```sh
gdt dotnet restore
```

### build

Build the C# project:

```sh
gdt dotnet build
```

### run

Run the project via `dotnet run`, falling back to the Godot engine if unavailable:

```sh
gdt dotnet run
```

## Environment Variables

| Variable | Description |
|---|---|
| `GDT_PROJECT_ROOT` | Path to the Godot project root (required for restore/build/run) |
| `GDT_GODOT_VERSION` | Current Godot version |
| `GDT_ENGINE_PATH` | Path to the Godot engine binary (fallback for run) |
| `GDT_HOME` | Path to gdt home directory |

## License

MIT
