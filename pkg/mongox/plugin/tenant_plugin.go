package plugin

import (
	"context"
	"encoding/json"
	"reflect"
	"strconv"
	"strings"
	"sync"

	"github.com/Duke1616/ecmdb/pkg/mongox"
	"github.com/Duke1616/eiam/pkg/ctxutil"
	"go.mongodb.org/mongo-driver/bson"
)

const (
	// IGNORE_TENANT_KEY 用于标识是否跳过租户隔离校验的 Context Key
	IGNORE_TENANT_KEY = "mongox:ignore_tenant"
)

// SharedConfig 共享规则配置
type SharedConfig struct {
	IsShared  bool   // 是否为共享资源，允许跨租户或与系统租户共享
	IsPrivate bool   // 是否为纯私有资源，强制隔离
	IsIgnore  bool   // 是否为忽略模型，免除多租户插件的一切拦截
	Condition bson.M // 共享的附加过滤条件（如 {"is_public": true}）
}

// TenantPlugin 自动在底层无感地隔离多租户数据。
// 完全向 GORM (gormx.TenantPlugin) 核心隔离及提权理念对齐。
type TenantPlugin struct {
	tenantField    string
	systemTenantID int64
	cache          sync.Map // key: reflect.Type -> value: SharedConfig
}

// TenantOption 定义插件配置函数
type TenantOption func(*TenantPlugin)

// WithTenantField 设置自定义的租户字段，默认 "tenant_id"
func WithTenantField(field string) TenantOption {
	return func(p *TenantPlugin) {
		p.tenantField = field
	}
}

// WithSystemTenantID 设置自定义的系统根租户 ID
func WithSystemTenantID(id int64) TenantOption {
	return func(p *TenantPlugin) {
		p.systemTenantID = id
	}
}

// NewTenantPlugin 实例化多租户隔离插件
func NewTenantPlugin(opts ...TenantOption) *TenantPlugin {
	p := &TenantPlugin{
		tenantField:    "tenant_id",
		systemTenantID: ctxutil.SystemTenantID,
	}
	for _, opt := range opts {
		opt(p)
	}
	return p
}

// Name 实现 Plugin 接口
func (p *TenantPlugin) Name() string {
	return "tenant_plugin"
}

// BeforeFind 拦截查询操作，织入智能多租户安全边界
func (p *TenantPlugin) BeforeFind(stmt *mongox.Statement) error {
	if IsIgnoreTenant(stmt.Context) {
		return nil
	}

	conf := p.getSharedConfig(stmt.Model)
	if conf.IsIgnore {
		return nil
	}

	tid := ctxutil.GetTenantID(stmt.Context).Int64()
	if tid == 0 {
		return nil // 无租户上下文，豁免隔离限制，支持后台全局检索
	}

	// 智能判定并构建对应多租户过滤条件
	cond := p.buildFindFilter(stmt.Context, conf, tid)
	if len(cond) > 0 {
		injectConditions(stmt.Filter, cond)
	}
	return nil
}

// BeforeUpdate 拦截更新操作，追加严格租户隔离，杜绝越权
func (p *TenantPlugin) BeforeUpdate(stmt *mongox.Statement) error {
	return p.runWriteBarrier(stmt)
}

// BeforeDelete 拦截删除操作，追加严格租户隔离，杜绝越权
func (p *TenantPlugin) BeforeDelete(stmt *mongox.Statement) error {
	return p.runWriteBarrier(stmt)
}

// runWriteBarrier 统一的变更写屏障校验逻辑，杜绝冗余代码
func (p *TenantPlugin) runWriteBarrier(stmt *mongox.Statement) error {
	if IsIgnoreTenant(stmt.Context) {
		return nil
	}

	conf := p.getSharedConfig(stmt.Model)
	if conf.IsIgnore {
		return nil
	}

	tid := ctxutil.GetTenantID(stmt.Context).Int64()
	if tid == 0 || tid == p.systemTenantID {
		return nil // 超管及后台全局任务免除写屏障校验
	}

	injectConditions(stmt.Filter, bson.M{p.tenantField: tid})
	return nil
}

// buildFindFilter 根据不同的隔离级别策略，扁平化地计算并返回应当注入的租户查询条件
func (p *TenantPlugin) buildFindFilter(ctx context.Context, conf SharedConfig, tid int64) bson.M {
	// 1. 纯私有模式：硬隔离，任何情况下仅限访问本租户数据
	if conf.IsPrivate {
		return bson.M{p.tenantField: tid}
	}

	// 2. 共享模式：支持本租户数据与系统空间共享数据的复合检索
	if conf.IsShared {
		// 上下文指定了强私有检索，或当前访问者就是系统根租户本身
		if ctxutil.IsPrivateOnly(ctx) || tid == p.systemTenantID {
			return bson.M{p.tenantField: tid}
		}

		// 构建系统共享空间条件
		systemCond := bson.M{p.tenantField: p.systemTenantID}
		for k, v := range conf.Condition {
			systemCond[k] = v
		}

		// 召回本租户数据或系统空间的共享配置
		return bson.M{
			"$or": []bson.M{
				{p.tenantField: tid},
				systemCond,
			},
		}
	}

	// 3. 默认常规普通模型（不带 eiam 标签）：
	// 超管上帝视角放行，非超管则默认注入私有隔离
	if tid == p.systemTenantID {
		return nil
	}
	return bson.M{p.tenantField: tid}
}

// BeforeInsert 拦截写入操作，反射安全填充当前租户 ID
func (p *TenantPlugin) BeforeInsert(stmt *mongox.Statement) error {
	if IsIgnoreTenant(stmt.Context) {
		return nil
	}

	conf := p.getSharedConfig(stmt.Model)
	if conf.IsIgnore {
		return nil
	}

	tid := ctxutil.GetTenantID(stmt.Context).Int64()
	if tid <= 0 {
		return nil
	}

	p.setTenantIDValue(stmt.Model, tid)
	return nil
}

// ==========================================
// 内部反射解析与填充逻辑
// ==========================================

func (p *TenantPlugin) getSharedConfig(model interface{}) SharedConfig {
	if model == nil {
		return SharedConfig{}
	}

	t := reflect.TypeOf(model)
	for t.Kind() == reflect.Ptr || t.Kind() == reflect.Slice {
		t = t.Elem()
	}
	for t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	if t.Kind() != reflect.Struct {
		return SharedConfig{}
	}

	if val, ok := p.cache.Load(t); ok {
		return val.(SharedConfig)
	}

	conf := p.parseStruct(t)
	p.cache.Store(t, conf)
	return conf
}

func (p *TenantPlugin) parseStruct(t reflect.Type) SharedConfig {
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)

		eiamTag := field.Tag.Get("eiam")
		if eiamTag != "" {
			return p.parseEiamTag(eiamTag)
		}

		if field.Anonymous && field.Type.Kind() == reflect.Struct {
			cfg := p.parseStruct(field.Type)
			if cfg.IsPrivate || cfg.IsShared || cfg.IsIgnore {
				return cfg
			}
		}
	}
	return SharedConfig{}
}

func (p *TenantPlugin) parseEiamTag(tag string) SharedConfig {
	conf := SharedConfig{}
	parts := strings.SplitN(tag, ":", 2)
	switch parts[0] {
	case "ignore":
		conf.IsIgnore = true
	case "private":
		conf.IsPrivate = true
	case "shared":
		conf.IsShared = true
		if len(parts) > 1 {
			conf.Condition = parseCondition(parts[1])
		}
	}
	return conf
}

func (p *TenantPlugin) setTenantIDValue(model interface{}, tenantID int64) {
	if model == nil {
		return
	}
	v := reflect.ValueOf(model)
	p.reflectAndSet(v, tenantID)
}

func (p *TenantPlugin) reflectAndSet(v reflect.Value, tenantID int64) {
	for v.Kind() == reflect.Ptr || v.Kind() == reflect.Interface {
		if v.IsNil() {
			return
		}
		v = v.Elem()
	}

	switch v.Kind() {
	case reflect.Slice, reflect.Array:
		for i := 0; i < v.Len(); i++ {
			p.reflectAndSet(v.Index(i), tenantID)
		}
	case reflect.Struct:
		p.setField(v, tenantID)
	}
}

func (p *TenantPlugin) setField(v reflect.Value, tenantID int64) {
	f := v.FieldByName("TenantID")
	if f.IsValid() && f.CanSet() && f.Int() == 0 {
		f.SetInt(tenantID)
		return
	}

	t := v.Type()
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		if field.Anonymous && field.Type.Kind() == reflect.Struct {
			p.setField(v.Field(i), tenantID)
		}
	}
}

// ==========================================
// 提权与隔离 Context 快捷辅助函数
// ==========================================

// IgnoreTenantContext 将跳过租户隔离标记注入 Context，允许业务服务层全局访问数据
func IgnoreTenantContext(ctx context.Context) context.Context {
	return context.WithValue(ctx, IGNORE_TENANT_KEY, true)
}

// IsIgnoreTenant 检查是否处于跳过隔离校验模式
func IsIgnoreTenant(ctx context.Context) bool {
	if ctx == nil {
		return false
	}
	val, ok := ctx.Value(IGNORE_TENANT_KEY).(bool)
	return ok && val
}

// ==========================================
// 内部辅助拼接函数
// ==========================================

func parseCondition(condStr string) bson.M {
	cond := bson.M{}
	if condStr == "" {
		return cond
	}
	// 1. 优先尝试以标准的 JSON 格式解析条件
	if strings.HasPrefix(condStr, "{") && strings.HasSuffix(condStr, "}") {
		if err := json.Unmarshal([]byte(condStr), &cond); err == nil {
			return cond
		}
	}
	// 2. 降级使用逗号分隔的名值对解析（如 "is_public=true,status=1"）
	for _, pair := range strings.Split(condStr, ",") {
		kv := strings.SplitN(pair, "=", 2)
		if len(kv) != 2 {
			continue
		}
		k := strings.TrimSpace(kv[0])
		vStr := strings.TrimSpace(kv[1])

		switch vStr {
		case "true":
			cond[k] = true
		case "false":
			cond[k] = false
		default:
			if val, err := strconv.ParseInt(vStr, 10, 64); err == nil {
				cond[k] = val
			} else if val, err := strconv.ParseFloat(vStr, 64); err == nil {
				cond[k] = val
			} else {
				cond[k] = vStr
			}
		}
	}
	return cond
}

func injectConditions(filter bson.M, cond bson.M) {
	if len(cond) == 0 {
		return
	}
	// 1. 若原 filter 为空，直接注入所有多租户条件
	if len(filter) == 0 {
		for k, v := range cond {
			filter[k] = v
		}
		return
	}

	// 2. 防御性深度集成：为了 100% 避免原 filter 里的 $or / $and 或同名键被覆盖或篡改，
	// 最安全且无语义歧义的做法是在顶层引入 $and，将“原 filter”和“租户隔离 cond”打包强强联合。
	// 这完全对标了关系型数据库的 (UserFilter) AND (TenantFilter) 隔离逻辑。

	// 如果原 filter 顶层已经是个 $and 节点，则扁平化追加，以维持 BSON 树的最优树深
	if andSlice, ok := filter["$and"]; ok {
		if slice, yes := andSlice.([]bson.M); yes {
			filter["$and"] = append(slice, cond)
			return
		}
		if slice, yes := andSlice.([]interface{}); yes {
			filter["$and"] = append(slice, cond)
			return
		}
	}

	// 否则，克隆打包原 filter，挂载至全新的顶层 $and 节点下
	orig := bson.M{}
	for k, v := range filter {
		orig[k] = v
		delete(filter, k)
	}
	filter["$and"] = []bson.M{orig, cond}
}
