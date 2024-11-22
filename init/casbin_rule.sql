/*
 Navicat Premium Data Transfer

 Source Server         : mysql
 Source Server Type    : MySQL
 Source Server Version : 80037
 Source Host           : 10.31.0.200:3306
 Source Schema         : cmdb

 Target Server Type    : MySQL
 Target Server Version : 80037
 File Encoding         : 65001

 Date: 22/11/2024 13:51:09
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
) ENGINE=InnoDB AUTO_INCREMENT=8201 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;

-- ----------------------------
-- Records of casbin_rule
-- ----------------------------
BEGIN;
INSERT INTO `casbin_rule` (`id`, `ptype`, `v0`, `v1`, `v2`, `v3`, `v4`, `v5`) VALUES (2225, 'g', '1', 'admin', '', '', '', '');
INSERT INTO `casbin_rule` (`id`, `ptype`, `v0`, `v1`, `v2`, `v3`, `v4`, `v5`) VALUES (2226, 'g', '1', 'audit', '', '', '', '');
INSERT INTO `casbin_rule` (`id`, `ptype`, `v0`, `v1`, `v2`, `v3`, `v4`, `v5`) VALUES (4239, 'g', '3', 'dev', '', '', '', '');
INSERT INTO `casbin_rule` (`id`, `ptype`, `v0`, `v1`, `v2`, `v3`, `v4`, `v5`) VALUES (4236, 'g', '5', 'dev', '', '', '', '');
INSERT INTO `casbin_rule` (`id`, `ptype`, `v0`, `v1`, `v2`, `v3`, `v4`, `v5`) VALUES (4285, 'g', '6', 'dev', '', '', '', '');
INSERT INTO `casbin_rule` (`id`, `ptype`, `v0`, `v1`, `v2`, `v3`, `v4`, `v5`) VALUES (4286, 'g', '7', 'dev', '', '', '', '');
INSERT INTO `casbin_rule` (`id`, `ptype`, `v0`, `v1`, `v2`, `v3`, `v4`, `v5`) VALUES (8149, 'p', 'admin', '/api/attribute/create', 'POST', 'allow', '', '');
INSERT INTO `casbin_rule` (`id`, `ptype`, `v0`, `v1`, `v2`, `v3`, `v4`, `v5`) VALUES (8154, 'p', 'admin', '/api/attribute/custom/field', 'POST', 'allow', '', '');
INSERT INTO `casbin_rule` (`id`, `ptype`, `v0`, `v1`, `v2`, `v3`, `v4`, `v5`) VALUES (8152, 'p', 'admin', '/api/attribute/delete', 'POST', 'allow', '', '');
INSERT INTO `casbin_rule` (`id`, `ptype`, `v0`, `v1`, `v2`, `v3`, `v4`, `v5`) VALUES (8148, 'p', 'admin', '/api/attribute/group/create', 'POST', 'allow', '', '');
INSERT INTO `casbin_rule` (`id`, `ptype`, `v0`, `v1`, `v2`, `v3`, `v4`, `v5`) VALUES (8156, 'p', 'admin', '/api/attribute/list', 'POST', 'allow', '', '');
INSERT INTO `casbin_rule` (`id`, `ptype`, `v0`, `v1`, `v2`, `v3`, `v4`, `v5`) VALUES (8160, 'p', 'admin', '/api/attribute/list/field', 'POST', 'allow', '', '');
INSERT INTO `casbin_rule` (`id`, `ptype`, `v0`, `v1`, `v2`, `v3`, `v4`, `v5`) VALUES (8175, 'p', 'admin', '/api/attribute/update', 'POST', 'allow', '', '');
INSERT INTO `casbin_rule` (`id`, `ptype`, `v0`, `v1`, `v2`, `v3`, `v4`, `v5`) VALUES (8100, 'p', 'admin', '/api/codebook/create', 'POST', 'allow', '', '');
INSERT INTO `casbin_rule` (`id`, `ptype`, `v0`, `v1`, `v2`, `v3`, `v4`, `v5`) VALUES (8102, 'p', 'admin', '/api/codebook/delete', 'POST', 'allow', '', '');
INSERT INTO `casbin_rule` (`id`, `ptype`, `v0`, `v1`, `v2`, `v3`, `v4`, `v5`) VALUES (8166, 'p', 'admin', '/api/codebook/list', 'POST', 'allow', '', '');
INSERT INTO `casbin_rule` (`id`, `ptype`, `v0`, `v1`, `v2`, `v3`, `v4`, `v5`) VALUES (8101, 'p', 'admin', '/api/codebook/update', 'POST', 'allow', '', '');
INSERT INTO `casbin_rule` (`id`, `ptype`, `v0`, `v1`, `v2`, `v3`, `v4`, `v5`) VALUES (8178, 'p', 'admin', '/api/department/create', 'POST', 'allow', '', '');
INSERT INTO `casbin_rule` (`id`, `ptype`, `v0`, `v1`, `v2`, `v3`, `v4`, `v5`) VALUES (8180, 'p', 'admin', '/api/department/delete', 'POST', 'allow', '', '');
INSERT INTO `casbin_rule` (`id`, `ptype`, `v0`, `v1`, `v2`, `v3`, `v4`, `v5`) VALUES (8177, 'p', 'admin', '/api/department/list/tree', 'POST', 'allow', '', '');
INSERT INTO `casbin_rule` (`id`, `ptype`, `v0`, `v1`, `v2`, `v3`, `v4`, `v5`) VALUES (8179, 'p', 'admin', '/api/department/update', 'POST', 'allow', '', '');
INSERT INTO `casbin_rule` (`id`, `ptype`, `v0`, `v1`, `v2`, `v3`, `v4`, `v5`) VALUES (8172, 'p', 'admin', '/api/endpoint/list', 'POST', 'allow', '', '');
INSERT INTO `casbin_rule` (`id`, `ptype`, `v0`, `v1`, `v2`, `v3`, `v4`, `v5`) VALUES (8093, 'p', 'admin', '/api/menu/create', 'POST', 'allow', '', '');
INSERT INTO `casbin_rule` (`id`, `ptype`, `v0`, `v1`, `v2`, `v3`, `v4`, `v5`) VALUES (8095, 'p', 'admin', '/api/menu/delete', 'POST', 'allow', '', '');
INSERT INTO `casbin_rule` (`id`, `ptype`, `v0`, `v1`, `v2`, `v3`, `v4`, `v5`) VALUES (8171, 'p', 'admin', '/api/menu/list/tree', 'POST', 'allow', '', '');
INSERT INTO `casbin_rule` (`id`, `ptype`, `v0`, `v1`, `v2`, `v3`, `v4`, `v5`) VALUES (8094, 'p', 'admin', '/api/menu/update', 'POST', 'allow', '', '');
INSERT INTO `casbin_rule` (`id`, `ptype`, `v0`, `v1`, `v2`, `v3`, `v4`, `v5`) VALUES (8157, 'p', 'admin', '/api/model/by_group', 'POST', 'allow', '', '');
INSERT INTO `casbin_rule` (`id`, `ptype`, `v0`, `v1`, `v2`, `v3`, `v4`, `v5`) VALUES (8144, 'p', 'admin', '/api/model/create', 'POST', 'allow', '', '');
INSERT INTO `casbin_rule` (`id`, `ptype`, `v0`, `v1`, `v2`, `v3`, `v4`, `v5`) VALUES (8146, 'p', 'admin', '/api/model/delete', 'POST', 'allow', '', '');
INSERT INTO `casbin_rule` (`id`, `ptype`, `v0`, `v1`, `v2`, `v3`, `v4`, `v5`) VALUES (8145, 'p', 'admin', '/api/model/group/create', 'POST', 'allow', '', '');
INSERT INTO `casbin_rule` (`id`, `ptype`, `v0`, `v1`, `v2`, `v3`, `v4`, `v5`) VALUES (8147, 'p', 'admin', '/api/model/group/delete', 'POST', 'allow', '', '');
INSERT INTO `casbin_rule` (`id`, `ptype`, `v0`, `v1`, `v2`, `v3`, `v4`, `v5`) VALUES (8150, 'p', 'admin', '/api/model/relation/create', 'POST', 'allow', '', '');
INSERT INTO `casbin_rule` (`id`, `ptype`, `v0`, `v1`, `v2`, `v3`, `v4`, `v5`) VALUES (8151, 'p', 'admin', '/api/model/relation/delete', 'POST', 'allow', '', '');
INSERT INTO `casbin_rule` (`id`, `ptype`, `v0`, `v1`, `v2`, `v3`, `v4`, `v5`) VALUES (8155, 'p', 'admin', '/api/model/relation/graph', 'POST', 'allow', '', '');
INSERT INTO `casbin_rule` (`id`, `ptype`, `v0`, `v1`, `v2`, `v3`, `v4`, `v5`) VALUES (8135, 'p', 'admin', '/api/model/relation/list', 'POST', 'allow', '', '');
INSERT INTO `casbin_rule` (`id`, `ptype`, `v0`, `v1`, `v2`, `v3`, `v4`, `v5`) VALUES (8122, 'p', 'admin', '/api/order/create', 'POST', 'allow', '', '');
INSERT INTO `casbin_rule` (`id`, `ptype`, `v0`, `v1`, `v2`, `v3`, `v4`, `v5`) VALUES (8125, 'p', 'admin', '/api/order/detail/process_inst_id', 'POST', 'allow', '', '');
INSERT INTO `casbin_rule` (`id`, `ptype`, `v0`, `v1`, `v2`, `v3`, `v4`, `v5`) VALUES (8183, 'p', 'admin', '/api/order/history', 'POST', 'allow', '', '');
INSERT INTO `casbin_rule` (`id`, `ptype`, `v0`, `v1`, `v2`, `v3`, `v4`, `v5`) VALUES (8130, 'p', 'admin', '/api/order/pass', 'POST', 'allow', '', '');
INSERT INTO `casbin_rule` (`id`, `ptype`, `v0`, `v1`, `v2`, `v3`, `v4`, `v5`) VALUES (8131, 'p', 'admin', '/api/order/reject', 'POST', 'allow', '', '');
INSERT INTO `casbin_rule` (`id`, `ptype`, `v0`, `v1`, `v2`, `v3`, `v4`, `v5`) VALUES (8184, 'p', 'admin', '/api/order/revoke', 'POST', 'allow', '', '');
INSERT INTO `casbin_rule` (`id`, `ptype`, `v0`, `v1`, `v2`, `v3`, `v4`, `v5`) VALUES (8124, 'p', 'admin', '/api/order/start/user', 'POST', 'allow', '', '');
INSERT INTO `casbin_rule` (`id`, `ptype`, `v0`, `v1`, `v2`, `v3`, `v4`, `v5`) VALUES (8126, 'p', 'admin', '/api/order/task/record', 'POST', 'allow', '', '');
INSERT INTO `casbin_rule` (`id`, `ptype`, `v0`, `v1`, `v2`, `v3`, `v4`, `v5`) VALUES (8129, 'p', 'admin', '/api/order/todo', 'POST', 'allow', '', '');
INSERT INTO `casbin_rule` (`id`, `ptype`, `v0`, `v1`, `v2`, `v3`, `v4`, `v5`) VALUES (8128, 'p', 'admin', '/api/order/todo/user', 'POST', 'allow', '', '');
INSERT INTO `casbin_rule` (`id`, `ptype`, `v0`, `v1`, `v2`, `v3`, `v4`, `v5`) VALUES (8090, 'p', 'admin', '/api/permission/change', 'POST', 'allow', '', '');
INSERT INTO `casbin_rule` (`id`, `ptype`, `v0`, `v1`, `v2`, `v3`, `v4`, `v5`) VALUES (8089, 'p', 'admin', '/api/permission/list', 'POST', 'allow', '', '');
INSERT INTO `casbin_rule` (`id`, `ptype`, `v0`, `v1`, `v2`, `v3`, `v4`, `v5`) VALUES (8153, 'p', 'admin', '/api/relation/create', 'POST', 'allow', '', '');
INSERT INTO `casbin_rule` (`id`, `ptype`, `v0`, `v1`, `v2`, `v3`, `v4`, `v5`) VALUES (8158, 'p', 'admin', '/api/relation/list', 'POST', 'allow', '', '');
INSERT INTO `casbin_rule` (`id`, `ptype`, `v0`, `v1`, `v2`, `v3`, `v4`, `v5`) VALUES (8133, 'p', 'admin', '/api/resource/create', 'POST', 'allow', '', '');
INSERT INTO `casbin_rule` (`id`, `ptype`, `v0`, `v1`, `v2`, `v3`, `v4`, `v5`) VALUES (8134, 'p', 'admin', '/api/resource/delete', 'POST', 'allow', '', '');
INSERT INTO `casbin_rule` (`id`, `ptype`, `v0`, `v1`, `v2`, `v3`, `v4`, `v5`) VALUES (8161, 'p', 'admin', '/api/resource/detail', 'POST', 'allow', '', '');
INSERT INTO `casbin_rule` (`id`, `ptype`, `v0`, `v1`, `v2`, `v3`, `v4`, `v5`) VALUES (8159, 'p', 'admin', '/api/resource/list', 'POST', 'allow', '', '');
INSERT INTO `casbin_rule` (`id`, `ptype`, `v0`, `v1`, `v2`, `v3`, `v4`, `v5`) VALUES (8137, 'p', 'admin', '/api/resource/list/ids', 'POST', 'allow', '', '');
INSERT INTO `casbin_rule` (`id`, `ptype`, `v0`, `v1`, `v2`, `v3`, `v4`, `v5`) VALUES (8141, 'p', 'admin', '/api/resource/relation/can_be_related', 'POST', 'allow', '', '');
INSERT INTO `casbin_rule` (`id`, `ptype`, `v0`, `v1`, `v2`, `v3`, `v4`, `v5`) VALUES (8142, 'p', 'admin', '/api/resource/relation/create', 'POST', 'allow', '', '');
INSERT INTO `casbin_rule` (`id`, `ptype`, `v0`, `v1`, `v2`, `v3`, `v4`, `v5`) VALUES (8143, 'p', 'admin', '/api/resource/relation/delete', 'POST', 'allow', '', '');
INSERT INTO `casbin_rule` (`id`, `ptype`, `v0`, `v1`, `v2`, `v3`, `v4`, `v5`) VALUES (8138, 'p', 'admin', '/api/resource/relation/graph', 'POST', 'allow', '', '');
INSERT INTO `casbin_rule` (`id`, `ptype`, `v0`, `v1`, `v2`, `v3`, `v4`, `v5`) VALUES (8140, 'p', 'admin', '/api/resource/relation/graph/add/left', 'POST', 'allow', '', '');
INSERT INTO `casbin_rule` (`id`, `ptype`, `v0`, `v1`, `v2`, `v3`, `v4`, `v5`) VALUES (8139, 'p', 'admin', '/api/resource/relation/graph/add/right', 'POST', 'allow', '', '');
INSERT INTO `casbin_rule` (`id`, `ptype`, `v0`, `v1`, `v2`, `v3`, `v4`, `v5`) VALUES (8136, 'p', 'admin', '/api/resource/relation/pipeline/all', 'POST', 'allow', '', '');
INSERT INTO `casbin_rule` (`id`, `ptype`, `v0`, `v1`, `v2`, `v3`, `v4`, `v5`) VALUES (8174, 'p', 'admin', '/api/resource/search', 'POST', 'allow', '', '');
INSERT INTO `casbin_rule` (`id`, `ptype`, `v0`, `v1`, `v2`, `v3`, `v4`, `v5`) VALUES (8132, 'p', 'admin', '/api/resource/secure', 'POST', 'allow', '', '');
INSERT INTO `casbin_rule` (`id`, `ptype`, `v0`, `v1`, `v2`, `v3`, `v4`, `v5`) VALUES (8176, 'p', 'admin', '/api/resource/update', 'POST', 'allow', '', '');
INSERT INTO `casbin_rule` (`id`, `ptype`, `v0`, `v1`, `v2`, `v3`, `v4`, `v5`) VALUES (8091, 'p', 'admin', '/api/role/create', 'POST', 'allow', '', '');
INSERT INTO `casbin_rule` (`id`, `ptype`, `v0`, `v1`, `v2`, `v3`, `v4`, `v5`) VALUES (8096, 'p', 'admin', '/api/role/delete', 'POST', 'allow', '', '');
INSERT INTO `casbin_rule` (`id`, `ptype`, `v0`, `v1`, `v2`, `v3`, `v4`, `v5`) VALUES (8170, 'p', 'admin', '/api/role/list', 'POST', 'allow', '', '');
INSERT INTO `casbin_rule` (`id`, `ptype`, `v0`, `v1`, `v2`, `v3`, `v4`, `v5`) VALUES (8092, 'p', 'admin', '/api/role/update', 'POST', 'allow', '', '');
INSERT INTO `casbin_rule` (`id`, `ptype`, `v0`, `v1`, `v2`, `v3`, `v4`, `v5`) VALUES (8099, 'p', 'admin', '/api/role/user/does_not_have', 'POST', 'allow', '', '');
INSERT INTO `casbin_rule` (`id`, `ptype`, `v0`, `v1`, `v2`, `v3`, `v4`, `v5`) VALUES (8098, 'p', 'admin', '/api/role/user/have', 'POST', 'allow', '', '');
INSERT INTO `casbin_rule` (`id`, `ptype`, `v0`, `v1`, `v2`, `v3`, `v4`, `v5`) VALUES (8199, 'p', 'admin', '/api/rota/create', 'POST', 'allow', '', '');
INSERT INTO `casbin_rule` (`id`, `ptype`, `v0`, `v1`, `v2`, `v3`, `v4`, `v5`) VALUES (8197, 'p', 'admin', '/api/rota/delete', 'POST', 'allow', '', '');
INSERT INTO `casbin_rule` (`id`, `ptype`, `v0`, `v1`, `v2`, `v3`, `v4`, `v5`) VALUES (8188, 'p', 'admin', '/api/rota/list', 'POST', 'allow', '', '');
INSERT INTO `casbin_rule` (`id`, `ptype`, `v0`, `v1`, `v2`, `v3`, `v4`, `v5`) VALUES (8192, 'p', 'admin', '/api/rota/rule/detail', 'POST', 'allow', '', '');
INSERT INTO `casbin_rule` (`id`, `ptype`, `v0`, `v1`, `v2`, `v3`, `v4`, `v5`) VALUES (8195, 'p', 'admin', '/api/rota/rule/shift_adjustment/add', 'POST', 'allow', '', '');
INSERT INTO `casbin_rule` (`id`, `ptype`, `v0`, `v1`, `v2`, `v3`, `v4`, `v5`) VALUES (8194, 'p', 'admin', '/api/rota/rule/shift_adjustment/delete', 'POST', 'allow', '', '');
INSERT INTO `casbin_rule` (`id`, `ptype`, `v0`, `v1`, `v2`, `v3`, `v4`, `v5`) VALUES (8196, 'p', 'admin', '/api/rota/rule/shift_adjustment/update', 'POST', 'allow', '', '');
INSERT INTO `casbin_rule` (`id`, `ptype`, `v0`, `v1`, `v2`, `v3`, `v4`, `v5`) VALUES (8190, 'p', 'admin', '/api/rota/rule/shift_scheduling/add', 'POST', 'allow', '', '');
INSERT INTO `casbin_rule` (`id`, `ptype`, `v0`, `v1`, `v2`, `v3`, `v4`, `v5`) VALUES (8193, 'p', 'admin', '/api/rota/rule/shift_scheduling/update', 'POST', 'allow', '', '');
INSERT INTO `casbin_rule` (`id`, `ptype`, `v0`, `v1`, `v2`, `v3`, `v4`, `v5`) VALUES (8189, 'p', 'admin', '/api/rota/schedule/preview', 'POST', 'allow', '', '');
INSERT INTO `casbin_rule` (`id`, `ptype`, `v0`, `v1`, `v2`, `v3`, `v4`, `v5`) VALUES (8198, 'p', 'admin', '/api/rota/update', 'POST', 'allow', '', '');
INSERT INTO `casbin_rule` (`id`, `ptype`, `v0`, `v1`, `v2`, `v3`, `v4`, `v5`) VALUES (8105, 'p', 'admin', '/api/runner/delete', 'POST', 'allow', '', '');
INSERT INTO `casbin_rule` (`id`, `ptype`, `v0`, `v1`, `v2`, `v3`, `v4`, `v5`) VALUES (8168, 'p', 'admin', '/api/runner/list', 'POST', 'allow', '', '');
INSERT INTO `casbin_rule` (`id`, `ptype`, `v0`, `v1`, `v2`, `v3`, `v4`, `v5`) VALUES (8114, 'p', 'admin', '/api/runner/list/tags', 'POST', 'allow', '', '');
INSERT INTO `casbin_rule` (`id`, `ptype`, `v0`, `v1`, `v2`, `v3`, `v4`, `v5`) VALUES (8103, 'p', 'admin', '/api/runner/register', 'POST', 'allow', '', '');
INSERT INTO `casbin_rule` (`id`, `ptype`, `v0`, `v1`, `v2`, `v3`, `v4`, `v5`) VALUES (8104, 'p', 'admin', '/api/runner/update', 'POST', 'allow', '', '');
INSERT INTO `casbin_rule` (`id`, `ptype`, `v0`, `v1`, `v2`, `v3`, `v4`, `v5`) VALUES (8169, 'p', 'admin', '/api/task/list', 'POST', 'allow', '', '');
INSERT INTO `casbin_rule` (`id`, `ptype`, `v0`, `v1`, `v2`, `v3`, `v4`, `v5`) VALUES (8108, 'p', 'admin', '/api/task/retry', 'POST', 'allow', '', '');
INSERT INTO `casbin_rule` (`id`, `ptype`, `v0`, `v1`, `v2`, `v3`, `v4`, `v5`) VALUES (8106, 'p', 'admin', '/api/task/update/args', 'POST', 'allow', '', '');
INSERT INTO `casbin_rule` (`id`, `ptype`, `v0`, `v1`, `v2`, `v3`, `v4`, `v5`) VALUES (8107, 'p', 'admin', '/api/task/update/variables', 'POST', 'allow', '', '');
INSERT INTO `casbin_rule` (`id`, `ptype`, `v0`, `v1`, `v2`, `v3`, `v4`, `v5`) VALUES (8109, 'p', 'admin', '/api/template/create', 'POST', 'allow', '', '');
INSERT INTO `casbin_rule` (`id`, `ptype`, `v0`, `v1`, `v2`, `v3`, `v4`, `v5`) VALUES (8112, 'p', 'admin', '/api/template/delete', 'POST', 'allow', '', '');
INSERT INTO `casbin_rule` (`id`, `ptype`, `v0`, `v1`, `v2`, `v3`, `v4`, `v5`) VALUES (8123, 'p', 'admin', '/api/template/detail', 'POST', 'allow', '', '');
INSERT INTO `casbin_rule` (`id`, `ptype`, `v0`, `v1`, `v2`, `v3`, `v4`, `v5`) VALUES (8110, 'p', 'admin', '/api/template/group/create', 'POST', 'allow', '', '');
INSERT INTO `casbin_rule` (`id`, `ptype`, `v0`, `v1`, `v2`, `v3`, `v4`, `v5`) VALUES (8165, 'p', 'admin', '/api/template/group/list', 'POST', 'allow', '', '');
INSERT INTO `casbin_rule` (`id`, `ptype`, `v0`, `v1`, `v2`, `v3`, `v4`, `v5`) VALUES (8163, 'p', 'admin', '/api/template/list', 'POST', 'allow', '', '');
INSERT INTO `casbin_rule` (`id`, `ptype`, `v0`, `v1`, `v2`, `v3`, `v4`, `v5`) VALUES (8162, 'p', 'admin', '/api/template/list/pipeline', 'POST', 'allow', '', '');
INSERT INTO `casbin_rule` (`id`, `ptype`, `v0`, `v1`, `v2`, `v3`, `v4`, `v5`) VALUES (8111, 'p', 'admin', '/api/template/update', 'POST', 'allow', '', '');
INSERT INTO `casbin_rule` (`id`, `ptype`, `v0`, `v1`, `v2`, `v3`, `v4`, `v5`) VALUES (8191, 'p', 'admin', '/api/user/find/by_ids', 'POST', 'allow', '', '');
INSERT INTO `casbin_rule` (`id`, `ptype`, `v0`, `v1`, `v2`, `v3`, `v4`, `v5`) VALUES (8118, 'p', 'admin', '/api/user/find/by_keyword', 'POST', 'allow', '', '');
INSERT INTO `casbin_rule` (`id`, `ptype`, `v0`, `v1`, `v2`, `v3`, `v4`, `v5`) VALUES (8181, 'p', 'admin', '/api/user/find/department_id', 'POST', 'allow', '', '');
INSERT INTO `casbin_rule` (`id`, `ptype`, `v0`, `v1`, `v2`, `v3`, `v4`, `v5`) VALUES (8200, 'p', 'admin', '/api/user/find/id', 'POST', 'allow', '', '');
INSERT INTO `casbin_rule` (`id`, `ptype`, `v0`, `v1`, `v2`, `v3`, `v4`, `v5`) VALUES (8117, 'p', 'admin', '/api/user/find/username', 'POST', 'allow', '', '');
INSERT INTO `casbin_rule` (`id`, `ptype`, `v0`, `v1`, `v2`, `v3`, `v4`, `v5`) VALUES (8116, 'p', 'admin', '/api/user/find/usernames', 'POST', 'allow', '', '');
INSERT INTO `casbin_rule` (`id`, `ptype`, `v0`, `v1`, `v2`, `v3`, `v4`, `v5`) VALUES (8088, 'p', 'admin', '/api/user/info', 'POST', 'allow', '', '');
INSERT INTO `casbin_rule` (`id`, `ptype`, `v0`, `v1`, `v2`, `v3`, `v4`, `v5`) VALUES (8187, 'p', 'admin', '/api/user/ldap/refresh_cache', 'POST', 'allow', '', '');
INSERT INTO `casbin_rule` (`id`, `ptype`, `v0`, `v1`, `v2`, `v3`, `v4`, `v5`) VALUES (8186, 'p', 'admin', '/api/user/ldap/search', 'POST', 'allow', '', '');
INSERT INTO `casbin_rule` (`id`, `ptype`, `v0`, `v1`, `v2`, `v3`, `v4`, `v5`) VALUES (8185, 'p', 'admin', '/api/user/ldap/sync', 'POST', 'allow', '', '');
INSERT INTO `casbin_rule` (`id`, `ptype`, `v0`, `v1`, `v2`, `v3`, `v4`, `v5`) VALUES (8173, 'p', 'admin', '/api/user/list', 'POST', 'allow', '', '');
INSERT INTO `casbin_rule` (`id`, `ptype`, `v0`, `v1`, `v2`, `v3`, `v4`, `v5`) VALUES (8115, 'p', 'admin', '/api/user/pipeline/department_id', 'POST', 'allow', '', '');
INSERT INTO `casbin_rule` (`id`, `ptype`, `v0`, `v1`, `v2`, `v3`, `v4`, `v5`) VALUES (8097, 'p', 'admin', '/api/user/role/bind', 'POST', 'allow', '', '');
INSERT INTO `casbin_rule` (`id`, `ptype`, `v0`, `v1`, `v2`, `v3`, `v4`, `v5`) VALUES (8182, 'p', 'admin', '/api/user/update', 'POST', 'allow', '', '');
INSERT INTO `casbin_rule` (`id`, `ptype`, `v0`, `v1`, `v2`, `v3`, `v4`, `v5`) VALUES (8167, 'p', 'admin', '/api/worker/list', 'POST', 'allow', '', '');
INSERT INTO `casbin_rule` (`id`, `ptype`, `v0`, `v1`, `v2`, `v3`, `v4`, `v5`) VALUES (8113, 'p', 'admin', '/api/workflow/create', 'POST', 'allow', '', '');
INSERT INTO `casbin_rule` (`id`, `ptype`, `v0`, `v1`, `v2`, `v3`, `v4`, `v5`) VALUES (8121, 'p', 'admin', '/api/workflow/delete', 'POST', 'allow', '', '');
INSERT INTO `casbin_rule` (`id`, `ptype`, `v0`, `v1`, `v2`, `v3`, `v4`, `v5`) VALUES (8119, 'p', 'admin', '/api/workflow/deploy', 'POST', 'allow', '', '');
INSERT INTO `casbin_rule` (`id`, `ptype`, `v0`, `v1`, `v2`, `v3`, `v4`, `v5`) VALUES (8127, 'p', 'admin', '/api/workflow/graph', 'POST', 'allow', '', '');
INSERT INTO `casbin_rule` (`id`, `ptype`, `v0`, `v1`, `v2`, `v3`, `v4`, `v5`) VALUES (8164, 'p', 'admin', '/api/workflow/list', 'POST', 'allow', '', '');
INSERT INTO `casbin_rule` (`id`, `ptype`, `v0`, `v1`, `v2`, `v3`, `v4`, `v5`) VALUES (8120, 'p', 'admin', '/api/workflow/update', 'POST', 'allow', '', '');
INSERT INTO `casbin_rule` (`id`, `ptype`, `v0`, `v1`, `v2`, `v3`, `v4`, `v5`) VALUES (2570, 'p', 'audit', '/api/model/group/create', 'POST', 'allow', '', '');
INSERT INTO `casbin_rule` (`id`, `ptype`, `v0`, `v1`, `v2`, `v3`, `v4`, `v5`) VALUES (4533, 'p', 'dev', '/api/attribute/list', 'POST', 'allow', '', '');
INSERT INTO `casbin_rule` (`id`, `ptype`, `v0`, `v1`, `v2`, `v3`, `v4`, `v5`) VALUES (4537, 'p', 'dev', '/api/attribute/list/field', 'POST', 'allow', '', '');
INSERT INTO `casbin_rule` (`id`, `ptype`, `v0`, `v1`, `v2`, `v3`, `v4`, `v5`) VALUES (4534, 'p', 'dev', '/api/model/by_group', 'POST', 'allow', '', '');
INSERT INTO `casbin_rule` (`id`, `ptype`, `v0`, `v1`, `v2`, `v3`, `v4`, `v5`) VALUES (4532, 'p', 'dev', '/api/model/relation/graph', 'POST', 'allow', '', '');
INSERT INTO `casbin_rule` (`id`, `ptype`, `v0`, `v1`, `v2`, `v3`, `v4`, `v5`) VALUES (4523, 'p', 'dev', '/api/model/relation/list', 'POST', 'allow', '', '');
INSERT INTO `casbin_rule` (`id`, `ptype`, `v0`, `v1`, `v2`, `v3`, `v4`, `v5`) VALUES (4519, 'p', 'dev', '/api/order/create', 'POST', 'allow', '', '');
INSERT INTO `casbin_rule` (`id`, `ptype`, `v0`, `v1`, `v2`, `v3`, `v4`, `v5`) VALUES (4535, 'p', 'dev', '/api/relation/list', 'POST', 'allow', '', '');
INSERT INTO `casbin_rule` (`id`, `ptype`, `v0`, `v1`, `v2`, `v3`, `v4`, `v5`) VALUES (4522, 'p', 'dev', '/api/resource/create', 'POST', 'allow', '', '');
INSERT INTO `casbin_rule` (`id`, `ptype`, `v0`, `v1`, `v2`, `v3`, `v4`, `v5`) VALUES (4538, 'p', 'dev', '/api/resource/detail', 'POST', 'allow', '', '');
INSERT INTO `casbin_rule` (`id`, `ptype`, `v0`, `v1`, `v2`, `v3`, `v4`, `v5`) VALUES (4536, 'p', 'dev', '/api/resource/list', 'POST', 'allow', '', '');
INSERT INTO `casbin_rule` (`id`, `ptype`, `v0`, `v1`, `v2`, `v3`, `v4`, `v5`) VALUES (4525, 'p', 'dev', '/api/resource/list/ids', 'POST', 'allow', '', '');
INSERT INTO `casbin_rule` (`id`, `ptype`, `v0`, `v1`, `v2`, `v3`, `v4`, `v5`) VALUES (4529, 'p', 'dev', '/api/resource/relation/can_be_related', 'POST', 'allow', '', '');
INSERT INTO `casbin_rule` (`id`, `ptype`, `v0`, `v1`, `v2`, `v3`, `v4`, `v5`) VALUES (4530, 'p', 'dev', '/api/resource/relation/create', 'POST', 'allow', '', '');
INSERT INTO `casbin_rule` (`id`, `ptype`, `v0`, `v1`, `v2`, `v3`, `v4`, `v5`) VALUES (4531, 'p', 'dev', '/api/resource/relation/delete', 'POST', 'allow', '', '');
INSERT INTO `casbin_rule` (`id`, `ptype`, `v0`, `v1`, `v2`, `v3`, `v4`, `v5`) VALUES (4526, 'p', 'dev', '/api/resource/relation/graph', 'POST', 'allow', '', '');
INSERT INTO `casbin_rule` (`id`, `ptype`, `v0`, `v1`, `v2`, `v3`, `v4`, `v5`) VALUES (4528, 'p', 'dev', '/api/resource/relation/graph/add/left', 'POST', 'allow', '', '');
INSERT INTO `casbin_rule` (`id`, `ptype`, `v0`, `v1`, `v2`, `v3`, `v4`, `v5`) VALUES (4527, 'p', 'dev', '/api/resource/relation/graph/add/right', 'POST', 'allow', '', '');
INSERT INTO `casbin_rule` (`id`, `ptype`, `v0`, `v1`, `v2`, `v3`, `v4`, `v5`) VALUES (4524, 'p', 'dev', '/api/resource/relation/pipeline/all', 'POST', 'allow', '', '');
INSERT INTO `casbin_rule` (`id`, `ptype`, `v0`, `v1`, `v2`, `v3`, `v4`, `v5`) VALUES (4540, 'p', 'dev', '/api/resource/search', 'POST', 'allow', '', '');
INSERT INTO `casbin_rule` (`id`, `ptype`, `v0`, `v1`, `v2`, `v3`, `v4`, `v5`) VALUES (4521, 'p', 'dev', '/api/resource/secure', 'POST', 'allow', '', '');
INSERT INTO `casbin_rule` (`id`, `ptype`, `v0`, `v1`, `v2`, `v3`, `v4`, `v5`) VALUES (4541, 'p', 'dev', '/api/resource/update', 'POST', 'allow', '', '');
INSERT INTO `casbin_rule` (`id`, `ptype`, `v0`, `v1`, `v2`, `v3`, `v4`, `v5`) VALUES (4520, 'p', 'dev', '/api/template/detail', 'POST', 'allow', '', '');
INSERT INTO `casbin_rule` (`id`, `ptype`, `v0`, `v1`, `v2`, `v3`, `v4`, `v5`) VALUES (4539, 'p', 'dev', '/api/template/list/pipeline', 'POST', 'allow', '', '');
INSERT INTO `casbin_rule` (`id`, `ptype`, `v0`, `v1`, `v2`, `v3`, `v4`, `v5`) VALUES (4638, 'p', 'dev', '/api/user/info', 'POST', 'allow', '', '');
COMMIT;

SET FOREIGN_KEY_CHECKS = 1;
