DROP DATABASE IF EXISTS trees;
CREATE DATABASE trees;
USE trees;

DROP TABLE IF EXISTS orgs;
CREATE TABLE orgs(
	id BINARY(16) NOT NULL,
    PRIMARY KEY (id)
);

DROP TABLE IF EXISTS orgMembers;
CREATE TABLE orgMembers(
	org BINARY(16) NOT NULL,
	id BINARY(16) NOT NULL,
    name VARCHAR(50) NOT NULL,
    totalRemainingTime BIGINT UNSIGNED NOT NULL DEFAULT 0,
    totalLoggedTime BIGINT UNSIGNED NOT NULL DEFAULT 0,
    isActive BOOL NOT NULL DEFAULT TRUE,
    role TINYINT UNSIGNED NOT NULL DEFAULT 2, #0 owner, 1 admin, 2 memberOfAllProjects, 3 memberOfOnlySpecificProjects
    PRIMARY KEY (org, isActive, role, name),
    UNIQUE INDEX (org, id)
);

DROP TABLE IF EXISTS orgActivities;
CREATE TABLE orgActivities(
	org BINARY(16) NOT NULL,
    id BINARY(16) NOT NULL,
    occurredOn DATETIME NOT NULL,
    item BINARY(16) NOT NULL,
    member BINARY(16) NOT NULL,
    itemType VARCHAR(100) NOT NULL,
    itemName VARCHAR(250) NOT NULL,
    action VARCHAR(100) NOT NULL,
    PRIMARY KEY (org, occurredOn, item, member),
    UNIQUE INDEX (org, item, occurredOn, member),
    UNIQUE INDEX (org, member, occurredOn, item)
);

DROP TABLE IF EXISTS projectMembers;
CREATE TABLE projectMembers(
	org BINARY(16) NOT NULL,
	project BINARY(16) NOT NULL,
    member BINARY(16) NOT NULL,
    totalRemainingTime BIGINT UNSIGNED NOT NULL,
    totalLoggedTime BIGINT UNSIGNED NOT NULL,
    role TINYINT UNSIGNED NOT NULL, #0 admin, 1 writer, 2 reader
    PRIMARY KEY (org, project, role, member),
    UNIQUE INDEX (org, project, member)
);

DROP TABLE IF EXISTS projectActivities;
CREATE TABLE projectActivities(
	org BINARY(16) NOT NULL,
    project BINARY(16) NOT NULL,
    occurredOn DATETIME NOT NULL,
    item BINARY(16) NOT NULL,
    member BINARY(16) NOT NULL,
    itemType VARCHAR(100) NOT NULL,
    itemName VARCHAR(250) NOT NULL,
    action VARCHAR(100) NOT NULL,
    PRIMARY KEY (org, project, occurredOn, item, member),
    UNIQUE INDEX (org, project, item, occurredOn, member),
    UNIQUE INDEX (org, project, member, occurredOn, item),
    UNIQUE INDEX (org, occurredOn, project, item, member)
);

DROP TABLE IF EXISTS projects;
CREATE TABLE projects(
	org BINARY(16) NOT NULL,
    id BINARY(16) NOT NULL,
    firstChild BINARY(16) NULL,
	name VARCHAR(250) NOT NULL,
	description VARCHAR(1250) NOT NULL,
    createdOn DATETIME NOT NULL,
    archivedOn DateTime NULL,
    startOn DATETIME NULL,
    dueOn DATETIME NULL,
    totalRemainingTime BIGINT UNSIGNED NOT NULL,
    totalLoggedTime BIGINT UNSIGNED NOT NULL,
    minimumRemainingTime BIGINT UNSIGNED NOT NULL,
    fileCount BIGINT UNSIGNED NOT NULL,
    fileSize BIGINT UNSIGNED NOT NULL,
    linkedFileCount BIGINT UNSIGNED NOT NULL,
    chatCount BIGINT UNSIGNED NOT NULL,
    isParallel BOOL NOT NULL DEFAULT FALSE,
    isPublic BOOL NOT NULL DEFAULT FALSE,
    PRIMARY KEY (org, id),
    INDEX(org, archivedOn, name, createdOn, id),
    INDEX(org, archivedOn, createdOn, name, id),
    INDEX(org, archivedOn, startOn, name, id),
    INDEX(org, archivedOn, dueOn, name, id)
);

DROP TABLE IF EXISTS nodes;
CREATE TABLE nodes(
	org BINARY(16) NOT NULL,
	project BINARY(16) NOT NULL,
    id BINARY(16) NOT NULL,
    parent BINARY(16) NOT NULL,
    firstChild BINARY(16) NULL,
    nextSibling BINARY(16) NULL,
    isAbstract BOOL NOT NULL,
	name VARCHAR(250) NOT NULL,
	description VARCHAR(1250) NOT NULL,
    createdOn DATETIME NOT NULL,
    totalRemainingTime BIGINT UNSIGNED NOT NULL,
    totalLoggedTime BIGINT UNSIGNED NOT NULL,
    minimumRemainingTime BIGINT UNSIGNED NULL,
    linkedFileCount INT UNSIGNED NOT NULL,
    chatCount BIGINT UNSIGNED NOT NULL,
    isParallel BOOL NOT NULL DEFAULT FALSE,
    member BINARY(16) NULL,
    PRIMARY KEY (org, project, id),
    UNIQUE INDEX(org, project, parent, id),
    UNIQUE INDEX(org, member, project, id),
    UNIQUE INDEX(org, project, member, id)
);

DROP PROCEDURE IF EXISTS registerAccount;
DELIMITER $$
CREATE PROCEDURE registerAccount(_id BINARY(16), _ownerId BINARY(16), _ownerName VARCHAR(50))
BEGIN
	INSERT INTO orgs (id) VALUES (_id);
    INSERT INTO orgMembers (org, id, name, totalRemainingTime, totalLoggedTime, isActive, role) VALUES (_id, _ownerId, _ownerName, 0, 0, true, 0);
END;
$$
DELIMITER ;

DROP PROCEDURE IF EXISTS deleteAccount;
DELIMITER $$
CREATE PROCEDURE deleteAccount(_id BINARY(16))
BEGIN
	DELETE FROM orgs WHERE id =_id;
	DELETE FROM orgMembers WHERE org =_id;
	DELETE FROM orgActivities WHERE org =_id;
	DELETE FROM projectMembers WHERE org =_id;
	DELETE FROM projectActivities WHERE org =_id;
	DELETE FROM projects WHERE org =_id;
END;
$$
DELIMITER ;

DROP USER IF EXISTS 'tc_rc_trees'@'%';
CREATE USER 'tc_rc_trees'@'%' IDENTIFIED BY 'T@sk-C3n-T3r-Tr335';
GRANT SELECT ON trees.* TO 'tc_rc_trees'@'%';
GRANT INSERT ON trees.* TO 'tc_rc_trees'@'%';
GRANT UPDATE ON trees.* TO 'tc_rc_trees'@'%';
GRANT DELETE ON trees.* TO 'tc_rc_trees'@'%';
GRANT EXECUTE ON trees.* TO 'tc_rc_trees'@'%';