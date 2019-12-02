create database monitor
  DEFAULT CHARACTER SET utf8
  DEFAULT COLLATE utf8_general_ci;
USE monitor;
SET NAMES utf8;

/* ---------------------------------------------uic--------------------------------------*/
DROP TABLE if exists team;
CREATE TABLE `team` (
  `id` int(10) unsigned NOT NULL AUTO_INCREMENT,
  `name` varchar(64) NOT NULL,
  `resume` varchar(255) not null default '',
  `creator` int(10) unsigned NOT NULL DEFAULT '0',
  `created` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `idx_team_name` (`name`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;

/**
 * role: -1:blocked 0:normal 1:admin 2:root
 */
DROP TABLE if exists `user`;
CREATE TABLE `user` (
  `id` int(10) unsigned NOT NULL AUTO_INCREMENT,
  `name` varchar(64) NOT NULL,
  `passwd` varchar(64) not null default '',
  `cnname` varchar(128) not null default '',
  `email` varchar(255) not null default '',
  `phone` varchar(16) not null default '',
  `im` varchar(32) not null default '',
  `qq` varchar(16) not null default '',
  `role` tinyint not null default 0,
  `creator` int(10) unsigned NOT NULL DEFAULT 0,
  `created` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `idx_user_name` (`name`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;

DROP TABLE if exists `rel_team_user`;
CREATE TABLE `rel_team_user` (
  `id` int(10) unsigned NOT NULL AUTO_INCREMENT,
  `tid` int(10) unsigned not null,
  `uid` int(10) unsigned not null,
  PRIMARY KEY (`id`),
  KEY `idx_rel_tid` (`tid`),
  KEY `idx_rel_uid` (`uid`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;

DROP TABLE if exists `session`;
CREATE TABLE `session` (
  `id` int(10) unsigned NOT NULL AUTO_INCREMENT,
  `uid` int(10) unsigned not null,
  `sig` varchar(32) not null,
  `expired` int(10) unsigned not null,
  PRIMARY KEY (`id`),
  KEY `idx_session_uid` (`uid`),
  KEY `idx_session_sig` (`sig`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;

/*900150983cd24fb0d6963f7d28e17f72*/
/*insert into `user`(`name`, `passwd`, `role`, `created`) values('root', md5('abc'), 2, now());*/

/*---------------------------------------------portal -----------------------------------------*/

/**
 * 这里的机器是从机器管理系统中同步过来的
 * 系统拿出来单独部署需要为hbs增加功能，心跳上来的机器写入host表
 */
DROP TABLE IF EXISTS host;
CREATE TABLE host
(
    id             INT UNSIGNED NOT NULL AUTO_INCREMENT,
    hostname       VARCHAR(255) NOT NULL DEFAULT '',
    ip             VARCHAR(16)  NOT NULL DEFAULT '',
    agent_version  VARCHAR(16)  NOT NULL DEFAULT '',
    plugin_version VARCHAR(128) NOT NULL DEFAULT '',
    maintain_begin INT UNSIGNED NOT NULL DEFAULT 0,
    maintain_end   INT UNSIGNED NOT NULL DEFAULT 0,
    update_at      TIMESTAMP    NOT NULL DEFAULT CURRENT_TIMESTAMP on update CURRENT_TIMESTAMP,
    PRIMARY KEY (id),
    UNIQUE KEY idx_host_hostname (hostname)
)
    ENGINE =InnoDB
    DEFAULT CHARSET =utf8
    COLLATE =utf8_unicode_ci;


/**
 * 机器分组信息
 * come_from 0: 从机器管理同步过来的；1: 从页面创建的
 */
DROP TABLE IF EXISTS grp;
CREATE TABLE `grp` (
                       id          INT(10) UNSIGNED NOT NULL AUTO_INCREMENT,
                       grp_name    VARCHAR(255)     NOT NULL DEFAULT '',
                       create_user VARCHAR(64)      NOT NULL DEFAULT '',
                       create_at   TIMESTAMP        NOT NULL DEFAULT CURRENT_TIMESTAMP,
                       come_from   TINYINT(4)       NOT NULL DEFAULT '0',
                       PRIMARY KEY (id),
                       UNIQUE KEY idx_host_grp_grp_name (grp_name)
)
    ENGINE =InnoDB
    DEFAULT CHARSET =utf8
    COLLATE =utf8_unicode_ci;


DROP TABLE IF EXISTS grp_host;
CREATE TABLE grp_host
(
    grp_id  INT UNSIGNED NOT NULL,
    host_id INT UNSIGNED NOT NULL,
    KEY idx_grp_host_grp_id (grp_id),
    KEY idx_grp_host_host_id (host_id)
)
    ENGINE =InnoDB
    DEFAULT CHARSET =utf8
    COLLATE =utf8_unicode_ci;


/**
 * 监控策略模板
 * tpl_name全局唯一，命名的时候可以适当带上一些前缀，比如：sa.falcon.base
 */
DROP TABLE IF EXISTS tpl;
CREATE TABLE tpl
(
    id          INT UNSIGNED NOT NULL AUTO_INCREMENT,
    tpl_name    VARCHAR(255) NOT NULL DEFAULT '',
    parent_id   INT UNSIGNED NOT NULL DEFAULT 0,
    action_id   INT UNSIGNED NOT NULL DEFAULT 0,
    create_user VARCHAR(64)  NOT NULL DEFAULT '',
    create_at   TIMESTAMP    NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (id),
    UNIQUE KEY idx_tpl_name (tpl_name),
    KEY idx_tpl_create_user (create_user)
)
    ENGINE =InnoDB
    DEFAULT CHARSET =utf8
    COLLATE =utf8_unicode_ci;


DROP TABLE IF EXISTS strategy;
CREATE TABLE `strategy` (
                            `id`          INT(10) UNSIGNED NOT NULL AUTO_INCREMENT,
                            `metric`      VARCHAR(128)     NOT NULL DEFAULT '',
                            `tags`        VARCHAR(256)     NOT NULL DEFAULT '',
                            `max_step`    INT(11)          NOT NULL DEFAULT '1',
                            `priority`    TINYINT(4)       NOT NULL DEFAULT '0',
                            `func`        VARCHAR(16)      NOT NULL DEFAULT 'all(#1)',
                            `op`          VARCHAR(8)       NOT NULL DEFAULT '',
                            `right_value` VARCHAR(64)      NOT NULL,
                            `note`        VARCHAR(128)     NOT NULL DEFAULT '',
                            `run_begin`   VARCHAR(16)      NOT NULL DEFAULT '',
                            `run_end`     VARCHAR(16)      NOT NULL DEFAULT '',
                            `tpl_id`      INT(10) UNSIGNED NOT NULL DEFAULT '0',
                            PRIMARY KEY (`id`),
                            KEY `idx_strategy_tpl_id` (`tpl_id`)
)
    ENGINE =InnoDB
    DEFAULT CHARSET =utf8
    COLLATE =utf8_unicode_ci;


DROP TABLE IF EXISTS expression;
CREATE TABLE `expression` (
                              `id`          INT(10) UNSIGNED NOT NULL AUTO_INCREMENT,
                              `expression`  VARCHAR(1024)    NOT NULL,
                              `func`        VARCHAR(16)      NOT NULL DEFAULT 'all(#1)',
                              `op`          VARCHAR(8)       NOT NULL DEFAULT '',
                              `right_value` VARCHAR(16)      NOT NULL DEFAULT '',
                              `max_step`    INT(11)          NOT NULL DEFAULT '1',
                              `priority`    TINYINT(4)       NOT NULL DEFAULT '0',
                              `note`        VARCHAR(1024)    NOT NULL DEFAULT '',
                              `action_id`   INT(10) UNSIGNED NOT NULL DEFAULT '0',
                              `create_user` VARCHAR(64)      NOT NULL DEFAULT '',
                              `pause`       TINYINT(1)       NOT NULL DEFAULT '0',
                              PRIMARY KEY (`id`)
)
    ENGINE =InnoDB
    DEFAULT CHARSET =utf8
    COLLATE =utf8_unicode_ci;


DROP TABLE IF EXISTS grp_tpl;
CREATE TABLE `grp_tpl` (
                           `grp_id`    INT(10) UNSIGNED NOT NULL,
                           `tpl_id`    INT(10) UNSIGNED NOT NULL,
                           `bind_user` VARCHAR(64)      NOT NULL DEFAULT '',
                           KEY `idx_grp_tpl_grp_id` (`grp_id`),
                           KEY `idx_grp_tpl_tpl_id` (`tpl_id`)
)
    ENGINE =InnoDB
    DEFAULT CHARSET =utf8
    COLLATE =utf8_unicode_ci;

CREATE TABLE `plugin_dir` (
                              `id`          INT(10) UNSIGNED NOT NULL AUTO_INCREMENT,
                              `grp_id`      INT(10) UNSIGNED NOT NULL,
                              `dir`         VARCHAR(255)     NOT NULL,
                              `create_user` VARCHAR(64)      NOT NULL DEFAULT '',
                              `create_at`   TIMESTAMP        NOT NULL DEFAULT CURRENT_TIMESTAMP,
                              PRIMARY KEY (`id`),
                              KEY `idx_plugin_dir_grp_id` (`grp_id`)
)
    ENGINE =InnoDB
    DEFAULT CHARSET =utf8
    COLLATE =utf8_unicode_ci;


DROP TABLE IF EXISTS action;
CREATE TABLE `action` (
                          `id`                   INT(10) UNSIGNED NOT NULL AUTO_INCREMENT,
                          `uic`                  VARCHAR(255)     NOT NULL DEFAULT '',
                          `url`                  VARCHAR(255)     NOT NULL DEFAULT '',
                          `callback`             TINYINT(4)       NOT NULL DEFAULT '0',
                          `before_callback_sms`  TINYINT(4)       NOT NULL DEFAULT '0',
                          `before_callback_mail` TINYINT(4)       NOT NULL DEFAULT '0',
                          `after_callback_sms`   TINYINT(4)       NOT NULL DEFAULT '0',
                          `after_callback_mail`  TINYINT(4)       NOT NULL DEFAULT '0',
                          PRIMARY KEY (`id`)
)
    ENGINE =InnoDB
    DEFAULT CHARSET =utf8
    COLLATE =utf8_unicode_ci;

/**
 * nodata mock config
 */
DROP TABLE IF EXISTS `mockcfg`;
CREATE TABLE `mockcfg` (
                           `id`       BIGINT(20) UNSIGNED NOT NULL AUTO_INCREMENT,
                           `name`     VARCHAR(255) NOT NULL DEFAULT '' COMMENT 'name of mockcfg, used for uuid',
                           `obj`      VARCHAR(10240) NOT NULL DEFAULT '' COMMENT 'desc of object',
                           `obj_type` VARCHAR(255) NOT NULL DEFAULT '' COMMENT 'type of object, host or group or other',
                           `metric`   VARCHAR(128) NOT NULL DEFAULT '',
                           `tags`     VARCHAR(1024) NOT NULL DEFAULT '',
                           `dstype`   VARCHAR(32)  NOT NULL DEFAULT 'GAUGE',
                           `step`     INT(11) UNSIGNED  NOT NULL DEFAULT 60,
                           `mock`     DOUBLE  NOT NULL DEFAULT 0  COMMENT 'mocked value when nodata occurs',
                           `creator`  VARCHAR(64)  NOT NULL DEFAULT '',
                           `t_create` DATETIME NOT NULL COMMENT 'create time',
                           `t_modify` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT 'last modify time',
                           PRIMARY KEY (`id`),
                           UNIQUE KEY `uniq_name` (`name`)
)
    ENGINE=InnoDB
    DEFAULT CHARSET=utf8
    COLLATE=utf8_unicode_ci;

/**
 *  aggregator cluster metric config table
 */
DROP TABLE IF EXISTS `cluster`;
CREATE TABLE `cluster` (
                           `id`          INT UNSIGNED   NOT NULL AUTO_INCREMENT,
                           `grp_id`      INT            NOT NULL,
                           `numerator`   VARCHAR(10240) NOT NULL,
                           `denominator` VARCHAR(10240) NOT NULL,
                           `endpoint`    VARCHAR(255)   NOT NULL,
                           `metric`      VARCHAR(255)   NOT NULL,
                           `tags`        VARCHAR(255)   NOT NULL,
                           `ds_type`     VARCHAR(255)   NOT NULL,
                           `step`        INT            NOT NULL,
                           `last_update` TIMESTAMP      NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
                           `creator`     VARCHAR(255)   NOT NULL,
                           PRIMARY KEY (`id`)
)
    ENGINE =InnoDB
    DEFAULT CHARSET=utf8
    COLLATE=utf8_unicode_ci;

/**
 * alert links
 */
DROP TABLE IF EXISTS alert_link;
CREATE TABLE alert_link
(
    id        INT UNSIGNED NOT NULL AUTO_INCREMENT,
    path      VARCHAR(16)  NOT NULL DEFAULT '',
    content   TEXT         NOT NULL,
    create_at TIMESTAMP    NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (id),
    UNIQUE KEY alert_path(path)
)
    ENGINE =InnoDB
    DEFAULT CHARSET =utf8
    COLLATE =utf8_unicode_ci;

/*      ------------------------------------dashboard----------------------------------  */


DROP TABLE IF EXISTS `dashboard_graph`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `dashboard_graph` (
                                   `id` int(11) unsigned NOT NULL AUTO_INCREMENT,
                                   `title` char(128) NOT NULL,
                                   `hosts` varchar(10240) NOT NULL DEFAULT '',
                                   `counters` varchar(1024) NOT NULL DEFAULT '',
                                   `screen_id` int(11) unsigned NOT NULL,
                                   `timespan` int(11) unsigned NOT NULL DEFAULT '3600',
                                   `graph_type` char(2) NOT NULL DEFAULT 'h',
                                   `method` char(8) DEFAULT '',
                                   `position` int(11) unsigned NOT NULL DEFAULT '0',
                                   `falcon_tags` varchar(512) NOT NULL DEFAULT '',
                                   PRIMARY KEY (`id`),
                                   KEY `idx_sid` (`screen_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `dashboard_screen`
--

DROP TABLE IF EXISTS `dashboard_screen`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `dashboard_screen` (
                                    `id` int(11) unsigned NOT NULL AUTO_INCREMENT,
                                    `pid` int(11) unsigned NOT NULL DEFAULT '0',
                                    `name` char(128) NOT NULL,
                                    `time` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
                                    PRIMARY KEY (`id`),
                                    KEY `idx_pid` (`pid`),
                                    UNIQUE KEY `idx_pid_n` (`pid`,`name`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `tmp_graph`
--

DROP TABLE IF EXISTS `tmp_graph`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `tmp_graph` (
                             `id` int(11) unsigned NOT NULL AUTO_INCREMENT,
                             `endpoints` varchar(10240) NOT NULL DEFAULT '',
                             `counters` varchar(10240) NOT NULL DEFAULT '',
                             `ck` varchar(32) NOT NULL,
                             `time_` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
                             PRIMARY KEY (`id`),
                             UNIQUE KEY `idx_ck` (`ck`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;
/*!40101 SET character_set_client = @saved_cs_client */;
/*!40103 SET TIME_ZONE=@OLD_TIME_ZONE */;

/*!40101 SET SQL_MODE=@OLD_SQL_MODE */;
/*!40014 SET FOREIGN_KEY_CHECKS=@OLD_FOREIGN_KEY_CHECKS */;
/*!40014 SET UNIQUE_CHECKS=@OLD_UNIQUE_CHECKS */;
/*!40101 SET CHARACTER_SET_CLIENT=@OLD_CHARACTER_SET_CLIENT */;
/*!40101 SET CHARACTER_SET_RESULTS=@OLD_CHARACTER_SET_RESULTS */;
/*!40101 SET COLLATION_CONNECTION=@OLD_COLLATION_CONNECTION */;
/*!40111 SET SQL_NOTES=@OLD_SQL_NOTES */;

/*-----------------------------------------------graph---------------------------------------*/

DROP TABLE if exists `endpoint`;
CREATE TABLE `endpoint` (
                                    `id` int(10) unsigned NOT NULL AUTO_INCREMENT,
                                    `endpoint` varchar(255) NOT NULL DEFAULT '',
                                    `ts` int(11) DEFAULT NULL,
                                    `t_create` DATETIME NOT NULL COMMENT 'create time',
                                    `t_modify` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT 'last modify time',
                                    PRIMARY KEY (`id`),
                                    UNIQUE KEY `idx_endpoint` (`endpoint`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;

DROP TABLE if exists `endpoint_counter`;
CREATE TABLE `endpoint_counter` (
                                            `id` int(10) unsigned NOT NULL AUTO_INCREMENT,
                                            `endpoint_id` int(10) unsigned NOT NULL,
                                            `counter` varchar(255) NOT NULL DEFAULT '',
                                            `step` int(11) not null default 60 comment 'in second',
                                            `type` varchar(16) not null comment 'GAUGE|COUNTER|DERIVE',
                                            `ts` int(11) DEFAULT NULL,
                                            `t_create` DATETIME NOT NULL COMMENT 'create time',
                                            `t_modify` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT 'last modify time',
                                            PRIMARY KEY (`id`),
                                            UNIQUE KEY `idx_endpoint_id_counter` (`endpoint_id`, `counter`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;

DROP TABLE if exists `tag_endpoint`;
CREATE TABLE `tag_endpoint` (
                                        `id` int(10) unsigned NOT NULL AUTO_INCREMENT,
                                        `tag` varchar(255) NOT NULL DEFAULT '' COMMENT 'srv=tv',
                                        `endpoint_id` int(10) unsigned NOT NULL,
                                        `ts` int(11) DEFAULT NULL,
                                        `t_create` DATETIME NOT NULL COMMENT 'create time',
                                        `t_modify` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT 'last modify time',
                                        PRIMARY KEY (`id`),
                                        UNIQUE KEY `idx_tag_endpoint_id` (`tag`, `endpoint_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;


/* --------------------------------------------------alarms------------------------------------------*/

/*
* 建立告警归档资料表, 主要存储各个告警的最后触发状况
*/
DROP TABLE IF EXISTS event_cases;
CREATE TABLE IF NOT EXISTS event_cases(
                                          id VARCHAR(50),
                                          endpoint VARCHAR(100) NOT NULL,
                                          metric VARCHAR(200) NOT NULL,
                                          func VARCHAR(50),
                                          cond VARCHAR(200) NOT NULL,
                                          note VARCHAR(500),
                                          max_step int(10) unsigned,
                                          current_step int(10) unsigned,
                                          priority INT(6) NOT NULL,
                                          status VARCHAR(20) NOT NULL,
                                          timestamp Timestamp NOT NULL,
                                          update_at Timestamp NULL DEFAULT NULL,
                                          closed_at Timestamp NULL DEFAULT NULL,
                                          closed_note VARCHAR(250),
                                          user_modified int(10) unsigned,
                                          tpl_creator VARCHAR(64),
                                          expression_id int(10) unsigned,
                                          strategy_id int(10) unsigned,
                                          template_id int(10) unsigned,
                                          process_note MEDIUMINT,
                                          process_status VARCHAR(20) DEFAULT 'unresolved',
                                          PRIMARY KEY (id),
                                          INDEX (endpoint, strategy_id, template_id)
)
    ENGINE =InnoDB
    DEFAULT CHARSET =utf8;


/*
* 建立告警归档资料表, 存储各个告警触发状况的历史状态
*/
DROP TABLE IF EXISTS events;
CREATE TABLE IF NOT EXISTS events (
                                      id int(10) NOT NULL AUTO_INCREMENT,
                                      event_caseId VARCHAR(50),
                                      step int(10) unsigned,
                                      cond VARCHAR(200) NOT NULL,
                                      status int(3) unsigned DEFAULT 0,
                                      timestamp Timestamp,
                                      PRIMARY KEY (id),
                                      INDEX(event_caseId),
                                      FOREIGN KEY (event_caseId) REFERENCES event_cases(id)
                                          ON DELETE CASCADE
                                          ON UPDATE CASCADE
)
    ENGINE =InnoDB
    DEFAULT CHARSET =utf8;

/*
* 告警留言表
*/
CREATE TABLE IF NOT EXISTS event_note (
                                          id MEDIUMINT NOT NULL AUTO_INCREMENT,
                                          event_caseId VARCHAR(50),
                                          note    VARCHAR(300),
                                          case_id VARCHAR(20),
                                          status VARCHAR(15),
                                          timestamp Timestamp,
                                          user_id int(10) unsigned,
                                          PRIMARY KEY (id),
                                          INDEX (event_caseId),
                                          FOREIGN KEY (event_caseId) REFERENCES event_cases(id)
                                              ON DELETE CASCADE
                                              ON UPDATE CASCADE,
                                          FOREIGN KEY (user_id) REFERENCES user(id)
                                              ON DELETE CASCADE
                                              ON UPDATE CASCADE
);