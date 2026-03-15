package easyflow

import (
	"encoding/json"
	"testing"

	"github.com/Bunny3th/easy-workflow/workflow/model"
	"github.com/stretchr/testify/require"
)

func TestLogicFlow_Deploy_Benchmark(t *testing.T) {
	workflow := Workflow{
		Name:  "环境发版",
		Owner: "admin",
		FlowData: LogicFlow{
			Nodes: []map[string]interface{}{
				{"id": "6f9f19dd-e986-46b0-aacb-b27f29462fcf", "type": "start", "properties": map[string]interface{}{"is_notify": true}},
				{"id": "e45dc192-147e-49c9-be02-22139fbef53c", "type": "end"},
				{"id": "f8c6bc45-0626-4288-a4a6-537259aa8ccb", "type": "selective"},
				{"id": "23367682-80f6-4c73-9b47-74e083d0ea19", "type": "parallel"},
				{"id": "1ee7a9b6-ee76-4a9a-b32e-edfa607553c4", "type": "user", "properties": map[string]interface{}{"name": "后端", "approved": []string{"yangwz", "lim"}, "rule": "template", "template_field": "name", "template_id": 35}},
				{"id": "a13994d0-7e6f-49db-adf9-adceb1971ab6", "type": "condition"},
				{"id": "dfcad5d3-d1ca-44f1-a629-19ab1a9b07bf", "type": "condition"},
				{"id": "ec102725-53ba-49ea-8f2d-c08808a32466", "type": "condition"},
				{"id": "36707c1c-111b-461b-8f1c-38c26a9d1305", "type": "user", "properties": map[string]interface{}{"name": "前端", "approved": []string{"luankz"}, "rule": "appoint"}},
				{"id": "ff438369-305e-4efd-b964-f48f89637f5e", "type": "user", "properties": map[string]interface{}{"name": "算法", "approved": []string{"changjj"}, "rule": "appoint"}},
				{"id": "6e483548-8445-4af7-8dce-b778600514ac", "type": "parallel"},
				{"id": "efb34b56-8fe3-4385-b06c-363a51cbc91c", "type": "user", "properties": map[string]interface{}{"name": "产品", "approved": []string{"langtt"}, "rule": "appoint"}},
				{"id": "8fa6f6b2-b9ca-4415-bcdc-cc227e21b9ee", "type": "user", "properties": map[string]interface{}{"name": "领导", "approved": []string{"chenggs"}, "rule": "appoint"}},
				{"id": "b4845f3c-43f6-4040-81f8-5f96309c0490", "type": "automation", "properties": map[string]interface{}{"name": "自动化-部署", "codebook_uid": "agent", "is_notify": false}},
				{"id": "80852ea2-0de3-421f-8c75-0527f9999067", "type": "user", "properties": map[string]interface{}{"name": "123", "is_cc": true, "rule": "founder"}},
				{"id": "f9f1fe1e-6882-4c2d-942b-24bf9f3fa039", "type": "chat", "properties": map[string]interface{}{"name": "群通知", "title": "Agent 发版执行结果", "mode": "existing"}},
			},
			Edges: []map[string]interface{}{
				{"id": "e1", "sourceNodeId": "6f9f19dd-e986-46b0-aacb-b27f29462fcf", "targetNodeId": "80852ea2-0de3-421f-8c75-0527f9999067"},
				{"id": "e2", "sourceNodeId": "80852ea2-0de3-421f-8c75-0527f9999067", "targetNodeId": "f8c6bc45-0626-4288-a4a6-537259aa8ccb"},
				{"id": "e3", "sourceNodeId": "f8c6bc45-0626-4288-a4a6-537259aa8ccb", "targetNodeId": "a13994d0-7e6f-49db-adf9-adceb1971ab6"},
				{"id": "e4", "sourceNodeId": "f8c6bc45-0626-4288-a4a6-537259aa8ccb", "targetNodeId": "dfcad5d3-d1ca-44f1-a629-19ab1a9b07bf"},
				{"id": "e5", "sourceNodeId": "f8c6bc45-0626-4288-a4a6-537259aa8ccb", "targetNodeId": "ec102725-53ba-49ea-8f2d-c08808a32466"},
				{"id": "o1", "sourceNodeId": "a13994d0-7e6f-49db-adf9-adceb1971ab6", "targetNodeId": "1ee7a9b6-ee76-4a9a-b32e-edfa607553c4", "properties": map[string]interface{}{"expression": "$environment in ('backend')"}},
				{"id": "o2", "sourceNodeId": "dfcad5d3-d1ca-44f1-a629-19ab1a9b07bf", "targetNodeId": "36707c1c-111b-461b-8f1c-38c26a9d1305", "properties": map[string]interface{}{"expression": "$environment in ('frontend')"}},
				{"id": "o3", "sourceNodeId": "ec102725-53ba-49ea-8f2d-c08808a32466", "targetNodeId": "ff438369-305e-4efd-b964-f48f89637f5e", "properties": map[string]interface{}{"expression": "$environment in ('ocr')"}},
				{"id": "j1", "sourceNodeId": "1ee7a9b6-ee76-4a9a-b32e-edfa607553c4", "targetNodeId": "23367682-80f6-4c73-9b47-74e083d0ea19"},
				{"id": "j2", "sourceNodeId": "36707c1c-111b-461b-8f1c-38c26a9d1305", "targetNodeId": "23367682-80f6-4c73-9b47-74e083d0ea19"},
				{"id": "j3", "sourceNodeId": "ff438369-305e-4efd-b964-f48f89637f5e", "targetNodeId": "23367682-80f6-4c73-9b47-74e083d0ea19"},
				{"id": "p1", "sourceNodeId": "23367682-80f6-4c73-9b47-74e083d0ea19", "targetNodeId": "efb34b56-8fe3-4385-b06c-363a51cbc91c"},
				{"id": "p2", "sourceNodeId": "23367682-80f6-4c73-9b47-74e083d0ea19", "targetNodeId": "8fa6f6b2-b9ca-4415-bcdc-cc227e21b9ee"},
				{"id": "p3", "sourceNodeId": "efb34b56-8fe3-4385-b06c-363a51cbc91c", "targetNodeId": "6e483548-8445-4af7-8dce-b778600514ac"},
				{"id": "p4", "sourceNodeId": "8fa6f6b2-b9ca-4415-bcdc-cc227e21b9ee", "targetNodeId": "6e483548-8445-4af7-8dce-b778600514ac"},
				{"id": "a1", "sourceNodeId": "6e483548-8445-4af7-8dce-b778600514ac", "targetNodeId": "b4845f3c-43f6-4040-81f8-5f96309c0490"},
				{"id": "c1", "sourceNodeId": "b4845f3c-43f6-4040-81f8-5f96309c0490", "targetNodeId": "f9f1fe1e-6882-4c2d-942b-24bf9f3fa039"},
				{"id": "e_end", "sourceNodeId": "f9f1fe1e-6882-4c2d-942b-24bf9f3fa039", "targetNodeId": "e45dc192-147e-49c9-be02-22139fbef53c"},
			},
		},
	}

	converter := NewDefaultConverterWithHandlers()

	process, err := converter.Convert(workflow)
	require.NoError(t, err)

	expectedJSON := `{
		"ProcessName": "环境发版",
		"Source": "工单系统",
		"RevokeEvents": ["EventRevoke"],
		"Nodes": [{"NodeID":"6f9f19dd-e986-46b0-aacb-b27f29462fcf","NodeName":"Start","NodeType":0,"PrevNodeIDs":null,"UserIDs":["$starter"],"Roles":null,"GWConfig":{"Conditions":null,"InevitableNodes":null,"WaitForAllPrevNode":0},"IsCosigned":0,"NodeStartEvents":null,"NodeEndEvents":["EventStart"],"TaskFinishEvents":null},{"NodeID":"e45dc192-147e-49c9-be02-22139fbef53c","NodeName":"End","NodeType":3,"PrevNodeIDs":["f9f1fe1e-6882-4c2d-942b-24bf9f3fa039"],"UserIDs":null,"Roles":null,"GWConfig":{"Conditions":null,"InevitableNodes":null,"WaitForAllPrevNode":0},"IsCosigned":0,"NodeStartEvents":["EventNotify"],"NodeEndEvents":null,"TaskFinishEvents":null},{"NodeID":"f8c6bc45-0626-4288-a4a6-537259aa8ccb","NodeName":"条件并行网关","NodeType":2,"PrevNodeIDs":["80852ea2-0de3-421f-8c75-0527f9999067"],"UserIDs":null,"Roles":null,"GWConfig":{"Conditions":null,"InevitableNodes":["a13994d0-7e6f-49db-adf9-adceb1971ab6","dfcad5d3-d1ca-44f1-a629-19ab1a9b07bf","ec102725-53ba-49ea-8f2d-c08808a32466"],"WaitForAllPrevNode":1},"IsCosigned":0,"NodeStartEvents":["EventSelectiveGatewaySplit"],"NodeEndEvents":null,"TaskFinishEvents":null},{"NodeID":"23367682-80f6-4c73-9b47-74e083d0ea19","NodeName":"并行网关","NodeType":2,"PrevNodeIDs":["1ee7a9b6-ee76-4a9a-b32e-edfa607553c4","36707c1c-111b-461b-8f1c-38c26a9d1305","ff438369-305e-4efd-b964-f48f89637f5e"],"UserIDs":null,"Roles":null,"GWConfig":{"Conditions":null,"InevitableNodes":["efb34b56-8fe3-4385-b06c-363a51cbc91c","8fa6f6b2-b9ca-4415-bcdc-cc227e21b9ee"],"WaitForAllPrevNode":1},"IsCosigned":0,"NodeStartEvents":null,"NodeEndEvents":null,"TaskFinishEvents":null},{"NodeID":"1ee7a9b6-ee76-4a9a-b32e-edfa607553c4","NodeName":"后端","NodeType":1,"PrevNodeIDs":["a13994d0-7e6f-49db-adf9-adceb1971ab6"],"UserIDs":["yangwz","lim"],"Roles":null,"GWConfig":{"Conditions":null,"InevitableNodes":null,"WaitForAllPrevNode":0},"IsCosigned":0,"NodeStartEvents":["EventNotify"],"NodeEndEvents":null,"TaskFinishEvents":["EventTaskParallelNodePass","EventConcurrentRejectCleanup"]},{"NodeID":"a13994d0-7e6f-49db-adf9-adceb1971ab6","NodeName":"","NodeType":2,"PrevNodeIDs":["f8c6bc45-0626-4288-a4a6-537259aa8ccb"],"UserIDs":null,"Roles":null,"GWConfig":{"Conditions":[{"Expression":"$environment in ('backend')","NodeID":"1ee7a9b6-ee76-4a9a-b32e-edfa607553c4"}],"InevitableNodes":[],"WaitForAllPrevNode":3},"IsCosigned":0,"NodeStartEvents":null,"NodeEndEvents":null,"TaskFinishEvents":null},{"NodeID":"dfcad5d3-d1ca-44f1-a629-19ab1a9b07bf","NodeName":"","NodeType":2,"PrevNodeIDs":["f8c6bc45-0626-4288-a4a6-537259aa8ccb"],"UserIDs":null,"Roles":null,"GWConfig":{"Conditions":[{"Expression":"$environment in ('frontend')","NodeID":"36707c1c-111b-461b-8f1c-38c26a9d1305"}],"InevitableNodes":[],"WaitForAllPrevNode":3},"IsCosigned":0,"NodeStartEvents":null,"NodeEndEvents":null,"TaskFinishEvents":null},{"NodeID":"ec102725-53ba-49ea-8f2d-c08808a32466","NodeName":"","NodeType":2,"PrevNodeIDs":["f8c6bc45-0626-4288-a4a6-537259aa8ccb"],"UserIDs":null,"Roles":null,"GWConfig":{"Conditions":[{"Expression":"$environment in ('ocr')","NodeID":"ff438369-305e-4efd-b964-f48f89637f5e"}],"InevitableNodes":[],"WaitForAllPrevNode":3},"IsCosigned":0,"NodeStartEvents":null,"NodeEndEvents":null,"TaskFinishEvents":null},{"NodeID":"36707c1c-111b-461b-8f1c-38c26a9d1305","NodeName":"前端","NodeType":1,"PrevNodeIDs":["dfcad5d3-d1ca-44f1-a629-19ab1a9b07bf"],"UserIDs":["luankz"],"Roles":null,"GWConfig":{"Conditions":null,"InevitableNodes":null,"WaitForAllPrevNode":0},"IsCosigned":0,"NodeStartEvents":["EventNotify"],"NodeEndEvents":null,"TaskFinishEvents":["EventTaskParallelNodePass","EventConcurrentRejectCleanup"]},{"NodeID":"ff438369-305e-4efd-b964-f48f89637f5e","NodeName":"算法","NodeType":1,"PrevNodeIDs":["ec102725-53ba-49ea-8f2d-c08808a32466"],"UserIDs":["changjj"],"Roles":null,"GWConfig":{"Conditions":null,"InevitableNodes":null,"WaitForAllPrevNode":0},"IsCosigned":0,"NodeStartEvents":["EventNotify"],"NodeEndEvents":null,"TaskFinishEvents":["EventTaskParallelNodePass","EventConcurrentRejectCleanup"]},{"NodeID":"6e483548-8445-4af7-8dce-b778600514ac","NodeName":"并行网关","NodeType":2,"PrevNodeIDs":["efb34b56-8fe3-4385-b06c-363a51cbc91c","8fa6f6b2-b9ca-4415-bcdc-cc227e21b9ee"],"UserIDs":null,"Roles":null,"GWConfig":{"Conditions":null,"InevitableNodes":["b4845f3c-43f6-4040-81f8-5f96309c0490"],"WaitForAllPrevNode":1},"IsCosigned":0,"NodeStartEvents":null,"NodeEndEvents":null,"TaskFinishEvents":null},{"NodeID":"efb34b56-8fe3-4385-b06c-363a51cbc91c","NodeName":"产品","NodeType":1,"PrevNodeIDs":["23367682-80f6-4c73-9b47-74e083d0ea19"],"UserIDs":["langtt"],"Roles":null,"GWConfig":{"Conditions":null,"InevitableNodes":null,"WaitForAllPrevNode":0},"IsCosigned":0,"NodeStartEvents":["EventNotify"],"NodeEndEvents":null,"TaskFinishEvents":["EventTaskParallelNodePass"]},{"NodeID":"8fa6f6b2-b9ca-4415-bcdc-cc227e21b9ee","NodeName":"领导","NodeType":1,"PrevNodeIDs":["23367682-80f6-4c73-9b47-74e083d0ea19"],"UserIDs":["chenggs"],"Roles":null,"GWConfig":{"Conditions":null,"InevitableNodes":null,"WaitForAllPrevNode":0},"IsCosigned":0,"NodeStartEvents":["EventNotify"],"NodeEndEvents":null,"TaskFinishEvents":["EventTaskParallelNodePass"]},{"NodeID":"b4845f3c-43f6-4040-81f8-5f96309c0490","NodeName":"自动化-部署","NodeType":1,"PrevNodeIDs":["6e483548-8445-4af7-8dce-b778600514ac"],"UserIDs":["automation"],"Roles":null,"GWConfig":{"Conditions":null,"InevitableNodes":null,"WaitForAllPrevNode":0},"IsCosigned":0,"NodeStartEvents":["EventAutomation"],"NodeEndEvents":["EventNotify"],"TaskFinishEvents":null},{"NodeID":"80852ea2-0de3-421f-8c75-0527f9999067","NodeName":"123","NodeType":1,"PrevNodeIDs":["6f9f19dd-e986-46b0-aacb-b27f29462fcf"],"UserIDs":[],"Roles":null,"GWConfig":{"Conditions":null,"InevitableNodes":null,"WaitForAllPrevNode":0},"IsCosigned":0,"NodeStartEvents":["EventCarbonCopy"],"NodeEndEvents":null,"TaskFinishEvents":null}
		,{"NodeID":"f9f1fe1e-6882-4c2d-942b-24bf9f3fa039","NodeName":"群通知","NodeType":1,"PrevNodeIDs":["b4845f3c-43f6-4040-81f8-5f96309c0490"],"UserIDs":["chat_group"],"Roles":null,"GWConfig":{"Conditions":null,"InevitableNodes":null,"WaitForAllPrevNode":0},"IsCosigned":0,"NodeStartEvents":["EventChatGroup"],"NodeEndEvents":null,"TaskFinishEvents":null}]
	}`

	var expectedProcess model.Process
	require.NoError(t, json.Unmarshal([]byte(expectedJSON), &expectedProcess))

	require.Equal(t, len(expectedProcess.Nodes), len(process.Nodes), "节点数量不一致")

	actualNodes := make(map[string]model.Node)
	for _, n := range process.Nodes {
		actualNodes[n.NodeID] = n
	}

	for _, exp := range expectedProcess.Nodes {
		act, ok := actualNodes[exp.NodeID]
		require.True(t, ok, "缺少期望的节点: %s", exp.NodeID)
		require.Equal(t, exp, act, "节点 %s (%s) 属性不匹配", exp.NodeID, exp.NodeName)
	}
}

func TestLogicFlow_Deploy_ParallelBenchmark(t *testing.T) {
	workflow := Workflow{
		Name:  "并行测试",
		Owner: "admin",
		FlowData: LogicFlow{
			Nodes: []map[string]interface{}{
				{"id": "86a7c39c-28a3-4875-b08d-56b4223e2903", "type": "start", "properties": map[string]interface{}{"height": 40, "width": 40}},
				{"id": "37a7885b-6b04-4bf5-9c26-53abc005d20b", "type": "end", "properties": map[string]interface{}{"height": 40, "width": 40}},
				{"id": "c21ae10c-6f62-4d0d-91cb-daaa7965807b", "type": "user", "properties": map[string]interface{}{"name": "领导", "approved": []string{"luankz"}, "rule": "appoint"}},
				{"id": "117fa0c6-e0f6-4222-adfc-7e7d3b5103e3", "type": "user", "properties": map[string]interface{}{"name": "张三", "approved": []string{"liusq"}, "rule": "appoint"}},
				{"id": "e2717c58-9946-4b26-99ab-2f034163365d", "type": "condition", "properties": map[string]interface{}{"name": "开始条件"}},
				{"id": "f4b1de0d-ccff-471e-aa2c-179ce9a905b5", "type": "user", "properties": map[string]interface{}{"name": "王五", "approved": []string{"luankz"}, "rule": "appoint"}},
				{"id": "3f1761ee-57c6-437c-83a5-53450f7966a9", "type": "user", "properties": map[string]interface{}{"name": "李四", "approved": []string{"liyy", "peicg"}, "rule": "appoint", "is_cosigned": true}},
				{"id": "9766c409-7415-4948-8957-4da0bca9b4bd", "type": "condition", "properties": map[string]interface{}{"name": "结束条件"}},
				{"id": "2d514c1e-b79d-46a4-aef8-65a83130495d", "type": "condition"},
				{"id": "95c6edc7-34b6-4e2a-a48f-5e5de0ef9282", "type": "user", "properties": map[string]interface{}{"name": "李四", "approved": []string{"liyy"}, "rule": "appoint"}},
				{"id": "fbe83138-b383-45ac-8c6f-a47a06bb648c", "type": "condition"},
				{"id": "8eb64352-d127-47ef-8a36-562986dc6c27", "type": "user", "properties": map[string]interface{}{"name": "提交人", "approved": []string{"chenggs"}, "rule": "appoint"}},
				{"id": "d1827b8a-0801-4cc1-ada8-73a42487de9f", "type": "parallel"},
				{"id": "6fb0e7ec-8fdf-46a2-9ef8-d629fdae65c0", "type": "parallel"},
				{"id": "f6a7a752-ed54-4bff-ae87-689b797dac92", "type": "user", "properties": map[string]interface{}{"name": "提交人", "rule": "founder"}},
			},
			Edges: []map[string]interface{}{
				{"id": "97333e79-ec3b-46ff-a548-888426e3c584", "sourceNodeId": "c21ae10c-6f62-4d0d-91cb-daaa7965807b", "targetNodeId": "37a7885b-6b04-4bf5-9c26-53abc005d20b"},
				{"id": "69fb51e5-e141-4a5e-b280-6b311a719326", "sourceNodeId": "e2717c58-9946-4b26-99ab-2f034163365d", "targetNodeId": "117fa0c6-e0f6-4222-adfc-7e7d3b5103e3", "properties": map[string]interface{}{"expression": "$name = '张三'"}},
				{"id": "4c0cd2ff-9c3b-4ae2-b977-4d2760a39062", "sourceNodeId": "e2717c58-9946-4b26-99ab-2f034163365d", "targetNodeId": "3f1761ee-57c6-437c-83a5-53450f7966a9", "properties": map[string]interface{}{"expression": "$name = '李四'"}},
				{"id": "bee9e46e-b13d-40d9-b008-01fc8c10d2c8", "sourceNodeId": "117fa0c6-e0f6-4222-adfc-7e7d3b5103e3", "targetNodeId": "9766c409-7415-4948-8957-4da0bca9b4bd"},
				{"id": "0b022bda-b351-41b7-94d2-8f7024ef0c15", "sourceNodeId": "3f1761ee-57c6-437c-83a5-53450f7966a9", "targetNodeId": "9766c409-7415-4948-8957-4da0bca9b4bd"},
				{"id": "9388ffb6-ef59-44c1-98f4-2cc16435cc64", "sourceNodeId": "95c6edc7-34b6-4e2a-a48f-5e5de0ef9282", "targetNodeId": "fbe83138-b383-45ac-8c6f-a47a06bb648c"},
				{"id": "0b92756b-14a1-47db-9dfa-b4c91e1c3ca0", "sourceNodeId": "86a7c39c-28a3-4875-b08d-56b4223e2903", "targetNodeId": "8eb64352-d127-47ef-8a36-562986dc6c27"},
				{"id": "3d4d8844-c08d-47e7-9603-51e771f1356c", "sourceNodeId": "6fb0e7ec-8fdf-46a2-9ef8-d629fdae65c0", "targetNodeId": "c21ae10c-6f62-4d0d-91cb-daaa7965807b"},
				{"id": "9067f40d-e3f8-4a17-8476-2b9994202cfa", "sourceNodeId": "f4b1de0d-ccff-471e-aa2c-179ce9a905b5", "targetNodeId": "6fb0e7ec-8fdf-46a2-9ef8-d629fdae65c0"},
				{"id": "1ac9be5c-589c-4914-b1e5-eb3641874b46", "sourceNodeId": "d1827b8a-0801-4cc1-ada8-73a42487de9f", "targetNodeId": "f4b1de0d-ccff-471e-aa2c-179ce9a905b5"},
				{"id": "fd000417-d31b-4b3d-b997-1f687658910f", "sourceNodeId": "d1827b8a-0801-4cc1-ada8-73a42487de9f", "targetNodeId": "e2717c58-9946-4b26-99ab-2f034163365d"},
				{"id": "7360fbda-f694-4b32-af0f-620acbaf1d02", "sourceNodeId": "8eb64352-d127-47ef-8a36-562986dc6c27", "targetNodeId": "d1827b8a-0801-4cc1-ada8-73a42487de9f"},
				{"id": "ccaaf5d9-f2c5-4f08-adc4-2f3d34080af0", "sourceNodeId": "9766c409-7415-4948-8957-4da0bca9b4bd", "targetNodeId": "6fb0e7ec-8fdf-46a2-9ef8-d629fdae65c0"},
				{"id": "4929ed1b-567b-41ae-8927-5e3ac97df38a", "sourceNodeId": "fbe83138-b383-45ac-8c6f-a47a06bb648c", "targetNodeId": "6fb0e7ec-8fdf-46a2-9ef8-d629fdae65c0"},
				{"id": "f65a2b5a-e936-4da2-9993-8665c1fa4f18", "sourceNodeId": "2d514c1e-b79d-46a4-aef8-65a83130495d", "targetNodeId": "95c6edc7-34b6-4e2a-a48f-5e5de0ef9282", "properties": map[string]interface{}{"expression": "$name = '李四'"}},
				{"id": "f379d8fc-bd85-426d-88d7-22995d378113", "sourceNodeId": "d1827b8a-0801-4cc1-ada8-73a42487de9f", "targetNodeId": "f6a7a752-ed54-4bff-ae87-689b797dac92"},
				{"id": "a4d1405e-b6eb-46a5-855b-295b9683fd95", "sourceNodeId": "f6a7a752-ed54-4bff-ae87-689b797dac92", "targetNodeId": "2d514c1e-b79d-46a4-aef8-65a83130495d"},
			},
		},
	}

	converter := NewDefaultConverterWithHandlers()

	process, err := converter.Convert(workflow)
	require.NoError(t, err)

	t.Logf("Total nodes: %d", len(process.Nodes))
	for _, n := range process.Nodes {
		t.Logf("NodeID: %s, NodeName: %s, Type: %d, PrevNodeIDs: %v", n.NodeID, n.NodeName, n.NodeType, n.PrevNodeIDs)
	}

	expectedJSON := `{"ProcessName":"Agent 发版_并行","Source":"工单系统","RevokeEvents":["EventRevoke"],"Nodes":[{"NodeID":"86a7c39c-28a3-4875-b08d-56b4223e2903","NodeName":"Start","NodeType":0,"PrevNodeIDs":null,"UserIDs":["$starter"],"Roles":null,"GWConfig":{"Conditions":null,"InevitableNodes":null,"WaitForAllPrevNode":0},"IsCosigned":0,"NodeStartEvents":null,"NodeEndEvents":["EventStart"],"TaskFinishEvents":null},{"NodeID":"37a7885b-6b04-4bf5-9c26-53abc005d20b","NodeName":"End","NodeType":3,"PrevNodeIDs":["c21ae10c-6f62-4d0d-91cb-daaa7965807b"],"UserIDs":null,"Roles":null,"GWConfig":{"Conditions":null,"InevitableNodes":null,"WaitForAllPrevNode":0},"IsCosigned":0,"NodeStartEvents":["EventNotify"],"NodeEndEvents":null,"TaskFinishEvents":null},{"NodeID":"c21ae10c-6f62-4d0d-91cb-daaa7965807b","NodeName":"领导","NodeType":1,"PrevNodeIDs":["6fb0e7ec-8fdf-46a2-9ef8-d629fdae65c0"],"UserIDs":["luankz"],"Roles":null,"GWConfig":{"Conditions":null,"InevitableNodes":null,"WaitForAllPrevNode":0},"IsCosigned":0,"NodeStartEvents":["EventNotify"],"NodeEndEvents":null,"TaskFinishEvents":["EventGatewayConditionReject"]},{"NodeID":"117fa0c6-e0f6-4222-adfc-7e7d3b5103e3","NodeName":"张三","NodeType":1,"PrevNodeIDs":["e2717c58-9946-4b26-99ab-2f034163365d"],"UserIDs":["liusq"],"Roles":null,"GWConfig":{"Conditions":null,"InevitableNodes":null,"WaitForAllPrevNode":0},"IsCosigned":0,"NodeStartEvents":["EventNotify"],"NodeEndEvents":null,"TaskFinishEvents":["EventConcurrentRejectCleanup"]},{"NodeID":"e2717c58-9946-4b26-99ab-2f034163365d","NodeName":"开始条件","NodeType":2,"PrevNodeIDs":["d1827b8a-0801-4cc1-ada8-73a42487de9f"],"UserIDs":null,"Roles":null,"GWConfig":{"Conditions":[{"Expression":"$name = '张三'","NodeID":"117fa0c6-e0f6-4222-adfc-7e7d3b5103e3"},{"Expression":"$name = '李四'","NodeID":"3f1761ee-57c6-437c-83a5-53450f7966a9"}],"InevitableNodes":[],"WaitForAllPrevNode":3},"IsCosigned":0,"NodeStartEvents":null,"NodeEndEvents":null,"TaskFinishEvents":null},{"NodeID":"f4b1de0d-ccff-471e-aa2c-179ce9a905b5","NodeName":"王五","NodeType":1,"PrevNodeIDs":["d1827b8a-0801-4cc1-ada8-73a42487de9f"],"UserIDs":["luankz"],"Roles":null,"GWConfig":{"Conditions":null,"InevitableNodes":null,"WaitForAllPrevNode":0},"IsCosigned":0,"NodeStartEvents":["EventNotify"],"NodeEndEvents":null,"TaskFinishEvents":["EventTaskParallelNodePass","EventUserNodeRejectProxyCleanup"]},{"NodeID":"3f1761ee-57c6-437c-83a5-53450f7966a9","NodeName":"李四","NodeType":1,"PrevNodeIDs":["e2717c58-9946-4b26-99ab-2f034163365d"],"UserIDs":["liyy","peicg"],"Roles":null,"GWConfig":{"Conditions":null,"InevitableNodes":null,"WaitForAllPrevNode":0},"IsCosigned":1,"NodeStartEvents":["EventNotify"],"NodeEndEvents":null,"TaskFinishEvents":["EventConcurrentRejectCleanup"]},{"NodeID":"9766c409-7415-4948-8957-4da0bca9b4bd","NodeName":"结束条件","NodeType":2,"PrevNodeIDs":["117fa0c6-e0f6-4222-adfc-7e7d3b5103e3","3f1761ee-57c6-437c-83a5-53450f7966a9"],"UserIDs":null,"Roles":null,"GWConfig":{"Conditions":[{"Expression":"1 = 1","NodeID":"proxy_9766c409-7415-4948-8957-4da0bca9b4bd_6fb0e7ec-8fdf-46a2-9ef8-d629fdae65c0"}],"InevitableNodes":[],"WaitForAllPrevNode":3},"IsCosigned":0,"NodeStartEvents":null,"NodeEndEvents":null,"TaskFinishEvents":null},{"NodeID":"2d514c1e-b79d-46a4-aef8-65a83130495d","NodeName":"","NodeType":2,"PrevNodeIDs":["f6a7a752-ed54-4bff-ae87-689b797dac92"],"UserIDs":null,"Roles":null,"GWConfig":{"Conditions":[{"Expression":"$name = '李四'","NodeID":"95c6edc7-34b6-4e2a-a48f-5e5de0ef9282"}],"InevitableNodes":[],"WaitForAllPrevNode":3},"IsCosigned":0,"NodeStartEvents":null,"NodeEndEvents":null,"TaskFinishEvents":null},{"NodeID":"95c6edc7-34b6-4e2a-a48f-5e5de0ef9282","NodeName":"李四","NodeType":1,"PrevNodeIDs":["2d514c1e-b79d-46a4-aef8-65a83130495d"],"UserIDs":["liyy"],"Roles":null,"GWConfig":{"Conditions":null,"InevitableNodes":null,"WaitForAllPrevNode":0},"IsCosigned":0,"NodeStartEvents":["EventNotify"],"NodeEndEvents":null,"TaskFinishEvents":null},{"NodeID":"fbe83138-b383-45ac-8c6f-a47a06bb648c","NodeName":"","NodeType":2,"PrevNodeIDs":["95c6edc7-34b6-4e2a-a48f-5e5de0ef9282"],"UserIDs":null,"Roles":null,"GWConfig":{"Conditions":[{"Expression":"1 = 1","NodeID":"proxy_fbe83138-b383-45ac-8c6f-a47a06bb648c_6fb0e7ec-8fdf-46a2-9ef8-d629fdae65c0"}],"InevitableNodes":[],"WaitForAllPrevNode":3},"IsCosigned":0,"NodeStartEvents":null,"NodeEndEvents":null,"TaskFinishEvents":null},{"NodeID":"8eb64352-d127-47ef-8a36-562986dc6c27","NodeName":"提交人","NodeType":1,"PrevNodeIDs":["86a7c39c-28a3-4875-b08d-56b4223e2903"],"UserIDs":["chenggs"],"Roles":null,"GWConfig":{"Conditions":null,"InevitableNodes":null,"WaitForAllPrevNode":0},"IsCosigned":0,"NodeStartEvents":["EventNotify"],"NodeEndEvents":null,"TaskFinishEvents":["EventTaskParallelNodePass"]},{"NodeID":"d1827b8a-0801-4cc1-ada8-73a42487de9f","NodeName":"并行网关","NodeType":2,"PrevNodeIDs":["8eb64352-d127-47ef-8a36-562986dc6c27"],"UserIDs":null,"Roles":null,"GWConfig":{"Conditions":null,"InevitableNodes":["f4b1de0d-ccff-471e-aa2c-179ce9a905b5","e2717c58-9946-4b26-99ab-2f034163365d","f6a7a752-ed54-4bff-ae87-689b797dac92"],"WaitForAllPrevNode":1},"IsCosigned":0,"NodeStartEvents":null,"NodeEndEvents":null,"TaskFinishEvents":null},{"NodeID":"proxy_9766c409-7415-4948-8957-4da0bca9b4bd_6fb0e7ec-8fdf-46a2-9ef8-d629fdae65c0","NodeName":"系统代理流转","NodeType":1,"PrevNodeIDs":["9766c409-7415-4948-8957-4da0bca9b4bd"],"UserIDs":["sys_auto"],"Roles":null,"GWConfig":{"Conditions":null,"InevitableNodes":null,"WaitForAllPrevNode":0},"IsCosigned":0,"NodeStartEvents":["EventNotify"],"NodeEndEvents":[],"TaskFinishEvents":["EventTaskParallelNodePass"]},{"NodeID":"proxy_fbe83138-b383-45ac-8c6f-a47a06bb648c_6fb0e7ec-8fdf-46a2-9ef8-d629fdae65c0","NodeName":"系统代理流转","NodeType":1,"PrevNodeIDs":["fbe83138-b383-45ac-8c6f-a47a06bb648c"],"UserIDs":["sys_auto"],"Roles":null,"GWConfig":{"Conditions":null,"InevitableNodes":null,"WaitForAllPrevNode":0},"IsCosigned":0,"NodeStartEvents":["EventNotify"],"NodeEndEvents":[],"TaskFinishEvents":["EventTaskParallelNodePass"]},{"NodeID":"6fb0e7ec-8fdf-46a2-9ef8-d629fdae65c0","NodeName":"并行网关","NodeType":2,"PrevNodeIDs":["f4b1de0d-ccff-471e-aa2c-179ce9a905b5","proxy_9766c409-7415-4948-8957-4da0bca9b4bd_6fb0e7ec-8fdf-46a2-9ef8-d629fdae65c0","proxy_fbe83138-b383-45ac-8c6f-a47a06bb648c_6fb0e7ec-8fdf-46a2-9ef8-d629fdae65c0"],"UserIDs":null,"Roles":null,"GWConfig":{"Conditions":null,"InevitableNodes":["c21ae10c-6f62-4d0d-91cb-daaa7965807b"],"WaitForAllPrevNode":1},"IsCosigned":0,"NodeStartEvents":null,"NodeEndEvents":null,"TaskFinishEvents":null},{"NodeID":"f6a7a752-ed54-4bff-ae87-689b797dac92","NodeName":"提交人","NodeType":1,"PrevNodeIDs":["d1827b8a-0801-4cc1-ada8-73a42487de9f"],"UserIDs":[],"Roles":null,"GWConfig":{"Conditions":null,"InevitableNodes":null,"WaitForAllPrevNode":0},"IsCosigned":0,"NodeStartEvents":["EventNotify"],"NodeEndEvents":null,"TaskFinishEvents":["EventUserNodeRejectProxyCleanup"]}]}`

	var expectedProcess model.Process
	require.NoError(t, json.Unmarshal([]byte(expectedJSON), &expectedProcess))

	require.Equal(t, len(expectedProcess.Nodes), len(process.Nodes), "节点数量不一致")

	actualNodes := make(map[string]model.Node)
	for _, n := range process.Nodes {
		actualNodes[n.NodeID] = n
	}

	for _, exp := range expectedProcess.Nodes {
		act, ok := actualNodes[exp.NodeID]
		require.True(t, ok, "缺少期望的节点: %s", exp.NodeID)
		require.Equal(t, exp, act, "节点 %s (%s) 属性不匹配", exp.NodeID, exp.NodeName)
	}
}
