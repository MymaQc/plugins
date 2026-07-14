// Command csharp-gen emits the supported C# surface directly from Dragonfly's Go AST.
package main

import (
	"bytes"
	"flag"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"sort"
	"strconv"
	"strings"

	"github.com/bedrock-gophers/plugins/internal/host"
	_ "github.com/df-mc/dragonfly/server/block"
	"github.com/df-mc/dragonfly/server/world"
)

type method struct {
	Name         string
	Parameters   []parameter
	Subscription uint64
}

type parameter struct {
	Name string
	Type string
}

type commandInterface struct {
	Name       string
	Embeddings []string
	Methods    []commandMethod
}

type commandMethod struct {
	Name       string
	Parameters []parameter
	ReturnType string
}

type generatedFile struct {
	Path    string
	Content []byte
}

type cubeSpec struct {
	Faces []string
}

type encodedBlock struct {
	Name          string
	Identifier    string
	PropertiesNBT []byte
}

type blockSpec struct {
	Stateless []encodedBlock
	Sand      [2]encodedBlock
}

var supportedPlayerHandlers = map[string]uint64{
	"HandleChat":         1 << 1,
	"HandleFoodLoss":     1 << 8,
	"HandleJump":         1 << 14,
	"HandleMove":         1,
	"HandlePunchAir":     1 << 17,
	"HandleQuit":         1 << 3,
	"HandleTeleport":     1 << 15,
	"HandleToggleSneak":  1 << 13,
	"HandleToggleSprint": 1 << 12,
}

var selectedCommandInterfaces = []string{
	"Runnable",
	"Allower",
	"Target",
	"NamedTarget",
	"Source",
	"Enum",
}

var selectedPlayerTextMethods = []string{
	"Message",
	"SendPopup",
	"SendTip",
	"SendJukeboxPopup",
	"SetNameTag",
	"Disconnect",
}

var selectedWorldTxMethods = []string{
	"Range",
	"SetBlock",
	"Block",
	"BlockLoaded",
}

func main() {
	root := flag.String("root", ".", "repository root")
	dragonfly := flag.String("dragonfly", "", "Dragonfly module directory")
	check := flag.Bool("check", false, "fail if generated output differs")
	flag.Parse()

	directory := *dragonfly
	if directory == "" {
		command := exec.Command("go", "list", "-m", "-f", "{{.Dir}}", "github.com/df-mc/dragonfly")
		command.Dir = *root
		output, err := command.Output()
		if err != nil {
			fatal(err)
		}
		directory = string(bytes.TrimSpace(output))
	}
	methods, err := playerHandlerMethods(filepath.Join(directory, "server", "player", "handler.go"))
	if err != nil {
		fatal(err)
	}
	playerMethods, err := playerTextMethods(filepath.Join(directory, "server", "player", "player.go"))
	if err != nil {
		fatal(err)
	}
	interfaces, err := commandInterfaces(filepath.Join(directory, "server", "cmd"))
	if err != nil {
		fatal(err)
	}
	cube, err := inspectCube(filepath.Join(directory, "server", "block", "cube"))
	if err != nil {
		fatal(err)
	}
	setOpts, err := inspectSetOpts(filepath.Join(directory, "server", "world", "world.go"))
	if err != nil {
		fatal(err)
	}
	worldTx, err := inspectWorldTx(filepath.Join(directory, "server", "world", "tx.go"))
	if err != nil {
		fatal(err)
	}
	blocks, err := inspectBlocks(filepath.Join(directory, "server", "block"))
	if err != nil {
		fatal(err)
	}
	files := []generatedFile{
		{
			Path:    filepath.Join(*root, "csharp", "Dragonfly", "Generated", "Player.Handler.g.cs"),
			Content: generatePlayerHandler(methods),
		},
		{
			Path:    filepath.Join(*root, "csharp", "Dragonfly", "Generated", "Player.Text.g.cs"),
			Content: generatePlayerTextMethods(playerMethods),
		},
		{
			Path:    filepath.Join(*root, "csharp", "Dragonfly", "Generated", "Cmd.Interfaces.g.cs"),
			Content: generateCommandInterfaces(interfaces),
		},
		{
			Path:    filepath.Join(*root, "csharp", "Dragonfly", "Generated", "Cube.g.cs"),
			Content: generateCube(cube),
		},
		{
			Path:    filepath.Join(*root, "csharp", "Dragonfly", "Generated", "World.Block.g.cs"),
			Content: generateWorldBlock(setOpts, worldTx),
		},
		{
			Path:    filepath.Join(*root, "csharp", "Dragonfly", "Generated", "Block.Types.g.cs"),
			Content: generateBlocks(blocks),
		},
	}
	if err := syncGeneratedFiles(files, *check); err != nil {
		fatal(err)
	}
}

func playerTextMethods(path string) ([]method, error) {
	file, err := parser.ParseFile(token.NewFileSet(), path, nil, 0)
	if err != nil {
		return nil, err
	}
	found := map[string]method{}
	for _, declaration := range file.Decls {
		function, ok := declaration.(*ast.FuncDecl)
		if !ok || !selectedPlayerTextMethod(function.Name.Name) || !playerMethod(function) {
			continue
		}
		if function.Type.Results != nil && len(function.Type.Results.List) != 0 {
			return nil, fmt.Errorf("player.Player.%s returns values", function.Name.Name)
		}
		parameters, err := translateParameters(function.Type.Params)
		if err != nil {
			return nil, fmt.Errorf("player.Player.%s: %w", function.Name.Name, err)
		}
		found[function.Name.Name] = method{Name: function.Name.Name, Parameters: parameters}
	}
	methods := make([]method, 0, len(selectedPlayerTextMethods))
	for _, name := range selectedPlayerTextMethods {
		definition, ok := found[name]
		if !ok {
			return nil, fmt.Errorf("Dragonfly player.Player has no supported %s method", name)
		}
		methods = append(methods, definition)
	}
	return methods, nil
}

func selectedPlayerTextMethod(name string) bool {
	for _, selected := range selectedPlayerTextMethods {
		if name == selected {
			return true
		}
	}
	return false
}

func playerMethod(function *ast.FuncDecl) bool {
	if function.Recv == nil || len(function.Recv.List) != 1 {
		return false
	}
	pointer, ok := function.Recv.List[0].Type.(*ast.StarExpr)
	if !ok {
		return false
	}
	receiver, ok := pointer.X.(*ast.Ident)
	return ok && receiver.Name == "Player"
}

func inspectCube(directory string) (cubeSpec, error) {
	if err := requireArrayType(filepath.Join(directory, "pos.go"), "Pos", 3, "int"); err != nil {
		return cubeSpec{}, err
	}
	if err := requireArrayType(filepath.Join(directory, "range.go"), "Range", 2, "int"); err != nil {
		return cubeSpec{}, err
	}
	file, err := parser.ParseFile(token.NewFileSet(), filepath.Join(directory, "face.go"), nil, 0)
	if err != nil {
		return cubeSpec{}, err
	}
	if !hasNamedType(file, "Face", "int") {
		return cubeSpec{}, fmt.Errorf("cube.Face is not backed by int")
	}
	var faces []string
	for _, declaration := range file.Decls {
		gen, ok := declaration.(*ast.GenDecl)
		if !ok || gen.Tok != token.CONST {
			continue
		}
		for _, raw := range gen.Specs {
			spec, ok := raw.(*ast.ValueSpec)
			if !ok {
				continue
			}
			for _, name := range spec.Names {
				if strings.HasPrefix(name.Name, "Face") {
					faces = append(faces, strings.TrimPrefix(name.Name, "Face"))
				}
			}
		}
	}
	want := []string{"Down", "Up", "North", "South", "West", "East"}
	if !reflect.DeepEqual(faces, want) {
		return cubeSpec{}, fmt.Errorf("cube.Face values changed: %v", faces)
	}
	return cubeSpec{Faces: faces}, nil
}

func requireArrayType(path, name string, length int64, element string) error {
	file, err := parser.ParseFile(token.NewFileSet(), path, nil, 0)
	if err != nil {
		return err
	}
	for _, declaration := range file.Decls {
		gen, ok := declaration.(*ast.GenDecl)
		if !ok {
			continue
		}
		for _, raw := range gen.Specs {
			spec, ok := raw.(*ast.TypeSpec)
			if !ok || spec.Name.Name != name {
				continue
			}
			array, ok := spec.Type.(*ast.ArrayType)
			if !ok || array.Len == nil {
				return fmt.Errorf("cube.%s is not a fixed array", name)
			}
			literal, ok := array.Len.(*ast.BasicLit)
			if !ok || literal.Kind != token.INT || literal.Value != strconv.FormatInt(length, 10) {
				return fmt.Errorf("cube.%s length changed", name)
			}
			identifier, ok := array.Elt.(*ast.Ident)
			if !ok || identifier.Name != element {
				return fmt.Errorf("cube.%s element changed", name)
			}
			return nil
		}
	}
	return fmt.Errorf("cube.%s not found", name)
}

func hasNamedType(file *ast.File, name, underlying string) bool {
	for _, declaration := range file.Decls {
		gen, ok := declaration.(*ast.GenDecl)
		if !ok {
			continue
		}
		for _, raw := range gen.Specs {
			spec, ok := raw.(*ast.TypeSpec)
			if !ok || spec.Name.Name != name {
				continue
			}
			identifier, ok := spec.Type.(*ast.Ident)
			return ok && identifier.Name == underlying
		}
	}
	return false
}

func inspectSetOpts(path string) ([]string, error) {
	file, err := parser.ParseFile(token.NewFileSet(), path, nil, 0)
	if err != nil {
		return nil, err
	}
	for _, declaration := range file.Decls {
		gen, ok := declaration.(*ast.GenDecl)
		if !ok {
			continue
		}
		for _, raw := range gen.Specs {
			spec, ok := raw.(*ast.TypeSpec)
			if !ok || spec.Name.Name != "SetOpts" {
				continue
			}
			structure, ok := spec.Type.(*ast.StructType)
			if !ok {
				return nil, fmt.Errorf("world.SetOpts is not a struct")
			}
			var fields []string
			for _, field := range structure.Fields.List {
				identifier, ok := field.Type.(*ast.Ident)
				if !ok || identifier.Name != "bool" {
					return nil, fmt.Errorf("world.SetOpts contains a non-bool field")
				}
				for _, name := range field.Names {
					if name.IsExported() {
						fields = append(fields, name.Name)
					}
				}
			}
			if len(fields) == 0 {
				return nil, fmt.Errorf("world.SetOpts has no exported fields")
			}
			return fields, nil
		}
	}
	return nil, fmt.Errorf("world.SetOpts not found")
}

func inspectWorldTx(path string) ([]commandMethod, error) {
	file, err := parser.ParseFile(token.NewFileSet(), path, nil, 0)
	if err != nil {
		return nil, err
	}
	found := map[string]commandMethod{}
	for _, declaration := range file.Decls {
		function, ok := declaration.(*ast.FuncDecl)
		if !ok || !selectedWorldTxMethod(function.Name.Name) || !pointerReceiver(function, "Tx") {
			continue
		}
		parameters, err := translateWorldTxParameters(function.Type.Params)
		if err != nil {
			return nil, fmt.Errorf("world.Tx.%s: %w", function.Name.Name, err)
		}
		result, err := translateWorldTxResult(function.Name.Name, function.Type.Results)
		if err != nil {
			return nil, fmt.Errorf("world.Tx.%s: %w", function.Name.Name, err)
		}
		found[function.Name.Name] = commandMethod{
			Name: function.Name.Name, Parameters: parameters, ReturnType: result,
		}
	}
	methods := make([]commandMethod, 0, len(selectedWorldTxMethods))
	for _, name := range selectedWorldTxMethods {
		definition, ok := found[name]
		if !ok {
			return nil, fmt.Errorf("Dragonfly world.Tx has no supported %s method", name)
		}
		methods = append(methods, definition)
	}
	return methods, nil
}

func selectedWorldTxMethod(name string) bool {
	for _, selected := range selectedWorldTxMethods {
		if name == selected {
			return true
		}
	}
	return false
}

func pointerReceiver(function *ast.FuncDecl, name string) bool {
	if function.Recv == nil || len(function.Recv.List) != 1 {
		return false
	}
	pointer, ok := function.Recv.List[0].Type.(*ast.StarExpr)
	if !ok {
		return false
	}
	receiver, ok := pointer.X.(*ast.Ident)
	return ok && receiver.Name == name
}

func translateWorldTxParameters(fields *ast.FieldList) ([]parameter, error) {
	var parameters []parameter
	if fields == nil {
		return parameters, nil
	}
	for _, field := range fields.List {
		typeName, ok := worldTxCSharpType(field.Type, true)
		if !ok {
			return nil, fmt.Errorf("unsupported parameter type %s", formatGoExpression(field.Type))
		}
		if len(field.Names) == 0 {
			return nil, fmt.Errorf("unnamed parameter of type %s", formatGoExpression(field.Type))
		}
		for _, name := range field.Names {
			parameters = append(parameters, parameter{Name: name.Name, Type: typeName})
		}
	}
	return parameters, nil
}

func translateWorldTxResult(method string, fields *ast.FieldList) (string, error) {
	if fields == nil || len(fields.List) == 0 {
		return "void", nil
	}
	var results []string
	for _, field := range fields.List {
		typeName, ok := worldTxCSharpType(field.Type, false)
		if !ok {
			return "", fmt.Errorf("unsupported return type %s", formatGoExpression(field.Type))
		}
		count := len(field.Names)
		if count == 0 {
			count = 1
		}
		for range count {
			results = append(results, typeName)
		}
	}
	if len(results) == 1 {
		return results[0], nil
	}
	if method == "BlockLoaded" {
		if !reflect.DeepEqual(results, []string{"Block", "bool"}) {
			return "", fmt.Errorf("expected (Block, bool), got (%s)", strings.Join(results, ", "))
		}
		return "(Block? Block, bool Ok)", nil
	}
	return "(" + strings.Join(results, ", ") + ")", nil
}

func worldTxCSharpType(expression ast.Expr, parameter bool) (string, bool) {
	switch value := expression.(type) {
	case *ast.StarExpr:
		typeName, ok := worldTxCSharpType(value.X, parameter)
		if !ok {
			return "", false
		}
		return strings.TrimSuffix(typeName, "?") + "?", true
	case *ast.Ident:
		typeName, ok := map[string]string{
			"Block":   "Block",
			"SetOpts": "SetOpts",
			"bool":    "bool",
		}[value.Name]
		if !ok {
			return "", false
		}
		if parameter && value.Name == "Block" {
			return typeName + "?", true
		}
		return typeName, true
	case *ast.SelectorExpr:
		packageName, ok := value.X.(*ast.Ident)
		if !ok {
			return "", false
		}
		typeName, ok := map[string]string{
			"cube.Pos":   "Cube.Pos",
			"cube.Range": "Cube.Range",
		}[packageName.Name+"."+value.Sel.Name]
		return typeName, ok
	default:
		return "", false
	}
}

func inspectBlocks(directory string) (blockSpec, error) {
	packages, err := parser.ParseDir(token.NewFileSet(), directory, func(info os.FileInfo) bool {
		return !strings.HasSuffix(info.Name(), "_test.go")
	}, 0)
	if err != nil {
		return blockSpec{}, err
	}
	pkg, ok := packages["block"]
	if !ok {
		return blockSpec{}, fmt.Errorf("Dragonfly block package not found")
	}
	declarations := map[string]*ast.TypeSpec{}
	for _, file := range pkg.Files {
		for _, declaration := range file.Decls {
			gen, ok := declaration.(*ast.GenDecl)
			if !ok {
				continue
			}
			for _, raw := range gen.Specs {
				typeSpec, ok := raw.(*ast.TypeSpec)
				if ok && typeSpec.Name.IsExported() {
					declarations[typeSpec.Name.Name] = typeSpec
				}
			}
		}
	}

	world.DefaultBlockRegistry.Finalize()
	registered := map[reflect.Type][]world.Block{}
	for _, value := range world.DefaultBlockRegistry.Blocks() {
		typeOf := reflect.TypeOf(value)
		if typeOf == nil {
			continue
		}
		if typeOf.Kind() == reflect.Pointer {
			typeOf = typeOf.Elem()
		}
		if typeOf.PkgPath() == "github.com/df-mc/dragonfly/server/block" && typeOf.Name() != "" {
			registered[typeOf] = append(registered[typeOf], value)
		}
	}

	var result blockSpec
	for typeOf, states := range registered {
		if declarations[typeOf.Name()] == nil {
			return blockSpec{}, fmt.Errorf("registered block %s has no exported AST declaration", typeOf.Name())
		}
		exported := 0
		for index := 0; index < typeOf.NumField(); index++ {
			if typeOf.Field(index).IsExported() {
				exported++
			}
		}
		if exported != 0 {
			continue
		}
		zero, ok := reflect.Zero(typeOf).Interface().(world.Block)
		if !ok {
			continue
		}
		encoded, err := encodeRegisteredBlock(typeOf.Name(), zero, states)
		if err != nil {
			return blockSpec{}, err
		}
		result.Stateless = append(result.Stateless, encoded)
	}
	sort.Slice(result.Stateless, func(i, j int) bool { return result.Stateless[i].Name < result.Stateless[j].Name })
	if len(result.Stateless) == 0 {
		return blockSpec{}, fmt.Errorf("no stateless Dragonfly blocks found")
	}

	var sandType reflect.Type
	var sandStates []world.Block
	for typeOf, states := range registered {
		if typeOf.Name() == "Sand" {
			sandType, sandStates = typeOf, states
			break
		}
	}
	if sandType == nil || !sandHasRedBool(declarations["Sand"]) {
		return blockSpec{}, fmt.Errorf("Dragonfly block.Sand.Red bool field not found")
	}
	for index, red := range []bool{false, true} {
		value := reflect.New(sandType).Elem()
		value.FieldByName("Red").SetBool(red)
		block := value.Interface().(world.Block)
		encoded, err := encodeRegisteredBlock("Sand", block, sandStates)
		if err != nil {
			return blockSpec{}, err
		}
		result.Sand[index] = encoded
	}
	return result, nil
}

func sandHasRedBool(spec *ast.TypeSpec) bool {
	if spec == nil {
		return false
	}
	structure, ok := spec.Type.(*ast.StructType)
	if !ok {
		return false
	}
	for _, field := range structure.Fields.List {
		identifier, ok := field.Type.(*ast.Ident)
		if !ok || identifier.Name != "bool" {
			continue
		}
		for _, name := range field.Names {
			if name.Name == "Red" {
				return true
			}
		}
	}
	return false
}

func encodeRegisteredBlock(typeName string, value world.Block, registered []world.Block) (encodedBlock, error) {
	identifier, properties := value.EncodeBlock()
	found := false
	for _, candidate := range registered {
		candidateIdentifier, candidateProperties := candidate.EncodeBlock()
		if identifier == candidateIdentifier && reflect.DeepEqual(properties, candidateProperties) {
			found = true
			break
		}
	}
	if !found {
		return encodedBlock{}, fmt.Errorf("block.%s zero state is not registered", typeName)
	}
	encoded, ok := host.EncodeBlockProperties(properties)
	if !ok {
		return encodedBlock{}, fmt.Errorf("block.%s has unsupported properties", typeName)
	}
	return encodedBlock{Name: typeName, Identifier: identifier, PropertiesNBT: encoded}, nil
}

func syncGeneratedFiles(files []generatedFile, check bool) error {
	for _, file := range files {
		if check {
			current, err := os.ReadFile(file.Path)
			if err != nil || !bytes.Equal(current, file.Content) {
				return fmt.Errorf("%s is stale; run make generate", file.Path)
			}
			continue
		}
		if err := os.MkdirAll(filepath.Dir(file.Path), 0o755); err != nil {
			return err
		}
		if err := os.WriteFile(file.Path, file.Content, 0o644); err != nil {
			return err
		}
	}
	return nil
}

func playerHandlerMethods(path string) ([]method, error) {
	file, err := parser.ParseFile(token.NewFileSet(), path, nil, 0)
	if err != nil {
		return nil, err
	}
	for _, declaration := range file.Decls {
		gen, ok := declaration.(*ast.GenDecl)
		if !ok {
			continue
		}
		for _, spec := range gen.Specs {
			typeSpec, ok := spec.(*ast.TypeSpec)
			if !ok || typeSpec.Name.Name != "Handler" {
				continue
			}
			interfaceType, ok := typeSpec.Type.(*ast.InterfaceType)
			if !ok {
				return nil, fmt.Errorf("player.Handler is not an interface")
			}
			var methods []method
			for _, field := range interfaceType.Methods.List {
				if len(field.Names) != 1 {
					continue
				}
				subscription, supported := supportedPlayerHandlers[field.Names[0].Name]
				if !supported {
					continue
				}
				function, ok := field.Type.(*ast.FuncType)
				if !ok {
					return nil, fmt.Errorf("player.Handler.%s is not a method", field.Names[0].Name)
				}
				translated, err := translateParameters(function.Params)
				if err != nil {
					return nil, fmt.Errorf("player.Handler.%s: %w", field.Names[0].Name, err)
				}
				methods = append(methods, method{Name: field.Names[0].Name, Parameters: translated, Subscription: subscription})
			}
			for name := range supportedPlayerHandlers {
				found := false
				for _, method := range methods {
					found = found || method.Name == name
				}
				if !found {
					return nil, fmt.Errorf("Dragonfly player.Handler has no supported %s method", name)
				}
			}
			return methods, nil
		}
	}
	return nil, fmt.Errorf("Dragonfly player.Handler interface not found")
}

func translateParameters(fields *ast.FieldList) ([]parameter, error) {
	var parameters []parameter
	for _, field := range fields.List {
		typeName, ok := csharpType(field.Type)
		if !ok {
			return nil, fmt.Errorf("unsupported parameter type %T", field.Type)
		}
		for _, name := range field.Names {
			parameters = append(parameters, parameter{Name: name.Name, Type: typeName})
		}
	}
	return parameters, nil
}

func csharpType(expression ast.Expr) (string, bool) {
	switch value := expression.(type) {
	case *ast.Ellipsis:
		typeName, ok := csharpType(value.Elt)
		if !ok {
			return "", false
		}
		return "params " + typeName + "[]", true
	case *ast.StarExpr:
		typeName, ok := csharpType(value.X)
		if !ok {
			return "", false
		}
		if typeName == "string" || typeName == "int" || typeName == "double" || typeName == "bool" {
			return "ref " + typeName, true
		}
		return typeName, true
	case *ast.Ident:
		typeName, ok := map[string]string{
			"any":     "object?",
			"bool":    "bool",
			"Context": "Player.Context",
			"int":     "int",
			"Player":  "Player",
			"string":  "string",
		}[value.Name]
		return typeName, ok
	case *ast.SelectorExpr:
		packageName, ok := value.X.(*ast.Ident)
		if !ok {
			return "", false
		}
		typeName, ok := map[string]string{
			"mgl64.Vec3":    "Vector3",
			"cube.Rotation": "Rotation",
		}[packageName.Name+"."+value.Sel.Name]
		return typeName, ok
	default:
		return "", false
	}
}

func commandInterfaces(directory string) ([]commandInterface, error) {
	packages, err := parser.ParseDir(token.NewFileSet(), directory, nil, 0)
	if err != nil {
		return nil, err
	}
	pkg, ok := packages["cmd"]
	if !ok {
		return nil, fmt.Errorf("Dragonfly cmd package not found")
	}
	found := map[string]commandInterface{}
	for _, file := range pkg.Files {
		for _, declaration := range file.Decls {
			gen, ok := declaration.(*ast.GenDecl)
			if !ok {
				continue
			}
			for _, spec := range gen.Specs {
				typeSpec, ok := spec.(*ast.TypeSpec)
				if !ok || !selectedCommandInterface(typeSpec.Name.Name) {
					continue
				}
				interfaceType, ok := typeSpec.Type.(*ast.InterfaceType)
				if !ok {
					return nil, fmt.Errorf("cmd.%s is not an interface", typeSpec.Name.Name)
				}
				translated, err := translateCommandInterface(typeSpec.Name.Name, interfaceType)
				if err != nil {
					return nil, err
				}
				found[typeSpec.Name.Name] = translated
			}
		}
	}
	interfaces := make([]commandInterface, 0, len(selectedCommandInterfaces))
	for _, name := range selectedCommandInterfaces {
		definition, ok := found[name]
		if !ok {
			return nil, fmt.Errorf("Dragonfly cmd.%s interface not found", name)
		}
		interfaces = append(interfaces, definition)
	}
	return interfaces, nil
}

func selectedCommandInterface(name string) bool {
	for _, selected := range selectedCommandInterfaces {
		if name == selected {
			return true
		}
	}
	return false
}

func translateCommandInterface(name string, interfaceType *ast.InterfaceType) (commandInterface, error) {
	definition := commandInterface{Name: name}
	for _, field := range interfaceType.Methods.List {
		if len(field.Names) == 0 {
			embedding, ok := field.Type.(*ast.Ident)
			if !ok || !selectedCommandInterface(embedding.Name) {
				return commandInterface{}, fmt.Errorf("cmd.%s has unsupported embedded interface", name)
			}
			definition.Embeddings = append(definition.Embeddings, embedding.Name)
			continue
		}
		if len(field.Names) != 1 {
			return commandInterface{}, fmt.Errorf("cmd.%s has unnamed method", name)
		}
		function, ok := field.Type.(*ast.FuncType)
		if !ok {
			return commandInterface{}, fmt.Errorf("cmd.%s.%s is not a method", name, field.Names[0].Name)
		}
		parameters, err := translateCommandParameters(function.Params)
		if err != nil {
			return commandInterface{}, fmt.Errorf("cmd.%s.%s: %w", name, field.Names[0].Name, err)
		}
		returnType, err := translateCommandResult(function.Results)
		if err != nil {
			return commandInterface{}, fmt.Errorf("cmd.%s.%s: %w", name, field.Names[0].Name, err)
		}
		definition.Methods = append(definition.Methods, commandMethod{
			Name:       field.Names[0].Name,
			Parameters: parameters,
			ReturnType: returnType,
		})
	}
	return definition, nil
}

func translateCommandParameters(fields *ast.FieldList) ([]parameter, error) {
	var parameters []parameter
	if fields == nil {
		return parameters, nil
	}
	for _, field := range fields.List {
		typeName, ok := commandCSharpType(field.Type)
		if !ok {
			return nil, fmt.Errorf("unsupported parameter type %s", formatGoExpression(field.Type))
		}
		for _, name := range field.Names {
			parameters = append(parameters, parameter{Name: name.Name, Type: typeName})
		}
	}
	return parameters, nil
}

func translateCommandResult(fields *ast.FieldList) (string, error) {
	if fields == nil || len(fields.List) == 0 {
		return "void", nil
	}
	if len(fields.List) != 1 || len(fields.List[0].Names) > 1 {
		return "", fmt.Errorf("multiple return values are unsupported")
	}
	typeName, ok := commandCSharpType(fields.List[0].Type)
	if !ok {
		return "", fmt.Errorf("unsupported return type %s", formatGoExpression(fields.List[0].Type))
	}
	return typeName, nil
}

func commandCSharpType(expression ast.Expr) (string, bool) {
	switch value := expression.(type) {
	case *ast.StarExpr:
		typeName, ok := commandCSharpType(value.X)
		if !ok {
			return "", false
		}
		if typeName == "World.Tx" {
			return typeName + "?", true
		}
		return typeName, true
	case *ast.ArrayType:
		if value.Len != nil {
			return "", false
		}
		element, ok := commandCSharpType(value.Elt)
		if !ok {
			return "", false
		}
		return "IReadOnlyList<" + element + ">", true
	case *ast.Ident:
		typeName, ok := map[string]string{
			"bool":   "bool",
			"Output": "Output",
			"Source": "Source",
			"string": "string",
		}[value.Name]
		return typeName, ok
	case *ast.SelectorExpr:
		packageName, ok := value.X.(*ast.Ident)
		if !ok {
			return "", false
		}
		typeName, ok := map[string]string{
			"mgl64.Vec3": "Vector3",
			"world.Tx":   "World.Tx",
		}[packageName.Name+"."+value.Sel.Name]
		return typeName, ok
	default:
		return "", false
	}
}

func formatGoExpression(expression ast.Expr) string {
	switch value := expression.(type) {
	case *ast.Ident:
		return value.Name
	case *ast.SelectorExpr:
		return formatGoExpression(value.X) + "." + value.Sel.Name
	case *ast.StarExpr:
		return "*" + formatGoExpression(value.X)
	case *ast.ArrayType:
		return "[]" + formatGoExpression(value.Elt)
	default:
		return fmt.Sprintf("%T", expression)
	}
}

func generatePlayerHandler(methods []method) []byte {
	var output bytes.Buffer
	output.WriteString("// Code generated from Dragonfly server/player/handler.go. DO NOT EDIT.\n")
	output.WriteString("namespace Dragonfly;\n\n")
	output.WriteString("public sealed partial class Player\n{\n    public interface Handler\n    {\n")
	for _, method := range methods {
		fmt.Fprintf(&output, "        void %s(%s);\n", method.Name, formatParameters(method.Parameters))
	}
	output.WriteString("    }\n}\n\n")
	output.WriteString("public abstract partial class Plugin\n{\n")
	for _, method := range methods {
		fmt.Fprintf(&output, "    [HandlerSubscription(%dUL)]\n", method.Subscription)
		fmt.Fprintf(&output, "    public virtual void %s(%s) { }\n", method.Name, formatParameters(method.Parameters))
	}
	output.WriteString("}\n")
	return output.Bytes()
}

func generatePlayerTextMethods(methods []method) []byte {
	var output bytes.Buffer
	output.WriteString("// Code generated from Dragonfly server/player/player.go. DO NOT EDIT.\n")
	output.WriteString("#nullable enable\n")
	output.WriteString("using Dragonfly.Native;\n\n")
	output.WriteString("namespace Dragonfly;\n\n")
	output.WriteString("public sealed partial class Player\n{\n")
	for _, method := range methods {
		fmt.Fprintf(&output, "    public void %s(%s) => ", method.Name, formatParameters(method.Parameters))
		switch method.Name {
		case "Message", "SendPopup", "SendTip", "SendJukeboxPopup", "Disconnect":
			kind := map[string]string{
				"Message":          "Message",
				"SendPopup":        "Popup",
				"SendTip":          "Tip",
				"SendJukeboxPopup": "JukeboxPopup",
				"Disconnect":       "Disconnect",
			}[method.Name]
			fmt.Fprintf(&output, "SendText(Abi.PlayerText%s, FormatArguments(%s));\n", kind, method.Parameters[0].Name)
		case "SetNameTag":
			fmt.Fprintf(&output, "SendText(Abi.PlayerTextNameTag, %s);\n", method.Parameters[0].Name)
		default:
			panic("unsupported player text method: " + method.Name)
		}
	}
	output.WriteString("}\n")
	return output.Bytes()
}

func generateCommandInterfaces(interfaces []commandInterface) []byte {
	var output bytes.Buffer
	output.WriteString("// Code generated from Dragonfly server/cmd Go AST. DO NOT EDIT.\n#nullable enable\n")
	output.WriteString("namespace Dragonfly;\n\n")
	output.WriteString("public static partial class Cmd\n{\n")
	for index, definition := range interfaces {
		fmt.Fprintf(&output, "    public interface %s", definition.Name)
		if len(definition.Embeddings) != 0 {
			fmt.Fprintf(&output, " : %s", strings.Join(definition.Embeddings, ", "))
		}
		output.WriteString("\n    {\n")
		for _, method := range definition.Methods {
			fmt.Fprintf(&output, "        %s %s(%s);\n", method.ReturnType, method.Name, formatParameters(method.Parameters))
		}
		output.WriteString("    }\n")
		if index != len(interfaces)-1 {
			output.WriteString("\n")
		}
	}
	output.WriteString("}\n")
	return output.Bytes()
}

func generateCube(spec cubeSpec) []byte {
	var output bytes.Buffer
	output.WriteString("// Code generated from Dragonfly server/block/cube Go AST. DO NOT EDIT.\n")
	output.WriteString("namespace Dragonfly;\n\n")
	output.WriteString("public static partial class Cube\n{\n")
	output.WriteString("    public enum Face\n    {\n")
	for index, face := range spec.Faces {
		fmt.Fprintf(&output, "        %s = %d,\n", face, index)
	}
	output.WriteString("    }\n\n")
	output.WriteString(`    public readonly record struct Range(int Minimum, int Maximum)
    {
        public int Min() => Minimum;
        public int Max() => Maximum;
        public int Height() => Maximum - Minimum;
    }

    public readonly record struct Pos
    {
        private readonly int _x;
        private readonly int _y;
        private readonly int _z;

        public Pos(int x, int y, int z) => (_x, _y, _z) = (x, y, z);

        public int X() => _x;
        public int Y() => _y;
        public int Z() => _z;
        public bool OutOfBounds(Range range) => _y > range.Max() || _y < range.Min();
        public bool Within(Pos min, Pos max) =>
            _x >= min._x && _x <= max._x &&
            _y >= min._y && _y <= max._y &&
            _z >= min._z && _z <= max._z;
        public Pos Add(Pos other) => new(_x + other._x, _y + other._y, _z + other._z);
        public Pos Sub(Pos other) => new(_x - other._x, _y - other._y, _z - other._z);
        public Vector3 Vec3() => new(_x, _y, _z);
        public Vector3 Vec3Middle() => new(_x + 0.5, _y, _z + 0.5);
        public Vector3 Vec3Centre() => new(_x + 0.5, _y + 0.5, _z + 0.5);

        public Pos Side(Face face) => face switch
        {
            Cube.Face.Up => new(_x, _y + 1, _z),
            Cube.Face.Down => new(_x, _y - 1, _z),
            Cube.Face.North => new(_x, _y, _z - 1),
            Cube.Face.South => new(_x, _y, _z + 1),
            Cube.Face.West => new(_x - 1, _y, _z),
            Cube.Face.East => new(_x + 1, _y, _z),
            _ => this,
        };

        public Face Face(Pos other) => NeighbourFace(other).Face;

        public (Face Face, bool Ok) NeighbourFace(Pos other) => other.Sub(this) switch
        {
            Pos { _x: 0, _y: 1, _z: 0 } => (Cube.Face.Up, true),
            Pos { _x: 0, _y: -1, _z: 0 } => (Cube.Face.Down, true),
            Pos { _x: 0, _y: 0, _z: -1 } => (Cube.Face.North, true),
            Pos { _x: 0, _y: 0, _z: 1 } => (Cube.Face.South, true),
            Pos { _x: -1, _y: 0, _z: 0 } => (Cube.Face.West, true),
            Pos { _x: 1, _y: 0, _z: 0 } => (Cube.Face.East, true),
            _ => (Cube.Face.Up, false),
        };

        public override string ToString() => $"({_x},{_y},{_z})";
    }

    public static Pos PosFromVec3(Vector3 value) => new(
        checked((int)Math.Floor(value.X)),
        checked((int)Math.Floor(value.Y)),
        checked((int)Math.Floor(value.Z)));

    public static Pos Min(Pos first, Pos second) => new(
        Math.Min(first.X(), second.X()),
        Math.Min(first.Y(), second.Y()),
        Math.Min(first.Z(), second.Z()));

    public static Pos Max(Pos first, Pos second) => new(
        Math.Max(first.X(), second.X()),
        Math.Max(first.Y(), second.Y()),
        Math.Max(first.Z(), second.Z()));
}
`)
	return output.Bytes()
}

func generateWorldBlock(setOpts []string, methods []commandMethod) []byte {
	var output bytes.Buffer
	output.WriteString("// Code generated from Dragonfly server/world Go AST. DO NOT EDIT.\n")
	output.WriteString("#nullable enable\n\n")
	output.WriteString("namespace Dragonfly;\n\n")
	output.WriteString("public sealed partial class World\n{\n")
	output.WriteString("    public interface Block { }\n\n")
	output.WriteString("    public sealed class SetOpts\n    {\n")
	for _, field := range setOpts {
		fmt.Fprintf(&output, "        public bool %s;\n", field)
	}
	output.WriteString("    }\n\n")
	output.WriteString("    public partial class Tx\n    {\n")
	for index, method := range methods {
		parameters := formatParameters(method.Parameters)
		if method.Name == "SetBlock" && len(method.Parameters) != 0 {
			last := method.Parameters[len(method.Parameters)-1]
			if last.Name != "opts" || last.Type != "SetOpts?" {
				panic("world.Tx.SetBlock final parameter is not opts *SetOpts")
			}
			parameters = strings.TrimSuffix(parameters, last.Type+" "+last.Name) + last.Type + " " + last.Name + " = null"
		}
		fmt.Fprintf(&output, "        public %s %s(%s) =>\n", method.ReturnType, method.Name, parameters)
		switch method.Name {
		case "Range":
			output.WriteString("            PluginBridge.Host.WorldRange(Invocation);\n")
		case "SetBlock":
			fmt.Fprintf(&output, "            PluginBridge.Host.SetWorldBlock(Invocation, %s, %s, %s);\n",
				method.Parameters[0].Name, method.Parameters[1].Name, method.Parameters[2].Name)
		case "Block":
			fmt.Fprintf(&output, "            PluginBridge.Host.WorldBlock(Invocation, %s);\n", method.Parameters[0].Name)
		case "BlockLoaded":
			fmt.Fprintf(&output, "            PluginBridge.Host.WorldBlockLoaded(Invocation, %s);\n", method.Parameters[0].Name)
		default:
			panic("unsupported world.Tx method: " + method.Name)
		}
		if index != len(methods)-1 {
			output.WriteByte('\n')
		}
	}
	output.WriteString("    }\n}\n")
	return output.Bytes()
}

func generateBlocks(spec blockSpec) []byte {
	var output bytes.Buffer
	output.WriteString("// Code generated from Dragonfly server/block Go AST and registry. DO NOT EDIT.\n")
	output.WriteString("#nullable enable\nusing System;\nusing Dragonfly;\n\n")
	output.WriteString("namespace Dragonfly\n{\n    public static partial class Block\n    {\n")
	for _, block := range spec.Stateless {
		fmt.Fprintf(&output, "        public readonly record struct %s : World.Block;\n", block.Name)
	}
	output.WriteString("        public readonly record struct Sand(bool Red = false) : World.Block;\n")
	output.WriteString("    }\n\n")
	output.WriteString("    internal static class BlockCodec\n    {\n")
	states := append([]encodedBlock(nil), spec.Stateless...)
	states = append(states, spec.Sand[:]...)
	for index, state := range states {
		fmt.Fprintf(&output, "        private static readonly byte[] State%d = %s;\n", index, csharpBytes(state.PropertiesNBT))
	}
	output.WriteString("\n        internal static bool TryEncode(World.Block block, out string identifier, out byte[] properties)\n        {\n")
	output.WriteString("            switch (block)\n            {\n")
	for index, block := range spec.Stateless {
		fmt.Fprintf(&output, "                case Block.%s _:\n                    identifier = %s; properties = State%d; return true;\n", block.Name, strconv.Quote(block.Identifier), index)
	}
	sandOffset := len(spec.Stateless)
	fmt.Fprintf(&output, "                case Block.Sand { Red: true }:\n                    identifier = %s; properties = State%d; return true;\n", strconv.Quote(spec.Sand[1].Identifier), sandOffset+1)
	fmt.Fprintf(&output, "                case Block.Sand:\n                    identifier = %s; properties = State%d; return true;\n", strconv.Quote(spec.Sand[0].Identifier), sandOffset)
	output.WriteString("                case EncodedBlock encoded:\n                    identifier = encoded.Identifier; properties = encoded.Properties; return true;\n")
	output.WriteString("                default:\n                    identifier = string.Empty; properties = Array.Empty<byte>(); return false;\n            }\n        }\n\n")
	output.WriteString("        internal static World.Block Decode(string identifier, ReadOnlySpan<byte> properties)\n        {\n")
	for index, block := range spec.Stateless {
		fmt.Fprintf(&output, "            if (identifier == %s && properties.SequenceEqual(State%d)) return new Block.%s();\n", strconv.Quote(block.Identifier), index, block.Name)
	}
	fmt.Fprintf(&output, "            if (identifier == %s && properties.SequenceEqual(State%d)) return new Block.Sand();\n", strconv.Quote(spec.Sand[0].Identifier), sandOffset)
	fmt.Fprintf(&output, "            if (identifier == %s && properties.SequenceEqual(State%d)) return new Block.Sand(true);\n", strconv.Quote(spec.Sand[1].Identifier), sandOffset+1)
	output.WriteString("            return new EncodedBlock(identifier, properties.ToArray());\n        }\n\n")
	output.WriteString("        private sealed record EncodedBlock(string Identifier, byte[] Properties) : World.Block;\n")
	output.WriteString("    }\n}\n")
	return output.Bytes()
}

func csharpBytes(value []byte) string {
	if len(value) == 0 {
		return "Array.Empty<byte>()"
	}
	parts := make([]string, len(value))
	for index, current := range value {
		parts[index] = fmt.Sprintf("0x%02x", current)
	}
	return "[" + strings.Join(parts, ", ") + "]"
}

func formatParameters(parameters []parameter) string {
	values := make([]string, len(parameters))
	for index, parameter := range parameters {
		values[index] = parameter.Type + " " + parameter.Name
	}
	return strings.Join(values, ", ")
}

func fatal(err error) {
	fmt.Fprintln(os.Stderr, err)
	os.Exit(1)
}
