/*
 Navicat Premium Data Transfer

 Source Server         : mysql
 Source Server Type    : MySQL
 Source Server Version : 80037
 Source Host           : 127.0.0.1:3306
 Source Schema         : cmdb

 Target Server Type    : MySQL
 Target Server Version : 80037
 File Encoding         : 65001

 Date: 19/08/2024 15:38:44
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
) ENGINE=InnoDB AUTO_INCREMENT=2571 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;

-- ----------------------------
-- Records of casbin_rule
-- ----------------------------
BEGIN;
INSERT INTO `casbin_rule` (`id`, `ptype`, `v0`, `v1`, `v2`, `v3`, `v4`, `v5`) VALUES (2225, 'g', '1', 'admin', '', '', '', '');
INSERT INTO `casbin_rule` (`id`, `ptype`, `v0`, `v1`, `v2`, `v3`, `v4`, `v5`) VALUES (2442, 'p', 'admin', '/api/attribute/create', 'POST', 'allow', '', '');
INSERT INTO `casbin_rule` (`id`, `ptype`, `v0`, `v1`, `v2`, `v3`, `v4`, `v5`) VALUES (2447, 'p', 'admin', '/api/attribute/custom/field', 'POST', 'allow', '', '');
INSERT INTO `casbin_rule` (`id`, `ptype`, `v0`, `v1`, `v2`, `v3`, `v4`, `v5`) VALUES (2445, 'p', 'admin', '/api/attribute/delete', 'POST', 'allow', '', '');
INSERT INTO `casbin_rule` (`id`, `ptype`, `v0`, `v1`, `v2`, `v3`, `v4`, `v5`) VALUES (2441, 'p', 'admin', '/api/attribute/group/create', 'POST', 'allow', '', '');
INSERT INTO `casbin_rule` (`id`, `ptype`, `v0`, `v1`, `v2`, `v3`, `v4`, `v5`) VALUES (2449, 'p', 'admin', '/api/attribute/list', 'POST', 'allow', '', '');
INSERT INTO `casbin_rule` (`id`, `ptype`, `v0`, `v1`, `v2`, `v3`, `v4`, `v5`) VALUES (2453, 'p', 'admin', '/api/attribute/list/field', 'POST', 'allow', '', '');
INSERT INTO `casbin_rule` (`id`, `ptype`, `v0`, `v1`, `v2`, `v3`, `v4`, `v5`) VALUES (2399, 'p', 'admin', '/api/codebook/create', 'POST', 'allow', '', '');
INSERT INTO `casbin_rule` (`id`, `ptype`, `v0`, `v1`, `v2`, `v3`, `v4`, `v5`) VALUES (2401, 'p', 'admin', '/api/codebook/delete', 'POST', 'allow', '', '');
INSERT INTO `casbin_rule` (`id`, `ptype`, `v0`, `v1`, `v2`, `v3`, `v4`, `v5`) VALUES (2459, 'p', 'admin', '/api/codebook/list', 'POST', 'allow', '', '');
INSERT INTO `casbin_rule` (`id`, `ptype`, `v0`, `v1`, `v2`, `v3`, `v4`, `v5`) VALUES (2400, 'p', 'admin', '/api/codebook/update', 'POST', 'allow', '', '');
INSERT INTO `casbin_rule` (`id`, `ptype`, `v0`, `v1`, `v2`, `v3`, `v4`, `v5`) VALUES (2465, 'p', 'admin', '/api/endpoint/list', 'POST', 'allow', '', '');
INSERT INTO `casbin_rule` (`id`, `ptype`, `v0`, `v1`, `v2`, `v3`, `v4`, `v5`) VALUES (2392, 'p', 'admin', '/api/menu/create', 'POST', 'allow', '', '');
INSERT INTO `casbin_rule` (`id`, `ptype`, `v0`, `v1`, `v2`, `v3`, `v4`, `v5`) VALUES (2394, 'p', 'admin', '/api/menu/delete', 'POST', 'allow', '', '');
INSERT INTO `casbin_rule` (`id`, `ptype`, `v0`, `v1`, `v2`, `v3`, `v4`, `v5`) VALUES (2464, 'p', 'admin', '/api/menu/list/tree', 'POST', 'allow', '', '');
INSERT INTO `casbin_rule` (`id`, `ptype`, `v0`, `v1`, `v2`, `v3`, `v4`, `v5`) VALUES (2393, 'p', 'admin', '/api/menu/update', 'POST', 'allow', '', '');
INSERT INTO `casbin_rule` (`id`, `ptype`, `v0`, `v1`, `v2`, `v3`, `v4`, `v5`) VALUES (2438, 'p', 'admin', '/api/model/create', 'POST', 'allow', '', '');
INSERT INTO `casbin_rule` (`id`, `ptype`, `v0`, `v1`, `v2`, `v3`, `v4`, `v5`) VALUES (2440, 'p', 'admin', '/api/model/delete', 'POST', 'allow', '', '');
INSERT INTO `casbin_rule` (`id`, `ptype`, `v0`, `v1`, `v2`, `v3`, `v4`, `v5`) VALUES (2439, 'p', 'admin', '/api/model/group/create', 'POST', 'allow', '', '');
INSERT INTO `casbin_rule` (`id`, `ptype`, `v0`, `v1`, `v2`, `v3`, `v4`, `v5`) VALUES (2450, 'p', 'admin', '/api/model/list/pipeline', 'POST', 'allow', '', '');
INSERT INTO `casbin_rule` (`id`, `ptype`, `v0`, `v1`, `v2`, `v3`, `v4`, `v5`) VALUES (2443, 'p', 'admin', '/api/model/relation/create', 'POST', 'allow', '', '');
INSERT INTO `casbin_rule` (`id`, `ptype`, `v0`, `v1`, `v2`, `v3`, `v4`, `v5`) VALUES (2444, 'p', 'admin', '/api/model/relation/delete', 'POST', 'allow', '', '');
INSERT INTO `casbin_rule` (`id`, `ptype`, `v0`, `v1`, `v2`, `v3`, `v4`, `v5`) VALUES (2448, 'p', 'admin', '/api/model/relation/graph', 'POST', 'allow', '', '');
INSERT INTO `casbin_rule` (`id`, `ptype`, `v0`, `v1`, `v2`, `v3`, `v4`, `v5`) VALUES (2429, 'p', 'admin', '/api/model/relation/list', 'POST', 'allow', '', '');
INSERT INTO `casbin_rule` (`id`, `ptype`, `v0`, `v1`, `v2`, `v3`, `v4`, `v5`) VALUES (2416, 'p', 'admin', '/api/order/create', 'POST', 'allow', '', '');
INSERT INTO `casbin_rule` (`id`, `ptype`, `v0`, `v1`, `v2`, `v3`, `v4`, `v5`) VALUES (2419, 'p', 'admin', '/api/order/detail/process_inst_id', 'POST', 'allow', '', '');
INSERT INTO `casbin_rule` (`id`, `ptype`, `v0`, `v1`, `v2`, `v3`, `v4`, `v5`) VALUES (2424, 'p', 'admin', '/api/order/pass', 'POST', 'allow', '', '');
INSERT INTO `casbin_rule` (`id`, `ptype`, `v0`, `v1`, `v2`, `v3`, `v4`, `v5`) VALUES (2425, 'p', 'admin', '/api/order/reject', 'POST', 'allow', '', '');
INSERT INTO `casbin_rule` (`id`, `ptype`, `v0`, `v1`, `v2`, `v3`, `v4`, `v5`) VALUES (2418, 'p', 'admin', '/api/order/start/user', 'POST', 'allow', '', '');
INSERT INTO `casbin_rule` (`id`, `ptype`, `v0`, `v1`, `v2`, `v3`, `v4`, `v5`) VALUES (2420, 'p', 'admin', '/api/order/task/record', 'POST', 'allow', '', '');
INSERT INTO `casbin_rule` (`id`, `ptype`, `v0`, `v1`, `v2`, `v3`, `v4`, `v5`) VALUES (2423, 'p', 'admin', '/api/order/todo', 'POST', 'allow', '', '');
INSERT INTO `casbin_rule` (`id`, `ptype`, `v0`, `v1`, `v2`, `v3`, `v4`, `v5`) VALUES (2422, 'p', 'admin', '/api/order/todo/user', 'POST', 'allow', '', '');
INSERT INTO `casbin_rule` (`id`, `ptype`, `v0`, `v1`, `v2`, `v3`, `v4`, `v5`) VALUES (2389, 'p', 'admin', '/api/permission/change', 'POST', 'allow', '', '');
INSERT INTO `casbin_rule` (`id`, `ptype`, `v0`, `v1`, `v2`, `v3`, `v4`, `v5`) VALUES (2388, 'p', 'admin', '/api/permission/list', 'POST', 'allow', '', '');
INSERT INTO `casbin_rule` (`id`, `ptype`, `v0`, `v1`, `v2`, `v3`, `v4`, `v5`) VALUES (2446, 'p', 'admin', '/api/relation/create', 'POST', 'allow', '', '');
INSERT INTO `casbin_rule` (`id`, `ptype`, `v0`, `v1`, `v2`, `v3`, `v4`, `v5`) VALUES (2451, 'p', 'admin', '/api/relation/list', 'POST', 'allow', '', '');
INSERT INTO `casbin_rule` (`id`, `ptype`, `v0`, `v1`, `v2`, `v3`, `v4`, `v5`) VALUES (2427, 'p', 'admin', '/api/resource/create', 'POST', 'allow', '', '');
INSERT INTO `casbin_rule` (`id`, `ptype`, `v0`, `v1`, `v2`, `v3`, `v4`, `v5`) VALUES (2428, 'p', 'admin', '/api/resource/delete', 'POST', 'allow', '', '');
INSERT INTO `casbin_rule` (`id`, `ptype`, `v0`, `v1`, `v2`, `v3`, `v4`, `v5`) VALUES (2454, 'p', 'admin', '/api/resource/detail', 'POST', 'allow', '', '');
INSERT INTO `casbin_rule` (`id`, `ptype`, `v0`, `v1`, `v2`, `v3`, `v4`, `v5`) VALUES (2452, 'p', 'admin', '/api/resource/list', 'POST', 'allow', '', '');
INSERT INTO `casbin_rule` (`id`, `ptype`, `v0`, `v1`, `v2`, `v3`, `v4`, `v5`) VALUES (2431, 'p', 'admin', '/api/resource/list/ids', 'POST', 'allow', '', '');
INSERT INTO `casbin_rule` (`id`, `ptype`, `v0`, `v1`, `v2`, `v3`, `v4`, `v5`) VALUES (2435, 'p', 'admin', '/api/resource/relation/can_be_related', 'POST', 'allow', '', '');
INSERT INTO `casbin_rule` (`id`, `ptype`, `v0`, `v1`, `v2`, `v3`, `v4`, `v5`) VALUES (2436, 'p', 'admin', '/api/resource/relation/create', 'POST', 'allow', '', '');
INSERT INTO `casbin_rule` (`id`, `ptype`, `v0`, `v1`, `v2`, `v3`, `v4`, `v5`) VALUES (2437, 'p', 'admin', '/api/resource/relation/delete', 'POST', 'allow', '', '');
INSERT INTO `casbin_rule` (`id`, `ptype`, `v0`, `v1`, `v2`, `v3`, `v4`, `v5`) VALUES (2432, 'p', 'admin', '/api/resource/relation/graph', 'POST', 'allow', '', '');
INSERT INTO `casbin_rule` (`id`, `ptype`, `v0`, `v1`, `v2`, `v3`, `v4`, `v5`) VALUES (2434, 'p', 'admin', '/api/resource/relation/graph/add/left', 'POST', 'allow', '', '');
INSERT INTO `casbin_rule` (`id`, `ptype`, `v0`, `v1`, `v2`, `v3`, `v4`, `v5`) VALUES (2433, 'p', 'admin', '/api/resource/relation/graph/add/right', 'POST', 'allow', '', '');
INSERT INTO `casbin_rule` (`id`, `ptype`, `v0`, `v1`, `v2`, `v3`, `v4`, `v5`) VALUES (2430, 'p', 'admin', '/api/resource/relation/pipeline/all', 'POST', 'allow', '', '');
INSERT INTO `casbin_rule` (`id`, `ptype`, `v0`, `v1`, `v2`, `v3`, `v4`, `v5`) VALUES (2426, 'p', 'admin', '/api/resource/secure', 'POST', 'allow', '', '');
INSERT INTO `casbin_rule` (`id`, `ptype`, `v0`, `v1`, `v2`, `v3`, `v4`, `v5`) VALUES (2390, 'p', 'admin', '/api/role/create', 'POST', 'allow', '', '');
INSERT INTO `casbin_rule` (`id`, `ptype`, `v0`, `v1`, `v2`, `v3`, `v4`, `v5`) VALUES (2395, 'p', 'admin', '/api/role/delete', 'POST', 'allow', '', '');
INSERT INTO `casbin_rule` (`id`, `ptype`, `v0`, `v1`, `v2`, `v3`, `v4`, `v5`) VALUES (2463, 'p', 'admin', '/api/role/list', 'POST', 'allow', '', '');
INSERT INTO `casbin_rule` (`id`, `ptype`, `v0`, `v1`, `v2`, `v3`, `v4`, `v5`) VALUES (2391, 'p', 'admin', '/api/role/update', 'POST', 'allow', '', '');
INSERT INTO `casbin_rule` (`id`, `ptype`, `v0`, `v1`, `v2`, `v3`, `v4`, `v5`) VALUES (2398, 'p', 'admin', '/api/role/user/does_not_have', 'POST', 'allow', '', '');
INSERT INTO `casbin_rule` (`id`, `ptype`, `v0`, `v1`, `v2`, `v3`, `v4`, `v5`) VALUES (2397, 'p', 'admin', '/api/role/user/have', 'POST', 'allow', '', '');
INSERT INTO `casbin_rule` (`id`, `ptype`, `v0`, `v1`, `v2`, `v3`, `v4`, `v5`) VALUES (2404, 'p', 'admin', '/api/runner/delete', 'POST', 'allow', '', '');
INSERT INTO `casbin_rule` (`id`, `ptype`, `v0`, `v1`, `v2`, `v3`, `v4`, `v5`) VALUES (2461, 'p', 'admin', '/api/runner/list', 'POST', 'allow', '', '');
INSERT INTO `casbin_rule` (`id`, `ptype`, `v0`, `v1`, `v2`, `v3`, `v4`, `v5`) VALUES (2402, 'p', 'admin', '/api/runner/register', 'POST', 'allow', '', '');
INSERT INTO `casbin_rule` (`id`, `ptype`, `v0`, `v1`, `v2`, `v3`, `v4`, `v5`) VALUES (2403, 'p', 'admin', '/api/runner/update', 'POST', 'allow', '', '');
INSERT INTO `casbin_rule` (`id`, `ptype`, `v0`, `v1`, `v2`, `v3`, `v4`, `v5`) VALUES (2462, 'p', 'admin', '/api/task/list', 'POST', 'allow', '', '');
INSERT INTO `casbin_rule` (`id`, `ptype`, `v0`, `v1`, `v2`, `v3`, `v4`, `v5`) VALUES (2407, 'p', 'admin', '/api/task/retry', 'POST', 'allow', '', '');
INSERT INTO `casbin_rule` (`id`, `ptype`, `v0`, `v1`, `v2`, `v3`, `v4`, `v5`) VALUES (2405, 'p', 'admin', '/api/task/update/args', 'POST', 'allow', '', '');
INSERT INTO `casbin_rule` (`id`, `ptype`, `v0`, `v1`, `v2`, `v3`, `v4`, `v5`) VALUES (2406, 'p', 'admin', '/api/task/update/variables', 'POST', 'allow', '', '');
INSERT INTO `casbin_rule` (`id`, `ptype`, `v0`, `v1`, `v2`, `v3`, `v4`, `v5`) VALUES (2408, 'p', 'admin', '/api/template/create', 'POST', 'allow', '', '');
INSERT INTO `casbin_rule` (`id`, `ptype`, `v0`, `v1`, `v2`, `v3`, `v4`, `v5`) VALUES (2411, 'p', 'admin', '/api/template/delete', 'POST', 'allow', '', '');
INSERT INTO `casbin_rule` (`id`, `ptype`, `v0`, `v1`, `v2`, `v3`, `v4`, `v5`) VALUES (2417, 'p', 'admin', '/api/template/detail', 'POST', 'allow', '', '');
INSERT INTO `casbin_rule` (`id`, `ptype`, `v0`, `v1`, `v2`, `v3`, `v4`, `v5`) VALUES (2409, 'p', 'admin', '/api/template/group/create', 'POST', 'allow', '', '');
INSERT INTO `casbin_rule` (`id`, `ptype`, `v0`, `v1`, `v2`, `v3`, `v4`, `v5`) VALUES (2458, 'p', 'admin', '/api/template/group/list', 'POST', 'allow', '', '');
INSERT INTO `casbin_rule` (`id`, `ptype`, `v0`, `v1`, `v2`, `v3`, `v4`, `v5`) VALUES (2456, 'p', 'admin', '/api/template/list', 'POST', 'allow', '', '');
INSERT INTO `casbin_rule` (`id`, `ptype`, `v0`, `v1`, `v2`, `v3`, `v4`, `v5`) VALUES (2455, 'p', 'admin', '/api/template/list/pipeline', 'POST', 'allow', '', '');
INSERT INTO `casbin_rule` (`id`, `ptype`, `v0`, `v1`, `v2`, `v3`, `v4`, `v5`) VALUES (2410, 'p', 'admin', '/api/template/update', 'POST', 'allow', '', '');
INSERT INTO `casbin_rule` (`id`, `ptype`, `v0`, `v1`, `v2`, `v3`, `v4`, `v5`) VALUES (2466, 'p', 'admin', '/api/user/list', 'POST', 'allow', '', '');
INSERT INTO `casbin_rule` (`id`, `ptype`, `v0`, `v1`, `v2`, `v3`, `v4`, `v5`) VALUES (2396, 'p', 'admin', '/api/user/role/bind', 'POST', 'allow', '', '');
INSERT INTO `casbin_rule` (`id`, `ptype`, `v0`, `v1`, `v2`, `v3`, `v4`, `v5`) VALUES (2460, 'p', 'admin', '/api/worker/list', 'POST', 'allow', '', '');
INSERT INTO `casbin_rule` (`id`, `ptype`, `v0`, `v1`, `v2`, `v3`, `v4`, `v5`) VALUES (2412, 'p', 'admin', '/api/workflow/create', 'POST', 'allow', '', '');
INSERT INTO `casbin_rule` (`id`, `ptype`, `v0`, `v1`, `v2`, `v3`, `v4`, `v5`) VALUES (2415, 'p', 'admin', '/api/workflow/delete', 'POST', 'allow', '', '');
INSERT INTO `casbin_rule` (`id`, `ptype`, `v0`, `v1`, `v2`, `v3`, `v4`, `v5`) VALUES (2413, 'p', 'admin', '/api/workflow/deploy', 'POST', 'allow', '', '');
INSERT INTO `casbin_rule` (`id`, `ptype`, `v0`, `v1`, `v2`, `v3`, `v4`, `v5`) VALUES (2421, 'p', 'admin', '/api/workflow/graph', 'POST', 'allow', '', '');
INSERT INTO `casbin_rule` (`id`, `ptype`, `v0`, `v1`, `v2`, `v3`, `v4`, `v5`) VALUES (2457, 'p', 'admin', '/api/workflow/list', 'POST', 'allow', '', '');
INSERT INTO `casbin_rule` (`id`, `ptype`, `v0`, `v1`, `v2`, `v3`, `v4`, `v5`) VALUES (2414, 'p', 'admin', '/api/workflow/update', 'POST', 'allow', '', '');
INSERT INTO `casbin_rule` (`id`, `ptype`, `v0`, `v1`, `v2`, `v3`, `v4`, `v5`) VALUES (2570, 'p', 'audit', '/api/model/group/create', 'POST', 'allow', '', '');
COMMIT;

SET FOREIGN_KEY_CHECKS = 1;
