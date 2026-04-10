/*
Navicat MySQL Data Transfer

Source Server         : localhost
Source Server Version : 50624
Source Host           : 127.0.0.1:3306
Source Database       : bw_user_base

Target Server Type    : MYSQL
Target Server Version : 50624
File Encoding         : 65001

Date: 2016-12-13 11:51:43
*/

SET FOREIGN_KEY_CHECKS=0;

-- ----------------------------
-- Table structure for t_api
-- ----------------------------
DROP TABLE IF EXISTS `t_api`;
CREATE TABLE `t_api` (
  `id` bigint(20) NOT NULL AUTO_INCREMENT,
  `create_date` datetime DEFAULT NULL,
  `create_user_id` varchar(255) DEFAULT NULL,
  `entity_no` varchar(255) DEFAULT NULL,
  `model_status` varchar(255) DEFAULT NULL,
  `modify_date` datetime DEFAULT NULL,
  `modify_user_id` varchar(255) DEFAULT NULL,
  `product_id` varchar(255) DEFAULT NULL,
  `tenant_id` varchar(255) DEFAULT NULL,
  `comment` varchar(255) DEFAULT NULL,
  `name` varchar(255) DEFAULT NULL,
  `service_name` varchar(255) DEFAULT NULL,
  `url` varchar(255) DEFAULT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;

-- ----------------------------
-- Records of t_api
-- ----------------------------

-- ----------------------------
-- Table structure for t_api_field_right
-- ----------------------------
DROP TABLE IF EXISTS `t_api_field_right`;
CREATE TABLE `t_api_field_right` (
  `id` bigint(20) NOT NULL AUTO_INCREMENT,
  `create_date` datetime DEFAULT NULL,
  `create_user_id` varchar(255) DEFAULT NULL,
  `entity_no` varchar(255) DEFAULT NULL,
  `model_status` varchar(255) DEFAULT NULL,
  `modify_date` datetime DEFAULT NULL,
  `modify_user_id` varchar(255) DEFAULT NULL,
  `product_id` varchar(255) DEFAULT NULL,
  `tenant_id` varchar(255) DEFAULT NULL,
  `field_name` varchar(255) DEFAULT NULL,
  `field_right_type` int(11) DEFAULT NULL,
  `need` bit(1) NOT NULL,
  `right_entity_no` varchar(255) DEFAULT NULL,
  `api` bigint(20) DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `FK_j18jeo9h5cq0js49w87wkydmc` (`api`),
  CONSTRAINT `FK_j18jeo9h5cq0js49w87wkydmc` FOREIGN KEY (`api`) REFERENCES `t_api` (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;

-- ----------------------------
-- Records of t_api_field_right
-- ----------------------------

-- ----------------------------
-- Table structure for t_api_res_filter
-- ----------------------------
DROP TABLE IF EXISTS `t_api_res_filter`;
CREATE TABLE `t_api_res_filter` (
  `id` bigint(20) NOT NULL AUTO_INCREMENT,
  `create_date` datetime DEFAULT NULL,
  `create_user_id` varchar(255) DEFAULT NULL,
  `entity_no` varchar(255) DEFAULT NULL,
  `model_status` varchar(255) DEFAULT NULL,
  `modify_date` datetime DEFAULT NULL,
  `modify_user_id` varchar(255) DEFAULT NULL,
  `product_id` varchar(255) DEFAULT NULL,
  `tenant_id` varchar(255) DEFAULT NULL,
  `need` bit(1) NOT NULL,
  `rep_reg` varchar(255) DEFAULT NULL,
  `right_entity_no` varchar(255) DEFAULT NULL,
  `src_reg` varchar(255) DEFAULT NULL,
  `api` bigint(20) DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `FK_nlj329fo2cq205txjoa85ovpt` (`api`),
  CONSTRAINT `FK_nlj329fo2cq205txjoa85ovpt` FOREIGN KEY (`api`) REFERENCES `t_api` (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;

-- ----------------------------
-- Records of t_api_res_filter
-- ----------------------------

-- ----------------------------
-- Table structure for t_api_right
-- ----------------------------
DROP TABLE IF EXISTS `t_api_right`;
CREATE TABLE `t_api_right` (
  `id` bigint(20) NOT NULL AUTO_INCREMENT,
  `create_date` datetime DEFAULT NULL,
  `create_user_id` varchar(255) DEFAULT NULL,
  `entity_no` varchar(255) DEFAULT NULL,
  `model_status` varchar(255) DEFAULT NULL,
  `modify_date` datetime DEFAULT NULL,
  `modify_user_id` varchar(255) DEFAULT NULL,
  `product_id` varchar(255) DEFAULT NULL,
  `tenant_id` varchar(255) DEFAULT NULL,
  `right_entity_no` varchar(255) DEFAULT NULL,
  `right_type` varchar(255) DEFAULT NULL,
  `url` varchar(255) DEFAULT NULL,
  `api` bigint(20) DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `FK_rt318fjdu8bh44lsbop6u1n7y` (`api`),
  CONSTRAINT `FK_rt318fjdu8bh44lsbop6u1n7y` FOREIGN KEY (`api`) REFERENCES `t_api` (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;

-- ----------------------------
-- Records of t_api_right
-- ----------------------------

-- ----------------------------
-- Table structure for t_exception
-- ----------------------------
DROP TABLE IF EXISTS `t_exception`;
CREATE TABLE `t_exception` (
  `id` bigint(20) NOT NULL AUTO_INCREMENT,
  `create_date` datetime DEFAULT NULL,
  `create_user_id` varchar(255) DEFAULT NULL,
  `entity_no` varchar(255) DEFAULT NULL,
  `model_status` varchar(255) DEFAULT NULL,
  `modify_date` datetime DEFAULT NULL,
  `modify_user_id` varchar(255) DEFAULT NULL,
  `product_id` varchar(255) DEFAULT NULL,
  `tenant_id` varchar(255) DEFAULT NULL,
  `exception` text DEFAULT NULL,
  `tenant` varchar(255) DEFAULT NULL,
  `uri` varchar(255) DEFAULT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;

-- ----------------------------
-- Records of t_exception
-- ----------------------------

-- ----------------------------
-- Table structure for t_level
-- ----------------------------
DROP TABLE IF EXISTS `t_level`;
CREATE TABLE `t_level` (
  `id` bigint(20) NOT NULL AUTO_INCREMENT,
  `create_date` datetime DEFAULT NULL,
  `create_user_id` varchar(255) DEFAULT NULL,
  `entity_no` varchar(255) DEFAULT NULL,
  `model_status` varchar(255) DEFAULT NULL,
  `modify_date` datetime DEFAULT NULL,
  `modify_user_id` varchar(255) DEFAULT NULL,
  `product_id` varchar(255) DEFAULT NULL,
  `tenant_id` varchar(255) DEFAULT NULL,
  `comment` varchar(255) DEFAULT NULL,
  `name` varchar(255) DEFAULT NULL,
  `sid` int(11) NOT NULL,
  `user_count` int(11) NOT NULL,
  `parent_id` bigint(20) DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `FK_t1d65pcqbqbgg9gt6w734b1y0` (`parent_id`),
  CONSTRAINT `FK_t1d65pcqbqbgg9gt6w734b1y0` FOREIGN KEY (`parent_id`) REFERENCES `t_level` (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;

-- ----------------------------
-- Records of t_level
-- ----------------------------

-- ----------------------------
-- Table structure for t_right
-- ----------------------------
DROP TABLE IF EXISTS `t_right`;
CREATE TABLE `t_right` (
  `id` bigint(20) NOT NULL AUTO_INCREMENT,
  `create_date` datetime DEFAULT NULL,
  `create_user_id` varchar(255) DEFAULT NULL,
  `entity_no` varchar(255) DEFAULT NULL,
  `model_status` varchar(255) DEFAULT NULL,
  `modify_date` datetime DEFAULT NULL,
  `modify_user_id` varchar(255) DEFAULT NULL,
  `product_id` varchar(255) DEFAULT NULL,
  `tenant_id` varchar(255) DEFAULT NULL,
  `comment` varchar(255) DEFAULT NULL,
  `flag` bigint(20) NOT NULL,
  `name` varchar(255) DEFAULT NULL,
  `type` varchar(255) DEFAULT NULL,
  `parent_id` bigint(20) DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `FK_ghqultn3sdyq1ckvpwg7tjq1b` (`parent_id`),
  CONSTRAINT `FK_ghqultn3sdyq1ckvpwg7tjq1b` FOREIGN KEY (`parent_id`) REFERENCES `t_right` (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;

-- ----------------------------
-- Records of t_right
-- ----------------------------

-- ----------------------------
-- Table structure for t_role
-- ----------------------------
DROP TABLE IF EXISTS `t_role`;
CREATE TABLE `t_role` (
  `id` bigint(20) NOT NULL AUTO_INCREMENT,
  `create_date` datetime DEFAULT NULL,
  `create_user_id` varchar(255) DEFAULT NULL,
  `entity_no` varchar(255) DEFAULT NULL,
  `model_status` varchar(255) DEFAULT NULL,
  `modify_date` datetime DEFAULT NULL,
  `modify_user_id` varchar(255) DEFAULT NULL,
  `product_id` varchar(255) DEFAULT NULL,
  `tenant_id` varchar(255) DEFAULT NULL,
  `belong_user_count` int(11) DEFAULT NULL,
  `comment` varchar(255) DEFAULT NULL,
  `name` varchar(255) DEFAULT NULL,
  `role_right_flags` bigint(20) DEFAULT NULL,
  `sub_role_count` int(11) NOT NULL,
  `parent_id` bigint(20) DEFAULT NULL,
  `role_type_id` varchar(32) NOT NULL DEFAULT '',
  PRIMARY KEY (`id`),
  KEY `FK_h1m17ucy0tpm76ku0xotsso24` (`parent_id`),
  CONSTRAINT `FK_h1m17ucy0tpm76ku0xotsso24` FOREIGN KEY (`parent_id`) REFERENCES `t_role` (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;

-- ----------------------------
-- Records of t_role
-- ----------------------------

-- ----------------------------
-- Table structure for t_role_right_relation
-- ----------------------------
DROP TABLE IF EXISTS `t_role_right_relation`;
CREATE TABLE `t_role_right_relation` (
  `t_role` bigint(20) NOT NULL,
  `role_rights` bigint(20) NOT NULL,
  KEY `FK_thr6x2vghqgblsg6fyxq7dh8b` (`role_rights`),
  KEY `FK_of4i58p0ljpvqxe9of5ock0kn` (`t_role`),
  CONSTRAINT `FK_of4i58p0ljpvqxe9of5ock0kn` FOREIGN KEY (`t_role`) REFERENCES `t_role` (`id`),
  CONSTRAINT `FK_thr6x2vghqgblsg6fyxq7dh8b` FOREIGN KEY (`role_rights`) REFERENCES `t_right` (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;

-- ----------------------------
-- Records of t_role_right_relation
-- ----------------------------

-- ----------------------------
-- Table structure for t_system_introduce
-- ----------------------------
DROP TABLE IF EXISTS `t_system_introduce`;
CREATE TABLE `t_system_introduce` (
  `id` bigint(20) NOT NULL AUTO_INCREMENT,
  `create_date` datetime DEFAULT NULL,
  `create_user_id` varchar(255) DEFAULT NULL,
  `entity_no` varchar(255) DEFAULT NULL,
  `model_status` varchar(255) DEFAULT NULL,
  `modify_date` datetime DEFAULT NULL,
  `modify_user_id` varchar(255) DEFAULT NULL,
  `product_id` varchar(255) DEFAULT NULL,
  `tenant_id` varchar(255) DEFAULT NULL,
  `enable` tinyint(1) NOT NULL DEFAULT '1' COMMENT '是否启用',
  `platform` varchar(50) NOT NULL DEFAULT 'Web',
  `type` varchar(100) DEFAULT 'Agent' COMMENT '推广类型，代理Agent 直客StraightGues',
  `name` varchar(255) DEFAULT NULL,
  `url` varchar(255) DEFAULT NULL,
  `display_url` varchar(512) DEFAULT NULL,
  `qr_code` varchar(512) DEFAULT NULL,
  `parameter_type` varchar(255) DEFAULT NULL,
  `bw_user_show` varchar(100) DEFAULT 'UserAllVisible' COMMENT '用户可见范围,UserAllVisible, UserPartVisible, UserNotVisible',
  `visible_user` text COMMENT '可见用户',
  `visible_user_name` text COMMENT '可见用户Name',
  `server_id` varchar(100) DEFAULT NULL,
  `vendor` varchar(100) DEFAULT 'MT4',
  `mt_group` varchar(100) DEFAULT NULL,
  `leverage` int(10) DEFAULT NULL COMMENT '杠杆',
  `account_group` varchar(100) DEFAULT '' COMMENT '账户组',
  `owner_id` varchar(255) DEFAULT NULL,
  `owner_type` varchar(255) DEFAULT NULL,
  `business_code` varchar(255) DEFAULT NULL,
  `os` varchar(100) DEFAULT NULL,
  `software_package` varchar(512) DEFAULT NULL,
  `invisible_user` text,
  `invisible_user_name` text,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=155 DEFAULT CHARSET=utf8;


-- ----------------------------
-- Records of t_system_introduce
-- ----------------------------

-- ----------------------------
-- Table structure for t_user_detail
-- ----------------------------
DROP TABLE IF EXISTS `t_user_detail`;
CREATE TABLE `t_user_detail` (
  `id` bigint(20) NOT NULL AUTO_INCREMENT,
  `create_date` datetime DEFAULT NULL,
  `create_user_id` varchar(255) DEFAULT NULL,
  `entity_no` varchar(255) DEFAULT NULL,
  `model_status` varchar(255) DEFAULT NULL,
  `modify_date` datetime DEFAULT NULL,
  `modify_user_id` varchar(255) DEFAULT NULL,
  `product_id` varchar(255) DEFAULT NULL,
  `tenant_id` varchar(255) DEFAULT NULL,
  `active` int(11) NOT NULL,
  `address` varchar(255) DEFAULT NULL,
  `birthday` varchar(255) DEFAULT NULL,
  `city` varchar(255) DEFAULT NULL,
  `comment` varchar(255) DEFAULT NULL,
  `country` varchar(255) DEFAULT NULL,
  `email` varchar(255) DEFAULT NULL,
  `head_image` varchar(255) DEFAULT NULL,
  `level_id` bigint(20) NOT NULL,
  `level_name` varchar(255) DEFAULT NULL,
  `login` varchar(255) DEFAULT NULL,
  `name` varchar(255) DEFAULT NULL,
  `need_init_pass` bit(1) NOT NULL,
  `nickname` varchar(255) DEFAULT NULL,
  `parent_id` varchar(255) DEFAULT NULL,
  `phone` varchar(255) DEFAULT NULL,
  `postcode` varchar(255) DEFAULT NULL,
  `province` varchar(255) DEFAULT NULL,
  `pub_user_id` varchar(255) DEFAULT NULL,
  `role_id` bigint(20) NOT NULL,
  `role_name` varchar(255) DEFAULT NULL,
  `sex` varchar(255) DEFAULT NULL,
  `sub_user_count` int(11) NOT NULL,
  `username` varchar(255) DEFAULT NULL,
  `vendor_server_id` varchar(255) DEFAULT NULL,
  `version` int(11) NOT NULL,
  `id_type` varchar(255) DEFAULT NULL,
  `id_num` varchar(255) DEFAULT NULL,
  `id_url1` varchar(255) DEFAULT NULL,
  `id_url2` varchar(255) DEFAULT NULL,
  `bank_account` varchar(255) DEFAULT NULL,
  `bank_branch` varchar(255) DEFAULT NULL,
  `account_no` varchar(255) DEFAULT NULL,
  `bank_card_file1` varchar(255) DEFAULT NULL,
  `bank_card_file2` varchar(255) DEFAULT NULL,
  `do_agency_business` varchar(255) DEFAULT NULL,
  `invest_experience` varchar(255) DEFAULT NULL,
  `agent` int(11) DEFAULT NULL,
  `field01` varchar(255) DEFAULT NULL,
  `field02` varchar(255) DEFAULT NULL,
  `field03` varchar(255) DEFAULT NULL,
  `field04` varchar(255) DEFAULT NULL,
  `field05` varchar(255) DEFAULT NULL,
  `field06` varchar(255) DEFAULT NULL,
  `field07` varchar(255) DEFAULT NULL,
  `field08` varchar(255) DEFAULT NULL,
  `field09` varchar(255) DEFAULT NULL,
  `field10` varchar(255) DEFAULT NULL,
  `field11` varchar(255) DEFAULT NULL,
  `field12` varchar(255) DEFAULT NULL,
  `field13` varchar(255) DEFAULT NULL,
  `field14` varchar(255) DEFAULT NULL,
  `field15` varchar(255) DEFAULT NULL,
  `field16` varchar(255) DEFAULT NULL,
  `field17` varchar(255) DEFAULT NULL,
  `field18` varchar(255) DEFAULT NULL,
  `field19` varchar(255) DEFAULT NULL,
  `field20` varchar(255) DEFAULT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;

-- ----------------------------
-- Records of t_user_detail
-- ----------------------------

-- ----------------------------
-- Table structure for t_user_detail_ip_white_black
-- ----------------------------
DROP TABLE IF EXISTS `t_user_detail_ip_white_black`;
CREATE TABLE `t_user_detail_ip_white_black` (
  `id` bigint(20) NOT NULL AUTO_INCREMENT,
  `create_date` datetime DEFAULT NULL,
  `create_user_id` varchar(255) DEFAULT NULL,
  `entity_no` varchar(255) DEFAULT NULL,
  `model_status` varchar(255) DEFAULT NULL,
  `modify_date` datetime DEFAULT NULL,
  `modify_user_id` varchar(255) DEFAULT NULL,
  `product_id` varchar(255) DEFAULT NULL,
  `tenant_id` varchar(255) DEFAULT NULL,
  `enable` bit(1) NOT NULL,
  `from_ip` varchar(255) DEFAULT NULL,
  `to_ip` varchar(255) DEFAULT NULL,
  `user` tinyblob,
  `white` bit(1) NOT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;

-- ----------------------------
-- Records of t_user_detail_ip_white_black
-- ----------------------------

-- ----------------------------
-- Table structure for t_user_group
-- ----------------------------
DROP TABLE IF EXISTS `t_user_group`;
CREATE TABLE `t_user_group` (
  `id` bigint(20) NOT NULL AUTO_INCREMENT,
  `group_name` varchar(255) DEFAULT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;

-- ----------------------------
-- Records of t_user_group
-- ----------------------------

-- ----------------------------
-- Table structure for t_user_list_field
-- ----------------------------
DROP TABLE IF EXISTS `t_user_list_field`;
CREATE TABLE `t_user_list_field` (
  `id` bigint(20) NOT NULL AUTO_INCREMENT,
  `create_date` datetime DEFAULT NULL,
  `create_user_id` varchar(255) DEFAULT NULL,
  `entity_no` varchar(255) DEFAULT NULL,
  `model_status` varchar(255) DEFAULT NULL,
  `modify_date` datetime DEFAULT NULL,
  `modify_user_id` varchar(255) DEFAULT NULL,
  `product_id` varchar(255) DEFAULT NULL,
  `tenant_id` varchar(255) DEFAULT NULL,
  `model` varchar(255) DEFAULT NULL,
  `name` varchar(255) DEFAULT NULL,
  `position` int(11) NOT NULL,
  `user_id` bigint(20) NOT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;

-- ----------------------------
-- Records of t_user_list_field
-- ----------------------------

-- ----------------------------

