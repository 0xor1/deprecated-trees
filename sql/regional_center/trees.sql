DROP DATABASE IF EXISTS trees;
CREATE DATABASE trees;
USE trees;

DROP TABLE IF EXISTS members;
CREATE TABLE members(
	org BINARY(16) NOT NULL,
	id BINARY(16) NOT NULL,
    name VARCHAR(50) NOT NULL,
    accessTask BINARY(16) NULL,
    totalRemainingTime BIGINT UNSIGNED NOT NULL,
    totalLoggedTime BIGINT UNSIGNED NOT NULL,
    isActive BOOL NOT NULL,
    isDeleted BOOL NOT NULL,
    role VARCHAR(10) NOT NULL,
    PRIMARY KEY (org, isActive, name),
    UNIQUE INDEX (org, id)
);

DROP TABLE IF EXISTS tasks;
CREATE TABLE tasks(
	org BINARY(16) NOT NULL,
	id BINARY(16) NOT NULL,
    name VARCHAR(50) NOT NULL,
    user BINARY(16) NULL,
    totalRemainingTime BIGINT UNSIGNED NOT NULL,
    totalLoggedTime BIGINT UNSIGNED NOT NULL,
    chatCount BIGINT UNSIGNED NOT NULL,
    fileCount BIGINT UNSIGNED NOT NULL,
    fileSize BIGINT UNSIGNED NOT NULL,
    created DATETIME NOT NULL,
    isAbstractTask BOOL NOT NULL,
    PRIMARY KEY (org, id)
);

DROP TABLE IF EXISTS abstractTasks;
CREATE TABLE abstractTasks(
	org BINARY(16) NOT NULL,
	id BINARY(16) NOT NULL,
	minimumRemainingTime BIGINT UNSIGNED NOT NULL,
	isParallel BOOL NOT NULL,
	childCount SMALLINT UNSIGNED NOT NULL,
	taskCount BIGINT UNSIGNED NOT NULL,
	subFileCount BIGINT UNSIGNED NOT NULL,
	subFileSize BIGINT UNSIGNED NOT NULL,
	archivedChildCount BIGINT UNSIGNED NOT NULL,
	archivedTaskCount BIGINT UNSIGNED NOT NULL,
	archivedSubFileCount BIGINT UNSIGNED NOT NULL,
	archivedSubFileSize BIGINT UNSIGNED NOT NULL,
    PRIMARY KEY (org, id)
);


DROP USER IF EXISTS 'tc_rc_trees'@'%';
CREATE USER 'tc_rc_trees'@'%' IDENTIFIED BY 'T@sk-C3n-T3r-Tr335';
GRANT SELECT ON accounts.* TO 'tc_rc_trees'@'%';
GRANT INSERT ON accounts.* TO 'tc_rc_trees'@'%';
GRANT UPDATE ON accounts.* TO 'tc_rc_trees'@'%';
GRANT DELETE ON accounts.* TO 'tc_rc_trees'@'%';
GRANT EXECUTE ON accounts.* TO 'tc_rc_trees'@'%';