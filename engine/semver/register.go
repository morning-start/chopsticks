package semver

import (
	"regexp"
	"strconv"

	"github.com/dop251/goja"
)

// Module 为脚本引擎提供 semver 注册。
type Module struct{}

// RegisterJS 向 JavaScript 运行时注册 semver 函数。
func (m *Module) RegisterJS(vm *goja.Runtime) {
	semverObj := vm.NewObject()

	semverObj.Set("parse", func(call goja.FunctionCall) goja.Value {
		version := call.Argument(0).String()
		v, err := Parse(version)
		if err != nil {
			return vm.ToValue(map[string]interface{}{"success": false, "error": err.Error()})
		}

		// 解析预发布编号
		prereleaseNum := 0
		if v.Pre != "" {
			// 尝试提取预发布编号，如 "beta4" -> 4
			re := regexp.MustCompile(`\d+$`)
			if match := re.FindString(v.Pre); match != "" {
				if num, err := strconv.Atoi(match); err == nil {
					prereleaseNum = num
				}
			}
		}

		// 构建数字段
		segments := []int{v.Major, v.Minor, v.Patch}

		return vm.ToValue(map[string]interface{}{
			"success":       true,
			"raw":           version,
			"normalized":    v.String(),
			"type":          "semver",
			"segments":      segments,
			"prerelease":    v.Pre,
			"prereleaseNum": prereleaseNum,
			"build":         v.Build,
			"comparable":    true,
		})
	})

	semverObj.Set("compare", func(call goja.FunctionCall) goja.Value {
		v1 := call.Argument(0).String()
		v2 := call.Argument(1).String()

		result, err := CompareStrings(v1, v2)
		if err != nil {
			return vm.ToValue(map[string]interface{}{"success": false, "error": err.Error()})
		}
		return vm.ToValue(map[string]interface{}{"success": true, "result": result})
	})

	semverObj.Set("gt", func(call goja.FunctionCall) goja.Value {
		v1 := call.Argument(0).String()
		v2 := call.Argument(1).String()

		ver1, err := Parse(v1)
		if err != nil {
			return vm.ToValue(map[string]interface{}{"success": false, "error": err.Error()})
		}
		ver2, err := Parse(v2)
		if err != nil {
			return vm.ToValue(map[string]interface{}{"success": false, "error": err.Error()})
		}
		return vm.ToValue(map[string]interface{}{"success": true, "result": ver1.GT(ver2)})
	})

	semverObj.Set("lt", func(call goja.FunctionCall) goja.Value {
		v1 := call.Argument(0).String()
		v2 := call.Argument(1).String()

		ver1, err := Parse(v1)
		if err != nil {
			return vm.ToValue(map[string]interface{}{"success": false, "error": err.Error()})
		}
		ver2, err := Parse(v2)
		if err != nil {
			return vm.ToValue(map[string]interface{}{"success": false, "error": err.Error()})
		}
		return vm.ToValue(map[string]interface{}{"success": true, "result": ver1.LT(ver2)})
	})

	semverObj.Set("eq", func(call goja.FunctionCall) goja.Value {
		v1 := call.Argument(0).String()
		v2 := call.Argument(1).String()

		ver1, err := Parse(v1)
		if err != nil {
			return vm.ToValue(map[string]interface{}{"success": false, "error": err.Error()})
		}
		ver2, err := Parse(v2)
		if err != nil {
			return vm.ToValue(map[string]interface{}{"success": false, "error": err.Error()})
		}
		return vm.ToValue(map[string]interface{}{"success": true, "result": ver1.EQ(ver2)})
	})

	semverObj.Set("gte", func(call goja.FunctionCall) goja.Value {
		v1 := call.Argument(0).String()
		v2 := call.Argument(1).String()

		ver1, err := Parse(v1)
		if err != nil {
			return vm.ToValue(map[string]interface{}{"success": false, "error": err.Error()})
		}
		ver2, err := Parse(v2)
		if err != nil {
			return vm.ToValue(map[string]interface{}{"success": false, "error": err.Error()})
		}
		return vm.ToValue(map[string]interface{}{"success": true, "result": ver1.GTE(ver2)})
	})

	semverObj.Set("lte", func(call goja.FunctionCall) goja.Value {
		v1 := call.Argument(0).String()
		v2 := call.Argument(1).String()

		ver1, err := Parse(v1)
		if err != nil {
			return vm.ToValue(map[string]interface{}{"success": false, "error": err.Error()})
		}
		ver2, err := Parse(v2)
		if err != nil {
			return vm.ToValue(map[string]interface{}{"success": false, "error": err.Error()})
		}
		return vm.ToValue(map[string]interface{}{"success": true, "result": ver1.LTE(ver2)})
	})

	semverObj.Set("satisfies", func(call goja.FunctionCall) goja.Value {
		version := call.Argument(0).String()
		constraint := call.Argument(1).String()

		ok, err := Satisfies(version, constraint)
		if err != nil {
			return vm.ToValue(map[string]interface{}{"success": false, "error": err.Error()})
		}
		return vm.ToValue(map[string]interface{}{"success": true, "result": ok})
	})

	semverObj.Set("normalize", func(call goja.FunctionCall) goja.Value {
		version := call.Argument(0).String()

		result, err := Normalize(version)
		if err != nil {
			return vm.ToValue(map[string]interface{}{"success": false, "error": err.Error()})
		}
		return vm.ToValue(map[string]interface{}{"success": true, "result": result})
	})

	semverObj.Set("detectType", func(call goja.FunctionCall) goja.Value {
		version := call.Argument(0).String()

		result := DetectType(version)
		return vm.ToValue(map[string]interface{}{"success": true, "result": result})
	})

	semverObj.Set("parseConstraint", func(call goja.FunctionCall) goja.Value {
		constraint := call.Argument(0).String()

		conType, conVersion, conOperator, err := ParseConstraint(constraint)
		if err != nil {
			return vm.ToValue(map[string]interface{}{"success": false, "error": err.Error()})
		}

		// 将 ConstraintType 转换为字符串表示
		typeStr := ""
		switch conType {
		case 0:
			typeStr = ">="
		case 1:
			typeStr = "<="
		case 2:
			typeStr = ">"
		case 3:
			typeStr = "<"
		case 4:
			typeStr = "^"
		case 5:
			typeStr = "~"
		case 6:
			typeStr = "="
		}

		return vm.ToValue(map[string]interface{}{
			"success":  true,
			"type":     typeStr,
			"version":  conVersion,
			"operator": conOperator,
		})
	})

	vm.Set("semver", semverObj)
}
