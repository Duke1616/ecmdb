package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"go/ast"
	"go/format"
	"go/printer"
	"go/token"
	"os"
	"strconv"
)

// MenuData 表示 JSON 中的菜单数据结构
type MenuData struct {
	ID        string `json:"id"`
	PID       string `json:"pid"`
	Path      string `json:"path"`
	Name      string `json:"name"`
	Sort      string `json:"sort"`
	Component string `json:"component"`
	Redirect  string `json:"redirect"`
	Status    string `json:"status"`
	Type      string `json:"type"`
	Meta      string `json:"meta"`
	Endpoints string `json:"endpoints"`
}

// MetaData 表示 meta 字段的结构
type MetaData struct {
	Title       string   `json:"title"`
	IsHidden    bool     `json:"is_hidden"`
	IsAffix     bool     `json:"is_affix"`
	IsKeepAlive bool     `json:"is_keepalive"`
	Icon        string   `json:"icon"`
	Platforms   []string `json:"platforms"`
}

// EndpointData 表示 endpoints 字段的结构
type EndpointData struct {
	Path     string `json:"path"`
	Method   string `json:"method"`
	Resource string `json:"resource"`
	Desc     string `json:"desc"`
}

func main() {
	if len(os.Args) != 2 {
		fmt.Println("用法: go run ast_generator.go <json_file>")
		os.Exit(1)
	}

	jsonFile := os.Args[1]

	// 读取 JSON 文件
	jsonData, err := os.ReadFile(jsonFile)
	if err != nil {
		fmt.Printf("读取 JSON 文件失败: %v\n", err)
		os.Exit(1)
	}

	// 解析 JSON 数据
	var menus []MenuData
	if err = json.Unmarshal(jsonData, &menus); err != nil {
		fmt.Printf("解析 JSON 数据失败: %v\n", err)
		os.Exit(1)
	}

	// 生成 AST
	file := generateMenuFile(menus)

	// 格式化并输出代码到 cmd/initial/menu/ 目录
	outputPath := "../../initial/menu/menu_data.go"
	if err = formatAndWrite(file, outputPath); err != nil {
		fmt.Printf("生成代码失败: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("✅ 菜单代码生成成功!")
	fmt.Println("📁 生成文件: cmd/initial/menu/menu_data.go")
}

// generateMenuFile 生成菜单文件的 AST
func generateMenuFile(menus []MenuData) *ast.File {
	// 创建文件 AST
	file := &ast.File{
		Name:  &ast.Ident{Name: "menu"},
		Decls: []ast.Decl{},
	}

	// 添加导入声明
	imports := &ast.GenDecl{
		Tok: token.IMPORT,
		Specs: []ast.Spec{
			&ast.ImportSpec{
				Path: &ast.BasicLit{Kind: token.STRING, Value: `"github.com/Duke1616/ecmdb/internal/menu"`},
			},
		},
	}
	file.Decls = append(file.Decls, imports)

	// 生成菜单数据变量
	menuVar := generateMenuVariable(menus)
	file.Decls = append(file.Decls, menuVar)

	// 生成获取菜单的函数
	getMenuFunc := generateGetMenuFunction()
	file.Decls = append(file.Decls, getMenuFunc)

	// 生成获取所有菜单IDs的函数
	getMenuIDsFunc := generateGetMenuIDsFunction(menus)
	file.Decls = append(file.Decls, getMenuIDsFunc)

	return file
}

// generateMenuVariable 生成菜单数据变量
func generateMenuVariable(menus []MenuData) ast.Decl {
	var elements []ast.Expr

	for _, menu := range menus {
		menuExpr := generateMenuStruct(menu)
		elements = append(elements, menuExpr)
	}

	return &ast.GenDecl{
		Tok: token.VAR,
		Specs: []ast.Spec{
			&ast.ValueSpec{
				Names: []*ast.Ident{{Name: "DefaultMenus"}},
				Type: &ast.ArrayType{
					Elt: &ast.SelectorExpr{
						X:   &ast.Ident{Name: "menu"},
						Sel: &ast.Ident{Name: "Menu"},
					},
				},
				Values: []ast.Expr{
					&ast.CompositeLit{
						Type: &ast.ArrayType{
							Elt: &ast.SelectorExpr{
								X:   &ast.Ident{Name: "menu"},
								Sel: &ast.Ident{Name: "Menu"},
							},
						},
						Elts: elements,
					},
				},
			},
		},
	}
}

// generateMenuStruct 生成单个菜单结构体
func generateMenuStruct(menu MenuData) ast.Expr {
	// 解析 ID
	id, _ := strconv.ParseInt(menu.ID, 10, 64)
	pid, _ := strconv.ParseInt(menu.PID, 10, 64)
	sort, _ := strconv.ParseInt(menu.Sort, 10, 64)
	status, _ := strconv.ParseInt(menu.Status, 10, 64)
	menuType, _ := strconv.ParseInt(menu.Type, 10, 64)

	// 解析 Meta 数据
	meta := parseMetaData(menu.Meta)

	// 解析 Endpoints 数据
	endpoints := parseEndpointsData(menu.Endpoints)

	return &ast.CompositeLit{
		Type: &ast.SelectorExpr{
			X:   &ast.Ident{Name: "menu"},
			Sel: &ast.Ident{Name: "Menu"},
		},
		Elts: []ast.Expr{
			&ast.KeyValueExpr{
				Key:   &ast.Ident{Name: "Id"},
				Value: &ast.BasicLit{Kind: token.INT, Value: strconv.FormatInt(id, 10)},
			},
			&ast.KeyValueExpr{
				Key:   &ast.Ident{Name: "Pid"},
				Value: &ast.BasicLit{Kind: token.INT, Value: strconv.FormatInt(pid, 10)},
			},
			&ast.KeyValueExpr{
				Key:   &ast.Ident{Name: "Path"},
				Value: &ast.BasicLit{Kind: token.STRING, Value: strconv.Quote(menu.Path)},
			},
			&ast.KeyValueExpr{
				Key:   &ast.Ident{Name: "Name"},
				Value: &ast.BasicLit{Kind: token.STRING, Value: strconv.Quote(menu.Name)},
			},
			&ast.KeyValueExpr{
				Key:   &ast.Ident{Name: "Sort"},
				Value: &ast.BasicLit{Kind: token.INT, Value: strconv.FormatInt(sort, 10)},
			},
			&ast.KeyValueExpr{
				Key:   &ast.Ident{Name: "Component"},
				Value: &ast.BasicLit{Kind: token.STRING, Value: strconv.Quote(menu.Component)},
			},
			&ast.KeyValueExpr{
				Key:   &ast.Ident{Name: "Redirect"},
				Value: &ast.BasicLit{Kind: token.STRING, Value: strconv.Quote(menu.Redirect)},
			},
			&ast.KeyValueExpr{
				Key: &ast.Ident{Name: "Status"},
				Value: &ast.CallExpr{
					Fun: &ast.SelectorExpr{
						X:   &ast.Ident{Name: "menu"},
						Sel: &ast.Ident{Name: "Status"},
					},
					Args: []ast.Expr{
						&ast.BasicLit{Kind: token.INT, Value: strconv.FormatInt(status, 10)},
					},
				},
			},
			&ast.KeyValueExpr{
				Key: &ast.Ident{Name: "Type"},
				Value: &ast.CallExpr{
					Fun: &ast.SelectorExpr{
						X:   &ast.Ident{Name: "menu"},
						Sel: &ast.Ident{Name: "Type"},
					},
					Args: []ast.Expr{
						&ast.BasicLit{Kind: token.INT, Value: strconv.FormatInt(menuType, 10)},
					},
				},
			},
			&ast.KeyValueExpr{
				Key:   &ast.Ident{Name: "Meta"},
				Value: generateMetaStruct(meta),
			},
			&ast.KeyValueExpr{
				Key:   &ast.Ident{Name: "Endpoints"},
				Value: generateEndpointsSlice(endpoints),
			},
		},
	}
}

// generateMetaStruct 生成 Meta 结构体
func generateMetaStruct(meta MetaData) ast.Expr {
	return &ast.CompositeLit{
		Type: &ast.SelectorExpr{
			X:   &ast.Ident{Name: "menu"},
			Sel: &ast.Ident{Name: "Meta"},
		},
		Elts: []ast.Expr{
			&ast.KeyValueExpr{
				Key:   &ast.Ident{Name: "Title"},
				Value: &ast.BasicLit{Kind: token.STRING, Value: strconv.Quote(meta.Title)},
			},
			&ast.KeyValueExpr{
				Key:   &ast.Ident{Name: "IsHidden"},
				Value: &ast.Ident{Name: strconv.FormatBool(meta.IsHidden)},
			},
			&ast.KeyValueExpr{
				Key:   &ast.Ident{Name: "IsAffix"},
				Value: &ast.Ident{Name: strconv.FormatBool(meta.IsAffix)},
			},
			&ast.KeyValueExpr{
				Key:   &ast.Ident{Name: "IsKeepAlive"},
				Value: &ast.Ident{Name: strconv.FormatBool(meta.IsKeepAlive)},
			},
			&ast.KeyValueExpr{
				Key:   &ast.Ident{Name: "Icon"},
				Value: &ast.BasicLit{Kind: token.STRING, Value: strconv.Quote(meta.Icon)},
			},
			&ast.KeyValueExpr{
				Key:   &ast.Ident{Name: "Platforms"},
				Value: generatePlatformsSlice(meta.Platforms),
			},
		},
	}
}

// generatePlatformsSlice 生成 Platforms 切片
func generatePlatformsSlice(platforms []string) ast.Expr {
	if len(platforms) == 0 {
		return &ast.Ident{Name: "nil"}
	}

	var elements []ast.Expr
	for _, platform := range platforms {
		elements = append(elements, &ast.BasicLit{
			Kind:  token.STRING,
			Value: strconv.Quote(platform),
		})
	}

	return &ast.CompositeLit{
		Type: &ast.ArrayType{
			Elt: &ast.Ident{Name: "string"},
		},
		Elts: elements,
	}
}

// generateEndpointsSlice 生成 Endpoints 切片
func generateEndpointsSlice(endpoints []EndpointData) ast.Expr {
	if len(endpoints) == 0 {
		return &ast.Ident{Name: "nil"}
	}

	var elements []ast.Expr
	for _, endpoint := range endpoints {
		endpointExpr := &ast.CompositeLit{
			Type: &ast.SelectorExpr{
				X:   &ast.Ident{Name: "menu"},
				Sel: &ast.Ident{Name: "Endpoint"},
			},
			Elts: []ast.Expr{
				&ast.KeyValueExpr{
					Key:   &ast.Ident{Name: "Path"},
					Value: &ast.BasicLit{Kind: token.STRING, Value: strconv.Quote(endpoint.Path)},
				},
				&ast.KeyValueExpr{
					Key:   &ast.Ident{Name: "Method"},
					Value: &ast.BasicLit{Kind: token.STRING, Value: strconv.Quote(endpoint.Method)},
				},
				&ast.KeyValueExpr{
					Key:   &ast.Ident{Name: "Resource"},
					Value: &ast.BasicLit{Kind: token.STRING, Value: strconv.Quote(endpoint.Resource)},
				},
				&ast.KeyValueExpr{
					Key:   &ast.Ident{Name: "Desc"},
					Value: &ast.BasicLit{Kind: token.STRING, Value: strconv.Quote(endpoint.Desc)},
				},
			},
		}
		elements = append(elements, endpointExpr)
	}

	return &ast.CompositeLit{
		Type: &ast.ArrayType{
			Elt: &ast.SelectorExpr{
				X:   &ast.Ident{Name: "menu"},
				Sel: &ast.Ident{Name: "Endpoint"},
			},
		},
		Elts: elements,
	}
}

// generateGetMenuFunction 生成获取菜单的函数
func generateGetMenuFunction() ast.Decl {
	return &ast.FuncDecl{
		Name: &ast.Ident{Name: "GetInjectMenus"},
		Type: &ast.FuncType{
			Results: &ast.FieldList{
				List: []*ast.Field{
					{
						Type: &ast.ArrayType{
							Elt: &ast.SelectorExpr{
								X:   &ast.Ident{Name: "menu"},
								Sel: &ast.Ident{Name: "Menu"},
							},
						},
					},
				},
			},
		},
		Body: &ast.BlockStmt{
			List: []ast.Stmt{
				&ast.ReturnStmt{
					Results: []ast.Expr{
						&ast.Ident{Name: "DefaultMenus"},
					},
				},
			},
		},
	}
}

// generateGetMenuIDsFunction 生成获取所有菜单IDs的函数
func generateGetMenuIDsFunction(menus []MenuData) ast.Decl {
	// 生成所有菜单ID的切片
	var idElements []ast.Expr
	for _, menu := range menus {
		id, _ := strconv.ParseInt(menu.ID, 10, 64)
		idElements = append(idElements, &ast.BasicLit{Kind: token.INT, Value: strconv.FormatInt(id, 10)})
	}

	return &ast.FuncDecl{
		Name: &ast.Ident{Name: "GetAllMenuIDs"},
		Type: &ast.FuncType{
			Results: &ast.FieldList{
				List: []*ast.Field{
					{
						Type: &ast.ArrayType{
							Elt: &ast.Ident{Name: "int64"},
						},
					},
				},
			},
		},
		Body: &ast.BlockStmt{
			List: []ast.Stmt{
				&ast.ReturnStmt{
					Results: []ast.Expr{
						&ast.CompositeLit{
							Type: &ast.ArrayType{
								Elt: &ast.Ident{Name: "int64"},
							},
							Elts: idElements,
						},
					},
				},
			},
		},
	}
}

// parseMetaData 解析 Meta 数据
func parseMetaData(metaStr string) MetaData {
	var meta MetaData
	if err := json.Unmarshal([]byte(metaStr), &meta); err != nil {
		return MetaData{
			Title:       "",
			IsHidden:    false,
			IsAffix:     false,
			IsKeepAlive: false,
			Icon:        "",
			Platforms:   []string{},
		}
	}
	return meta
}

// parseEndpointsData 解析 Endpoints 数据
func parseEndpointsData(endpointsStr string) []EndpointData {
	if endpointsStr == "[ ]" || endpointsStr == "[]" {
		return []EndpointData{}
	}

	var endpoints []EndpointData
	if err := json.Unmarshal([]byte(endpointsStr), &endpoints); err != nil {
		return []EndpointData{}
	}
	return endpoints
}

// formatAndWrite 格式化并写入文件
func formatAndWrite(file *ast.File, filename string) error {
	fSet := token.NewFileSet()
	var buf bytes.Buffer

	// 先用 printer 打印 AST
	cfg := &printer.Config{Mode: printer.UseSpaces | printer.TabIndent, Tabwidth: 4}
	if err := cfg.Fprint(&buf, fSet, file); err != nil {
		return err
	}

	// 再走 format.Source 重新解析并格式化
	src, err := format.Source(buf.Bytes())
	if err != nil {
		return err
	}

	return os.WriteFile(filename, src, 0644)
}
