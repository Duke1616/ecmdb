package service

import (
	"os"

	"github.com/Duke1616/ecmdb/internal/bootstrap/structure"
	"gopkg.in/yaml.v3"
)

// Parser YAML 配置解析器
type Parser interface {
	// ParseFile 从文件解析配置
	ParseFile(filePath string) (*structure.Config, error)
	// Parse 从字节数据解析配置
	Parse(data []byte) (*structure.Config, error)
}

type parser struct{}

// NewParser 创建解析器
func NewParser() Parser {
	return &parser{}
}

// ParseFile 从文件解析配置
func (p *parser) ParseFile(filePath string) (*structure.Config, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}
	return p.Parse(data)
}

// Parse 从字节数据解析配置
func (p *parser) Parse(data []byte) (*structure.Config, error) {
	var cfg structure.Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}
