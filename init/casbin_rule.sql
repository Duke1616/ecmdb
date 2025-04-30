/*
 Navicat Premium Dump SQL

 Source Server         : mysql
 Source Server Type    : MySQL
 Source Server Version : 80037 (8.0.37)
 Source Host           : 127.0.0.1:3306
 Source Schema         : cmdb

 Target Server Type    : MySQL
 Target Server Version : 80037 (8.0.37)
 File Encoding         : 65001

 Date: 30/04/2025 14:14:54
*/

SET NAMES utf8mb4;
SET FOREIGN_KEY_CHECKS = 0;

-- ----------------------------
-- Table structure for casbin_rule
-- ----------------------------
DROP TABLE IF EXISTS `casbin_rule`;
CREATE TABLE `casbin_rule` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `ptype` varchar(100) COLLATE utf8mb4_general_ci DEFAULT NULL,
  `v0` varchar(100) COLLATE utf8mb4_general_ci DEFAULT NULL,
  `v1` varchar(100) COLLATE utf8mb4_general_ci DEFAULT NULL,
  `v2` varchar(100) COLLATE utf8mb4_general_ci DEFAULT NULL,
  `v3` varchar(100) COLLATE utf8mb4_general_ci DEFAULT NULL,
  `v4` varchar(100) COLLATE utf8mb4_general_ci DEFAULT NULL,
  `v5` varchar(100) COLLATE utf8mb4_general_ci DEFAULT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `idx_casbin_rule` (`ptype`,`v0`,`v1`,`v2`,`v3`,`v4`,`v5`)
) ENGINE=InnoDB AUTO_INCREMENT=9065 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;

-- ----------------------------
-- Records of casbin_rule
-- ----------------------------
BEGIN;
INSERT INTO `casbin_rule` (`id`, `ptype`, `v0`, `v1`, `v2`, `v3`, `v4`, `v5`) VALUES (2225, 'g', '1', 'admin', '', '', '', '');
INSERT INTO `casbin_rule` (`id`, `ptype`, `v0`, `v1`, `v2`, `v3`, `v4`, `v5`) VALUES (2226, 'g', '1', 'audit', '', '', '', '');
INSERT INTO `casbin_rule` (`id`, `ptype`, `v0`, `v1`, `v2`, `v3`, `v4`, `v5`) VALUES (8888, 'g', '16', 'start_order', '', '', '', '');
INSERT INTO `casbin_rule` (`id`, `ptype`, `v0`, `v1`, `v2`, `v3`, `v4`, `v5`) VALUES (4239, 'g', '3', 'dev', '', '', '', '');
INSERT INTO `casbin_rule` (`id`, `ptype`, `v0`, `v1`, `v2`, `v3`, `v4`, `v5`) VALUES (4236, 'g', '5', 'dev', '', '', '', '');
INSERT INTO `casbin_rule` (`id`, `ptype`, `v0`, `v1`, `v2`, `v3`, `v4`, `v5`) VALUES (4285, 'g', '6', 'dev', '', '', '', '');
INSERT INTO `casbin_rule` (`id`, `ptype`, `v0`, `v1`, `v2`, `v3`, `v4`, `v5`) VALUES (4286, 'g', '7', 'dev', '', '', '', '');
INSERT INTO `casbin_rule` (`id`, `ptype`, `v0`, `v1`, `v2`, `v3`, `v4`, `v5`) VALUES (8975, 'p', 'admin', '/api/attribute/create', 'POST', 'allow', '', '');
INSERT INTO `casbin_rule` (`id`, `ptype`, `v0`, `v1`, `v2`, `v3`, `v4`, `v5`) VALUES (8980, 'p', 'admin', '/api/attribute/custom/field', 'POST', 'allow', '', '');
INSERT INTO `casbin_rule` (`id`, `ptype`, `v0`, `v1`, `v2`, `v3`, `v4`, `v5`) VALUES (8978, 'p', 'admin', '/api/attribute/delete', 'POST', 'allow', '', '');
INSERT INTO `casbin_rule` (`id`, `ptype`, `v0`, `v1`, `v2`, `v3`, `v4`, `v5`) VALUES (8974, 'p', 'admin', '/api/attribute/group/create', 'POST', 'allow', '', '');
INSERT INTO `casbin_rule` (`id`, `ptype`, `v0`, `v1`, `v2`, `v3`, `v4`, `v5`) VALUES (8982, 'p', 'admin', '/api/attribute/list', 'POST', 'allow', '', '');
INSERT INTO `casbin_rule` (`id`, `ptype`, `v0`, `v1`, `v2`, `v3`, `v4`, `v5`) VALUES (8986, 'p', 'admin', '/api/attribute/list/field', 'POST', 'allow', '', '');
INSERT INTO `casbin_rule` (`id`, `ptype`, `v0`, `v1`, `v2`, `v3`, `v4`, `v5`) VALUES (9002, 'p', 'admin', '/api/attribute/update', 'POST', 'allow', '', '');
INSERT INTO `casbin_rule` (`id`, `ptype`, `v0`, `v1`, `v2`, `v3`, `v4`, `v5`) VALUES (8923, 'p', 'admin', '/api/codebook/create', 'POST', 'allow', '', '');
INSERT INTO `casbin_rule` (`id`, `ptype`, `v0`, `v1`, `v2`, `v3`, `v4`, `v5`) VALUES (8925, 'p', 'admin', '/api/codebook/delete', 'POST', 'allow', '', '');
INSERT INTO `casbin_rule` (`id`, `ptype`, `v0`, `v1`, `v2`, `v3`, `v4`, `v5`) VALUES (8993, 'p', 'admin', '/api/codebook/list', 'POST', 'allow', '', '');
INSERT INTO `casbin_rule` (`id`, `ptype`, `v0`, `v1`, `v2`, `v3`, `v4`, `v5`) VALUES (8924, 'p', 'admin', '/api/codebook/update', 'POST', 'allow', '', '');
INSERT INTO `casbin_rule` (`id`, `ptype`, `v0`, `v1`, `v2`, `v3`, `v4`, `v5`) VALUES (9005, 'p', 'admin', '/api/department/create', 'POST', 'allow', '', '');
INSERT INTO `casbin_rule` (`id`, `ptype`, `v0`, `v1`, `v2`, `v3`, `v4`, `v5`) VALUES (9007, 'p', 'admin', '/api/department/delete', 'POST', 'allow', '', '');
INSERT INTO `casbin_rule` (`id`, `ptype`, `v0`, `v1`, `v2`, `v3`, `v4`, `v5`) VALUES (9004, 'p', 'admin', '/api/department/list/tree', 'POST', 'allow', '', '');
INSERT INTO `casbin_rule` (`id`, `ptype`, `v0`, `v1`, `v2`, `v3`, `v4`, `v5`) VALUES (9006, 'p', 'admin', '/api/department/update', 'POST', 'allow', '', '');
INSERT INTO `casbin_rule` (`id`, `ptype`, `v0`, `v1`, `v2`, `v3`, `v4`, `v5`) VALUES (9060, 'p', 'admin', '/api/discovery/create', 'POST', 'allow', '', '');
INSERT INTO `casbin_rule` (`id`, `ptype`, `v0`, `v1`, `v2`, `v3`, `v4`, `v5`) VALUES (9057, 'p', 'admin', '/api/discovery/delete', 'POST', 'allow', '', '');
INSERT INTO `casbin_rule` (`id`, `ptype`, `v0`, `v1`, `v2`, `v3`, `v4`, `v5`) VALUES (9062, 'p', 'admin', '/api/discovery/list/by_template_id', 'POST', 'allow', '', '');
INSERT INTO `casbin_rule` (`id`, `ptype`, `v0`, `v1`, `v2`, `v3`, `v4`, `v5`) VALUES (9064, 'p', 'admin', '/api/discovery/sync', 'POST', 'allow', '', '');
INSERT INTO `casbin_rule` (`id`, `ptype`, `v0`, `v1`, `v2`, `v3`, `v4`, `v5`) VALUES (9058, 'p', 'admin', '/api/discovery/update', 'POST', 'allow', '', '');
INSERT INTO `casbin_rule` (`id`, `ptype`, `v0`, `v1`, `v2`, `v3`, `v4`, `v5`) VALUES (8999, 'p', 'admin', '/api/endpoint/list', 'POST', 'allow', '', '');
INSERT INTO `casbin_rule` (`id`, `ptype`, `v0`, `v1`, `v2`, `v3`, `v4`, `v5`) VALUES (8916, 'p', 'admin', '/api/menu/create', 'POST', 'allow', '', '');
INSERT INTO `casbin_rule` (`id`, `ptype`, `v0`, `v1`, `v2`, `v3`, `v4`, `v5`) VALUES (8918, 'p', 'admin', '/api/menu/delete', 'POST', 'allow', '', '');
INSERT INTO `casbin_rule` (`id`, `ptype`, `v0`, `v1`, `v2`, `v3`, `v4`, `v5`) VALUES (8998, 'p', 'admin', '/api/menu/list/tree', 'POST', 'allow', '', '');
INSERT INTO `casbin_rule` (`id`, `ptype`, `v0`, `v1`, `v2`, `v3`, `v4`, `v5`) VALUES (8917, 'p', 'admin', '/api/menu/update', 'POST', 'allow', '', '');
INSERT INTO `casbin_rule` (`id`, `ptype`, `v0`, `v1`, `v2`, `v3`, `v4`, `v5`) VALUES (8983, 'p', 'admin', '/api/model/by_group', 'POST', 'allow', '', '');
INSERT INTO `casbin_rule` (`id`, `ptype`, `v0`, `v1`, `v2`, `v3`, `v4`, `v5`) VALUES (8963, 'p', 'admin', '/api/model/by_uids', 'POST', 'allow', '', '');
INSERT INTO `casbin_rule` (`id`, `ptype`, `v0`, `v1`, `v2`, `v3`, `v4`, `v5`) VALUES (8970, 'p', 'admin', '/api/model/create', 'POST', 'allow', '', '');
INSERT INTO `casbin_rule` (`id`, `ptype`, `v0`, `v1`, `v2`, `v3`, `v4`, `v5`) VALUES (8972, 'p', 'admin', '/api/model/delete', 'POST', 'allow', '', '');
INSERT INTO `casbin_rule` (`id`, `ptype`, `v0`, `v1`, `v2`, `v3`, `v4`, `v5`) VALUES (8971, 'p', 'admin', '/api/model/group/create', 'POST', 'allow', '', '');
INSERT INTO `casbin_rule` (`id`, `ptype`, `v0`, `v1`, `v2`, `v3`, `v4`, `v5`) VALUES (8973, 'p', 'admin', '/api/model/group/delete', 'POST', 'allow', '', '');
INSERT INTO `casbin_rule` (`id`, `ptype`, `v0`, `v1`, `v2`, `v3`, `v4`, `v5`) VALUES (8976, 'p', 'admin', '/api/model/relation/create', 'POST', 'allow', '', '');
INSERT INTO `casbin_rule` (`id`, `ptype`, `v0`, `v1`, `v2`, `v3`, `v4`, `v5`) VALUES (8977, 'p', 'admin', '/api/model/relation/delete', 'POST', 'allow', '', '');
INSERT INTO `casbin_rule` (`id`, `ptype`, `v0`, `v1`, `v2`, `v3`, `v4`, `v5`) VALUES (8981, 'p', 'admin', '/api/model/relation/graph', 'POST', 'allow', '', '');
INSERT INTO `casbin_rule` (`id`, `ptype`, `v0`, `v1`, `v2`, `v3`, `v4`, `v5`) VALUES (8960, 'p', 'admin', '/api/model/relation/list', 'POST', 'allow', '', '');
INSERT INTO `casbin_rule` (`id`, `ptype`, `v0`, `v1`, `v2`, `v3`, `v4`, `v5`) VALUES (8945, 'p', 'admin', '/api/order/create', 'POST', 'allow', '', '');
INSERT INTO `casbin_rule` (`id`, `ptype`, `v0`, `v1`, `v2`, `v3`, `v4`, `v5`) VALUES (8948, 'p', 'admin', '/api/order/detail/process_inst_id', 'POST', 'allow', '', '');
INSERT INTO `casbin_rule` (`id`, `ptype`, `v0`, `v1`, `v2`, `v3`, `v4`, `v5`) VALUES (9010, 'p', 'admin', '/api/order/history', 'POST', 'allow', '', '');
INSERT INTO `casbin_rule` (`id`, `ptype`, `v0`, `v1`, `v2`, `v3`, `v4`, `v5`) VALUES (8953, 'p', 'admin', '/api/order/pass', 'POST', 'allow', '', '');
INSERT INTO `casbin_rule` (`id`, `ptype`, `v0`, `v1`, `v2`, `v3`, `v4`, `v5`) VALUES (8954, 'p', 'admin', '/api/order/reject', 'POST', 'allow', '', '');
INSERT INTO `casbin_rule` (`id`, `ptype`, `v0`, `v1`, `v2`, `v3`, `v4`, `v5`) VALUES (9011, 'p', 'admin', '/api/order/revoke', 'POST', 'allow', '', '');
INSERT INTO `casbin_rule` (`id`, `ptype`, `v0`, `v1`, `v2`, `v3`, `v4`, `v5`) VALUES (8947, 'p', 'admin', '/api/order/start/user', 'POST', 'allow', '', '');
INSERT INTO `casbin_rule` (`id`, `ptype`, `v0`, `v1`, `v2`, `v3`, `v4`, `v5`) VALUES (8949, 'p', 'admin', '/api/order/task/record', 'POST', 'allow', '', '');
INSERT INTO `casbin_rule` (`id`, `ptype`, `v0`, `v1`, `v2`, `v3`, `v4`, `v5`) VALUES (8952, 'p', 'admin', '/api/order/todo', 'POST', 'allow', '', '');
INSERT INTO `casbin_rule` (`id`, `ptype`, `v0`, `v1`, `v2`, `v3`, `v4`, `v5`) VALUES (8951, 'p', 'admin', '/api/order/todo/user', 'POST', 'allow', '', '');
INSERT INTO `casbin_rule` (`id`, `ptype`, `v0`, `v1`, `v2`, `v3`, `v4`, `v5`) VALUES (8913, 'p', 'admin', '/api/permission/change', 'POST', 'allow', '', '');
INSERT INTO `casbin_rule` (`id`, `ptype`, `v0`, `v1`, `v2`, `v3`, `v4`, `v5`) VALUES (8912, 'p', 'admin', '/api/permission/list', 'POST', 'allow', '', '');
INSERT INTO `casbin_rule` (`id`, `ptype`, `v0`, `v1`, `v2`, `v3`, `v4`, `v5`) VALUES (8979, 'p', 'admin', '/api/relation/create', 'POST', 'allow', '', '');
INSERT INTO `casbin_rule` (`id`, `ptype`, `v0`, `v1`, `v2`, `v3`, `v4`, `v5`) VALUES (8984, 'p', 'admin', '/api/relation/list', 'POST', 'allow', '', '');
INSERT INTO `casbin_rule` (`id`, `ptype`, `v0`, `v1`, `v2`, `v3`, `v4`, `v5`) VALUES (8956, 'p', 'admin', '/api/resource/create', 'POST', 'allow', '', '');
INSERT INTO `casbin_rule` (`id`, `ptype`, `v0`, `v1`, `v2`, `v3`, `v4`, `v5`) VALUES (8959, 'p', 'admin', '/api/resource/delete', 'POST', 'allow', '', '');
INSERT INTO `casbin_rule` (`id`, `ptype`, `v0`, `v1`, `v2`, `v3`, `v4`, `v5`) VALUES (8988, 'p', 'admin', '/api/resource/detail', 'POST', 'allow', '', '');
INSERT INTO `casbin_rule` (`id`, `ptype`, `v0`, `v1`, `v2`, `v3`, `v4`, `v5`) VALUES (8985, 'p', 'admin', '/api/resource/list', 'POST', 'allow', '', '');
INSERT INTO `casbin_rule` (`id`, `ptype`, `v0`, `v1`, `v2`, `v3`, `v4`, `v5`) VALUES (8962, 'p', 'admin', '/api/resource/list/ids', 'POST', 'allow', '', '');
INSERT INTO `casbin_rule` (`id`, `ptype`, `v0`, `v1`, `v2`, `v3`, `v4`, `v5`) VALUES (8967, 'p', 'admin', '/api/resource/relation/can_be_related', 'POST', 'allow', '', '');
INSERT INTO `casbin_rule` (`id`, `ptype`, `v0`, `v1`, `v2`, `v3`, `v4`, `v5`) VALUES (8968, 'p', 'admin', '/api/resource/relation/create', 'POST', 'allow', '', '');
INSERT INTO `casbin_rule` (`id`, `ptype`, `v0`, `v1`, `v2`, `v3`, `v4`, `v5`) VALUES (8969, 'p', 'admin', '/api/resource/relation/delete', 'POST', 'allow', '', '');
INSERT INTO `casbin_rule` (`id`, `ptype`, `v0`, `v1`, `v2`, `v3`, `v4`, `v5`) VALUES (8964, 'p', 'admin', '/api/resource/relation/graph', 'POST', 'allow', '', '');
INSERT INTO `casbin_rule` (`id`, `ptype`, `v0`, `v1`, `v2`, `v3`, `v4`, `v5`) VALUES (8966, 'p', 'admin', '/api/resource/relation/graph/add/left', 'POST', 'allow', '', '');
INSERT INTO `casbin_rule` (`id`, `ptype`, `v0`, `v1`, `v2`, `v3`, `v4`, `v5`) VALUES (8965, 'p', 'admin', '/api/resource/relation/graph/add/right', 'POST', 'allow', '', '');
INSERT INTO `casbin_rule` (`id`, `ptype`, `v0`, `v1`, `v2`, `v3`, `v4`, `v5`) VALUES (8961, 'p', 'admin', '/api/resource/relation/pipeline/all', 'POST', 'allow', '', '');
INSERT INTO `casbin_rule` (`id`, `ptype`, `v0`, `v1`, `v2`, `v3`, `v4`, `v5`) VALUES (9001, 'p', 'admin', '/api/resource/search', 'POST', 'allow', '', '');
INSERT INTO `casbin_rule` (`id`, `ptype`, `v0`, `v1`, `v2`, `v3`, `v4`, `v5`) VALUES (8955, 'p', 'admin', '/api/resource/secure', 'POST', 'allow', '', '');
INSERT INTO `casbin_rule` (`id`, `ptype`, `v0`, `v1`, `v2`, `v3`, `v4`, `v5`) VALUES (9028, 'p', 'admin', '/api/resource/set_custom_field', 'POST', 'allow', '', '');
INSERT INTO `casbin_rule` (`id`, `ptype`, `v0`, `v1`, `v2`, `v3`, `v4`, `v5`) VALUES (9003, 'p', 'admin', '/api/resource/update', 'POST', 'allow', '', '');
INSERT INTO `casbin_rule` (`id`, `ptype`, `v0`, `v1`, `v2`, `v3`, `v4`, `v5`) VALUES (8914, 'p', 'admin', '/api/role/create', 'POST', 'allow', '', '');
INSERT INTO `casbin_rule` (`id`, `ptype`, `v0`, `v1`, `v2`, `v3`, `v4`, `v5`) VALUES (8919, 'p', 'admin', '/api/role/delete', 'POST', 'allow', '', '');
INSERT INTO `casbin_rule` (`id`, `ptype`, `v0`, `v1`, `v2`, `v3`, `v4`, `v5`) VALUES (8997, 'p', 'admin', '/api/role/list', 'POST', 'allow', '', '');
INSERT INTO `casbin_rule` (`id`, `ptype`, `v0`, `v1`, `v2`, `v3`, `v4`, `v5`) VALUES (8915, 'p', 'admin', '/api/role/update', 'POST', 'allow', '', '');
INSERT INTO `casbin_rule` (`id`, `ptype`, `v0`, `v1`, `v2`, `v3`, `v4`, `v5`) VALUES (8922, 'p', 'admin', '/api/role/user/does_not_have', 'POST', 'allow', '', '');
INSERT INTO `casbin_rule` (`id`, `ptype`, `v0`, `v1`, `v2`, `v3`, `v4`, `v5`) VALUES (8921, 'p', 'admin', '/api/role/user/have', 'POST', 'allow', '', '');
INSERT INTO `casbin_rule` (`id`, `ptype`, `v0`, `v1`, `v2`, `v3`, `v4`, `v5`) VALUES (9026, 'p', 'admin', '/api/rota/create', 'POST', 'allow', '', '');
INSERT INTO `casbin_rule` (`id`, `ptype`, `v0`, `v1`, `v2`, `v3`, `v4`, `v5`) VALUES (9024, 'p', 'admin', '/api/rota/delete', 'POST', 'allow', '', '');
INSERT INTO `casbin_rule` (`id`, `ptype`, `v0`, `v1`, `v2`, `v3`, `v4`, `v5`) VALUES (9015, 'p', 'admin', '/api/rota/list', 'POST', 'allow', '', '');
INSERT INTO `casbin_rule` (`id`, `ptype`, `v0`, `v1`, `v2`, `v3`, `v4`, `v5`) VALUES (9019, 'p', 'admin', '/api/rota/rule/detail', 'POST', 'allow', '', '');
INSERT INTO `casbin_rule` (`id`, `ptype`, `v0`, `v1`, `v2`, `v3`, `v4`, `v5`) VALUES (9022, 'p', 'admin', '/api/rota/rule/shift_adjustment/add', 'POST', 'allow', '', '');
INSERT INTO `casbin_rule` (`id`, `ptype`, `v0`, `v1`, `v2`, `v3`, `v4`, `v5`) VALUES (9021, 'p', 'admin', '/api/rota/rule/shift_adjustment/delete', 'POST', 'allow', '', '');
INSERT INTO `casbin_rule` (`id`, `ptype`, `v0`, `v1`, `v2`, `v3`, `v4`, `v5`) VALUES (9023, 'p', 'admin', '/api/rota/rule/shift_adjustment/update', 'POST', 'allow', '', '');
INSERT INTO `casbin_rule` (`id`, `ptype`, `v0`, `v1`, `v2`, `v3`, `v4`, `v5`) VALUES (9017, 'p', 'admin', '/api/rota/rule/shift_scheduling/add', 'POST', 'allow', '', '');
INSERT INTO `casbin_rule` (`id`, `ptype`, `v0`, `v1`, `v2`, `v3`, `v4`, `v5`) VALUES (9020, 'p', 'admin', '/api/rota/rule/shift_scheduling/update', 'POST', 'allow', '', '');
INSERT INTO `casbin_rule` (`id`, `ptype`, `v0`, `v1`, `v2`, `v3`, `v4`, `v5`) VALUES (9016, 'p', 'admin', '/api/rota/schedule/preview', 'POST', 'allow', '', '');
INSERT INTO `casbin_rule` (`id`, `ptype`, `v0`, `v1`, `v2`, `v3`, `v4`, `v5`) VALUES (9025, 'p', 'admin', '/api/rota/update', 'POST', 'allow', '', '');
INSERT INTO `casbin_rule` (`id`, `ptype`, `v0`, `v1`, `v2`, `v3`, `v4`, `v5`) VALUES (8928, 'p', 'admin', '/api/runner/delete', 'POST', 'allow', '', '');
INSERT INTO `casbin_rule` (`id`, `ptype`, `v0`, `v1`, `v2`, `v3`, `v4`, `v5`) VALUES (8995, 'p', 'admin', '/api/runner/list', 'POST', 'allow', '', '');
INSERT INTO `casbin_rule` (`id`, `ptype`, `v0`, `v1`, `v2`, `v3`, `v4`, `v5`) VALUES (9063, 'p', 'admin', '/api/runner/list/by_ids', 'POST', 'allow', '', '');
INSERT INTO `casbin_rule` (`id`, `ptype`, `v0`, `v1`, `v2`, `v3`, `v4`, `v5`) VALUES (9061, 'p', 'admin', '/api/runner/list/by_workflow_id', 'POST', 'allow', '', '');
INSERT INTO `casbin_rule` (`id`, `ptype`, `v0`, `v1`, `v2`, `v3`, `v4`, `v5`) VALUES (9049, 'p', 'admin', '/api/runner/list/tags', 'POST', 'allow', '', '');
INSERT INTO `casbin_rule` (`id`, `ptype`, `v0`, `v1`, `v2`, `v3`, `v4`, `v5`) VALUES (8926, 'p', 'admin', '/api/runner/register', 'POST', 'allow', '', '');
INSERT INTO `casbin_rule` (`id`, `ptype`, `v0`, `v1`, `v2`, `v3`, `v4`, `v5`) VALUES (8927, 'p', 'admin', '/api/runner/update', 'POST', 'allow', '', '');
INSERT INTO `casbin_rule` (`id`, `ptype`, `v0`, `v1`, `v2`, `v3`, `v4`, `v5`) VALUES (9029, 'p', 'admin', '/api/sessions/tunnel', 'GET', 'allow', '', '');
INSERT INTO `casbin_rule` (`id`, `ptype`, `v0`, `v1`, `v2`, `v3`, `v4`, `v5`) VALUES (8996, 'p', 'admin', '/api/task/list', 'POST', 'allow', '', '');
INSERT INTO `casbin_rule` (`id`, `ptype`, `v0`, `v1`, `v2`, `v3`, `v4`, `v5`) VALUES (9047, 'p', 'admin', '/api/task/list/by_instance_id', 'POST', 'allow', '', '');
INSERT INTO `casbin_rule` (`id`, `ptype`, `v0`, `v1`, `v2`, `v3`, `v4`, `v5`) VALUES (8931, 'p', 'admin', '/api/task/retry', 'POST', 'allow', '', '');
INSERT INTO `casbin_rule` (`id`, `ptype`, `v0`, `v1`, `v2`, `v3`, `v4`, `v5`) VALUES (8929, 'p', 'admin', '/api/task/update/args', 'POST', 'allow', '', '');
INSERT INTO `casbin_rule` (`id`, `ptype`, `v0`, `v1`, `v2`, `v3`, `v4`, `v5`) VALUES (8930, 'p', 'admin', '/api/task/update/variables', 'POST', 'allow', '', '');
INSERT INTO `casbin_rule` (`id`, `ptype`, `v0`, `v1`, `v2`, `v3`, `v4`, `v5`) VALUES (8932, 'p', 'admin', '/api/template/create', 'POST', 'allow', '', '');
INSERT INTO `casbin_rule` (`id`, `ptype`, `v0`, `v1`, `v2`, `v3`, `v4`, `v5`) VALUES (8935, 'p', 'admin', '/api/template/delete', 'POST', 'allow', '', '');
INSERT INTO `casbin_rule` (`id`, `ptype`, `v0`, `v1`, `v2`, `v3`, `v4`, `v5`) VALUES (8946, 'p', 'admin', '/api/template/detail', 'POST', 'allow', '', '');
INSERT INTO `casbin_rule` (`id`, `ptype`, `v0`, `v1`, `v2`, `v3`, `v4`, `v5`) VALUES (9054, 'p', 'admin', '/api/template/get_by_workflow_id', 'POST', 'allow', '', '');
INSERT INTO `casbin_rule` (`id`, `ptype`, `v0`, `v1`, `v2`, `v3`, `v4`, `v5`) VALUES (9046, 'p', 'admin', '/api/template/group/by_ids', 'POST', 'allow', '', '');
INSERT INTO `casbin_rule` (`id`, `ptype`, `v0`, `v1`, `v2`, `v3`, `v4`, `v5`) VALUES (8933, 'p', 'admin', '/api/template/group/create', 'POST', 'allow', '', '');
INSERT INTO `casbin_rule` (`id`, `ptype`, `v0`, `v1`, `v2`, `v3`, `v4`, `v5`) VALUES (9045, 'p', 'admin', '/api/template/group/list', 'POST', 'allow', '', '');
INSERT INTO `casbin_rule` (`id`, `ptype`, `v0`, `v1`, `v2`, `v3`, `v4`, `v5`) VALUES (9043, 'p', 'admin', '/api/template/list', 'POST', 'allow', '', '');
INSERT INTO `casbin_rule` (`id`, `ptype`, `v0`, `v1`, `v2`, `v3`, `v4`, `v5`) VALUES (8989, 'p', 'admin', '/api/template/list/pipeline', 'POST', 'allow', '', '');
INSERT INTO `casbin_rule` (`id`, `ptype`, `v0`, `v1`, `v2`, `v3`, `v4`, `v5`) VALUES (9055, 'p', 'admin', '/api/template/rules/by_workflow_id', 'POST', 'allow', '', '');
INSERT INTO `casbin_rule` (`id`, `ptype`, `v0`, `v1`, `v2`, `v3`, `v4`, `v5`) VALUES (8934, 'p', 'admin', '/api/template/update', 'POST', 'allow', '', '');
INSERT INTO `casbin_rule` (`id`, `ptype`, `v0`, `v1`, `v2`, `v3`, `v4`, `v5`) VALUES (8987, 'p', 'admin', '/api/tools/minio/get_presigned_url', 'POST', 'allow', '', '');
INSERT INTO `casbin_rule` (`id`, `ptype`, `v0`, `v1`, `v2`, `v3`, `v4`, `v5`) VALUES (8958, 'p', 'admin', '/api/tools/minio/object/remove', 'POST', 'allow', '', '');
INSERT INTO `casbin_rule` (`id`, `ptype`, `v0`, `v1`, `v2`, `v3`, `v4`, `v5`) VALUES (8957, 'p', 'admin', '/api/tools/minio/put_presigned_url', 'POST', 'allow', '', '');
INSERT INTO `casbin_rule` (`id`, `ptype`, `v0`, `v1`, `v2`, `v3`, `v4`, `v5`) VALUES (9018, 'p', 'admin', '/api/user/find/by_ids', 'POST', 'allow', '', '');
INSERT INTO `casbin_rule` (`id`, `ptype`, `v0`, `v1`, `v2`, `v3`, `v4`, `v5`) VALUES (9053, 'p', 'admin', '/api/user/find/by_keyword', 'POST', 'allow', '', '');
INSERT INTO `casbin_rule` (`id`, `ptype`, `v0`, `v1`, `v2`, `v3`, `v4`, `v5`) VALUES (9008, 'p', 'admin', '/api/user/find/department_id', 'POST', 'allow', '', '');
INSERT INTO `casbin_rule` (`id`, `ptype`, `v0`, `v1`, `v2`, `v3`, `v4`, `v5`) VALUES (9027, 'p', 'admin', '/api/user/find/id', 'POST', 'allow', '', '');
INSERT INTO `casbin_rule` (`id`, `ptype`, `v0`, `v1`, `v2`, `v3`, `v4`, `v5`) VALUES (9052, 'p', 'admin', '/api/user/find/username', 'POST', 'allow', '', '');
INSERT INTO `casbin_rule` (`id`, `ptype`, `v0`, `v1`, `v2`, `v3`, `v4`, `v5`) VALUES (9051, 'p', 'admin', '/api/user/find/usernames', 'POST', 'allow', '', '');
INSERT INTO `casbin_rule` (`id`, `ptype`, `v0`, `v1`, `v2`, `v3`, `v4`, `v5`) VALUES (8911, 'p', 'admin', '/api/user/info', 'POST', 'allow', '', '');
INSERT INTO `casbin_rule` (`id`, `ptype`, `v0`, `v1`, `v2`, `v3`, `v4`, `v5`) VALUES (9014, 'p', 'admin', '/api/user/ldap/refresh_cache', 'POST', 'allow', '', '');
INSERT INTO `casbin_rule` (`id`, `ptype`, `v0`, `v1`, `v2`, `v3`, `v4`, `v5`) VALUES (9013, 'p', 'admin', '/api/user/ldap/search', 'POST', 'allow', '', '');
INSERT INTO `casbin_rule` (`id`, `ptype`, `v0`, `v1`, `v2`, `v3`, `v4`, `v5`) VALUES (9012, 'p', 'admin', '/api/user/ldap/sync', 'POST', 'allow', '', '');
INSERT INTO `casbin_rule` (`id`, `ptype`, `v0`, `v1`, `v2`, `v3`, `v4`, `v5`) VALUES (9000, 'p', 'admin', '/api/user/list', 'POST', 'allow', '', '');
INSERT INTO `casbin_rule` (`id`, `ptype`, `v0`, `v1`, `v2`, `v3`, `v4`, `v5`) VALUES (9050, 'p', 'admin', '/api/user/pipeline/department_id', 'POST', 'allow', '', '');
INSERT INTO `casbin_rule` (`id`, `ptype`, `v0`, `v1`, `v2`, `v3`, `v4`, `v5`) VALUES (8920, 'p', 'admin', '/api/user/role/bind', 'POST', 'allow', '', '');
INSERT INTO `casbin_rule` (`id`, `ptype`, `v0`, `v1`, `v2`, `v3`, `v4`, `v5`) VALUES (9009, 'p', 'admin', '/api/user/update', 'POST', 'allow', '', '');
INSERT INTO `casbin_rule` (`id`, `ptype`, `v0`, `v1`, `v2`, `v3`, `v4`, `v5`) VALUES (8994, 'p', 'admin', '/api/worker/list', 'POST', 'allow', '', '');
INSERT INTO `casbin_rule` (`id`, `ptype`, `v0`, `v1`, `v2`, `v3`, `v4`, `v5`) VALUES (9048, 'p', 'admin', '/api/workflow/create', 'POST', 'allow', '', '');
INSERT INTO `casbin_rule` (`id`, `ptype`, `v0`, `v1`, `v2`, `v3`, `v4`, `v5`) VALUES (8944, 'p', 'admin', '/api/workflow/delete', 'POST', 'allow', '', '');
INSERT INTO `casbin_rule` (`id`, `ptype`, `v0`, `v1`, `v2`, `v3`, `v4`, `v5`) VALUES (8942, 'p', 'admin', '/api/workflow/deploy', 'POST', 'allow', '', '');
INSERT INTO `casbin_rule` (`id`, `ptype`, `v0`, `v1`, `v2`, `v3`, `v4`, `v5`) VALUES (8950, 'p', 'admin', '/api/workflow/graph', 'POST', 'allow', '', '');
INSERT INTO `casbin_rule` (`id`, `ptype`, `v0`, `v1`, `v2`, `v3`, `v4`, `v5`) VALUES (9044, 'p', 'admin', '/api/workflow/list', 'POST', 'allow', '', '');
INSERT INTO `casbin_rule` (`id`, `ptype`, `v0`, `v1`, `v2`, `v3`, `v4`, `v5`) VALUES (9037, 'p', 'admin', '/api/workflow/update', 'POST', 'allow', '', '');
INSERT INTO `casbin_rule` (`id`, `ptype`, `v0`, `v1`, `v2`, `v3`, `v4`, `v5`) VALUES (2570, 'p', 'audit', '/api/model/group/create', 'POST', 'allow', '', '');
INSERT INTO `casbin_rule` (`id`, `ptype`, `v0`, `v1`, `v2`, `v3`, `v4`, `v5`) VALUES (8389, 'p', 'dev', '/api/attribute/list', 'POST', 'allow', '', '');
INSERT INTO `casbin_rule` (`id`, `ptype`, `v0`, `v1`, `v2`, `v3`, `v4`, `v5`) VALUES (8393, 'p', 'dev', '/api/attribute/list/field', 'POST', 'allow', '', '');
INSERT INTO `casbin_rule` (`id`, `ptype`, `v0`, `v1`, `v2`, `v3`, `v4`, `v5`) VALUES (8390, 'p', 'dev', '/api/model/by_group', 'POST', 'allow', '', '');
INSERT INTO `casbin_rule` (`id`, `ptype`, `v0`, `v1`, `v2`, `v3`, `v4`, `v5`) VALUES (8881, 'p', 'dev', '/api/model/by_uids', 'POST', 'allow', '', '');
INSERT INTO `casbin_rule` (`id`, `ptype`, `v0`, `v1`, `v2`, `v3`, `v4`, `v5`) VALUES (8388, 'p', 'dev', '/api/model/relation/graph', 'POST', 'allow', '', '');
INSERT INTO `casbin_rule` (`id`, `ptype`, `v0`, `v1`, `v2`, `v3`, `v4`, `v5`) VALUES (8878, 'p', 'dev', '/api/model/relation/list', 'POST', 'allow', '', '');
INSERT INTO `casbin_rule` (`id`, `ptype`, `v0`, `v1`, `v2`, `v3`, `v4`, `v5`) VALUES (8373, 'p', 'dev', '/api/order/create', 'POST', 'allow', '', '');
INSERT INTO `casbin_rule` (`id`, `ptype`, `v0`, `v1`, `v2`, `v3`, `v4`, `v5`) VALUES (8391, 'p', 'dev', '/api/relation/list', 'POST', 'allow', '', '');
INSERT INTO `casbin_rule` (`id`, `ptype`, `v0`, `v1`, `v2`, `v3`, `v4`, `v5`) VALUES (8376, 'p', 'dev', '/api/resource/create', 'POST', 'allow', '', '');
INSERT INTO `casbin_rule` (`id`, `ptype`, `v0`, `v1`, `v2`, `v3`, `v4`, `v5`) VALUES (8395, 'p', 'dev', '/api/resource/detail', 'POST', 'allow', '', '');
INSERT INTO `casbin_rule` (`id`, `ptype`, `v0`, `v1`, `v2`, `v3`, `v4`, `v5`) VALUES (8392, 'p', 'dev', '/api/resource/list', 'POST', 'allow', '', '');
INSERT INTO `casbin_rule` (`id`, `ptype`, `v0`, `v1`, `v2`, `v3`, `v4`, `v5`) VALUES (8880, 'p', 'dev', '/api/resource/list/ids', 'POST', 'allow', '', '');
INSERT INTO `casbin_rule` (`id`, `ptype`, `v0`, `v1`, `v2`, `v3`, `v4`, `v5`) VALUES (8385, 'p', 'dev', '/api/resource/relation/can_be_related', 'POST', 'allow', '', '');
INSERT INTO `casbin_rule` (`id`, `ptype`, `v0`, `v1`, `v2`, `v3`, `v4`, `v5`) VALUES (8386, 'p', 'dev', '/api/resource/relation/create', 'POST', 'allow', '', '');
INSERT INTO `casbin_rule` (`id`, `ptype`, `v0`, `v1`, `v2`, `v3`, `v4`, `v5`) VALUES (8387, 'p', 'dev', '/api/resource/relation/delete', 'POST', 'allow', '', '');
INSERT INTO `casbin_rule` (`id`, `ptype`, `v0`, `v1`, `v2`, `v3`, `v4`, `v5`) VALUES (8382, 'p', 'dev', '/api/resource/relation/graph', 'POST', 'allow', '', '');
INSERT INTO `casbin_rule` (`id`, `ptype`, `v0`, `v1`, `v2`, `v3`, `v4`, `v5`) VALUES (8384, 'p', 'dev', '/api/resource/relation/graph/add/left', 'POST', 'allow', '', '');
INSERT INTO `casbin_rule` (`id`, `ptype`, `v0`, `v1`, `v2`, `v3`, `v4`, `v5`) VALUES (8383, 'p', 'dev', '/api/resource/relation/graph/add/right', 'POST', 'allow', '', '');
INSERT INTO `casbin_rule` (`id`, `ptype`, `v0`, `v1`, `v2`, `v3`, `v4`, `v5`) VALUES (8879, 'p', 'dev', '/api/resource/relation/pipeline/all', 'POST', 'allow', '', '');
INSERT INTO `casbin_rule` (`id`, `ptype`, `v0`, `v1`, `v2`, `v3`, `v4`, `v5`) VALUES (8519, 'p', 'dev', '/api/resource/search', 'POST', 'allow', '', '');
INSERT INTO `casbin_rule` (`id`, `ptype`, `v0`, `v1`, `v2`, `v3`, `v4`, `v5`) VALUES (8375, 'p', 'dev', '/api/resource/secure', 'POST', 'allow', '', '');
INSERT INTO `casbin_rule` (`id`, `ptype`, `v0`, `v1`, `v2`, `v3`, `v4`, `v5`) VALUES (8398, 'p', 'dev', '/api/resource/update', 'POST', 'allow', '', '');
INSERT INTO `casbin_rule` (`id`, `ptype`, `v0`, `v1`, `v2`, `v3`, `v4`, `v5`) VALUES (8374, 'p', 'dev', '/api/template/detail', 'POST', 'allow', '', '');
INSERT INTO `casbin_rule` (`id`, `ptype`, `v0`, `v1`, `v2`, `v3`, `v4`, `v5`) VALUES (8396, 'p', 'dev', '/api/template/list/pipeline', 'POST', 'allow', '', '');
INSERT INTO `casbin_rule` (`id`, `ptype`, `v0`, `v1`, `v2`, `v3`, `v4`, `v5`) VALUES (8394, 'p', 'dev', '/api/tools/minio/get_presigned_url', 'POST', 'allow', '', '');
INSERT INTO `casbin_rule` (`id`, `ptype`, `v0`, `v1`, `v2`, `v3`, `v4`, `v5`) VALUES (8378, 'p', 'dev', '/api/tools/minio/object/remove', 'POST', 'allow', '', '');
INSERT INTO `casbin_rule` (`id`, `ptype`, `v0`, `v1`, `v2`, `v3`, `v4`, `v5`) VALUES (8377, 'p', 'dev', '/api/tools/minio/put_presigned_url', 'POST', 'allow', '', '');
INSERT INTO `casbin_rule` (`id`, `ptype`, `v0`, `v1`, `v2`, `v3`, `v4`, `v5`) VALUES (8372, 'p', 'dev', '/api/user/info', 'POST', 'allow', '', '');
INSERT INTO `casbin_rule` (`id`, `ptype`, `v0`, `v1`, `v2`, `v3`, `v4`, `v5`) VALUES (8907, 'p', 'start_order', '/api/order/create', 'POST', 'allow', '', '');
INSERT INTO `casbin_rule` (`id`, `ptype`, `v0`, `v1`, `v2`, `v3`, `v4`, `v5`) VALUES (8908, 'p', 'start_order', '/api/template/detail', 'POST', 'allow', '', '');
INSERT INTO `casbin_rule` (`id`, `ptype`, `v0`, `v1`, `v2`, `v3`, `v4`, `v5`) VALUES (8518, 'p', 'work_order', '/api/model/by_uids', 'POST', 'allow', '', '');
INSERT INTO `casbin_rule` (`id`, `ptype`, `v0`, `v1`, `v2`, `v3`, `v4`, `v5`) VALUES (8517, 'p', 'work_order', '/api/resource/search', 'POST', 'allow', '', '');
COMMIT;

SET FOREIGN_KEY_CHECKS = 1;
